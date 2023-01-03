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
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_BuildDefinition(t *testing.T) {
	path := "../testdata/build-definition.json"
	bdBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Could not read the JSON file in %q:\n%v", path, err)
	}

	var got BuildDefinition
	if err := json.Unmarshal(bdBytes, &got); err != nil {
		t.Fatalf("Could not unmarshal the JSON file in %q as a BuildDefinition:\n%v", path, err)
	}

	wantSource := ArtifactReference{
		URI:    "git+https://github.com/project-oak/transparent-release",
		Digest: map[string]string{"sha1": "9b5f98310dbbad675834474fa68c37d880687cb9"},
	}

	wantBuilderImage := ArtifactReference{
		URI:    "bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
		Digest: map[string]string{"sha256": "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"},
	}

	want := BuildDefinition{
		BuildType: "https://slsa.dev/container-based-build/v0.1?draft",
		ExternalParameters: ParameterCollection{
			Artifacts: map[string]ArtifactReference{"source": wantSource, "builderImage": wantBuilderImage},
			Values:    map[string]string{"configFile": "internal/builders/docker/testdata/config.toml"},
		},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}
