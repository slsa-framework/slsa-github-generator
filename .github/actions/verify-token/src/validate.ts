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

import { githubObj } from "../src/types";

export function validateGitHubFields(gho: githubObj): void {
  // actor
  validateFieldNonEmpty("github.actor", gho.actor);
  validateField("github.actor", gho.actor, process.env.GITHUB_ACTOR);

  // actor_id
  validateFieldNonEmpty("github.actor_id", gho.actor_id);
  validateField("github.actor_id", gho.actor_id, process.env.GITHUB_ACTOR_ID);

  // event_name
  validateFieldNonEmpty("github.event_name", gho.event_name);
  validateField(
    "github.event_name",
    gho.event_name,
    process.env.GITHUB_EVENT_NAME
  );

  // event_path
  validateFieldNonEmpty("github.event_path", gho.event_path);
  validateField(
    "github.event_path",
    gho.event_path,
    process.env.GITHUB_EVENT_PATH
  );

  // job
  validateFieldNonEmpty("github.job", gho.job);
  validateField("github.job", gho.job, process.env.GITHUB_JOB);

  // ref
  validateFieldNonEmpty("github.ref", gho.ref);
  validateField("github.ref", gho.ref, process.env.GITHUB_REF);

  // ref_type
  validateFieldNonEmpty("github.ref_type", gho.ref_type);
  validateField("github.ref_type", gho.ref_type, process.env.GITHUB_REF_TYPE);

  // repository
  validateFieldNonEmpty("github.repository", gho.repository);
  validateField(
    "github.repository",
    gho.repository,
    process.env.GITHUB_REPOSITORY
  );

  // repository_id
  validateFieldNonEmpty("github.repository_id", gho.repository_id);
  validateField(
    "github.repository_id",
    gho.repository_id,
    process.env.GITHUB_REPOSITORY_ID
  );

  // repository_owner
  validateFieldNonEmpty("github.repository_owner", gho.repository_owner);
  validateField(
    "github.repository_owner",
    gho.repository_owner,
    process.env.GITHUB_REPOSITORY_OWNER
  );

  // repository_owner_id
  validateFieldNonEmpty("github.repository_owner_id", gho.repository_owner_id);
  validateField(
    "github.repository_owner_id",
    gho.repository_owner_id,
    process.env.GITHUB_REPOSITORY_OWNER_ID
  );

  // run_attempt
  validateFieldNonEmpty("github.run_attempt", gho.run_attempt);
  validateField(
    "github.run_attempt",
    gho.run_attempt,
    process.env.GITHUB_RUN_ATTEMPT
  );

  // run_id
  validateFieldNonEmpty("github.run_id", gho.run_id);
  validateField("github.run_id", gho.run_id, process.env.GITHUB_RUN_ID);

  // run_number
  validateFieldNonEmpty("github.run_number", gho.run_number);
  validateField(
    "github.run_number",
    gho.run_number,
    process.env.GITHUB_RUN_NUMBER
  );

  // sha
  validateFieldNonEmpty("github.sha", gho.sha);
  validateField("github.sha", gho.sha, process.env.GITHUB_SHA);

  // workflow_ref
  validateFieldNonEmpty("github.workflow_ref", gho.workflow_ref);
  validateField(
    "github.workflow_ref",
    gho.workflow_ref,
    process.env.GITHUB_WORKFLOW_REF
  );

  // workflow_sha
  validateFieldNonEmpty("github.workflow_sha", gho.workflow_sha);
  validateField(
    "github.workflow_sha",
    gho.workflow_sha,
    process.env.GITHUB_WORKFLOW_SHA
  );
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

export function validateField<T>(name: string, actual: T, expected: T): void {
  if (actual !== expected) {
    throw new Error(
      `mismatch ${name}: got '${actual}', expected '${expected}'.`
    );
  }
}

export function validateFieldNonEmpty(name: string, actual: string): void {
  if (actual === "" || actual === null || actual === undefined) {
    throw new Error(`empty ${name}, expected non-empty value.`);
  }
}
