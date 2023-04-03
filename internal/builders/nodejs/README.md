# Generation of SLSA3+ provenance for Node.js packages

This document explains how to generate SLSA provenance for Node.js packages.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Node.js builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the Node.js builder](#referencing-the-nodejs-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Verification](#verification)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Node.js builder will generate a non-forgeable attestation to the
Node.js package using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the package they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Node.js builder is currently under active development. The API could change
while approaching an initial release. You can track progress towards General
Availability via [this
milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/8).

## Generating Provenance

The Node.js builder uses a Github Actions reusable workflow to build your
package and generate the provenance.

### Getting Started

Let's assume you have a `package.json` for your TypeScript project that looks
something like this. Here we have a `build` script that runs `tsc` to compile
our package's TypeScript code.

We also define a `ci` script that runs `npm ci`. We will see later that define
a `ci` script so we can install our `devDependencies` before running our
`build` script.

In order to publish to `npmjs.com`, packages should define a `repository`
field. This should match the repository used to run the Node.js builder.

```json
{
  "name": "mypackage",
  "version": "0.0.1",
  "description": "My Package",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": ["/dist"],
  "scripts": {
    "ci": "npm ci",
    "build": "tsc"
  },
  "devDependencies": {
    "typescript": "^4.8.4"
  },
  "repository": {
    "type": "git",
    "url": "git+https://github.com/ianlewis/mypackage.git"
  }
}
```

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.
This assumes that the `package.json` is in the root directory of your
repository.

```yaml
jobs:
  build:
    permissions:
      actions: read # for detecting the Github Actions environment.
      id-token: write # for creating OIDC tokens for signing.
      packages: write # for uploading attestations.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@v1.5.0
    with:
      run-scripts: "ci, build"
    secrets:
      node-auth-token: ${{ secrets.NPM_TOKEN }}
```

The `run-scripts` are a set of comma separated build scripts to run before
publishing the package. This scripts run in order.

After the build scripts are run, the Node.js builder runs `npm publish` to
publish your package to the npm registry. We provide a `node-auth-token` so
that we can authenticate with `npmjs.com`.

### Referencing the Node.js builder

At present, the builder **MUST** be referenced by a tag of the form `@vX.Y.Z`,
because the build will fail if you reference it via a shorter tag like `@vX.Y`
or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure
renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
https://rekor.sigstore.dev/. This entry includes the repository name. This will cause the
private repository name to leak and be discoverable via the public Rekor API
server.

If this is ok with you, you can set the `rekor-log-public` flag in order to
opt in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  rekor-log-public: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue #372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

### Supported Triggers

The following [GitHub trigger
events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows)
are fully supported and tested:

- `schedule`
- `push` (including new tags)
- `release`
- Manual run via `workflow_dispatch`

However, in practice, most triggers should work with the exception of
`pull_request`. If you would like support for `pull_request`, please tell us
about your use case on [issue
#358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a [new
issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The Node.js builder accepts the following inputs:

Inputs:

| Name              | Required | Default                                                          | Description                                                                                                                                                                                                                                         |
| ----------------- | -------- | ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| access            | No       | 'restricted' for scoped packages, 'public' for unscoped packages | The access level for the package. Valid values are "public" or "restricted".                                                                                                                                                                        |
| directory         | No       | `github.workspace`                                               | The root directory of the package (i.e. where the `package.json` is located)                                                                                                                                                                        |
| npm-publish       | No       | `true` if pushing a new tag, `false` otherwise                   | Publish to the npm package repository.                                                                                                                                                                                                              |
| node-version      | No       |                                                                  | The version of Node.js to use. If no value is supplied, the `node` version from `$PATH` is used.                                                                                                                                                    |
| node-version-file | No       |                                                                  | File containing the version Spec of the version to use. Examples: .nvmrc, .node-version, .tool-versions.                                                                                                                                            |
| rekor-log-public  | No       | false                                                            | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false` |
| run-scripts       | No       |                                                                  | A comma separated ordered list of npm scripts to run before running `npm publish`. See [scripts](https://docs.npmjs.com/cli/v9/using-npm/scripts) for more information. \                                                                           |
| dist-tag          | No       | latest                                                           | The package dist-tag to attach. See `npm help dist-tag` for more information on tags.                                                                                                                                                               |

Secrets:

| Name              | Description                                                                                                                                              |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `node-auth-token` | An npm access token used to authenticated with the registry. See [About access tokens](https://docs.npmjs.com/about-access-tokens) for more information. |

### Workflow Outputs

The workflow does not have any outputs.

### Provenance Format

Provenance is generated as an [in-toto](https://in-toto.io/) statement with a
SLSA v0.2 predicate.

| Name           | Value                                                          | Description                                                                                    |
| -------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `subject.name` | Package url ([purl](https://github.com/package-url/purl-spec)) | The subject identifies the package in [purl](https://github.com/package-url/purl-spec) format. |

The project generates SLSA v0.2 provenance predicate with the following values.

| Name                         | Value                                                                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `builder.id`                 | `https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@refs/tags/v1.5.0"` | Identifies the Node.js builder                                                                                                                                                                                         |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0"`                                       | Identifies a the GitHub Actions build.                                                                                                                                                                                 |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                                                                             | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

### Provenance Example

The following is an example of the generated provenance.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "subject": [
    {
      "name": "pkg:npm/%40ianlewis/actions-test@0.1.77",
      "digest": {
        "sha512": "3bbabe6803f7f9fd8a22dd3e85ae015664a75f277459d3bc2fc6aacd7054eb8af432188d92fca653af49460f9209562c0d298b959b169db98d369c4dcfe8406e"
      }
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@refs/tags/v1.5.0"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/tags/v0.1.77",
        "digest": {
          "sha1": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91"
        },
        "entryPoint": ".github/workflows/nodejs.yml"
      },
      "parameters": {
        "inputs": {
          "access": "",
          "directory": "nodejs",
          "node-version": "16",
          "node-version-file": "",
          "rekor-log-public": false,
          "run-scripts": "build",
          "dist-tag": "latest"
        }
      },
      "environment": {
        "GITHUB_ACTOR_ID": "49289",
        "GITHUB_EVENT_NAME": "push",
        "GITHUB_REF": "refs/tags/v0.1.77",
        "GITHUB_REF_TYPE": "tag",
        "GITHUB_REPOSITORY": "ianlewis/actions-test",
        "GITHUB_REPOSITORY_ID": "474793590",
        "GITHUB_REPOSITORY_OWNER_ID": "49289",
        "GITHUB_RUN_ATTEMPT": "1",
        "GITHUB_RUN_ID": "4527998016",
        "GITHUB_RUN_NUMBER": "64",
        "GITHUB_SHA": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91",
        "GITHUB_TRIGGERING_ACTOR_ID": "49289",
        "GITHUB_WORKFLOW_REF": "ianlewis/actions-test/.github/workflows/nodejs.yml@refs/tags/v0.1.77",
        "GITHUB_WORKFLOW_SHA": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91",
        "IMAGE_OS": "ubuntu22",
        "IMAGE_VERSION": "20230317.1",
        "RUNNER_ARCH": "X64",
        "RUNNER_NAME": "GitHub Actions 12",
        "RUNNER_OS": "Linux"
      }
    },
    "metadata": {
      "buildInvocationId": "4527998016-1",
      "completeness": {
        "parameters": true
      }
    },
    "materials": [
      {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/tags/v0.1.77",
        "digest": {
          "sha1": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91"
        }
      }
    ]
  }
}
```

## Verification

Verification of Node.js packages is currently a work in progress. Please see
[`slsa-verifier`](https://github.com/slsa-framework/slsa-verifier) issue
[#449](https://github.com/slsa-framework/slsa-verifier/issues/449).
