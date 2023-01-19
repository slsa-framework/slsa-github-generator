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
import type { WorkflowDispatchEvent } from "@octokit/webhooks-types";

import type { Endpoints } from "@octokit/types";

export type ApiWorkflowRun =
  Endpoints["GET /repos/{owner}/{repo}/actions/runs/{run_id}"]["response"]["data"];

export interface Builder {
  id: string;
  version?: string;
  builderDependencies?: ArtifactReference[];
}

export interface DigestSet {
  [key: string]: string;
}

export interface Metadata {
  invocationId?: string;
  startedOn?: Date;
  finishedOn?: Date;
}

export interface ParameterValue_Artifact {
  artifact: ArtifactReference;
}

export interface ParameterValue_String {
  value: string;
}

export type ParameterValue = ParameterValue_Artifact | ParameterValue_String;

export interface ArtifactReference {
  uri: string;
  digest: DigestSet;
  localName?: string;
  downloadLocation?: string;
  mediaType?: string;
}

export interface BuildDefinition {
  // buildType is a TypeURI that unambiguously indicates the type of this message and how to initiate the build.
  buildType: string;

  // externalParameters is the set of top-level external inputs to the build.
  externalParameters: { [key: string]: ParameterValue };

  // systemParameters describes parameters of the build environment provided by the `builder`.
  systemParameters?: { [key: string]: ParameterValue };

  // resolvedDependencies are dependencies needed at build time.
  resolvedDependencies?: ArtifactReference[];
}

export interface RunDetails {
  builder: Builder;

  metadata: Metadata;

  byproducts?: ArtifactReference[];
}

export interface SLSAv1Predicate {
  // buildDefinition describes the inputs to the build.
  buildDefinition: BuildDefinition;

  // runDetails includes details specific to this particular execution of the build.
  runDetails: RunDetails;
}

// addGitHubContext
export function addGitHubContext(
  predicate: SLSAv1Predicate,
  currentRun: ApiWorkflowRun
): SLSAv1Predicate {
  const { env } = process;

  if (!predicate.buildDefinition.systemParameters) {
    predicate.buildDefinition.systemParameters = {};
  }

  // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1505):
  // Add GitHub event payload.
  const systemParams = predicate.buildDefinition.systemParameters;
  systemParams["GITHUB_EVENT_NAME"] = { value: env.GITHUB_EVENT_NAME || "" };
  systemParams["GITHUB_JOB"] = { value: env.GITHUB_JOB || "" };
  systemParams["GITHUB_REF"] = { value: env.GITHUB_REF || "" };
  systemParams["GITHUB_REF_TYPE"] = { value: env.GITHUB_REF_TYPE || "" };
  systemParams["GITHUB_REPOSITORY"] = { value: env.GITHUB_REPOSITORY || "" };
  systemParams["GITHUB_RUN_ATTEMPT"] = { value: env.GITHUB_RUN_ATTEMPT || "" };
  systemParams["GITHUB_RUN_ID"] = { value: env.GITHUB_RUN_ID || "" };
  systemParams["GITHUB_RUN_NUMBER"] = { value: env.GITHUB_RUN_NUMBER || "" };
  systemParams["GITHUB_SHA"] = { value: env.GITHUB_SHA || "" };
  systemParams["GITHUB_WORKFLOW"] = { value: env.GITHUB_WORKFLOW || "" };
  systemParams["GITHUB_ACTOR_ID"] = {
    value: String(currentRun.actor?.id || ""),
  };
  systemParams["GITHUB_REPOSITORY_ID"] = {
    value: String(currentRun.repository.id || ""),
  };
  systemParams["GITHUB_REPSITORY_OWNER_ID"] = {
    value: String(currentRun.repository.owner.id || ""),
  };
  systemParams["GITHUB_WORKFLOW_REF"] = {
    value: env.GITHUB_WORKFLOW_REF || "",
  };
  systemParams["GITHUB_WORKFLOW_SHA"] = {
    value: env.GITHUB_WORKFLOW_SHA || "",
  };
  systemParams["IMAGE_OS"] = { value: env.ImageOS || "" };
  systemParams["IMAGE_VERSION"] = { value: env.ImageVersion || "" };
  systemParams["RUNNER_ARCH"] = { value: env.RUNNER_ARCH || "" };
  systemParams["RUNNER_NAME"] = { value: env.RUNNER_NAME || "" };
  systemParams["RUNNER_OS"] = { value: env.RUNNER_OS || "" };

  if (env.GITHUB_EVENT_NAME === "workflow_dispatch") {
    if (env.GITHUB_EVENT_PATH) {
      const ghEvent: WorkflowDispatchEvent = JSON.parse(
        fs.readFileSync(env.GITHUB_EVENT_PATH).toString()
      );

      for (const input in ghEvent.inputs) {
        // The invocation parameters belong here and are the top-level GitHub
        // workflow inputs.
        predicate.buildDefinition.externalParameters[`input_${input}`] = {
          value: String(ghEvent.inputs[input] || ""),
        };
      }
    }
  }

  return predicate;
}

// createURI creates the fully qualified URI out of the repository
export function createURI(repository: string, ref: string): string {
  if (!repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  let refVal = "";
  if (ref) {
    refVal = `@${ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}
