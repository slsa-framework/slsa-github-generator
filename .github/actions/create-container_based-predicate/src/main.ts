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

import * as core from "@actions/core";
import type { BuildDefinition, ResourceDescriptor } from "./predicate";
import { generatePredicate } from "./predicate";
import * as gh from "./github";
import * as utils from "./utils";
import * as tscommon from "tscommon";

async function run(): Promise<void> {
  try {
    /* Test locally. Requires a GitHub token:
        $ env INPUT_BUILD-DEFINITION="testdata/build_definition.json" \
        INPUT_OUTPUT-FILE="predicate.json" \
        INPUT_BINARY-SHA256="0982432e54df5f3eb6b25c6c1ae77a45c242ad5a81a485c1fc225ae5ac472be3" \
        INPUT_BINARY-URI="git+https://github.com/asraa/slsa-github-generator@refs/heads/refs/heads/main" \
        INPUT_TOKEN="$(gh auth token)" \
        INPUT_BUILDER-ID="https://github.com/asraa/slsa-github-generator/.github/workflows/builder_container-based_slsa3.yml@refs/tags/v0.0.1" \
        GITHUB_EVENT_NAME="workflow_dispatch" \
        GITHUB_RUN_ATTEMPT="1" \
        GITHUB_RUN_ID="4128571590" \
        GITHUB_RUN_NUMBER="38" \
        GITHUB_WORKFLOW="pre-submit e2e container-based default" \
        GITHUB_WORKFLOW_REF="asraa/slsa-github-generator/.github/workflows/pre-submit.e2e.container-based.default.yml@refs/heads/main" \
        GITHUB_SHA="97f1bfd54b02d1c7b632da907676a7d30d2efc02" \
        GITHUB_REPOSITORY="asraa/slsa-github-generator" \
        GITHUB_REPOSITORY_ID="479129389" \
        GITHUB_REPOSITORY_OWNER="asraa" \
        GITHUB_REPOSITORY_OWNER_ID="5194569" \
        GITHUB_ACTOR_ID="5194569" \
        GITHUB_REF="refs/heads/main" \
        GITHUB_BASE_REF="" \
        GITHUB_REF_TYPE="branch" \
        GITHUB_ACTOR="asraa" \
        GITHUB_WORKSPACE="$(pwd)" \
        nodejs ./dist/index.js
    */
    const bdPath = core.getInput("build-definition");
    const outputFile = core.getInput("output-file");
    const binaryDigest = core.getInput("binary-sha256");
    const binaryURI = core.getInput("binary-uri");
    const jobWorkflowRef = core.getInput("builder-id");
    const token = core.getInput("token");
    if (!token) {
      throw new Error("token not provided");
    }

    if (!tscommon.safeExistsSync(bdPath)) {
      throw new Error("build-definition file does not exist");
    }

    // Read SLSA build definition
    const buffer = tscommon.safeReadFileSync(bdPath);
    const bd: BuildDefinition = JSON.parse(buffer.toString());

    // Get builder binary artifact reference.
    const builderBinaryRef: ResourceDescriptor = {
      uri: binaryURI,
      digest: {
        sha256: binaryDigest,
      },
    };

    // Generate the predicate.
    const ownerRepo = utils.getEnv("GITHUB_REPOSITORY");
    const currentWorkflowRun = await gh.getWorkflowRun(
      ownerRepo,
      Number(process.env.GITHUB_RUN_ID),
      token
    );
    const predicate = generatePredicate(
      bd,
      builderBinaryRef,
      jobWorkflowRef,
      currentWorkflowRun
    );

    // Write output predicate
    tscommon.safeWriteFileSync(outputFile, JSON.stringify(predicate));
    core.debug(`Wrote predicate to ${outputFile}`);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}
run();
