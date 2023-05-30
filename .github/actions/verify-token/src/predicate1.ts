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
import {
  createSourceURI,
  getTriggerRepository,
  getTriggerRef,
  getSourceSha1,
  getTriggerPath,
} from "./utils";
import { SLSAPredicate } from "./slsatypes1";

const DELEGATOR_BUILD_TYPE_V0 =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

export async function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string,
  token: string,
  isGenerator: boolean
): Promise<SLSAPredicate> {
  // Trigger information.
  const triggerPath: string = getTriggerPath(rawTokenObj);
  const triggerRef: string = getTriggerRef(rawTokenObj);
  const triggerRepository: string = getTriggerRepository(rawTokenObj);
  // Source information.
  const sourceURI: string = createSourceURI(rawTokenObj);
  const sourceSha1: string = getSourceSha1(rawTokenObj);

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
      externalParameters: {},
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
      },
      resolvedDependencies: [
        {
          uri: sourceURI,
          digest: {
            gitCommit: sourceSha1,
          },
        },
      ],
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

  // Construct the predicate according to the type of builder.
  if (isGenerator) {
    predicate.buildDefinition.externalParameters = {
      workflow: {
        ref: triggerRef,
        repository: `git+https://github.com/${triggerRepository}`,
        path: triggerPath,
      },
      // TODO(#1555): record the vars.
      vars: {},
      // TODO(#2164): record the inputs, depending on the type of trigger events.
      inputs: {},
    };
    // Throw an error for now. We have no generators using v1.0 yet
    // and it's not supported in the slsa-verifier.
    throw new Error("not supported: #2164, #1555");
  } else {
    // NOTE: the workflow information is available in the internalParameters.GITHUB_WORKFLOW_REF.
    predicate.buildDefinition.externalParameters = {
      // Inputs to the TRW, which define the interface of the builder for the
      // BYOB framework. Some of these values may be masked by the TRW.
      // NOTE: the Map object needs to be converted to an object to serialize to JSON.
      inputs: Object.fromEntries(rawTokenObj.tool.inputs),
      // Variables are always empty for BYOB / builders.
      vars: {},
      source: {
        uri: sourceURI,
        digest: {
          gitCommit: sourceSha1,
        },
      },
    };
  }

  // Put GitHub event payload into internalParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  // NOTE: Contents of event_path have been pre-validated.
  predicate.buildDefinition.internalParameters.GITHUB_EVENT_PAYLOAD =
    JSON.parse(
      tscommon.safeReadFileSync(process.env.GITHUB_EVENT_PATH || "").toString()
    );

  return predicate;
}
