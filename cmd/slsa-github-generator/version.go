package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/slsa-framework/slsa-github-generator/version"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}
}
