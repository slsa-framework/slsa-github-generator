import * as github from "@actions/github";
import * as core from "@actions/core";
import type { Endpoints } from "@octokit/types";
import path from "path";

type ApiWorkflowRun =
  Endpoints["GET /repos/{owner}/{repo}/actions/runs/{run_id}"]["response"]["data"];

type githubClaimsType = {
  aud?: string;
  job_workflow_ref?: string;
};

function decodeToken(federatedToken: string): githubClaimsType {
  const tokenPayload = federatedToken.split(".")[1];
  const bufferObj = Buffer.from(tokenPayload, "base64");
  const decoded = JSON.parse(bufferObj.toString("utf8"));
  return decoded;
}

async function run(): Promise<void> {
  /* Test locally. Requires a GitHub token:
        $ env INPUT_TOKEN="$(gh auth token)" \
        GITHUB_RUN_ID="4289844197" \
        GITHUB_REPOSITORY="konstruktoid/ansible-role-dns" \
        nodejs ./dist/index.js
    */
  const token = core.getInput("token");
  const octokit = github.getOctokit(token);

  const repoName = process.env.GITHUB_REPOSITORY;
  if (!repoName) {
    core.setFailed("No repository detected.");
    return;
  }

  // Set outputs.
  let repository = "";
  let ref = "";
  let workflow = "";

  try {
    // Use the OIDC token when available.
    const aud = path.join(repoName, "detect-workflow-js");
    const id_token = await core.getIDToken(aud);
    const decoded = decodeToken(id_token);
    if (!decoded.aud || decoded.aud !== aud) {
      core.setFailed("invalid audience from OIDC token.");
      return;
    }

    // Use job_workflow_ref to extract the outputs.
    const jobWorkflowRef = decoded.job_workflow_ref;
    if (!jobWorkflowRef) {
      core.setFailed("job_workflow_ref missing from OIDC token.");
      return;
    }

    const [workflowPath, workflowRef] = jobWorkflowRef.split("@", 2);
    const [workflowOwner, workflowRepo, ...workflowArray] =
      workflowPath.split("/");
    repository = [workflowOwner, workflowRepo].join("/");
    ref = workflowRef;
    workflow = workflowArray.join("/");
  } catch {
    // Otherwise, try to use the referenced workflows from the current workflow run.
    core.info(
      "Failed to retrieve OIDC token. This may be due to missing permissions."
    );
    const [owner, repo] = repoName.split("/");
    const res = await octokit.rest.actions.getWorkflowRun({
      owner,
      repo,
      run_id: Number(process.env.GITHUB_RUN_ID),
    });
    const workflowData: ApiWorkflowRun = res.data;

    if (!workflowData.referenced_workflows) {
      core.setFailed("No reusable workflows detected.");
      return;
    }

    // There is only one referenced reusable workflow, this must be it.
    // Otherwise, there are multiple reusable workflows referenced in the caller
    // workflow and we cannot yet determine which is the current one.
    // TODO(https://github.com/actions/runner/issues/2417): When
    // GITHUB_JOB_WORKFLOW_SHA becomes fully functional, the OIDC token
    // above can be removed and we can match the current reusable workflow
    // through the sha of a referenced workflow, fully supporting all triggers.
    if (workflowData.referenced_workflows.length === 1) {
      const reusableWorkflow = workflowData.referenced_workflows[0];
      const workflowPath = reusableWorkflow.path.split("@", 1);
      const [workflowOwner, workflowRepo, ...workflowArray] =
        workflowPath[0].split("/");
      repository = [workflowOwner, workflowRepo].join("/");
      ref = reusableWorkflow.ref || "";
      workflow = workflowArray.join("/");
    }
  }

  if (!repository) {
    core.setFailed("No repository detected.");
    return;
  }
  if (!ref) {
    core.setFailed("No ref detected.");
    return;
  }
  if (!workflow) {
    core.setFailed("No workflow detected.");
    return;
  }

  core.setOutput("repository", repository);
  core.setOutput("ref", ref);
  core.setOutput("workflow", workflow);
}
run();
