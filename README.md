# Generation of SLSA3+ provenance for native GitHub projects

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator/badge)](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6503/badge)](https://bestpractices.coreinfrastructure.org/projects/6503)
[![Go Report Card](https://goreportcard.com/badge/github.com/slsa-framework/slsa-github-generator)](https://goreportcard.com/report/github.com/slsa-framework/slsa-github-generator)
[![Slack](https://slack.babeljs.io/badge.svg)](https://slack.com/app_redirect?team=T019QHUBYQ3&channel=slsa-tooling)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

This repository contains tools for generating non-forgeable [SLSA provenance](https://slsa.dev/) on GitHub that meets the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements) and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels).

Use of the provided [GitHub Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)s only is not sufficient to meet all of the requirements at SLSA level 3. Specifically, the [source requirements](https://slsa.dev/spec/v0.1/requirements#source-requirements) are not covered by these workflows and must be handled explicitly to meet all requirements at SLSA level 3+.

This repository contains the code, examples and technical design for system described in the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Roadmap](#roadmap)
- [Generation of provenance](#generation-of-provenance)
  - [Referencing SLSA builders and generators](#referencing-slsa-builders-and-generators)
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
- [Contributing](#contributing)

<!-- tocstop -->

---

## Roadmap

The project roadmap is tracked via milestones. You can track progress and open
issues via the [milestones page](https://github.com/slsa-framework/slsa-github-generator/milestones?direction=asc&sort=due_date&state=open).
Each milestone includes a description of what is being worked on and a rough
timeline for completion.

## Generation of provenance

Below we describe the various builders and generators in this repository. They let you build and / or generate non-forgeable provenance
using a trusted / isolated re-usable workflow. You can read up on the design in our [technical design document](#technical-design).

### Referencing SLSA builders and generators

At present, the GitHub Actions provided in this repository as builders and generators **MUST** be referenced
by tag in order for the `slsa-verifier` to be able to verify the ref of the trusted builder/generator's
reusable workflow. It also needs to be referred as `@vX.Y.Z`, because the build will fail if you reference it via a shorter tag like `@vX.Y` or `@vX`.

This is contrary to the [GitHub best practice for third-party actions](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions) which recommends referencing by digest, but intentional due to limits in GitHub Actions.
The desire to be able to verify reusable workflows pinned by hash, and the reasons for the current status, are tracked as [Issue #12](https://github.com/slsa-framework/slsa-verifier/issues/12) in the slsa-verifier project.

For guidance on how to configure renovate see [RENOVATE.md](RENOVATE.md).

### Builders

Builders build and generate provenance. They let you meet the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements)
and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA Level 3 and above](https://slsa.dev/spec/v0.1/levels).

Builders are able to report the commands used to generate your artifact in the provenance.

This repository hosts the following builders:

1. [Go Builder SLSA Level 3](internal/builders/go/README.md). **Status**: [available since v1.0.0](https://github.com/slsa-framework/slsa-github-generator/milestone/1).
   This builder builds and generates provenance for your [Go](https://go.dev/) projects. To use it,
   follow the [Go builder's README.md](internal/builders/go/README.md).
1. _Container Builder SLSA Level 3_. **Status**: [WIP, expected release in Nov 2022](https://github.com/slsa-framework/slsa-github-generator/milestone/5).
   This builder will build your container image and generate provenance. The generated provenance will be compatible with
   [cosign](https://github.com/sigstore/cosign)'s attestation format.
1. _Dockerfile-based Builder SLSA Level 3_. **Status**: [WIP](https://github.com/slsa-framework/slsa-github-generator/milestone/4). This builder will build arbitrary
   artifacts using building steps defined in a Dockerfile.

If you would rather build your project yourself, use the generators instead as explained in the next section.

### Provenance-only generators

Provenance-only generators let you build your artifact, and only generate provenance for you.
They let you meet the [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements
for [SLSA Level 3](https://slsa.dev/spec/v0.1/levels).

Generators create an attestation to a software artifact coming from your repository.

Generators are _not_ able to report the commands used to generate your artifact in the provenance.

This repository hosts the following generators:

1. [Generic generator SLSA Level 3](internal/builders/generic/README.md). **Status**: [available since v1.2.0](https://github.com/slsa-framework/slsa-github-generator/milestone/2).
   This generator generates provenance for arbitrary artifacts of your choice. To use it,
   follow the [Generic generator's README.md](internal/builders/generic/README.md).
1. [Container generator SLSA Level 3](internal/builders/container/README.md). **Status**: [WIP, expected release Oct 2022](https://github.com/slsa-framework/slsa-github-generator/milestone/3).
   This generator will generate provenance for container images. The generated provenance will be compatible with
   [cosign](https://github.com/sigstore/cosign)'s attestation format.

## Verification of provenance

To verify the provenance, use the [github.com/slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier) project.

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

## Contributing

Please see the [Contributor Guide](CONTRIBUTING.md) for more info.
