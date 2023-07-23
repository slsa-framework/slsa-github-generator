# SLSA GitHub Generator

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator/badge)](https://api.securityscorecards.dev/projects/github.com/slsa-framework/slsa-github-generator)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6503/badge)](https://bestpractices.coreinfrastructure.org/projects/6503)
[![Go Report Card](https://goreportcard.com/badge/github.com/slsa-framework/slsa-github-generator)](https://goreportcard.com/report/github.com/slsa-framework/slsa-github-generator)
[![Slack](https://img.shields.io/static/v1?label=openssf.slack.com&message=%23slsa-tooling&color=4A154B&logo=slack)](https://slack.openssf.org/)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

<img align="right" src="https://slsa.dev/images/logo-mono.svg" width="140" height="140">

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Overview](#overview)
  - [What is SLSA?](#what-is-slsa)
  - [What is provenance?](#what-is-provenance)
  - [What is slsa-github-generator?](#what-is-slsa-github-generator)
  - [Hall of Fame](#hall-of-fame)
    - [Generation of Provenance](#generation-of-provenance)
    - [Builder Creation](#builder-creation)
- [Roadmap](#roadmap)
- [Generation of provenance](#generation-of-provenance)
  - [Referencing SLSA builders and generators](#referencing-slsa-builders-and-generators)
  - [Builders](#builders)
  - [Provenance-only generators](#provenance-only-generators)
- [Verification of provenance](#verification-of-provenance)
  - [Installation](#installation)
  - [Inputs](#inputs)
  - [Command line examples](#command-line-examples)
- [Build Your Own Builder](#build-your-own-builder)
- [Technical design](#technical-design)
  - [Specifications](#specifications)
  - [Provenance format](#provenance-format)
- [Contributing](#contributing)

<!-- tocstop -->

## Overview

### What is SLSA?

[Supply chain Levels for Software Artifacts](https://slsa.dev), or SLSA (salsa),
is a security framework, a check-list of standards and controls to prevent
tampering, improve integrity, and secure packages and infrastructure in your
projects, businesses or enterprises.

SLSA defines an incrementally adoptable set of levels which are defined in
terms of increasing compliance and assurance. SLSA levels are like a common
language to talk about how secure software, supply chains and their component
parts really are.

### What is provenance?

Provenance is information, or metadata, about how a software artifact was
created. This could include information about what source code, build system,
and build steps were used, as well as who and why the build was initiated.
Provenance can be used to determine the authenticity and trustworthiness of
software artifacts that you use.

As part of the framework, SLSA defines a
[provenance format](https://slsa.dev/provenance/) which can be used to hold this
metadata.

### What is slsa-github-generator?

slsa-github-generator is a set of tools for generation of SLSA3+ provenance for
native GitHub projects. It allows projects to generate
[SLSA provenance](https://slsa.dev/provenance/) safely and accurately using
[GitHub Actions](https://github.com/features/actions).

Specifically, this repository contains:

- tools for generating non-forgeable
SLSA provenance on GitHub for your projects. The generated provenance meets the
[provenance generation](https://slsa.dev/spec/v1.0/requirements#provenance-generation)
and [isolation](https://slsa.dev/spec/v1.0/requirements#isolation-strength)
requirements for [SLSA Build level 3 and above](https://slsa.dev/spec/v1.0/levels).
See some [popular projects](#generation-of-provenance) generating provenance using this project.
- tools for building a SLSA builder on GitHub using the [Build-Your-Own-Builder](#build-your-own-builder) framework. With this framework, you can "wrap" an existing GitHub Action into a SLSA builder. The SLSA builder will generate non-forgeable provenance meeting the
[provenance generation](https://slsa.dev/spec/v1.0/requirements#provenance-generation)
and [isolation](https://slsa.dev/spec/v1.0/requirements#isolation-strength)
requirements for [SLSA Build level 3 and above](https://slsa.dev/spec/v1.0/levels).
See some [builders](#builder-creation) created using the BYOB framework.

While slsa-github-generator can help you achieve SLSA Build level 3, use of the provided
[GitHub Actions reusable workflows](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
only is not sufficient to meet all of the requirements at SLSA Build level 3.
Specifically, these workflows do not address provenance
[distribution](https://slsa.dev/spec/v1.0/distributing-provenance) or
[verification](https://slsa.dev/spec/v1.0/verifying-artifacts). Those requirements
must be handled separately to meet SLSA Build level 3+.

### Hall of Fame

#### Generation of Provenance

Below is a non-exhaustive list of projects that use the builders in this repository to generate provenance:

[![flask stars](https://img.shields.io/github/stars/pallets/flask?logo=github&label=pallets/flask)](https://github.com/pallets/flask) [![flatbuffers stars](https://img.shields.io/github/stars/google/flatbuffers?logo=github&label=google/flatbuffers)]((https://github.com/google/flatbuffers)) [![grpc-gateway stars](https://img.shields.io/github/stars/grpc-ecosystem/grpc-gateway?logo=github&label=grpc-ecosystem/grpc-gateway)](https://github.com/grpc-ecosystem/grpc-gateway) [![click stars](https://img.shields.io/github/stars/pallets/click?logo=github&label=pallets/click)](https://github.com/pallets/click) [![argo-cd stars](https://img.shields.io/github/stars/argoproj/argo-cd?logo=github&label=argoproj/argo-cd)](https://github.com/argoproj/argo-cd) [![jib stars](https://img.shields.io/github/stars/GoogleContainerTools/jib?logo=github&label=GoogleContainerTools/jib)](https://github.com/GoogleContainerTools/jib) [![jinja stars](https://img.shields.io/github/stars/pallets/jinja?logo=github&label=pallets/jinja)](https://github.com/pallets/jinja) [![docker-bench-security stars](https://img.shields.io/github/stars/docker/docker-bench-security?logo=github&label=docker/docker-bench-security)](https://github.com/docker/docker-bench-security) [![sentencepiece stars](https://img.shields.io/github/stars/google/sentencepiece?logo=github&label=google/sentencepiece)](https://github.com/google/sentencepiece) [![werkzeug stars](https://img.shields.io/github/stars/pallets/werkzeug?logo=github&label=pallets/werkzeug)](https://github.com/pallets/werkzeug) [![ko stars](https://img.shields.io/github/stars/ko-build/ko?logo=github&label=ko-build/ko)](https://github.com/ko-build/ko) [![micronaut-core stars](https://img.shields.io/github/stars/micronaut-projects/micronaut-core?logo=github&label=micronaut-projects/micronaut-core)](https://github.com/micronaut-projects/micronaut-core) [![kubeedge stars](https://img.shields.io/github/stars/kubeedge/kubeedge?logo=github&label=kubeedge/kubeedge)](https://github.com/kubeedge/kubeedge) [![osv-scanner stars](https://img.shields.io/github/stars/google/osv-scanner?logo=github&label=google/osv-scanner)](https://github.com/google/osv-scanner) [![flux2 stars](https://img.shields.io/github/stars/fluxcd/flux2?logo=github&label=fluxcd/flux2)](https://github.com/fluxcd/flux2) [![kyverno stars](https://img.shields.io/github/stars/kyverno/kyverno?logo=github&label=kyverno/kyverno)](https://github.com/kyverno/kyverno) [![flask-sqlalchemy stars](https://img.shields.io/github/stars/pallets-eco/flask-sqlalchemy?logo=github&label=pallets-eco/flask-sqlalchemy)](https://github.com/pallets-eco/flask-sqlalchemy) [![scorecard stars](https://img.shields.io/github/stars/ossf/scorecard?logo=github&label=ossf/scorecard)](https://github.com/ossf/scorecard) [![urllib3 stars](https://img.shields.io/github/stars/urllib3/urllib3?logo=github&label=urllib3/urllib3)](https://github.com/urllib3/urllib3) [![pdns stars](https://img.shields.io/github/stars/PowerDNS/pdns?logo=github&label=PowerDNS/pdns)](https://github.com/PowerDNS/pdns) [![powertools-lambda-python stars](https://img.shields.io/github/stars/aws-powertools/powertools-lambda-python?logo=github&label=aws-powertools/powertools-lambda-python)](https://github.com/aws-powertools/powertools-lambda-python)

#### Builder Creation

Several builders have been built using the "Build Your Own Builder" (BYOB):

1. [nodejs builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/nodejs#readme), by @ianlewis
2. [JReleaser builder](https://github.com/jreleaser/release-action/tree/java#slsa-builder), by @aalmiray
3. [Maven builder](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/maven/README.md), by @AdamKorcz
4. [Gradle builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/gradle/README.md), by @AdamKorcz
4. Coming soon! [Bazel builder](https://github.com/slsa-framework/slsa-github-generator/tree/main/internal/builders/bazel/README.md), by @enteraga6


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

Build platforms build and generate provenance. They let you meet the
[provenance generation](https://slsa.dev/spec/v1.0/requirements#provenance-generation) and
[isolation strength](https://slsa.dev/spec/v1.0/requirements#isolation-strength)
requirements for [SLSA Build level 3 and above](https://slsa.dev/spec/v1.0/levels).

This repository hosts the following builders:

1. [Go Builder SLSA Level 3](internal/builders/go/README.md). **Status**: [available since v1.0.0](https://github.com/slsa-framework/slsa-github-generator/milestone/1).
   This builder builds and generates provenance for your [Go](https://go.dev/) projects. To use it,
   follow the [Go builder's README.md](internal/builders/go/README.md).
2. [Node.js Builder SLSA Level 3](internal/builders/nodejs/README.md).
   **Status**: [Beta since v1.6.0](https://github.com/slsa-framework/slsa-github-generator/milestone/8).
   [Expected GA release Jun 2023](https://github.com/slsa-framework/slsa-github-generator/milestone/17).
3. _Container-based Builder SLSA Level 3_. **Status**: [Beta release since v1.7.0](https://github.com/slsa-framework/slsa-github-generator/milestone/4). This builder builds arbitrary artifacts by executing a user-supplied container image.
4. [Maven builder](internal/builders/maven/README.md). **Status**: [Beta since v1.8.0](https://github.com/slsa-framework/slsa-github-generator/milestone/14). This builder builds [Maven](https://maven.apache.org/) packages. The package and its attestations can be uploaded to [Maven central](https://search.maven.org).
5. [Gradle builder](internal/builders/gradle/README.md). **Status**: [Beta since v1.8.0](https://github.com/slsa-framework/slsa-github-generator/milestone/15). This builder builds [Gradle](https://gradle.org/) projects. The Maven package and its attestations can be uploaded to Maven central.
5. [Bazel builder](internal/builders/bazel/README.md). **Status**: [WIP](https://github.com/slsa-framework/slsa-github-generator/milestone/16).
   [Expected beta-release Sept 2023]This builder builds [Bazel](https://bazel.build/) projects.
7. _Container Builder SLSA Level 3_. **Status**: [WIP](https://github.com/slsa-framework/slsa-github-generator/milestone/5).
   This builder will build your container image and generate provenance. The generated provenance will be compatible with
   [cosign](https://github.com/sigstore/cosign)'s attestation format.

If you would rather build your project yourself, use the generators instead as explained in the next section.

There are other builders built using this repository's [BYOB framework](#build-your-own-builderld Your Own Builder) and not hosted in this repository:

1. [JReleaser builder](https://github.com/jreleaser/release-action/tree/java#slsa-builder).


### Provenance-only generators

Provenance-only generators let you build your artifact, and only generate provenance for you.
They let you meet the [provenance generation](https://slsa.dev/spec/v1.0/requirements#provenance-generation) requirements
for [SLSA Build level 3](https://slsa.dev/spec/v1.0/levels).

Generators create an attestation to a software artifact coming from your repository.

Generators are _not_ able to report the commands used to generate your artifact in the provenance.

This repository hosts the following generators:

1. [Generic generator SLSA Level 3](internal/builders/generic/README.md). **Status**: [available since v1.2.0](https://github.com/slsa-framework/slsa-github-generator/milestone/2).
   This generator generates provenance for arbitrary artifacts of your choice. To use it,
   follow the [Generic generator's README.md](internal/builders/generic/README.md).
2. [Container generator SLSA Level 3](internal/builders/container/README.md). **Status**: [available since v1.4.0](https://github.com/slsa-framework/slsa-github-generator/milestone/3).
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

## Build Your Own Builder

If you want to build your own builder, use the [BYOB framework](BYOB.md). The framework lets you create your own SLSA3 builder on GitHub. For example, you can wrap an existing GitHub Action intoto a SLSA3 builder. For verification, your users can use the [slsa-verifier](#verification-of-provenance).

## Technical design

The initial technical design was described in the blog post
"[Improving software supply chain security with tamper-proof builds](https://security.googleblog.com/2022/04/improving-software-supply-chain.html)".

### Specifications

For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).

### Provenance format

The format of the provenance is available in [PROVENANCE_FORMAT.md](./PROVENANCE_FORMAT.md).

## Contributing

Please see the [Contributor Guide](CONTRIBUTING.md) for more info.
