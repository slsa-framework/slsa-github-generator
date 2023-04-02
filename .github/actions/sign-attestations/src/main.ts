import * as core from "@actions/core";
import { sigstore } from "sigstore";
import * as path from "path";
import * as tscommon from "tscommon";

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
};

async function run(): Promise<void> {
  try {
    /* Test locally:
        $ env INPUT_ATTESTATIONS="testdata/attestations" \
        INPUT_OUTPUT-FOLDER="outputs" \
        GITHUB_WORKSPACE="$(pwd)" \
        nodejs ./dist/index.js
    */

    // Attestations
    const attestationFolder = core.getInput("attestations");
    const payloadType = core.getInput("payload-type");

    // Output folder
    const outputFolder = core.getInput("output-folder");
    tscommon.safeMkdirSync(outputFolder, { recursive: true });

    const files = await tscommon.safePromises_readdir(attestationFolder);
    for (const file of files) {
      const fpath = path.join(attestationFolder, file);
      const stat = await tscommon.safePromises_stat(fpath);
      if (stat.isFile()) {
        core.debug(`Signing ${fpath}...`);
        const buffer = tscommon.safeReadFileSync(fpath);
        const bundle = await sigstore.attest(buffer, payloadType, signOptions);
        const bundleStr = JSON.stringify(bundle);
        const outputPath = path.join(
          outputFolder,
          `${path.basename(fpath)}.sigstore`
        );
        // We detect path traversal for outputPath in safeWriteFileSync.
        tscommon.safeWriteFileSync(outputPath, bundleStr);
        core.debug(`Wrote signed attestation to '${outputPath}.`);
      }
    }
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}

run();
