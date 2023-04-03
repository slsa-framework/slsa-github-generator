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
import { sigstore } from "sigstore";
import * as process from "process";

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
};

async function run(): Promise<void> {
  try {
    /* Test locally:
        $ env INPUT_SLSA-WORKFLOW-RECIPIENT="laurentsimon/slsa-delegated-tool" \
        INPUT_SLSA-REKOR-LOG-PUBLIC=true \
        INPUT_SLSA-RUNNER-LABEL="ubuntu-latest" \
        INPUT_SLSA-BUILD-ACTION-PATH="./actions/build-artifacts-composite" \
        INPUT_SLSA-WORKFLOW-INPUTS="{\"name1\":\"value1\",\"name2\":\"value2\",\"name3\":\"value3\",\"name4\":\"value4\"}" \
        INPUT_SLSA-WORKFLOW-INPUTS-MASK="name2, name4" \
        nodejs ./dist/index.js
    */

    const slsaVersion = core.getInput("slsa-version");
    if (!["v1-rc1", "v0.2"].includes(slsaVersion)) {
      throw new Error(`Unsupported slsa-version: ${slsaVersion}`);
    }
    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const rekorLogPublic = core.getInput("slsa-rekor-log-public");
    const runnerLabel = core.getInput("slsa-runner-label");
    const buildArtifactsActionPath = core.getInput("slsa-build-action-path");
    const workflowsInputsMask = core.getInput("slsa-workflow-masked-inputs");
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

    const workflowMaskedInputs = getMaskedInputs(workflowsInputsMask);
    core.info(`maskedInputs: `);
    for (const value of workflowMaskedInputs) {
      core.info(` ${value}`);
    }

    const payload = JSON.stringify(github.context.payload, undefined, 2);
    core.debug(`The event payload: ${payload}`);

    // Construct an unsigned SLSA token.
    const unsignedSlsaToken = {
      version: 1,
      slsaVersion,
      context: "SLSA delegator framework",
      builder: {
        rekor_log_public: rekorLogPublic,
        runner_label: runnerLabel,
        audience: workflowRecipient,
      },
      github: {
        actor_id: process.env.GITHUB_ACTOR_ID,
        event_name: process.env.GITHUB_EVENT_NAME,
        event_path: process.env.GITHUB_EVENT_PATH,
        ref: process.env.GITHUB_REF,
        ref_type: process.env.GITHUB_REF_TYPE,
        repository: process.env.GITHUB_REPOSITORY,
        repository_id: process.env.GITHUB_REPOSITORY_ID,
        repository_owner_id: process.env.GITHUB_REPOSITORY_OWNER_ID,
        run_attempt: process.env.GITHUB_RUN_ATTEMPT,
        run_id: process.env.GITHUB_RUN_ID,
        run_number: process.env.GITHUB_RUN_NUMBER,
        sha: process.env.GITHUB_SHA,
        workflow_ref: process.env.GITHUB_WORKFLOW_REF,
        workflow_sha: process.env.GITHUB_WORKFLOW_SHA,
      },
      image: {
        os: process.env.ImageOS,
        version: process.env.ImageVersion,
      },
      runner: {
        arch: process.env.RUNNER_ARCH,
        name: process.env.RUNNER_NAME,
        os: process.env.RUNNER_OS,
      },
      tool: {
        actions: {
          build_artifacts: {
            path: buildArtifactsActionPath,
          },
        },
        inputs: workflowInputs,
        masked_inputs: workflowMaskedInputs,
      },
    };

    // Prepare the base64 unsigned token.
    const unsignedToken = JSON.stringify(unsignedSlsaToken, undefined);
    const unsignedB64Token = Buffer.from(unsignedToken).toString("base64");
    core.info(`unsignedToken: ${unsignedToken}`);
    core.info(`unsignedB64Token: ${unsignedB64Token}`);

    // Sign and prepare the base64 bundle.
    const bundle = await sigstore.sign(
      Buffer.from(unsignedB64Token),
      signOptions
    );

    // Verify just to double check.
    // NOTE: this is an offline verification.
    // TODO(#1668): re-enable verification.
    // await sigstore.verify(bundle, Buffer.from(unsignedB64Token));
    const bundleStr = JSON.stringify(bundle);

    const bundleB64 = Buffer.from(bundleStr).toString("base64");
    core.info(`bundleStr: ${bundleStr}`);
    core.info(`bundleB64: ${bundleB64}`);
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

function getMaskedInputs(inputsStr: string): string[] {
  const ret = [];
  const inputArr = inputsStr.split(",");
  for (const input of inputArr) {
    ret.push(input.trim());
  }
  return ret;
}

run();
