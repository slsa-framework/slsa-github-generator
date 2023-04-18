import * as core from "@actions/core";
import * as tscommon from "tscommon";

function run(): void {
  // Get the path to the untrusted file from ENV variable 'UNTRUSTED_PATH'
  const untrustedPath = core.getInput("path");
  core.info(`Computing sha256 of ${untrustedPath}`);
  const sha = tscommon.safeFileSha256(untrustedPath);
  core.info(`Computed sha256 of ${untrustedPath} as ${sha}`);
  core.setOutput("sha256", sha);
}
run();
