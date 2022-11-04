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

	"github.com/slsa-framework/slsa-github-generator/signing"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// provenanceCmd runs the 'provenance' command.
func provenanceCmd(provider slsa.ClientProvider, check func(error), signer signing.Signer, tlog signing.TransparencyLog) *cobra.Command {
	var provenanceDirectory string

	c := &cobra.Command{
		Use:   "provenance",
		Short: "Run provenance command",
		Long:  `Run provenance command to generate and sign provenance file.`,

		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	c.Flags().StringVarP(&provenanceDirectory, "directory", "d", "", "Working directory to issue commands.")

	return c
}
