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
	"github.com/spf13/cobra"
)

// buildCmd runs the 'build' command.
func buildCmd(check func(error)) *cobra.Command {
	var dryRun bool

	c := &cobra.Command{
		Use:   "build",
		Short: "build a project",
		Long:  `build a project. Example: ./binary build --dry-run"`,

		Run: func(cmd *cobra.Command, args []string) {
			if err := build(dryRun); err != nil {
				panic(err)
			}
		},
	}

	c.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"Perform a dry run only. Do not build. Output provenance metadata (steps and provenance filenames)")
	return c
}

func build(dryRun bool) error {
	// 1. Retrieve the workflow inputs from env variable SLSA_WORKFLOW_INPUTS.

	// 2. Install dependencies via `npm ci <inputs.ci-arguments>`

	// 3. Build via `npm run <inputs.run-scripts>`

	// 4. Create the final package tarball.
	/* TODO: only run if non-empty.
	   Note: pack-destination only supported version 7.x above.
	   https://docs.npmjs.com/cli/v7/commands/npm-pack.
	   This outputs a .tgz. Before running this command, let's record the .tgz
	   files and their hashes, so that we can identify the new file without the need to parse
	   the manifest.json.
	   echo "npm pack --pack-destination="./out"
	   copy tarball to upper folder to make the tarball accessible to next step.
	*/
	// TODO: output the list of artifacts and their corresponding build steps.
	// The tarball name into a step output: echo "filename=$TARBALL" >> "$GITHUB_OUTPUT"

	if dryRun {
		// Output the proveance metadata in a format:
		/* METADATA={
			"provenance1.intoto.jsonl":{
				"artifact-1":{
					"buildSteps":[]Steps{
						"workingDir": string,
						"env": map[string]string,
						"command": []string
					}
				},
				"artifact-2":{
					"buildSteps":[]Steps{
						"workingDir": string,
						"env": map[string]string,
						"command": []string
					}
				}
			},
			"provenance2.intoto.jsonl":{
				"artifact-3":{
					"buildSteps":[]Steps{
						"workingDir": string,
						"env": map[string]string,
						"command": []string
					}
				},
			}
		}
		*/
		return nil
	}
	// 4. Output the list of artifacts in a format TBD (sha256sum?).

	return nil
}
