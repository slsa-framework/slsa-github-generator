/**
 * @fileoverview Tests for utils.ts
 */

import { getWorkflowPath } from "../src/utils";
import { githubObj } from "../src/types";

describe("getWorkflowPath", () => {
  it("extracts path from workflow ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_path: "",
      job: "",
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

    const path = getWorkflowPath(obj);
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });

  it("extracts path from workflow ref without git ref", () => {
    const obj: githubObj = {
      actor_id: "",
      event_name: "",
      event_path: "",
      job: "",
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

    const path = getWorkflowPath(obj);
    expect(path).toEqual(".github/workflows/my-workflow.yml");
  });
});
