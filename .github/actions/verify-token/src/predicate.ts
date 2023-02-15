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
import * as process from "process";
import * as fs from "fs";
import type { Endpoints } from "@octokit/types";

const DELEGATOR_BUILD_TYPE =
  "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0";

type ApiWorkflowRun =
  Endpoints["GET /repos/{owner}/{repo}/actions/runs/{run_id}"]["response"]["data"];

interface Builder {
  id: string;
  version?: string;
  builderDependencies?: ArtifactReference[];
}

interface DigestSet {
  [key: string]: string;
}

interface Metadata {
  invocationId?: string;
  startedOn?: Date;
  finishedOn?: Date;
}

interface ArtifactReference {
  uri: string;
  digest: DigestSet;
  localName?: string;
  downloadLocation?: string;
  mediaType?: string;
}

interface BuildDefinition {
  // buildType is a TypeURI that unambiguously indicates the type of this message and how to initiate the build.
  buildType: string;

  // externalParameters is the set of top-level external inputs to the build.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  externalParameters: any;

  // systemParameters describes parameters of the build environment provided by the `builder`.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  systemParameters?: any;

  // resolvedDependencies are dependencies needed at build time.
  resolvedDependencies?: ArtifactReference[];
}

interface RunDetails {
  builder: Builder;

  metadata: Metadata;

  byproducts?: ArtifactReference[];
}

interface SLSAv1Predicate {
  // buildDefinition describes the inputs to the build.
  buildDefinition: BuildDefinition;

  // runDetails includes details specific to this particular execution of the build.
  runDetails: RunDetails;
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

export function createPredicate(
  rawTokenObj: rawTokenInterface,
  toolURI: string,
  currentRun: ApiWorkflowRun
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

  const payload = github.context;
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
          ref: env.GITHUB_REF,
          repository: env.GITHUB_REPOSITORY,
          // TODO(#1643): extract path from GITHUB_WORKFLOW_REF.
          path: currentRun.path,
        },
        // We only use source here because the source contained the source
        // repository and the build configuration.
        source: sourceRef,
      },
      systemParameters: {
        GITHUB_EVENT_NAME: payload.eventName,
        GITHUB_JOB: payload.job,
        GITHUB_REF: payload.ref,
        GITHUB_REF_TYPE: env.GITHUB_REF_TYPE || "",
        GITHUB_REPOSITORY: env.GITHUB_REPOSITORY || "",
        GITHUB_RUN_ATTEMPT: env.GITHUB_RUN_ATTEMPT || "",
        GITHUB_RUN_ID: payload.runId,
        GITHUB_RUN_NUMBER: payload.runNumber,
        GITHUB_SHA: payload.sha,
        GITHUB_WORKFLOW: payload.workflow,
        GITHUB_ACTOR_ID: String(currentRun.actor?.id || ""),
        GITHUB_REPOSITORY_ID: String(currentRun.repository.id || ""),
        GITHUB_REPOSITORY_OWNER_ID: String(
          currentRun.repository.owner.id || ""
        ),
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
        invocationId: `https://github.com/${currentRun.repository.full_name}/actions/runs/${currentRun.id}/attempts/${currentRun.run_attempt}`,
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
