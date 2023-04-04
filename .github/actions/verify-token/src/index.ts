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

import * as core from "@actions/core";
import * as sigstore from "sigstore";
import {
  validateField,
  validateFieldAnyOf,
  validateFieldNonEmpty,
  validateGitHubFields,
  validateAndMaskInputs,
} from "./validate";
import { createPredicate as createPredicate_v1 } from "./predicate1";
import { createPredicate as createPredicate_v02 } from "./predicate02";
import { rawTokenInterface } from "./types";
import { filterWorkflowInputs } from "./inputs";
import { parseCertificate } from "./utils";
import * as tscommon from "tscommon";

async function run(): Promise<void> {
  try {
    /* Test locally. Requires a GitHub token:
        $ rm -f client.cert && rm -f predicate.json && \
        env INPUT_SLSA-WORKFLOW-RECIPIENT="delegator_generic_slsa3.yml" \
        INPUT_SLSA-UNVERIFIED-TOKEN="$(cat testdata/slsa-token)" \
        INPUT_SLSA-VERSION="1.0-rc1" \
        INPUT_TOKEN="$(echo $GH_TOKEN)" \
        INPUT_OUTPUT-PREDICATE="predicate.json" \
        GITHUB_EVENT_NAME="push" \
        GITHUB_RUN_ATTEMPT="1" \
        GITHUB_RUN_ID="4386810663" \
        GITHUB_RUN_NUMBER="74" \
        GITHUB_WORKFLOW="delegate release project" \
        GITHUB_WORKFLOW_REF="laurentsimon/slsa-delegate-project/.github/workflows/anchor-sbom.yml@refs/tags/v0.0.2" \
        GITHUB_WORKFLOW_SHA="66a665d98ad0b990bbcb1dfc57891a63182459ea" \
        GITHUB_SHA="66a665d98ad0b990bbcb1dfc57891a63182459ea" \
        GITHUB_REPOSITORY="laurentsimon/slsa-delegate-project" \
        GITHUB_REPOSITORY_ID="567955265" \
        GITHUB_REPOSITORY_OWNER="laurentsimon" \
        GITHUB_REPOSITORY_OWNER_ID="64505099" \
        GITHUB_ACTOR_ID="64505099" \
        GITHUB_REF="refs/tags/v0.0.2" \
        GITHUB_EVENT_PATH="/home/runner/work/_temp/_github_workflow/event.json" \
        GITHUB_BASE_REF="" \
        GITHUB_REF_TYPE="tag" \
        GITHUB_ACTOR="laurentsimon" \
        GITHUB_WORKSPACE="$(pwd)" \
        nodejs ./dist/dist/index.js
    */

    const ghToken = core.getInput("token");
    if (!ghToken) {
      throw new Error("token not provided");
    }
    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const unverifiedToken = core.getInput("slsa-unverified-token");

    const outputPredicate = core.getInput("output-predicate");
    if (!outputPredicate) {
      // detect if output predicate is null or empty string.
      throw new Error("output-predicate must be supplied");
    }

    if (tscommon.safeExistsSync(outputPredicate)) {
      throw new Error("output-predicate file already exists");
    }

    // Log the inputs for troubleshooting.
    core.debug(`workflowRecipient: ${workflowRecipient}`);
    core.debug(`unverifiedToken: ${unverifiedToken}`);
    const parts = unverifiedToken.split(".");
    if (parts.length !== 2) {
      throw new Error(`malformed token: ${parts.length}`);
    }

    const bundleStr = Buffer.from(parts[0], "base64").toString("utf-8");
    const b64Token = parts[1];
    const bundle = JSON.parse(bundleStr);

    // First, verify the signature, i.e., that it is signed by a certificate that
    // chains up to Fulcio.
    await sigstore.sigstore.verify(bundle, Buffer.from(b64Token));

    const rawToken = Buffer.from(b64Token, "base64");
    core.debug(`bundle: ${bundleStr}`);
    core.debug(`token: ${rawToken}`);

    const rawTokenStr = rawToken.toString();
    const rawTokenObj: rawTokenInterface = JSON.parse(rawTokenStr);

    // Verify the version.
    validateField("version", rawTokenObj.version, 1);

    // Validate the slsaVersion
    validateFieldAnyOf("slsaVersion", rawTokenObj.slsaVersion, [
      "v1-rc1",
      "v0.2",
    ]);

    // Verify the context of the signature.
    validateField("context", rawTokenObj.context, "SLSA delegator framework");

    // Verify the intended recipient.
    validateField(
      "builder.audience",
      rawTokenObj.builder.audience,
      workflowRecipient
    );

    // Verify the runner label.
    validateFieldAnyOf(
      "builder.runner_label",
      rawTokenObj.builder.runner_label,
      ["ubuntu-latest"]
    );

    // Verify the GitHub event information.
    validateGitHubFields(rawTokenObj.github);

    // Validate the build Action is not empty.
    validateFieldNonEmpty(
      "tool.actions.build_artifacts.path",
      rawTokenObj.tool.actions.build_artifacts.path
    );

    // No validation needed for the builder inputs,
    // they may be empty.
    // TODO(#1780): test empty inputs.

    // Extract certificate information.
    const [toolURI, toolRepository, toolRef, toolSha, toolPath] =
      parseCertificate(bundle);

    // Extract the inputs.
    // See https://github.com/slsa-framework/slsa-github-generator/issues/1737.
    const rawFinalTokenObj = await filterWorkflowInputs(
      rawTokenObj,
      ghToken,
      toolRepository,
      toolSha,
      toolPath
    );
    core.debug(
      `workflow inputs: ${JSON.stringify(
        Object.fromEntries(rawFinalTokenObj.tool.inputs)
      )}`
    );

    // Validate the masked inputs and update the token.
    const rawMaskedTokenObj = validateAndMaskInputs(rawFinalTokenObj);
    core.debug(
      `masked inputs: ${JSON.stringify(
        Object.fromEntries(rawMaskedTokenObj.tool.inputs)
      )}`
    );

    core.debug(`slsa-verified-token: ${rawTokenStr}`);

    // Now generate the SLSA predicate using the verified token and the GH context.
    // NOTE: we create the predicate using the token with masked inputs.
    let predicateStr = "";
    switch (rawMaskedTokenObj.slsaVersion) {
      case "v1-rc1": {
        const predicate_v1 = await createPredicate_v1(
          rawMaskedTokenObj,
          toolURI,
          ghToken
        );
        predicateStr = JSON.stringify(predicate_v1);
        break;
      }
      case "v0.2": {
        const predicate_v02 = await createPredicate_v02(
          rawMaskedTokenObj,
          toolURI,
          ghToken
        );
        predicateStr = JSON.stringify(predicate_v02);
        break;
      }
      default: {
        throw new Error(
          `Unsupported slsa-version: ${rawMaskedTokenObj.slsaVersion}`
        );
      }
    }
    tscommon.safeWriteFileSync(outputPredicate, predicateStr);
    core.debug(`predicate: ${predicateStr}`);
    core.debug(`Wrote predicate to ${outputPredicate}`);

    core.setOutput("tool-repository", toolRepository);
    core.setOutput("tool-ref", toolRef);
    // NOTE: we output the token with unmasked inputs because the inputs
    // are needed by the wrapper Action.
    core.setOutput("slsa-verified-token", rawTokenStr);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}

run();
