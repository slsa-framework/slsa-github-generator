package slsa

import (
	"testing"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func TestNewWorkflowRun(t *testing.T) {
	t.Run("entrypoint", func(t *testing.T) {
		r := NewWorkflowRun(nil, github.WorkflowContext{
			ServerURL:  "https://github.com",
			Repository: "slsa-framework/slsa-github-generator",
			Event: map[string]interface{}{
				"workflow": ".github/workflows/workflow.yml",
			},
		})

		if want, got := "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/workflow.yml", r.Invocation.ConfigSource.EntryPoint; want != got {
			t.Errorf("unexpected entrypoint, want: %q, got: %q", want, got)
		}
	})
}
