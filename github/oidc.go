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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
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
func RequestOIDCToken(audience string) (*OIDCToken, error) {
	urlKey := os.Getenv(requestURLEnvKey)
	if urlKey == "" {
		return nil, errURLEnvKeyEmpty
	}

	url := urlKey + "&audience=" + audience
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Authorization", "bearer "+os.Getenv(requestTokenEnvKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	var payload struct {
		Value string `json:"value"`
	}

	// Extract the value from JSON payload.
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		return nil, errResponseJSON
	}

	// This is a JWT token with 3 parts.
	parts := strings.Split(payload.Value, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	content := parts[1]

	// Base64-decode the content.
	token, err := base64.RawURLEncoding.DecodeString(content)
	if err != nil {
		return nil, errInvalidTokenB64
	}

	var oidc OIDCToken
	if err := json.Unmarshal(token, &oidc); err != nil {
		return nil, errInvalidTokenJSON
	}

	return &oidc, nil
}
