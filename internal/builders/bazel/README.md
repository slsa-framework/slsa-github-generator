# Generation of SLSA3+ provenance for Bazel builds

This document explains how to generate SLSA provenance for Bazel artifact(s).

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the artifact(s) and generate SLSA provenance. We'll call this
workflow the "Bazel builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Runfile Support](#runfile-support)
  - [Java Artifact Support (and Caveats)](#java-artifact-support-and-caveats)
  - [Referencing the Bazel builder](#referencing-the-bazel-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Bazel builder will generate a non-forgeable attestation to the
Bazel build artifacts produced using the identity of the GitHub workflow. This can be used to
create a positive attestation to an artifact coming from your repository.

That means that once your users verify the artifact they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Bazel builder is currently in development.

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Generating Provenance

The Bazel builder uses a Github Actions reusable workflow to build your
artifact(s) and generate the provenance.

### Getting Started

Let's assume you have a Bazel project which includes a
[WORKSPACE file](https://bazel.build/reference/glossary#workspace-file)
at the root of your repository as well as one or more
[BUILD files](https://bazel.build/reference/glossary#build-file). Also included
at the root of the directory is the
[.bazelversion file](https://github.com/bazelbuild/bazelisk#how-does-bazelisk-know-which-bazel-version-to-run)
, [.bazelrc file](https://bazel.build/run/bazelrc)
, and a
[MODULE.bazel file](https://bazel.build/rules/lib/globals/module).

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.
This assumes that the files above are in the root directory of your
repository.

We also define the targets to be built as 'fib' and 'hello', and
we define the flag of the build to be '--strip=always'

The following reusable workflow call will build the targets and upload the artifacts into a .zip and
generate provenance attestations which will be uploaded as another .zip to the
workflow run.

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@v2.0.0
    with:
      targets: "//src:fib //src:hello"
      flags: "--strip=always"
```

The `targets` are a set of space separated build targets to be built. Each target must include the `//` workspace root identifier and package target identifier (`:your_target`). Because of this each target should be of the form `//path/from/root/to/target:your_target`.

Targets can also be referred to with general glob patterns such as `//src/...` or `//src/internal:all`. Generic glob patterns that have an intersection are allowed as well.

Once the targets are built, the Bazel builder creates a folder for the artifacts
and another for the provenance attestations which are uploaded as artifacts to the workflow run.

### Runfile Support

If the artifact(s) built need the runfiles generated along with it to function properly, then they can be added with the artifact in the attestation. In the following resuable workflow call, the flag `needs-runfiles` will be set to `true`
in order to package the artifacts with their runfiles.

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@v2.0.0
    with:
      targets: "//src:fib //src:hello"
      flags: "--strip=always"
      needs-runfiles: true
```

In the artifact folder that gets uploaded to Github, with `needs-runfiles` set to true, there will be a folder for each artifact which contains the artifact and the folder of its runfiles.
With the `needs-runfiles` flag set to true, each target specified in the workflow call will be packaged with their respective runfiles.

### Java Artifact Support (and Caveats)

If the targets being built includes Java targets, then the flag `includes-java` must be set to true. Additionally, if a specific distribution and version of Java is needed,
that can be designated through the `user-java-distribution` and `user-java-version` flags. Note that the default Java distribution is Oracle and default Java version is 17.
For more info on configuring the Java distribution and version go [here](https://github.com/actions/setup-java). This flag usage can be seen in the following resuable workflow call:

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@v2.0.0
    with:
      targets: "//src:fib //src:hello"
      flags: "--strip=always"
      includes-java: true
      user-java-distribution: "oracle"
      user-java-version: "17"
```

Each Java target will be outputed in its own directory inside the artifact folder that gets uploaded. Inside each respective artifact directory will be a JAR that can be ran on its own using the run-script that is
packaged with it. For instance if there is a Java target named Main it would be uploaded as its own directory with tree looking like the following:

├── Main <br />
│   ├── Main # This is the run-script <br />
│   └── Main_deploy.jar <br />

Each Java target, whether specified as in the targets input as a `_deploy.jar` or not, will be built as a [_deploy.jar](https://bazel.build/reference/be/java) which contains all classes found by classloader and native libraries for dependencies.
Since the artifact is built on a Github Runner, the run-script has the VM's Java bin path hardcoded in. However, the run-script has been modified to include an additional flag, `--local_javabin` to change the Java Bin path to the user's. To run the JAR using
the run-script the `--singlejar` flag must be specified to signal to the run-script that the JAR is a `_deploy.jar`. Additionally, `--local_javabin` must be set to the path of the user's Java Bin to run it. Therefore running the JAR would look like the following:

`./Main --singlejar --local_javabin="path/to/user/bin/java"`

Note that Java targets do not need to have the `needs-runfiles` flag to be true in order to create the _deploy.jar and run-script for it.

### Referencing the Bazel builder

At present, the builder **MUST** be referenced by a tag of the form `@vX.Y.Z`,
because the build will fail if you reference it via a shorter tag like `@vX.Y`
or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure
renovatebot, see the main repository [README.md](../../../README.md).

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

### Supported Triggers

Only the following [event types] are supported:

| Supported event type  | Event description                          |
| --------------------- | ------------------------------------------ |
| [`create`]            | Creation of a git tag or branch.           |
| [`release`]           | Creation or update of a GitHub release.    |
| [`push`]              | Creation or update of a git tag or branch. |
| [`workflow_dispatch`] | Manual trigger of a workflow.              |

`pull_request` events are currently not supported. If you would like support for
`pull_request`, please tell us about your use case on
[issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a
[new issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

[event types]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
[`create`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#create
[`release`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#release
[`push`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#push
[`workflow_dispatch`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch

### Workflow Inputs

The Bazel builder accepts the following inputs:

Inputs:

| Name              | Required | Default            | Description                                                                                                                                                                                                                                         |
| ----------------- | -------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| rekor-log-public  | No       | false              | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false` |
| targets           | Yes      |                    | A space separated list of targets to build and generate artifacts for. See [targets](https://bazel.build/concepts/build-ref#targets) for more information. \                                                                                        |
| flags             | No       | ""                 | A space separated list of flags to modify the build by. See [flags](https://bazel.build/docs/user-manual#build-options) for more information. \                                                                                                     |

### Workflow Outputs

The Bazel builder produces the following outputs:

| Name                       | Description                                                            |
| -------------------------- | ---------------------------------------------------------------------- |
| provenance-download-name   | The name of the provenance attestation uploaded to the workflow run.   |

### Provenance Format

Provenance is generated as an [in-toto](https://in-toto.io/) statement with a
SLSA v1.0 predicate.

| Name           | Value                                                          | Description                                                                                    |
| -------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `subject.name` | Artifact Name | The subject identifies the artifact built by the builder. The provenance attests the build process for the artifact listed.                     |

The project generates SLSA v1.0 provenance predicate with the following values.

| Name                         | Value                                                                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `runDetails.builder.id`      | `https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@refs/tags/v1.7.0"`  | Identifies the Bazel builder                                                                                                                                                                                           |
| `buildDefinition.buildType`  | `"https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0"`                                       | Identifies a the GitHub Actions build.                                                                                                                                                                                 |
| `runDetails.metadata.invocationID` | `"[run_id]-[run_attempt]"`                                                                                       | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |
| `externalParameters.inputs   | User Inputted Values                                                                                                   | Identifies the inputs to the Bazel Builder that were passed in                                                                                                                                                         |

### Provenance Example

The following is an example of the generated provenance.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "subject": [
    {
      "name": "fib",
      "digest": {
        "sha256": "caaadba2846905ac477c777e96a636e1c2e067fdf6fed90ec9eeca4df18d6ed9"
      }
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v1",
  "predicate": {
    "buildDefinition": {
      "buildType": "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0",
      "externalParameters": {
        "inputs": {
          "rekor-log-public": false,
          "targets": "//src:fib //src:hello",
          "flags": "--strip=always"
        },
        "vars": {}
      },
      "internalParameters": {
        "GITHUB_ACTOR_ID": "78953604",
        "GITHUB_EVENT_NAME": "workflow_dispatch",
        "GITHUB_REF": "refs/heads/main",
        "GITHUB_REF_TYPE": "branch",
        "GITHUB_REPOSITORY": "enteraga6/cc_user_test",
        "GITHUB_REPOSITORY_ID": "647514313",
        "GITHUB_REPOSITORY_OWNER_ID": "78953604",
        "GITHUB_RUN_ATTEMPT": "1",
        "GITHUB_RUN_ID": "5214546713",
        "GITHUB_RUN_NUMBER": "1",
        "GITHUB_SHA": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6",
        "GITHUB_TRIGGERING_ACTOR_ID": "78953604",
        "GITHUB_WORKFLOW_REF": "enteraga6/cc_user_test/.github/workflows/flag_test.yml@refs/heads/main",
        "GITHUB_WORKFLOW_SHA": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6",
        "GITHUB_EVENT_PAYLOAD": {
          ...
      },
      "resolvedDependencies": [
        {
          "uri": "git+https://github.com/enteraga6/cc_user_test@refs/heads/main",
          "digest": {
            "gitCommit": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6"
          }
        }
      ]
    },
    "runDetails": {
      "builder": {
        "id": "https://github.com/enteraga6/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@refs/heads/bazel-builder"
      },
      "metadata": {
        "invocationId": "https://github.com/enteraga6/cc_user_test/actions/runs/5214546713/attempts/1"
      }
    }
  }
}
```
