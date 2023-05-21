package slsa

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	slsa1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"

	"github.com/slsa-framework/slsa-github-generator/github"
)

var (
	testBuildType   = "http://example.com/v1"
	testBuildConfig = "test build config"
)

type TestBuild struct {
	*GithubActionsBuild
}

func (*TestBuild) URI() string {
	return testBuildType
}

func (tB *TestBuild) BuildDefinition(ctx context.Context) (slsa1.ProvenanceBuildDefinition, error) {
	buildDefinition, err := tB.GithubActionsBuild.BuildDefinition(ctx)
	if err != nil {
		return slsa1.ProvenanceBuildDefinition{}, err
	}

	workflowBuildDef := buildDefinition.ExternalParameters.(WorkflowExternalParameters)
	workflowBuildDef.Config = testBuildConfig
	buildDefinition.ExternalParameters = workflowBuildDef

	return buildDefinition, nil
}

func TestHostedActionsProvenance(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		b        BuildType
		token    *github.OIDCToken
		expected *intoto.ProvenanceStatementSLSA1
		name     string
	}{
		{
			name: "empty",
			b: &TestBuild{
				GithubActionsBuild: NewGithubActionsBuild(nil, &github.WorkflowContext{}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{""},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatementSLSA1{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa1.PredicateSLSAProvenance,
				},
				Predicate: slsa1.ProvenancePredicate{
					BuildDefinition: slsa1.ProvenanceBuildDefinition{
						BuildType: testBuildType,
						InternalParameters: map[string]interface{}{
							"github_run_id":           "",
							"github_run_attempt":      "",
							"github_actor":            "",
							"github_base_ref":         "",
							"github_event_name":       "",
							"github_head_ref":         "",
							"github_ref":              "",
							"github_ref_type":         "",
							"github_repository_owner": "",
							"github_run_number":       "",
							"github_sha1":             "",
						},
						ExternalParameters: WorkflowExternalParameters{
							Config: testBuildConfig,
						},
					},
					RunDetails: slsa1.ProvenanceRunDetails{
						Builder: slsa1.Builder{
							ID: GithubHostedActionsBuilderID,
						},
					},
				},
			},
		},
		{
			name: "invocation env",
			b: &TestBuild{
				GithubActionsBuild: NewGithubActionsBuild(nil, &github.WorkflowContext{
					RunID:      "12345",
					RunAttempt: "1",
					EventName:  "pull_request",
					SHA:        "abcde",
					RefType:    "branch",
					Ref:        "some/ref",
					BaseRef:    "some/base_ref",
					HeadRef:    "some/head_ref",
					RunNumber:  "102937",
					Actor:      "user",
				}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},

			expected: &intoto.ProvenanceStatementSLSA1{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa1.PredicateSLSAProvenance,
				},
				Predicate: slsa1.ProvenancePredicate{
					BuildDefinition: slsa1.ProvenanceBuildDefinition{
						BuildType: testBuildType,
						InternalParameters: map[string]interface{}{
							"github_run_id":           "12345",
							"github_run_attempt":      "1",
							"github_actor":            "user",
							"github_base_ref":         "some/base_ref",
							"github_event_name":       "pull_request",
							"github_head_ref":         "some/head_ref",
							"github_ref":              "some/ref",
							"github_ref_type":         "branch",
							"github_repository_owner": "",
							"github_run_number":       "102937",
							"github_sha1":             "abcde",
						},
						ExternalParameters: WorkflowExternalParameters{
							Config: testBuildConfig,
						},
						ResolvedDependencies: []slsa1.ResourceDescriptor{
							{
								Digest: slsacommon.DigestSet{
									"sha1": "abcde",
								},
							},
						},
					},
					RunDetails: slsa1.ProvenanceRunDetails{
						Builder: slsa1.Builder{
							ID: GithubHostedActionsBuilderID,
						},
						BuildMetadata: slsa1.BuildMetadata{
							InvocationID: "12345-1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewHostedActionsGenerator(tc.b).WithClients(&NilClientProvider{})

			if p, err := g.Generate(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else {
				if want, got := tc.expected, p; !cmp.Equal(want, got) {
					t.Errorf("unexpected result\nwant: %#v\ngot:  %#v\ndiff: %v", want, got, cmp.Diff(want, got))
				}
			}
		})
	}
}
