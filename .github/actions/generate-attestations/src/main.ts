import * as core from "@actions/core";
import { writeAttestations } from "./attestation";
import path from "path";
import * as tscommon from "tscommon";

export function run(): void {
  try {
    /*
    Test:
      env INPUT_SLSA-LAYOUT-FILE=layout.json \
      INPUT_PREDICATE-TYPE=https://slsa.dev/provenance/v1.0?draft \
      INPUT_PREDICATE-FILE=predicate.json \
      INPUT_OUTPUT-FOLDER=out-folder \
      GITHUB_WORKSPACE=$PWD \
      nodejs ./dist/index.js
    */

    // SLSA subjects layout file.
    const slsaOutputs = core.getInput("slsa-layout-file");
    core.debug(`Using SLSA output file at ${slsaOutputs}!`);

    // Predicate.
    const predicateFile = core.getInput("predicate-file");
    core.debug(`Inputs: Predicate file ${predicateFile}!`);

    // Predicate type
    const predicateType = core.getInput("predicate-type");
    core.debug(`Inputs: Predicate type ${predicateType}!`);

    // Attach subjects and generate attestation files
    const outputFolder = core.getInput("output-folder");
    const attestations = writeAttestations(
      slsaOutputs,
      predicateType,
      predicateFile
    );

    // Write attestations
    tscommon.safeMkdirSync(outputFolder, { recursive: true });
    for (const att in attestations) {
      const outputFile = path.join(outputFolder, att);
      tscommon.safeWriteFileSync(outputFile, attestations[att]);
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
