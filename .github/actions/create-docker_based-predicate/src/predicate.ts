// TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1470):
// Share this code with BYO predicate definitions.

import type { Endpoints } from "@octokit/types";
import { addGitHubParameters, getInvocationID } from "./github";

import {
  BuildDefinition,
  ArtifactReference,
  SLSAv1Predicate,
} from "predicate-utils";

export type ApiWorkflowRun =
  Endpoints["GET /repos/{owner}/{repo}/actions/runs/{run_id}"]["response"]["data"];

export function generatePredicate(
  bd: BuildDefinition,
  binaryRef: ArtifactReference,
  jobWorkflowRef: string,
  currentRun: ApiWorkflowRun
): SLSAv1Predicate {
  let pred: SLSAv1Predicate = {
    buildDefinition: bd,
    runDetails: {
      builder: {
        id: jobWorkflowRef,
      },
      metadata: {
        invocationId: getInvocationID(currentRun),
      },
    },
  };
  // Add the builder binary to the resolved dependencies.
  pred.buildDefinition.resolvedDependencies = [binaryRef];

  // Update the parameters with the GH context, including workflow
  // inputs.
  pred = addGitHubParameters(pred, currentRun);

  return pred;
}
