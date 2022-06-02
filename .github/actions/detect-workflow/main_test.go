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

	"github.com/slsa-framework/slsa-github-generator/github"
)

func Test_getStr(t *testing.T) {
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
			if want, got := tc.expected, getStr(tc.m, tc.key); want != got {
				t.Errorf("unexpected response, want: %q, got: %q", want, got)
			}
		})
	}
}

func Test_getRepoRef(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		now := time.Date(2022, 5, 3, 14, 49, 0, 0, time.UTC)
		os.Setenv("GITHUB_CONTEXT", `{"repository": "githubuser/reponame"}`)
		s, c := github.NewTestOIDCServer(t, now, &github.OIDCToken{
			Audience:       []string{"githubuser/reponame/detect-workflow"},
			Expiry:         time.Date(2022, 5, 4, 0, 0, 0, 0, time.UTC),
			JobWorkflowRef: "githubuser/reponame/path/to/workflow@refs/heads/main",
		})
		defer s.Close()

		repo, ref, err := getRepoRef(context.Background(), c)
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
		os.Setenv("GITHUB_CONTEXT", `{"repository": "githubuser/reponame", "head_ref": "refs/heads/mybranch", "event_name": "pull_request", "event": {"pull_request": {"head": {"repo": {"full_name": "otheruser/reponame"}}}}}`)
		s, c := github.NewTestOIDCServer(t, now, nil)
		defer s.Close()

		repo, ref, err := getRepoRef(context.Background(), c)
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
