package slsa

import (
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/slsa-framework/slsa-github-generator/github"
)

// GithubHostedActionsBuilderID is the builder ID for Github hosted actions.
const GithubHostedActionsBuilderID = "https://github.com/Attestations/GitHubHostedActions@v1"

// WorkflowRun contains information about the build run including the builder,
// build invocation, and parameters.
type WorkflowRun struct {
	// Subjects is information about the generated artifacts.
	Subjects []intoto.Subject

	// BuildType indicates the type of build that was done. More importantly it
	// also specifies the format of the BuildConfig.
	BuildType string

	// BuildConfig is metadata about the build.
	BuildConfig interface{}

	// GithubContext is the context for the workflow run.
	GithubContext github.WorkflowContext
}

var (
	parametersVersion = 1
)

// WorkflowParametersV1 contains parameters given to the workflow invocation in v1 format.
type WorkflowParametersV1 struct {
	// Version is the version of the
	Version int `json:"version"`

	// EventName is the name of the event that initiated the workflow run.
	EventName string `json:"event_name,omitempty"`

	// EventPayload is the full event payload.
	// See: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	EventPayload interface{} `json:"event_payload,omitempty"`

	// RefType is type of ref that triggered the workflow run.
	RefType string `json:"ref_type,omitempty"`

	// Ref is the ref that triggered the workflow run.
	Ref string `json:"ref,omitempty"`

	// BaseRef is the base ref or base branch of the pull request in a workflow
	// run.
	BaseRef string `json:"base_ref,omitempty"`

	// HeadRef is ref or source branch of the pull request in a workflow run.
	HeadRef string `json:"head_ref,omitempty"`

	// Actor is the username of the user that initiated the workflow run.
	Actor string `json:"actor,omitempty"`

	// SHA1 is the commit SHA that triggered the workflow run.
	SHA1 string `json:"sha1,omitempty"`
}

// HostedActionsProvenance generates an in-toto provenance statement in the SLSA
// v0.2 format for a workflow run on a Github actions hosted runner.
func HostedActionsProvenance(w WorkflowRun) (intoto.ProvenanceStatement, error) {
	return intoto.ProvenanceStatement{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa.PredicateSLSAProvenance,
			Subject:       w.Subjects,
		},
		Predicate: slsa.ProvenancePredicate{
			BuildType: w.BuildType,
			Builder: slsa.ProvenanceBuilder{
				ID: GithubHostedActionsBuilderID,
			},
			Invocation: slsa.ProvenanceInvocation{
				ConfigSource: slsa.ConfigSource{
					EntryPoint: w.GithubContext.Workflow,
					URI:        w.GithubContext.RepositoryURI(),
					Digest: slsa.DigestSet{
						"sha1": w.GithubContext.SHA,
					},
				},
				// Non user-controllable environment vars needed to reproduce the build.
				Environment: map[string]interface{}{
					// NOTE: Hosted runners are always amd64.
					"arch":               "amd64",
					"github_event_name":  w.GithubContext.EventName,
					"github_run_number":  w.GithubContext.RunNumber,
					"github_run_id":      w.GithubContext.RunID,
					"github_run_attempt": w.GithubContext.RunAttempt,
				},
				// Parameters coming from the trigger event.
				Parameters: WorkflowParametersV1{
					Version:      parametersVersion,
					EventName:    w.GithubContext.EventName,
					Ref:          w.GithubContext.Ref,
					BaseRef:      w.GithubContext.BaseRef,
					HeadRef:      w.GithubContext.HeadRef,
					RefType:      w.GithubContext.RefType,
					Actor:        w.GithubContext.Actor,
					SHA1:         w.GithubContext.SHA,
					EventPayload: w.GithubContext.Event,
				},
			},
			BuildConfig: w.BuildConfig,
			Materials: []slsa.ProvenanceMaterial{
				{
					URI: w.GithubContext.RepositoryURI(),
					Digest: slsa.DigestSet{
						"sha1": w.GithubContext.SHA,
					},
				},
			},
		},
	}, nil
}
