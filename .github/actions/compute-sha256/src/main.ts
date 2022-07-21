import * as core from "@actions/core";
import * as fs from "fs";
import * as crypto from "crypto";

function shasum256(untrustedPath: string): string {
  if (!fs.existsSync(untrustedPath)) {
    throw new Error(`File ${untrustedPath} not present`);
  }
  const untrustedFile = fs.readFileSync(untrustedPath);
  return crypto.createHash("sha256").update(untrustedFile).digest("hex");
}

function run(): void {
  // Get the path to the untrusted file from ENV variable 'UNTRUSTED_PATH'
  const untrustedPath = core.getInput("path");
  core.info(`Computing sha256 of ${untrustedPath}`);
  const sha = shasum256(untrustedPath);
  core.info(`Computed sha256 of ${untrustedPath} as ${sha}`);
  core.setOutput("sha256", sha);
}
run();
