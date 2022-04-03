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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsav02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

func parseSubjects(subjectsCSV string) ([]intoto.Subject, error) {
	var parsed []intoto.Subject

	subjects := strings.Split(subjectsCSV, "|")
	for _, s := range subjects {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) == 0 {
			return nil, errors.New("missing subject name")
		}
		if len(parts) == 1 {
			return nil, fmt.Errorf("expected sha256 hash for subject %q", parts[0])
		}

		parsed = append(parsed, intoto.Subject{
			Name: parts[0],
			Digest: slsav02.DigestSet{
				"sha256": parts[1],
			},
		})
	}
	return parsed, nil
}

func getFile(path string) (io.Writer, error) {
	if path == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
}

// attestCmd returns the 'attest' command.
func attestCmd() *cobra.Command {
	var attPath string
	var subjects string

	c := &cobra.Command{
		Use:   "attest",
		Short: "Create a signed SLSA attestation from a Github Action",
		Long: `Generate and sign SLSA provenance from a Github Action to form an attestation
and upload to a Rekor transparency log. This command assumes that it is being
run in the context of a Github Actions workflow.`,

		Run: func(cmd *cobra.Command, args []string) {
			ghContext, err := github.GetWorkflowContext()
			check(err)

			parsedSubjects, err := parseSubjects(subjects)
			check(err)

			if len(parsedSubjects) == 0 {
				check(errors.New("expected at least one subject"))
			}

			p, err := slsa.HostedActionsProvenance(slsa.NewWorkflowRun(parsedSubjects, ghContext))
			check(err)

			if attPath != "" {
				ctx := context.Background()

				s := sigstore.NewDefaultSigner()
				att, err := s.Sign(ctx, p)
				check(err)

				_, err = s.Upload(ctx, att)
				check(err)

				f, err := getFile(attPath)
				check(err)

				_, err = f.Write(att.Bytes())
				check(err)

			}
		},
	}

	c.Flags().StringVarP(&attPath, "signature", "g", "attestation.intoto.jsonl", "Path to write the signed attestation")
	c.Flags().StringVarP(&subjects, "subjects", "s", "", "Formatted list of subjects of the form NAME:SHA256[|NAME:SHA256[|...]]")

	return c
}
