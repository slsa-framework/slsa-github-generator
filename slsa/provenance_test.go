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

func TestHostedActionsProvenance(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		r        WorkflowRun
		token    *github.OIDCToken
		expected *intoto.ProvenanceStatement
	}{
		{
			name: "empty",
			r:    WorkflowRun{},
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
					Builder: slsa.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					Metadata: &slsa.ProvenanceMetadata{},
				},
			},
		},
		{
			name: "invocation id",
			r: WorkflowRun{
				GithubContext: github.WorkflowContext{
					RunID:      "12345",
					RunAttempt: "1",
				},
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
					Metadata: &slsa.ProvenanceMetadata{
						BuildInvocationID: "12345-1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, c := github.NewTestOIDCServer(t, now, tc.token)
			defer s.Close()

			if p, err := HostedActionsProvenance(context.Background(), tc.r, c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else {
				if want, got := tc.expected, p; !cmp.Equal(want, got) {
					t.Errorf("unexpected result\nwant: %#v\ngot:  %#v\ndiff: %v", want, got, cmp.Diff(want, got))
				}
			}
		})
	}
}
