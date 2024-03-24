// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"gopkg.in/square/go-jose.v2"
)

type jsonToken struct {
	Issuer            string   `json:"iss"`
	JobWorkflowRef    string   `json:"job_workflow_ref"`
	RepositoryID      string   `json:"repository_id"`
	RepositoryOwnerID string   `json:"repository_owner_id"`
	ActorID           string   `json:"actor_id"`
	Audience          []string `json:"aud"`
	Expiry            int64    `json:"exp"`
}

// testKeySet is an oidc.KeySet that can be used in tests.
type testKeySet struct{}

// VerifySignature implements oidc.KeySet.VerifySignature.
func (ks *testKeySet) VerifySignature(_ context.Context, jwt string) ([]byte, error) {
	// NOTE: Doesn't actually verify, just parses out the payload from the token.
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("jwt parts: %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("jwt payload: %w", err)
	}
	return payload, nil
}

// NewTestOIDCServer returns a httptest.Server that can be used as the OIDC
// server, and an OIDClient that will use the test server. The server returns the
// given token when queried. Now is the time used for token expiration
// verification by the client.
func NewTestOIDCServer(t *testing.T, now time.Time, token *OIDCToken) (*httptest.Server, *OIDCClient) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// FIXME: Fix creating a test server that can return tokens that can be verified.
	var issuerURL string
	s, c := newTestOIDCServer(t, now, func(w http.ResponseWriter, _ *http.Request) {
		// Allow the token to override the issuer for verification testing.
		issuer := issuerURL
		if token.Issuer != "" {
			issuer = token.Issuer
		}

		b, err := json.Marshal(jsonToken{
			Issuer:            issuer,
			Audience:          token.Audience,
			Expiry:            token.Expiry.Unix(),
			JobWorkflowRef:    token.JobWorkflowRef,
			RepositoryID:      token.RepositoryID,
			RepositoryOwnerID: token.RepositoryOwnerID,
			ActorID:           token.ActorID,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		object, err := signer.Sign(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		value, err := object.CompactSerialize()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `{"value": "%s"}`, value)
	})
	issuerURL = s.URL

	return s, c
}

func newRawTestOIDCServer(t *testing.T, now time.Time, status int, raw string) (*httptest.Server, *OIDCClient) {
	return newTestOIDCServer(t, now, func(w http.ResponseWriter, _ *http.Request) {
		// Respond with a very basic 3-part JWT token.
		w.WriteHeader(status)
		fmt.Fprintln(w, raw)
	})
}

func newTestOIDCServer(t *testing.T, now time.Time, f http.HandlerFunc) (*httptest.Server, *OIDCClient) {
	var issuerURL string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			f(w, r)
		case "/.well-known/openid-configuration":
			// Return very basic provider info in case it's requested.
			fmt.Fprintf(w, `{"issuer": %q, "token_endpoint": %q}`, issuerURL, issuerURL)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	issuerURL = s.URL

	requestURL, err := url.ParseRequestURI(s.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := OIDCClient{
		requestURL: requestURL,
		verifierFunc: func(_ context.Context) (*oidc.IDTokenVerifier, error) {
			return oidc.NewVerifier(s.URL, &testKeySet{}, &oidc.Config{
				Now:               func() time.Time { return now },
				SkipClientIDCheck: true,
			}), nil
		},
	}

	return s, &c
}
