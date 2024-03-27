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

import { Octokit } from "@octokit/rest";

export async function getJobLabels(
  octokit: Octokit,
  owner: string,
  repo: string,
): Promise<Set<string>> {
  const jobs = await octokit.paginate(
    octokit.rest.actions.listJobsForWorkflowRun,
    {
      owner,
      repo,
      run_id: Number(process.env.GITHUB_RUN_ID),
    },
  );
  const jobLabels: Set<string> = new Set<string>(
    jobs.map((job) => job.labels).flat(),
  );
  return jobLabels;
}

export async function getSelfHostedRunnerLabels(
  octokit: Octokit,
  owner: string,
  repo: string,
): Promise<Set<string>> {
  const selfHostedRunners = await octokit.paginate(
    octokit.rest.actions.listSelfHostedRunnersForRepo,
    {
      owner,
      repo,
    },
  );
  const selfHostedRunnerLabels: Set<string> = new Set<string>(
    selfHostedRunners
      .map((runner) => runner.labels.map((label) => label.name))
      .flat(),
  );
  return selfHostedRunnerLabels;
}
