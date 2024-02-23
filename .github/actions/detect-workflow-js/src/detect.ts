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

import type { Endpoints } from "@octokit/types";
import * as github from "@actions/github";
import * as core from "@actions/core";
import { Octokit as OctokitRest } from "@octokit/rest";

type ApiWorkflowRun =
  Endpoints["GET /repos/{owner}/{repo}/actions/runs/{run_id}"]["response"]["data"];

type githubClaimsType = {
  aud?: string;
  job_workflow_ref?: string;
};

export function decodeToken(federatedToken: string): githubClaimsType {
  const tokenPayload = federatedToken.split(".")[1];
  const bufferObj = Buffer.from(tokenPayload, "base64");
  const decoded = JSON.parse(bufferObj.toString("utf8"));
  return decoded;
}

export async function detectWorkflowFromOIDC(
  aud: string,
): Promise<[string, string, string]> {
  const id_token = await core.getIDToken(aud);
  const decoded = decodeToken(id_token);
  if (!decoded.aud || decoded.aud !== aud) {
    return Promise.reject(Error("invalid audience from OIDC token."));
  }

  // Use job_workflow_ref to extract the outputs.
  const jobWorkflowRef = decoded.job_workflow_ref;
  if (!jobWorkflowRef) {
    return Promise.reject(Error("job_workflow_ref missing from OIDC token."));
  }

  // In some cases, the job_workflow_ref field may contain multiple `@`s
  // (e.g. `vitejs/vite/.github/workflows/publish.yml@refs/tags/create-vite@5.0.0-beta.0`).
  // In this case, the workflow ref contains an `@`, so we can't simply use `.split`.
  const firstAtIndex = jobWorkflowRef.indexOf("@");
  const workflowPath = jobWorkflowRef.slice(0, firstAtIndex);
  const workflowRef = jobWorkflowRef.slice(firstAtIndex + 1);
  const [workflowOwner, workflowRepo, ...workflowArray] =
    workflowPath.split("/");
  const repository = [workflowOwner, workflowRepo].join("/");
  const workflow = workflowArray.join("/");
  return [repository, workflowRef, workflow];
}

export async function detectWorkflowFromContext(
  repoName: string,
  token: string,
): Promise<[string, string, string]> {
  const [owner, repo] = repoName.split("/");
  const octokit = github.getOctokit(token);
  const res = await octokit.rest.actions.getWorkflowRun({
    owner,
    repo,
    run_id: Number(process.env.GITHUB_RUN_ID),
  });
  const workflowData: ApiWorkflowRun = res.data;
  core.info(`workflow data: ${JSON.stringify(workflowData)}`);

  if (!workflowData.referenced_workflows) {
    return Promise.reject(
      Error(`No reusable workflows detected ${JSON.stringify(workflowData)}.`),
    );
  }

  let [repository, ref, workflow] = ["", "", ""];

  // If this is a pull request on the main repository
  // (slsa-framework/slsa-github-generator), then look for the repo and head
  // SHA from the pull_request event value. Pull requests on forks are not
  // supported.
  if (
    (workflowData.event === "pull_request" ||
      workflowData.event === "merge_group") &&
    workflowData.repository.full_name === "slsa-framework/slsa-github-generator"
  ) {
    ref = workflowData.head_sha;
    repository = workflowData.head_repository.full_name;
    workflow = workflowData.path;
  } else {
    // Otherwise this is an external repository.
    // Filter referenced_workflows for slsa-github-generator repositories.
    // TODO(https://github.com/actions/runner/issues/2417): When
    // GITHUB_JOB_WORKFLOW_SHA becomes fully functional, the OIDC token
    // detection can be removed and we can identify the current reusable
    // workflow through the sha of a referenced workflow, fully supporting all
    // triggers without the repository filter.
    for (const reusableWorkflow of workflowData.referenced_workflows) {
      const workflowPath = reusableWorkflow.path.split("@", 1);
      const [workflowOwner, workflowRepo, ...workflowArray] =
        workflowPath[0].split("/");
      if (workflowRepo === "slsa-github-generator") {
        if (!reusableWorkflow.ref) {
          return Promise.reject(
            Error(
              "Referenced slsa-github-generator workflow missing ref: was the workflow invoked by digest?",
            ),
          );
        }
        const tmpRepository = [workflowOwner, workflowRepo].join("/");
        const tmpRef = reusableWorkflow.ref;
        const tmpWorkflow = workflowArray.join("/");
        // If there are multiple invocations of reusable workflows in
        // a single caller workflow, ensure that the repositories and refs are
        // the same.
        if (repository !== "" && repository !== tmpRepository) {
          return Promise.reject(Error("Unexpected mismatch of repositories"));
        }
        if (ref !== "" && ref !== tmpRef) {
          return Promise.reject(Error("Unexpected mismatch of reference"));
        }
        repository = tmpRepository;
        ref = tmpRef;
        workflow = tmpWorkflow;
      }
    }
  }

  return [repository, ref, workflow];
}

export async function ensureOnlyGithubHostedRunners(
  repoName: string,
  token: string,
): Promise<void> {
  // We need to confirm that all jobs in the Workflow are using Github-hosted Runners.
  const [owner, repo] = repoName.split("/");
  // const octokit = github.getOctokit(token);
  const octokit = new OctokitRest({ auth: token });
  const jobsData = await octokit.paginate(
    octokit.rest.actions.listJobsForWorkflowRun,
    {
      owner,
      repo,
      run_id: Number(process.env.GITHUB_RUN_ID),
    },
  );
  const selfHostedLabel = "self-hosted";
  const jobsUsingSelfHostedRunners: string[] = [];
  for (const job of jobsData.jobs) {
    if (job.labels.includes(selfHostedLabel)) {
      jobsUsingSelfHostedRunners.push(job.name);
    }
  }
  if (jobsUsingSelfHostedRunners.length) {
    return Promise.reject(
      Error(
        `Self-hosted Runners are not allowed: jobs: ${jobsUsingSelfHostedRunners}`,
      ),
    );
  }
  return;
}
