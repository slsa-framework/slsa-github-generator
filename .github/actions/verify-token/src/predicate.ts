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
  config_source: ConfigSource;
  parameters: Object;
  environment: Environment;
}

interface Environment {
  [key: string]: Object;
}

interface WorkflowParameters {
  event_inputs?: Object;
}

interface DigestSet {
  [key: string]: string;
}

interface ConfigSource {
  uri: string;
  digest: DigestSet;
  entry_point: string;
}

interface Steps {
  working_dir: string;
  command: string[];
  env: string[];
}

interface BuildConfig {
  inputs: Object;
  steps?: Steps[];
  version: number;
}

interface Metadata {
  build_invocation_id?: string;
  build_started_on?: Date;
  build_finished_on?: Date;
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
  build_type?: string;

  // Invocation information including the entry point.
  invocation: Invocation;

  // A JSON object describing the BuildConfig.
  build_config?: BuildConfig;

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
  const workflowInputs: WorkflowParameters = {};
  const callerRepo: string = createURI(
    process.env.GITHUB_REPOSITORY || "",
    process.env.GITHUB_REF || ""
  );
  // getEntryPoint via GitHub API via runID and repository
  const predicate: SLSAv02Predicate = {
    builder: { id: toolURI },
    build_type: DELEGATOR_BUILD_TYPE,
    invocation: {
      parameters: workflowInputs, // The caller's workflow inputs.
      config_source: {
        uri: callerRepo,
        entry_point: process.env.GITHUB_WORKFLOW || "",
        digest: {
          sha1: process.env.GITHUB_SHA || "",
        },
      },
      environment: {
        github_run_number: process.env.GITHUB_RUN_NUMBER || "",
        github_run_id: process.env.GITHUB_RUN_ID || "",
        github_run_attempt: process.env.GITHUB_RUN_ATTEMPT || "",
        github_event_name: process.env.GITHUB_EVENT_NAME || "",
        github_ref_type: process.env.GITHUB_REF_TYPE || "",
        github_ref: process.env.GITHUB_REF || "",
        github_base_ref: process.env.GITHUB_BASE_REF || "",
        github_head_ref: process.env.GITHUB_HEAD_REF || "",
        github_actor: process.env.GITHUB_ACTOR || "",
        github_sha1: process.env.GITHUB_SHA || "",
        github_repository_owner: process.env.GITHUB_REPOSITORY_OWNER || "",
        github_repository_owner_id:
          process.env.GITHUB_REPOSITORY_OWNER_ID || "",
        github_actor_id: process.env.GITHUB_ACTOR_ID || "",
        github_repository_id: process.env.GITHUB_REPOSITORY_ID || "",
      },
    },
    build_config: {
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
  if (process.env.GITHUB_EVENT_PATH !== undefined) {
    const ghEvent = JSON.parse(
      fs.readFileSync(process.env.GITHUB_EVENT_PATH).toString()
    );
    workflowInputs.event_inputs = ghEvent.inputs;
    predicate.invocation.parameters = workflowInputs;
    predicate.invocation.config_source.entry_point = ghEvent.workflow;
    predicate.invocation.environment["github_event_payload"] = ghEvent;
  }

  return predicate;
}

// createURI creates the fully qualified URI out of the repository
function createURI(repository: string, ref: string): string {
  return `git+https://github.com/${repository}@${ref}`;
}
