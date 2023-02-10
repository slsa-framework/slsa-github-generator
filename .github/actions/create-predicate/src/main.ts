import * as core from "@actions/core";
import type { BuildDefinition, ArtifactReference } from "./predicate";
import { generatePredicate } from "./predicate";
import * as gh from "./github";
import * as utils from "./utils";
import * as fs from "fs";

async function run(): Promise<void> {
  try {
    /* Test locally. Requires a GitHub token:
        $ env INPUT_BUILD-DEFINITION="testdata/build_definition.json" \
        INPUT_OUTPUT-FILE="predicate.json" \
        INPUT_BINARY-SHA256="0982432e54df5f3eb6b25c6c1ae77a45c242ad5a81a485c1fc225ae5ac472be3" \
        INPUT_BINARY-URI="git+https://github.com/asraa/slsa-github-generator@refs/heads/refs/heads/main" \
        INPUT_TOKEN="$(gh auth token)" \
        INPUT_BUILDER-ID="https://github.com/asraa/slsa-github-generator/.github/workflows/builder_docker-baed_slsa3.yml@refs/tags/v0.0.1" \
        GITHUB_EVENT_NAME="workflow_dispatch" \
        GITHUB_RUN_ATTEMPT="1" \
        GITHUB_RUN_ID="4128571590" \
        GITHUB_RUN_NUMBER="38" \
        GITHUB_WORKFLOW="pre-submit e2e docker-based default" \
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
    const wd = utils.getEnv("GITHUB_WORKSPACE");

    const bdPath = core.getInput("build-definition");
    const outputFile = core.getInput("output-file");
    const binaryDigest = core.getInput("binary-sha256");
    const binaryURI = core.getInput("binary-uri");
    const jobWorkflowRef = core.getInput("builder-id");
    const token = core.getInput("token");
    if (!token) {
      throw new Error("token not provided");
    }

    const safeBdPath = utils.resolvePathInput(bdPath, wd);
    // TODO(#1513): Use a common utility to harden file writes.
    if (!fs.existsSync(safeBdPath)) {
      throw new Error("build-definition file does not exist");
    }

    // Read SLSA build definition
    const buffer = fs.readFileSync(safeBdPath);
    const bd: BuildDefinition = JSON.parse(buffer.toString());

    // Get builder binary artifact reference.
    const builderBinaryRef: ArtifactReference = {
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
    const safeOutput = utils.resolvePathInput(outputFile, wd);
    fs.writeFileSync(safeOutput, JSON.stringify(predicate), {
      flag: "ax",
      mode: 0o600,
    });
    core.debug(`Wrote predicate to ${safeOutput}`);
  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed(`Unexpected error: ${error}`);
    }
  }
}
run();
