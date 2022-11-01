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

// ciCmd runs the 'ci' command.
func ciCmd(check func(error)) *cobra.Command {
	var ciArguments []string
	var ciDirectory string

	c := &cobra.Command{
		Use:   "ci",
		Short: "Run ci command",
		Long:  `Run ci command to install dependencies of a node project.`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(ciArguments)
		},
	}

	c.Flags().StringArrayVarP(&ciArguments, "arguments", "a", []string{}, "Arguments to pass to ci command.")
	c.Flags().StringVarP(&ciDirectory, "directory", "d", "", "Working directory to issue commands.")

	return c
}
