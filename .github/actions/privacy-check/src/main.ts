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

import * as github from "@actions/github";
import * as core from "@actions/core";

async function run(): Promise<void> {
  const override = core.getInput("override") === "true";
  const message = core.getInput("error_message");
  const token = core.getInput("token");
  const octokit = github.getOctokit(token);

  const repoName = process.env.GITHUB_REPOSITORY;
  if (!repoName) {
    core.setFailed("No repository detected.");
    return;
  }

  const parts = repoName.split("/");
  const owner = parts[0];
  const repo = parts[1];

  const repoResp = await octokit.rest.repos.get({
    owner,
    repo,
  });

  core.setOutput("is_private", repoResp.data.private);

  if (repoResp.data.private && !override) {
    core.setFailed(`${repoName}: ${message}`);
    return;
  }
}
run();
