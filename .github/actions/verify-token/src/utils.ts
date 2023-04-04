import { githubObj } from "./types";
import * as core from "@actions/core";
import * as sigstore from "sigstore";
import * as child_process from "child_process";
import * as tscommon from "tscommon";

// createURI creates the fully qualified URI out of the repository
export function createURI(repository: string, ref: string): string {
  if (!repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  let refVal = "";
  if (ref) {
    refVal = `@${ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}

// getWorkflowPath returns the workflow's path from the workflow_ref.
export function getWorkflowPath(obj: githubObj): string {
  // GITHUB_WORKFLOW_REF contains the repository name in the path. We will trim
  // it out.
  // e.g. 'octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch'
  // Strip off the repo name and git ref from the workflow path.
  return obj.workflow_ref
    .substring(`${obj.repository}/`.length)
    .split("@", 1)[0];
}

export function parseCertificate(
  bundle: sigstore.sigstore.Bundle
): [string, string, string, string, string] {
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

  tscommon.safeWriteFileSync(clientCertPath, clientCertDer);

  // https://stackabuse.com/executing-shell-commands-with-node-js/
  // NOTE: it's also included in the '1.3.6.1.4.1.57264.1.9' extension field.
  // https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md#1361415726419--build-signer-uri
  // The SAN from the certificate looks like:
  // `
  //  X509v3 Subject Alternative Name: critical\n
  //      URI:https://github.com/laurentsimon/slsa-delegated-tool/.github/workflows/tool1_slsa3.yml@refs/heads/main\n
  // `
  const resultSAN = child_process
    .execSync(`openssl x509 -in ${clientCertPath} -noout -ext subjectAltName`)
    .toString();
  const indexSAN = resultSAN.indexOf("URI:");
  if (indexSAN === -1) {
    throw new Error("error: cannot find URI in subjectAltName");
  }
  const toolURI = resultSAN.slice(indexSAN + 4).replace("\n", "");
  core.debug(`tool-uri: ${toolURI}`);

  // NOTE: we can use the job_workflow_ref and job_workflow_sha when they become available.
  const [toolRepository, toolRef] = extractIdentifyFromSAN(toolURI);
  core.debug(`tool-repository: ${toolRepository}`);
  core.debug(`tool-ref: ${toolRef}`);

  // https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
  // We use the more recent Fulcio claims to extract the tool information.
  // The commit sha for the tool is stored in
  // a v3 extension with oid '1.3.6.1.4.1.57264.1.9'.
  // https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md#1361415726419--build-signer-uri
  // `
  //    1.3.6.1.4.1.57264.1.3:\n
  //        8cbf4d422367d8499d5980a837cb9cc8e1e67001
  // `
  const textCertPath = "./client.txt";
  const shaOid = "1.3.6.1.4.1.57264.1.10";
  child_process
    .execSync(
      `openssl x509 -in ${clientCertPath} -noout -text -out ${textCertPath}`
    )
    .toString();
  const resultSha = child_process
    .execSync(`grep -A 1 '${shaOid}:' ${textCertPath}`)
    .toString();
  const indexSha = resultSha.indexOf(shaOid);
  if (indexSha === -1) {
    throw new Error(`error: cannot find oid '${shaOid}' in certificate`);
  }

  core.debug(`resultSha: ${resultSha}`);
  const sha1ByteLen = 20;
  const toolSha = resultSha
    .slice(indexSha + `${shaOid}:`.length)
    .replace("\n", "")
    .trim()
    .slice(-(sha1ByteLen * 2));
  core.debug(`tool-sha: ${toolSha}`);

  const toolPath = removeSuffix(
    removePrefix(toolURI, `https://github.com/${toolRepository}/`),
    `@${toolRef}`
  );
  core.debug(`tool-path: ${toolPath}`);
  return [toolURI, toolRepository, toolRef, toolSha, toolPath];
}

function removeSuffix(s: string, suffix: string): string {
  if (!s.endsWith(suffix)) {
    throw new Error(`error: no suffix '${suffix}' in '${s}'`);
  }
  return s.slice(0, -suffix.length);
}

function removePrefix(s: string, prefix: string): string {
  if (!s.startsWith(prefix)) {
    throw new Error(`error: no prefix '${prefix}' in '${s}'`);
  }
  return s.slice(prefix.length);
}

function extractIdentifyFromSAN(URI: string): [string, string] {
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
