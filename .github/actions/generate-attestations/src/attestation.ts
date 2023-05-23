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

import * as types from "./intoto";
import path from "path";
import * as tscommon from "tscommon";

// Maximum number of attestations to be written.
const MAX_ATTESTATION_COUNT = 50;

export function createStatement(
  subjects: types.Subject[],
  type: string,
  predicate: object
): types.IntotoStatement {
  return {
    _type: types.INTOTO_TYPE,
    subject: subjects,
    predicateType: type,
    predicate,
  };
}

export function writeAttestations(
  layoutFile: string,
  predicateType: string,
  predicateFile: string
): Record<string, string> {
  // Read SLSA output layout file.
  const buffer = tscommon.safeReadFileSync(layoutFile);
  const layout: types.Layout = JSON.parse(buffer.toString());

  if (layout.version !== 1) {
    throw Error(`SLSA outputs layout invalid version: ${layout.version}`);
  }

  const count = Object.keys(layout.attestations).length;
  if (count > MAX_ATTESTATION_COUNT) {
    throw Error(`SLSA outputs layout had too many attestations: ${count}`);
  }

  // Read predicate
  const predicateBuffer = tscommon.safeReadFileSync(predicateFile);
  const predicateJson = JSON.parse(predicateBuffer.toString());

  // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1422): Add other predicate validations.

  // Iterate through SLSA output layout and create attestations
  const ret: Record<string, string> = {};
  if (layout.attestations.length === 0) {
    throw Error("attestation list is empty");
  }
  for (const att of layout.attestations) {
    // Validate that attestation path is not nested.
    if (path.dirname(att.name) !== ".") {
      throw Error(`attestation filename must not be nested ${att}`);
    }

    const subjectJson: types.Subject[] = JSON.parse(
      JSON.stringify(att.subjects)
    );

    const attestationJSON = createStatement(
      subjectJson,
      predicateType,
      predicateJson
    );

    ret[att.name] = JSON.stringify(attestationJSON);
  }
  return ret;
}
