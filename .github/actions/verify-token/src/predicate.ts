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

import * as pred from "predicate-utils";
import type { ApiWorkflowRun } from "predicate-utils";

const DELEGATOR_BUILD_TYPE =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

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
    inputs: Map<string, string | number | boolean>;
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

export async function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string,
  currentRun: ApiWorkflowRun
): Promise<pred.SLSAv1Predicate> {
  const predicate: pred.SLSAv1Predicate = {
    buildDefinition: {
      buildType: DELEGATOR_BUILD_TYPE,
      externalParameters: {
        // NOTE: This is equivalent to the v0.2 entryPoint.
        workflowPath: currentRun.path,
        // We only use source here because the source contained the source
        // repository and the build configuration.
        source: {
          artifact: pred.getSourceArtifactReference(),
        },
      },
      systemParameters: {},
    },
    runDetails: {
      // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1504):
      // Record raw token inputs (tool action inputs).
      builder: {
        id: toolURI,
      },
      metadata: {
        invocationId: pred.getInvocationID(currentRun),
      },
    },
  };

  // Add trusted GitHub context to systemParameters.
  return pred.addGitHubSystemParameters(predicate, currentRun);
}
