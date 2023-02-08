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

import * as fs from "fs";
import * as process from "process";
import * as types from "./types";
import * as github from "@actions/github";
import type { ApiWorkflowRun } from "./types";

// getWorkflowRun retrieves the current WorkflowRun given the repository (owner/repo)
// and run ID.
export async function getWorkflowRun(
  repository: string,
  run_id: number,
  token: string
): Promise<ApiWorkflowRun> {
  const octokit = github.getOctokit(token);
  const [owner, repo] = repository.split("/");
  const res = await octokit.rest.actions.getWorkflowRun({
    owner,
    repo,
    run_id: Number(process.env.GITHUB_RUN_ID),
  });
  return res.data as ApiWorkflowRun;
}

// addGitHubSystemParameters adds trusted GitHub context to system paramters.
export function addGitHubSystemParameters(
  predicate: types.SLSAv1Predicate,
  currentRun: ApiWorkflowRun
): types.SLSAv1Predicate {
  const { env } = process;
  const ctx = github.context;

  if (!predicate.buildDefinition.systemParameters) {
    predicate.buildDefinition.systemParameters = {};
  }
  const systemParams = predicate.buildDefinition.systemParameters;

  // Put GitHub context and env vars into systemParameters.
  systemParams.GITHUB_EVENT_NAME = ctx.eventName;
  systemParams.GITHUB_JOB = ctx.job;
  systemParams.GITHUB_REF = ctx.ref;
  systemParams.GITHUB_REF_TYPE = env.GITHUB_REF_TYPE || "";
  systemParams.GITHUB_REPOSITORY = env.GITHUB_REPOSITORY || "";
  systemParams.GITHUB_RUN_ATTEMPT = env.GITHUB_RUN_ATTEMPT || "";
  systemParams.GITHUB_RUN_ID = ctx.runId;
  systemParams.GITHUB_RUN_NUMBER = ctx.runNumber;
  systemParams.GITHUB_SHA = ctx.sha;
  systemParams.GITHUB_WORKFLOW = ctx.workflow;
  systemParams.GITHUB_WORKFLOW_REF = env.GITHUB_WORKFLOW_REF || "";
  systemParams.GITHUB_WORKFLOW_SHA = env.GITHUB_WORKFLOW_SHA || "";
  systemParams.IMAGE_OS = env.ImageOS || "";
  systemParams.IMAGE_VERSION = env.ImageVersion || "";
  systemParams.RUNNER_ARCH = env.RUNNER_ARCH || "";
  systemParams.RUNNER_NAME = env.RUNNER_NAME || "";
  systemParams.RUNNER_OS = env.RUNNER_OS || "";
  systemParams.GITHUB_ACTOR_ID = String(currentRun.actor?.id || "");
  systemParams.GITHUB_REPOSITORY_ID = String(currentRun.repository.id || "");
  systemParams.GITHUB_REPOSITORY_OWNER_ID = String(
    currentRun.repository.owner.id || ""
  );

  // Put GitHub event payload into systemParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  if (env.GITHUB_EVENT_PATH) {
    const ghEvent = JSON.parse(
      fs.readFileSync(env.GITHUB_EVENT_PATH).toString()
    );
    systemParams.GITHUB_EVENT_PAYLOAD = ghEvent;
  }

  return predicate;
}

export type InputMap = {
  [k: string]: unknown;
} | null;

// getWorkflowInputs gets the workflow runs' inputs (only populated on workflow dispatch).
export function getWorkflowInputs(): InputMap {
  const { env } = process;

  if (env.GITHUB_EVENT_NAME === "workflow_dispatch") {
    return github.context.payload.inputs;
  }

  return null;
}

// getInvocationID returns the URI describing the globally unique invocation ID.
export function getInvocationID(currentRun: ApiWorkflowRun): string {
  return `https://github.com/${currentRun.repository.full_name}/actions/runs/${currentRun.id}/attempts/${currentRun.run_attempt}`;
}

// getSourceArtifactReference retrieves the artifact reference based on the
// caller repository.
export function getSourceArtifactReference(): types.ArtifactReference {
  const { env } = process;

  const callerRepo: string = createURI(
    env.GITHUB_REPOSITORY || "",
    env.GITHUB_REF || ""
  );

  return {
    uri: callerRepo,
    digest: {
      sha1: env.GITHUB_SHA || "",
    },
  };
}

// createURI creates the fully qualified URI out of the repository
function createURI(repository: string, ref: string): string {
  if (!repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  let refVal = "";
  if (ref) {
    refVal = `@${ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}
