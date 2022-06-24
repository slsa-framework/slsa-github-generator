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

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func Test_action_getEventValue(t *testing.T) {
	cases := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "empty map",
			m:        nil,
			key:      "test.foo.bar",
			expected: "",
		},
		{
			name: "empty key",
			m: map[string]interface{}{
				"test": "hoge",
			},
			key:      "",
			expected: "",
		},
		{
			name: "shallow",
			m: map[string]interface{}{
				"test": "hoge",
			},
			key:      "test",
			expected: "hoge",
		},
		{
			name: "deep",
			m: map[string]interface{}{
				"test": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "hoge",
					},
				},
			},
			key:      "test.foo.bar",
			expected: "hoge",
		},
		{
			name: "value type",
			m: map[string]interface{}{
				"test": map[string]interface{}{
					"foo": map[string]string{
						"bar": "hoge",
					},
				},
			},
			key:      "test.foo.bar",
			expected: "hoge",
		},
		{
			name: "partial key",
			m: map[string]interface{}{
				"test": map[string]interface{}{
					"foo": map[string]string{
						"bar": "hoge",
					},
				},
			},
			key:      "test.foo",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := action{
				event: tc.m,
			}
			if want, got := tc.expected, a.getEventValue(tc.key); want != got {
				t.Errorf("unexpected response, want: %q, got: %q", want, got)
			}
		})
	}
}

func Test_action_getRepoRef(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		now := time.Date(2022, 5, 3, 14, 49, 0, 0, time.UTC)
		t.Setenv("GITHUB_CONTEXT", `{"repository": "githubuser/reponame"}`)
		s, c := github.NewTestOIDCServer(t, now, &github.OIDCToken{
			Audience:          []string{"githubuser/reponame/detect-workflow"},
			Expiry:            time.Date(2022, 5, 4, 0, 0, 0, 0, time.UTC),
			JobWorkflowRef:    "githubuser/reponame/path/to/workflow@refs/heads/main",
			RepositoryOwnerID: "1",
			ActorID:           "1",
			RepositoryID:      "1",
		})
		defer s.Close()

		a := action{
			getenv: func(k string) string {
				if k == "GITHUB_REPOSITORY" {
					return "githubuser/reponame"
				}
				return ""
			},
			client: c,
		}

		repo, ref, err := a.getRepoRef(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want, got := "githubuser/reponame", repo; want != got {
			t.Errorf("unexpected repository, want: %q, got: %q", want, got)
		}
		if want, got := "refs/heads/main", ref; want != got {
			t.Errorf("unexpected ref, want: %q, got: %q", want, got)
		}
	})

	t.Run("pull_request", func(t *testing.T) {
		now := time.Date(2022, 5, 3, 14, 49, 0, 0, time.UTC)
		s, c := github.NewTestOIDCServer(t, now, nil)
		defer s.Close()

		a := action{
			getenv: func(k string) string {
				env := map[string]string{
					"GITHUB_REPOSITORY": "githubuser/reponame",
					"GITHUB_EVENT_NAME": "pull_request",
					"GITHUB_HEAD_REF":   "refs/heads/mybranch",
				}
				return env[k]
			},
			client: c,
			event: map[string]any{
				"pull_request": map[string]any{
					"head": map[string]any{
						"repo": map[string]any{
							"full_name": "otheruser/reponame",
						},
					},
				},
			},
		}

		repo, ref, err := a.getRepoRef(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want, got := "otheruser/reponame", repo; want != got {
			t.Errorf("unexpected repository, want: %q, got: %q", want, got)
		}
		if want, got := "refs/heads/mybranch", ref; want != got {
			t.Errorf("unexpected ref, want: %q, got: %q", want, got)
		}
	})
}

func Test_newAction(t *testing.T) {
	type args struct {
		getenv func(string) string
		c      *github.OIDCClient
	}
	tests := []struct {
		name    string
		args    args
		want    *action
		wantErr bool
	}{
		{
			name: "failure with empty string",
			args: args{
				getenv: func(k string) string {
					return ""
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "success",
			args: args{
				getenv: func(k string) string {
					// create a temp file with key:value

					f, err := os.CreateTemp("", "")
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					defer f.Close()
					if _, err := f.Write([]byte(`{"test":"hoge"}`)); err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					return f.Name()
				},
			},
			want: &action{
				getenv: func(k string) string {
					return "test:hoge"
				},
				event: map[string]any{
					"test": "hoge",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAction(tt.args.getenv, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var diff string
			if tt.want != nil && tt.want.event != nil {
				if diff = cmp.Diff(got.event, tt.want.event); diff != "" {
					t.Errorf("newAction() event mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
