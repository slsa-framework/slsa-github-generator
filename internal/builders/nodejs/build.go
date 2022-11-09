// Copyright 2022 slsa Authors
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

func build() error {
	integration, err := slsaIntegrationNew()
	if err != nil {
		return err
	}

	r := runner.CommandRunner{}

	// Ci command.
	if err := createCiCommands(&r, integration.Inputs); err != nil {
		return nil
	}

	// Run command.
	if err := createRunCommands(&r, integration.Inputs); err != nil {
		return err
	}

	// Pack command.
	if err := createPackCommands(&r, integration.Inputs); err != nil {
		return err
	}

	if integration.Inputs.DryRun {
		return runDry(integration, r)
	}

	return run(integration, r)
}

func runDry(integration *slsaIntegration,
	r runner.CommandRunner,
) error {
	// This builder supports a single provenance file generation.
	if err := validateArtifacts(integration.Inputs.Artifacts); err != nil {
		return nil
	}
	// Run dry.
	steps, err := r.Dry()
	if err != nil {
		return err
	}

	// Generate the provenance metadata.
	artifact := integration.Inputs.Artifacts[0]
	name := strings.TrimSuffix(filepath.Base(artifact.Path), ".tgz")
	metadata := slsaDryRunOutput{
		name + ".intoto.jsonl": []slsaDryMetadata{
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

func validateArtifacts(artifacts []slsaArtifact) error {
	if len(artifacts) != 1 {
		return fmt.Errorf("%w: only 1 artifact is supported", errorInvalidField)
	}

	if artifacts[0].Path == "" {
		return fmt.Errorf("%w: artifact path empty", errorInvalidField)
	}

	if len(artifacts[0].Digests) == 0 {
		return fmt.Errorf("%w: artifact digest empty", errorInvalidField)
	}

	return nil
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

func run(integration *slsaIntegration,
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
	artifacts := slsaArtifact{
		Path: path,
		Digests: slsaDigests{
			"sha256": hex.EncodeToString(bs),
		},
	}

	return writeOutput(integration.OutputPath, artifacts)
}

func createPackCommands(r *runner.CommandRunner, inputs *slsaInputs) error {
	// We currently do not
	// support arguments to pack, e.g. `--pack-destination`.
	// Note: pack-destination only supported version 7.x above.
	// https://docs.npmjs.com/cli/v7/commands/npm-pack.
	cmd := []string{"npm", "pack", "--json"}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}

	r.Steps = append(r.Steps,
		&runner.CommandStep{
			Command:    cmd,
			Env:        nil,
			WorkingDir: workingDir,
		},
	)
	return nil
}

func createRunCommands(r *runner.CommandRunner, inputs *slsaInputs) error {
	script, present := inputs.WorkflowInputs["run-scripts"]
	if !present {
		return fmt.Errorf("%w: 'run-scripts' not present", errorInvalidField)
	}
	cmd := []string{"npm", "run"}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}

	scripts := strings.Split(script, ",")
	for i := range scripts {
		s := strings.TrimSpace(string(scripts[i]))
		r.Steps = append(r.Steps,
			&runner.CommandStep{
				Command:    append(cmd, s),
				WorkingDir: workingDir,
			},
		)
	}

	return nil
}

func createCiCommands(r *runner.CommandRunner, inputs *slsaInputs) error {
	args, present := inputs.WorkflowInputs["ci-arguments"]
	if !present {
		return fmt.Errorf("%w: 'ci-arguments' not present", errorInvalidField)
	}
	cmd := []string{"npm", "ci"}
	if args != "" {
		cmd = append(cmd, strings.Split(args, ",")...)
	}

	workingDir, present := inputs.WorkflowInputs["working-directory"]
	if !present {
		return fmt.Errorf("%w: 'working-directory' not present", errorInvalidField)
	}

	r.Steps = append(r.Steps, &runner.CommandStep{
		Command:    cmd,
		Env:        nil,
		WorkingDir: workingDir,
	})

	return nil
}
