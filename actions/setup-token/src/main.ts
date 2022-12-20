import * as github from "@actions/github";
import * as core from "@actions/core";
import * as sigstore from "sigstore";
import { connected } from "process";
import * as fs from 'fs';

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
  rekorBaseURL: sigstore.sigstore.DEFAULT_REKOR_BASE_URL,
};

async function run(): Promise<void> {
  try {
    /* Test locally:
        $ env INPUT_SLSA-WORKFLOW-RECIPIENT="laurentsimon/slsa-delegated-tool" \
        INPUT_SLSA-PRIVATE-REPOSITORY=true \
        INPUT_SLSA-RUNNER-LABEL="ubuntu-latest" \
        INPUT_SLSA-BUILD-ACTION-PATH="./actions/build-artifacts-composite" \
        INPUT_SLSA-WORKFLOW-INPUTS="{\"name1\":\"value1\",\"name2\":\"value2\",\"private-repository\":true}" \
        nodejs ./dist/index.js
    */
   
    const workflowRecipient = core.getInput("slsa-workflow-recipient");
    const privateRepository = core.getInput("slsa-private-repository");
    const runnerLabel = core.getInput("slsa-runner-label");
    const buildArtifactsActionPath = core.getInput("slsa-build-action-path");
    // The workflow inputs are represented as a JSON object theselves.
    const workflowsInputsText = core.getInput("slsa-workflow-inputs");

    // Log the inputs for troubleshooting.
    core.info(`workflowsInputsText: ${workflowsInputsText}`);
    core.info(`workfowInputs: `);
    const workflowInputs = JSON.parse(workflowsInputsText);
    const workflowInputsMap = new Map(Object.entries(workflowInputs));
    workflowInputsMap.forEach((value, key) => {
      core.info(` ${key}: ${value}`);
    });

    // const payload = JSON.stringify(github.context.payload, undefined, 2);
    // core.info(`The event payload: ${payload}`);

    // Construct an unsigned SLSA token.
    const unsignedSlsaToken = {
      version: 1,
      context: "SLSA delegator framework",
      builder: {
        "private-repository": true,
        "runner-label": runnerLabel,
        audience: workflowRecipient,
      },
      github: {
        context: github.context,
      },
      tool: {
        actions: {
          "build-artifacts": {
            path: buildArtifactsActionPath,
          },
        },
        // TODO: grab the calling workflow here ?
        // We need it for policy authz and we should report it
        // somewhere. Where?
        inputs: workflowInputs,
      },
    };

<<<<<<< Updated upstream
    const token = JSON.stringify(rawSlsaToken, undefined);
    const b64Token = Buffer.from(token).toString('base64');
    // Log for troublehooting.
    core.info(`Base64 raw SLSA token: ${b64Token}`);
    core.info(`Raw SLSA token: ${token}`);
=======
    // Prepare the base64 unsigned token.
    const unsignedToken = JSON.stringify(unsignedSlsaToken, undefined);
    const unsignedB64Token = Buffer.from(unsignedToken).toString("base64");
    core.info(`unsignedToken: ${unsignedToken}`);
    core.info(`unsignedB64Token: ${unsignedB64Token}`);
>>>>>>> Stashed changes

    // Sign and prepare the base64 bundle.
    const bundle = await sigstore.sigstore.sign(
      Buffer.from(unsignedB64Token),
      signOptions
    );
    const bundleStr = JSON.stringify(bundle);
    const bundleB64 = Buffer.from(bundleStr).toString("base64");
    core.info(`bundleStr: ${bundleStr}`);
    core.info(`bundleB64: ${bundleB64}`);

    // Verify just to double check.
    // NOTE: this is an offline verification.
    await sigstore.sigstore.verify(bundle, Buffer.from(unsignedB64Token));

    // Output the signed token.
    core.info(`slsa-token: ${bundleB64}.${unsignedB64Token}`);
    core.setOutput("slsa-token", `${bundleB64}.${unsignedB64Token}`);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.info(`Unexpected error: ${error}`);
    }
  }
}

run();
