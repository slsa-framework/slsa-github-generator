// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/coreos/go-oidc"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

var defaultActionsProviderURL = "https://token.actions.githubusercontent.com"

const (
	requestTokenEnvKey = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	requestURLEnvKey   = "ACTIONS_ID_TOKEN_REQUEST_URL"
)

// OIDCToken represents the contents of a Github OIDC JWT token.
type OIDCToken struct {
	// Issuer is the token issuer.
	Issuer string

	// Audience is the audience for which the token was granted.
	Audience []string

	// Expiry is the expiration date of the token.
	Expiry time.Time

	// JobWorkflowRef is a reference to the current job workflow.
	JobWorkflowRef string `json:"job_workflow_ref"`
}

// errRequestError indicates an error requesting the token from the issuer.
type errRequestError struct {
	errors.WrappableError
}

// errToken indicates an error in the format of the token.
type errToken struct {
	errors.WrappableError
}

// errVerify indicates an error in the token verification process.
type errVerify struct {
	errors.WrappableError
}

// OIDCClient is a client for the GitHub OIDC provider.
type OIDCClient struct {
	// requestURL is the GitHub URL to request a OIDC token.
	requestURL *url.URL

	// bearerToken is used to request an ID token.
	bearerToken string

	// verifierFunc is a factory to generate an oidc.IDTokenVerifier for token verification.
	// This is used for tests.
	verifierFunc func(context.Context) (*oidc.IDTokenVerifier, error)
}

// NewOIDCClient returns new GitHub OIDC provider client.
func NewOIDCClient() (*OIDCClient, error) {
	requestURL := os.Getenv(requestURLEnvKey)
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, fmt.Errorf("invalid request URL %q: %w", requestURL, err)
	}

	c := OIDCClient{
		requestURL:  parsedURL,
		bearerToken: os.Getenv(requestTokenEnvKey),
	}
	c.verifierFunc = func(ctx context.Context) (*oidc.IDTokenVerifier, error) {
		provider, err := oidc.NewProvider(ctx, defaultActionsProviderURL)
		if err != nil {
			return nil, err
		}
		return provider.Verifier(&oidc.Config{
			// NOTE: Disable ClientID check.
			// ClientID is normally checked to be part of the audience but we
			// don't use a ClientID when requesting a token.
			SkipClientIDCheck: true,
		}), nil
	}
	return &c, nil
}

func (c *OIDCClient) newRequestURL(audience []string) string {
	requestURL := *c.requestURL
	q := requestURL.Query()
	for _, a := range audience {
		q.Add("audience", a)
	}
	requestURL.RawQuery = q.Encode()
	return requestURL.String()
}

func (c *OIDCClient) requestToken(ctx context.Context, audience []string) ([]byte, error) {
	// Request the token.
	req, err := http.NewRequest("GET", c.newRequestURL(audience), nil)
	if err != nil {
		return nil, errors.Errorf(&errRequestError{}, "creating request: %w", err)
	}
	req.Header.Add("Authorization", "bearer "+c.bearerToken)
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf(&errRequestError{}, "request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response.
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf(&errRequestError{}, "reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Errorf(&errRequestError{}, "response: %s: %s", resp.Status, string(b))
	}
	return b, nil
}

func (c *OIDCClient) decodePayload(b []byte) (string, error) {
	// Extract the raw token from JSON payload.
	var payload struct {
		Value string `json:"value"`
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(&payload); err != nil {
		return "", errors.Errorf(&errToken{}, "parsing JSON: %w", err)
	}
	return payload.Value, nil
}

// verifyToken verifies the token contents and signature.
func (c *OIDCClient) verifyToken(ctx context.Context, audience []string, payload string) (*oidc.IDToken, error) {
	// Verify the token.
	verifier, err := c.verifierFunc(ctx)
	if err != nil {
		return nil, errors.Errorf(&errVerify{}, "creating verifier: %w", err)
	}

	t, err := verifier.Verify(ctx, payload)
	if err != nil {
		return nil, errors.Errorf(&errVerify{}, "could not verify token: %w", err)
	}

	// Verify the audience received is the one we requested.
	if !compareStringSlice(audience, t.Audience) {
		return nil, errors.Errorf(&errVerify{}, "audience not equal %q != %q", audience, t.Audience)
	}

	return t, nil
}

func (c *OIDCClient) decodeToken(token *oidc.IDToken) (*OIDCToken, error) {
	var t OIDCToken
	t.Issuer = token.Issuer
	t.Audience = token.Audience
	t.Expiry = token.Expiry

	if err := token.Claims(&t); err != nil {
		return nil, errors.Errorf(&errToken{}, "getting claims: %w", err)
	}

	return &t, nil
}

// Token requests an OIDC token from Github's provider, verifies it, and
// returns the token.
func (c *OIDCClient) Token(ctx context.Context, audience []string) (*OIDCToken, error) {
	tokenBytes, err := c.requestToken(ctx, audience)
	if err != nil {
		return nil, err
	}

	tokenPayload, err := c.decodePayload(tokenBytes)
	if err != nil {
		return nil, err
	}

	t, err := c.verifyToken(ctx, audience, tokenPayload)
	if err != nil {
		return nil, err
	}

	token, err := c.decodeToken(t)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func compareStringSlice(s1, s2 []string) bool {
	// Verify the audience received is the one we requested.
	if len(s1) != len(s2) {
		return false
	}

	c1 := make([]string, len(s1))
	copy(s1, c1)
	sort.Strings(c1)

	c2 := make([]string, len(s2))
	copy(s2, c2)
	sort.Strings(c2)

	for i := range c1 {
		if c1[i] != c2[i] {
			return false
		}
	}

	return true
}
