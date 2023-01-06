export const INTOTO_TYPE = "https://in-toto.io/Statement/v0.1";

export type Subject = {
  digest: { [key: string]: string };
  name: string;
};

export type Attestation = {
  name: string;
  subjects: Subject[];
};

export interface IntotoStatement {
  _type: string;
  subject: Subject[];
  predicateType: string;
  predicate: object;
}

export type Layout = {
  version: number;
  attestations: Attestation[];
};
