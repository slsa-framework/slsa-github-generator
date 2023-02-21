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

import {
  githubObj,
  rawTokenInterface,
  ArtifactReference,
  SLSAv1Predicate,
} from "./types";

const DELEGATOR_BUILD_TYPE =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

// getWorkflowPath returns the workflow's path from the workflow_ref.
export function getWorkflowPath(obj: githubObj): string {
  // GITHUB_WORKFLOW_REF contains the repository name in the path. We will trim
  // it out.
  // e.g. 'octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch'
  // Strip off the repo name and git ref from the workflow path.
  return obj.workflow_ref
    .substring(`${obj.repository}/`.length)
    .split("@", 1)[0];
}

export function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string
): SLSAv1Predicate {
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

  // NOTE: see example at https://github.com/slsa-framework/slsa/blob/main/docs/github-actions-workflow/examples/v0.1/example.json.
  const predicate: SLSAv1Predicate = {
    buildDefinition: {
      buildType: DELEGATOR_BUILD_TYPE,
      externalParameters: {
        // Inputs to the TRW, which define the interface of the builder for the
        // BYOB framework.
        // TODO(#1575): mask inputs.
        inputs: rawTokenObj.tool.inputs,
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
      systemParameters: {
        GITHUB_ACTOR: rawTokenObj.github.actor,
        GITHUB_ACTOR_ID: rawTokenObj.github.actor_id,
        GITHUB_EVENT_NAME: rawTokenObj.github.event_name,
        GITHUB_JOB: rawTokenObj.github.job,
        GITHUB_REF: rawTokenObj.github.ref,
        GITHUB_REF_TYPE: rawTokenObj.github.ref_type,
        GITHUB_REPOSITORY: rawTokenObj.github.repository,
        GITHUB_REPOSITORY_ID: rawTokenObj.github.repository_id,
        GITHUB_REPOSITORY_OWNER: rawTokenObj.github.repository_owner,
        GITHUB_REPOSITORY_OWNER_ID: rawTokenObj.github.repository_owner_id,
        GITHUB_RUN_ATTEMPT: rawTokenObj.github.run_attempt,
        GITHUB_RUN_ID: rawTokenObj.github.run_id,
        GITHUB_RUN_NUMBER: rawTokenObj.github.run_number,
        GITHUB_SHA: rawTokenObj.github.sha,
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

  // Put GitHub event payload into systemParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  if (rawTokenObj.github.event_path) {
    // NOTE: event_path has been validated as the same as env.GITHUB_EVENT_PATH
    const ghEvent = JSON.parse(
      fs.readFileSync(rawTokenObj.github.event_path).toString()
    );
    predicate.buildDefinition.systemParameters.GITHUB_EVENT_PAYLOAD = ghEvent;
  }

  return predicate;
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
