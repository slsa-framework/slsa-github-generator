/*
Copyright 2022 SLSA Authors
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
import * as core from "@actions/core";
import * as sigstore from "sigstore";
import * as process from "process";

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
  rekorBaseURL: sigstore.sigstore.DEFAULT_REKOR_BASE_URL,
};

async function run(): Promise<void> {
  try {
    /* Test locally:
        $ env INPUT_SLSA-WORKFLOW-RECIPIENT="laurentsimon/slsa-delegated-tool" \
        INPUT_SLSA-REKOR-LOG-PUBLIC=true \
        INPUT_SLSA-RUNNER-LABEL="ubuntu-latest" \
        INPUT_SLSA-BUILD-ACTION-PATH="./actions/build-artifacts-composite" \
        INPUT_SLSA-WORKFLOW-INPUTS="{\"name1\":\"value1\",\"name2\":\"value2\"}" \
        nodejs ./dist/index.js
    */

    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const rekorLogPublic = core.getInput("slsa-rekor-log-public");
    const runnerLabel = core.getInput("slsa-runner-label");
    const buildArtifactsActionPath = core.getInput("slsa-build-action-path");
    // The workflow inputs are represented as a JSON object theselves.
    const workflowsInputsText = core.getInput("slsa-workflow-inputs");

    // Log the inputs for troubleshooting.
    core.debug(`workflowsInputsText: ${workflowsInputsText}`);
    core.debug(`workfowInputs: `);
    const workflowInputs = JSON.parse(workflowsInputsText);
    const workflowInputsMap = new Map(Object.entries(workflowInputs));
    for (const [key, value] of workflowInputsMap) {
      core.info(` ${key}: ${value}`);
    }

    const payload = JSON.stringify(github.context.payload, undefined, 2);
    core.debug(`The event payload: ${payload}`);

    // Construct an unsigned SLSA token.
    const unsignedSlsaToken = {
      version: 1,
      context: "SLSA delegator framework",
      builder: {
        rekor_log_public: rekorLogPublic,
        runner_label: runnerLabel,
        audience: workflowRecipient,
      },
      github: {
        // TODO(#1410): Re-evaluate if we need the context.
        context: github,
        event_name: process.env.GITHUB_EVENT_NAME,
        run_attempt: process.env.GITHUB_RUN_ATTEMPT,
        run_id: process.env.GITHUB_RUN_ID,
        run_number: process.env.GITHUB_RUN_NUMBER,
        workflow: process.env.GITHUB_WORKFLOW,
        sha: process.env.GITHUB_SHA,
        repository: process.env.GITHUB_REPOSITORY,
        repository_owner: process.env.GITHUB_REPOSITORY_OWNER,
        // TODO(#1411): Record if these become available.
        // repository_id: process.env.GITHUB_REPOSITORY_ID,
        // repository_owner_id: process.env.GITHUB_REPOSITORY_OWNER_ID,
        ref: process.env.GITHUB_REF,
        ref_type: process.env.GITHUB_REF_TYPE,
        actor: process.env.GITHUB_ACTOR,
      },
      tool: {
        actions: {
          build_artifacts: {
            path: buildArtifactsActionPath,
          },
        },
        inputs: workflowInputs,
      },
    };

    // Prepare the base64 unsigned token.
    const unsignedToken = JSON.stringify(unsignedSlsaToken, undefined);
    const unsignedB64Token = Buffer.from(unsignedToken).toString("base64");
    core.info(`unsignedToken: ${unsignedToken}`);
    core.info(`unsignedB64Token: ${unsignedB64Token}`);

    // Sign and prepare the base64 bundle.
    const eventName = process.env.GITHUB_EVENT_NAME || "";
    let bundleStr = "";
    if (eventName === "pull_request") {
      bundleStr = "PLACEHOLDER_SIGNATURE";
    } else {
      const bundle = await sigstore.sigstore.sign(
        Buffer.from(unsignedB64Token),
        signOptions
      );
      bundleStr = JSON.stringify(bundle);
    }

    const bundleB64 = Buffer.from(bundleStr).toString("base64");
    core.info(`bundleStr: ${bundleStr}`);
    core.info(`bundleB64: ${bundleB64}`);

    // Verify just to double check.
    // NOTE: this is an offline verification.
    // TODO(https://github.com/sigstore/sigstore-js/issues/215): Re-enable when fixed.
    // await sigstore.sigstore.verify(bundle, Buffer.from(unsignedB64Token));

    // Output the signed token.
    core.info(`slsa-token: ${bundleB64}.${unsignedB64Token}`);
    core.setOutput("slsa-token", `${bundleB64}.${unsignedB64Token}`);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}

run();
