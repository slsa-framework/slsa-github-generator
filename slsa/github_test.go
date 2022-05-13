package slsa

import (
	"testing"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func TestNewWorkflowRun(t *testing.T) {
	t.Run("entrypoint", func(t *testing.T) {
		r := NewWorkflowRun(nil, github.WorkflowContext{
			// TODO(github.com/slsa-framework/slsa-github-generator/issues/11): Use path to workflow file.
			Workflow: "Workflow Name",
		})

		if want, got := "Workflow Name", r.Invocation.ConfigSource.EntryPoint; want != got {
			t.Errorf("unexpected entrypoint, want: %q, got: %q", want, got)
		}
	})
}
