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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

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
	// JobWorkflowRef is a reference to the current job workflow.
	JobWorkflowRef string `json:"job_workflow_ref"`

	// rawToken holds the full JWT token in compact serialization format for
	// later verification.
	rawToken string
}

type errRequestError struct {
	errors.WrappableError
}

type errToken struct {
	errors.WrappableError
}

type errVerify struct {
	errors.WrappableError
}

// OIDCClient is a client for the GitHub OIDC provider.
type OIDCClient struct {
	actionsProviderURL string
	requestURL         *url.URL
	conf               *oidc.Config
}

// NewOIDCClient returns new GitHub OIDC provider client.
func NewOIDCClient(audience string) (*OIDCClient, error) {
	parsedURL, err := url.Parse(os.Getenv(requestURLEnvKey))
	if err != nil {
		return nil, fmt.Errorf("invalid request URL %q: %w", parsedURL, err)
	}
	q := parsedURL.Query()
	q.Add("audience", audience)
	parsedURL.RawQuery = q.Encode()

	return &OIDCClient{
		actionsProviderURL: defaultActionsProviderURL,
		requestURL:         parsedURL,
		conf: &oidc.Config{
			ClientID: audience,
		},
	}, nil
}

// Token requests an OIDC token from Github's provider and returns
// the token.
// FIXME: Don't return an OIDCToken. Either return a rawToken as a string, or verify on the spot.
func (c *OIDCClient) Token(ctx context.Context) (*OIDCToken, error) {
	// Request the token.
	req, err := http.NewRequest("GET", c.requestURL.String(), nil)
	if err != nil {
		return nil, errors.Errorf(&errRequestError{}, "creating request: %w", err)
	}
	req.Header.Add("Authorization", "bearer "+os.Getenv(requestTokenEnvKey))
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
	if resp.StatusCode >= 400 {
		return nil, errors.Errorf(&errRequestError{}, "response: %s: %s", resp.Status, string(b))
	}

	// Extract the raw token from JSON payload.
	var payload struct {
		Value string `json:"value"`
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(&payload); err != nil {
		return nil, errors.Errorf(&errToken{}, "parsing JSON: %w", err)
	}

	parts := strings.Split(payload.Value, ".")
	if len(parts) != 3 {
		return nil, errors.Errorf(&errToken{}, "invalid token, expected 3 parts got %d", len(parts))
	}

	// Base64-decode the content.
	token, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.Errorf(&errToken{}, "invalid token payload: %w", err)
	}

	var t OIDCToken
	if err := json.Unmarshal(token, &t); err != nil {
		return nil, errors.Errorf(&errToken{}, "invalid token payload: %w", err)
	}
	t.rawToken = payload.Value

	return &t, nil
}

// Verify verifies the token contents and signature.
func (c *OIDCClient) Verify(ctx context.Context, t *OIDCToken) error {
	// Verify the token.

	// FIXME: create a verifier using NewVerifier.
	// FIXME: allow setting a dummy KeySet to the verifier for testing.
	provider, err := oidc.NewProvider(ctx, c.actionsProviderURL)
	if err != nil {
		return errors.Errorf(&errVerify{}, "retrieving provider info: %w", err)
	}

	verifier := provider.Verifier(c.conf)
	if _, err = verifier.Verify(ctx, t.rawToken); err != nil {
		return errors.Errorf(&errVerify{}, "could not verify token: %w", err)
	}

	return nil
}
