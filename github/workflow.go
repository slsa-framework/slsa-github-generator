package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// WorkflowContext is the 'github' context given to workflows that contains
// information about the Github Actions workflow run.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#github-context.
type WorkflowContext struct {
	Repository string      `json:"repository"`
	ActionPath string      `json:"action_path"`
	Workflow   string      `json:"workflow"`
	EventName  string      `json:"event_name"`
	Event      interface{} `json:"event"`
	SHA        string      `json:"sha"`
	RefType    string      `json:"ref_type"`
	Ref        string      `json:"ref"`
	BaseRef    string      `json:"base_ref"`
	HeadRef    string      `json:"head_ref"`
	Actor      string      `json:"actor"`
	RunNumber  string      `json:"run_number"`
	ServerURL  string      `json:"server_url"`
	RunID      string      `json:"run_id"`
	RunAttempt string      `json:"run_attempt"`
	// TODO: try removing this token:
	// `omitting Token from the struct causes an unexpected end of line from encoding/json`
	// Token string `json:"token,omitempty"`
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
		"git+%s/%s%s.git",
		c.ServerURL,
		c.Repository,
		ref,
	)
}

// GetWorkflowContext returns the current Github Actions 'github' context.
func GetWorkflowContext() (WorkflowContext, error) {
	w := WorkflowContext{}
	ghContext, ok := os.LookupEnv("GITHUB_CONTEXT")
	if !ok {
		return w, errors.New("GITHUB_CONTEXT environment variable not set")
	}

	err := json.Unmarshal([]byte(ghContext), &w)
	return w, err
}
