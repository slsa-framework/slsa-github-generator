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

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_LoadBuildConfigFromFile(t *testing.T) {
	got, err := loadBuildConfigFromFile("../testdata/config.toml")
	if err != nil {
		t.Fatalf("couldn't load config file: %v", err)
	}

	want := BuildConfig{
		Command:      []string{"cp", "internal/builders/docker/testdata/config.toml", "config.toml"},
		ArtifactPath: "config.toml",
	}

	if diff := cmp.Diff(*got, want); diff != "" {
		t.Errorf(diff)
	}
}

func Test_NewDockerBuildConfig(t *testing.T) {
	io := &InputOptions{
		BuildConfigPath: "testdata/config.toml",
		SourceRepo:      "https://github.com/project-oak/transparent-release",
		GitCommitHash:   "sha1:9b5f98310dbbad675834474fa68c37d880687cb9",
		BuilderImage:    "bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
	}
	got, err := NewDockerBuildConfig(io)
	if err != nil {
		t.Fatalf("invalid inputs: %v", err)
	}

	want := DockerBuildConfig{
		SourceRepo: io.SourceRepo,
		SourceDigest: Digest{
			Alg:   "sha1",
			Value: "9b5f98310dbbad675834474fa68c37d880687cb9",
		},
		BuilderImage: DockerImage{
			URI: "bash",
			Digest: Digest{
				Alg:   "sha256",
				Value: "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
			},
		},
		BuildConfigPath: io.BuildConfigPath,
	}

	if diff := cmp.Diff(*got, want); diff != "" {
		t.Errorf(diff)
	}
}
