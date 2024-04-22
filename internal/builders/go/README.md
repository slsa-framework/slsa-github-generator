# Generation of SLSA3+ provenance for Go projects

This document explains how to use the builder for [Go](https://go.dev/) projects.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Generation](#generation)
  - [Referencing the SLSA builder](#referencing-the-slsa-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Configuration File](#configuration-file)
  - [Migration from GoReleaser](#migration-from-goreleaser)
  - [Multi-platform builds](#multi-platform-builds)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Workflow Example](#workflow-example)
  - [Provenance Example](#provenance-example)
  - [BuildConfig Format](#buildconfig-format)
- [Known Issues](#known-issues)
  - [error updating to TUF remote mirror: tuf: invalid key](#error-updating-to-tuf-remote-mirror-tuf-invalid-key)
  - [Compatibility with `actions/download-artifact`](#compatibility-with-actionsdownload-artifact)

<!-- tocstop -->

---

## Generation

The Go builder workflow uses a GitHub Actions reusable workflow to generate the
provenance.

### Referencing the SLSA builder

At present, the trusted builder **MUST** be referenced
by a tag of the form `@vX.Y.Z`, because the build will fail if you reference it via a shorter tag like `@vX.Y` or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
https://rekor.sigstore.dev/. This entry includes the repository name. This will cause the
private repository name to leak and be discoverable via the public Rekor API
server.

If this is ok with you, you can set the `private-repository` flag in order to
opt in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  private-repository: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue #372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

### Supported Triggers

The following [GitHub trigger events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows) are fully supported and tested:

- `schedule`
- `push` (including new tags)
- `release`
- Manual run via `workflow_dispatch`

In practice, most triggers should work with the exception of `pull_request`. If you would like support for `pull_request`, please tell us about your use case on [issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If you have an issue with any other triggers please submit a [new issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Configuration File

Define a configuration file called `.slsa-goreleaser.yml` in the root of your project.

```yaml
# Version for this file.
version: 1

# (Optional) List of env variables used during compilation.
env:
  - GO111MODULE=on
  - CGO_ENABLED=0

# (Optional) Flags for the compiler.
flags:
  - -trimpath
  - -tags=netgo

# The OS to compile for. `GOOS` env variable will be set to this value.
goos: linux

# The architecture to compile for. `GOARCH` env variable will be set to this value.
goarch: amd64

# (Optional) Entrypoint to compile.
# main: ./path/to/main.go

# (Optional) Working directory. (default: root of the project)
# dir: ./relative/path/to/dir

# Binary output name.
# {{ .Os }} will be replaced by goos field in the config file.
# {{ .Arch }} will be replaced by goarch field in the config file.
binary: binary-{{ .Os }}-{{ .Arch }}

# (Optional) ldflags generated dynamically in the workflow, and set as the `evaluated-envs` input variables in the workflow.
ldflags:
  - "-X main.Version={{ .Env.VERSION }}"
  - "-X main.Commit={{ .Env.COMMIT }}"
  - "-X main.CommitDate={{ .Env.COMMIT_DATE }}"
  - "-X main.TreeState={{ .Env.TREE_STATE }}"
```

### Migration from GoReleaser

If you are already using GoReleaser, you may be able to migrate to our builder using multiple config files for each build. However, this is cumbersome and we are working on supporting multiple builds in a single config file for future releases.

In the meantime, you can use both GoReleaser and this builder in the same repository. For example, you can select one build you would like to start generating provenance for. GoReleaser and this builder can co-exist without interfering with one another, so long as the resulting binaries have different names (e.g., when building for different OS/Arch). If you want to keep the same name, you can use the GoReleaser `ignore` option in the `.goreleaser.yml`:

```yaml
builds:
# ...
goos:
  - windows
  - linux
  - darwin
goarch:
  - amd64
  - arm64
  - s390x
# This instructs GoReleaser to not build for linux amd64.
ignore:
  - goos: linux
    goarch: amd64
```

The configuration file accepts many of the common fields GoReleaser uses, as you can see in the [example](#configuration-file). The configuration file also supports two variables: `{{ .Os }}` and `{{ .Arch }}`. Other variables can be set manually as shown in the table below, in combination with the builder's `evaluated-envs`:

| Name                 | Value                                                                                                                            | Example                                    |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| `{{ .CommitDate }}`  | `date -d @$(git log --date=iso8601-strict -1 --pretty=%ct)`                                                                      | `Mon Jun 13 01:23:36 AM UTC 2022`          |
| `{{ .FullCommit }}`  | `$GITHUB_SHA` or `$(git rev-parse HEAD)`                                                                                         | `b2a980888f359b8cef22cb61f153746e1a06deb0` |
| `{{ .ShortCommit }}` | `$(echo $GITHUB_SHA \| cut -c1-8)` or `$(git rev-parse HEAD \| cut -c1-8)`                                                       | `b2a98088`                                 |
| `{{ .Version }}`     | `$(git describe --tags --always --dirty \| cut -c2-)` or `$(echo $GITHUB_REF_NAME \| cut -c2-)` on new tags and release triggers | `1.2.3-alpha+b2a98088`                     |
| `{{ .Tag }}`         | `$GITHUB_REF_NAME` (on `release` and `push` new tag triggers) or `$(git describe --tags --always --dirty \| cut -c2-)`           | `v1.2.3-alpha+b2a98088`                    |
| `{{ .Major }}`       | `$(git describe --tags --always --dirty \| cut -d '.' -f1 \| cut -c2-)`                                                          | `1`                                        |
| `{{ .Minor }}`       | `$(git describe --tags --always --dirty \| cut -d '.' -f2`                                                                       | `2`                                        |
| `{{ .Patch }}`       | `$(git describe --tags --always --dirty \| cut -d '.' -f3 \| cut -d '-' -f1 \| cut -d '+' -f1`                                   | `3`                                        |

If you think you need support for other variables, please [open an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new).

### Multi-platform builds

It's easy to generate binaries for multiple platforms. To accomplish this, we can use the [maxtrix functionality](https://docs.github.com/en/actions/using-jobs/using-a-matrix-for-your-jobs) of github actions and several config files. In the below code sample, you'll see that we have a `strategy` section which lists the platforms and architectures to build for. These reference a `config-file` property at the bottom which will select the correct config for that platform.

```yaml
build:
  permissions:
    id-token: write # To sign the provenance.
    contents: write # To upload assets to release.
    actions: read # To read the workflow path.
  strategy:
    matrix:
      os:
        - linux
        - windows
        - darwin
      arch:
        - amd64
        - arm64
  uses: slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@v2.0.0
  with:
    go-version: 1.19
    config-file: .slsa-goreleaser/${{matrix.os}}-${{matrix.arch}}.yml
    # ... your other stuff here.
```

### Workflow Inputs

The builder workflow [slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/builder_go_slsa3.yml) accepts the following inputs:

| Name                 | Required | Default                                 | Description                                                                                                                                                                                                                                               |
| -------------------- | -------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config-file`        | no       | `.github/workflows/slsa-goreleaser.yml` | The configuration file for the builder. A path within the calling repository.                                                                                                                                                                             |
| `evaluated-envs`     | no       | empty value                             | A list of environment variables, seperated by `,`: `VAR1: value, VAR2: value`. This is typically used to pass dynamically-generated values, such as `ldflags`. Note that only environment variables with names starting with `CGO_` or `GO` are accepted. |
| `go-version`         | no       |                                         | The go version for your project. This value is passed, unchanged, to the [actions/setup-go](https://github.com/actions/setup-go) action when setting up the environment. One of `go-version` or `go-version-file` is required.                            |
| `go-version-file`    | no       |                                         | The go version file (e.g. `go.mod`) for your project. This value is passed, unchanged, to the [actions/setup-go](https://github.com/actions/setup-go) action when setting up the environment. One of `go-version` or `go-version-file` is required.       |
| `upload-assets`      | no       | true on new tags                        | Whether to upload assets to a GitHub release or not.                                                                                                                                                                                                      |
| `upload-tag-name`    | no       |                                         | If specified and `upload-assets` is set to true, the provenance will be uploaded to a Github release identified by the tag-name regardless of the triggering event.                                                                                       |
| `prerelease`         | no       |                                         | If specified and `upload-assets` is set to true, the release is created as prerelease.                                                                                                                                                                    |
| `private-repository` | no       | false                                   | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).                           |
| `draft-release`      | no       | false                                   | If true, the release is created as a draft                                                                                                                                                                                                                |

### Workflow Outputs

The builder workflow [slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/builder_go_slsa3.yml) provides the following outputs:

| Name                 | Description                                                                           |
| -------------------- | ------------------------------------------------------------------------------------- |
| `go-binary-name`     | The name of the generated binary uploaded to the artifact registry.                   |
| `go-provenance-name` | The artifact name of the signed provenance. (A file with the intoto.jsonl extension). |

### Workflow Example

Create a new workflow, e.g., `.github/workflows/slsa-goreleaser.yml`.

**Note**: Make sure that you reference the trusted builder with a semantic version of the form `@vX.Y.Z`. More information [here](/README.md#referencing-slsa-builders-and-generators).

```yaml
name: SLSA go releaser
on:
  workflow_dispatch:
  push:
    tags:
      - "*"

permissions: read-all

jobs:
  # Generate ldflags dynamically.
  # Optional: only needed for ldflags.
  args:
    runs-on: ubuntu-latest
    outputs:
      commit-date: ${{ steps.ldflags.outputs.commit-date }}
      commit: ${{ steps.ldflags.outputs.commit }}
      version: ${{ steps.ldflags.outputs.version }}
      tree-state: ${{ steps.ldflags.outputs.tree-state }}
    steps:
      - id: checkout
        uses: actions/checkout@ec3a7ce113134d7a93b817d10a8272cb61118579 # tag=v2.3.4
        with:
          fetch-depth: 0
      - id: ldflags
        run: |
          echo "commit-date=$(git log --date=iso8601-strict -1 --pretty=%ct)" >> "$GITHUB_OUTPUT"
          echo "commit=$GITHUB_SHA" >> "$GITHUB_OUTPUT"
          echo "version=$(git describe --tags --always --dirty | cut -c2-)" >> "$GITHUB_OUTPUT"
          echo "tree-state=$(if git diff --quiet; then echo "clean"; else echo "dirty"; fi)" >> "$GITHUB_OUTPUT"

  # Trusted builder.
  build:
    permissions:
      id-token: write # To sign the provenance.
      contents: write # To upload assets to release.
      actions: read # To read the workflow path.
    needs: args
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@v2.0.0
    with:
      go-version: 1.17
      # Optional: only needed if using ldflags.
      evaluated-envs: "COMMIT_DATE:${{needs.args.outputs.commit-date}}, COMMIT:${{needs.args.outputs.commit}}, VERSION:${{needs.args.outputs.version}}, TREE_STATE:${{needs.args.outputs.tree-state}}"
```

### Provenance Example

An example of the provenance generated from this repo is below:

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "binary-linux-amd64",
      "digest": {
        "sha256": "7bf2e6ebb97e1bdb669d9df73048247f141e2f8e72ab59f23d456f1bc5a041dc"
      }
    }
  ],
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@v1.4.0"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator/go@v1",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/heads/main",
        "digest": {
          "sha1": "d29d1701b47bbbe489e94b053611e5a7bf6d9414"
        },
        "entryPoint": ".github/workflows/release.yml"
      },
      "parameters": {},
      "environment": {
        "github_actor": "ianlewis",
        "github_actor_id": "123456",
        "github_base_ref": "",
        "github_event_name": "workflow_dispatch",
        "github_event_payload": ...,
        "github_head_ref": "",
        "github_ref": "refs/heads/main",
        "github_ref_type": "branch",
        "github_repository_id": "8923542",
        "github_repository_owner": "ianlewis",
        "github_repository_owner_id": "123456",
        "github_run_attempt": "1",
        "github_run_id": "2193104371",
        "github_run_number": "16",
        "github_sha1": "d29d1701b47bbbe489e94b053611e5a7bf6d9414"
      }
    },
    "buildConfig": {
      "version": 1,
      "steps": [
        {
          "command": [
            "/opt/hostedtoolcache/go/1.17.10/x64/bin/go",
            "mod",
            "vendor"
          ],
          "env": null,
          "workingDir": "/home/runner/work/ianlewis/actions-test"
        },
        {
          "command": [
            "/opt/hostedtoolcache/go/1.17.10/x64/bin/go",
            "build",
            "-mod=vendor",
            "-trimpath",
            "-tags=netgo",
            "-o",
            "binary-linux-amd64-config1"
          ],
          "env": [
            "GOOS=linux",
            "GOARCH=amd64",
            "GO111MODULE=on",
            "CGO_ENABLED=0"
          ],
          "workingDir": "/home/runner/work/ianlewis/actions-test"
        }
      ]

    },
    "metadata": {
      "completeness": {
        "parameters": true,
        "environment": false,
        "materials": false
      },
      "reproducible": false
    },
    "materials": [
      {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/heads/main",
        "digest": {
          "sha1": "d29d1701b47bbbe489e94b053611e5a7bf6d9414"
        }
      }
    ]
  }
}
```

### BuildConfig Format

The `BuildConfig` contains the following fields:

`version`: The version of the `BuildConfig` format.

`steps`: The steps that were performed in the build.

`steps[*].command`: The list of commands that were executed in a step.

```json
  "command": [
"/opt/hostedtoolcache/go/1.17.10/x64/bin/go",
"mod",
"vendor"
],
```

`steps[*].env`: Any environment variables used in the command, including any OS environment variables and those set in the configuration file.

```json
  "env": [
"GOOS=linux",
"GOARCH=amd64",
"GO111MODULE=on",
"CGO_ENABLED=0"
],
```

`steps[*].workingDir`: The working directory where the steps were performed in the runner.

```json
  "workingDir": "/home/runner/work/ianlewis/actions-test"
```

## Known Issues

### error updating to TUF remote mirror: tuf: invalid key

**Affected versions:** v1.2.x

Workflows are currently failing with the error:

```text
validating log entry: unable to fetch Rekor public keys from TUF repository, and not trusting the Rekor API for fetching public keys: updating local metadata and targets: error updating to TUF remote mirror: tuf: invalid key
```

This issue is currently tracked by [issue #1163](https://github.com/slsa-framework/slsa-github-generator/issues/1163)

You can work around this error by setting `compile-builder` input flag.

```yaml
with:
  compile-builder: true
```

This will compile the builder binary used by the workflow instead of downloading
the latest release. Make sure you continue to reference the workflow using a
release tag in order to allow verification by `slsa-verifier`.

```yaml
uses: slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@v2.0.0
```

### Compatibility with `actions/download-artifact`

To download provenance (e.g., if you don't use `upload-assets`) you have to
use [`actions/download-artifact@v3`](https://github.com/actions/download-artifact).
The workflow uses [`actions/upload-artifact@3`](https://github.com/actions/upload-artifact)
which is
[not compatible](https://github.com/actions/download-artifact?tab=readme-ov-file#breaking-changes)
with `actions/download-artifact@v4`.
