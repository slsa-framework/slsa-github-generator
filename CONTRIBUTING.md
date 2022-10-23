# CONTRIBUTING

- [Contributor Guide](#contributor-guide)
  - [Prerequisites](#prerequisites)
    - [Create a GitHub account](#create-a-github-account)
    - [Sign the CLA](#sign-the-cla)
    - [Code of Conduct](#code-of-conduct)
    - [Setting up your development environment](#setting-up-your-development-environment)
  - [Development](#development)
  - [Testing](#testing)
    - [Pre-submits and Unit Tests](#pre-submits-and-unit-tests)
    - [End to End (e2e) Tests](#end-to-end-e2e-tests)
    - [Linters](#linters)
  - [Community](#community)
    - [Communication](#communication)

## Contributor Guide

This guide will help you understand the overall organization of the
[slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator)
project, and direct you to the best places to get started contributing. You'll
be able to pick up issues, write code to fix them, and get your work reviewed
and merged.

This document is the single source of truth for how to contribute to the code
base. Feel free to browse the [open issues](https://github.com/slsa-framework/slsa-github-generator/issues) and file new ones, all feedback
is welcome!

### Prerequisites

Before submitting code, you should first complete the following
prerequisites.

#### Create a GitHub account

Before you get started, you will need to [signup](http://github.com/signup) for a GitHub user account.

#### Sign the CLA

> TODO: is this needed?

Before you can contribute to Kubernetes, you will need to sign the
[Contributor License Agreement].

#### Code of Conduct

> TODO: Is there a CoC for the OpenSSF that we can point to?

Please make sure to read and observe the [Code of Conduct] and
[Community Values]

#### Setting up your development environment

It is not required to set up a developer environment in order to contribute to
slsa-github-generator but it may be required for code changes.

slsa-github-generator uses primarily [Go](https://go.dev/) and
[TypeScript](https://www.typescriptlang.org/) programming languages. However
much of the logic of the project is implemented in GitHub Actions workflows that
are written in [YAML](https://yaml.org/) and make heavy use of Bash scripts.

- [Install Go](https://go.dev/doc/install)
- [Install Node.js](https://nodejs.org/en/download/)

This project also uses several linters in order to maintain code quality. If you
wish to run these linters locally, follow the instructions for each of these to
install them on your development machine.

- [yamllint](https://yamllint.readthedocs.io/)
- [golangci-lint](https://golangci-lint.run/)
- [shellcheck](https://www.shellcheck.net/)
- [eslint](https://eslint.org/) (NOTE: eslint is installed automatically so you
  don't need to install it)

### Development

Since this project includes reusable workflows for use on GitHub Actions local
development is limited to building and testing the binaries used by the reusable
workflows. The workflows themselves must be tested in your own fork.

Local commands that can be used for development are defined in the
[Makefile](./Makefile). You can list the available targets by running `make`.

```
make
```

Most workflows are actually run when pushing to GitHub so in order to test that
a code change is working you may want to set up a GitHub repository for testing.
This repository should have some workflows that call the actions or reusable
workflows you are testing.

Some example test repos:

- [actions-test](https://github.com/ianlewis/actions-test/tree/main/.github/workflows)
- [slsa-on-github-test](https://github.com/laurentsimon/slsa-on-github-test/tree/main/.github/workflows)

### Testing

A number of automated tests and linters are used to maintain stability and
good code quality. New PRs that include new functionality should include
automated tests for that functionality.

#### Pre-submits and Unit Tests

Pre-submits run on each Pull Request and will block it from being merged if
they fail. These tests are located in the [.github/workflows](.github/workflows)
directory and begin with the prefix `pre-sumbit`.

Unit tests are run as pre-submit tests in the
[pre-submit.units.yml](.github/workflows/pre-submit.units.yml) file. You can run
unit tests locally using `make`. This requires that the Go runtime be installed.

```
make unit-test
```

#### End-to-End (e2e) Tests

This project has a number of End-to-End tests that are scheduled to run daily.
These tests are located in the
[example-package](https://github.com/slsa-framework/example-package/tree/main/.github/workflows)
repository and include a number of testing workflows. Please read the e2e
testing
[README.md](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/README.md)
for more information about e2e tests.

#### Linters

You can run all linters using `make`.

```
make lint
```

These linters will also run as GitHub checks for pull requests via
[pre-submit.lint.yml](.github/workflows/pre-submit.lint.yml) file.

### Releasing

Please see [RELEASE.md](RELEASE.md).

## Community

This project is part of the [SLSA Community](https://slsa.dev/community) working
with the SLSA Tooling SIG.

### Communication

The `#slsa-tooling` channel in the [OpenSSF Slack](https://slack.openssf.org/)
is used for communication and sharing ideas.

Communication is also done over
[GitHub issues](https://github.com/slsa-framework/slsa-github-generator/issues).

[open issues]: https://github.com/slsa-framework/slsa-github-generator/issues?q=is%3Aissue+is%3Aopen+label%3Atype%2Fdocumentation
