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

import { getInput, setFailed, info } from "@actions/core";
import { Octokit } from "@octokit/rest";
import { getJobLabels, getSelfHostedRunnerLabels } from "./labels";

/**
Confirm that all jobs in the calling workflow are not using self-hosted Runners.
This check is only needed when calling from the generic builders, check if we're using a "generic builder", 
where the user may attempt to supply artifacts produced on self-hosted runners.

This check requires the caller to specify a github token with an additional `administration:read` permissions.
example:
```
uses: slsa-framewrok/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml'
  secrets:
    token: ${{ secrets.MY_TOKEN_WITH_EXTRA_PERM }}
 ```
*/
export async function getUsedSelfHostedRunnerLabels(
  token: string,
  owner: string,
  repo: string,
): Promise<string[]> {
  const octokit = new Octokit({ auth: token });
  const [jobLabels, selfHostedRunnerLabels] = await Promise.all([
    getJobLabels(octokit, owner, repo),
    getSelfHostedRunnerLabels(octokit, owner, repo),
  ]);
  const commonLabels = [...jobLabels].filter((label) =>
    selfHostedRunnerLabels.has(label),
  );
  return commonLabels;
}

export async function run(): Promise<void> {
  const token = getInput("token");
  let labels: string[] = [];
  try {
    const [owner, repo] = `${process.env.GITHUB_REPOSITORY}`.split("/");
    labels = await getUsedSelfHostedRunnerLabels(token, owner, repo);
  } catch (error) {
    if (error instanceof Error) {
      setFailed(error.message);
    } else {
      setFailed(`Unexpected error: ${error}`);
    }
  }
  if (labels.length) {
    setFailed(
      `Self-hosted runners are not allowed in SLSA Level 3 workflows. labels: ${labels}`,
    );
  } else {
    info("No self-hosted runners detected");
  }
}

run();
