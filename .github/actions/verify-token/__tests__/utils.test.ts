/**
 * @fileoverview Tests for utils.ts
 */

import {
  createSourceURI,
  createTriggerURI,
  getSourceSha1,
  getTriggerPath,
  getTriggerSha1,
  getTriggerRef,
  getTriggerRepository,
} from "../src/utils";
import { githubObj, rawTokenInterface } from "../src/types";

describe("getTriggerPath", () => {
  it("extracts path from workflow ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "octocat/hello-world",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "",
      workflow_ref:
        "octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch",
      workflow_sha: "",
    };

    const token = createToken(obj, "12345");
    const path = getTriggerPath(token);
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });

  it("extracts path from workflow ref without git ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "octocat/hello-world",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "",
      workflow_ref: "octocat/hello-world/.github/workflows/my-workflow.yml",
      workflow_sha: "",
    };

    const token = createToken(obj, "12345");
    const path = getTriggerPath(token);
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });
});

function createToken(obj: githubObj, sha1: string): rawTokenInterface {
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
        sha1: sha1,
      },
    },
    github: obj,
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
      inputs: new Map(),
      masked_inputs: [],
    },
  };
  return token;
}

describe("getTriggerSha1", () => {
  it("get trigger sha1", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "12345");
    const sha1 = getTriggerSha1(token);
    expect(sha1).toEqual("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");
  });

  it("invalid length", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    expect(() => {
      const token = createToken(obj, "12345");
      const sha1 = getTriggerSha1(token);
    }).toThrow();
  });

  it("invalid character", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaZ",
      workflow_ref: "",
      workflow_sha: "",
    };

    expect(() => {
      const token = createToken(obj, "12345");
      const sha1 = getTriggerSha1(token);
    }).toThrow();
  });
});

describe("getTriggerRepository", () => {
  it("get trigger repo", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "12345");
    const repo = getTriggerRepository(token);
    expect(repo).toEqual("org/name");
  });
});

describe("getTriggerRef", () => {
  it("get trigger ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "the-ref",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "12345");
    const ref = getTriggerRef(token);
    expect(ref).toEqual("the-ref");
  });
});

describe("getSourceSha1", () => {
  it("default value", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "");
    const sha1 = getSourceSha1(token);
    expect(sha1).toEqual("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");
  });

  it("overwriten by TRW", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "1111111111111111111111111111111111111111");
    const sha1 = getSourceSha1(token);
    expect(sha1).toEqual("1111111111111111111111111111111111111111");
  });

  it("invalid length", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aa",
      workflow_ref: "",
      workflow_sha: "",
    };

    expect(() => {
      const token = createToken(obj, "");
      const sha1 = getSourceSha1(token);
    }).toThrow();
  });

  it("invalid character", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaZ",
      workflow_ref: "",
      workflow_sha: "",
    };

    expect(() => {
      const token = createToken(obj, "");
      const sha1 = getSourceSha1(token);
    }).toThrow();
  });
});

describe("createTriggerURI", () => {
  it("with ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "the-ref",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "1111111111111111111111111111111111111111");
    const uri = createTriggerURI(token);
    expect(uri).toEqual("git+https://github.com/org/name@the-ref");
  });
  it("no ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "1111111111111111111111111111111111111111");
    const uri = createTriggerURI(token);
    expect(uri).toEqual("git+https://github.com/org/name");
  });
});

describe("createSourceURI", () => {
  it("with ref no overwrite", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "the-ref",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "");
    const uri = createSourceURI(token);
    expect(uri).toEqual("git+https://github.com/org/name@the-ref");
  });
  it("no ref no overwrite", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "");
    const uri = createSourceURI(token);
    expect(uri).toEqual("git+https://github.com/org/name");
  });
  it("ref and overwrite", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "the-ref",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "1111111111111111111111111111111111111111");
    const uri = createSourceURI(token);
    expect(uri).toEqual("git+https://github.com/org/name");
  });
  it("no ref and overwrite", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_payload_sha256: "",
      ref: "",
      ref_type: "",
      repository: "org/name",
      repository_id: "",
      repository_owner_id: "",
      run_attempt: "",
      run_id: "",
      run_number: "",
      sha: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      workflow_ref: "",
      workflow_sha: "",
    };

    const token = createToken(obj, "1111111111111111111111111111111111111111");
    const uri = createSourceURI(token);
    expect(uri).toEqual("git+https://github.com/org/name");
  });
});
