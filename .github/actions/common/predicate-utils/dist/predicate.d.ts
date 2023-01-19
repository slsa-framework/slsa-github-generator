import type { ApiWorkflowRun } from "./types";
export interface Builder {
  id: string;
  version?: string;
  builderDependencies?: ArtifactReference[];
}
export interface DigestSet {
  [key: string]: string;
}
export interface Metadata {
  invocationId?: string;
  startedOn?: Date;
  finishedOn?: Date;
}
export interface ParameterValue_Artifact {
  artifact: ArtifactReference;
}
export interface ParameterValue_String {
  value: string;
}
export type ParameterValue = ParameterValue_Artifact | ParameterValue_String;
export interface ArtifactReference {
  uri: string;
  digest: DigestSet;
  localName?: string;
  downloadLocation?: string;
  mediaType?: string;
}
export interface BuildDefinition {
  buildType: string;
  externalParameters: {
    [key: string]: ParameterValue;
  };
  systemParameters?: {
    [key: string]: ParameterValue;
  };
  resolvedDependencies?: ArtifactReference[];
}
export interface RunDetails {
  builder: Builder;
  metadata: Metadata;
  byproducts?: ArtifactReference[];
}
export interface SLSAv1Predicate {
  buildDefinition: BuildDefinition;
  runDetails: RunDetails;
}
export declare function addGitHubContext(
  predicate: SLSAv1Predicate,
  currentRun: ApiWorkflowRun
): SLSAv1Predicate;
export declare function createURI(repository: string, ref: string): string;
//# sourceMappingURL=predicate.d.ts.map
