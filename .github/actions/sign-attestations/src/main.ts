import * as core from "@actions/core";
import * as fs from "fs";
import * as sigstore from "sigstore";
import * as path from "path";

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
  rekorBaseURL: sigstore.sigstore.DEFAULT_REKOR_BASE_URL,
};

async function run(): Promise<void> {
  try {
    // Attestations
    const attestationFolder = core.getInput("attestations");
    const payloadType = core.getInput("payload-type");

    // Output folder
    const outputFolder = core.getInput("output-folder");

    const files = await fs.promises.readdir(attestationFolder);
    fs.mkdirSync(outputFolder, { recursive: true });
    for (const file of files) {
      const fpath = path.join(attestationFolder, file);
      const stat = await fs.promises.stat(fpath);
      if (stat.isFile()) {
        console.log("Signing '%s'.", fpath);
        const buffer = fs.readFileSync(fpath);
        const bundle = await sigstore.sigstore.signAttestation(
          buffer,
          payloadType,
          signOptions
        );
        const bundleStr = JSON.stringify(bundle);
        const outputPath = path.join(
          outputFolder,
          path.basename(file),
          ".sigstore"
        );
        fs.writeFileSync(outputPath, bundleStr);
        console.log("Wrote signed attestation to '%s'.", outputPath);
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
