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
    - [Multi-Project Builds](#multi-project-builds)
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
[this milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/15).

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Limitations

The Gradle builder currently has the following limitations:

1. The project must be buildable by way of `./gradlew build`. If you need the option for flags, profiles or something else to define more granular builds, please open an issue.
2. The project must include a gradle wrapper (`gradlew`). The Gradle builder does not include an installation of gradle.
3. The project's build scripts must place the artifacts into `./build`, relative to the `directory` workflow input. If you are doing [multi-project builds](https://docs.gradle.org/current/userguide/intro_multi_project_builds.html), you may need to follow the [example below](#multi-project-builds)

## Generating Provenance

The Gradle builder uses a Github Actions reusable workflow to build your
package and generate the provenance.

### Getting Started

Let's say you have the following build set up:

1. You can build your artifacts by way of `./gradlew build`.
2. You release artifacts via Github Actions.

To add provenance to releases is easy. Simply use the following workflow in `.github/workflows` in your repository:

```yaml
name: SLSA Provenance with Gradle builder
on:
  - workflow_dispatch

permissions: read-all

jobs:
  build:
    permissions:
      id-token: write
      contents: read
      actions: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_gradle_slsa3.yml@v2.0.0
    with:
      artifact-list: >-
        ./build/artifact1.jar,
        ./build/artifact2.jar
```

Now, when you invoke this workflow, the Gradle builder will build both your artifacts and the provenance files for them.

The Gradle builder requires you to specify the artifacts that you wish to attest to. To do so, you add a comma-separated list of paths to the artifacts as shown in the example. The paths are relative from the root of your project directory.

#### Multi-Project Builds

If you are using [multi-project builds](https://docs.gradle.org/current/userguide/intro_multi_project_builds.html), where each of your sub-projects' `src` are in separate subfolders, then you will need to add a task to copy over the artifact files to the root `./build` folder.

See this example to add to your sub-projects' `build.gradle.kts` file.

```kotlin
tasks.register<Copy>("copySubProjectBuild") {
    from(layout.buildDirectory)
    into("${rootProject.projectDir}/build/${project.name}")
}

tasks.named("build") {
    finalizedBy("copySubProjectBuild")
}
```

This, for example, will move `./app1/build/` and `./app2/build/` to `./build/app1/` and `./build/app2/`. You must then alter your input to `artifact-list`.

```yaml
...
      artifact-list: >-
        ./build/app1/libs/app.jar,
        ./build/app2/libs/app.jar,
...
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
