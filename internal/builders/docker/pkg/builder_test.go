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
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	intoto "github.com/in-toto/in-toto-golang/in_toto"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

func Test_CreateBuildDefinition(t *testing.T) {
	config := &DockerBuildConfig{
		SourceRepo:   "git+https://github.com/slsa-framework/slsa-github-generator@refs/heads/main",
		SourceDigest: Digest{Alg: "sha1", Value: "cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00"},
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

func Test_GitClient_sourceWithRef(t *testing.T) {
	// This tests that specifying a source repository with a ref resolves
	// to a valid GitClient source repository

	config := &DockerBuildConfig{
		// Use a small repo for test
		SourceRepo: "git+https://github.com/project-oak/transparent-release@refs/heads/main",
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

	// We expect that the sourceRef is set to refs/heads/main
	if gc.sourceRef == nil || *gc.sourceRef != "refs/heads/main" {
		t.Errorf("expected sourceRef to be refs/heads/main, got %s", *gc.sourceRef)
	}
}

func Test_GitClient_invalidSourceRef(t *testing.T) {
	// This tests that specifying a source repository with a ref resolves
	// to a valid GitClient source repository

	config := &DockerBuildConfig{
		// Use a small repo for test
		SourceRepo: "git+https://github.com/project-oak/transparent-release@refs/heads/main@invalid",
		// The digest value does not matter for the test
		SourceDigest:    Digest{Alg: "sha1", Value: "does-no-matter"},
		BuildConfigPath: "internal/builders/docker/testdata/config.toml",
		ForceCheckout:   false,
		// BuilderImage field is not relevant, so it is omitted
	}
	_, err := newGitClient(config, 1)
	if err == nil {
		t.Fatalf("expected error creating GitClient")
	}
	if !strings.Contains(err.Error(), "invalid source repository format") {
		t.Fatalf("expected invalid source ref error creating GitClient, got %s", err)
	}
}

func Test_inspectArtifacts(t *testing.T) {
	// Note: If the files in ../testdata/ change, this test must be updated.
	pattern := "../testdata/*"
	out := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	got, err := inspectAndWriteArtifacts(pattern, out, filepath.Dir(wd))
	if err != nil {
		t.Fatalf("failed to inspect artifacts: %v", err)
	}

	s1 := intoto.Subject{
		Name:   "build-definition.json",
		Digest: map[string]string{"sha256": "b1c74863007166aadca8ff54a0e647047696bee38e8e8a25a1290f494e3abc46"},
	}
	s2 := intoto.Subject{
		Name:   "config.toml",
		Digest: map[string]string{"sha256": "975a0582b8c9607f3f20a6b8cfef01b25823e68c5c3658e6e1ccaaced2a3255d"},
	}

	s3 := intoto.Subject{
		Name:   "slsa1-provenance.json",
		Digest: map[string]string{"sha256": "f472aaf04468ae881ab502f1f02f23476fe0d4dbb7a8a4b5d3eae9b2843e2ecd"},
	}

	s4 := intoto.Subject{
		Name:   "wildcard-config.toml",
		Digest: map[string]string{"sha256": "d9b8670f1b9616db95b0dc84cbc68062c691ef31bb9240d82753de0739c59194"},
	}

	want := []intoto.Subject{s1, s2, s3, s4}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

// When running in the checkout of the repository root, root == "".
func Test_inspectArtifactsNoRoot(t *testing.T) {
	// Note: If the files in ../testdata/ change, this test must be updated.
	pattern := "testdata/*"
	out := t.TempDir()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}

	got, err := inspectAndWriteArtifacts(pattern, out, "")
	if err != nil {
		t.Fatalf("failed to inspect artifacts: %v", err)
	}

	s1 := intoto.Subject{
		Name:   "build-definition.json",
		Digest: map[string]string{"sha256": "b1c74863007166aadca8ff54a0e647047696bee38e8e8a25a1290f494e3abc46"},
	}
	s2 := intoto.Subject{
		Name:   "config.toml",
		Digest: map[string]string{"sha256": "975a0582b8c9607f3f20a6b8cfef01b25823e68c5c3658e6e1ccaaced2a3255d"},
	}

	s3 := intoto.Subject{
		Name:   "slsa1-provenance.json",
		Digest: map[string]string{"sha256": "f472aaf04468ae881ab502f1f02f23476fe0d4dbb7a8a4b5d3eae9b2843e2ecd"},
	}

	s4 := intoto.Subject{
		Name:   "wildcard-config.toml",
		Digest: map[string]string{"sha256": "d9b8670f1b9616db95b0dc84cbc68062c691ef31bb9240d82753de0739c59194"},
	}

	want := []intoto.Subject{s1, s2, s3, s4}

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
		SourceDigest: Digest{Alg: "sha1", Value: "cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00"},
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

func Test_ParseProvenance(t *testing.T) {
	provenance := loadProvenance(t)
	got := &provenance.Predicate.BuildDefinition

	want, err := loadBuildDefinitionFromFile("../testdata/build-definition.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func Test_ProvenanceStatementSLSA1_ToDockerBuildConfig(t *testing.T) {
	provenance := loadProvenance(t)
	got, err := provenance.ToDockerBuildConfig(true)
	if err != nil {
		t.Fatalf("%v", err)
	}

	want := &DockerBuildConfig{
		SourceRepo: "git+https://github.com/slsa-framework/slsa-github-generator@refs/heads/main",
		SourceDigest: Digest{
			Alg:   "sha1",
			Value: "cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00",
		},
		BuilderImage: DockerImage{
			Name: "bash",
			Digest: Digest{
				Alg:   "sha256",
				Value: "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
			},
		},
		BuildConfigPath: "internal/builders/docker/testdata/config.toml",
		ForceCheckout:   true,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func loadProvenance(t *testing.T) ProvenanceStatementSLSA1 {
	bytes, err := os.ReadFile("../testdata/slsa1-provenance.json")
	if err != nil {
		t.Fatalf("Reading the provenance file: %v", err)
	}

	provenance, err := ParseProvenance(bytes)
	if err != nil {
		t.Fatalf("Parsing the provenance file: %v", err)
	}
	return *provenance
}

func checkError[T error](t *testing.T, got error, want T) {
	if !errors.As(got, &want) {
		t.Errorf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}
