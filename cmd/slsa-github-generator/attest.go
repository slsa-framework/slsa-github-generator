package main

import (
	"context"
	"errors"
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

func parseSubjects(subjects []string) ([]intoto.Subject, error) {
	var parsed []intoto.Subject
	for _, s := range subjects {
		subject := intoto.Subject{}
		parts := strings.SplitN(s, "@", 2)
		if len(parts) == 0 {
			return nil, errors.New("missing subject name")
		}

		subject.Name = parts[0]
		if len(parts) > 1 {
			subject.Digest = slsav02.DigestSet{
				"sha256": parts[1],
			}
		}
		parsed = append(parsed, subject)
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
	var subjects []string

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

			p, err := slsa.HostedActionsProvenance(slsa.WorkflowRun{
				Subjects:      parsedSubjects,
				BuildType:     "",
				BuildConfig:   nil,
				GithubContext: ghContext,
			})
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

	// TODO: add flag for config file
	c.Flags().StringVarP(&attPath, "signature", "g", "attestation.intoto.jsonl", "Path to write the signed attestation")
	c.Flags().StringSliceVarP(&subjects, "subject", "j", nil, "Subject of the form NAME[@SHA256HEX]")

	return c
}
