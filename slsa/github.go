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

package slsa

import (
	"fmt"
	"path"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/slsa-framework/slsa-github-generator/github"
)

const (
	provenanceOnlyBuildType = "https://github.com/slsa-framework/slsa-github-generator@v1"
)

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

	// oidcClient overrides the GitHub OIDC client.
	oidcClient *github.OIDCClient
}

// WorkflowParameters contains parameters given to the workflow invocation.
type WorkflowParameters struct {
	// EventInputs is the inputs for the event that triggered the workflow.
	EventInputs interface{} `json:"event_inputs,omitempty"`
}

// NewWorkflowRun returns a generic WorkflowRun based on the
// github context without special knowledge of the build.
func NewWorkflowRun(s []intoto.Subject, c github.WorkflowContext) WorkflowRun {
	// Create the entrypoint from the repository URI @ workflow path in the
	// event. `workflow` is not used from the github context because it includes
	// the workflow name rather than the path.
	// NOTE: ServerURL does not have a trailing slash and Repository does not
	// have a leading slash.
	entryPoint := fmt.Sprintf("%s/%s",
		c.ServerURL,
		path.Join(
			c.Repository,
			fmt.Sprintf("%s", c.Event["workflow"]),
		),
	)

	return WorkflowRun{
		Subjects:  s,
		BuildType: provenanceOnlyBuildType,
		Invocation: slsa.ProvenanceInvocation{
			ConfigSource: slsa.ConfigSource{
				EntryPoint: entryPoint,
				URI:        c.RepositoryURI(),
				Digest: slsa.DigestSet{
					"sha1": c.SHA,
				},
			},
			// Builder-controlled environment vars needed
			// to reproduce the build.
			Environment: map[string]interface{}{
				// TODO(https://github.com/slsa-framework/slsa-github-generator/issues/5): set "arch"
				"github_run_number":  c.RunNumber,
				"github_run_id":      c.RunID,
				"github_run_attempt": c.RunAttempt,
				// github_event_name is the name of the event that initiated the
				// workflow run.
				"github_event_name": c.EventName,
				// github_event_payload is the full event payload.
				"github_event_payload": c.Event,
				// github_ref_type is type of ref that triggered the
				// workflow run.
				"github_ref_type": c.RefType,
				// github_ref is the ref that triggered the workflow run.
				"github_ref": c.Ref,
				// github_base_ref is the base ref or base branch of the
				// pull request in a workflow run.
				"github_base_ref": c.BaseRef,
				// github_head_ref is ref or source branch of the pull
				// request in a workflow run.
				"github_head_ref": c.HeadRef,
				// github_actor is the username of the user that initiated
				// the workflow run.
				"github_actor": c.Actor,
				// github_sha1 is the commit SHA that triggered the
				// workflow run.
				"github_sha1": c.SHA,
			},
			// Parameters coming from the trigger event.
			Parameters: WorkflowParameters{
				EventInputs: c.Event["inputs"],
			},
		},
		Materials: []slsa.ProvenanceMaterial{
			{
				URI: c.RepositoryURI(),
				Digest: slsa.DigestSet{
					"sha1": c.SHA,
				},
			},
		},
		Completeness: slsa.ProvenanceComplete{
			Parameters: true,
		},
		GithubContext: c,
	}
}
