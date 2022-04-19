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
	"fmt"
	"regexp"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/slsa-framework/slsa-github-generator/github"
)

const (
	// GithubHostedActionsBuilderID is the builder ID for Github hosted actions.
	GithubHostedActionsBuilderID = "https://github.com/Attestations/GitHubHostedActions@v1"
)

var (
	githubComReplace = regexp.MustCompile(`^(https?://)?github\.com/?`)
)

// HostedActionsProvenance generates an in-toto provenance statement in the SLSA
// v0.2 format for a workflow run on a Github actions hosted runner.
func HostedActionsProvenance(ctx context.Context, w WorkflowRun, c *github.OIDCClient) (*intoto.ProvenanceStatement, error) {
	// NOTE: Use buildType as the audience as that closely matches the intended
	// recipient of the OIDC token.
	// NOTE: GitHub doesn't allow github.com in the audience so remove it.
	audience := githubComReplace.ReplaceAllString(w.BuildType, "")

	t, err := c.Token(ctx, []string{audience})
	if err != nil {
		return nil, err
	}

	builderID := GithubHostedActionsBuilderID
	if t.JobWorkflowRef != "" {
		builderID = fmt.Sprintf("https://github.com/%s", t.JobWorkflowRef)
	}

	buildInvocationID := w.GithubContext.RunID
	if w.GithubContext.RunAttempt != "" {
		// NOTE: RunID does not get updated on re-runs so we need to include RunAttempt.
		buildInvocationID = fmt.Sprintf("%s-%s", w.GithubContext.RunID, w.GithubContext.RunAttempt)
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
				ID: builderID,
			},
			Invocation:  w.Invocation,
			BuildConfig: w.BuildConfig,
			Materials:   w.Materials,
			// TODO(https://github.com/slsa-framework/slsa-github-generator/issues/8): support more metadata fields.
			Metadata: &slsa.ProvenanceMetadata{
				BuildInvocationID: buildInvocationID,
				Completeness:      w.Completeness,
			},
		},
	}, nil
}
