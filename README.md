# Generation of SLSA3+ provenance for native GitHub projects

This repository contains the implementation for generating non-forgeable [SLSA provenance](https://slsa.dev/) that meets the requirement for the [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels). The SLSA generation makes use of a [Github Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows).

This repository contains the code, examples and technical design for the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Generation of provenance](#generation-of-provenance)
  - [Golang projects](#golang-projects)
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

### Golang projects

To generate SLSA provenance for your Golang project, follow [internal/builders/go/README.md](internal/builders/go/README.md).

### Other projects

To generate SLSA provenance for other programming languages, follow [internal/builders/generic/README.md](internal/builders/generic/README.md).
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

