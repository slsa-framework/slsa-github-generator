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

// This file contains structs for the slsa provenance V1.0, and will be
// replaced once the format is finalized.

// TODO(#1191): Update to the final BuildType URI.
const DockerBasedBuildType = "https://slsa.dev/container-based-build/v0.1?draft"

// BuildDefinition contains the information required for building an artifact using a Docker image.
// Based on BuildDefinition in https://github.com/slsa-framework/slsa/pull/525.
type BuildDefinition struct {
	// BuildType indicates how to unambiguously interpret this BuildDefinition.
	BuildType string `json:"buildType"`

	// The set of top-level external inputs to the build. This SHOULD contain all
	// the information necessary and sufficient to initialize the build and begin
	// execution. "Top-level" means that it is not derived from another input.
	//
	// REQUIRED for SLSA Build L1.
	ExternalParameters ParameterCollection `json:"externalParameters"`

	// Parameters of the build environment that were provided by the `builder` and
	// not under external control. The primary intention of this field is for
	// debugging, incident response, and vulnerability management. The values here
	// MAY be necessary for reproducing the build.
	//
	// OPTIONAL.
	SystemParameters ParameterCollection `json:"systemParameters,omitempty"`

	// Resolved dependencies needed at build time.
	//
	// OPTIONAL.
	ResolvedDependencies []ArtifactReference `json:"resolvedDependencies,omitempty"`
}

// ParameterCollection is a collection of parameters that appear in a build definition.
type ParameterCollection struct {
	// References to the top-level, independent input artifacts to the build. In
	// many cases, this is a singular "source" artifact to be built.
	//
	// The key is a name whose interpretation depends on `buildType`. If there is
	// only one input, it SHOULD be named "source".
	Artifacts map[string]ArtifactReference `json:"artifacts,omitempty"`

	// Other parameters that are not artifact references. Like `artifacts`, the
	// key is a name whose interpretation depends on `buildType`.
	Values map[string]string `json:"values,omitempty"`
}

// ArtifactReference contains details about an artifact.
type ArtifactReference struct {
	// [URI] describing where this artifact came from. When possible, this SHOULD
	// be a universal and stable identifier, such as a source location or Package
	// URL ([purl]).
	//
	// Example: `pkg:pypi/pyyaml@6.0`
	//
	// REQUIRED.
	URI string `json:"uri"`

	// A map of cryptographic digests for the contents of this artifact.
	// The key indicates the cryptographic algorithm used for computing the digest.
	//
	// REQUIRED.
	Digest map[string]string `json:"digest"`

	// The name for this artifact local to the build.
	//
	// Example: `PyYAML-6.0.tar.gz`
	//
	// OPTIONAL.
	LocalName string `json:"localName,omitempty"`

	//nolint:lll
	// [URI] identifying the location that this artifact was downloaded from, if
	// different and not derivable from `uri`.
	//
	// Example: `https://files.pythonhosted.org/packages/36/2b/61d51a2c4f25ef062ae3f74576b01638bebad5e045f747ff12643df63844/PyYAML-6.0.tar.gz`
	//
	// OPTIONAL.
	DownloadLocation string `json:"downloadLocation,omitempty"`

	// [Media Type] (aka MIME type) of this artifact.
	//
	// OPTIONAL.
	MediaType string `json:"mediaType,omitempty"`
}
