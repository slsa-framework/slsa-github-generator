/*
Copyright 2024 SLSA Authors
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

import * as github from "@actions/github";
import * as core from "@actions/core";
import * as inputs from "./inputs";
import * as attestation from "./attestation";
import * as crypto from "crypto";
import * as tscommon from "tscommon";

async function run(): Promise<void> {
  /* Test locally. Requires a GitHub token:
    $ env INPUT_SLSA-WORKFLOW-INPUTS='{"formats":"zip tar.gz"}' \
    GITHUB_REPOSITORY="slsa-framework/slsa-policy" \
    GITHUB_REF_NAME="v0.0.1" \
    SLSA_OUTPUTS_ARTIFACTS_FILE="output.json" \
    GITHUB_TOKEN="$(echo $GH_TOKEN)" \
    nodejs ./dist/index.js
  */
  try {
    // Get token.
    const ghToken = inputs.getEnvVariable("GITHUB_TOKEN");
    // Get workflow inputs.
    const workflowInputs = inputs.getActionInput("slsa-workflow-inputs");
    const jsonInputs = JSON.parse(workflowInputs);
    const f = jsonInputs["formats"];
    if (!f) {
      throw new Error("formats empty");
    }
    const formats = inputs.parseFormats(f);
    const apiFormats = inputs.formatsToAPI(formats);

    // Verify that upload-assets is set to true.
    const uploadAssets = jsonInputs["upload-assets"].trim();
    if (uploadAssets !== "true") {
      throw new Error(`expected upload-assets to be true, got ${uploadAssets}`);
    }

    // Get layout file.
    const layoutFile = inputs.getActionInput("slsa-layout-file");
    // See https://docs.github.com/en/actions/learn-github-actions/variables.
    // Get repository.
    const ghRepo = inputs.getEnvVariable("GITHUB_REPOSITORY");
    const parts = ghRepo.split("/");
    if (parts.length !== 2) {
      throw new Error(`malformed repo: ${ghRepo}`);
    }
    // Get ref.
    const ghRef = inputs.getEnvVariable("GITHUB_REF_NAME");

    core.debug(`slsa-workflow-inputs: ${workflowInputs}`);
    core.debug(`slsa-layout-file: ${layoutFile}`);
    core.debug(`formats: ${apiFormats}`);
    core.debug(`ref: ${ghRef}`);
    core.debug(`owner: ${parts[0]}`);
    core.debug(`repo: ${parts[1]}`);

    // Create octokit instance.
    const octokit = github.getOctokit(ghToken);
    let i = 0;
    const hashes: string[] = [];
    for (const format of apiFormats) {
      core.debug(`format: ${format}`);
      // https://docs.github.com/en/rest/repos/contents?apiVersion=2022-11-28#download-a-repository-archive-tar
      const result = await octokit.request(
        `GET /repos/{owner}/{repo}/${format}/{ref}`,
        {
          owner: parts[0],
          repo: parts[1],
          ref: ghRef,
          headers: {
            "X-GitHub-Api-Version": "2022-11-28",
          },
        },
      );
      // Compute the hash.
      hashes.push(
        crypto
          .createHash("sha256")
          .update(Buffer.from(result.data))
          .digest("hex"),
      );

      // Get release information.
      const releaseInfo = await octokit.request(
        "GET /repos/{owner}/{repo}/releases/tags/{tag}",
        {
          owner: parts[0],
          repo: parts[1],
          tag: ghRef,
          headers: {
            "X-GitHub-Api-Version": "2022-11-28",
          },
        },
      );
      core.debug(`release.id: ${releaseInfo.data.id}`);
      // NOTE: The code snippet from https://docs.github.com/en/rest/releases/assets?apiVersion=2022-11-28#upload-a-release-asset
      // did not work and return HTTP error Not Found.
      octokit.rest.repos.uploadReleaseAsset({
        owner: parts[0],
        repo: parts[1],
        release_id: releaseInfo.data.id,
        name: `${parts[1]}-${ghRef}.${formats[i]}`,
        data: result.data,
      });
      i += 1;
    }

    // Now create the results.
    i = 0;
    const subj: attestation.subjectObject[] = [];
    for (const format of formats) {
      subj.push({
        name: `${parts[1]}-${ghRef}.${format}`,
        digest: {
          sha256: hashes[i],
        },
      });
      i += 1;
    }
    const att: attestation.result = {
      version: 1,
      attestations: [
        {
          name: `${parts[1]}-${ghRef}`,
          subjects: subj,
        },
      ],
    };
    core.debug(`att: ${Buffer.from(JSON.stringify(att))}`);
    const attFile = inputs.getEnvVariable("SLSA_OUTPUTS_ARTIFACTS_FILE");
    tscommon.safeWriteFileSync(attFile, Buffer.from(JSON.stringify(att)));
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}

run();
