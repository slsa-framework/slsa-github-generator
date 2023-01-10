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

package runner

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// clearEnv clears everything and sets a basic path.
func clearEnv() func() {
	oldEnv := map[string]string{}
	for _, e := range os.Environ() {
		k, v, found := strings.Cut(e, "=")
		if found {
			_ = os.Unsetenv(k)
			oldEnv[k] = v
		}
	}
	// Set a basic PATH
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	// PWD is required for Go's implementation of os.Getwd
	os.Setenv("PWD", oldEnv["PWD"])

	return func() {
		for k, v := range oldEnv {
			_ = os.Setenv(k, v)
		}
	}
}

func TestCommandRunner_StepEnv(t *testing.T) {
	t.Cleanup(clearEnv())

	// The current Environ should be excluded from provenance..
	t.Setenv("TESTVAR", "VALUE")

	// Set GitHub env var. These shouldn't be output in the provenance.
	t.Setenv("GITHUB_FOO", "BAR")
	t.Setenv("RUNNER_HOGE", "FUGA")
	t.Setenv("CI", "true")

	out := &strings.Builder{}
	r := CommandRunner{
		Env: []string{"TEST=hoge"},
		Steps: []*CommandStep{
			{
				Command: []string{"bash", "-c", "echo -n $TEST"},
				// NOTE: this overrides other env var.
				Env: []string{"TEST=fuga"},
				// NOTE: WorkingDir default to CWD
			},
		},
		Stdout: out,
	}

	steps, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("unexpected number of steps: %v", len(steps))
	}

	if diff := cmp.Diff(steps[0].Command, []string{"bash", "-c", "echo -n $TEST"}); diff != "" {
		t.Fatalf("unexpected command: %v", diff)
	}

	sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(steps[0].Env, []string{"TEST=fuga"}, sorted); diff != "" {
		t.Fatalf("unexpected env: %v", diff)
	}

	if diff := cmp.Diff(steps[0].WorkingDir, pwd); diff != "" {
		t.Fatalf("unexpected working dir: %v", diff)
	}

	if want, got := "fuga", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func TestCommandRunner_RunnerEnv(t *testing.T) {
	t.Cleanup(clearEnv())

	// The current Environ should be excluded from provenance..
	t.Setenv("TESTVAR", "VALUE")

	// Set GitHub env var. These shouldn't be output in the provenance.
	t.Setenv("GITHUB_FOO", "BAR")
	t.Setenv("RUNNER_HOGE", "FUGA")
	t.Setenv("CI", "true")

	out := &strings.Builder{}
	r := CommandRunner{
		Env: []string{"RUNNER=hoge"},
		Steps: []*CommandStep{
			{
				Command: []string{"bash", "-c", "echo -n $STEP"},
				// NOTE: this overrides other env var.
				Env: []string{"STEP=fuga"},
				// NOTE: WorkingDir default to CWD
			},
		},
		Stdout: out,
	}

	steps, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("unexpected number of steps: %v", len(steps))
	}

	if diff := cmp.Diff(steps[0].Command, []string{"bash", "-c", "echo -n $STEP"}); diff != "" {
		t.Fatalf("unexpected command: %v", diff)
	}

	sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(steps[0].Env, []string{"RUNNER=hoge", "STEP=fuga"}, sorted); diff != "" {
		t.Fatalf("unexpected env: %v", diff)
	}

	if diff := cmp.Diff(steps[0].WorkingDir, pwd); diff != "" {
		t.Fatalf("unexpected working dir: %v", diff)
	}

	if want, got := "fuga", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func TestCommandRunner_GlobalEnv(t *testing.T) {
	t.Cleanup(clearEnv())

	// The current Environ should be available to the build but excluded
	// from provenance.
	t.Setenv("TESTVAR", "VALUE")

	out := &strings.Builder{}
	r := CommandRunner{
		Env: []string{"RUNNER=hoge"},
		Steps: []*CommandStep{
			{
				Command: []string{"bash", "-c", "echo -n $TESTVAR"},
			},
		},
		Stdout: out,
	}

	steps, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("unexpected number of steps: %v", len(steps))
	}

	if diff := cmp.Diff(steps[0].Command, []string{"bash", "-c", "echo -n $TESTVAR"}); diff != "" {
		t.Fatalf("unexpected command: %v", diff)
	}

	sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	// NOTE: TESTVAR not included in provenance.
	if diff := cmp.Diff(steps[0].Env, []string{"RUNNER=hoge"}, sorted); diff != "" {
		t.Fatalf("unexpected env: %v", diff)
	}

	if want, got := "VALUE", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func TestCommandRunner_RunnerMulti(t *testing.T) {
	t.Cleanup(clearEnv())

	out := &strings.Builder{}
	r := CommandRunner{
		Steps: []*CommandStep{
			{
				Command: []string{"bash", "-c", "echo $STEP1"},
				Env:     []string{"STEP1=hoge"},
			},
			{
				Command: []string{"bash", "-c", "echo $STEP2"},
				Env:     []string{"STEP2=fuga"},
			},
		},
		Stdout: out,
	}

	steps, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff := cmp.Diff(steps, []*CommandStep{
		{
			Command:    []string{"bash", "-c", "echo $STEP1"},
			Env:        []string{"STEP1=hoge"},
			WorkingDir: pwd,
		},
		{
			Command:    []string{"bash", "-c", "echo $STEP2"},
			Env:        []string{"STEP2=fuga"},
			WorkingDir: pwd,
		},
	})
	if diff != "" {
		t.Fatalf("unexpected result: %v", diff)
	}

	if want, got := "hoge\nfuga\n", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func Test_dedupEnv(t *testing.T) {
	tests := map[string]struct {
		input    []string
		expected []string
	}{
		"with duplicate": {
			input: []string{
				"FOO=hoge",
				"FOO=bar",
			},
			expected: []string{
				"FOO=bar",
			},
		},
		"reverse order": {
			input: []string{
				"FOO=bar",
				"FOO=hoge",
			},
			expected: []string{
				"FOO=hoge",
			},
		},
		"with extra": {
			input: []string{
				"FOO=hoge",
				"EXTRA=fuga",
				"FOO=bar",
			},
			expected: []string{
				"EXTRA=fuga",
				// NOTE: The active variable appears after EXTRA
				"FOO=bar",
			},
		},
		"no duplicate": {
			input: []string{
				"FOO=bar",
				"HOGE=fuga",
			},
			expected: []string{
				"FOO=bar",
				"HOGE=fuga",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(dedupEnv(tc.input), tc.expected); diff != "" {
				t.Fatalf("unexpected result: %v", diff)
			}
		})
	}
}
