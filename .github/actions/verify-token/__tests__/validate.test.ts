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
 * @fileoverview Tests for validate.ts
 */

import {
  validateField,
  validateFieldAnyOf,
  validateFieldNonEmpty,
  validateGitHubFields,
  validateAndMaskInputs,
} from "../src/validate";
import { githubObj, rawTokenInterface } from "../src/types";

// Import `fs` using commonJS so it can be mocked.
const tscommon = require("tscommon");

jest.mock("tscommon");

afterEach(() => {
  jest.clearAllMocks();
});

describe("validateField", () => {
  it("validates equal values", () => {
    validateField("foo", "foo", "foo");
  });

  it("validates equal empty values", () => {
    validateField("foo", "", "", true);
  });

  expect(() => {
    validateField("foo", "", "");
  }).toThrow();

  it("does not validate unequal values", () => {
    expect(() => {
      validateField("foo", "foo", "bar");
    }).toThrow();

    expect(() => {
      validateField("foo", "bar", "foo");
    }).toThrow();
  });

  it("does not validate unequal types", () => {
    expect(() => {
      validateField("foo", null, undefined);
    }).toThrow();

    expect(() => {
      validateField("foo", "", undefined);
    }).toThrow();

    expect(() => {
      validateField("foo", "undefined", undefined);
    }).toThrow();

    expect(() => {
      validateField("foo", "", null);
    }).toThrow();

    expect(() => {
      validateField("foo", "null", null);
    }).toThrow();
  });

  it("does not validate equal empty values", () => {
    expect(() => {
      validateField("foo", "", "");
    }).toThrow();

    expect(() => {
      validateField("foo", undefined, undefined);
    }).toThrow();

    expect(() => {
      validateField("foo", null, null);
    }).toThrow();
  });
});

describe("validateFieldAnyOf", () => {
  it("validates from a single choice", () => {
    validateFieldAnyOf("foo", "foo", ["foo"]);
  });

  it("validates from a multiple choices", () => {
    validateFieldAnyOf("foo", "foo", ["bar", "foo"]);
  });

  it("does not validate from unequal choices", () => {
    expect(() => {
      validateFieldAnyOf("foo", "foo", ["bar", "baz"]);
    }).toThrow();
  });

  it("does not validate from no choices", () => {
    expect(() => {
      validateFieldAnyOf("foo", "foo", []);
    }).toThrow();
  });

  it("does not validate from unequal types", () => {
    expect(() => {
      validateFieldAnyOf("foo", "", [null, undefined]);
    }).toThrow();

    expect(() => {
      validateFieldAnyOf("foo", "null", [null]);
    }).toThrow();

    expect(() => {
      validateFieldAnyOf("foo", "undefined", [undefined]);
    }).toThrow();
  });
});

describe("validateFieldNonEmpty", () => {
  it("validates from a non-empty value", () => {
    validateFieldNonEmpty("foo", "foo");
    validateFieldNonEmpty("foo", "0");
    validateFieldNonEmpty("foo", "null");
    validateFieldNonEmpty("foo", "undefined");
  });

  it("does not validate empty value", () => {
    expect(() => {
      validateFieldNonEmpty("foo", "");
    }).toThrow();
  });
});

function createToken(
  inputs: Map<string, string | number | boolean>,
  masked: string[],
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
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "456",
      repository_owner_id: "789",
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
      masked_inputs: masked,
    },
  };
  return token;
}

describe("validateAndMaskInputs", () => {
  it("valid masked inputs", () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const masked = ["name2", "name3", "name4"];
    const token = createToken(inputs, masked);
    expect(validateAndMaskInputs(token).tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name2", "***"],
        ["name3", "***"],
        ["name4", "***"],
      ]),
    );
  });

  it("single masked inputs", () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const masked = ["name2"];
    const token = createToken(inputs, masked);
    expect(validateAndMaskInputs(token).tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name2", "***"],
        ["name3", ""],
        ["name4", true],
      ]),
    );
  });

  it("empty masked inputs", () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const masked = [""];
    const token = createToken(inputs, masked);
    expect(validateAndMaskInputs(token).tool.inputs).toEqual(
      new Map<string, string | number | boolean>([
        ["name1", "value1"],
        ["name2", 2],
        ["name3", ""],
        ["name4", true],
      ]),
    );
  });

  it("invalid masked input name", () => {
    const inputs = JSON.parse(
      '{"name1": "value1", "name2": 2, "name3": "", "name4": true}',
    );
    const masked = ["does-not-exist"];
    const token = createToken(inputs, masked);
    expect(() => {
      validateAndMaskInputs(token);
    }).toThrow();
  });
});

describe("validateGitHubFields", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
  });

  afterEach(() => {
    process.env = env;
  });

  it("validates from githubObj", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_BASE_REF = "base-ref";
    process.env.GITHUB_REF = "refs/heads/mybranch";
    process.env.GITHUB_REF_TYPE = "branch";
    process.env.GITHUB_REPOSITORY = "ianlewis/byob-test";
    process.env.GITHUB_REPOSITORY_ID = "456";
    process.env.GITHUB_REPOSITORY_OWNER_ID = "789";
    process.env.GITHUB_RUN_ATTEMPT = "1";
    process.env.GITHUB_RUN_ID = "123456789";
    process.env.GITHUB_RUN_NUMBER = "1";
    process.env.GITHUB_SHA = "deadbeef";
    process.env.GITHUB_WORKFLOW_REF =
      "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_WORKFLOW_SHA = "deadbeef";

    const obj: githubObj = {
      actor_id: "123",
      event_name: "workflow_dispatch",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "456",
      repository_owner_id: "789",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    };

    tscommon.safeFileSha256.mockReturnValueOnce(obj.event_payload_sha256);

    validateGitHubFields(obj);
  });

  it("does not validate githubObj with empty fields", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_BASE_REF = "base-ref";
    process.env.GITHUB_REF = "refs/heads/mybranch";
    process.env.GITHUB_REF_TYPE = "branch";
    process.env.GITHUB_REPOSITORY = "ianlewis/byob-test";
    process.env.GITHUB_REPOSITORY_ID = "";
    process.env.GITHUB_REPOSITORY_OWNER_ID = "789";
    process.env.GITHUB_RUN_ATTEMPT = "1";
    process.env.GITHUB_RUN_ID = "123456789";
    process.env.GITHUB_RUN_NUMBER = "1";
    process.env.GITHUB_SHA = "deadbeef";
    process.env.GITHUB_WORKFLOW_REF =
      "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_WORKFLOW_SHA = "deadbeef";

    const obj: githubObj = {
      actor_id: "123",
      event_name: "workflow_dispatch",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "", // NOTE: Empty
      repository_owner_id: "789",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    };

    tscommon.safeFileSha256.mockReturnValueOnce(obj.event_payload_sha256);

    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });

  it("does not validate githubObj with unequal repository", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_JOB = "234";
    process.env.GITHUB_BASE_REF = "base-ref";
    process.env.GITHUB_REF = "refs/heads/mybranch";
    process.env.GITHUB_REF_TYPE = "branch";
    process.env.GITHUB_REPOSITORY = "ianlewis/byob-test";
    process.env.GITHUB_REPOSITORY_ID = "456";
    process.env.GITHUB_REPOSITORY_OWNER_ID = "789";
    process.env.GITHUB_RUN_ATTEMPT = "1";
    process.env.GITHUB_RUN_ID = "123456789";
    process.env.GITHUB_RUN_NUMBER = "1";
    process.env.GITHUB_SHA = "deadbeef";
    process.env.GITHUB_WORKFLOW = ".github/workflow/myworkflow.yml";
    process.env.GITHUB_WORKFLOW_REF =
      "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_WORKFLOW_SHA = "deadbeef";

    const obj: githubObj = {
      actor_id: "123",
      event_name: "workflow_dispatch",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "asraa/byob-test", // NOTE: Not equal
      repository_id: "456",
      repository_owner_id: "789",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    };

    tscommon.safeFileSha256.mockReturnValueOnce(obj.event_payload_sha256);

    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });

  it("does not validate githubObj with unequal event_payload_sha256", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_JOB = "234";
    process.env.GITHUB_BASE_REF = "base-ref";
    process.env.GITHUB_REF = "refs/heads/mybranch";
    process.env.GITHUB_REF_TYPE = "branch";
    process.env.GITHUB_REPOSITORY = "ianlewis/byob-test";
    process.env.GITHUB_REPOSITORY_ID = "456";
    process.env.GITHUB_REPOSITORY_OWNER_ID = "789";
    process.env.GITHUB_RUN_ATTEMPT = "1";
    process.env.GITHUB_RUN_ID = "123456789";
    process.env.GITHUB_RUN_NUMBER = "1";
    process.env.GITHUB_SHA = "deadbeef";
    process.env.GITHUB_WORKFLOW = ".github/workflow/myworkflow.yml";
    process.env.GITHUB_WORKFLOW_REF =
      "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_WORKFLOW_SHA = "deadbeef";

    const obj: githubObj = {
      actor_id: "123",
      event_name: "workflow_dispatch",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "456",
      repository_owner_id: "789",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    };

    tscommon.safeFileSha256.mockReturnValueOnce("incorrect-sha"); // NOTE: Not equal

    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });

  it("does not validate githubObj with unequal base_ref", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_BASE_REF = "not-base-ref";
    process.env.GITHUB_REF = "refs/heads/mybranch";
    process.env.GITHUB_REF_TYPE = "branch";
    process.env.GITHUB_REPOSITORY = "ianlewis/byob-test";
    process.env.GITHUB_REPOSITORY_ID = "456";
    process.env.GITHUB_REPOSITORY_OWNER_ID = "789";
    process.env.GITHUB_RUN_ATTEMPT = "1";
    process.env.GITHUB_RUN_ID = "123456789";
    process.env.GITHUB_RUN_NUMBER = "1";
    process.env.GITHUB_SHA = "deadbeef";
    process.env.GITHUB_WORKFLOW_REF =
      "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_WORKFLOW_SHA = "deadbeef";

    const obj: githubObj = {
      actor_id: "123",
      event_name: "workflow_dispatch",
      event_payload_sha256:
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      base_ref: "base-ref",
      ref: "refs/heads/mybranch",
      ref_type: "branch",
      repository: "ianlewis/byob-test",
      repository_id: "456",
      repository_owner_id: "789",
      run_attempt: "1",
      run_id: "123456789",
      run_number: "1",
      sha: "deadbeef",
      workflow_ref:
        "ianlewis/byob-test/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "deadbeef",
    };

    tscommon.safeFileSha256.mockReturnValueOnce(obj.event_payload_sha256);

    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });
});
