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

import * as core from "@actions/core";
import * as YAML from "yaml";
import { rawTokenInterface, gitHubWorkflowInterface } from "../src/types";
import { fetchToolWorkflow, asMap } from "./utils";

// This function removes the fields from the workflow_dispatch,
// and keeps only the inputs from the re-usable workflow.
// We need to filter out event's inputs because GitHub
// reports all of them via the `${{ inputs }}`,
// see https://github.com/actions/runner/issues/2274.
export async function filterWorkflowInputs(
  slsaToken: rawTokenInterface,
  ghToken: string,
  repoName: string,
  hash: string,
  workflowPath: string
): Promise<rawTokenInterface> {
  const content = await fetchToolWorkflow(
    ghToken,
    repoName,
    hash,
    workflowPath
  );
  return updateSLSAToken(content, slsaToken);
}

export function updateSLSAToken(
  content: string,
  slsaToken: rawTokenInterface
): rawTokenInterface {
  const ret = Object.create(slsaToken);
  const workflow: gitHubWorkflowInterface = YAML.parse(content);
  slsaToken.tool.inputs = asMap<string | number | boolean>(
    slsaToken.tool.inputs
  );

  // NOTE: We need to check the presence of workflow_call but it could be empty
  // (e.g. no inputs).
  if (workflow.on?.workflow_call === undefined) {
    throw new Error("no 'workflow_call' field");
  }

  // No inputs field defined.
  if (!workflow.on?.workflow_call?.inputs) {
    core.info("no input defined in the workflow");
    ret.tool.inputs = new Map();
    return ret;
  }

  // Inputs defined.
  const wInputsMap = asMap<Object>(workflow.on.workflow_call.inputs);

  // No fields defined.
  if (wInputsMap.size === 0) {
    ret.tool.inputs = new Map();
    return ret;
  }

  // Fields defined.
  const names = [...slsaToken.tool.inputs.keys()];
  for (const name of names) {
    core.info(`Processing name: ${name}`);
    if (!wInputsMap.has(name)) {
      core.info(" - Removed");
      slsaToken.tool.inputs.delete(name);
    }
  }

  return ret;
}
