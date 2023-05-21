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

	slsa1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
)

const (
	// GithubHostedActionsBuilderID is a default builder ID for Github hosted actions.
	GithubHostedActionsBuilderID = "https://github.com/Attestations/GitHubHostedActions@v1"
)

var githubComReplace = regexp.MustCompile(`^(https?://)?github\.com/?`)

// HostedActionsGenerator is a SLSA provenance generator for Github Hosted
// Actions. Provenance is generated based on a "build type" which defines the
// format for many of the fields in the provenance metadata. Builders for
// different ecosystems (languages etc.) can implement a build type from
// scratch or by extending GithubActionsBuild.
type HostedActionsGenerator struct {
	buildType BuildType
	clients   ClientProvider
}

// NewHostedActionsGenerator returns a SLSA provenance generator for the given build type.
func NewHostedActionsGenerator(bt BuildType) *HostedActionsGenerator {
	return &HostedActionsGenerator{
		buildType: bt,
		clients:   &DefaultClientProvider{},
	}
}

// Generate generates an in-toto provenance statement in SLSA v0.2 format.
func (g *HostedActionsGenerator) Generate(ctx context.Context) (*intoto.ProvenanceStatementSLSA1, error) {
	// NOTE: Use buildType as the audience as that closely matches the intended
	// recipient of the OIDC token.
	// NOTE: GitHub doesn't allow github.com in the audience so remove it.
	audience := githubComReplace.ReplaceAllString(g.buildType.URI(), "")

	oidcClient, err := g.clients.OIDCClient()
	if err != nil {
		return nil, err
	}

	// We allow nil OIDC client to support e2e tests on pull requests.
	builderID := GithubHostedActionsBuilderID
	if oidcClient != nil {
		t, err := oidcClient.Token(ctx, []string{audience})
		if err != nil {
			return nil, err
		}

		if t.JobWorkflowRef != "" {
			builderID = fmt.Sprintf("https://github.com/%s", t.JobWorkflowRef)
		}
	}

	subject, err := g.buildType.Subject(ctx)
	if err != nil {
		return nil, err
	}

	buildDefinition, err := g.buildType.BuildDefinition(ctx)
	if err != nil {
		return nil, err
	}
	buildDefinition.BuildType = g.buildType.URI()

	runDetails, err := g.buildType.RunDetails(ctx)
	if err != nil {
		return nil, err
	}
	runDetails.Builder.ID = builderID

	return &intoto.ProvenanceStatementSLSA1{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa1.PredicateSLSAProvenance,
			Subject:       subject,
		},
		Predicate: slsa1.ProvenancePredicate{
			BuildDefinition: buildDefinition,
			RunDetails:      runDetails,
		},
	}, nil
}

// WithClients overrides the default ClientProvider. Useful for tests where
// clients are not available.
func (g *HostedActionsGenerator) WithClients(c ClientProvider) *HostedActionsGenerator {
	g.clients = c
	return g
}
