/*
Copyright 2023 SLSA Authors
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

import * as process from "process";
import * as fs from "fs";

const DELEGATOR_BUILD_TYPE =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

interface Builder {
  id: string;
  version?: string;
  builderDependencies?: ArtifactReference[];
}

interface DigestSet {
  [key: string]: string;
}

interface Metadata {
  invocationId?: string;
  startedOn?: Date;
  finishedOn?: Date;
}

interface ParameterValue_Artifact {
  artifact: ArtifactReference;
}

interface ParameterValue_String {
  value: string | undefined;
}

type ParameterValue = ParameterValue_Artifact | ParameterValue_String;

interface ArtifactReference {
  uri: string;
  digest: DigestSet;
  localName?: string;
  downloadLocation?: string;
  mediaType?: string;
}

interface BuildDefinition {
  // URI indicating the type of build.
  buildType: string;

  // external
  externalParameters: { [key: string]: ParameterValue };

  systemParameters?: { [key: string]: ParameterValue };

  resolvedDependencies?: ArtifactReference[];
}

interface RunDetails {
  builder: Builder;

  metadata: Metadata;

  byproducts?: ArtifactReference[];
}

interface SLSAv1Predicate {
  buildDefinition: BuildDefinition;

  runDetails: RunDetails;
}

export interface rawTokenInterface {
  version: number;
  context: string;
  builder: {
    private_repository: boolean;
    runner_label: string;
    audience: string;
  };
  github: githubObj;
  tool: {
    actions: {
      build_artifacts: {
        path: string;
      };
    };
    // NOTE: reusable workflows only support inputs of type
    // boolean, number, or string.
    // https://docs.github.com/en/actions/using-workflows/reusing-workflows#passing-inputs-and-secrets-to-a-reusable-workflow.
    inputs: Map<string, Object>;
  };
}

export interface githubObj {
  event_name: string;
  run_attempt: string;
  run_id: string;
  run_number: string;
  workflow: string;
  sha: string;
  repository: string;
  repository_owner: string;
  // TODO(#1411): Record if these become available.
  // repository_id: string;
  // repository_owner_id: string;
  // actor_id: string;
  ref: string;
  ref_type: string;
  actor: string;
}

export function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string,
  workflow_name: string
): SLSAv1Predicate {
  const { env } = process;

  const callerRepo: string = createURI(
    env.GITHUB_REPOSITORY || "",
    env.GITHUB_REF || ""
  );

  const predicate: SLSAv1Predicate = {
    buildDefinition: {
      buildType: DELEGATOR_BUILD_TYPE,
      externalParameters: {
        // This is the v0.2 entryPoint.
        // TODO: Get path via GitHub API via runID and repository
        workflow: { value: workflow_name },
        // We use source here because the source contained the source
        // repository and the build configuration.
        source: {
          artifact: {
            uri: callerRepo,
            digest: {
              sha1: env.GITHUB_SHA || "",
            },
          },
        },
      },
      systemParameters: {
        // environment
        // TODO: Add GitHub event payload.
        GITHUB_ACTOR_ID: { value: env.GITHUB_ACTOR_ID },
        GITHUB_EVENT_NAME: { value: env.GITHUB_EVENT_NAME },
        GITHUB_JOB: { value: env.GITHUB_JOB },
        GITHUB_REF: { value: env.GITHUB_REF },
        GITHUB_REF_TYPE: { value: env.GITHUB_REF_TYPE },
        GITHUB_REPOSITORY: { value: env.GITHUB_REPOSITORY },
        GITHUB_REPOSITORY_ID: { value: env.GITHUB_REPOSITORY_ID },
        GITHUB_REPOSITORY_OWNER_ID: { value: env.GITHUB_REPOSITORY_OWNER_ID },
        GITHUB_RUN_ATTEMPT: { value: env.GITHUB_RUN_ATTEMPT },
        GITHUB_RUN_ID: { value: env.GITHUB_RUN_ID },
        GITHUB_RUN_NUMBER: { value: env.GITHUB_RUN_NUMBER },
        GITHUB_SHA: { value: env.GITHUB_SHA },
        GITHUB_WORKFLOW: { value: env.GITHUB_WORKFLOW },
        GITHUB_WORKFLOW_REF: { value: env.GITHUB_WORKFLOW_REF },
        GITHUB_WORKFLOW_SHA: { value: env.GITHUB_WORKFLOW_SHA },
        IMAGE_OS: { value: env.ImageOS },
        IMAGE_VERSION: { value: env.ImageVersion },
        RUNNER_ARCH: { value: env.RUNNER_ARCH },
        RUNNER_NAME: { value: env.RUNNER_NAME },
        RUNNER_OS: { value: env.RUNNER_OS },
      },
    },
    runDetails: {
      // TODO: Where do the raw token inputs (the tool inputs) go?
      builder: {
        id: toolURI,
      },
      metadata: {
        invocationId: env.GITHUB_RUN_ID,
      },
    },
  };

  if (env.GITHUB_EVENT_PATH !== undefined) {
    const ghEvent = JSON.parse(
      fs.readFileSync(env.GITHUB_EVENT_PATH).toString()
    );
    for (const input in ghEvent.inputs) {
      // The invocation parameters belong here and are the top-level GitHub
      // workflow inputs.
      predicate.buildDefinition.externalParameters[`input_${input}`] = {
        value: ghEvent.inputs.input,
      };
    }
  }

  return predicate;
}

// createURI creates the fully qualified URI out of the repository
function createURI(repository: string, ref: string): string {
  return `git+https://github.com/${repository}@${ref}`;
}
