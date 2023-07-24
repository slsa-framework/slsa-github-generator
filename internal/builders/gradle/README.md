# Generation of SLSA3+ provenance for `.jar` artifacts built with Gradle

This document explains how to generate SLSA provenance for artifacts built with Gradle.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Gradle builder" from now on.

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

Using the Gradle builder will generate a non-forgeable attestation to the
Gradle package using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the package they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Gradle builder is currently in alpha. The API could change while approaching
a Generally Available (GA) release. You can track progress towards General
Availability via
[this milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/17).

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Limitations

The Gradle builder currently has the following limitations:

1. The project must be build'able by way of `gradle build`. If you need the option for flags, profiles or something else to define more granular builds, please open an issue.

## Generating Provenance

The Gradle builder uses a Github Actions reusable workflow to build your
package and generate the provenance.

### Getting Started

Let's say you have the following build set up:

1. You can build your artifacts by way of `gradle build`.
2. You release artifacts via Github Actions.

To add provenance to releases is easy. Simply use the following workflow in `.github/workflows` in your repository:

```yaml
name: SLSA Provenance with Gradle builder
on:
  - workflow_dispatch

permissions: read-all

jobs:
  usetrw:
    permissions:
      contents: read
      id-token: write
      actions: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_gradle_slsa3.yml@main
    with:
      rekor-log-public: true
      artifact-list: ./artifact1.jar,./artifact2.jar
```

Now, when you invoke this workflow, the Gradle builder will build both your artifacts and the provenance files for them.

The Gradle builder requires you to specify the artifacts that you wish to attest to. To do so, you add a comma-separated list of paths to the artifacts as shown in the example. The paths are relative from the root of your project directory.

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
