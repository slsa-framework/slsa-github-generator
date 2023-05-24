// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
