// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * @fileoverview Tests for inputs.ts
 */

import { updateSLSAToken } from "../src/inputs";
import { rawTokenInterface } from "../src/types";

describe("updateSLSAToken", () => {
  it("no inputs, no vars", async () => {
    const inputs = new Map<string, string>();
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>(),
    );
  });

  it("remove bool input", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
        name1:
            required: false
        name2:
            required: false
        name3:
            required: false
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name2", 2],
        ["name3", ""],
      ]),
    );
  });

  it("remove empty string input", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
        name1:
            required: false
        name2:
            required: false
        name4:
            required: false
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name2", 2],
        ["name4", true],
      ]),
    );
  });

  it("remove integer input", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
        name1:
            required: false
        name3:
            required: false
        name4:
            required: false
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name3", ""],
        ["name4", true],
      ]),
    );
  });

  it("remove string input", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
        name2:
            required: false
        name3:
            required: false
        name4:
            required: false
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name2", 2],
        ["name3", ""],
        ["name4", true],
      ]),
    );
  });

  it("no 'on' field", async () => {
    const inputs = new Map<string, string>();
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
ona:
  workflow_call:
`;

    expect(() => {
      updateSLSAToken(content, token);
    }).toThrow();
  });

  it("no 'workflow_call' field", async () => {
    const inputs = new Map<string, string>();
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_calla:
`;

    expect(() => {
      updateSLSAToken(content, token);
    }).toThrow();
  });

  it("no 'inputs' field", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>(),
    );
  });

  it("no fields under 'inputs'", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>(),
    );
  });

  it("no 'inputs' field no workflow inputs", async () => {
    const inputs = new Map<string, string>();
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>(),
    );
  });

  it("empty 'inputs' field no workflow inputs", async () => {
    const inputs = new Map<string, string>();
    const vars = new Map<string, string>();
    const token = createToken(inputs, vars);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>(),
    );
  });
});

function createToken(
  inputs: Map<string, string | number | boolean>,
  vars: Map<string, string>,
): rawTokenInterface {
  const token: rawTokenInterface = {
    version: 1,
    slsaVersion: "1.0-rc1",
    context: "the context",
    builder: {
      private_repository: true,
      runner_label: "the label",
      audience: "the audience",
    },
    source: {
      checkout: {
        fetch_depth: 3,
        sha1: "abcdef",
      },
    },
    github: {
      actor_id: "123",
      event_name: "workflow_dispatch",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "456",
      repository_owner_id: "789",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    },
    runner: {
      arch: "arch",
      name: "name",
      os: "os",
    },
    image: {
      os: "os",
      version: "version",
    },
    tool: {
      actions: {
        build_artifacts: {
          path: "path",
        },
      },
      inputs: inputs,
      masked_inputs: [],
      vars: vars,
      masked_vars: [],
    },
  };
  return token;
}
