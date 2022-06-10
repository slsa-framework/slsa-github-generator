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
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Command is a docker command.
type Command struct {
	cmd *exec.Cmd
}

// NewCommand returns a docker command.
func NewCommand(args ...string) (*Command, error) {
	// TODO(github.com/slsa-framework/slsa-github-generator/issues/57): implement build.
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("unable to find the docker command: %w", err)
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &Command{
		cmd: cmd,
	}, nil
}

// WithFlag adds a flag to the command.
func (c *Command) WithFlag(name, value string) {
	flg := name
	if !strings.HasPrefix("-", flg) {
		flg = "--" + flg
	}
	if value != "" {
		flg += "=" + value
	}
	c.cmd.Args = append(c.cmd.Args, flg)
}

// Run runs the command.
func (c *Command) Run() error {
	return c.cmd.Run()
}
