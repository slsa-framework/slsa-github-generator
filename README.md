# Generation of SLSA3+ provenance for native GitHub projects

This repository contains tools for generating non-forgeable [SLSA provenance](https://slsa.dev/) on GitHub that meets the [build](https://slsa.dev/spec/v0.1/requirements#build-requirements) and [provenance](https://slsa.dev/spec/v0.1/requirements#provenance-requirements) requirements for [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels).

Use of the provided [Github Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)s only is not sufficient to meet all of the requirements at SLSA level 3. Specifically, the [source requirements](https://slsa.dev/spec/v0.1/requirements#source-requirements) are not covered by these workflows and must be handled explicitly to meet all requirements at SLSA level 3+.

This repository contains the code, examples and technical design for system described in the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Generation of provenance](#generation-of-provenance)
  - [Go projects](#go-projects)
  - [Other projects](#other-projects)
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

### Go projects

To generate SLSA provenance for your [Go](https://go.dev/) project, follow
[internal/builders/go/README.md](internal/builders/go/README.md).

### Other projects

To generate SLSA provenance for other programming languages, follow
[internal/builders/generic/README.md](internal/builders/generic/README.md).
This is a pre-release only and we will have the official release in July 2022.

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
