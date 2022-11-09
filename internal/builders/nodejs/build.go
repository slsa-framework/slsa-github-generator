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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/internal/runner"

	"github.com/spf13/cobra"
)

// buildCmd runs the 'build' command.
func buildCmd(check func(error)) *cobra.Command {
	c := &cobra.Command{
		Use:   "build",
		Short: "build a project",
		Long:  `build a project. Example: ./binary build --dry-run"`,

		Run: func(cmd *cobra.Command, args []string) {
			if err := build(); err != nil {
				panic(err)
			}
		},
	}
	return c
}

type dryRunOutputMetadata map[string][]metadata

type metadata struct {
	Name    string                `json:"name"`
	Digests SLSADigests           `json:"digests"`
	Steps   []*runner.CommandStep `json:"steps"`
}

func build() error {
	integration, err := SLSAIntegrationNew()
	if err != nil {
		return err
	}
	// DEBUG
	// encoder := json.NewEncoder(os.Stdout)
	// if err := encoder.Encode(*integration); err != nil {
	// 	return err
	// }

	r := runner.CommandRunner{}

	// Ci command.
	c, err := createCiCommands(integration.Inputs)
	r.Steps = append(r.Steps, c...)

	// Run command.
	c, err = createRunCommands(integration.Inputs)
	r.Steps = append(r.Steps, c...)

	// Pack command. Note: we currently do not
	// support arguments to pack, e.g. `--pack-destination`.
	// Note: pack-destination only supported version 7.x above.
	// https://docs.npmjs.com/cli/v7/commands/npm-pack.
	c, err = createPackCommands(integration.Inputs)
	r.Steps = append(r.Steps, c...)

	if integration.Inputs.DryRun {
		return runDry(integration, r)
	}

	return run(integration, r)
}

func runDry(integration *SLSAIntegration,
	r runner.CommandRunner,
) error {
	// This builder supports a single provenance file generation.
	if len(integration.Inputs.Artifacts) != 1 {
		return fmt.Errorf("%w: only 1 artifact is supported", errorInvalidField)
	}
	// Run dry.
	steps, err := r.Dry()
	if err != nil {
		return err
	}

	// Generate the provenance metadata.
	artifact := integration.Inputs.Artifacts[0]
	name := strings.TrimSuffix(filepath.Base(artifact.Path), ".tgz")
	metadata := dryRunOutputMetadata{
		name + ".intoto.jsonl": []metadata{
			{
				Name:    name,
				Digests: artifact.Digests,
				Steps:   steps,
			},
		},
	}

	// Write the metadata to the output file.
	return writeOutput(integration.OutputPath, metadata)
}

func writeOutput(path string, i interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(i); err != nil {
		return err
	}

	return nil
}

func run(integration *SLSAIntegration,
	r runner.CommandRunner,
) error {
	_, err := r.Run(context.Background())
	if err != nil {
		return err
	}

	workingDir, present := integration.Inputs.WorkflowInputs["working-directory"]
	if !present {
		return fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}
	// TODO: extract the name of the project from manifest
	// by parsing the ouput of `pack --json`.
	path := filepath.Join(workingDir, "hello-1.0.0.tgz")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	artifacts := SLSAArtifact{
		Path: path,
		Digests: SLSADigests{
			"sha256": hex.EncodeToString(bs),
		},
	}

	return writeOutput(integration.OutputPath, artifacts)
}

func createPackCommands(inputs *SLSAInputs) ([]*runner.CommandStep, error) {
	cmd := []string{"npm", "pack", "--json"}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return nil, fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}
	return []*runner.CommandStep{
		{
			Command:    cmd,
			Env:        nil,
			WorkingDir: workingDir,
		},
	}, nil
}

func createRunCommands(inputs *SLSAInputs) ([]*runner.CommandStep, error) {
	script, present := inputs.WorkflowInputs["run-scripts"]
	if !present {
		return nil, fmt.Errorf("%w: 'run-scripts' not present", errorInvalidField)
	}
	cmd := []string{"npm", "run"}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return nil, fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}
	steps := []*runner.CommandStep{}
	scripts := strings.Split(script, ",")
	for i := range scripts {
		s := strings.TrimSpace(string(scripts[i]))
		steps = append(steps,
			&runner.CommandStep{
				Command:    append(cmd, s),
				WorkingDir: workingDir,
			},
		)
	}
	return steps, nil
}

func createCiCommands(inputs *SLSAInputs) ([]*runner.CommandStep, error) {
	args, present := inputs.WorkflowInputs["ci-arguments"]
	if !present {
		return nil, fmt.Errorf("%w: 'ci-arguments' not present", errorInvalidField)
	}
	cmd := []string{"npm", "ci"}
	if args != "" {
		cmd = append(cmd, strings.Split(args, ",")...)
	}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return nil, fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}
	return []*runner.CommandStep{
		{
			Command:    cmd,
			Env:        nil,
			WorkingDir: workingDir,
		},
	}, nil
}
