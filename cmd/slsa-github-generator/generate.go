package main

import (
	"context"
	"io/ioutil"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// GenerateCmd returns the 'generate' command.
func generateCmd() *cobra.Command {
	var attPath string

	c := &cobra.Command{
		Use:   "generate",
		Short: "Generate SLSA provenance from a Github Action",
		Long: `Generate signed SLSA provenance from a Github Action and upload to a Rekor
transparency log. This command assumes that it is being run in the context of
a Github Actions workflow.`,

		Run: func(cmd *cobra.Command, args []string) {
			ghContext, err := github.GetWorkflowContext()
			check(err)

			p, err := slsa.HostedActionsProvenance(slsa.WorkflowRun{
				// TODO: Get subject names and digests
				Subjects:      []intoto.Subject{},
				BuildType:     "",
				BuildConfig:   nil,
				GithubContext: ghContext,
			})
			check(err)

			ctx := context.Background()

			s := sigstore.NewDefaultSigner()
			att, err := s.Sign(ctx, p)
			check(err)

			_, err = s.Upload(ctx, att)
			check(err)

			check(ioutil.WriteFile(attPath, att.Bytes(), 0600))
		},
	}

	// TODO: add flag for config file
	// TODO: add flag for getting subject names and digests
	c.Flags().StringVarP(&attPath, "output", "o", "attestation.intoto.jsonl", "Path to write the attestation")

	return c
}
