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
	"errors"

	// TODO: Allow use of other OIDC providers?
	// Enable the github OIDC auth provider.
	_ "github.com/sigstore/cosign/v2/pkg/providers/github"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "slsa-generator-generic",
		Short: "Generate SLSA provenance for Github Actions",
		Long: `Generate SLSA provenance for Github Actions.
For more information on SLSA, visit https://slsa.dev`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("expected command")
		},
	}
	c.AddCommand(versionCmd())
	c.AddCommand(attestCmd(nil, checkExit, sigstore.NewDefaultFulcio(), sigstore.NewDefaultRekor()))
	return c
}

func main() {
	checkExit(rootCmd().Execute())
}
