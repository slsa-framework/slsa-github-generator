/*
Copyright 2022 SLSA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WIHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

export interface Builder {
  id: string;
  version?: string;
  builderDependencies?: ResourceDescriptor[];
}

/**
 * DigestSet implements an in-toto version v1.0 DigestSet.
 *
 * See: https://github.com/in-toto/attestation/blob/main/spec/v1/digest_set.md
 */
export interface DigestSet {
  [key: string]: string;
}

export interface Metadata {
  invocationId?: string;
  startedOn?: Date;
  finishedOn?: Date;
}

/**
 * ResourceDescriptor implements an in-toto version v1.0 ResourceDescriptor.
 *
 * See: https://github.com/in-toto/attestation/blob/main/spec/v1/resource_descriptor.md
 */
export interface ResourceDescriptor {
  name?: string;
  uri?: string;
  digest?: DigestSet;
  content?: Uint8Array; // content is bytes.
  downloadLocation?: string;
  mediaType?: string;
  annotations?: [key: string]: string;
}

export interface BuildDefinition {
  // buildType is a TypeURI that unambiguously indicates the type of this message and how to initiate the build.
  buildType: string;

  // externalParameters is the set of top-level external inputs to the build.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  externalParameters: any;

  // internalParameters describes parameters of the build environment provided by the `builder`.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  internalParameters?: any;

  // resolvedDependencies are dependencies needed at build time.
  resolvedDependencies?: ResourceDescriptor[];
}

export interface RunDetails {
  builder: Builder;

  metadata: Metadata;

  byproducts?: ResourceDescriptor[];
}

// SLSAPredicate is a SLSA v1.0-rc1 provenance predicate.
export interface SLSAPredicate {
  // buildDefinition describes the inputs to the build.
  buildDefinition: BuildDefinition;

  // runDetails includes details specific to this particular execution of the build.
  runDetails: RunDetails;
}
