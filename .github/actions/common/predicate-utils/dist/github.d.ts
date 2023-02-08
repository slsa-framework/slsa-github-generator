import * as types from "./types";
import type { ApiWorkflowRun } from "./types";
export declare function getWorkflowRun(repository: string, run_id: number, token: string): Promise<ApiWorkflowRun>;
export declare function addGitHubSystemParameters(predicate: types.SLSAv1Predicate, currentRun: ApiWorkflowRun): types.SLSAv1Predicate;
export type InputMap = {
    [k: string]: unknown;
} | null;
export declare function getWorkflowInputs(): InputMap;
export declare function getInvocationID(currentRun: ApiWorkflowRun): string;
export declare function getSourceArtifactReference(): types.ArtifactReference;
