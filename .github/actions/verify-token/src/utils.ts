// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { rawTokenInterface } from "./types";
import * as core from "@actions/core";
import { sigstore } from "sigstore";
import * as child_process from "child_process";
import * as tscommon from "tscommon";
import * as github from "@actions/github";

// createTriggerURI creates the fully qualified URI out of the trigger repository.
export function createTriggerURI(rawTokenObj: rawTokenInterface): string {
  if (!rawTokenObj.github.repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  const repository = rawTokenObj.github.repository;
  const ref = rawTokenObj.github.ref;
  let refVal = "";
  if (ref) {
    refVal = `@${ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}

// createSourceURI creates the fully qualified URI out of the checked out repository.
export function createSourceURI(rawTokenObj: rawTokenInterface): string {
  if (!rawTokenObj.github.repository) {
    throw new Error(`cannot create URI: repository undefined`);
  }
  const repository = rawTokenObj.github.repository;

  // TRW may overwrite the commit sha to build.
  // For example, users of JReleaser may push a commit
  // before building. See discussion at
  // https://github.com/slsa-framework/slsa-github-generator/issues/2043.
  // If the TRW passed in a sha1, we don't know the ref
  // so we never report it.
  if (rawTokenObj.source.checkout.sha1) {
    return `git+https://github.com/${repository}`;
  }

  let refVal = "";
  if (rawTokenObj.github.ref) {
    refVal = `@${rawTokenObj.github.ref}`;
  }
  return `git+https://github.com/${repository}${refVal}`;
}

// getTriggerSha1 returns the sha1 of the trigger repository.
export function getTriggerSha1(rawTokenObj: rawTokenInterface): string {
  const sha1 = rawTokenObj.github.sha;
  validateSha1(sha1);
  return sha1;
}

// getTriggerRepository returns the repository of the trigger.
export function getTriggerRepository(rawTokenObj: rawTokenInterface): string {
  return rawTokenObj.github.repository;
}

// getTriggerRef returns the ref of the trigger.
export function getTriggerRef(rawTokenObj: rawTokenInterface): string {
  return rawTokenObj.github.ref;
}

function validateSha1(sha1: string): void {
  // 40 characters in hex format.
  if (!/[a-fA-F0-9]{40}/.test(sha1)) {
    throw new Error(`invalid sha1: ${sha1}`);
  }
}

// getSourceSha1 returns the sha1 of the source that is checked out.
export function getSourceSha1(rawTokenObj: rawTokenInterface): string {
  // The checkout.sha1 takes precedence over the default GitHub event.
  // TRW may overwrite the commit sha to build.
  // For example, users of JReleaser may push a commit
  // before building. See discussion at
  // https://github.com/slsa-framework/slsa-github-generator/issues/2043.
  const sha1 = rawTokenObj.source.checkout.sha1 || rawTokenObj.github.sha;
  validateSha1(sha1);
  return sha1;
}

// getTriggerPath returns the workflow's path from the workflow_ref.
export function getTriggerPath(rawTokenObj: rawTokenInterface): string {
  // GITHUB_WORKFLOW_REF contains the repository name in the path. We will trim
  // it out.
  // e.g. 'octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch'
  // Strip off the repo name and git ref from the workflow path.
  return rawTokenObj.github.workflow_ref
    .substring(`${rawTokenObj.github.repository}/`.length)
    .split("@", 1)[0];
}

// This function takes an Object and
// creates a Map. Both JSON.parse() and YAML.parse()
// return Object for fields that are declared as Map, so we need
// to create a Map for it.
export function asMap<T>(inputs: Map<string, T>): Map<string, T> {
  if (!inputs) {
    return new Map<string, T>();
  }
  if (!(inputs instanceof Map)) {
    return new Map<string, T>(Object.entries(inputs));
  }
  return inputs;
}

export function parseCertificate(
  bundle: sigstore.Bundle
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

export async function fetchToolWorkflow(
  ghToken: string,
  repoName: string,
  hash: string,
  workflowPath: string
): Promise<string> {
  const octokit = github.getOctokit(ghToken);
  const [o, r] = repoName.split("/", 2);
  const response = await octokit.rest.repos.getContent({
    owner: o,
    repo: r,
    path: workflowPath,
    ref: hash,
  });

  if (!("content" in response.data)) {
    throw new Error("no data");
  }

  // Content is base64 encoded.
  const content = Buffer.from(response.data["content"], "base64").toString();
  core.info(`content: ${content}`);
  return content;
}
