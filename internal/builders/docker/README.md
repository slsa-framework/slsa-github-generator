# Generation of SLSA3+ provenance for artifacts created in a Docker container

This document explains how to generate SLSA provenance for artifacts created by
running a command inside a user-supplied container.

The container-based workflow uses a GitHub Actions [reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the artifacts and generate the SLSA provenance. We'll call this
workflow the "container-based workflow" from now on.

This workflow differs from the [generic container
workflow](../container/README.md) which attests to a container image, not
artifacts that are built from a process within a Docker image. This also differs
from the [generic artifact workflow](../generic/README.md) as it performs the
build as well as generates the provenance. This workflow uses a distinct build
type that provides the full details on the build process.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the SLSA builder](#referencing-the-slsa-builder)
  - [Private Repositories](#private-repositories)
  - [Registry Authentication](#registry-authentication)
  - [Supported Triggers](#supported-triggers)
  - [Configuration File](#configuration-file)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Example](#workflow-example)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Command line tool](#command-line-tool)
  - [The `dry-run` subcommand](#the-dry-run-subcommand)
  - [The `build` subcommand](#the-build-subcommand)
  - [The `verify` command](#the-verify-command)

<!-- tocstop -->

---

## Benefits of Provenance

Using this workflow to build artifacts will generate a non-forgeable attestation
to specified artifacts using the identity of the GitHub workflow. This can be
used to create a positive attestation to artifacts built inside a container
image coming from your repository.

In addition, the provenance contains detailed information about the build
process: the base image, the command that was run to generate the artifact, the
produced artifacts, and the environment where the execution was preformed.

That means that once your users verify the artifacts they have downloaded they
can be sure that the artifacts was created by your repository's workflow and
hasn't been tampered with. The details in the provenance allow users to create
policy on or verify properties of the build and even provide enough information
to reproduce the artifact.

## Generating Provenance

The container-based workflow uses a Github Actions reusable workflow to generate
the provenance.

### Getting Started

### Referencing the SLSA builder

At present, the generator **MUST** be referenced by a tag of the form `@vX.Y.Z`,
because the build will fail if you reference it via a shorter tag like `@vX.Y`
or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure
renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
<https://rekor.sigstore.dev/>. This entry includes the repository name. This
will cause the private repository name to leak and be discoverable via the
public Rekor API server.

If this is ok with you, you can set the `rekor-log-public` flag in order to opt
in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  rekor-log-public: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue
#372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

### Registry Authentication

This workflow support authentication against a Docker registry for private base images
through an input `registry-username` and `registry-password` secret. Authentication is
not required for public base images.
See [Workflow Inputs](#workflow-inputs) for more.

Authentication to GCR using GCP workflow identity federation is also supported using
`gcp-workload-identity-provider` and `gcp-service-account`.

### Supported Triggers

The following [GitHub trigger
events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows)
are fully supported and tested:

- `schedule`
- Manual run via `workflow_dispatch`

<!-- 

TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1656):
Support artifact and provenance upload to a release.
- `push` (including new tags)
- `release`

-->

In practice, most triggers should work. For events that do not have access to
the `id-token: write` permission, like `pull_request`, the workflow will run the
container-based build and produce an *unsigned* DSSE attestation for the
purposes of testing.

If you have an issue with any other triggers please submit a [new
issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Configuration File

The user must supply a configuration file location in their source repository
that contains the details of the build.

```toml
# (Required) Docker run command.
command = ["cp", "internal/builders/docker/testdata/config.toml", "config.toml"]

# Path to the file generated by the command above.
artifact_path = "**.toml"
```

The output artifact path supports wildcard characters. All matching files will
be measured and recorded as attestation subjects. The subject names will be the
basenames of the matching files.

### Workflow Inputs

The [container-based
workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/builder_docker-based_slsa3.yml)
accepts the following inputs:

Inputs:

| Name                             | Description                                                                                                                                                                                                                                                                             |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `builder-image`                          | **(Required)** The OCI image name of the builder image in which the build execution will be run. This must not include a tag or digest.                                                                                                                                                                                                               |
| `builder-digest`                         | **(Required)** The OCI image digest of the builder-image. The image digest of the form '<algorithm>:<digest>' (e.g. 'sha256:abcdef...')                                                                                                                                                                      |
| `config-path`              | **(Required)** Path to a configuration file relative to the root of the repository containing the command that the builder image should be invoked with and the path to the output artifacts. See [Configuration File](#configuration-file).                                                                                                                                                                  |
| `compile-builder`              | Whether to build the builder from source. This increases build time by ~2m.<br>Default: `false`.                                                                                                                                                                                      |
| `provenance-name`             | The artifact name of the signed provenance. The file must have the `.intoto` extension.<br>Defaults to `<filename>.intoto` for single artifact or `multiple.intoto.jsonl` for multiple artifacts.                                    |
| `rekor-log-public`             | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false`                                     |
| `registry-username`              | Username to log in the container registry.                                                                                                                                                                 |
| `gcp-workload-identity-provider` | The full identifier of the Workload Identity Provider, including the project number, pool name, and provider name. If provided, this must be the full identifier which includes all parts:<br>`projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider` |
| `gcp-service-account`            | Email address or unique identifier of the Google Cloud service account for which to generate credentials. For example:<br>`my-service-account@my-project.iam.gserviceaccount.com`                                                                                                       |
| `upload-assets`            |  Uploads the artifact and provenance to a GitHub release.<br>If the `upload-tag-name` was provided, then the assets are uploaded to the provided input tag. This can be used for workflow_dispatch events. Otherwise, if a new tag triggered the workflow, then the assets are uploaded to the triggering tag.           |
| `upload-tag-name`            | If non-empty and `upload-assets` is set to true, the provenance is uploaded to the GitHub release identified by the tag name. If a workflow is run on a new tag and `upload-tag-name` is non-empty, the new tag is ignored and the value of `upload-tag-name` is used instead to upload the assets.              |
| `prerelease`            | EIf true, GitHub Release is created as a pre-release.                                                                                                     |

Secrets:

| Name                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `registry-username` | Username to log in the container registry. This should only be used for high entropy values such as AWS Access Key as described [here](https://github.com/docker/login-action#aws-elastic-container-registry-ecr). Normal username values could match other input values and cause them to be ignored by GitHub Actions and causing your build to fail. In those cases, use the `registry-username` input instead. |
| `registry-password` | Password to log in the container registry. Required if a `registry-username` is provided |

### Workflow Example

Create a new workflow, e.g., `.github/workflows/slsa-build.yml`.

```yaml
name: SLSA container-based releaser
on:
  workflow_dispatch:
  push:
    tags:
      - "*"

permissions: read-all

jobs:
  # Trusted builder.
  build:
    permissions:
      id-token: write # To sign the provenance.
      contents: write # To upload assets to release.
      actions: read # To read the workflow path.
    needs: args
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@v1.5.0
    with:
      builder-image: "bash"
      builder-digest: "sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"
      config-path: ".github/configs-docker/config.toml"
```

### Workflow Outputs

The [container-based
workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/builder_docker-based_slsa3.yml) produces the following outputs:

| Name               | Description                                                                            |
| ------------------ | -------------------------------------------------------------------------------------- |
| `build-outputs-name`  | The name of the artifact where the generated artifacts are uploaded to the artifact registry.       |
| `attestations-download-name`          | Name of the artifact to download all the attestations. When run on a `pull_request` trigger, attestations are not signed and have an `.intoto` extension. When run on other triggers, attestations are signed and have an `.intoto.sigstore` extension. |

### Provenance Format

The `buildDefinition` contains the following fields:

| Name                         | Value                                                                    | Description                                                                                                                                                                                                            |
| -------------------- | ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://slsa.dev/container-based-build/v0.1?draft"` | Identifies the container-based build type.                                                                                                                                                                                 |
| `externalParameters.source` | `slsa.ArtifactReference`                                               | An artifact reference specifying the location of the source repository. |
| `externalParameters.builderImage` | `slsa.ArtifactReference`                                               | An artifact reference specifying the container base image used to build the artifacts. |
| `externalParameters.configPath` | `".github/configs-docker/config.toml"`                                               | The location of the configuration file, relative to the root of the source repository. |
| `externalParameters.buildConfig` | JSON object                                               | An object describing the build configuration. |
| `externalParameters.buildConfig.ArtifactPath` | `"dist/**"`                                               | The path describing the output artifacts to attest to and upload. |
| `externalParameters.buildConfig.Command` | `"["npm", "run", "all"]"`                                               | The build command invoked in the container image to produce the output artifacts. |
| `externalParameters.resolvedDependencies` | `slsa.ArtifactReference`                                               | An artifact reference specifying the binary used by the reusable workflow to build the artifact and generate the build definition. See the [CLI tool](#command-line-tool) below. |

### Provenance Example

The following is an example of the generated provenance. Provenance is generated
as an [in-toto](https://in-toto.io/) statement with a SLSA predicate.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "subject": [
    {
      "name": "example.js",
      "digest": {
        "sha256": "5d672b0dbb696a3289632bf241cc4bb08dbb32c3e9559ea7e9f96b0490209891"
      }
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v1.0?draft",
  "predicate": {
    "buildDefinition": {
      "buildType": "https://slsa.dev/container-based-build/v0.1?draft",
      "externalParameters": {
        "source": {
          "uri": "git+https://github.com/slsa-framework/example-package@refs/heads/main",
          "digest": {
            "sha1": "ca220e54c07b6fcdd758184a12c132ee3ae531f1"
          }
        },
        "builderImage": {
          "uri": "bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9",
          "digest": {
            "sha256": "9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9"
          }
        },
        "configPath": ".github/configs-docker/config.toml",
        "buildConfig": {
          "ArtifactPath": "bin/**",
          "Command": [
            "npm",
            "run",
            "all"
          ]
        }
      },
      "resolvedDependencies": [
        {
          "uri": "git+https://github.com/slsa-framework/slsa-github-generator@refs/tags/v1.5.0",
          "digest": {
            "sha256": "f590fb74e36c9ed8026cceab415a5bebab9bdf3268ace6bd4765c4361c4415c1"
          }
        }
      ],
      "systemParameters": {
        "GITHUB_EVENT_NAME": "workflow_dispatch",
        "GITHUB_JOB": "provenance",
        "GITHUB_REF": "refs/heads/main",
        "GITHUB_REF_TYPE": "branch",
        "GITHUB_REPOSITORY": "slsa-framework/example-package",
        "GITHUB_RUN_ATTEMPT": "1",
        "GITHUB_RUN_ID": 4310284899,
        "GITHUB_RUN_NUMBER": 50,
        "GITHUB_SHA": "ca220e54c07b6fcdd758184a12c132ee3ae531f1",
        "GITHUB_WORKFLOW": ".github/workflows/e2e.docker-based.workflow_dispatch.main.default.slsa3.yml",
        "GITHUB_WORKFLOW_REF": "slsa-framework/example-package/.github/workflows/e2e.docker-based.workflow_dispatch.main.default.slsa3.yml@refs/heads/main",
        "GITHUB_WORKFLOW_SHA": "ca220e54c07b6fcdd758184a12c132ee3ae531f1",
        "IMAGE_OS": "ubuntu22",
        "IMAGE_VERSION": "20230227.2",
        "RUNNER_ARCH": "X64",
        "RUNNER_NAME": "GitHub Actions 9",
        "RUNNER_OS": "Linux",
        "GITHUB_ACTOR_ID": "41898282",
        "GITHUB_REPOSITORY_ID": "486325809",
        "GITHUB_REPOSITORY_OWNER_ID": "80431187",
        "GITHUB_EVENT_PAYLOAD": {}
      }
    },
    "runDetails": {
      "builder": {
        "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_docker-based_slsa3.yml@refs/tags/v1.5.0"
      },
      "metadata": {
        "invocationId": "https://github.com/slsa-framework/example-package/actions/runs/4310284899/attempts/1"
      }
    }
  }
}
```

## Command line tool

This folder contains a command line tool for building artifacts using a Docker
image.

It is meant to be used as part of a GitHub Actions reusable workflow for
generating SLSA provenances. However, users can also run the command locally to
test their builds. When a build is intended to be reproducible, consumers may
also use the `verify` sub-command to reproduce the build from the SLSA
provenance.

The command line tool provides three sub-commands, namely `dry-run`, `build`, and
`verify`.

### The `dry-run` subcommand

The `dry-run` subcommand can be used to validate the inputs. If the inputs are
valid, then the tool creates a `BuildDefinition` and stores that as a JSON
document in the output path that must be provided as one of the flags to the
command. The following is an example, which assumes you are running the code in
`internal/builders/docker`:

```bash
go run *.go  dry-run \
  --build-config-path internal/builders/docker/testdata/config.toml \
  --builder-image bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9 \
  --git-commit-digest sha1:cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00 \
  --source-repo git+https://github.com/slsa-framework/slsa-github-generator \
  --build-definition-path bd.json \
  --force-checkout
```

The output of this is a JSON document stored in `bd.json`.

### The `build` subcommand

The `build` subcommand takes more or less the same inputs as the `dry-run`
subcommand, but actually builds the artifacts. To successfully run this
command, you need to have [rootless Docker installed](https://docs.docker.com/engine/security/rootless/).

The following is an example:

```bash
go run *.go build \
  --build-config-path internal/builders/docker/testdata/config.toml \
  --builder-image bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9 \
  --git-commit-digest sha1:cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00 \
  --source-repo git+https://github.com/slsa-framework/slsa-github-generator \
  --subjects-path subjects.json \
  --output-folder /tmp/build-outputs \
  --force-checkout
```

If the build is successful, this command will generate `subjects.json`
containing a JSON-encoded list of generated artifacts and their SHA256 digests.
It also writes all artifacts to the `output-folder`.

### The `verify` command

The `verify` subcommand takes the path to a SLSAv1.0 provenance and verifies it,
by rebuilding the artifacts using the build definition in the provenance, and
checking that the resulting artifacts have the same names and subjects as the
ones in the provenance subject.

Here is an example:

```bash
go run *.go verify --provenance-path testdata/slsa1-provenance.json
```
