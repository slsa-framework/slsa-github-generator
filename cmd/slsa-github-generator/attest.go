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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsav02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

var (
	// shaCheck verifies a hash is has only hexidecimal digits and is 64
	// characters long.
	shaCheck = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

	// wsSplit is used to split lines in the subjects input.
	wsSplit = regexp.MustCompile(`[\t ]`)
)

// parseSubjects parses the value given to the subjects option.
func parseSubjects(subjectsStr string) ([]intoto.Subject, error) {
	var parsed []intoto.Subject

	scanner := bufio.NewScanner(strings.NewReader(subjectsStr))
	for scanner.Scan() {
		// Split by whitespace, and get values.
		parts := wsSplit.Split(strings.TrimSpace(scanner.Text()), 2)

		// Lowercase the sha digest to comply with the SLSA spec.
		shaDigest := strings.ToLower(strings.TrimSpace(parts[0]))
		if shaDigest == "" {
			// Ignore empty lines.
			continue
		}
		// Do a sanity check on the SHA to make sure it's a proper hex digest.
		if !shaCheck.MatchString(shaDigest) {
			return nil, fmt.Errorf("unexpected sha256 hash %q", shaDigest)
		}

		// Check for the subject name.
		if len(parts) == 1 {
			return nil, fmt.Errorf("expected subject name for hash %q", shaDigest)
		}
		name := strings.TrimSpace(parts[1])

		for _, p := range parsed {
			if p.Name == name {
				return nil, fmt.Errorf("duplicate subject: %q", name)
			}
		}

		parsed = append(parsed, intoto.Subject{
			Name: name,
			Digest: slsav02.DigestSet{
				"sha256": shaDigest,
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
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

			ctx := context.Background()

			c, err := github.NewOIDCClient()
			check(err)

			p, err := slsa.HostedActionsProvenance(ctx, slsa.NewWorkflowRun(parsedSubjects, ghContext), c)
			check(err)

			if attPath != "" {
				s := sigstore.NewDefaultSigner()
				att, err := s.Sign(ctx, p)
				check(err)

				check(s.Upload(ctx, att))

				f, err := getFile(attPath)
				check(err)

				_, err = f.Write(att.Bytes())
				check(err)
			}
		},
	}

	c.Flags().StringVarP(&attPath, "signature", "g", "attestation.intoto.jsonl", "Path to write the signed attestation")
	c.Flags().StringVarP(&subjects, "subjects", "s", "", "Formatted list of subjects in the same format as sha256sum")

	return c
}
