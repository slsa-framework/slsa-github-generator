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
  - [Releasing to Maven Central](#releasing-to-maven-central)
    - [Action requirements](#action-requirements)
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
      id-token: write
      contents: read
      actions: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_maven_slsa3.yml@v2.0.0
```

Now, when you invoke this workflow, the Maven builder will build both your artifacts and the provenance files for them.

### Releasing to Maven Central

You can also release artifacts to Maven Central with [the slsa-github-generator Maven publish action](https://github.com/slsa-framework/slsa-github-generator/blob/main/actions/maven/publish/README.md).

#### Action requirements

Besides adding the above workflow to your CI pipeline, you also need to add the following plugin to your `pom.xml`:

```xml
<plugin>
    <groupId>io.github.slsa-framework.slsa-github-generator</groupId>
    <artifactId>hash-maven-plugin</artifactId>
    <version>0.0.1</version>
    <executions>
        <execution>
            <goals>
                <goal>hash-jarfile</goal>
            </goals>
        </execution>
    </executions>
    <configuration>
        <outputJsonPath>${SLSA_OUTPUTS_ARTIFACTS_FILE}</outputJsonPath>
    </configuration>
</plugin>
```

### Private Repositories

The builder records all provenance signatures in the [Rekor](https://github.com/sigstore/rekor) public transparency log. This record includes the repository name. To acknowledge you're aware that your repository name will be public, set the flag `rekor-log-public: true` when calling the builder:

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
