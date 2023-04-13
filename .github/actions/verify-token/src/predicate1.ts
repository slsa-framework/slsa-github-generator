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

import * as github from "@actions/github";
import * as tscommon from "tscommon";

import { rawTokenInterface } from "./types";
import { createURI, getWorkflowPath } from "./utils";
import { SLSAPredicate, ArtifactReference } from "./slsatypes1";

const DELEGATOR_BUILD_TYPE_V0 =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

export async function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string,
  token: string
): Promise<SLSAPredicate> {
  const callerRepo: string = createURI(
    rawTokenObj.github.repository,
    rawTokenObj.github.ref
  );

  const sourceRef: ArtifactReference = {
    uri: callerRepo,
    digest: {
      sha1: rawTokenObj.github.sha,
    },
  };

  // NOTE: We get the triggering_actor_id from the workflow run via the API.
  // We can trust this value as we have validated the run_id (as much as we can
  // trust the GitHub API on GitHub Actions anyway).
  const octokit = github.getOctokit(token);
  const [owner, repo] = rawTokenObj.github.repository.split("/");
  const { data: current_run } = await octokit.rest.actions.getWorkflowRun({
    owner,
    repo,
    run_id: Number(rawTokenObj.github.run_id),
  });

  // NOTE: see example at https://github.com/slsa-framework/slsa/blob/main/docs/github-actions-workflow/examples/v0.1/example.json.
  const predicate: SLSAPredicate = {
    buildDefinition: {
      buildType: DELEGATOR_BUILD_TYPE_V0,
      externalParameters: {
        // Inputs to the TRW, which define the interface of the builder for the
        // BYOB framework. Some of these values may be masked by the TRW.
        // NOTE: the Map object needs to be converted to an object to serialize to JSON.
        inputs: Object.fromEntries(rawTokenObj.tool.inputs),
        // Variables are always empty for BYOB / builders.
        // TODO(#1555): add support for generators.
        vars: {},
        // NOTE: This is equivalent to the v0.2 entryPoint.
        workflow: {
          ref: rawTokenObj.github.ref,
          repository: rawTokenObj.github.repository,
          path: getWorkflowPath(rawTokenObj.github),
        },
        // We only use source here because the source contained the source
        // repository and the build configuration.
        source: sourceRef,
      },
      internalParameters: {
        GITHUB_ACTOR_ID: rawTokenObj.github.actor_id,
        GITHUB_EVENT_NAME: rawTokenObj.github.event_name,
        GITHUB_REF: rawTokenObj.github.ref,
        GITHUB_REF_TYPE: rawTokenObj.github.ref_type,
        GITHUB_REPOSITORY: rawTokenObj.github.repository,
        GITHUB_REPOSITORY_ID: rawTokenObj.github.repository_id,
        GITHUB_REPOSITORY_OWNER_ID: rawTokenObj.github.repository_owner_id,
        GITHUB_RUN_ATTEMPT: rawTokenObj.github.run_attempt,
        GITHUB_RUN_ID: rawTokenObj.github.run_id,
        GITHUB_RUN_NUMBER: rawTokenObj.github.run_number,
        GITHUB_SHA: rawTokenObj.github.sha,
        // NOTE: the triggering_actor should be returned by the API but the
        // TypeScript type indicates that it could be undefined. If that is
        // the case, then we'll fall back to the actor_id.
        GITHUB_TRIGGERING_ACTOR_ID:
          (current_run.triggering_actor &&
            String(current_run.triggering_actor.id)) ||
          rawTokenObj.github.actor_id,
        GITHUB_WORKFLOW_REF: rawTokenObj.github.workflow_ref,
        GITHUB_WORKFLOW_SHA: rawTokenObj.github.workflow_sha,

        IMAGE_OS: rawTokenObj.image.os,
        IMAGE_VERSION: rawTokenObj.image.version,

        RUNNER_ARCH: rawTokenObj.runner.arch,
        RUNNER_NAME: rawTokenObj.runner.name,
        RUNNER_OS: rawTokenObj.runner.os,
      },
    },
    runDetails: {
      // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1504):
      // Record raw token inputs (tool action inputs).
      builder: {
        id: toolURI,
      },
      metadata: {
        invocationId: `https://github.com/${rawTokenObj.github.repository}/actions/runs/${rawTokenObj.github.run_id}/attempts/${rawTokenObj.github.run_attempt}`,
      },
    },
  };

  // Put GitHub event payload into internalParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  // NOTE: Contents of event_path have been pre-validated.
  predicate.buildDefinition.internalParameters.GITHUB_EVENT_PAYLOAD =
    JSON.parse(
      tscommon.safeReadFileSync(process.env.GITHUB_EVENT_PATH || "").toString()
    );

  return predicate;
}
