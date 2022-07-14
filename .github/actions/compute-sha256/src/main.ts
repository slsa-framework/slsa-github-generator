import * as core from "@actions/core";
import * as fs from "fs";
import { execSync } from "child_process";

function shasum256(untrustedPath: string): string {
  if (!fs.existsSync(untrustedPath)) {
    throw new Error(`File ${untrustedPath} not present`);
  }
  const sha = execSync(`sha256sum ${untrustedPath}`).toString().split(" ")[0];
  return sha;
}

async function run(): Promise<void> {
  // Get the path to the untrusted file from ENV variable 'UNTRUSTED_PATH'
  const untrustedPath = core.getInput("untrusted_path");
  const sha = shasum256(untrustedPath);
  core.setOutput("sha256", sha);
}
run();
