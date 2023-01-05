import * as fs from "fs";
import * as child_process from "child_process";

import * as core from "@actions/core";
import * as sigstore from "sigstore";

import { githubContext } from "./validate";

export interface RawToken {
  version: number;
  context: string;
  builder: {
    private_repository: boolean;
    runner_label: string;
    audience: string;
  };
  github: githubContext;
  tool: {
    actions: {
      build_artifacts: {
        path: string;
      };
    };
    // NOTE: reusable workflows only support inputs of type
    // boolean, number, or string.
    // https://docs.github.com/en/actions/using-workflows/reusing-workflows#passing-inputs-and-secrets-to-a-reusable-workflow.
    inputs: Map<string, Object>;
  };
}

/**
 * extractIdentityFromSAN extracts the repository name and ref from the given SAN URI.
 */
export function extractIdentityFromSAN(URI: string): [string, string] {
  // NOTE: the URI looks like:
  // https://github.com/laurentsimon/slsa-delegated-tool/.github/workflows/tool1_slsa3.yml@refs/heads/main.
  // We want to extract:
  // - the repository: laurentsimon/slsa-delegated-tool
  // - the ref: refs/heads/main
  const parts = URI.split("@");
  if (parts.length !== 2) {
    throw new Error(`invalid URI (1): ${URI}`);
  }
  const ref = parts[1];
  const url = parts[0];
  const gitHubURL = "https://github.com/";
  if (!url.startsWith(gitHubURL)) {
    throw new Error(`not a GitHub URI: ${URI}`);
  }
  // NOTE: we omit the gitHubURL from the URL.
  const parts2 = url.slice(gitHubURL.length).split("/");
  if (parts2.length <= 2) {
    throw new Error(`invalid URI (2): ${URI}`);
  }
  const repo = `${parts2[0]}/${parts2[1]}`;
  return [repo, ref];
}

/**
 * parseCertificateIdentity parses the sigstore.Bundle's certificate identity
 * and returns the Tool Reusable Workflow's URI, repository, and ref as a three
 * tuple.
 */
export function parseCertificateIdentity(
  bundle: sigstore.sigstore.Bundle
): [string, string, string] {
  if (bundle === undefined) {
    throw new Error(`undefined bundle.`);
  }
  if (bundle.verificationMaterial === undefined) {
    throw new Error(`undefined bundle.verificationMaterial.`);
  }
  if (bundle.verificationMaterial.x509CertificateChain === undefined) {
    throw new Error(
      `undefined bundle.verificationMaterial.x509CertificateChain.`
    );
  }
  if (
    bundle.verificationMaterial.x509CertificateChain.certificates.length === 0
  ) {
    throw new Error(
      `bundle.verificationMaterial.x509CertificateChaincertificates is empty.`
    );
  }
  // NOTE: the first certificate is the client certificate.
  const clientCertDer = Buffer.from(
    bundle.verificationMaterial.x509CertificateChain.certificates[0].rawBytes,
    "base64"
  );
  const clientCertPath = "client.cert";
  fs.writeFileSync(clientCertPath, clientCertDer);

  // https://stackabuse.com/executing-shell-commands-with-node-js/
  // The SAN from the certificate looks like:
  // `
  //  X509v3 Subject Alternative Name: critical\n
  //      URI:https://github.com/laurentsimon/slsa-delegated-tool/.github/workflows/tool1_slsa3.yml@refs/heads/main\n
  // `
  const result = child_process
    .execSync(`openssl x509 -in ${clientCertPath} -noout -ext subjectAltName`)
    .toString();
  const index = result.indexOf("URI:");
  if (index === -1) {
    throw new Error("error: cannot find URI in subjectAltName");
  }
  const toolURI = result.slice(index + 4).replace("\n", "");
  core.debug(`tool-uri: ${toolURI}`);

  // NOTE: we can use the job_workflow_ref and job_workflow_sha when they become available.
  const [toolRepository, toolRef] = extractIdentityFromSAN(toolURI);
  core.debug(`tool-repository: ${toolRepository}`);
  core.debug(`tool-ref: ${toolRef}`);

  return [toolURI, toolRepository, toolRef];
}
