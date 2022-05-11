package slsa

import (
	"testing"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func TestNewWorkflowRun(t *testing.T) {
	t.Run("entrypoint", func(t *testing.T) {
		workflowPath := ".github/workflows/workflow.yml"

		r := NewWorkflowRun(nil, github.WorkflowContext{
			Event: map[string]interface{}{
				"workflow": workflowPath,
			},
		})

		if want, got := workflowPath, r.Invocation.ConfigSource.EntryPoint; want != got {
			t.Errorf("unexpected entrypoint, want: %q, got: %q", want, got)
		}
	})
}
