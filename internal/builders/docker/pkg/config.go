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

// This file contains functionality and structs for validating and
// representing user inputs and configuration files.

import (
	"fmt"
	"net/url"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

// BuildConfig is a collection of parameters to use for building the artifact.
type BuildConfig struct {
	// The path, relative to the root of the git repository, where the artifact
	// built by the `docker run` command is expected to be found.
	ArtifactPath string `toml:"artifact_path"`

	// TODO(#1191): Add env and options if needed.
	// Command to pass to `docker run`. The command is taken as an array
	// instead of a single string to avoid unnecessary parsing. See
	// https://docs.docker.com/engine/reference/builder/#cmd and
	// https://man7.org/linux/man-pages/man3/exec.3.html for more details.
	Command []string `toml:"command"`
}

// Digest specifies a digest values, including the name of the hash function
// that was used for computing the digest.
type Digest struct {
	Alg   string
	Value string
}

// DockerImage fully specifies a docker image by a URI (e.g., including the
// docker image name and registry), and its digest.
type DockerImage struct {
	Name   string
	Digest Digest
}

// ToString returns the builder image in the form of NAME@ALG:VALUE.
func (bi *DockerImage) ToString() string {
	return fmt.Sprintf("%s@%s:%s", bi.Name, bi.Digest.Alg, bi.Digest.Value)
}

// DockerBuildConfig is a convenience class for holding validated user inputs.
type DockerBuildConfig struct {
	SourceRepo      string
	SourceDigest    Digest
	BuilderImage    DockerImage
	BuildConfigPath string
	ForceCheckout   bool
	Verbose         bool
}

// NewDockerBuildConfig validates the inputs and generates an instance of
// DockerBuildConfig.
func NewDockerBuildConfig(io *InputOptions) (*DockerBuildConfig, error) {
	if err := validateURI(io.SourceRepo); err != nil {
		return nil, err
	}

	sourceRepoDigest, err := validateDigest(io.GitCommitHash)
	if err != nil {
		return nil, err
	}

	dockerImage, err := validateDockerImage(io.BuilderImage)
	if err != nil {
		return nil, err
	}

	if err = utils.PathIsUnderCurrentDirectory(io.BuildConfigPath); err != nil {
		return nil, fmt.Errorf("invalid build config path: %v", err)
	}

	return &DockerBuildConfig{
		SourceRepo:      io.SourceRepo,
		SourceDigest:    *sourceRepoDigest,
		BuilderImage:    *dockerImage,
		BuildConfigPath: io.BuildConfigPath,
		ForceCheckout:   io.ForceCheckout,
		Verbose:         io.Verbose,
	}, nil
}

func validateURI(input string) error {
	_, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("could not parse string (%q) as URI: %v", input, err)
	}
	return nil
}

func validateDigest(input string) (*Digest, error) {
	// We expect the input to be of the form ALG:VALUE
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("got %s, want ALG:VALUE format", input)
	}
	digest := Digest{
		Alg:   parts[0],
		Value: parts[1],
	}
	return &digest, nil
}

func validateDockerImage(image string) (*DockerImage, error) {
	imageParts := strings.Split(image, "@")
	if len(imageParts) != 2 {
		return nil, fmt.Errorf("got %s, want NAME@DIGEST format", image)
	}

	if err := validateURI(imageParts[0]); err != nil {
		return nil, fmt.Errorf("docker image name (%q) is not a valid URI: %v", imageParts[0], err)
	}

	digest, err := validateDigest(imageParts[1])
	if err != nil {
		return nil, fmt.Errorf("docker image digest (%q) is malformed: %v", imageParts[1], err)
	}

	dockerImage := DockerImage{
		Name:   imageParts[0],
		Digest: *digest,
	}

	return &dockerImage, nil
}

// ToMap returns this instance as a mapping between the algorithm and value.
func (d *Digest) ToMap() map[string]string {
	return map[string]string{d.Alg: d.Value}
}

// LoadBuildConfigFromFile loads build configuration from a toml file specified
// by the BuildConfigPath of this DockerBuildConfig. An instance of BuildConfig
// is returned on success.
func (dbc *DockerBuildConfig) LoadBuildConfigFromFile() (*BuildConfig, error) {
	return loadBuildConfigFromFile(dbc.BuildConfigPath)
}

// loadBuildConfigFromFile does not validate the input path, and is therefore
// not exposed. The corresponding method LoadBuildConfigFromFile must be called
// on an instance of DockerBuildConfig which has a validated BuildConfigPath.
func loadBuildConfigFromFile(path string) (*BuildConfig, error) {
	tomlBytes, err := utils.SafeReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't load toml file: %v", err)
	}
	tomlTree, err := toml.LoadBytes(tomlBytes)
	if err != nil {
		return nil, fmt.Errorf("couldn't create toml tree: %v", err)
	}

	config := BuildConfig{}
	if err := tomlTree.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal toml file: %v", err)
	}

	return &config, nil
}
