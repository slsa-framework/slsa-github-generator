package main

import (
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// provenanceCmd returns the 'provenance' command.
func provenanceCmd() *cobra.Command {
	var provPath string
	var subjects string

	c := &cobra.Command{
		Use:   "provenance",
		Short: "Generate SLSA provenance for a Github Action",
		Long: `Generate SLSA provenance from a Github Action and save to a file. This command
assumes that it is being run in the context of a Github Actions workflow.`,

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
				BuildType:     provenanceOnlyBuildType,
				BuildConfig:   nil,
				GithubContext: ghContext,
			})
			check(err)

			b, err := json.Marshal(p)
			check(err)

			f, err := getFile(provPath)
			check(err)

			_, err = f.Write(b)
			check(err)
		},
	}

	c.Flags().StringVarP(&provPath, "provenance", "p", "provenance.intoto.json", "path to output the SLSA provenance JSON")
	c.Flags().StringVarP(&subjects, "subjects", "s", "", "Comma separated list of subjects of the form NAME@SHA256")

	return c
}
