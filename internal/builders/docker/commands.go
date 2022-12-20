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

// InputOptions are the common options for the dry run and build command.
type InputOptions struct {
	BuildConfigPath string
	SourceRepo      string
	GitCommitHash   string
	BuilderImage    string
}

// AddFlags adds input flags to the given command.
func (o *InputOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.BuildConfigPath, "build-config-path", "c", "",
		"Required - Path to a toml file containing the build configs.")

	cmd.Flags().StringVarP(&o.SourceRepo, "source-repo", "s", "",
		"Required - URL of the source repo.")

	cmd.Flags().StringVarP(&o.GitCommitHash, "git-commit-hash", "g", "",
		"Required - SHA1 Git commit digest of the revision of the source code to build the artefact from.")

	cmd.Flags().StringVarP(&o.BuilderImage, "builder-image", "b", "",
		"Required - URL indicating the Docker builder image, including a URI and image digest.")
}

// DryRunCmd validates the input flags, generates a BuildDefinition from them.
func DryRunCmd(check func(error)) *cobra.Command {
	o := &InputOptions{}
	var buildDefinitionPath string

	cmd := &cobra.Command{
		Use:   "dry-run [FLAGS]",
		Short: "Generates and stores a JSON-formatted BuildDefinition based on the input arguments.",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO(#1191): Parse the input arguments into an instance of BuildDefinition.
			bd := &pkg.BuildDefinition{}
			check(writeBuildDefinitionToFile(*bd, buildDefinitionPath))
		},
	}

	o.AddFlags(cmd)

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

// BuildCmd builds the artifacts using the input flags, and prints out their digests, or exists with an error.
func BuildCmd(check func(error)) *cobra.Command {
	o := &InputOptions{}

	cmd := &cobra.Command{
		Use:   "build [FLAGS]",
		Short: "Builds the artifacts using the build config, source repo, and the builder image.",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO(#1191): Set up build state and build the artifact.
			artifacts := "To be implemented"
			log.Printf("Generated artifacts are: %v\n", artifacts)
			// TODO(#1191): Write subjects to file.
		},
	}

	o.AddFlags(cmd)

	return cmd
}
