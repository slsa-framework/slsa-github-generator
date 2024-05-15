# Contributor Guide

This guide will help you understand the overall organization of the
[slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator)
project, and direct you to the best places to get started contributing. You'll
be able to pick up issues, write code to fix them, and get your work reviewed
and merged.

This document is the single source of truth for how to contribute to the code
base. Feel free to browse the [open issues] and file new ones, all feedback is
welcome!

<!-- markdown-toc --bullets="-" -i CONTRIBUTING.md -->

<!-- toc -->

- [How can I help?](#how-can-i-help)
- [Development](#development)
  - [Prerequisites](#prerequisites)
    - [Create a GitHub account](#create-a-github-account)
    - [Code of Conduct](#code-of-conduct)
    - [Setting up your development environment](#setting-up-your-development-environment)
  - [Making Changes](#making-changes)
    - [Create a fork](#create-a-fork)
    - [Create a test repo](#create-a-test-repo)
    - [Development Basics](#development-basics)
    - [Make your changes](#make-your-changes)
    - [Run tests](#run-tests)
    - [Submit a PR](#submit-a-pr)
  - [Preferred Languages](#preferred-languages)
- [Testing](#testing)
  - [Pre-submits and Unit Tests](#pre-submits-and-unit-tests)
  - [End-to-End (e2e) Tests](#end-to-end-e2e-tests)
  - [Linters](#linters)
- [Releasing](#releasing)
- [Community](#community)
- [Communication](#communication)

<!-- tocstop -->

## How can I help?

There are many areas in the project that need help. These are managed in GitHub
issues. Please let us know if you are willing to work on the issue and how you
can contribute.

- For new developers and contributors, please see issues labeled
  [good first issue]. These issues should require minimal background knowledge
  to contribute.
- For slightly more involved changes that may require some background knowledge,
  please see issues labeled [help wanted]
- For experienced developers, any of our [open issues] is open to contribution.

If you don't find an existing issue for your contribution feel free to [create a
new
one](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

## Development

This section covers what you need to set up a development environment, how
to make changes, and how to submit them.

Before you decide to make a change, please take some time to identify an
existing issue where you would like to contribute, or create a new issue. See
[How can I help?](#how-can-i-help) for more info.

### Prerequisites

Before submitting code, you should first complete the following
prerequisites.

#### Create a GitHub account

Before you get started, you will need to [signup](http://github.com/signup) for
a GitHub user account.

#### Code of Conduct

Please make sure to read and observe the [Code of Conduct].

#### Setting up your development environment

It is not required to set up a developer environment in order to contribute to
slsa-github-generator but it may be required for code changes.

slsa-github-generator uses primarily [Go](https://go.dev/) and
[TypeScript](https://www.typescriptlang.org/) programming languages. However
much of the logic of the project is implemented in GitHub Actions workflows that
are written in [YAML](https://yaml.org/) and make heavy use of Bash scripts.

- [Install Go](https://go.dev/doc/install)
- [Install Node.js](https://nodejs.org/en/download/)

Some Node.js projects require Jest as a testing framework.

- [Install Jest](https://jestjs.io/docs/getting-started)

This project also uses several linters in order to maintain code quality. If you
wish to run these linters locally, follow the instructions for each of these to
install them on your development machine.

- [yamllint](https://yamllint.readthedocs.io/)
- [golangci-lint](https://golangci-lint.run/)
- [shellcheck](https://www.shellcheck.net/)
- [eslint](https://eslint.org/) (NOTE: eslint is installed automatically so you
  don't need to install it)

### Making Changes

This section will cover how to make changes and test them.

#### Create a fork

Since this project includes reusable workflows for use on GitHub Actions local
development is limited to building and testing the binaries used by the reusable
workflows. The workflows themselves must be tested in your own fork. You can
fork the slsa-github-generator repo via the [Fork
button](https://github.com/slsa-framework/slsa-github-generator/fork).

#### Create a test repo

Most workflows are actually run when pushing to GitHub so in order to test that
a code change is working you may want to set up a GitHub repository for testing.
This repository should have some workflows that call the actions or reusable
workflows you are testing.

Some example test repos:

- [actions-test](https://github.com/ianlewis/actions-test/tree/main/.github/workflows)
- [slsa-on-github-test](https://github.com/laurentsimon/slsa-on-github-test/tree/main/.github/workflows)

#### Development Basics

Local commands that can be used for development are defined in the
[`Makefile`](./Makefile). You can list the available targets by running `make`.

```shell
make
```

#### Make your changes

When making changes in your fork, start by making a new branch.

```shell
git checkout -b my-feature-branch
```

Make your changes and then commit them to your branch. Make sure to sign-off on
your commits. This establishes [Developer Certificate of Origin] which indicates
that you are the author of the changes and are able to submit it under the
project [LICENSE].

```shell
git commit -s
```

You may periodically need to merge with the main branch. Be sure to sign-off on
the merges as well.

```shell
git merge --signoff main
```

#### Run tests

Before you submit your change run the unit tests and linters to ensure your
changes are ready to go. See the [Testing](#testing) section for more info.

#### Updating Github Actions Dependencies

##### Renovate-Bot PRs

`renovate-bot` will periodically send PRs to update the `package.json` and `package-lock.json` in the Github Actions of this repo.
But, it will not also automatically recompile the packages into `.js` files.

We use a Workflow [Update actions dist post-commit](../.github/workflows/update-actions-dist-post-commit.yml) to
help maintainers easily recompile the Github Actions against a PR.

Use the UI to invoke the workflow

[update-actions-dist-post-commit.yml](https://github.com/slsa-framework/slsa-verifier/actions/workflows/update-actions-dist-post-commit.yml)

or invoke with

```shell
gh workflow run update-actions-dist-post-commit.yml -F pr_number=<pull request number>
```

#### Submit a PR

Once your change is ready you can submit a PR via the website.

The title of your PR should include a [Conventional Commits] prefix label. For
example, bug fixes should be prefixed with "fix:". PRs for new features should
be prefixed with "feat:" etc.

Please also include as much information as your can about your change and any
relevant issues in the PR description.

### Preferred Languages

We prefer the use of the following programming languages:

- [Go](https://go.dev/) - For builders/generators.
- [TypeScript](https://www.typescriptlang.org/) - For GitHub actions.
- [Bash](https://www.gnu.org/software/bash/) - For scripts in GitHub actions and workflows.

We do not have a strict rule that proscribes the use of a specific language.
However, any PRs that introduce a new language should include justification for
adding the new language as well as a plan for adding the necessary tooling and
testing infrastructure.

We will generally reject PRs that introduce the use of memory-unsafe
languages or use of existing languages in memory-unsafe ways (e.g. the Go
[unsafe](https://pkg.go.dev/unsafe) package). See the
[Security Policy](./SECURITY.md) for more info.

## Testing

A number of automated tests and linters are used to maintain stability and
good code quality. New PRs that include new functionality should include
automated tests for that functionality.

### Pre-submits and Unit Tests

Pre-submits run on each Pull Request and will block it from being merged if
they fail. These tests are located in the [`.github/workflows`](.github/workflows)
directory and begin with the prefix `pre-sumbit`.

Unit tests are run as pre-submit tests in the
[`pre-submit.units.yml`](.github/workflows/pre-submit.units.yml) file. You can run
unit tests locally using `make`. This requires that the Go runtime be installed.

```shell
make unit-test
```

Tests for GitHub Actions created with Typescript require [Jest](https://jestjs.io).

```shell
make ts-test
```

### End-to-End (e2e) Tests

This project has a number of End-to-End tests that are scheduled to run daily.
These tests are located in the
[`example-package`](https://github.com/slsa-framework/example-package/tree/main/.github/workflows)
repository and include a number of testing workflows. Please read the e2e
testing
[`README.md`](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/README.md)
for more information about e2e tests.

### Linters

You can run all linters using `make`.

```shell
make lint
```

These linters will also run as GitHub checks for pull requests via
[`pre-submit.lint.yml`](.github/workflows/pre-submit.lint.yml) file.

## Releasing

Please see [`RELEASE.md`](RELEASE.md).

## Community

This project is part of the [SLSA Community](https://slsa.dev/community) working
with the SLSA Tooling SIG.

## Communication

The [`#slsa-tooling`](https://slack.com/app_redirect?team=T019QHUBYQ3&channel=slsa-tooling)
channel in the [OpenSSF Slack](https://slack.openssf.org/) is used for
communication and sharing ideas.

Communication about bugs, usage, and new feature development is also done on
[GitHub issues](https://github.com/slsa-framework/slsa-github-generator/issues).

[code of conduct]: https://github.com/slsa-framework/governance/blob/main/8._Code_of_Conduct.md
[conventional commits]: https://www.conventionalcommits.org/en/v1.0.0/
[developer certificate of origin]: https://developercertificate.org/
[license]: ./LICENSE
[good first issue]: https://github.com/slsa-framework/slsa-github-generator/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22
[help wanted]: https://github.com/slsa-framework/slsa-github-generator/issues?q=is%3Aissue+is%3Aopen+label%3A%22status%3Ahelp+wanted%22
[open issues]: https://github.com/slsa-framework/slsa-github-generator/issues
