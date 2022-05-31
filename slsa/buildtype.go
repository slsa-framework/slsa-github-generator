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
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/slsa-framework/slsa-github-generator/github"
)

var errorInvalidOIDCToken = errors.New("invalid OIDC token")

// BuildType implements generation of buildType specific elements of SLSA
// provenance. Each BuildType instance represents a specific build.
type BuildType interface {
	// URI returns the build type's URI.
	URI() string

	// Subject returns a set of artifacts created by the build.
	Subject(context.Context) ([]intoto.Subject, error)

	// BuildConfig returns the buildConfig for this build type.
	BuildConfig(context.Context) (interface{}, error)

	// Invocation returns an invocation for this build type.
	Invocation(context.Context) (slsa.ProvenanceInvocation, error)

	// Materials returns materials as defined by this build type.
	Materials(context.Context) ([]slsa.ProvenanceMaterial, error)

	// Metadata returns a metadata about the build.
	Metadata(context.Context) (*slsa.ProvenanceMetadata, error)
}

// GithubActionsBuild is a basic build type for builders running in Github Actions.
type GithubActionsBuild struct {
	subject []intoto.Subject
	Context github.WorkflowContext
	Clients ClientProvider
}

// WorkflowParameters contains parameters given to the workflow invocation.
type WorkflowParameters struct {
	// EventInputs is the inputs for the event that triggered the workflow.
	EventInputs interface{} `json:"event_inputs,omitempty"`
}

// NewGithubActionsBuild returns a new GithubActionsBuild that uses the
// github context to generate information.
func NewGithubActionsBuild(s []intoto.Subject, c github.WorkflowContext) *GithubActionsBuild {
	return &GithubActionsBuild{
		subject: s,
		Context: c,
		Clients: &DefaultClientProvider{},
	}
}

// Subject implements BuildType.Subject.
func (b *GithubActionsBuild) Subject(context.Context) ([]intoto.Subject, error) {
	return b.subject, nil
}

// BuildConfig implements BuildType.BuildConfig.
func (b *GithubActionsBuild) BuildConfig(context.Context) (interface{}, error) {
	// The default build config is nil.
	return nil, nil
}

func addEnvKeyString(m map[string]interface{}, k string, v string) {
	if v != "" {
		m[k] = v
	}
}

// getEntryPoint retrieves the path to the user workflow that initiated the
// workflow run. The `github` context contains the path in `workflow` but it
// will be the name of the workflow if it's set. The name will not uniquely
// identify the workflow so we need to retrieve the path via the Github API to
// get it reliably.
func (b *GithubActionsBuild) getEntryPoint(ctx context.Context) (string, error) {
	ghClient, err := b.Clients.GithubClient(ctx)
	if err != nil {
		return "", fmt.Errorf("github client: %w", err)
	}
	if ghClient == nil {
		return "", nil
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

// Invocation implements BuildType.Invocation. An invocation is returned that
// describes the workflow run.
// TODO: Document the basic invocation format.
func (b *GithubActionsBuild) Invocation(ctx context.Context) (slsa.ProvenanceInvocation, error) {
	i := slsa.ProvenanceInvocation{}

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
		return i, fmt.Errorf("oidc client: %w", err)
	}

	fmt.Println("oidcClient", oidcClient)
	if oidcClient != nil {
		t, err := oidcClient.Token(ctx, []string{b.Context.Repository})
		if err != nil {
			return i, err
		}

		if t.RepositoryID == "" {
			return i, fmt.Errorf("%w: repository ID is empty", errorInvalidOIDCToken)
		}
		if t.RepositoryOwnerID == "" {
			return i, fmt.Errorf("%w: repository owner ID is empty", errorInvalidOIDCToken)
		}
		if t.ActorID == "" {
			return i, fmt.Errorf("%w: actor ID is empty", errorInvalidOIDCToken)
		}

		// github_repository_id is the unique ID of the repository.
		addEnvKeyString(env, "github_repository_id", t.RepositoryID)

		// github_actor_id is the unique ID of the repository.
		addEnvKeyString(env, "github_actor_id", t.ActorID)

		// github_repository_owner_id is the unique ID of the owner
		// of the repository.
		addEnvKeyString(env, "github_repository_owner_id", t.RepositoryOwnerID)
	}

	if len(env) > 0 {
		i.Environment = env
	}

	// ConfigSource
	entryPoint, err := b.getEntryPoint(ctx)
	if err != nil {
		return i, fmt.Errorf("getting entrypoint: %w", err)
	}

	i.ConfigSource.EntryPoint = entryPoint
	i.ConfigSource.URI = b.Context.RepositoryURI()
	if b.Context.SHA != "" {
		i.ConfigSource.Digest = slsa.DigestSet{
			"sha1": b.Context.SHA,
		}
	}

	if b.Context.Event != nil {
		// Parameters coming from the trigger event.
		i.Parameters = WorkflowParameters{
			EventInputs: b.Context.Event["inputs"],
		}
	}

	return i, nil
}

// Materials implements BuildType.Materials. It returns a list of materials
// that includes the repository that triggered the Github Actions workflow.
func (b *GithubActionsBuild) Materials(ctx context.Context) ([]slsa.ProvenanceMaterial, error) {
	var material []slsa.ProvenanceMaterial
	if b.Context.RepositoryURI() != "" {
		material = append(material, slsa.ProvenanceMaterial{
			URI: b.Context.RepositoryURI(),
			Digest: slsa.DigestSet{
				"sha1": b.Context.SHA,
			},
		})
	}
	return material, nil
}

// Metadata implements BuildType.Metadata. It specifies that parameters
// are complete.
func (b *GithubActionsBuild) Metadata(context.Context) (*slsa.ProvenanceMetadata, error) {
	metadata := slsa.ProvenanceMetadata{}

	metadata.BuildInvocationID = b.Context.RunID
	if b.Context.RunAttempt != "" {
		// NOTE: RunID does not get updated on re-runs so we need to include RunAttempt.
		metadata.BuildInvocationID = fmt.Sprintf("%s-%s", b.Context.RunID, b.Context.RunAttempt)
	}

	if b.Context.Event != nil {
		// Parameters come from the trigger event.
		// If we have the event then mark parameters as complete.
		metadata.Completeness.Parameters = true
	}

	return &metadata, nil
}

// WithClients overrides the build type's default client provider. This is
// useful for tests where APIs are not available.
func (b *GithubActionsBuild) WithClients(p ClientProvider) *GithubActionsBuild {
	fmt.Println("setting clients")
	b.Clients = p
	return b
}
