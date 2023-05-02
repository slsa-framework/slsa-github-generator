/**
 * @fileoverview Tests for inputs.ts
 */

import { updateSLSAToken } from "../src/inputs";
import { rawTokenInterface } from "../src/types";

describe("updateSLSAToken", () => {
  it("remove bool", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
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
      ])
    );
  });

  it("remove empty string", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
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
      ])
    );
  });

  it("remove integer", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
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
      ])
    );
  });

  it("remove string", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
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
      ])
    );
  });

  it("no 'on' field", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
    const content = `
ona:
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

    expect(() => {
      updateSLSAToken(content, token);
    }).toThrow();
  });

  it("no 'workflow_call' field", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
    const content = `
on:
  workflow_calla:
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

    expect(() => {
      updateSLSAToken(content, token);
    }).toThrow();
  });

  it("no 'inputs' field", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>()
    );
  });

  it("no fields under 'inputs'", async () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}'
    );
    const token = createToken(inputs);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>()
    );
  });

  it("no 'inputs' field no workflow inputs", async () => {
    const inputs = JSON.parse("{}");
    const token = createToken(inputs);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>()
    );
  });

  it("empty 'inputs' field no workflow inputs", async () => {
    const inputs = JSON.parse("{}");
    const token = createToken(inputs);
    const content = `
on:
  workflow_call:
    secrets:
      registry-password:
    inputs:
`;
    const ret = updateSLSAToken(content, token);
    expect(ret.tool.inputs).toEqual(
      new Map<string, string | number | boolean>()
    );
  });
});

function createToken(
  inputs: Map<string, string | number | boolean>
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
    },
  };
  return token;
}
