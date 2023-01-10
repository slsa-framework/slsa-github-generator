<!-- markdown-toc --bullets="-" -i CHANGELOG.md -->

<!-- toc -->

- [v1.5.0](#v150)
  - [Summary of changes](#summary-of-changes)
    - [Go builder](#go-builder)
      - [New Features](#new-features)
    - [Generic generator](#generic-generator)
      - [New Features](#new-features-1)
    - [Container generator](#container-generator)
      - [New Features](#new-features-2)
  - [Changelog since v1.4.0](#changelog-since-v140)

<!-- tocstop -->

# v1.5.0

## Summary of changes

### Go builder

#### New Features

- A new [`upload-tag-name`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/generic/README.md#workflow-inputs) input was added to allow users to specify the tag name for the release when `upload-assets` is set to `true`.
- The environment variables included in provenance output were changed to include only those variables that are specified by the user in the [slsa-goreleaser.yml configuration file](https://github.com/slsa-framework/slsa-github-generator/tree/v1.5.0-rc.0/internal/builders/go#configuration-file) in order to improve reproducibility. See [#822](https://github.com/slsa-framework/slsa-github-generator/issues/822) for more information and background.

### Generic generator

#### New Features

- A new boolean [`continue-on-error`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/generic/README.md#workflow-inputs) input was added which, when set to `true`, prevents the workflow from failing when a step fails. If set to true, the result of the reusable workflow will be return in the [`outcome`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/generic/README.md#workflow-outputs) output.
- A new [`upload-tag-name`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/generic/README.md#workflow-inputs) input was added to allow users to specify the tag name for the release when `upload-assets` is set to `true`.

### Container generator

#### New Features

- A new boolean [`continue-on-error`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/container/README.md#workflow-inputs) input was added which, when set to `true`, prevents the workflow from failing when a step fails. If set to true, the result of the reusable workflow will be return in the [`outcome`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/container/README.md#workflow-outputs) output.
- TODO: Add `upload-tag-name`
- A new [`repository-username`](https://github.com/slsa-framework/slsa-github-generator/blob/v1.5.0-rc.0/internal/builders/container/README.md#workflow-inputs) secret input was added to allow users to pass their repository username that is stored in a [Github Actions encrypted secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets).

## Changelog since v1.4.0

https://github.com/slsa-framework/slsa-github-generator/compare/v1.4.0...v1.5.0-rc.0
