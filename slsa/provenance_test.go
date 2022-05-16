package slsa

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/slsa-framework/slsa-github-generator/github"
)

var testBuildType = "http://example.com/v1"
var testBuildConfig = "test build config"

type TestBuild struct {
	GithubActionsBuild
}

func (*TestBuild) URI() string {
	return testBuildType
}

func (*TestBuild) BuildConfig(context.Context) (interface{}, error) {
	return testBuildConfig, nil
}

func TestHostedActionsProvenance(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		b        BuildType
		token    *github.OIDCToken
		expected *intoto.ProvenanceStatement
	}{
		{
			name: "empty",
			b:    NewGithubActionsBuild(nil, github.WorkflowContext{}).WithClients(&NilClientProvider{}),
			token: &github.OIDCToken{
				Audience: []string{""},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa.PredicateSLSAProvenance,
				},
				Predicate: slsa.ProvenancePredicate{
					BuildType: provenanceOnlyBuildType,
					Builder: slsa.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					Metadata: &slsa.ProvenanceMetadata{},
				},
			},
		},
		{
			name: "empty",
			b: &TestBuild{
				GithubActionsBuild: *NewGithubActionsBuild(nil, github.WorkflowContext{}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{""},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa.PredicateSLSAProvenance,
				},
				Predicate: slsa.ProvenancePredicate{
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Builder: slsa.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					Metadata: &slsa.ProvenanceMetadata{},
				},
			},
		},
		{
			name: "invocation id",
			b: NewGithubActionsBuild(nil, github.WorkflowContext{
				RunID:      "12345",
				RunAttempt: "1",
			}).WithClients(&NilClientProvider{}),
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa.PredicateSLSAProvenance,
				},
				Predicate: slsa.ProvenancePredicate{
					Builder: slsa.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType: provenanceOnlyBuildType,
					Invocation: slsa.ProvenanceInvocation{
						Environment: map[string]interface{}{
							"github_run_id":      "12345",
							"github_run_attempt": "1",
						},
					},
					Metadata: &slsa.ProvenanceMetadata{
						BuildInvocationID: "12345-1",
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
