# Generation of SLSA3+ provenance for native GitHub projects

This repository contains tools for generating non-forgeable [SLSA provenance](https://slsa.dev/) on GitHub that meets the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements) and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels).

Use of the provided [GitHub Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)s only is not sufficient to meet all of the requirements at SLSA level 3. Specifically, the [source requirements](https://slsa.dev/spec/v0.1/requirements#source-requirements) are not covered by these workflows and must be handled explicitly to meet all requirements at SLSA level 3+.

This repository contains the code, examples and technical design for system described in the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Generation of provenance](#generation-of-provenance)
  - [Builders](#builders)
  - [Provenance-only Generators](#provenance-only-generators)
- [Verification of provenance](#verification-of-provenance)
  - [Installation](#installation)
  - [Inputs](#inputs)
  - [Command line examples](#command-line-examples)
- [Technical design](#technical-design)
  - [Blog post](#blog-post)
  - [Specifications](#specifications)
  - [Provenance format](#provenance-format)

---

## Generation of provenance

### Builders 

Builders build and generate provenance. They let you meet the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements)
and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA Level 3 and above](https://slsa.dev/spec/v0.1/levels).

Builders are able to report the exact commands used to generate your artifact in the provenance.

The following builders are available:

1. [Go Builder SLSA Level 3](internal/builders/go/README.md): To generate SLSA provenance for your [Go](https://go.dev/) project, follow
[internal/builders/go/README.md](internal/builders/go/README.md)


### Provenance-only Generators

Provenance-only generators let you build your artifact, and only generate provenance for you. 
They let you meet the [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements 
for [SLSA Level 3](https://slsa.dev/spec/v0.1/levels).

Generators create an attestation to a software artifact coming from your repository.

Generators are *not* able to report the exact commands used to generate your artifact in the provenance.

To generate SLSA provenance using the provenance-only generator, follow
[internal/builders/generic/README.md](internal/builders/generic/README.md).
This is a pre-release only and we will have the official release in July 2022.

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
