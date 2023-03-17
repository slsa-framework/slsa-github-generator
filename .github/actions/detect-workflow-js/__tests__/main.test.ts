import { create } from "domain";

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

    const [repo, ref, workflow] = await detect.detectWorkflowFromOIDC(
      "some/audience"
    );
    expect(repo).toBe("octo-org/octo-automation");
    expect(ref).toBe("refs/heads/main");
    expect(workflow).toBe(".github/workflows/oidc.yml");
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
      detect.detectWorkflowFromOIDC("other/audience")
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
      detect.detectWorkflowFromOIDC("some/audience")
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
      },
      pulls: {
        get: jest.fn(),
      },
    },
  };
}

describe("detectWorkflowFromContext", () => {
  github.getOctokit.mockClear();
  const octokit = createOctokitMock();
  github.getOctokit.mockReturnValue(octokit);

  it("no workflow run", async () => {
    octokit.rest.actions.getWorkflowRun.mockReturnValue(
      Promise.resolve({ data: { conclusion: "failure" } })
    );

    expect(
      detect.detectWorkflowFromContext("unused", "unused")
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
      })
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused"
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
      })
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused"
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
      })
    );
    expect(
      detect.detectWorkflowFromContext("unused", "unused")
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
            }
          ],
          repository: {
            name: "slsa-github-generator",
            full_name: "slsa-framework/slsa-github-generator"
          },
          head_repository: {
            name: "slsa-github-generator",
            full_name: "asraa/slsa-github-generator"
          }
        },
      })
    );
    const [repo, ref, workflow] = await detect.detectWorkflowFromContext(
      "unused",
      "unused"
    );
    expect(repo).toBe("asraa/slsa-github-generator");
    expect(ref).toBe("088d04f305bd32ad4594d82e8c1571507acf03d5");
    expect(workflow).toBe(".github/workflows/pre-submit.e2e.docker-based.default.yml");
  });
});
