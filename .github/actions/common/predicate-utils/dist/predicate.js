"use strict";
/*
Copyright 2023 SLSA Authors
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
Object.defineProperty(exports, "__esModule", { value: true });
exports.createURI = exports.addGitHubContext = void 0;
const process = __importStar(require("process"));
const fs = __importStar(require("fs"));
// addGitHubContext
function addGitHubContext(predicate, currentRun) {
    var _a;
    const { env } = process;
    if (!predicate.buildDefinition.systemParameters) {
        predicate.buildDefinition.systemParameters = {};
    }
    // TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1505):
    // Add GitHub event payload.
    const systemParams = predicate.buildDefinition.systemParameters;
    systemParams["GITHUB_EVENT_NAME"] = { value: env.GITHUB_EVENT_NAME || "" };
    systemParams["GITHUB_JOB"] = { value: env.GITHUB_JOB || "" };
    systemParams["GITHUB_REF"] = { value: env.GITHUB_REF || "" };
    systemParams["GITHUB_REF_TYPE"] = { value: env.GITHUB_REF_TYPE || "" };
    systemParams["GITHUB_REPOSITORY"] = { value: env.GITHUB_REPOSITORY || "" };
    systemParams["GITHUB_RUN_ATTEMPT"] = { value: env.GITHUB_RUN_ATTEMPT || "" };
    systemParams["GITHUB_RUN_ID"] = { value: env.GITHUB_RUN_ID || "" };
    systemParams["GITHUB_RUN_NUMBER"] = { value: env.GITHUB_RUN_NUMBER || "" };
    systemParams["GITHUB_SHA"] = { value: env.GITHUB_SHA || "" };
    systemParams["GITHUB_WORKFLOW"] = { value: env.GITHUB_WORKFLOW || "" };
    systemParams["GITHUB_ACTOR_ID"] = {
        value: String(((_a = currentRun.actor) === null || _a === void 0 ? void 0 : _a.id) || ""),
    };
    systemParams["GITHUB_REPOSITORY_ID"] = {
        value: String(currentRun.repository.id || ""),
    };
    systemParams["GITHUB_REPSITORY_OWNER_ID"] = {
        value: String(currentRun.repository.owner.id || ""),
    };
    systemParams["GITHUB_WORKFLOW_REF"] = {
        value: env.GITHUB_WORKFLOW_REF || "",
    };
    systemParams["GITHUB_WORKFLOW_SHA"] = {
        value: env.GITHUB_WORKFLOW_SHA || "",
    };
    systemParams["IMAGE_OS"] = { value: env.ImageOS || "" };
    systemParams["IMAGE_VERSION"] = { value: env.ImageVersion || "" };
    systemParams["RUNNER_ARCH"] = { value: env.RUNNER_ARCH || "" };
    systemParams["RUNNER_NAME"] = { value: env.RUNNER_NAME || "" };
    systemParams["RUNNER_OS"] = { value: env.RUNNER_OS || "" };
    if (env.GITHUB_EVENT_NAME === "workflow_dispatch") {
        if (env.GITHUB_EVENT_PATH) {
            const ghEvent = JSON.parse(fs.readFileSync(env.GITHUB_EVENT_PATH).toString());
            for (const input in ghEvent.inputs) {
                // The invocation parameters belong here and are the top-level GitHub
                // workflow inputs.
                predicate.buildDefinition.externalParameters[`input_${input}`] = {
                    value: String(ghEvent.inputs[input] || ""),
                };
            }
        }
    }
    return predicate;
}
exports.addGitHubContext = addGitHubContext;
// createURI creates the fully qualified URI out of the repository
function createURI(repository, ref) {
    if (!repository) {
        throw new Error(`cannot create URI: repository undefined`);
    }
    let refVal = "";
    if (ref) {
        refVal = `@${ref}`;
    }
    return `git+https://github.com/${repository}${refVal}`;
}
exports.createURI = createURI;
