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
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/internal/builders/docker/pkg"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

// DryRunCmd returns a new *cobra.Command that validates the input flags, and
// generates a BuildDefinition from them, or terminates with an error.
func DryRunCmd(check func(error)) *cobra.Command {
	inputOptions := &pkg.InputOptions{}
	var buildDefinitionPath string

	cmd := &cobra.Command{
		Use:   "dry-run [FLAGS]",
		Short: "Generates and stores a JSON-formatted BuildDefinition based on the input arguments.",
		Run: func(cmd *cobra.Command, args []string) {
			w, err := utils.CreateNewFileUnderCurrentDirectory(buildDefinitionPath, os.O_WRONLY)
			check(err)

			config, err := pkg.NewDockerBuildConfig(inputOptions)
			check(err)

			builder, err := pkg.NewBuilderWithGitFetcher(config)
			check(err)

			db, err := builder.SetUpBuildState()
			check(err)
			// Remove any temporary files that were fetched during the setup.
			defer db.RepoInfo.Cleanup()

			check(writeJSONToFile(*db.CreateBuildDefinition(), w))
		},
	}

	inputOptions.AddFlags(cmd)
	cmd.Flags().StringVarP(&buildDefinitionPath, "build-definition-path", "o", "",
		"Required - Path to store the generated BuildDefinition to.")

	return cmd
}

// BuildCmd returns a new *cobra.Command that builds the artifacts using the
// input flags, and prints out their digests, or terminates with an error.
func BuildCmd(check func(error)) *cobra.Command {
	inputOptions := &pkg.InputOptions{}
	var subjectsPath string
	var outputFolder string

	cmd := &cobra.Command{
		Use:   "build [FLAGS]",
		Short: "Builds the artifacts using the build config, source repo, and the builder image.",
		Run: func(cmd *cobra.Command, args []string) {
			absoluteOutputFolder, err := filepath.Abs(outputFolder)
			check(err)

			w, err := utils.CreateNewFileUnderCurrentDirectory(subjectsPath, os.O_WRONLY)
			check(err)
			config, err := pkg.NewDockerBuildConfig(inputOptions)
			check(err)

			builder, err := pkg.NewBuilderWithGitFetcher(config)
			check(err)

			db, err := builder.SetUpBuildState()
			check(err)
			// Remove any temporary files that were generated during the setup.
			defer db.RepoInfo.Cleanup()

			// Build artifacts and write them to the output folder.
			artifacts, err := db.BuildArtifacts(absoluteOutputFolder)
			check(err)
			check(writeJSONToFile(artifacts, w))
		},
	}

	inputOptions.AddFlags(cmd)
	cmd.Flags().StringVarP(&subjectsPath, "subjects-path", "o", "",
		"Required - Path to store a JSON-encoded array of subjects of the generated artifacts.")
	cmd.Flags().StringVar(&outputFolder, "output-folder", "",
		"Required - Path to a folder to store the generated artifacts.")
	cmd.MarkFlagRequired("output-folder")

	return cmd
}

func writeJSONToFile[T any](obj T, w io.Writer) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling the object failed: %w", err)
	}

	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("writing to file failed: %w", err)
	}
	return nil
}
