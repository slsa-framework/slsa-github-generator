// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandSteps is a series of steps executed when running a build. These steps
// can be included in the buildConfig in provenance statements.
type CommandSteps []*CommandStep

// Run executes each step. If any step returns an error the error is returned
// and subsequent steps are not executed.
func (s CommandSteps) Run() error {
	for _, step := range s {
		if err := step.Run(); err != nil {
			return err
		}
	}
	return nil
}

// CommandStep is a command executed by the builder.
type CommandStep struct {
	Command    []string `json:"command"`
	Env        []string `json:"env"`
	WorkingDir string   `json:"workingDir"`
}

// New returns a docker command.
func New(args ...string) (*CommandStep, error) {
	// TODO(github.com/slsa-framework/slsa-github-generator/issues/57): implement build.
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("unable to find the docker command: %w", err)
	}

	command := []string{path}
	command = append(command, args...)

	return &CommandStep{
		Command: command,
	}, nil
}

// WithFlag adds a flag to the command.
func (c *CommandStep) WithFlag(name, value string) {
	flg := name
	if !strings.HasPrefix("-", flg) {
		flg = "--" + flg
	}
	if value != "" {
		flg += "=" + value
	}
	c.Command = append(c.Command, flg)
}

func (c *CommandStep) newCmd() (*exec.Cmd, error) {
	if len(c.Command) == 0 {
		return nil, errors.New("command is empty")
	}
	cmd := exec.Command(c.Command[0], c.Command[1:]...)
	cmd.Env = c.Env
	cmd.Dir = c.WorkingDir
	return cmd, nil
}

// Run runs the command routing output to stdout and stderr.
func (c *CommandStep) Run() error {
	cmd, err := c.newCmd()
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Output runs the command and returns the stdout.
func (c *CommandStep) Output() ([]byte, error) {
	cmd, err := c.newCmd()
	if err != nil {
		return nil, err
	}
	return cmd.Output()
}
