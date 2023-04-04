/*
Copyright 2022 SLSA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WIHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import * as tscommon from "tscommon";
import { githubObj, rawTokenInterface } from "../src/types";

export function validateGitHubFields(gho: githubObj): void {
  // actor_id
  validateField("github.actor_id", gho.actor_id, process.env.GITHUB_ACTOR_ID);

  // event_name
  validateField(
    "github.event_name",
    gho.event_name,
    process.env.GITHUB_EVENT_NAME
  );

  // event_payload_sha256
  const eventPath = process.env.GITHUB_EVENT_PATH || "";
  // NOTE: validate GITHUB_EVENT_PATH is non-empty to provide a better error
  // message.
  validateFieldNonEmpty("GITHUB_EVENT_PATH", eventPath);
  validateField(
    "github.event_payload_sha256",
    gho.event_payload_sha256,
    tscommon.safeFileSha256(eventPath)
  );

  // ref
  validateField("github.ref", gho.ref, process.env.GITHUB_REF);

  // ref_type
  validateField("github.ref_type", gho.ref_type, process.env.GITHUB_REF_TYPE);

  // repository
  validateField(
    "github.repository",
    gho.repository,
    process.env.GITHUB_REPOSITORY
  );

  // repository_id
  validateField(
    "github.repository_id",
    gho.repository_id,
    process.env.GITHUB_REPOSITORY_ID
  );

  // repository_owner_id
  validateField(
    "github.repository_owner_id",
    gho.repository_owner_id,
    process.env.GITHUB_REPOSITORY_OWNER_ID
  );

  // run_attempt
  validateField(
    "github.run_attempt",
    gho.run_attempt,
    process.env.GITHUB_RUN_ATTEMPT
  );

  // run_id
  validateField("github.run_id", gho.run_id, process.env.GITHUB_RUN_ID);

  // run_number
  validateField(
    "github.run_number",
    gho.run_number,
    process.env.GITHUB_RUN_NUMBER
  );

  // sha
  validateField("github.sha", gho.sha, process.env.GITHUB_SHA);

  // workflow_ref
  validateField(
    "github.workflow_ref",
    gho.workflow_ref,
    process.env.GITHUB_WORKFLOW_REF
  );
  validateFieldStartsWith(
    "github.workflow_ref",
    gho.workflow_ref,
    `${process.env.GITHUB_REPOSITORY}/`
  );

  // workflow_sha
  validateField(
    "github.workflow_sha",
    gho.workflow_sha,
    process.env.GITHUB_WORKFLOW_SHA
  );
}

export function validateAndMaskInputs(
  token: rawTokenInterface
): rawTokenInterface {
  const maskedMapInputs = new Map(Object.entries(token.tool.inputs));
  const toolInputs = token.tool.masked_inputs;
  if (
    toolInputs === undefined ||
    // If TRW provides an empty argument, it's a 1-length array
    // with an empty string value.
    (toolInputs.length === 1 && toolInputs[0].length === 0)
  ) {
    token.tool.inputs = maskedMapInputs;
    return token;
  }

  for (const key of toolInputs) {
    // verify non-empty keys.
    if (key === undefined || key.trim().length === 0) {
      throw new Error("empty key in the input map");
    }

    if (!maskedMapInputs.has(key)) {
      throw new Error(`input '${key}' does not exist in the input map`);
    }

    // NOTE: This mask is the same used by GitHub for encrypted secrets and masked values.
    maskedMapInputs.set(key, "***");
  }
  token.tool.inputs = maskedMapInputs;
  return token;
}

export function validateFieldAnyOf<T>(
  name: string,
  actual: T,
  expected: T[]
): void {
  for (const value of expected) {
    if (actual === value) {
      // Found a match.
      return;
    }
  }
  throw new Error(
    `mismatch ${name}: got '${actual}', expected one of '${expected.join(
      ","
    )}'.`
  );
}

/**
 * validateField validates that the value of the named field matches the
 * expected value and is non-empty.
 * @param name - the name of the value
 * @param actual - the actual value of the field
 * @param expected - the expected value of the field
 * @throws Error - if actual and expected don't match or are empty.
 */
export function validateField<T>(name: string, actual: T, expected: T): void {
  if (actual !== expected) {
    throw new Error(
      `mismatch ${name}: got '${actual}', expected '${expected}'.`
    );
  }
  if (!actual) {
    throw new Error(`empty ${name}, expected non-empty value.`);
  }
}

export function validateFieldStartsWith(
  name: string,
  actual: string,
  prefix: string
): void {
  if (!actual.startsWith(prefix)) {
    throw new Error(
      `invalid ${name}: expected '${actual}' to start with '${prefix}'.`
    );
  }
}

/**
 * validateFieldNonEmpty validates that the value of the named field is not
 * empty.
 * @param name - the name of the value
 * @param actual - the actual value of the field
 * @throws Error - if actual is empty.
 */
export function validateFieldNonEmpty(name: string, actual: string): void {
  if (!actual) {
    throw new Error(`empty ${name}, expected non-empty value.`);
  }
}
