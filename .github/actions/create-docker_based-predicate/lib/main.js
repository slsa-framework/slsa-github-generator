"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const core = __importStar(require("@actions/core"));
const predicate_1 = require("./predicate");
const gh = __importStar(require("./github"));
const utils = __importStar(require("./utils"));
const tscommon = __importStar(require("tscommon"));
function run() {
    return __awaiter(this, void 0, void 0, function* () {
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
                GITHUB_WORKFLOW_REF="asraa/slsa-github-generator/.github/workflows/pre-submit.e2e.docker-based.default.yml@refs/heads/main" \
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
            const bd = JSON.parse(buffer.toString());
            // Get builder binary artifact reference.
            const builderBinaryRef = {
                uri: binaryURI,
                digest: {
                    sha256: binaryDigest,
                },
            };
            // Generate the predicate.
            const ownerRepo = utils.getEnv("GITHUB_REPOSITORY");
            const currentWorkflowRun = yield gh.getWorkflowRun(ownerRepo, Number(process.env.GITHUB_RUN_ID), token);
            const predicate = (0, predicate_1.generatePredicate)(bd, builderBinaryRef, jobWorkflowRef, currentWorkflowRun);
            // Write output predicate
            tscommon.safeWriteFileSync(outputFile, JSON.stringify(predicate));
            core.debug(`Wrote predicate to ${outputFile}`);
        }
        catch (error) {
            if (error instanceof Error) {
                core.setFailed(error.message);
            }
            else {
                core.setFailed(`Unexpected error: ${error}`);
            }
        }
    });
}
run();
