import { githubObj } from "./types";

// createURI creates the fully qualified URI out of the repository
export function createURI(repository: string, ref: string): string {
  if (!repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  let refVal = "";
  if (ref) {
    refVal = `@${ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}

// getWorkflowPath returns the workflow's path from the workflow_ref.
export function getWorkflowPath(obj: githubObj): string {
  // GITHUB_WORKFLOW_REF contains the repository name in the path. We will trim
  // it out.
  // e.g. 'octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch'
  // Strip off the repo name and git ref from the workflow path.
  return obj.workflow_ref
    .substring(`${obj.repository}/`.length)
    .split("@", 1)[0];
}
