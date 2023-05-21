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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	slsa1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"

	"github.com/slsa-framework/slsa-github-generator/github"
)

// BuildType implements generation of buildType specific elements of SLSA
// provenance. Each BuildType instance represents a specific build.
type BuildType interface {
	// URI returns the build type's URI.
	URI() string

	// Subject returns a set of artifacts created by the build.
	Subject(context.Context) ([]intoto.Subject, error)

	// BuildDefinition returns the input to the build.
	BuildDefinition(context.Context) (slsa1.ProvenanceBuildDefinition, error)

	// RunDetails returns the Details of the build execution.
	RunDetails(context.Context) (slsa1.ProvenanceRunDetails, error)
}

// GithubActionsBuild is a basic build type for builders running in GitHub Actions.
type GithubActionsBuild struct {
	Context github.WorkflowContext
	Clients ClientProvider
	subject []intoto.Subject
}

// WorkflowExternalParameters contains parameters given to the workflow invocation.
type WorkflowExternalParameters struct {
	// EventInputs is the inputs for the event that triggered the workflow.
	EventInputs interface{} `json:"event_inputs,omitempty"`
	EntryPoint  string      `json:"entry_point,omitempty"`
	Source      string      `json:"source,omitempty"`
	Config      string      `json:"config,omitempty"`
}

// NewGithubActionsBuild returns a new GithubActionsBuild that uses the
// GitHub context to generate information.
func NewGithubActionsBuild(s []intoto.Subject, c *github.WorkflowContext) *GithubActionsBuild {
	return &GithubActionsBuild{
		subject: s,
		Context: *c,
		Clients: &DefaultClientProvider{},
	}
}

// BuildDefinition implements BuildType.BuildDefinition.
func (b *GithubActionsBuild) BuildDefinition(ctx context.Context) (slsa1.ProvenanceBuildDefinition, error) {

	resDependencies, err := b.resolvedDependencies(ctx)
	if err != nil {
		return slsa1.ProvenanceBuildDefinition{}, err
	}

	exParameters, err := b.externalParameters(ctx)
	if err != nil {
		return slsa1.ProvenanceBuildDefinition{}, err
	}

	inParameters, err := b.internalParameters(ctx)
	if err != nil {
		return slsa1.ProvenanceBuildDefinition{}, err
	}

	buildDefinition := slsa1.ProvenanceBuildDefinition{
		ExternalParameters:   exParameters,
		InternalParameters:   inParameters,
		ResolvedDependencies: resDependencies,
	}
	return buildDefinition, nil
}

// RunDetails implements BuildType.RunDetails.
func (b *GithubActionsBuild) RunDetails(ctx context.Context) (slsa1.ProvenanceRunDetails, error) {

	metadata, err := b.Metadata(ctx)
	if err != nil {
		return slsa1.ProvenanceRunDetails{}, err
	}

	runDetails := slsa1.ProvenanceRunDetails{
		BuildMetadata: metadata,
	}

	return runDetails, nil
}

// Subject implements BuildType.Subject.
func (b *GithubActionsBuild) Subject(context.Context) ([]intoto.Subject, error) {
	return b.subject, nil
}

func addEnvKeyString(m map[string]interface{}, k, v string) {
	// Always record the value, even if it's empty. Let
	// the consumer/verifier decide how to interpret their meaning.
	m[k] = v
}

// getEntryPoint retrieves the path to the user workflow that initiated the
// workflow run. The `github` context contains the path in `workflow` but it
// will be the name of the workflow if it's set. The name will not uniquely
// identify the workflow, so we need to retrieve the path via the GitHub API to
// get it reliably.
func (b *GithubActionsBuild) getEntryPoint(ctx context.Context) (string, error) {
	ghClient, err := b.Clients.GithubClient(ctx)
	if err != nil {
		return "", fmt.Errorf("github client: %w", err)
	}
	if ghClient == nil {
		// If no client is provided, return the name of the workflow.
		return b.Context.Workflow, nil
	}

	runID, err := strconv.ParseInt(b.Context.RunID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parsing run ID %q: %w", b.Context.RunID, err)
	}

	repo := strings.SplitN(b.Context.Repository, "/", 2)
	if len(repo) < 2 {
		return "", fmt.Errorf("unexpected repository: %q", b.Context.Repository)
	}
	owner := repo[0]
	repoName := repo[1]

	wr, _, err := ghClient.Actions.GetWorkflowRunByID(ctx, owner, repoName, runID)
	if err != nil {
		return "", fmt.Errorf("getting workflow run: %w", err)
	}

	wf, _, err := ghClient.Actions.GetWorkflowByID(ctx, owner, repoName, wr.GetWorkflowID())
	if err != nil {
		return "", fmt.Errorf("getting workflow: %w", err)
	}
	if wf.Path == nil {
		return "", errors.New("workflow path not found")
	}

	return *wf.Path, nil
}

func (b *GithubActionsBuild) externalParameters(ctx context.Context) (interface{}, error) {
	// ConfigSource
	var externalParameters WorkflowExternalParameters
	entryPoint, err := b.getEntryPoint(ctx)
	if err != nil {
		return externalParameters, fmt.Errorf("getting entrypoint: %w", err)
	}

	externalParameters.EntryPoint = entryPoint

	externalParameters.Source = b.Context.RepositoryURI()

	if b.Context.Event != nil {
		// Parameters coming from the trigger event.
		externalParameters.EventInputs = b.Context.Event["inputs"]
	}

	return externalParameters, nil
}

func (b *GithubActionsBuild) internalParameters(ctx context.Context) (interface{}, error) {

	// Builder-controlled environment vars needed
	// to reproduce the build.
	env := map[string]interface{}{}

	// TODO(github.com/slsa-framework/slsa-github-generator/issues/5): set "arch" in environment.
	addEnvKeyString(env, "github_run_number", b.Context.RunNumber)
	addEnvKeyString(env, "github_run_id", b.Context.RunID)
	addEnvKeyString(env, "github_run_attempt", b.Context.RunAttempt)

	// github_event_name is the name of the event that initiated the
	// workflow run.
	addEnvKeyString(env, "github_event_name", b.Context.EventName)

	// github_event_payload is the full event payload.
	if b.Context.Event != nil {
		env["github_event_payload"] = b.Context.Event
	}

	// github_ref_type is type of ref that triggered the
	// workflow run.
	addEnvKeyString(env, "github_ref_type", b.Context.RefType)

	// github_ref is the ref that triggered the workflow run.
	addEnvKeyString(env, "github_ref", b.Context.Ref)

	// github_base_ref is the base ref or base branch of the
	// pull request in a workflow run.
	addEnvKeyString(env, "github_base_ref", b.Context.BaseRef)

	// github_head_ref is ref or source branch of the pull
	// request in a workflow run.
	addEnvKeyString(env, "github_head_ref", b.Context.HeadRef)

	// github_actor is the username of the user that initiated
	// the workflow run.
	addEnvKeyString(env, "github_actor", b.Context.Actor)

	// github_sha1 is the commit SHA that triggered the
	// workflow run.
	addEnvKeyString(env, "github_sha1", b.Context.SHA)

	// github_repository_owner is the owner of the repository.
	addEnvKeyString(env, "github_repository_owner", b.Context.RepositoryOwner)

	oidcClient, err := b.Clients.OIDCClient()
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("oidc client: %w", err)
	}

	if oidcClient != nil {
		t, err := oidcClient.Token(ctx, []string{b.Context.Repository})
		if err != nil {
			return map[string]interface{}{}, err
		}

		// github_repository_id is the unique ID of the repository.
		addEnvKeyString(env, "github_repository_id", t.RepositoryID)

		// github_actor_id is the unique ID of the repository.
		addEnvKeyString(env, "github_actor_id", t.ActorID)

		// github_repository_owner_id is the unique ID of the owner
		// of the repository.
		addEnvKeyString(env, "github_repository_owner_id", t.RepositoryOwnerID)
	}

	// Set the env.
	return env, nil
}

// ResolvedDependencies implements BuildType.ResolvedDependencies. It returns a list of ResourceDescriptor
// that includes the repository that triggered the GitHub Actions workflow.
func (b *GithubActionsBuild) resolvedDependencies(context.Context) ([]slsa1.ResourceDescriptor, error) {
	var resources []slsa1.ResourceDescriptor
	var targetResource slsa1.ResourceDescriptor
	if b.Context.RepositoryURI() != "" {
		targetResource.URI = b.Context.RepositoryURI()
	}

	if b.Context.SHA != "" {
		targetResource.Digest = slsacommon.DigestSet{
			"sha1": b.Context.SHA,
		}
	}

	if b.Context.RepositoryURI() != "" || b.Context.SHA != "" {
		resources = append(resources, targetResource)
	}

	return resources, nil
}

// Metadata implements BuildType.Metadata. It specifies that parameters
// are complete.
func (b *GithubActionsBuild) Metadata(context.Context) (slsa1.BuildMetadata, error) {
	metadata := slsa1.BuildMetadata{}

	metadata.InvocationID = b.Context.RunID
	if b.Context.RunAttempt != "" {
		// NOTE: RunID does not get updated on re-runs so we need to include RunAttempt.
		metadata.InvocationID = fmt.Sprintf("%s-%s", b.Context.RunID, b.Context.RunAttempt)
	}

	// if b.Context.Event != nil {
	// 	// Parameters come from the trigger event.
	// 	// If we have the event then mark parameters as complete.
	// 	metadata.Completeness.Parameters = true
	// }

	return metadata, nil
}

// WithClients overrides the build type's default client provider. This is
// useful for tests where APIs are not available.
func (b *GithubActionsBuild) WithClients(p ClientProvider) *GithubActionsBuild {
	b.Clients = p
	return b
}
