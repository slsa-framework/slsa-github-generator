# Generation of SLSA3+ provenance for native GitHub projects

This repository contains the implementation for generating non-forgeable [SLSA provenance](https://slsa.dev/) that meets the requirement for the [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels). The SLSA generation makes use of a [Github Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows).

This repository contains the code, examples and technical design for the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Golang projects](#golang-projects)
- [Other projects](#other-projects)
- [Technical design](#technical-design)
  - [Blog post](#blog-post)
  - [Specifications](#specifications)

---

## Golang projects

To generate SLSA provenance for your Golang project, follow [internal/builders/go/README.md](internal/builders/go/README.md)

## Other projects

To generate SLSA provenance for other programming languages, follow [internal/builders/generic/README.md](internal/builders/generic/README.md)

## Technical design

### Blog post

Find our blog post series [here](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

### Specifications

For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).
