import * as core from "@actions/core";
import * as fs from "fs";
import * as sigstore from "sigstore";
import * as path from "path";

const signOptions = {
  oidcClientID: "sigstore",
  oidcIssuer: "https://oauth2.sigstore.dev/auth",
  rekorBaseURL: sigstore.sigstore.DEFAULT_REKOR_BASE_URL,
};

// Detect directory traversal for input file.
function resolvePathInput(input: string, wd: string): string {
  const safeJoin = path.resolve(path.join(wd, input));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}

async function run(): Promise<void> {
  try {
    const wd = process.env.GITHUB_WORKSPACE;
    if (!wd) {
      core.setFailed("No repository detected.");
      return;
    }

    // Attestations
    const attestationFolder = core.getInput("attestations");
    const safeAttestationFolder = resolvePathInput(attestationFolder, wd);
    const payloadType = core.getInput("payload-type");

    // Output folder
    const outputFolder = core.getInput("output-folder");
    const safeOutputFolder = resolvePathInput(outputFolder, wd);
    fs.mkdirSync(safeOutputFolder, { recursive: true });

    const files = await fs.promises.readdir(safeAttestationFolder);
    for (const file of files) {
      const fpath = resolvePathInput(
        path.join(safeAttestationFolder, file),
        wd
      );
      const stat = await fs.promises.stat(fpath);
      if (stat.isFile()) {
        core.debug(`Signing ${fpath}...`);
        const buffer = fs.readFileSync(fpath);
        const bundle = await sigstore.sigstore.signAttestation(
          buffer,
          payloadType,
          signOptions
        );
        const bundleStr = JSON.stringify(bundle);
        // We detect path traversal for safeOutputFolder, so this should be safe.
        const outputPath = path.join(
          safeOutputFolder,
          path.basename(file),
          ".sigstore"
        );
        fs.writeFileSync(outputPath, bundleStr, {
          flag: "ax",
          mode: 0o600,
        });
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
