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
import { rawTokenInterface, GitHubWorkflowInterface } from "../src/types";
import { fetchToolWorkflow } from "./utils";

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
  const wokflowInputs = new Map(Object.entries(slsaToken.tool.inputs));
  const workflow: GitHubWorkflowInterface = YAML.parse(content);
  if (!workflow.on) {
    throw new Error("no 'on' field");
  }
  if (!workflow.on.workflow_call) {
    throw new Error("no 'workflow_call' field");
  }

  // No inputs field defined.
  if (!workflow.on.workflow_call.inputs) {
    core.info("no input defined in the workflow");
    ret.tool.inputs = new Map();
    return ret;
  }

  // Inputs defined.
  const wInputsMap = new Map(Object.entries(workflow.on.workflow_call.inputs));

  // No fields defined.
  if (wInputsMap.size === 0) {
    ret.tool.inputs = new Map();
    return ret;
  }

  // Fields defined.
  const names = [...wokflowInputs.keys()];
  for (const name of names) {
    core.info(`Processing name: ${name}`);
    if (!wInputsMap.has(name)) {
      core.info(" - Removed");
      wokflowInputs.delete(name);
    }
  }

  // Update the inputs to record.
  ret.tool.inputs = wokflowInputs;
  return ret;
}
