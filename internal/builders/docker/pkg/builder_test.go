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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	intoto "github.com/in-toto/in-toto-golang/in_toto"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

func Test_CreateBuildDefinition(t *testing.T) {
	config := &DockerBuildConfig{
		SourceRepo:   "git+https://github.com/project-oak/transparent-release",
		SourceDigest: Digest{Alg: "sha1", Value: "9b5f98310dbbad675834474fa68c37d880687cb9"},
		BuilderImage: DockerImage{
			Name:   "bash",
			Digest: Digest{Alg: "sha256", Value: "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"},
		},
		BuildConfigPath: "internal/builders/docker/testdata/config.toml",
	}

	db := &DockerBuild{
		config: config,
		buildConfig: &BuildConfig{
			Command:      []string{"cp", "internal/builders/docker/testdata/config.toml", "config.toml"},
			ArtifactPath: "config.toml",
		},
	}

	got := db.CreateBuildDefinition()

	want, err := loadBuildDefinitionFromFile("../testdata/build-definition.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func Test_GitClient_verifyOrFetchRepo(t *testing.T) {
	config := &DockerBuildConfig{
		// Use a small repo for test
		SourceRepo: "git+https://github.com/project-oak/transparent-release",
		// The digest value does not matter for the test
		SourceDigest:    Digest{Alg: "sha1", Value: "does-not-matter"},
		BuildConfigPath: "internal/builders/docker/testdata/config.toml",
		ForceCheckout:   false,
		// BuilderImage field is not relevant, so it is omitted
	}
	gc, err := newGitClient(config, 1)
	if err != nil {
		t.Fatalf("Could create GitClient: %v", err)
	}

	// We expect it to fail at verifyCommit
	want := &errGitCommitMismatch{}
	err = gc.verifyOrFetchRepo()
	checkError(t, err, want)
}

func Test_GitClient_fetchSourcesFromGitRepo(t *testing.T) {
	// The call to fetchSourcesFromGitRepo will change directory. Here we store
	// the current working directory, and change back to it when the test ends.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("couldn't get current working directory: %v", err)
	}

	config := &DockerBuildConfig{
		// Use a small repo for test
		SourceRepo: "git+https://github.com/project-oak/transparent-release",
		// The digest value does not matter for the test
		SourceDigest:    Digest{Alg: "sha1", Value: "does-no-matter"},
		BuildConfigPath: "internal/builders/docker/testdata/config.toml",
		ForceCheckout:   false,
		// BuilderImage field is not relevant, so it is omitted
	}
	gc, err := newGitClient(config, 1)
	if err != nil {
		t.Fatalf("Could not create GitClient: %v", err)
	}

	// We expect the checkout to fail
	want := &errGitCheckout{}
	err = gc.fetchSourcesFromGitRepo()
	checkError(t, err, want)

	// Cleanup
	gc.cleanupAllFiles()
	// Recover the original test state.
	if err := os.Chdir(cwd); err != nil {
		t.Errorf("couldn't change directory to %q: %v", cwd, err)
	}
}

func Test_inspectArtifacts(t *testing.T) {
	// Note: If the files in ../testdata/ change, this test must be updated.
	pattern := "../testdata/*"
	out := t.TempDir()
	got, err := inspectAndWriteArtifacts(pattern, out, "..")
	if err != nil {
		t.Fatalf("failed to inspect artifacts: %v", err)
	}

	s1 := intoto.Subject{
		Name:   "build-definition.json",
		Digest: map[string]string{"sha256": "1a60da949ad34d060ac2650bc4d7cba287cb7d2ffda4e8f4a65459c77801e2d5"},
	}
	s2 := intoto.Subject{
		Name:   "config.toml",
		Digest: map[string]string{"sha256": "975a0582b8c9607f3f20a6b8cfef01b25823e68c5c3658e6e1ccaaced2a3255d"},
	}
	s3 := intoto.Subject{
		Name:   "wildcard-config.toml",
		Digest: map[string]string{"sha256": "d9b8670f1b9616db95b0dc84cbc68062c691ef31bb9240d82753de0739c59194"},
	}

	want := []intoto.Subject{s1, s2, s3}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

type testFetcher struct{}

func (testFetcher) Fetch() (*RepoCheckoutInfo, error) {
	return &RepoCheckoutInfo{}, nil
}

func Test_Builder_SetUpBuildState(t *testing.T) {
	config := DockerBuildConfig{
		SourceRepo:   "git+https://github.com/project-oak/transparent-release",
		SourceDigest: Digest{Alg: "sha1", Value: "9b5f98310dbbad675834474fa68c37d880687cb9"},
		BuilderImage: DockerImage{
			Name:   "bash",
			Digest: Digest{Alg: "sha256", Value: "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"},
		},
		BuildConfigPath: "../testdata/config.toml",
	}

	f := testFetcher{}
	b := Builder{
		repoFetcher: f,
		config:      config,
	}

	db, err := b.SetUpBuildState()
	if err != nil {
		t.Fatalf("couldn't set up build state: %v", err)
	}
	if db == nil {
		t.Error("db is null")
	}
}

func checkError[T error](t *testing.T, got error, want T) {
	if !errors.As(got, &want) {
		t.Errorf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}
