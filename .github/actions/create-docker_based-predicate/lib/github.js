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
exports.getInvocationID = exports.getWorkflowInputs = exports.addGitHubParameters = exports.getWorkflowRun = void 0;
const process = __importStar(require("process"));
const github = __importStar(require("@actions/github"));
const tscommon = __importStar(require("tscommon"));
// getWorkflowRun retrieves the current WorkflowRun given the repository (owner/repo)
// and run ID.
function getWorkflowRun(repository, run_id, token) {
    return __awaiter(this, void 0, void 0, function* () {
        const octokit = github.getOctokit(token);
        const [owner, repo] = repository.split("/");
        const res = yield octokit.rest.actions.getWorkflowRun({
            owner,
            repo,
            run_id: Number(process.env.GITHUB_RUN_ID),
        });
        return res.data;
    });
}
exports.getWorkflowRun = getWorkflowRun;
// addGitHubParameters adds trusted GitHub context to system paramters
// and external parameters.
function addGitHubParameters(predicate, currentRun) {
    var _a;
    const { env } = process;
    const ctx = github.context;
    if (!predicate.buildDefinition.systemParameters) {
        predicate.buildDefinition.systemParameters = {};
    }
    const systemParams = predicate.buildDefinition.systemParameters;
    // Put GitHub context and env vars into systemParameters.
    systemParams.GITHUB_EVENT_NAME = ctx.eventName;
    systemParams.GITHUB_JOB = ctx.job;
    systemParams.GITHUB_REF = ctx.ref;
    systemParams.GITHUB_BASE_REF = env.GITHUB_BASE_REF || "";
    systemParams.GITHUB_REF_TYPE = env.GITHUB_REF_TYPE || "";
    systemParams.GITHUB_REPOSITORY = env.GITHUB_REPOSITORY || "";
    systemParams.GITHUB_RUN_ATTEMPT = env.GITHUB_RUN_ATTEMPT || "";
    systemParams.GITHUB_RUN_ID = ctx.runId;
    systemParams.GITHUB_RUN_NUMBER = ctx.runNumber;
    systemParams.GITHUB_SHA = ctx.sha;
    systemParams.GITHUB_WORKFLOW = ctx.workflow;
    systemParams.GITHUB_WORKFLOW_REF = env.GITHUB_WORKFLOW_REF || "";
    systemParams.GITHUB_WORKFLOW_SHA = env.GITHUB_WORKFLOW_SHA || "";
    systemParams.IMAGE_OS = env.ImageOS || "";
    systemParams.IMAGE_VERSION = env.ImageVersion || "";
    systemParams.RUNNER_ARCH = env.RUNNER_ARCH || "";
    systemParams.RUNNER_NAME = env.RUNNER_NAME || "";
    systemParams.RUNNER_OS = env.RUNNER_OS || "";
    systemParams.GITHUB_ACTOR_ID = String(((_a = currentRun.actor) === null || _a === void 0 ? void 0 : _a.id) || "");
    systemParams.GITHUB_REPOSITORY_ID = String(currentRun.repository.id || "");
    systemParams.GITHUB_REPOSITORY_OWNER_ID = String(currentRun.repository.owner.id || "");
    // Put GitHub event payload into systemParameters.
    // TODO(github.com/slsa-framework/slsa-github-generator/issues/1575): Redact sensitive information.
    if (env.GITHUB_EVENT_PATH) {
        const ghEvent = JSON.parse(tscommon.safeReadFileSync(env.GITHUB_EVENT_PATH || "").toString());
        systemParams.GITHUB_EVENT_PAYLOAD = ghEvent;
    }
    predicate.buildDefinition.systemParameters = systemParams;
    if (!env.GITHUB_WORKFLOW_REF) {
        throw new Error("missing GITHUB_WORKFLOW_REF");
    }
    return predicate;
}
exports.addGitHubParameters = addGitHubParameters;
// getWorkflowInputs gets the workflow runs' inputs (only populated on workflow dispatch).
function getWorkflowInputs() {
    const { env } = process;
    if (env.GITHUB_EVENT_NAME === "workflow_dispatch") {
        return github.context.payload.inputs;
    }
    return null;
}
exports.getWorkflowInputs = getWorkflowInputs;
// getInvocationID returns the URI describing the globally unique invocation ID.
function getInvocationID(currentRun) {
    return `https://github.com/${currentRun.repository.full_name}/actions/runs/${currentRun.id}/attempts/${currentRun.run_attempt}`;
}
exports.getInvocationID = getInvocationID;
