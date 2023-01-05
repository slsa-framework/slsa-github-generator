import * as github from "@actions/github";

/**
 * privacyCheck returns a two tuple of two boolean values. The first is true if
 * the repository is private. The second is true if the privacy check passes
 * (Repo is public or override is true).
 */
export async function privacyCheck(
  repoName: string,
  token: string,
  override: boolean
): Promise<[boolean, boolean]> {
  const priv = await repoIsPrivate(repoName, token);
  if (override) {
    return [priv, true];
  }
  return [priv, !priv];
}

/**
 * repoIsPrivate returns true if the repository is private.
 */
export async function repoIsPrivate(
  repoName: string,
  token: string
): Promise<boolean> {
  const octokit = github.getOctokit(token);
  if (!repoName) {
    throw new Error("No repository detected.");
  }

  const parts = repoName.split("/");
  const owner = parts[0];
  const repo = parts[1];

  const repoResp = await octokit.rest.repos.get({
    owner,
    repo,
  });

  return repoResp.data.private;
}
