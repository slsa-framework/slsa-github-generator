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
