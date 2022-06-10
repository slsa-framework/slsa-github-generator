// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

// BuildOpts are options for an image build.
type BuildOpts struct {
	ContextDir string
	File       string
	Tags       []string
}

// Build builds a docker image.
func Build(opts BuildOpts) error {
	contextDir := opts.ContextDir
	if contextDir == "" {
		contextDir = "."
	}

	cmd, err := New("build", contextDir)
	if err != nil {
		return err
	}

	for _, tag := range opts.Tags {
		cmd.WithFlag("tag", tag)
	}

	if opts.File != "" {
		cmd.WithFlag("file", opts.File)
	}

	return cmd.Run()
}
