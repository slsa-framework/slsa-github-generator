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
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	slsa1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1.0"
)

func Test_BuildDefinition(t *testing.T) {
	path := "../testdata/build-definition.json"
	got, err := loadBuildDefinitionFromFile(path)
	if err != nil {
		t.Fatalf("%v", err)
	}

	wantSource := slsa1.ArtifactReference{
		URI:    "git+https://github.com/project-oak/transparent-release",
		Digest: map[string]string{"sha1": "9b5f98310dbbad675834474fa68c37d880687cb9"},
	}

	wantBuilderImage := slsa1.ArtifactReference{
		URI:    "bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
		Digest: map[string]string{"sha256": "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"},
	}

	want := &slsa1.ProvenanceBuildDefinition{
		BuildType: "https://slsa.dev/container-based-build/v0.1?draft",
		ExternalParameters: DockerBasedExternalParmaters{
			Source:       wantSource,
			BuilderImage: wantBuilderImage,
			ConfigPath:   "internal/builders/docker/testdata/config.toml",
			Config: BuildConfig{
				ArtifactPath: "config.toml",
				Command: []string{
					"cp",
					"internal/builders/docker/testdata/config.toml",
					"config.toml",
				},
			},
		},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func loadBuildDefinitionFromFile(path string) (*slsa1.ProvenanceBuildDefinition, error) {
	bdBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the JSON file in %q: %w", path, err)
	}

	var bd slsa1.ProvenanceBuildDefinition
	if err := json.Unmarshal(bdBytes, &bd); err != nil {
		return nil, fmt.Errorf("could not unmarshal the JSON file in %q as a BuildDefinition: %w", path, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bdBytes, &result); err != nil {
		return nil, fmt.Errorf("could not unmarshal the JSON file in %q as a map[string]interface{}: %w", path, err)
	}

	ep, ok := result["externalParameters"]
	if !ok {
		return nil, fmt.Errorf("missing externalParameters in BuildDefinition")
	}

	epBytes, err := json.Marshal(ep)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the external params in %q: %w", path, err)
	}

	var dockerEp DockerBasedExternalParmaters
	if err := json.Unmarshal(epBytes, &dockerEp); err != nil {
		return nil, fmt.Errorf("could not unmarshal the JSON file in %q as a DockerBasedExternalParameters: %w", path, err)
	}

	bd.ExternalParameters = dockerEp

	return &bd, nil
}
