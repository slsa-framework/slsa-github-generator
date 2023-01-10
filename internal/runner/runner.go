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
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommandRunner runs commands and returns the build steps that were run.
type CommandRunner struct {
	// Stdout is the Writer used for Stdout. If nil then os.Stdout is used.
	Stdout io.Writer

	// Stderr is the Writer used for Stderr. If nil then os.Stderr is used.
	Stderr io.Writer

	// Env is global environment variables passed to all commands.
	Env []string

	// Steps are the steps to execute.
	Steps []*CommandStep
}

// CommandStep is a command that was executed by the builder.
type CommandStep struct {
	// WorkingDir is the working directory the command was executed in.
	WorkingDir string `json:"workingDir"`

	// Command is the command that was run.
	Command []string `json:"command"`

	// Env are the environment variables passed to the command.
	Env []string `json:"env"`
}

// Dry returns the command steps as they would be executed by the runner
// without actually executing the commands. This allows builders to get an
// accurate set of steps in a trusted environment as executing commands will
// execute untrusted code.
func (r *CommandRunner) Dry() (steps []*CommandStep, err error) {
	ctx := context.Background()
	for _, step := range r.Steps {
		var runStep *CommandStep
		runStep, err = r.runStep(ctx, step, true)
		if err != nil {
			return // steps, err
		}
		steps = append(steps, runStep)
	}
	return // steps, err
}

// Run executes a series of commands and returns the steps that were executed
// successfully. Commands are run in sequence and are expected to return a zero
// exit status.
//
// Global environment variables are merged with steps environment variables in
// the returned steps. In the case of duplicates the last occurrence has precidence.
// Environment variables are *not* inherited from the current process.
//
// The returned CommandSteps should be included in the buildConfig provenance.
// These are *not* the same as the runner commands. Env vars are sanitized, pwd
// is changed to the absolute path, and only commands that executed
// successfully are returned.
func (r *CommandRunner) Run(ctx context.Context) (steps []*CommandStep, err error) {
	for _, step := range r.Steps {
		var runStep *CommandStep
		runStep, err = r.runStep(ctx, step, false)
		if err != nil {
			return // steps, err
		}
		steps = append(steps, runStep)
	}
	return // steps, err
}

// runStep runs the build step and returns the CommandStep configuration
// actually used to run the command. If dry is true then the CommandStep is
// returned without executing the command.
func (r *CommandRunner) runStep(ctx context.Context, step *CommandStep, dry bool) (*CommandStep, error) {
	if len(step.Command) == 0 {
		return nil, errors.New("command is empty")
	}

	name := step.Command[0]
	args := step.Command[1:]

	cmd := exec.CommandContext(ctx, name, args...)
	pwd, err := filepath.Abs(step.WorkingDir)
	if err != nil {
		return nil, err
	}
	cmd.Dir = pwd
	cmd.Stdout = os.Stdout
	if r.Stdout != nil {
		cmd.Stdout = r.Stdout
	}
	cmd.Stderr = os.Stderr
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	}

	// We will copy over environment variables from the builder when executing
	// the command, However, we won't include the builder's environment
	// variables into the provenance as they are environment specific and
	// inhibit reproducibility.
	// See: https://github.com/slsa-framework/slsa-github-generator/issues/822

	var userEnv []string
	userEnv = append(userEnv, r.Env...)
	userEnv = append(userEnv, step.Env...)
	userEnv = dedupEnv(userEnv)

	cmdEnv := os.Environ()
	cmdEnv = append(cmdEnv, userEnv...)

	// Set the environment for the command. Duplicates that appear later in the
	// list override earlier entries. This is enforced by the stdlib exec package.
	cmd.Env = cmdEnv

	if !dry {
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	}

	return &CommandStep{
		Command:    append([]string{name}, args...),
		Env:        userEnv,
		WorkingDir: pwd,
	}, nil
}

func dedupEnv(env []string) []string {
	var deduped []string
	seen := map[string]bool{}
	for i := len(env) - 1; i >= 0; i-- {
		k, _, found := strings.Cut(env[i], "=")
		if found && !seen[k] {
			deduped = append(deduped, env[i])
			seen[k] = true
		}
	}
	return deduped
}
