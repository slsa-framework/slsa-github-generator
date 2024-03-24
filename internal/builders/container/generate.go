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
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/builders/common"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// generateCmd returns the 'generate' command.
func generateCmd(provider slsa.ClientProvider, check func(error)) *cobra.Command {
	var predicatePath string

	c := &cobra.Command{
		Use:   "generate",
		Short: "Create a SLSA provenance predicate from a GitHub Action",
		Long: `Generate SLSA provenance predicate from a GitHub Action. This command assumes
that it is being run in the context of a Github Actions workflow.`,

		Run: func(_ *cobra.Command, _ []string) {
			ghContext, err := github.GetWorkflowContext()
			check(err)

			ctx := context.Background()

			b := common.GenericBuild{
				// NOTE: Subjects are nil because we are only writing the predicate.
				GithubActionsBuild: slsa.NewGithubActionsBuild(nil, &ghContext),
				BuildTypeURI:       containerBuildType,
			}

			if provider != nil {
				b.WithClients(provider)
			} else if utils.IsPresubmitTests() {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				b.WithClients(&slsa.NilClientProvider{})
			}

			g := slsa.NewHostedActionsGenerator(&b)
			if provider != nil {
				g.WithClients(provider)
			} else if utils.IsPresubmitTests() {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				g.WithClients(&slsa.NilClientProvider{})
			}

			p, err := g.Generate(ctx)
			check(err)

			pb, err := json.Marshal(p.Predicate)
			check(err)

			pf, err := utils.CreateNewFileUnderCurrentDirectory(predicatePath, os.O_WRONLY)
			check(err)

			_, err = pf.Write(pb)
			check(err)
		},
	}

	c.Flags().StringVarP(
		&predicatePath,
		"predicate", "p", "predicate.json",
		"Path to write the unsigned provenance predicate.",
	)

	return c
}
