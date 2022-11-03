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
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd runs the 'run' command.
func runCmd(check func(error)) *cobra.Command {
	var runScripts []string
	var runDirectory string

	c := &cobra.Command{
		Use:   "run",
		Short: "Run scripts",
		Long:  `Run scripts. Example: ./binary run --scripts="script1, script2"`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(runScripts)
		},
	}

	c.Flags().StringSliceVarP(&runScripts, "scripts", "s", []string{}, "List of scripts to run.")
	c.Flags().StringVarP(&runDirectory, "directory", "d", "", "Working directory to issue commands.")
	return c
}
