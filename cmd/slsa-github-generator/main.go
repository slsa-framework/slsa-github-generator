package main

import (
	"errors"
	"fmt"
	"os"

	// TODO: Allow use of other OIDC providers?
	// Enable the github OIDC auth provider.
	_ "github.com/sigstore/cosign/pkg/providers/github"

	"github.com/spf13/cobra"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "slsa-github-generator",
		Short: "Generate SLSA provenance for Github Actions",
		Long: `Generate SLSA provenance for Github Actions.
For more information on SLSA, visit https://slsa.dev`,
		Run: func(cmd *cobra.Command, args []string) {
			check(errors.New("expected command"))
		},
	}
	c.AddCommand(versionCmd())
	c.AddCommand(generateCmd())
	return c
}

func main() {
	check(rootCmd().Execute())
}
