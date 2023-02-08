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
	slsa1 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1.0"
)

// This file contains structs for the slsa provenance V1.0, and will be
// replaced once the format is finalized.

const (
	// DockerBasedBuildType is type for container-based builds.
	// TODO(#1191): Update to the final BuildType URI.
	DockerBasedBuildType = "https://slsa.dev/container-based-build/v0.1?draft"
	// SourceKey is the lookup key for source repository in ExternalParameters.
	SourceKey = "source"
	// BuilderImageKey is the lookup key for builder image in ExternalParameters.
	BuilderImageKey = "builderImage"
	// ConfigFileKey is the lookup key for the config file in ExternalParameters.
	ConfigFileKey = "configFile"
	// ArtifactPathKey is the lookup key for the artifacts path in ExternalParameters.
	ArtifactPathKey = "artifactPath"
	// CommandKey is the lookup key for the command in ExternalParameters.
	CommandKey = "command"
)

// DockerBasedExternalParmaters is a representation of the top level inputs to a
// docker-based build.
type DockerBasedExternalParmaters struct {
	// The source GitHub repo
	Source slsa1.ArtifactReference `json:"source"`

	// The Docker builder image
	BuilderImage slsa1.ArtifactReference `json:"builderImage"`

	// Path to a configuration file relative to the root of the repository.
	ConfigPath string `json:"configPath"`

	// Unpacked build config parameters
	Config BuildConfig `json:"buildConfig"`
}
