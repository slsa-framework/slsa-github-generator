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

export interface githubObj {
  actor_id: string;
  event_name: string;
  event_payload_sha256: string;
  ref: string;
  ref_type: string;
  repository: string;
  repository_id: string;
  repository_owner_id: string;
  run_attempt: string;
  run_id: string;
  run_number: string;
  sha: string;
  workflow_ref: string;
  workflow_sha: string;
}

export interface imageObj {
  os: string;
  version: string;
}

export interface runnerObj {
  arch: string;
  name: string;
  os: string;
}

export interface rawTokenInterface {
  version: number;
  slsaVersion: string;
  context: string;
  builder: {
    private_repository: boolean;
    runner_label: string;
    audience: string;
  };
  github: githubObj;
  runner: runnerObj;
  image: imageObj;
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
    // masked_inputs is a list of input names who's value should be masked in the provenance.
    masked_inputs: string[];
  };
}
