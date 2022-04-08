package slsa

import (
	"context"
	"reflect"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func TestHostedActionsProvenance(t *testing.T) {
	testCases := []struct {
		name     string
		r        WorkflowRun
		expected *intoto.ProvenanceStatement
	}{
		{
			name: "empty",
			r:    WorkflowRun{},
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
	}

	for _, tc := range testCases {
		_, stop := github.NewTestOIDCServer(nil)
		defer stop()

		t.Run(tc.name, func(t *testing.T) {
			if p, err := HostedActionsProvenance(context.Background(), tc.r); err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else {
				if want, got := tc.expected, p; !reflect.DeepEqual(want, got) {
					t.Errorf("unexpected result, want: %#v, got: %#v", want, got)
				}
			}
		})
	}
}
