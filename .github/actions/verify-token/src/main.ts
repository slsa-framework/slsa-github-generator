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
import * as process from "process";
import * as fs from "fs";
import * as child_process from "child_process";
import { githubObj, rawTokenInterface, createPredicate } from "./predicate";
import { getEnv, resolvePathInput } from "./utils";
import * as pred from "predicate-utils";

async function run(): Promise<void> {
  try {
    /* Test locally. Requires a GitHub token:
        $ env INPUT_SLSA-WORKFLOW-RECIPIENT="delegator_generic_slsa3.yml" \
        INPUT_SLSA-UNVERIFIED-TOKEN="$(cat testdata/slsa-token)" \
        INPUT_TOKEN="$(gh auth token)" \
        INPUT_OUTPUT-PREDICATE="predicate.json" \
        GITHUB_EVENT_NAME="workflow_dispatch" \
        GITHUB_RUN_ATTEMPT="1" \
        GITHUB_RUN_ID="3790385865" \
        GITHUB_RUN_NUMBER="200" \
        GITHUB_WORKFLOW="delegate release project" \
        GITHUB_SHA="8cbf4d422367d8499d5980a837cb9cc8e1e67001" \
        GITHUB_REPOSITORY="laurentsimon/slsa-delegate-project" \
        GITHUB_REPOSITORY_ID="567955265" \
        GITHUB_REPOSITORY_OWNER="laurentsimon" \
        GITHUB_REPOSITORY_OWNER_ID="64505099" \
        GITHUB_ACTOR_ID="64505099" \
        GITHUB_REF="refs/heads/main" \
        GITHUB_BASE_REF="" \
        GITHUB_REF_TYPE="branch" \
        GITHUB_ACTOR="laurentsimon" \
        GITHUB_WORKSPACE="$(pwd)" \
        nodejs ./dist/index.js
    */

    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const unverifiedToken = core.getInput("slsa-unverified-token");

    const outputPredicate = core.getInput("output-predicate");
    if (!outputPredicate) {
      // detect if output predicate is null or empty string.
      throw new Error("output-predicate must be supplied");
    }
    const wd = getEnv("GITHUB_WORKSPACE");
    const safeOutput = resolvePathInput(outputPredicate, wd);
    // TODO(#1513): Use a common utility to harden file writes.
    if (fs.existsSync(safeOutput)) {
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

    // Now generate the SLSA predicate using the verified token and the GH context.
    const token = core.getInput("token");
    if (!token) {
      throw new Error("token not provided");
    }
    const ownerRepo = getEnv("GITHUB_REPOSITORY");

    const currentWorkflowRun = await pred.getWorkflowRun(
      ownerRepo,
      Number(process.env.GITHUB_RUN_ID),
      token
    );

    const predicate = await createPredicate(
      rawTokenObj,
      toolURI,
      currentWorkflowRun
    );

    fs.writeFileSync(safeOutput, JSON.stringify(predicate), {
      flag: "ax",
      mode: 0o600,
    });
    core.debug(`predicate: ${JSON.stringify(predicate)}`);
    core.debug(`Wrote predicate to ${safeOutput}`);

    core.setOutput("tool-repository", toolRepository);
    core.setOutput("tool-ref", toolRef);
    core.setOutput("slsa-verified-token", rawTokenStr);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}

function parseCertificateIdentity(
  bundle: sigstore.sigstore.Bundle
): [string, string, string] {
  if (bundle === undefined) {
    throw new Error(`undefined bundle.`);
  }
  if (bundle.verificationMaterial === undefined) {
    throw new Error(`undefined bundle.verificationMaterial.`);
  }
  if (bundle.verificationMaterial.x509CertificateChain === undefined) {
    throw new Error(
      `undefined bundle.verificationMaterial.x509CertificateChain.`
    );
  }
  if (
    bundle.verificationMaterial.x509CertificateChain.certificates.length === 0
  ) {
    throw new Error(
      `bundle.verificationMaterial.x509CertificateChaincertificates is empty.`
    );
  }
  // NOTE: the first certificate is the client certificate.
  const clientCertDer = Buffer.from(
    bundle.verificationMaterial.x509CertificateChain.certificates[0].rawBytes,
    "base64"
  );
  const clientCertPath = "client.cert";
  fs.writeFileSync(clientCertPath, clientCertDer);

  // https://stackabuse.com/executing-shell-commands-with-node-js/
  // The SAN from the certificate looks like:
  // `
  //  X509v3 Subject Alternative Name: critical\n
  //      URI:https://github.com/laurentsimon/slsa-delegated-tool/.github/workflows/tool1_slsa3.yml@refs/heads/main\n
  // `
  const result = child_process
    .execSync(`openssl x509 -in ${clientCertPath} -noout -ext subjectAltName`)
    .toString();
  const index = result.indexOf("URI:");
  if (index === -1) {
    throw new Error("error: cannot find URI in subjectAltName");
  }
  const toolURI = result.slice(index + 4).replace("\n", "");
  core.debug(`tool-uri: ${toolURI}`);

  // NOTE: we can use the job_workflow_ref and job_workflow_sha when they become available.
  const [toolRepository, toolRef] = extractIdentifyFromSAN(toolURI);
  core.debug(`tool-repository: ${toolRepository}`);
  core.debug(`tool-ref: ${toolRef}`);

  return [toolURI, toolRepository, toolRef];
}

function extractIdentifyFromSAN(URI: string): [string, string] {
  // NOTE: the URI looks like:
  // https://github.com/laurentsimon/slsa-delegated-tool/.github/workflows/tool1_slsa3.yml@refs/heads/main.
  // We want to extract:
  // - the repository: laurentsimon/slsa-delegated-tool
  // - the ref: refs/heads/main
  const parts = URI.split("@");
  if (parts.length !== 2) {
    throw new Error(`invalid URI (1): ${URI}`);
  }
  const ref = parts[1];
  const url = parts[0];
  const gitHubURL = "https://github.com/";
  if (!url.startsWith(gitHubURL)) {
    throw new Error(`not a GitHub URI: ${URI}`);
  }
  // NOTE: we omit the gitHubURL from the URL.
  const parts2 = url.slice(gitHubURL.length).split("/");
  if (parts2.length <= 2) {
    throw new Error(`invalid URI (2): ${URI}`);
  }
  const repo = `${parts2[0]}/${parts2[1]}`;
  return [repo, ref];
}

function validateGitHubFields(gho: githubObj): void {
  validateField(
    "github.event_name",
    gho.event_name,
    process.env.GITHUB_EVENT_NAME
  );
  validateField(
    "github.run_attempt",
    gho.run_attempt,
    process.env.GITHUB_RUN_ATTEMPT
  );
  validateField("github.run_id", gho.run_id, process.env.GITHUB_RUN_ID);
  validateField(
    "github.run_number",
    gho.run_number,
    process.env.GITHUB_RUN_NUMBER
  );
  validateField("github.workflow", gho.workflow, process.env.GITHUB_WORKFLOW);
  validateField("github.sha", gho.sha, process.env.GITHUB_SHA);
  validateField(
    "github.repository",
    gho.repository,
    process.env.GITHUB_REPOSITORY
  );
  validateField(
    "github.repository_owner",
    gho.repository_owner,
    process.env.GITHUB_REPOSITORY_OWNER
  );
  validateField("github.ref", gho.ref, process.env.GITHUB_REF);
  validateField("github.ref_type", gho.ref_type, process.env.GITHUB_REF_TYPE);
  validateField("github.actor", gho.actor, process.env.GITHUB_ACTOR);
  // TODO(#1411): Record if these become available.
  // repository_id: process.env.GITHUB_REPOSITORY_ID,
  // repository_owner_id: process.env.GITHUB_REPOSITORY_OWNER_ID,
  // repository_actor_id: process.env.GITHUB_ACTOR_ID,
}

function validateFieldAnyOf<T>(name: string, actual: T, expected: T[]): void {
  for (const value of expected) {
    if (actual === value) {
      // Found a match.
      return;
    }
  }
  throw new Error(
    `mismatch ${name}: got '${actual}', expected one of '${expected.join(
      ","
    )}'.`
  );
}

function validateField<T>(name: string, actual: T, expected: T): void {
  if (actual !== expected) {
    throw new Error(
      `mismatch ${name}: got '${actual}', expected '${expected}'.`
    );
  }
}

function validateNonEmptyField(name: string, actual: string): void {
  if (actual === "") {
    throw new Error(`empty ${name}, expected non-empty value.`);
  }
}

run();
