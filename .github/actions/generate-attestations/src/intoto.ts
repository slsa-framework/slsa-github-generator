export const INTOTO_TYPE = "https://in-toto.io/Statement/v0.1";

export type Subject = {
  digest: { [key: string]: string };
  name: string;
};

export interface IntotoStatement {
  _type: string;
  subject: Subject[];
  predicateType: string;
  predicate: object;
}
