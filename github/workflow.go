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
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const (
	githubContextEnvKey = "GITHUB_CONTEXT"
)

// WorkflowContext is the 'github' context given to workflows that contains
// information about the Github Actions workflow run.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#github-context.
type WorkflowContext struct {
	Repository string                 `json:"repository"`
	ActionPath string                 `json:"action_path"`
	Workflow   string                 `json:"workflow"`
	EventName  string                 `json:"event_name"`
	Event      map[string]interface{} `json:"event"`
	SHA        string                 `json:"sha"`
	RefType    string                 `json:"ref_type"`
	Ref        string                 `json:"ref"`
	BaseRef    string                 `json:"base_ref"`
	HeadRef    string                 `json:"head_ref"`
	Actor      string                 `json:"actor"`
	RunNumber  string                 `json:"run_number"`
	ServerURL  string                 `json:"server_url"`
	RunID      string                 `json:"run_id"`
	RunAttempt string                 `json:"run_attempt"`
}

// RepositoryURI returns a full repository URI for the repo that triggered the workflow.
func (c WorkflowContext) RepositoryURI() string {
	if c.ServerURL == "" || c.Repository == "" {
		return ""
	}
	var ref string
	if c.Ref != "" {
		ref = "@" + c.Ref
	}
	return fmt.Sprintf(
		"git+%s/%s%s",
		c.ServerURL,
		c.Repository,
		ref,
	)
}

// GetWorkflowContext returns the current Github Actions 'github' context.
func GetWorkflowContext() (WorkflowContext, error) {
	w := WorkflowContext{}
	ghContext, ok := os.LookupEnv(githubContextEnvKey)
	if !ok {
		return w, errors.New("GITHUB_CONTEXT environment variable not set")
	}

	err := json.Unmarshal([]byte(ghContext), &w)
	return w, err
}

// GetToken gets the Github Actions token.
// See: https://docs.github.com/en/actions/security-guides/automatic-token-authentication
func GetToken() (string, error) {
	var w struct {
		Token string `json:"token,omitempty"`
	}
	ghContext, ok := os.LookupEnv(githubContextEnvKey)
	if !ok {
		return "", errors.New("GITHUB_CONTEXT environment variable not set")
	}

	err := json.Unmarshal([]byte(ghContext), &w)
	return w.Token, err
}
