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

import { create } from "domain";
import { OctokitResponse } from "@octokit/types";

const core = require("@actions/core");
const detect = require("../src/detect");
const github = require("@actions/github");

describe("decodeToken", () => {
  it("return email and job_workflow_ref", () => {
    const job_workflow_ref =
      "octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main";
    const payload = {
      iss: "some_issuer",
      aud: "some/audience",
      job_workflow_ref: job_workflow_ref,
    };

    const jwt = `.${Buffer.from(JSON.stringify(payload)).toString("base64")}.`;

    const decoded = detect.decodeToken(jwt);
    expect(decoded.aud).toBe("some/audience");
    expect(decoded.job_workflow_ref).toBe(job_workflow_ref);
  });
});

jest.mock("@actions/core");

describe("detectWorkflowFromOIDC", () => {
  it("success", async () => {
    const job_workflow_ref =
      "octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main";
    const payload = {
      iss: "some_issuer",
      aud: "some/audience",
      job_workflow_ref: job_workflow_ref,
    };

    const jwt = `.${Buffer.from(JSON.stringify(payload)).toString("base64")}.`;
    core.getIDToken.mockClear();
    core.getIDToken.mockReturnValueOnce(jwt);

    const [repo, ref, workflow] =
      await detect.detectWorkflowFromOIDC("some/audience");
    expect(repo).toBe("octo-org/octo-automation");
    expect(ref).toBe("refs/heads/main");
    expect(workflow).toBe(".github/workflows/oidc.yml");
  });
  it("workflow ref can contain '@'", async () => {
    const job_workflow_ref =
      "vitejs/vite/.github/workflows/publish.yml@refs/tags/create-vite@5.0.0-beta.0";
    const payload = {
      iss: "some_issuer",
      aud: "some/audience",
      job_workflow_ref: job_workflow_ref,
    };

    const jwt = `.${Buffer.from(JSON.stringify(payload)).toString("base64")}.`;
    core.getIDToken.mockClear();
    core.getIDToken.mockReturnValueOnce(jwt);

    const [repo, ref, workflow] =
      await detect.detectWorkflowFromOIDC("some/audience");
    expect(repo).toBe("vitejs/vite");
    expect(ref).toBe("refs/tags/create-vite@5.0.0-beta.0");
    expect(workflow).toBe(".github/workflows/publish.yml");
  });
  it("invalid audience", async () => {
    const job_workflow_ref =
      "octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main";
    const payload = {
      iss: "some_issuer",
      aud: "some/audience",
      job_workflow_ref: job_workflow_ref,
    };

    const jwt = `.${Buffer.from(JSON.stringify(payload)).toString("base64")}.`;
    core.getIDToken.mockClear();
    core.getIDToken.mockReturnValueOnce(jwt);

    await expect(
      detect.detectWorkflowFromOIDC("other/audience"),
    ).rejects.toThrow();
  });
  it("missing job_workflow_ref", async () => {
    const payload = {
      iss: "some_issuer",
      aud: "some/audience",
    };

    const jwt = `.${Buffer.from(JSON.stringify(payload)).toString("base64")}.`;
    core.getIDToken.mockClear();
    core.getIDToken.mockReturnValueOnce(jwt);

    await expect(
      detect.detectWorkflowFromOIDC("some/audience"),
    ).rejects.toThrow();
  });
});

jest.mock("@actions/github");

function createOctokitMock() {
  return {
    rest: {
      reactions: {
        createForIssueComment: jest.fn(),
      },
      actions: {
        listRepoWorkflows: jest.fn(),
        listWorkflowRuns: jest.fn(),
        getWorkflowRun: jest.fn(),
        reRunWorkflow: jest.fn(),
        listJobsForWorkflowRun: jest.fn(),
      },
      pulls: {
        get: jest.fn(),
      },
    },
  };
}

github.getOctokit.mockClear();
const octokit = createOctokitMock();
github.getOctokit.mockReturnValue(octokit);

describe("detectWorkflowFromContext", () => {
  it("no workflow run", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({ data: { conclusion: "failure" } }),
    );

    expect(
      detect.detectWorkflowFromContext("unused", "unused"),
    ).rejects.toThrow();
  });

  it("success", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          referenced_workflows: [
            {
              path: "project-oak/oak/.github/workflows/reusable_slsav1_provenance.yaml@b397021ff9b2c55236b9980698a189490eed2244",
              sha: "b397021ff9b2c55236b9980698a189490eed2244",
              ref: "refs/pull/3669/merge",
            },
            {
              path: "slsa-framework/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@v1.5.0",
              sha: "7f4fdb871876c23e455853d694197440c5a91506",
              ref: "refs/tags/v1.5.0",
            },
          ],
        },
      }),
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused",
    );
    expect(repo).toBe("slsa-framework/slsa-github-generator");
    expect(ref).toBe("refs/tags/v1.5.0");
    expect(workflow).toBe(".github/workflows/builder_docker-based_slsa3.yml");
  });

  it("multiple workflows", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          referenced_workflows: [
            {
              path: "slsa-framework/slsa-github-generator/.github/workflows/generic_generator_slsa3.yml@v1.5.0",
              sha: "7f4fdb871876c23e455853d694197440c5a91506",
              ref: "refs/tags/v1.5.0",
            },
            {
              path: "slsa-framework/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@v1.5.0",
              sha: "7f4fdb871876c23e455853d694197440c5a91506",
              ref: "refs/tags/v1.5.0",
            },
          ],
        },
      }),
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused",
    );
    expect(repo).toBe("slsa-framework/slsa-github-generator");
    expect(ref).toBe("refs/tags/v1.5.0");
    expect(workflow).toBe(".github/workflows/builder_docker-based_slsa3.yml");
  });

  it("multiple workflows mismatch", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          referenced_workflows: [
            {
              path: "slsa-framework/slsa-github-generator/.github/workflows/generic_generator_slsa3.yml@v1.5.0",
              sha: "7f4fdb871876c23e455853d694197440c5a91506",
              ref: "refs/tags/v1.5.0",
            },
            {
              path: "fork/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@v1.5.0",
              sha: "7f4fdb871876c23e455853d694197440c5a91506",
              ref: "refs/tags/v1.5.0",
            },
          ],
        },
      }),
    );
    expect(
      detect.detectWorkflowFromContext("unused", "unused"),
    ).rejects.toThrow();
  });

  it("success - PR on slsa-github-generator repo", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          event: "pull_request",
          head_sha: "088d04f305bd32ad4594d82e8c1571507acf03d5",
          path: ".github/workflows/pre-submit.e2e.docker-based.default.yml",
          referenced_workflows: [
            {
              path: "slsa-framework/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@9929152897cce5842f58221572911e18dd937808",
              sha: "9929152897cce5842f58221572911e18dd937808",
              ref: "refs/pull/3669/merge",
            },
          ],
          repository: {
            name: "slsa-github-generator",
            full_name: "slsa-framework/slsa-github-generator",
          },
          head_repository: {
            name: "slsa-github-generator",
            full_name: "asraa/slsa-github-generator",
          },
        },
      }),
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused",
    );
    expect(repo).toBe("asraa/slsa-github-generator");
    expect(ref).toBe("088d04f305bd32ad4594d82e8c1571507acf03d5");
    expect(workflow).toBe(
      ".github/workflows/pre-submit.e2e.docker-based.default.yml",
    );
  });
});

// mock `Octokit` from `@octokit/rest"`, doing pagination pass-through to our existing `octokit` mock
jest.mock("@octokit/rest", () => {
  return {
    Octokit: jest.fn().mockImplementation(() => {
      type octokitResponse = OctokitResponse<object, number>;
      type octokitArg = object;
      return {
        // re-use our existing mock
        ...octokit,
        // `paginate` takes a target method and it's arguments
        paginate: async (
          method: (arg: octokitArg) => Promise<octokitResponse>,
          arg: octokitArg,
        ): Promise<octokitResponse["data"]> => {
          // passthrough to the original method
          const resp = await method(arg);
          const data = resp.data;
          // paginator will always return the contents of `OctokitResponse.data`
          return data;
        },
      };
    }),
  };
});

describe("ensureOnlyGithubHostedRunners", () => {
  it("no workflow run", async () => {
    octokit.rest.actions.listJobsForWorkflowRun.mockReturnValue(
      Promise.resolve({ data: { conclusion: "failure" } }),
    );

    expect(
      detect.ensureOnlyGithubHostedRunners("unused", "unused"),
    ).rejects.toThrow();
  });

  it("success", async () => {
    octokit.rest.actions.listJobsForWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          total_count: 1,
          jobs: [
            {
              id: 399444496,
              run_id: 29679449,
              name: "myjob",
              labels: ["foo", "bar"],
            },
          ],
        },
      }),
    );
    expect(
      detect.ensureOnlyGithubHostedRunners("unused", "unused"),
    ).resolves.toBeUndefined();
  });

  it("failure", async () => {
    octokit.rest.actions.listJobsForWorkflowRun.mockReturnValue(
      Promise.resolve({
        data: {
          total_count: 1,
          jobs: [
            {
              id: 399444496,
              run_id: 29679449,
              name: "myjob",
              labels: ["self-hosted", "foo", "bar"],
            },
            {
              id: 399444497,
              run_id: 29679449,
              name: "otherjob",
              labels: ["self-hosted", "baz"],
            },
          ],
        },
      }),
    );
    expect(
      detect.ensureOnlyGithubHostedRunners("unused", "unused"),
    ).rejects.toThrow(
      new Error(`Self-hosted Runners are not allowed: jobs: myjob,otherjob`),
    );
  });
});
