"use strict";
// TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1470):
// Share this code with BYO predicate definitions.
Object.defineProperty(exports, "__esModule", { value: true });
exports.generatePredicate = void 0;
const github_1 = require("./github");
function generatePredicate(bd, binaryRef, jobWorkflowRef, currentRun) {
    let pred = {
        buildDefinition: bd,
        runDetails: {
            builder: {
                id: jobWorkflowRef,
            },
            metadata: {
                invocationId: (0, github_1.getInvocationID)(currentRun),
            },
        },
    };
    // Add the builder binary to the resolved dependencies.
    pred.buildDefinition.resolvedDependencies = [binaryRef];
    // Update the parameters with the GH context, including workflow
    // inputs.
    pred = (0, github_1.addGitHubParameters)(pred, currentRun);
    return pred;
}
exports.generatePredicate = generatePredicate;
