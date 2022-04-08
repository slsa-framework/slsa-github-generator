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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
)

const (
	actionsProviderURL = "https://token.actions.githubusercontent.com"
	requestTokenEnvKey = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	requestURLEnvKey   = "ACTIONS_ID_TOKEN_REQUEST_URL"
)

// OIDCToken represents the contents of a Github OIDC JWT token.
type OIDCToken struct {
	// JobWorkflowRef is a reference to the current job workflow.
	JobWorkflowRef string `json:"job_workflow_ref"`
}

var (
	errURLEnvKeyEmpty   = fmt.Errorf("%q env var is empty", requestURLEnvKey)
	errResponseJSON     = errors.New("invalid response JSON")
	errInvalidToken     = errors.New("invalid JWT token")
	errInvalidTokenB64  = errors.New("invalid JWT token base64")
	errInvalidTokenJSON = errors.New("invalid JWT token JSON")
)

// RequestOIDCToken requests an OIDC token from Github's provider and returns
// the token.
func RequestOIDCToken(ctx context.Context, audience string) (*OIDCToken, error) {
	requestURL := os.Getenv(requestURLEnvKey)
	if requestURL == "" {
		return nil, errURLEnvKeyEmpty
	}

	req, err := http.NewRequest("GET", requestURL+"&audience="+audience, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Authorization", "bearer "+os.Getenv(requestTokenEnvKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	rawIDToken, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Verify the token.
	provider, err := oidc.NewProvider(ctx, actionsProviderURL)
	if err != nil {
		return nil, fmt.Errorf("retrieving provider info: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: audience})
	idToken, err := verifier.Verify(ctx, string(rawIDToken))
	if err != nil {
		return nil, fmt.Errorf("could not verify token: %w", err)
		// return nil, errInvalidTokenJSON
	}

	var token OIDCToken
	if err := idToken.Claims(&token); err != nil {
		return nil, fmt.Errorf("invalid claims: %w", err)
		// return nil, errInvalidTokenJSON
	}

	return &token, nil
}
