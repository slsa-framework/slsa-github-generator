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

import { parseCertificateIdentity, RawToken } from "./token";
import {
  validateGitHubFields,
  validateField,
  validateFieldAnyOf,
  validateNonEmptyField,
} from "./validate";

async function run(): Promise<void> {
  try {
    /* Test locally:
        $ env INPUT_SLSA-WORKFLOW-RECIPIENT="delegator_generic_slsa3.yml" \
        INPUT_SLSA-UNVERIFIED-TOKEN="$(cat testdata/slsa-token)" \
        GITHUB_EVENT_NAME="workflow_dispatch" \
        GITHUB_RUN_ATTEMPT="1" \
        GITHUB_RUN_ID="3790385865" \
        GITHUB_RUN_NUMBER="200" \
        GITHUB_WORKFLOW="delegate release project" \
        GITHUB_SHA="8cbf4d422367d8499d5980a837cb9cc8e1e67001" \
        GITHUB_REPOSITORY="laurentsimon/slsa-delegate-project" \
        GITHUB_REPOSITORY_OWNER="laurentsimon" \
        GITHUB_REF="refs/heads/main" \
        GITHUB_REF_TYPE="branch" \
        GITHUB_ACTOR="laurentsimon" \
        nodejs ./dist/index.js
    */

    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const unverifiedToken = core.getInput("slsa-unverified-token");

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
    const rawTokenObj: RawToken = JSON.parse(rawTokenStr);

    // Verify the version.
    validateField("version", rawTokenObj.version, 1);

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
    validateNonEmptyField(
      "tool.actions.build_artifacts.path",
      rawTokenObj.tool.actions.build_artifacts.path
    );

    // No validation needed for the builder inputs.
    // They may be empty.

    // Extract certificate information.
    const [toolURI, toolRepository, toolRef] = parseCertificateIdentity(bundle);

    core.debug(`slsa-verified-token: ${rawTokenStr}`);

    core.setOutput("tool-repository", toolRepository);
    core.setOutput("tool-ref", toolRef);
    core.setOutput("tool-uri", toolURI);
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
