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

func checkExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slsa-docker-based-generator [COMMAND] [FLAGS]",
		Short: "Generate SLSA provenance for artifacts built using a docker builder image",
		Long: `Generate SLSA provenance for artifacts built using a docker builder image.
For more information on SLSA, visit https://slsa.dev`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("expecting a command")
		},
	}
	cmd.AddCommand(DryRunCmd(checkExit))
	cmd.AddCommand(BuildCmd(checkExit))
	cmd.AddCommand(VerifyCmd(checkExit))
	return cmd
}

func main() {
	checkExit(rootCmd().Execute())
}
