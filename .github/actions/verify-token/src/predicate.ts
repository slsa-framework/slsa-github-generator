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

import { rawTokenInterface, ArtifactReference, SLSAv1Predicate } from "./types";

const DELEGATOR_BUILD_TYPE =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

// getWorkflowPath returns the workflow's path from the workflow_ref.
export function getWorkflowPath(): string {
  // GITHUB_WORKFLOW_REF contains the repository name in the path. We will trim
  // it out.
  // e.g. 'octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch'
  const ref = (process.env.GITHUB_WORKFLOW_REF || "").trim();
  const repo = (process.env.GITHUB_REPOSITORY || "").trim();
  const repoPrefix = `${repo}/`;
  if (!ref.startsWith(repoPrefix)) {
    throw new Error(
      `expected workflow ref '${ref}' to start with repository name '${repo}'.`
    );
  }

  // Strip off the repo name and git ref from the workflow path.
  return ref.substring(repoPrefix.length).split("@", 1)[0];
}

export function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string
): SLSAv1Predicate {
  const { env } = process;

  const callerRepo: string = createURI(
    env.GITHUB_REPOSITORY || "",
    env.GITHUB_REF || ""
  );

  const sourceRef: ArtifactReference = {
    uri: callerRepo,
    digest: {
      sha1: env.GITHUB_SHA || "",
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
          ref: env.GITHUB_REF || "",
          repository: env.GITHUB_REPOSITORY || "",
          path: getWorkflowPath(),
        },
        // We only use source here because the source contained the source
        // repository and the build configuration.
        source: sourceRef,
      },
      systemParameters: {
        GITHUB_EVENT_NAME: env.GITHUB_EVENT_NAME || "",
        GITHUB_JOB: env.GITHUB_JOB || "",
        GITHUB_REF: env.GITHUB_REF || "",
        GITHUB_REF_TYPE: env.GITHUB_REF_TYPE || "",
        GITHUB_REPOSITORY: env.GITHUB_REPOSITORY || "",
        GITHUB_RUN_ATTEMPT: env.GITHUB_RUN_ATTEMPT || "",
        GITHUB_RUN_ID: env.GITHUB_RUN_ID || "",
        GITHUB_RUN_NUMBER: env.GITHUB_RUN_NUMBER || "",
        GITHUB_SHA: env.GITHUB_SHA || "",
        GITHUB_WORKFLOW: env.GITHUB_WORKFLOW || "",
        GITHUB_ACTOR_ID: env.GITHUB_ACTOR_ID || "",
        GITHUB_REPOSITORY_ID: env.GITHUB_REPOSITORY_ID || "",
        GITHUB_REPOSITORY_OWNER_ID: env.GITHUB_REPOSITORY_OWNER_ID || "",
        GITHUB_WORKFLOW_REF: env.GITHUB_WORKFLOW_REF || "",
        GITHUB_WORKFLOW_SHA: env.GITHUB_WORKFLOW_SHA || "",
        IMAGE_OS: env.ImageOS || "",
        IMAGE_VERSION: env.ImageVersion || "",
        RUNNER_ARCH: env.RUNNER_ARCH || "",
        RUNNER_NAME: env.RUNNER_NAME || "",
        RUNNER_OS: env.RUNNER_OS || "",
      },
    },
    runDetails: {
      // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1504):
      // Record raw token inputs (tool action inputs).
      builder: {
        id: toolURI,
      },
      metadata: {
        invocationId: `https://github.com/${env.GITHUB_REPOSITORY}/actions/runs/${env.GITHUB_RUN_ID}/attempts/${env.GITHUB_RUN_ATTEMPT}`,
      },
    },
  };

  // Put GitHub event payload into systemParameters.
  // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
  if (env.GITHUB_EVENT_PATH) {
    const ghEvent = JSON.parse(
      fs.readFileSync(env.GITHUB_EVENT_PATH).toString()
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
