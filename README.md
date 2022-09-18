[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator/badge)](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator)

# Generation of SLSA3+ provenance for native GitHub projects

This repository contains tools for generating non-forgeable [SLSA provenance](https://slsa.dev/) on GitHub that meets the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements) and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels).

Use of the provided [GitHub Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)s only is not sufficient to meet all of the requirements at SLSA level 3. Specifically, the [source requirements](https://slsa.dev/spec/v0.1/requirements#source-requirements) are not covered by these workflows and must be handled explicitly to meet all requirements at SLSA level 3+.

This repository contains the code, examples and technical design for system described in the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Roadmap](#roadmap)
- [Generation of provenance](#generation-of-provenance)
  - [Builders](#builders)
  - [Provenance-only generators](#provenance-only-generators)
- [Verification of provenance](#verification-of-provenance)
  - [Installation](#installation)
  - [Inputs](#inputs)
  - [Command line examples](#command-line-examples)
- [Technical design](#technical-design)
  - [Blog post](#blog-post)
  - [Specifications](#specifications)
  - [Provenance format](#provenance-format)
- [Development](#development)
  - [Unit Tests](#unit-tests)
  - [Linters](#linters)

---

## Roadmap

The project roadmap is tracked via milestones. You can track progress and open
issues via the [milestones page](https://github.com/slsa-framework/slsa-github-generator/milestones?direction=asc&sort=due_date&state=open).
Each milestone includes a description of what is being worked on and a rough
timeline for completion.

## Generation of provenance

Below we describe the various builders and generators in this repository. They let you build and / or generate non-forgeable provenance
using a trusted / isolated re-usable workflow. You can read up on the design in our [technical design document](#technical-design).

**Note**: At present the GitHub Actions provided in this repository as builders and generators **MUST** be referenced by
a tag that correpsonds to a semantic version of the form `@vX.Y.Z`. The build will fail
if you reference it via a shorter tag like `@vX.Y` or `@vX` or if you reference it by a tag of a different form (e.g., `@main`).

### Builders

Builders build and generate provenance. They let you meet the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements)
and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA Level 3 and above](https://slsa.dev/spec/v0.1/levels).

Builders are able to report the commands used to generate your artifact in the provenance.

This repository hosts the following builders:

1. [Go Builder SLSA Level 3](internal/builders/go/README.md). **Status**: available since v1.0.0.
   This builder builds and generates provenance for your [Go](https://go.dev/) projects. To use it,
   follow the [Go builder's README.md](internal/builders/go/README.md).
1. [Container Builder SLSA Level 3](TODO). **Status**: WIP, expected release in Sept 2022.
   This builder will build your container image and generate provenance. The generated provenance will be compatible with
   [cosign](https://github.com/sigstore/cosign)'s attestation format.
1. [Dockerfile-based Builder SLSA Level 3](TODO). **Status**: WIP, see [#23](https://github.com/slsa-framework/slsa-github-generator/issues/23).
   This builder will build arbitrary artifacts using building steps defined in a Dockerfile.

If you would rather build your project yourself, use the generators instead as explained in the next section.

### Provenance-only generators

Provenance-only generators let you build your artifact, and only generate provenance for you.
They let you meet the [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements
for [SLSA Level 3](https://slsa.dev/spec/v0.1/levels).

Generators create an attestation to a software artifact coming from your repository.

Generators are _not_ able to report the commands used to generate your artifact in the provenance.

This repository hosts the following generators:

1. [Generic generator SLSA Level 3](internal/builders/generic/README.md). **Status**: available since v1.2.0.
   This generator generates provenance for arbitrary artifacts of your choice. To use it,
   follow the [Generic generator's README.md](internal/builders/generic/README.md).
1. [Container generator SLSA Level 3](TODO). **Status**: WIP, expected release Aug-Sept 2022, see [#409](https://github.com/slsa-framework/slsa-github-generator/issues/409).
   This generator will generate provenance for container images. The generated provenance will be compatible with
   [cosign](https://github.com/sigstore/cosign)'s attestation format.

## Verification of provenance

To verify the provenance, use the [github.com/slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier) project.

**Note**: At present the GitHub Actions provided in this repository as builders and generators **MUST** be referenced by tag in order for the `slsa-verifier` to be able to verify the ref of the trusted builder/generator's reusable workflow.

This is contrary to the [best practice](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions) which recommends referencing by digest, but intentional due to limits in GitHub Actions.
The desire to be able to verify reusable workflows pinned by hash, and the reasons for the current status, are tracked as [Issue #12](https://github.com/slsa-framework/slsa-verifier/issues/12) in the slsa-verifier project.

### Installation

To install the verifier, see [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation).

### Inputs

The inputs of the verifier are described in [slsa-framework/slsa-verifier#available-options](https://github.com/slsa-framework/slsa-verifier#available-options).

### Command line examples

A command line example is provided in [slsa-framework/slsa-verifier#example](https://github.com/slsa-framework/slsa-verifier#example).

## Technical design

### Blog post

Find our blog post series [here](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

### Specifications

For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).

### Provenance format

The format of the provenance is available in [PROVENANCE_FORMAT.md](./PROVENANCE_FORMAT.md).

## Development

Since this project includes reusable workflows for use on GitHub Actions local
development is limited to building and testing the binaries used by the reusable
workflows. The workflows themselves must be tested in your own fork.

Local commands that can be used for development are defined in the
[Makefile](./Makefile). You can list the available targets by running `make`.

```
make
```

### Unit Tests

You can run unit tests locally using `make`. This requires that the Go runtime
be installed.

```
make unit-test
```

### Linters

This project uses several linters in order to maintain code quality. If you wish
to run these linters locally, follow the instructions for each of these to
install them on your development machine.

- [yamllint](https://yamllint.readthedocs.io/)
- [golangci-lint](https://golangci-lint.run/)
- [shellcheck](https://www.shellcheck.net/)
- [eslint](https://eslint.org/) (NOTE: eslint is installed automatically so you
  don't need to install it)

Once each of these are installed you can run the linters using `make`.

```
make lint
```

These linters will also run as GitHub checks for pull requests.
