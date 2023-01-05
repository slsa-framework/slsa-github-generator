export interface githubContext {
  event_name: string;
  run_attempt: string;
  run_id: string;
  run_number: string;
  workflow: string;
  sha: string;
  repository: string;
  repository_owner: string;
  // TODO(#1411): Record if these become available.
  // repository_id: string;
  // repository_owner_id: string;
  // actor_id: string;
  ref: string;
  ref_type: string;
  actor: string;
}

export function validateGitHubFields(gho: githubContext): void {
  validateField(
    "github.event_name",
    gho.event_name,
    process.env.GITHUB_EVENT_NAME
  );
  validateField(
    "github.run_attempt",
    gho.run_attempt,
    process.env.GITHUB_RUN_ATTEMPT
  );
  validateField("github.run_id", gho.run_id, process.env.GITHUB_RUN_ID);
  validateField(
    "github.run_number",
    gho.run_number,
    process.env.GITHUB_RUN_NUMBER
  );
  validateField("github.workflow", gho.workflow, process.env.GITHUB_WORKFLOW);
  validateField("github.sha", gho.sha, process.env.GITHUB_SHA);
  validateField(
    "github.repository",
    gho.repository,
    process.env.GITHUB_REPOSITORY
  );
  validateField(
    "github.repository_owner",
    gho.repository_owner,
    process.env.GITHUB_REPOSITORY_OWNER
  );
  validateField("github.ref", gho.ref, process.env.GITHUB_REF);
  validateField("github.ref_type", gho.ref_type, process.env.GITHUB_REF_TYPE);
  validateField("github.actor", gho.actor, process.env.GITHUB_ACTOR);
  // TODO(#1411): Record if these become available.
  // repository_id: process.env.GITHUB_REPOSITORY_ID,
  // repository_owner_id: process.env.GITHUB_REPOSITORY_OWNER_ID,
  // repository_actor_id: process.env.GITHUB_ACTOR_ID,
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

export function validateNonEmptyField(name: string, actual: string): void {
  if (actual === "") {
    throw new Error(`empty ${name}, expected non-empty value.`);
  }
}
