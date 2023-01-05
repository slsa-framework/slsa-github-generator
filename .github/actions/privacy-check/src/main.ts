import * as github from "@actions/github";
import * as core from "@actions/core";

import { privacyCheck } from "./private.ts";

async function run(): Promise<void> {
  const override = core.getInput("override") === "true";
  const message = core.getInput("error_message");
  const token = core.getInput("token");

  const repoName = process.env.GITHUB_REPOSITORY;
  const [priv, check] = privacyCheck(repoName, token, override);
  core.setOutput("is_private", priv);
  if (!check) {
    core.setFailed(`${repoName}: ${message}`);
  }
}
run();
