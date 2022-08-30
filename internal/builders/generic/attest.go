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

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/errors"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// attestCmd returns the 'attest' command.
func attestCmd(provider slsa.ClientProvider, check func(error)) *cobra.Command {
	var attPath string
	var subjects string

	c := &cobra.Command{
		Use:   "attest",
		Short: "Create a signed SLSA provenance attestation from a Github Action",
		Long: `Generate and sign SLSA provenance from a Github Action to form an attestation
and upload to a Rekor transparency log. This command assumes that it is being
run in the context of a Github Actions workflow.`,

		Run: func(cmd *cobra.Command, args []string) {
			ghContext, err := github.GetWorkflowContext()
			check(err)

			// Verify the extension path and extension.
			err = utils.VerifyAttestationPath(attPath)
			check(err)

			parsedSubjects, err := parseSubjects(subjects)
			check(err)

			if len(parsedSubjects) == 0 {
				check(errors.New("expected at least one subject"))
			}

			ctx := context.Background()

			b := provenanceOnlyBuild{
				GithubActionsBuild: slsa.NewGithubActionsBuild(parsedSubjects, ghContext),
			}
			if provider != nil {
				b.WithClients(provider)
			} else {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				if utils.IsPresubmitTests() {
					b.WithClients(&slsa.NilClientProvider{})
				}
			}

			g := slsa.NewHostedActionsGenerator(&b)
			if provider != nil {
				g.WithClients(provider)
			} else {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				if utils.IsPresubmitTests() {
					g.WithClients(&slsa.NilClientProvider{})
				}
			}

			p, err := g.Generate(ctx)
			check(err)

			// Note: the path is validated within CreateNewFileUnderCurrentDirectory().
			var attBytes []byte
			if utils.IsPresubmitTests() {
				attBytes, err = json.Marshal(p)
				check(err)
			} else {
				s := sigstore.NewDefaultFulcio()
				att, err := s.Sign(ctx, &intoto.Statement{
					StatementHeader: p.StatementHeader,
					Predicate:       p.Predicate,
				})
				check(err)

				r := sigstore.NewDefaultRekor()
				_, err = r.Upload(ctx, att)
				check(err)

				attBytes = att.Bytes()
			}

			f, err := utils.CreateNewFileUnderCurrentDirectory(attPath, os.O_WRONLY)
			check(err)

			_, err = f.Write(attBytes)
			check(err)
		},
	}

	c.Flags().StringVarP(&attPath, "signature", "g", "", "Path to write the signed provenance.")
	c.Flags().StringVarP(&subjects, "subjects", "s", "", "Formatted list of subjects in the same format as sha256sum (base64 encoded).")

	return c
}
