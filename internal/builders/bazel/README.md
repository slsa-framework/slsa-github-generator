# Generation of SLSA3+ provenance for Bazel builds

This document explains how to generate SLSA provenance for Bazel packages.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Bazel builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the Bazel builder](#referencing-the-bazel-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Verification](#verification)
  - [npm audit signatures](#npm-audit-signatures)
  - [slsa-verifier](#slsa-verifier)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Bazel builder will generate a non-forgeable attestation to the
Bazel build artifacts produced using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the artifact they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Bazel builder is currently in beta. The API could change while approaching
a Generally Available (GA) release.

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Generating Provenance

The Bazel builder uses a Github Actions reusable workflow to build your
artifact(s) and generate the provenance.

### Getting Started

Let's assume you have a Bazel project includes a
[WORKSPACE file](https://bazel.build/reference/glossary#workspace-file)
at the root of your repository as well as one or more
[BUILD files](https://bazel.build/reference/glossary#build-file). Also included
at the root of the directory is the
[.bazelversion file](https://github.com/bazelbuild/bazelisk#how-does-bazelisk-know-which-bazel-version-to-run)
, [.bazelrc file](https://bazel.build/run/bazelrc)
, and a
[MODULE.bazel file](https://bazel.build/rules/lib/globals/module).

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.
This assumes that the files above are in the root directory of your
repository.

We also define the targets to be built as 'target1' and 'target2', and
we define the flag of the build to be '--strip=always'

The following reusable workflow call will build the targets and upload the artifacts into a .zip and
generate provenance attestations which will be uploaded as another .zip to the
workflow run.

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@v1.6.0
    with:
      targets: "//src:target1 //src:target2"
      flags: "--strip=always"
```

The `targets` are a set of space separated build targets to be built.

Once the targets are built, the Bazel builder creates a folder for the artifacts
and another for the provenance attestations which are uploaded as artifacts to the workflow run.

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
about your use case on
[issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a
[new issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The Bazel builder accepts the following inputs:

Inputs:

| Name              | Required | Default            | Description                                                                                                                                                                                                                                         |
| ----------------- | -------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| rekor-log-public  | No       | false              | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false` |
| targets           | Yes      |                    | A space separated list of targets to build and generate artifacts for. See [targets](https://bazel.build/concepts/build-ref#targets) for more information. \                                                                                        |
| flags             | No       | ""                 | A space separated list of flags to modify the build by. See [flags](https://bazel.build/docs/user-manual#build-options) for more information. \                                                                                                     |

### Workflow Outputs

The Bazel builder produces the following outputs:

| Name                       | Description                                                            |
| -------------------------- | ---------------------------------------------------------------------- |
| provenance-download-name   | The name of the provenance attestation uploaded to the workflow run.   |

### Provenance Format

**Look into to double check**


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

DOUBLE CHECK IF SAME

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

TODO
