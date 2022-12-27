import * as types from "./intoto";
import fs from "fs";
import path from "path";

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

export async function writeAttestations(
  layoutFile: string,
  predicateType: string,
  predicateFile: string
): Promise<Record<string, string>> {
  // Read SLSA output layout file.
  const buffer = fs.readFileSync(layoutFile);
  const layout: types.Layout = JSON.parse(buffer.toString());

  if (layout.version !== 1) {
    throw Error(`SLSA outputs layout invalid version: ${layout.version}`);
  }

  const count = Object.keys(layout.attestations).length;
  if (count > MAX_ATTESTATION_COUNT) {
    throw Error(`SLSA outputs layout had too many attestations: ${count}`);
  }

  // Read predicate
  const predicateBuffer = fs.readFileSync(predicateFile);
  const predicateJson = JSON.parse(predicateBuffer.toString());

  // TODO: Add other predicate validations.

  // Iterate through SLSA output layout and create attestations
  const ret: Record<string, string> = {};
  for (const att of layout.attestations) {
    // Validate that attestation path is not nested.
    if (path.dirname(att.name) !== ".") {
      throw Error(`attestation filename must not be nested ${att}`);
    }

    // TODO: How to cast directly into types.Subject[]?
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
