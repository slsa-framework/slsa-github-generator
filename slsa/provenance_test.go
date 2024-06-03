// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slsa

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	slsa02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
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

func (*TestBuild) BuildConfig(context.Context) (any, error) {
	return testBuildConfig, nil
}

func TestHostedActionsProvenance(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		b        BuildType
		token    *github.OIDCToken
		expected *intoto.ProvenanceStatement
		name     string
	}{
		{
			name: "empty",
			b: &TestBuild{
				GithubActionsBuild: NewGithubActionsBuild(
					nil, &github.WorkflowContext{}, github.VarsContext{}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{""},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa02.PredicateSLSAProvenance,
				},
				Predicate: slsa02.ProvenancePredicate{
					Builder: slsacommon.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Invocation: slsa02.ProvenanceInvocation{
						Environment: map[string]any{
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
						Parameters: WorkflowParameters{
							VarsContext: github.VarsContext{},
						},
					},
					Metadata: &slsa02.ProvenanceMetadata{},
				},
			},
		},
		{
			name: "invocation complete",
			b: &TestBuild{
				GithubActionsBuild: NewGithubActionsBuild(nil, &github.WorkflowContext{
					RunID:      "12345",
					RunAttempt: "1",
					EventName:  "pull_request",
					Event: map[string]any{
						"inputs": map[string]any{
							"key1": "value1",
							"key2": 2,
							"key3": true,
						},
					},
					SHA:       "abcde",
					RefType:   "branch",
					Ref:       "some/ref",
					BaseRef:   "some/base_ref",
					HeadRef:   "some/head_ref",
					RunNumber: "102937",
					Actor:     "user",
				}, github.VarsContext{
					"REPO_VAR": "value",
				}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa02.PredicateSLSAProvenance,
				},
				Predicate: slsa02.ProvenancePredicate{
					Builder: slsacommon.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Invocation: slsa02.ProvenanceInvocation{
						Environment: map[string]any{
							"github_run_id":      "12345",
							"github_run_attempt": "1",
							"github_actor":       "user",
							"github_base_ref":    "some/base_ref",
							"github_event_name":  "pull_request",
							"github_event_payload": map[string]any{
								"inputs": map[string]any{
									"key1": "value1",
									"key2": 2,
									"key3": true,
								},
							},
							"github_head_ref":         "some/head_ref",
							"github_ref":              "some/ref",
							"github_ref_type":         "branch",
							"github_repository_owner": "",
							"github_run_number":       "102937",
							"github_sha1":             "abcde",
						},
						Parameters: WorkflowParameters{
							EventInputs: map[string]any{
								"key1": "value1",
								"key2": 2,
								"key3": true,
							},
							VarsContext: github.VarsContext{
								"REPO_VAR": "value",
							},
						},
						ConfigSource: slsa02.ConfigSource{
							Digest: slsacommon.DigestSet{
								"sha1": "abcde",
							},
						},
					},
					Metadata: &slsa02.ProvenanceMetadata{
						BuildInvocationID: "12345-1",
						Completeness: slsa02.ProvenanceComplete{
							Parameters: true,
						},
					},
				},
			},
		},
		{
			name: "empty invocation parameters",
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
				}, nil).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa02.PredicateSLSAProvenance,
				},
				Predicate: slsa02.ProvenancePredicate{
					Builder: slsacommon.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Invocation: slsa02.ProvenanceInvocation{
						Environment: map[string]any{
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
						ConfigSource: slsa02.ConfigSource{
							Digest: slsacommon.DigestSet{
								"sha1": "abcde",
							},
						},
					},
					Metadata: &slsa02.ProvenanceMetadata{
						BuildInvocationID: "12345-1",
					},
				},
			},
		},
		{
			name: "invocation with inputs",
			b: &TestBuild{
				GithubActionsBuild: NewGithubActionsBuild(nil, &github.WorkflowContext{
					RunID:      "12345",
					RunAttempt: "1",
					EventName:  "pull_request",
					Event: map[string]any{
						"inputs": map[string]any{
							"key1": "value1",
							"key2": 2,
							"key3": true,
						},
					},
					SHA:       "abcde",
					RefType:   "branch",
					Ref:       "some/ref",
					BaseRef:   "some/base_ref",
					HeadRef:   "some/head_ref",
					RunNumber: "102937",
					Actor:     "user",
				}, nil).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa02.PredicateSLSAProvenance,
				},
				Predicate: slsa02.ProvenancePredicate{
					Builder: slsacommon.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Invocation: slsa02.ProvenanceInvocation{
						Environment: map[string]any{
							"github_run_id":      "12345",
							"github_run_attempt": "1",
							"github_actor":       "user",
							"github_base_ref":    "some/base_ref",
							"github_event_name":  "pull_request",
							"github_event_payload": map[string]any{
								"inputs": map[string]any{
									"key1": "value1",
									"key2": 2,
									"key3": true,
								},
							},
							"github_head_ref":         "some/head_ref",
							"github_ref":              "some/ref",
							"github_ref_type":         "branch",
							"github_repository_owner": "",
							"github_run_number":       "102937",
							"github_sha1":             "abcde",
						},
						Parameters: WorkflowParameters{
							EventInputs: map[string]any{
								"key1": "value1",
								"key2": 2,
								"key3": true,
							},
						},
						ConfigSource: slsa02.ConfigSource{
							Digest: slsacommon.DigestSet{
								"sha1": "abcde",
							},
						},
					},
					Metadata: &slsa02.ProvenanceMetadata{
						BuildInvocationID: "12345-1",
						Completeness: slsa02.ProvenanceComplete{
							Parameters: true,
						},
					},
				},
			},
		},
		{
			name: "invocation with vars",
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
				}, github.VarsContext{
					"REPO_VAR": "value",
				}).WithClients(&NilClientProvider{}),
			},
			token: &github.OIDCToken{
				Audience: []string{"hoge"},
				Expiry:   now.Add(1 * time.Hour),
			},
			expected: &intoto.ProvenanceStatement{
				StatementHeader: intoto.StatementHeader{
					Type:          intoto.StatementInTotoV01,
					PredicateType: slsa02.PredicateSLSAProvenance,
				},
				Predicate: slsa02.ProvenancePredicate{
					Builder: slsacommon.ProvenanceBuilder{
						ID: GithubHostedActionsBuilderID,
					},
					BuildType:   testBuildType,
					BuildConfig: testBuildConfig,
					Invocation: slsa02.ProvenanceInvocation{
						Environment: map[string]any{
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
						Parameters: WorkflowParameters{
							VarsContext: github.VarsContext{
								"REPO_VAR": "value",
							},
						},
						ConfigSource: slsa02.ConfigSource{
							Digest: slsacommon.DigestSet{
								"sha1": "abcde",
							},
						},
					},
					Metadata: &slsa02.ProvenanceMetadata{
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
