import * as fs from "fs";
import * as process from "process";
import * as types from "./predicate";
import * as github from "@actions/github";
import type { ApiWorkflowRun } from "./predicate";

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
  return res.data;
}

// addGitHubParameters adds trusted GitHub context to system paramters
// and external parameters.
export function addGitHubParameters(
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

  predicate.buildDefinition.systemParameters = systemParams;

  if (!env.GITHUB_WORKFLOW_REF) {
    throw new Error("missing GITHUB_WORKFLOW_REF");
  }
  const [workflowPath, workflowRef] = env.GITHUB_WORKFLOW_REF.split("@", 2);
  const [, , ...path] = workflowPath.split("/");

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
