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

package pkg

import "github.com/spf13/cobra"

// InputOptions are the common options for the dry run and build command.
type InputOptions struct {
	BuildConfigPath string
	SourceRepo      string
	GitCommitHash   string
	BuilderImage    string
	ForceCheckout   bool
	Verbose         bool
}

// AddFlags adds input flags to the given command.
func (io *InputOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&io.BuildConfigPath, "build-config-path", "c", "",
		"Required - Path to a toml file containing the build configs.")

	cmd.Flags().StringVarP(&io.SourceRepo, "source-repo", "s", "",
		"Required - URL of the source repo.")

	cmd.Flags().StringVarP(&io.GitCommitHash, "git-commit-digest", "d", "",
		"Required - SHA1 Git commit digest of the revision of the source code to build the artefact from.")

	cmd.Flags().StringVarP(&io.BuilderImage, "builder-image", "i", "",
		"Required - URL indicating the Docker builder image, including a URI and image digest.")

	cmd.Flags().BoolVarP(&io.ForceCheckout, "force-checkout", "f", false,
		"Optional - Forces checking out the source code from the given Git repo.")

	cmd.Flags().BoolVarP(&io.Verbose, "verbose", "v", false,
		"Optional - Prints all logs and errors in console.")
}
