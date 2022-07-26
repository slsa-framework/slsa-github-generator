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

import "github.com/spf13/cobra"

// attestCmd returns the 'attest' command.
func attestCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "attest",
		Short: "Generate signed SLSA provenance for a Docker image.",
		Long: `Generate and sign SLSA provenance for a Docker image and upload it to the
Docker registry. This command assumes that it is being run in the context of a
Github Actions workflow.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO(github.com/slsa-framework/slsa-github-generator/issues/57): implement attest command
			panic("not implemented")
		},
	}

	// TODO(github.com/slsa-framework/slsa-github-generator/issues/57): flags

	return c
}
