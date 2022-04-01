package github

import (
	"encoding/base64"
	"encoding/json"
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

// RequestOIDCToken requests an OIDC token from Github's provider and returns
// the token.
func RequestOIDCToken(audience string) (*OIDCToken, error) {
	urlKey := os.Getenv(requestURLEnvKey)
	if urlKey == "" {
		return nil, fmt.Errorf("requestURLEnvKey is empty")
	}

	url := urlKey + "&audience=" + audience
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "bearer "+os.Getenv(requestTokenEnvKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Value string `json:"value"`
	}

	// Extract the value from JSON payload.
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}

	// This is a JWT token with 3 parts.
	parts := strings.Split(payload.Value, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt token: found %d parts", len(parts))
	}

	content := parts[1]

	// Base64-decode the content.
	token, err := base64.RawURLEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("base64.RawURLEncoding.DecodeString: %w", err)
	}

	var oidc OIDCToken
	if err := json.Unmarshal(token, &oidc); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &oidc, nil
}
