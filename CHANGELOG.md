# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- markdown-toc --bullets="-" -i CHANGELOG.md -->

<!-- toc -->

- [v2.0.0](#v200)
  - [v2.0.0: Breaking Change: upload-artifact and download-artifact](#v200-breaking-change-upload-artifact-and-download-artifact)
  - [v2.0.0: Breaking Change: attestation-name Workflow Input and Output](#v200-breaking-change-attestation-name-workflow-input-and-output)
  - [v2.0.0: DSSE Rekor Type](#v200-dsse-rekor-type)
- [v1.10.0](#v1100)
  - [v1.10.0: TUF fix](#v1100-tuf-fix)
  - [v1.10.0: Gradle Builder](#v1100-gradle-builder)
  - [v1.10.0: Go Builder](#v1100-go-builder)
  - [v1.10.0: Container Generator](#v1100-container-generator)
- [v1.9.0](#v190)
  - [v1.9.0: BYOB framework (beta)](#v190-byob-framework-beta)
  - [v1.9.0: Maven builder (beta)](#v190-maven-builder-beta)
  - [v1.9.0: Gradle builder (beta)](#v190-gradle-builder-beta)
  - [v1.9.0: JReleaser builder](#v190-jreleaser-builder)
- [v1.8.0](#v180)
  - [v1.8.0: Generic Generator](#v180-generic-generator)
  - [v1.8.0: Node.js Builder (beta)](#v180-nodejs-builder-beta)
- [v1.7.0](#v170)
  - [v1.7.0: Go builder](#v170-go-builder)
- [v1.6.0](#v160)
  - [Summary of changes](#summary-of-changes)
    - [Go builder](#go-builder)
      - [New Features](#new-features)
    - [Generic generator](#generic-generator)
      - [New Features](#new-features-1)
    - [Container generator](#container-generator)
  - [Changelog since v1.5.0](#changelog-since-v150)
- [v1.5.0](#v150)
  - [Summary of changes](#summary-of-changes-1)
    - [Go builder](#go-builder-1)
      - [New Features](#new-features-2)
    - [Generic generator](#generic-generator-1)
      - [New Features](#new-features-3)
    - [Container generator](#container-generator-1)
      - [New Features](#new-features-4)
  - [Changelog since v1.4.0](#changelog-since-v140)
- [v1.4.0](#v140)
  - [What's Changed](#whats-changed)
    - [Generic Generator](#generic-generator)
      - [Bug fixes](#bug-fixes)
    - [Go Builder](#go-builder)
      - [Bug fixes](#bug-fixes-1)
  - [New Contributors](#new-contributors)
  - [Full Changelog](#full-changelog)
- [v1.4.0-rc.2](#v140-rc2)
  - [What's Changed](#whats-changed-1)
  - [New Contributors](#new-contributors-1)
  - [Full Changelog](#full-changelog-1)
- [v1.4.0-rc.1](#v140-rc1)
  - [What's Changed](#whats-changed-2)
  - [New Contributors](#new-contributors-2)
  - [Full Changelog](#full-changelog-2)
- [v1.4.0-rc.0](#v140-rc0)
  - [What's Changed](#whats-changed-3)
  - [New Contributors](#new-contributors-3)
  - [Full Changelog](#full-changelog-3)
- [v1.2.2](#v122)
  - [What's Changed](#whats-changed-4)
  - [New Contributors](#new-contributors-4)
  - [Full Changelog](#full-changelog-4)
- [v1.2.1](#v121)
  - [What's Changed](#whats-changed-5)
    - [Generic generator](#generic-generator-2)
      - [buildType](#buildtype)
      - [Provenance file names](#provenance-file-names)
      - [Explicit opt-in for private repos](#explicit-opt-in-for-private-repos)
    - [Go builder](#go-builder-2)
      - [Support private repos](#support-private-repos)
  - [New Contributors](#new-contributors-5)
  - [Full Changelog](#full-changelog-5)
- [v1.2.0](#v120)
  - [What's Changed](#whats-changed-6)
    - [Generic generator](#generic-generator-3)
    - [Go builder](#go-builder-3)
  - [New Contributors](#new-contributors-6)
  - [Full Changelog](#full-changelog-6)
- [v1.1.1](#v111)
  - [What's Changed](#whats-changed-7)
  - [New Contributors](#new-contributors-7)
  - [Full Changelog](#full-changelog-7)
- [v1.0.0](#v100)
  - [What's Changed](#whats-changed-8)
  - [Contributors](#contributors)

<!-- tocstop -->

<!--
Information on the next release will be added here.

Use the format "X.Y.Z: Go builder" etc. for format headers to avoid header name
duplication."
-->

## v2.0.0

### v2.0.0: Breaking Change: upload-artifact and download-artifact

- Our workflows now use the new `@v4`s of `actions/upload-artifact` and
  `actions/download-artifact`, which are incompatiblle with the prior `@v3`. See
  Our docs on the [generic generator](./internal/builders/generic/README.md#compatibility-with-actionsdownload-artifact)
  for more information and how to upgrade.

### v2.0.0: Breaking Change: attestation-name Workflow Input and Output

- `attestation-name` as a workflow input to
  `.github/workflows/generator_generic_slsa3.yml` is now removed. Use
  `provenance-name` instead.

### v2.0.0: DSSE Rekor Type

- When uploading signed provenance to the log, the entry created in the log is now
  a DSSE Rekor type. This fixes a bug where the current intoto type does not
  persist provenance signatures. The attestation will no longer be persisted
  in Rekor ([#3299](https://github.com/slsa-framework/slsa-github-generator/issues/3299))

## v1.10.0

Release [v1.10.0](https://github.com/slsa-framework/slsa-github-generator/releases/tag/v1.10.0) includes bug fixes and new features.

See the [full change list](https://github.com/slsa-framework/slsa-github-generator/compare/v1.9.0...v1.10.0).

### v1.10.0: TUF fix

- The cosign TUF roots were fixed ([#3350](https://github.com/slsa-framework/slsa-github-generator/issues/3350)).
  More details [here](https://github.com/slsa-framework/slsa-github-generator/blob/v1.10.0/README.md#error-updating-to-tuf-remote-mirror-invalid).

### v1.10.0: Gradle Builder

- The Gradle Builder was fixed when the project root is the same as the
  repository root ([#2727](https://github.com/slsa-framework/slsa-github-generator/issues/2727))

### v1.10.0: Go Builder

- The `go-version-file` input was fixed so that it can find the `go.mod` file
  ([#2661](https://github.com/slsa-framework/slsa-github-generator/issues/2661))

### v1.10.0: Container Generator

- A new `provenance-repository` input was added to allow reading provenance from
  a different container repository than the image itself ([#2956](https://github.com/slsa-framework/slsa-github-generator/issues/2956))

## v1.9.0

Release [v1.9.0] includes bug fixes and new features.

See the [full change list](https://github.com/slsa-framework/slsa-github-generator/compare/v1.8.0...v1.9.0).

### v1.9.0: BYOB framework (beta)

- **New**: A [new framework](https://github.com/slsa-framework/slsa-github-generator/blob/main/BYOB.md) to turn GitHub Actions into SLSA compliant builders.

### v1.9.0: Maven builder (beta)

- **New**: A [Maven builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/maven) to build Java projects and publish to Maven central.

### v1.9.0: Gradle builder (beta)

- **New**: A [Gradle builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/gradle) to build Java projects and publish to Maven central.

### v1.9.0: JReleaser builder

- **New**: A [JReleaser builder](https://github.com/jreleaser/release-action/tree/v1.0.0-java) that wraps the official [JReleaser Action](https://github.com/jreleaser/release-action/tree/v1.0.0-java).

## v1.8.0

Release [v1.8.0] includes bug fixes and new features.

See the [full change list](https://github.com/slsa-framework/slsa-github-generator/compare/v1.7.0...v1.8.0).

### v1.8.0: Generic Generator

- **Added**: A new
  [`base64-subjects-as-file`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.8.0/internal/builders/generic/README.md#workflow-inputs)
  was added to allow for specifying a large subject list.

### v1.8.0: Node.js Builder (beta)

- **Fixed**: Publishing for non-scoped packages was fixed (See
  [#2359](https://github.com/slsa-framework/slsa-github-generator/issues/2359))
- **Fixed**: Documentation was updated to clarify that the GitHub Actions
  `deployment` event is not supported.
- **Changed**: The file extension for the generated provenance file was changed
  from `.sigstore` to `.build.slsa` in order to make it easier to identify
  provenance files regardless of file format.
- **Fixed**: The publish action was fixed to address an issue with the package
  name when using Node 16.

## v1.7.0

This release includes the first beta release of the
[Container-based builder](https://github.com/slsa-framework/slsa-github-generator/tree/v1.7.0/internal/builders/docker).
The Container-based builder provides a GitHub Actions reusable workflow that can
be used to invoke a container image with a user-specified command to generate an
artifact and SLSA Build L3 compliant provenance.

### v1.7.0: Go builder

- **Added**: A new
  [go-version-file](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/go/README.md#workflow-inputs)
  input was added. This allows you to specify a go.mod file in order to track
  which version of Go is used for your project.

## v1.6.0

This release includes the first beta release of the
[Node.js builder](https://github.com/slsa-framework/slsa-github-generator/tree/v1.6.0/internal/builders/nodejs).
The Node.js builder provides a GitHub Actions reusable workflow that can be
called to build a Node.js package, generate SLSA Build L3 compliant provenance,
and publish it to the npm registry along with the package.

### Summary of changes

#### Go builder

##### New Features

- A new
  [`prerelease`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.6.0/internal/builders/go/README.md#workflow-inputs)
  input was added to allow users to create releases marked as prerelease when
  `upload-assets` is set to `true`.
- A new input [`draft-release`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.6.0/internal/builders/go/README.md#workflow-inputs) was added to allow users to create releases marked
  as draft when `upload-assets` is set to `true`.
- A new output [`go-provenance-name`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.6.0/internal/builders/go/README.md#workflow-outputs) added which can be used to retrieve the name
  of the provenance file generated by the builder.

#### Generic generator

##### New Features

- A new input [`draft-release`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.6.0/internal/builders/generic/README.md#workflow-inputs) was added to allow users to create releases marked
  as draft when `upload-assets` is set to `true`.

#### Container generator

The Container Generator was updated to use `cosign` v2.0.0. No changes to the
workflow's inputs or outputs were made.

### Changelog since v1.5.0

https://github.com/slsa-framework/slsa-github-generator/compare/v1.5.0...v1.6.0

## v1.5.0

<!-- Information on the next release will be added here. -->

### Summary of changes

#### Go builder

##### New Features

- A new [`upload-tag-name`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/generic/README.md#workflow-inputs) input was added to allow users to specify the tag name for the release when `upload-assets` is set to `true`.
- The environment variables included in provenance output were changed to include only those variables that are specified by the user in the [slsa-goreleaser.yml configuration file](https://github.com/slsa-framework/slsa-github-generator/tree/v1.5.0/internal/builders/go#configuration-file) in order to improve reproducibility. See [#822](https://github.com/slsa-framework/slsa-github-generator/issues/822) for more information and background.

#### Generic generator

##### New Features

- A new boolean [`continue-on-error`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/generic/README.md#workflow-inputs) input was added which, when set to `true`, prevents the workflow from failing when a step fails. If set to true, the result of the reusable workflow will be return in the [`outcome`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/generic/README.md#workflow-outputs) output.
- A new [`upload-tag-name`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/generic/README.md#workflow-inputs) input was added to allow users to specify the tag name for the release when `upload-assets` is set to `true`.

#### Container generator

##### New Features

- A new boolean [`continue-on-error`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/container/README.md#workflow-inputs) input was added which, when set to `true`, prevents the workflow from failing when a step fails. If set to true, the result of the reusable workflow will be return in the [`outcome`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/container/README.md#workflow-outputs) output.
- A new [`repository-username`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/container/README.md#workflow-inputs) secret input was added to allow users to pass their repository username that is stored in a [Github Actions encrypted secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets). This secret input should only be used for high-entropy registry username values such as AWS Access Key.
- Support was added for authenticating with [Google Artifact Registry](https://cloud.google.com/artifact-registry) and [Google Container Registry](https://cloud.google.com/container-registry) using [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation). Users can use this new feature by using the [`gcp-workload-identity-provider` and `gcp-service-account` inputs](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0/internal/builders/container/README.md#workflow-inputs)

### Changelog since v1.4.0

https://github.com/slsa-framework/slsa-github-generator/compare/v1.4.0...v1.5.0

## v1.4.0

### What's Changed

This release is the first Generally Available version of the [Container Generator workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/container). The Container Generator workflow is now considered stable and can be included in your production GitHub Actions workflows

This is also the first release (technically the second) with support for the [generally available version of sigstore](https://blog.sigstore.dev/sigstore-ga-ddd6ba67894d)!!
We hope to have fewer issues with sigstore infrastructure moving forward.

#### Generic Generator

##### Bug fixes

1. Allow users of the [Generic Generator](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/generic) to generate provenance for artifacts created in a project subdirectory (#1225)

#### Go Builder

##### Bug fixes

1. Allow environment variables to contain '=' characters in the [Go builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/go) (#1231)

### New Contributors

- @cfergeau made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1232
- @DanAlbert made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1239
- @gal-legit made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1252

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.2...v1.4.0

## v1.4.0-rc.2

\*_This is a pre-release. It is not meant for general consumption. The following is the proposed release notes for the official release._

### What's Changed

This release is the first Generally Available version of the [generic container workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/container). The generic container workflow is now considered stable and can be included in your production GitHub Actions workflows

This is also the first release with support for the [generally available version of sigstore](https://blog.sigstore.dev/sigstore-ga-ddd6ba67894d)!

This release also includes a couple of bug fixes:

1. Allow users of the [generic generator workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/generic) to generate provenance using for artifacts created in a project subdirectory (#1225)
2. Allow environment variables to contain '=' characters in the [Go workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/go) (#1231)

### New Contributors

- @cfergeau made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1232
- @DanAlbert made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1239
- @gal-legit made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1252

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.2...v1.4.0-rc.2

## v1.4.0-rc.1

\*_This is a pre-release. It is not meant for general consumption. The following is the proposed release notes for the official release._

### What's Changed

This release is the first Generally Available version of the [generic container workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/container). The generic container workflow is now considered stable and can be included in your production GitHub Actions workflows

This is also the first release with support for the [generally available version of sigstore](https://blog.sigstore.dev/sigstore-ga-ddd6ba67894d)!

This release also includes a couple of bug fixes:

1. Allow users of the [generic generator workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/generic) to generate provenance using for artifacts created in a project subdirectory (#1225)
2. Allow environment variables to contain '=' characters in the [Go workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/go) (#1231)

### New Contributors

- @cfergeau made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1232
- @DanAlbert made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1239
- @gal-legit made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1252

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.2...v1.4.0-rc.1

## v1.4.0-rc.0

**This is a pre-release. It is not meant for general consumption. The following is the proposed release notes for the official release.**

### What's Changed

This release is the first Generally Available version of the [generic container workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/container). The generic container workflow is now considered stable and can be included in your production GitHub Actions workflows

This is also the first release with support for the [generally available version of sigstore](https://blog.sigstore.dev/sigstore-ga-ddd6ba67894d)!

This release also includes a couple of bug fixes:

1. Allow users of the [generic generator workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/generic) to generate provenance using for artifacts created in a project subdirectory (#1225)
2. Allow environment variables to contain '=' characters in the [Go workflow](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/go) (#1231)

### New Contributors

- @cfergeau made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1232
- @DanAlbert made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1239
- @gal-legit made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1252

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.2...v1.4.0-rc.0

## v1.2.2

### What's Changed

This release fixes issues with signing provenance due to a change in Sigstore TUF root certificates (#1163). This release also includes better handling of transient errors from the Rekor transparency logs.

### New Contributors

- @suzuki-shunsuke made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1061
- @datosh made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1074
- @pnacht made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1187
- @dongheelee92 made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/1209

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.1...v1.2.2

## v1.2.1

**DO NOT USE THIS RELEASE. This version will no longer work and is not supported due to errors described in #1163. Please upgrade to [v1.2.2](https://github.com/slsa-framework/slsa-github-generator/releases/tag/v1.2.2) or later.**

### What's Changed

This release fixes an error that occurs on the "Generate Builder" step for various workflows.

```text
FAILED: SLSA verification failed: could not find a matching valid signature entry
```

See #942

#### Generic generator

##### buildType

This release changes the [`buildType`](https://slsa.dev/provenance/v0.2#buildType) used in provenance created by the generic generator.

The previous value was:

```json
"buildType": "https://github.com/slsa-framework/slsa-github-generator@v1",
```

The new value is:

```json
"buildType": "https://github.com/slsa-framework/slsa-github-generator/generic@v1",
```

See #627

##### Provenance file names

Previously the default file name for provenance was `attestation.intoto.jsonl`. This has been updated to be in line with [intoto attestation file naming conventions](https://github.com/in-toto/attestation/blob/main/spec/bundle.md#file-naming-convention). The file name now defaults to `<artifact filename>.intoto.jsonl` if there is a single artifact, or `multiple.intoto.jsonl` if there are multiple artifacts.

See #654

##### Explicit opt-in for private repos

Private repository support was enhanced to required the `private-repository` input field as the repository name will be made public in the public Rekor transparency log.

Please add the following to your workflows if you opt into allowing repository names to be recorded in the public Rekor transparency log.

```yaml
with:
  private-repository: true
```

See #823

#### Go builder

##### Support private repos

Support for private repositories was fixed. If using a private repository you must specify the `private-repository` input field as the repository name will be made public in the public Rekor transparency log.

Please add the following to your workflows if you opt into allowing repository names to be recorded in the public Rekor transparency log.

```yaml
with:
  private-repository: true
```

See #823

### New Contributors

- @sethmlarson made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/758
- @yunginnanet made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/776
- @diogoteles08 made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/957

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.2.0...v1.2.1

## v1.2.0

**DO NOT USE THIS RELEASE. This version will no longer work and is not supported due to errors described in #942. Please upgrade to [v1.2.2](https://github.com/slsa-framework/slsa-github-generator/releases/tag/v1.2.2) or later.**

### What's Changed

#### Generic generator

The highlight of this release is a new re-usable workflow called the "Generic generator". It lets users build artifacts on their own and generate a provenance that satisfies SLSA provenance 3 requirement. It's perfect to get started with SLSA with minimal changes to an existing build workflow. To use it, check the [README.md](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/generic/README.md)!

#### Go builder

No changes.

### New Contributors

- @naveensrinivasan made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/352
- @renovate-bot made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/401
- @rarkins made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/489
- @developer-guy made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/497
- @loosebazooka made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/573

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.1.1...v1.2.0

## v1.1.1

### What's Changed

- Improve documentation
- Fix filename issue when resolving it with variables
- Add support for environment variables in artifact filename

### New Contributors

- @joshuagl made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/199
- @mihaimaruseac made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/202
- @MarkLodato made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/312
- @chipzoller made their first contribution in https://github.com/slsa-framework/slsa-github-generator/pull/354

### Full Changelog

https://github.com/slsa-framework/slsa-github-generator/compare/v1.0.0...v1.1.1

## v1.0.0

### What's Changed

This is the first official release of the generator. The first builder we are releasing is for Golang projects.
To learn how to use it, see [./README.md#golang-projects](https://github.com/slsa-framework/slsa-github-generator#golang-projects)

### Contributors

@asraa @ianlewis @MarkLodato @joshuagl @laurentsimon

[v1.8.0]: https://github.com/slsa-framework/slsa-github-generator/releases/tag/v1.8.0
[v1.9.0]: https://github.com/slsa-framework/slsa-github-generator/releases/tag/v1.9.0
