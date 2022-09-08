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
)

func TestCommandRunner_StepEnv(t *testing.T) {
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

	diff := cmp.Diff(steps, []*CommandStep{
		{
			Command: []string{"bash", "-c", "echo -n $TEST"},
			// TODO(https://github.com/slsa-framework/slsa-github-generator/issues/782): de-duplicate env.
			Env:        []string{"TEST=hoge", "TEST=fuga"},
			WorkingDir: pwd,
		},
	})
	if diff != "" {
		t.Fatalf("unexpected result: %v", diff)
	}

	if want, got := "fuga", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func TestCommandRunner_RunnerEnv(t *testing.T) {
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

	diff := cmp.Diff(steps, []*CommandStep{
		{
			Command:    []string{"bash", "-c", "echo -n $STEP"},
			Env:        []string{"RUNNER=hoge", "STEP=fuga"},
			WorkingDir: pwd,
		},
	})
	if diff != "" {
		t.Fatalf("unexpected result: %v", diff)
	}

	if want, got := "fuga", out.String(); want != got {
		t.Fatalf("unexpected env var value, want %q, got: %q", want, got)
	}
}

func TestCommandRunner_RunnerMulti(t *testing.T) {
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
