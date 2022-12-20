import * as github from "@actions/github";
import * as core from "@actions/core";

function snakeToCamel(str: string): string {
  return str
    .toLowerCase()
    .replace(/([-_][a-z])/g, (group) =>
      group.toUpperCase().replace("-", "").replace("_", "")
    );
}

async function run(): Promise<void> {
  try {
    /* Test locally: 
      export ACTION_INPUTS="$(cat ./ACTION_INPUTS.txt | jq -c)"
      export WORKFLOW_INPUTS="$(cat ./WORKFLOW_INPUTS.txt | jq -c)"
      TOOL_REPOSITORY=laurentsimon/slsa-delegated-tool
      REF=main
    */

    // Read the Action inputs.
    const actionInputs = process.env.ACTION_INPUTS;
    if (!actionInputs) {
      core.setFailed("No actionInputs found.");
      return;
    }
    core.info(`Found Action inputs: ${actionInputs}`);

    // Parse the Action inputs.
    interface inputsObjInterface {
      slsaPrivateRepository: boolean;
      slsaRunnerLabel: string;
      slsaBuildArtifactsActionPath: string;
      slsaWorkflowRecipient: string;
      slsaWorkflowInputs: string;
    }

    const inputsObj: inputsObjInterface = JSON.parse(
      actionInputs,
      function (key, value) {
        const camelCaseKey = snakeToCamel(key);
        // See https://stackoverflow.com/questions/68337817/is-it-possible-to-use-json-parse-to-change-the-keys-from-underscore-to-camelcase.
        if (this instanceof Array || camelCaseKey === key) {
          return value;
        } else {
          this[camelCaseKey] = value;
        }
      }
    );

    /* Log for troubleshooting */
    const inputs = new Map(Object.entries(inputsObj));
    for (const [key, value] of inputs) {
      core.info(`${key}: ${value}`);
    }

    const workflowRecipient = inputsObj.slsaWorkflowRecipient;
    const privateRepository = inputsObj.slsaPrivateRepository;
    const runnerLabel = inputsObj.slsaRunnerLabel;
    const buildArtifactsActionPath = inputsObj.slsaBuildArtifactsActionPath;
    const tmpWorkflowInputs = inputsObj.slsaWorkflowInputs;
    // The workflow inputs are represented as a JSON object theselves.
    const workflowInputs: Map<string, string> = JSON.parse(tmpWorkflowInputs);
    
    // Log the inputs for troubleshooting.
    core.info(`privateRepository: ${privateRepository}`);
    core.info(`runnerLabel: ${runnerLabel}`);
    core.info(`workflowRecipient: ${workflowRecipient}`);
    core.info(`buildArtifactsActionPath: ${buildArtifactsActionPath}`);
    core.info(`workfowInputs:`);
    const workflowInputsMap = new Map(Object.entries(workflowInputs));
    for (const [key, value] of workflowInputsMap) {
      core.info(` ${key}: ${value}`);
    }
     
    // Log for troublehooting.
    const payload = JSON.stringify(github.context.payload, undefined, 2);
    core.info(`The event payload: ${payload}`);

    // Construct our raw token.
    const rawSlsaToken = {
      version: 1,
      context: "SLSA integration framework",
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
        inputs: workflowInputs,
      },
    };

    const token = JSON.stringify(rawSlsaToken, undefined);
    const b64Token = Buffer.from(token).toString('base64');
    // Log for troublehooting.
    core.info(`Base64 raw SLSA token: ${b64Token}`);
    core.info(`Raw SLSA token: ${token}`);

    // Set the output of the Action.
    core.setOutput("base64-slsa-token", b64Token);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.info(`Unexpected error: ${error}`);
    }
  }
}

run();
