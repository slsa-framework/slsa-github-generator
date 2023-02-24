/**
 * @fileoverview Tests for predicate.ts
 */

import {
  validateField,
  validateFieldAnyOf,
  validateFieldNonEmpty,
  validateGitHubFields,
} from "../src/validate";
import { githubObj } from "../src/types";

describe("validateField", () => {
  it("validates equal values", () => {
    validateField("foo", "foo", "foo");
  });

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
    process.env.GITHUB_JOB = "234";
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
      event_path: "/path/to/event.json",
      job: "234",
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
    validateGitHubFields(obj);
  });

  it("does not validate githubObj with empty fields", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_JOB = "234";
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
      event_path: "/path/to/event.json",
      job: "234",
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
    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });

  it("does not validate githubObj with unequal fields", () => {
    process.env.GITHUB_ACTOR_ID = "123";
    process.env.GITHUB_EVENT_NAME = "workflow_dispatch";
    process.env.GITHUB_EVENT_PATH = "/path/to/event.json";
    process.env.GITHUB_JOB = "234";
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
      event_path: "/path/to/event.json",
      job: "234",
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
    expect(() => {
      validateGitHubFields(obj);
    }).toThrow();
  });
});
