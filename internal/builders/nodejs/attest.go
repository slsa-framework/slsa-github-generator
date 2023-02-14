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

// attestCmd runs the 'attest' command.
func attestCmd(provider slsa.ClientProvider, check func(error),
	signer signing.Signer, tlog signing.TransparencyLog,
) *cobra.Command {
	c := &cobra.Command{
		Use:   "attest",
		Short: "Run attest command",
		Long:  `Run attest command to generate and sign attest file.`,

		Run: func(cmd *cobra.Command, args []string) {
			if err := attest(); err != nil {
				panic(err)
			}
		},
	}

	return c
}

func attest() error {
	// 1. Retrieve the provenance metadata from env variable UNTRUSTED_PROVENANCE_METADATA.

	// 4. Create / sign a provenance file as per UNTRUSTED_PROVENANCE_METADATA.

	// 5. Output the list of provenance files generated in a format TBD (sha256sum?).
	return nil
}
