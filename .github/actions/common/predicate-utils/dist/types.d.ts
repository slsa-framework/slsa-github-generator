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
export interface ArtifactReference {
    uri: string;
    digest: DigestSet;
    localName?: string;
    downloadLocation?: string;
    mediaType?: string;
}
export interface BuildDefinition {
    buildType: string;
    externalParameters: any;
    systemParameters?: any;
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
