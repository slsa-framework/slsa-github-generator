/**
 * @fileoverview Tests for predicate.ts
 */

import { getWorkflowPath } from "../src/predicate";

describe("getWorkflowPath", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
  });

  afterEach(() => {
    process.env = env;
  });

  it("extracts path from workflow ref", () => {
    process.env.GITHUB_WORKFLOW_REF =
      "octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch";
    process.env.GITHUB_REPOSITORY = "octocat/hello-world";

    const path = getWorkflowPath();
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });

  it("extracts path from workflow ref without git ref", () => {
    process.env.GITHUB_WORKFLOW_REF =
      "octocat/hello-world/.github/workflows/my-workflow.yml";
    process.env.GITHUB_REPOSITORY = "octocat/hello-world";

    const path = getWorkflowPath();
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });

  it("fails if workflow ref does not start with repository", () => {
    process.env.GITHUB_WORKFLOW_REF =
      ".github/workflows/my-workflow.yml@ref/heads/my_branch";
    process.env.GITHUB_REPOSITORY = "octocat/hello-world";

    expect(() => {
      const path = getWorkflowPath();
    }).toThrow();
  });
});
