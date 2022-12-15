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

// This file contains the structs and functionality for building artifacts
// using a builder Docker image, and user-provided build configurations.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

const (
	SourceKey       = "source"
	BuilderImageKey = "builderImage"
	ConfigFileKey   = "configFile"
)

// DockerBuild represents a state in the process of building the artifacts
// where the source repository is checked out and the config file is loaded and
// parsed, and we are ready for running the `docker run` command.
type DockerBuild struct {
	BuildDefinition *BuildDefinition
	BuildConfig     BuildConfig
	RepoInfo        RepoCheckoutInfo
}

// RepoCheckoutInfo contains info about the location of a locally checked out
// repository.
type RepoCheckoutInfo struct {
	// Path to the root of the repo
	RepoRoot string
}

// CreateBuildDefinition creates a BuildDefinition from the given DockerBuildConfig.
func CreateBuildDefinition(config *DockerBuildConfig) *BuildDefinition {
	artifacts := make(map[string]ArtifactReference)
	artifacts[SourceKey] = sourceArtifact(config)
	artifacts[BuilderImageKey] = builderImage(config)

	ep := ParameterCollection{
		Artifacts: artifacts,
		Values:    map[string]string{ConfigFileKey: config.BuildConfigPath},
	}

	// Currently we don't have any SystemParameters or ResolvedDependencies.
	// So these fields are left empty.
	return &BuildDefinition{
		BuildType:          DockerBasedBuildType,
		ExternalParameters: ep,
	}
}

// sourceArtifact returns the source repo and its digest as an instance of ArtifactReference.
func sourceArtifact(config *DockerBuildConfig) ArtifactReference {
	return ArtifactReference{
		URI:    config.SourceRepo,
		Digest: config.SourceDigest.ToMap(),
	}
}

// builderImage returns the builder image as an instance of ArtifactReference.
func builderImage(config *DockerBuildConfig) ArtifactReference {
	return ArtifactReference{
		URI:    config.BuilderImage.ToString(),
		Digest: config.BuilderImage.Digest.ToMap(),
	}
}

// SetUpBuildState sets up the build by checking out the source repository and
// loading the config file. It returns an instance of DockerBuild, or an error
// if setting up the build state fails.
func SetUpBuildState(config *DockerBuildConfig) (*DockerBuild, error) {
	// 1. Check out the repo, or verify that it is checked out.
	repoInfo, err := verifyOrFetchRepo(config)
	if err != nil {
		return nil, fmt.Errorf("couldn't verify or fetch source repo: %v", err)
	}

	// 2. Load and parse the config file.
	bc, err := config.LoadBuildConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("couldn't load config file from %q: %v", config.BuildConfigPath, err)
	}

	// 3. Check that the ArtifactPath pattern does not match any existing files,
	// so that we don't accidentally generate provenances for the wrong files.
	if err := checkExistingFiles(bc.ArtifactPath); err != nil {
		return nil, err
	}

	db := &DockerBuild{
		BuildDefinition: CreateBuildDefinition(config),
		BuildConfig:     *bc,
		RepoInfo:        *repoInfo,
	}
	return db, nil
}

// BuildArtifact builds the artifacts based on the user-provided inputs, and
// returns the names and SHA256 digests of the generated artifacts.
func (db *DockerBuild) BuildArtifact() ([]intoto.Subject, error) {
	// Get the current working directory. We will mount it as a Docker volume.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("couldn't get the current working directory: %v", err)
	}

	defaultDockerRunFlags := []string{
		// Mount the current working directory to workspace.
		fmt.Sprintf("--volume=%s:/workspace", cwd),
		"--workdir=/workspace",
		// Remove the container file system after the container exits.
		"--rm",
	}

	var args []string
	args = append(args, "run")
	args = append(args, defaultDockerRunFlags...)
	args = append(args, db.BuildDefinition.ExternalParameters.Artifacts[BuilderImageKey].URI)
	args = append(args, db.BuildConfig.Command...)
	cmd := exec.Command("docker", args...)

	outFileName, err := saveToTempFile(cmd.StdoutPipe)
	if err != nil {
		return nil, fmt.Errorf("couldn't save logs to file: %v", err)
	}

	errFileName, err := saveToTempFile(cmd.StderrPipe)
	if err != nil {
		return nil, fmt.Errorf("couldn't save error logs to file: %v", err)
	}

	log.Printf("Running command: %q.", cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("couldn't start the command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("failed to complete the command: %v; see %s for logs, and %s for errors",
			err, outFileName, errFileName)
	}

	return getArtifacts(db.BuildConfig.ArtifactPath)
}

// verifyOrFetchRepo checks that the current working directly is a Git repository
// at the expected Git commit hash; fetches the repo, if this is not the case.
func verifyOrFetchRepo(config *DockerBuildConfig) (*RepoCheckoutInfo, error) {
	if config.SourceDigest.Alg != "sha1" {
		return nil, fmt.Errorf("git commit digest must be a sha1 digest")
	}
	commitHash := config.SourceDigest.Value
	repoIsCheckedOut, err := verifyCommit(commitHash)
	if err != nil {
		return nil, err
	}
	if !repoIsCheckedOut {
		info, err := fetchSourcesFromGitRepo(config.SourceRepo, commitHash)
		if err != nil {
			return nil, fmt.Errorf("couldn't fetch sources from %q at commit %q: %v", config.SourceRepo, commitHash, err)
		}
		// Return info about the checked out repo for future cleanup.
		return info, nil
	}
	// Repo is checked out at the right commit; no future cleanup needed.
	return &RepoCheckoutInfo{}, nil
}

// verifyCommit checks that the current working directory is the root of a Git
// repository at the given commit hash. Returns an error if the working
// directory is a Git repository at a different commit.
func verifyCommit(commitHash string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	lastCommitIDBytes, err := cmd.Output()
	if err != nil {
		// The current working directory is not a git repo
		return false, nil
	}
	lastCommitID := strings.TrimSpace(string(lastCommitIDBytes))

	if lastCommitID != commitHash {
		return false, fmt.Errorf("the repo is already checked out at a different commit (%q)", lastCommitID)
	}

	return true, nil
}

// fetchSourcesFromGitRepo fetches a repo from the given URL into a temporary
// directory, and checks out the specified commit. Returns an instance of
// RepoCheckoutInfo containing the absolute path of the root of the repo.
func fetchSourcesFromGitRepo(repoURL, commitHash string) (*RepoCheckoutInfo, error) {
	// create a temp folder in the current directory for fetching the repo.
	targetDir, err := os.MkdirTemp("", "release-*")
	if err != nil {
		return nil, fmt.Errorf("couldn't create temp directory: %v", err)
	}
	log.Printf("Checking out the repo in %q.", targetDir)

	// Make targetDir and its parents, and cd to it.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("couldn't create directories at %q: %v", targetDir, err)
	}
	if err := os.Chdir(targetDir); err != nil {
		return nil, fmt.Errorf("couldn't change directory to %q: %v", targetDir, err)
	}

	// Clone the repo.
	err = cloneGitRepo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't clone the Git repo: %v", err)
	}

	// Change directory to the root of the cloned repo.
	repoName := path.Base(repoURL)
	if err := os.Chdir(repoName); err != nil {
		return nil, fmt.Errorf("couldn't change directory to %q: %v", repoName, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("couldn't get current working directory: %v", err)
	}

	// Checkout the commit.
	err = checkoutGitCommit(commitHash)
	if err != nil {
		return nil, fmt.Errorf("couldn't checkout the Git commit %q: %v", commitHash, err)
	}

	info := RepoCheckoutInfo{
		RepoRoot: cwd,
	}

	return &info, nil
}

func cloneGitRepo(repo string) error {
	cmd := exec.Command("git", "clone", repo)
	log.Printf("Cloning the repo from %s...", repo)

	outFileName, err := saveToTempFile(cmd.StdoutPipe)
	if err != nil {
		return fmt.Errorf("couldn't save logs to file: %v", err)
	}
	errFileName, err := saveToTempFile(cmd.StderrPipe)
	if err != nil {
		return fmt.Errorf("couldn't save errors to file: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("couldn't start the 'git clone' command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to complete the command: %v; see %q for logs, and %q for errors",
			err, outFileName, errFileName)
	}
	log.Printf("'git clone' completed. See %q, and %q for logs, and errors.", outFileName, errFileName)

	return nil
}

func checkoutGitCommit(commitHash string) error {
	cmd := exec.Command("git", "checkout", commitHash)

	outFileName, err := saveToTempFile(cmd.StdoutPipe)
	if err != nil {
		return fmt.Errorf("cannot save logs to file: %v", err)
	}
	errFileName, err := saveToTempFile(cmd.StderrPipe)
	if err != nil {
		return fmt.Errorf("cannot save errors to file: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("couldn't start the 'git checkout' command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to complete the command: %v; see %q for logs, and %q for errors",
			err, outFileName, errFileName)
	}

	return nil
}

// saveToTempFile creates a tempfile in `/tmp` and writes the content of the
// given reader to that file.
func saveToTempFile(pipe func() (io.ReadCloser, error)) (string, error) {
	reader, err := pipe()
	if err != nil {
		return "", fmt.Errorf("couldn't get a pipe: %v", err)
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	tmpfile, err := os.CreateTemp("", "log-*.txt")
	if err != nil {
		return "", fmt.Errorf("couldn't create tempfile: %v", err)
	}

	if _, err := tmpfile.Write(bytes); err != nil {
		tmpfile.Close()
		return "", fmt.Errorf("couldn't write bytes to tempfile: %v", err)
	}

	return tmpfile.Name(), nil
}

// Checks if any files match the given pattern, and returns an error if so.
func checkExistingFiles(pattern string) error {
	matches, err := filepath.Glob(pattern)
	// The only possible error is ErrBadPattern
	if err != nil {
		return fmt.Errorf("the pattern (%q) is malformed: %v", pattern, err)
	}

	if len(matches) == 0 {
		return nil
	}
	return fmt.Errorf("the specified pattern (%q) matches %d existing files; expected no matches", pattern, len(matches))
}

// Finds all files matching the given pattern, measures the SHA256 digest of
// each file, and returns filenames and digests as an array of intoto.Subject.
// Precondition: The pattern is a relative file path pattern.
func getArtifacts(pattern string) ([]intoto.Subject, error) {
	matches, err := filepath.Glob(pattern)
	// The only possible error is ErrBadPattern
	if err != nil {
		return nil, fmt.Errorf("the pattern (%q) is malformed: %v", pattern, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no files matching the pattern %q: %v", pattern, err)
	}

	var subjects []intoto.Subject
	for _, path := range matches {
		subject, err := toIntotoSubject(path)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, *subject)
	}

	return subjects, nil
}

// Reads the file in the given path and returns its name and digest wrapped in
// an intoto.Subject.
func toIntotoSubject(path string) (*intoto.Subject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %q: %v", path, err)
	}

	sum256 := sha256.Sum256(data)
	digest := hex.EncodeToString(sum256[:])
	name := filepath.Base(path)
	subject := &intoto.Subject{
		Name:   name,
		Digest: map[string]string{"sha256": digest},
	}
	return subject, nil
}

// Cleanup removes the generated temp files. But it might not be able to remove
// all the files, for instance the ones generated by the build script.
func (info *RepoCheckoutInfo) Cleanup() {
	// Some files are generated by the build toolchain (e.g., cargo), and cannot
	// be removed. We still want to remove all other files to avoid taking up
	// too much space, particularly when running locally.
	if len(info.RepoRoot) == 0 {
		return
	}
	if err := os.RemoveAll(info.RepoRoot); err != nil {
		log.Printf("failed to remove the temp files: %v", err)
	}
}
