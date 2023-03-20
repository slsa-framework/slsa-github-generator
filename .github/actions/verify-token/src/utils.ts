import path from "path";
import { githubObj } from "./types";

export function getEnv(name: string): string {
  const res = process.env[name];
  if (!res) {
    throw new Error(`missing env: ${name}`);
  }
  return String(res);
}

export function resolvePathInput(untrustedInput: string, wd: string): string {
  const safeJoin = path.resolve(path.join(wd, untrustedInput));
  if (!(safeJoin + path.sep).startsWith(wd + path.sep)) {
    throw Error(`unsafe path ${safeJoin}`);
  }
  return safeJoin;
}

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
