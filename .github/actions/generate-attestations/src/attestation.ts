import * as types from "./intoto";
import fs from "fs";
import path from "path";

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
  const layout = JSON.parse(buffer.toString());

  // Read predicate
  const predicateBuffer = fs.readFileSync(predicateFile);
  const predicateJson = JSON.parse(predicateBuffer.toString());

  // Iterate through SLSA output layout and create attestations
  const ret: Record<string, string> = {};
  for (const att in layout) {
    if (att !== "version") {
      // Validate that attestation path is not nested.
      if (path.dirname(att) !== ".") {
        throw Error(`attestation filename must not be nested ${att}`);
      }

      // Validate that attestation is not already written.
      if (ret[att] !== undefined) {
        throw Error(`duplicate attestation file ${att}`);
      }

      // TODO: How to cast directly into types.Subject[]?
      const subjectJson: types.Subject[] = JSON.parse(
        JSON.stringify(layout[att])
      );

      const attestationJSON = createStatement(
        subjectJson,
        predicateType,
        predicateJson
      );

      ret[att] = JSON.stringify(attestationJSON, null, "\t");
    }
  }
  return ret;
}
