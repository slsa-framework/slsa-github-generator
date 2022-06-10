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
	"fmt"
	"os"

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
		Use:   os.Args[0],
		Short: "SLSA provenance for Docker images on Github Actions",
		Long: `Build Docker images and generate SLSA provenance on Github Actions.
For more information on SLSA, visit https://slsa.dev`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("expected command")
		},
	}
	c.AddCommand(versionCmd())
	c.AddCommand(buildCmd())
	c.AddCommand(attestCmd())
	return c
}

func main() {
	check(rootCmd().Execute())
}
