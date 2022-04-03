package slsa

import (
	"fmt"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/slsa-framework/slsa-github-generator/github"
)

// GithubHostedActionsBuilderID is the builder ID for Github hosted actions.
const GithubHostedActionsBuilderID = "https://github.com/Attestations/GitHubHostedActions@v1"

// WorkflowRun contains information about the build run including the builder,
// build invocation, materials, and environment.
type WorkflowRun struct {
	// Subjects is information about the generated artifacts.
	Subjects []intoto.Subject

	// BuildType indicates the type of build that was done. More importantly it
	// also specifies the format of the BuildConfig.
	BuildType string

	// BuildConfig is metadata about the build.
	BuildConfig interface{}

	// Invocation is the provenance invocation.
	Invocation slsa.ProvenanceInvocation

	// Materials is the materials used in the build run.
	Materials []slsa.ProvenanceMaterial

	// Completeness holds info on the completeness of
	// provenance data.
	Completeness slsa.ProvenanceComplete

	// GithubContext is the context for the workflow run.
	GithubContext github.WorkflowContext
}

// WorkflowParameters contains parameters given to the workflow invocation.
type WorkflowParameters struct {
	// EventInputs is the inputs for the event that triggered the workflow.
	EventInputs interface{} `json:"event_inputs,omitempty"`
}

var (
	audience = "slsa-framework"
)

// HostedActionsProvenance generates an in-toto provenance statement in the SLSA
// v0.2 format for a workflow run on a Github actions hosted runner.
func HostedActionsProvenance(w WorkflowRun) (*intoto.ProvenanceStatement, error) {
	t, err := github.RequestOIDCToken(audience)
	if err != nil {
		return nil, err
	}

	return &intoto.ProvenanceStatement{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa.PredicateSLSAProvenance,
			Subject:       w.Subjects,
		},
		Predicate: slsa.ProvenancePredicate{
			BuildType: w.BuildType,
			Builder: slsa.ProvenanceBuilder{
				ID: fmt.Sprintf("https://github.com/%s", t.JobWorkflowRef),
			},
			Invocation:  w.Invocation,
			BuildConfig: w.BuildConfig,
			Materials:   w.Materials,
			Metadata: &slsa.ProvenanceMetadata{
				Completeness: w.Completeness,
			},
		},
	}, nil
}
