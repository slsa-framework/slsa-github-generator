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
}

export interface DigestSet {
  [key: string]: string;
}

export interface ConfigSource {
  uri?: string;
  digest?: DigestSet;
  entryPoint?: string;
}

export interface Invocation {
  configSource?: ConfigSource;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  parameters?: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  environment?: any;
}

export interface Completeness {
  parameters?: boolean;
  environment?: boolean;
  materials?: boolean;
}

export interface Metadata {
  buildInvocationId?: string;
  buildStartedOn?: string;
  completeness?: Completeness;
  reproducible?: boolean;
}

export interface Material {
  uri: string;
  digest: DigestSet;
}

// SLSAPredicate is a SLSA v0.2 provenance predicate.
export interface SLSAPredicate {
  builder: Builder;
  buildType: string;
  invocation?: Invocation;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  buildConfig?: any;
  metadata: Metadata;
  materials?: Material[];
}
