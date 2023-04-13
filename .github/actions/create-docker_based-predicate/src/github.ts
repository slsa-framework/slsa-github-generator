import * as process from "process";
import * as types from "./predicate";
import * as github from "@actions/github";
import type { ApiWorkflowRun } from "./predicate";
import * as tscommon from "tscommon";

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

// addGitHubParameters adds trusted GitHub context to internal paramters
// and external parameters.
export function addGitHubParameters(
  predicate: types.SLSAv1Predicate,
  currentRun: ApiWorkflowRun
): types.SLSAv1Predicate {
  const { env } = process;
  const ctx = github.context;

  if (!predicate.buildDefinition.internalarameters) {
    predicate.buildDefinition.internalParameters = {};
  }
  const internalParams = predicate.buildDefinition.internalParameters;

  // Put GitHub context and env vars into internalParameters.
  internalParams.GITHUB_EVENT_NAME = ctx.eventName;
  internalParams.GITHUB_JOB = ctx.job;
  internalParams.GITHUB_REF = ctx.ref;
  internalParams.GITHUB_BASE_REF = env.GITHUB_BASE_REF || "";
  internalParams.GITHUB_REF_TYPE = env.GITHUB_REF_TYPE || "";
  internalParams.GITHUB_REPOSITORY = env.GITHUB_REPOSITORY || "";
  internalParams.GITHUB_RUN_ATTEMPT = env.GITHUB_RUN_ATTEMPT || "";
  internalParams.GITHUB_RUN_ID = ctx.runId;
  internalParams.GITHUB_RUN_NUMBER = ctx.runNumber;
  internalParams.GITHUB_SHA = ctx.sha;
  internalParams.GITHUB_WORKFLOW = ctx.workflow;
  internalParams.GITHUB_WORKFLOW_REF = env.GITHUB_WORKFLOW_REF || "";
  internalParams.GITHUB_WORKFLOW_SHA = env.GITHUB_WORKFLOW_SHA || "";
  internalParams.IMAGE_OS = env.ImageOS || "";
  internalParams.IMAGE_VERSION = env.ImageVersion || "";
  internalParams.RUNNER_ARCH = env.RUNNER_ARCH || "";
  internalParams.RUNNER_NAME = env.RUNNER_NAME || "";
  internalParams.RUNNER_OS = env.RUNNER_OS || "";
  internalParams.GITHUB_ACTOR_ID = String(currentRun.actor?.id || "");
  internalParams.GITHUB_REPOSITORY_ID = String(currentRun.repository.id || "");
  internalParams.GITHUB_REPOSITORY_OWNER_ID = String(
    currentRun.repository.owner.id || ""
  );

  // Put GitHub event payload into internalParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  if (env.GITHUB_EVENT_PATH) {
    const ghEvent = JSON.parse(
      tscommon.safeReadFileSync(env.GITHUB_EVENT_PATH || "").toString()
    );
    internalParams.GITHUB_EVENT_PAYLOAD = ghEvent;
  }

  predicate.buildDefinition.internalParameters = internalParams;

  if (!env.GITHUB_WORKFLOW_REF) {
    throw new Error("missing GITHUB_WORKFLOW_REF");
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
