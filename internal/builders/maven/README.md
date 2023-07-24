# Generation of SLSA3+ provenance for `.jar` artifacts built with Maven

This document explains how to generate SLSA provenance for `.jar` artifacts built with Maven.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Maven builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Limitations](#limitations)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Private Repositories](#private-repositories)
- [Verification](#verification)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Maven builder will generate a non-forgeable attestation to the
Maven package using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the package they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Maven builder is currently in alpha. The API could change while approaching
a Generally Available (GA) release. You can track progress towards General
Availability via
[this milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/17).

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Limitations

The Maven builder currently has the following limitations:

1. The project must be build'able by way of `mvn package`. If you need the option for flags, profiles or something else to define more granular builds, please open an issue.
2. The Maven publisher is limited to projects that output artifacts in a `target` directory - which is the default way used by the vast majority of projects.

## Generating Provenance

The Maven builder uses a Github Actions reusable workflow to build your
package and generate the provenance.

### Getting Started

Let's say you have the following build set up:

1. You can build your artifacts by way of `mvn package`.
2. You release artifacts via Github Actions.

To add provenance to releases is easy. Simply use the following workflow in `.github/workflows` in your repository:

```yaml
name: Build with provenance
on:
  - workflow_dispatch

permissions: read-all

jobs:
  build:
    permissions:
      contents: write
      id-token: write
      actions: read
      packages: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_maven_slsa3.yml@main
    with:
      rekor-log-public: true
```

Now, when you invoke this workflow, the Maven builder will build both your artifacts and the provenance files for them.

You can also release artifacts to Maven Central by adding the following step to your workflow:

```yaml
  publish:
    needs: build
    permissions:
      contents: write
      id-token: write
      actions: read
      packages: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/publish_maven.yml@main
    with:
      provenance-download-name: "${{ needs.build.outputs.provenance-download-name }}"
      provenance-download-sha256: "${{ needs.build.outputs.provenance-download-sha256 }}"
      target-download-sha256: "${{ needs.build.outputs.target-download-sha256 }}"
    secrets:
      maven-username: ${{ secrets.OSSRH_USERNAME }}
      maven-password: ${{ secrets.OSSRH_PASSWORD }}
      gpg-key-pass: ${{ secrets.GPG_PASSPHRASE }}
      gpg-private-key: ${{ secrets.GPG_PRIVATE_KEY }}
```

Now your workflow will build your artifacts and publish them to a staging repository in Maven Central.

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
https://rekor.sigstore.dev/. This entry includes the repository name. This will cause the
private repository name to leak and be discoverable via the public Rekor API
server.

If this is ok with you, you can set the `rekor-log-public` flag in order to
opt in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  rekor-log-public: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue #372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

## Verification

The `slsa-verifier` tool can verify `.jar` artifacts against their provenance statements.

Please see the [documentation](https://github.com/slsa-framework/slsa-verifier)
for more information.
