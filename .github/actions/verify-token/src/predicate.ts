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
}

interface Invocation {
  configSource: ConfigSource;
  parameters: Object;
  environment: Environment;
}

interface Environment {
  [key: string]: Object | undefined;
}

interface WorkflowParameters {
  eventInputs?: Object;
}

interface DigestSet {
  [key: string]: string;
}

interface ConfigSource {
  uri: string;
  digest: DigestSet;
  entryPoint: string;
}

interface Steps {
  workingDir: string;
  command: string[];
  env: string[];
}

interface BuildConfig {
  inputs: Object;
  steps?: Steps[];
  version: number;
}

interface Metadata {
  buildInvocationId?: string;
  buildStartedOn?: Date;
  buildFinishedOn?: Date;
  completeness?: Completeness;
  reproducible?: boolean;
}

interface Completeness {
  parameters: boolean;
  environment: boolean;
  materials: boolean;
}

interface Material {
  uri: string;
  digest: DigestSet;
}

interface SLSAv02Predicate {
  // URI identifying the builder making the attestation.
  builder: Builder;

  // URI indicating the type of build.
  buildType: string;

  // Invocation information including the entry point.
  invocation: Invocation;

  // A JSON object describing the BuildConfig.
  buildConfig?: BuildConfig;

  // Build Metadata.
  metadata: Metadata;

  // Materials required for the build.
  materials: Material[];
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
  toolURI: string
): SLSAv02Predicate {
  const { env } = process;

  const workflowInputs: WorkflowParameters = {};
  const callerRepo: string = createURI(
    env.GITHUB_REPOSITORY || "",
    env.GITHUB_REF || ""
  );
  // getEntryPoint via GitHub API via runID and repository
  const predicate: SLSAv02Predicate = {
    builder: { id: toolURI },
    buildType: DELEGATOR_BUILD_TYPE,
    invocation: {
      parameters: workflowInputs, // The caller's workflow inputs.
      configSource: {
        uri: callerRepo,
        entryPoint: env.GITHUB_WORKFLOW || "",
        digest: {
          sha1: env.GITHUB_SHA || "",
        },
      },
      environment: {
        GITHUB_ACTOR_ID: env.GITHUB_ACTOR_ID,
        GITHUB_EVENT_NAME: env.GITHUB_EVENT_NAME,
        GITHUB_JOB: env.GITHUB_JOB,
        GITHUB_REF: env.GITHUB_REF,
        GITHUB_REF_TYPE: env.GITHUB_REF_TYPE,
        GITHUB_REPOSITORY: env.GITHUB_REPOSITORY,
        GITHUB_REPOSITORY_ID: env.GITHUB_REPOSITORY_ID,
        GITHUB_REPOSITORY_OWNER_ID: env.GITHUB_REPOSITORY_OWNER_ID,
        GITHUB_RUN_ATTEMPT: env.GITHUB_RUN_ATTEMPT,
        GITHUB_RUN_ID: env.GITHUB_RUN_ID,
        GITHUB_RUN_NUMBER: env.GITHUB_RUN_NUMBER,
        GITHUB_SHA: env.GITHUB_SHA,
        GITHUB_WORKFLOW: env.GITHUB_WORKFLOW,
        GITHUB_WORKFLOW_REF: env.GITHUB_WORKFLOW_REF,
        GITHUB_WORKFLOW_SHA: env.GITHUB_WORKFLOW_SHA,
        IMAGE_OS: env.ImageOS,
        IMAGE_VERSION: env.ImageVersion,
        RUNNER_ARCH: env.RUNNER_ARCH,
        RUNNER_NAME: env.RUNNER_NAME,
        RUNNER_OS: env.RUNNER_OS,
      },
    },
    buildConfig: {
      version: 1,
      inputs: rawTokenObj.tool.inputs,
    },
    materials: [
      {
        uri: callerRepo,
        digest: {
          sha1: process.env.GITHUB_SHA || "",
        },
      },
    ],
    metadata: {
      reproducible: false,
      completeness: {
        parameters: true,
        environment: false,
        materials: false, // This may be true here.
      },
    },
  };
  if (env.GITHUB_EVENT_PATH !== undefined) {
    const ghEvent = JSON.parse(
      fs.readFileSync(env.GITHUB_EVENT_PATH).toString()
    );
    workflowInputs.eventInputs = ghEvent.inputs;
    predicate.invocation.parameters = workflowInputs;
    predicate.invocation.configSource.entryPoint = ghEvent.workflow;
    predicate.invocation.environment["GITHUB_EVENT_PAYLOAD"] = ghEvent;
  }

  return predicate;
}

// createURI creates the fully qualified URI out of the repository
function createURI(repository: string, ref: string): string {
  return `git+https://github.com/${repository}@${ref}`;
}
