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

// This file contains definitions of the subcommands of the
// `slsa-docker-based-generator` command.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/slsa-framework/slsa-github-generator/internal/builders/docker/pkg"
	"github.com/spf13/cobra"
)

// DryRunCmd returns a new *cobra.Command that validates the input flags, and
// generates a BuildDefinition from them, or terminates with an error.
func DryRunCmd(check func(error)) *cobra.Command {
	io := &pkg.InputOptions{}
	var buildDefinitionPath string

	cmd := &cobra.Command{
		Use:   "dry-run [FLAGS]",
		Short: "Generates and stores a JSON-formatted BuildDefinition based on the input arguments.",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := pkg.NewDockerBuildConfig(io)
			check(err)

			bd := pkg.CreateBuildDefinition(config)
			check(writeBuildDefinitionToFile(*bd, buildDefinitionPath))
		},
	}

	io.AddFlags(cmd)

	cmd.Flags().StringVarP(&buildDefinitionPath, "build-definition-path", "o", "",
		"Required - Path to store the generated BuildDefinition to.")

	return cmd
}

func writeBuildDefinitionToFile(bd pkg.BuildDefinition, path string) error {
	bytes, err := json.Marshal(bd)
	if err != nil {
		return fmt.Errorf("couldn't marshal the BuildDefinition: %v", err)
	}

	if err := os.WriteFile(path, bytes, 0o600); err != nil {
		return fmt.Errorf("couldn't write BuildDefinition to file: %v", err)
	}
	return nil
}

// BuildCmd returns a new *cobra.Command that builds the artifacts using the
// input flags, and prints out their digests, or terminates with an error.
func BuildCmd(check func(error)) *cobra.Command {
	io := &pkg.InputOptions{}
	var forceCheckout bool

	cmd := &cobra.Command{
		Use:   "build [FLAGS]",
		Short: "Builds the artifacts using the build config, source repo, and the builder image.",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := pkg.NewDockerBuildConfig(io)
			check(err)

			builder, err := pkg.NewBuilderWithGitFetcher(*config, forceCheckout)
			check(err)

			db, err := builder.SetUpBuildState()
			// Remove any temporary files that were generated during the setup.
			defer db.RepoInfo.Cleanup()
			check(err)

			artifacts, err := db.BuildArtifact()
			check(err)

			log.Printf("Generated artifacts are: %v\n", artifacts)
			// TODO(#1191): Write subjects to a file.
		},
	}

	io.AddFlags(cmd)
	cmd.Flags().BoolVarP(&forceCheckout, "force-checkout", "f", false,
		"Optional - Forces checking out the source code from the given Git repo.")

	return cmd
}
