import * as core from "@actions/core";
import fs from "fs";
import path from "path";
import { writeAttestations } from "./attestation";

// Detect directory traversal for input file.
function resolvePathInput(input: string, wd: string): string {
  const safeJoin = path.join(wd, input);
  if (!safeJoin.startsWith(wd)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}

export async function run(): Promise<void> {
  try {
    const wd = process.env[`GITHUB_WORKSPACE`] || "";

    // SLSA subjects layout file.
    const slsaOutputs = core.getInput("slsa-outputs-file");
    const safeSlsaOutputs = resolvePathInput(slsaOutputs, wd);
    core.debug(`Using SLSA output file at ${safeSlsaOutputs}!`);

    // Predicate.
    const predicateFile = core.getInput("predicate-file");
    const safePredicateFile = resolvePathInput(predicateFile, wd);
    core.debug(`Inputs: Predicate file ${safePredicateFile}!`);

    // Predicate type
    const predicateType = core.getInput("predicate-type");
    core.debug(`Inputs: Predicate type ${predicateType}!`);

    // Attach subjects and generate attestation files
    const outputFolder = core.getInput("output-folder");
    const attestations = await writeAttestations(
      safeSlsaOutputs,
      predicateType,
      safePredicateFile
    );

    // Write attestations
    fs.mkdirSync(outputFolder, { recursive: true });
    for (const att in attestations) {
      const outputFile = `${outputFolder}/${att}`;
      fs.writeFileSync(outputFile, attestations[att]);
    }

    core.setOutput("output-folder", outputFolder);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.info(`Unexpected error: ${error}`);
    }
  }
}

run();
