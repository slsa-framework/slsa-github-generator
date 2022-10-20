# Generation of SLSA3+ provenance for arbitrary projects

This document explains how to generate SLSA provenance for projects for which
there is no language or ecosystem specific builder available.

This can be done by adding an additional step to your existing Github Actions
workflow to call a [reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to generate generic SLSA provenance. We'll call this workflow the "generic
workflow" from now on.

The generic workflow differs from ecosystem specific builders (like the [Go
builder](../go)) which build the artifacts as well as generate provenance. This
project simply generates provenance as a separate step in an existing workflow.

---

- [Benefits of Provenance](#benefits-of-provenance)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the SLSA generator](#referencing-the-slsa-generator)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Integration With Other Build Systems](#integration-with-other-build-systems)
  - [Provenance for GoReleaser](#provenance-for-goreleaser)
  - [Provenance for Bazel](#provenance-for-bazel)
  - [Provenance for Java](#provenance-for-java)
  - [Provenance for Rust](#provenance-for-rust)
  - [Provenance for Haskell](#provenance-for-haskell)

---

## Benefits of Provenance

Using the generic workflow will generate a non-forgeable attestation to the
artifacts' digests using the identity of the GitHub workflow. This can be used
to create a positive attestation to a software artifact coming from your
repository.

That means that once your users verify the artifacts they have downloaded they
can be sure that the artifacts were created by your repository's workflow and
haven't been tampered with.

## Generating Provenance

The generic workflow uses a GitHub Actions reusable workflow to generate the
provenance.

### Getting Started

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing GitHub Actions workflow to build your project.

Add a step to your workflow after you have built your project to generate a
sha256 hash of your artifacts and base64 encode it.

Assuming you have a binary called `binary-linux-amd64` you can use the
`sha256sum` and `base64` commands to create the digest. Here we use the `-w0` to
output the encoded data on one line and make it easier to use as a GitHub Actions
output:

```shell
$ sha256sum artifact1 artifact2 ... | base64 -w0
```

This workflow expects the `base64-subjects` input to decode to a string conforming to the expected output of the `sha256sum` command. Specifically, the decoded output is expected to be comprised of a hash value followed by a space followed by the artifact name.

After you have encoded your digest, add a new job to call the reusable workflow.

```yaml
provenance:
  permissions:
    actions: read # Needed for detection of GitHub Actions environment.
    id-token: write # Needed for provenance signing and ID
    contents: write # Needed for release uploads
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
```

**Note**: Make sure that you reference the generator with a semantic version of the form `@vX.Y.Z`. 
More information [here](/README.md#referencing-slsa-builders-and-generators).

Here's an example of what it might look like all together.

```yaml
jobs:
  # This step builds our artifacts, uploads them to the workflow run, and
  # outputs their digest.
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
    runs-on: ubuntu-latest
    steps:
      - name: Build artifacts
        run: |
          # These are some amazing artifacts.
          echo "foo" > artifact1
          echo "bar" > artifact2

      - name: Generate hashes
        shell: bash
        id: hash
        run: |
          # sha256sum generates sha256 hash for all artifacts.
          # base64 -w0 encodes to base64 and outputs on a single line.
          # sha256sum artifact1 artifact2 ... | base64 -w0
          echo "hashes=$(sha256sum artifact1 artifact2 | base64 -w0)" >> "$GITHUB_OUTPUT"

      - name: Upload artifact1
        uses: actions/upload-artifact@3cea5372237819ed00197afe530f5a7ea3e805c8 # tag=v3.1.0
        with:
          name: artifact1
          path: artifact1
          if-no-files-found: error
          retention-days: 5

      - name: Upload artifact2
        uses: actions/upload-artifact@3cea5372237819ed00197afe530f5a7ea3e805c8 # tag=v3.1.0
        with:
          name: artifact2
          path: artifact2
          if-no-files-found: error
          retention-days: 5

  # This step calls the generic workflow to generate provenance.
  provenance:
    needs: [build]
    permissions:
      actions: read
      id-token: write
      contents: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      # Upload provenance to a new release
      upload-assets: true

  # This step uploads our artifacts to the tagged GitHub release.
  release:
    needs: [build, provenance]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/')
    steps:
      - name: Download artifact1
        uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741 # tag=v2.1.0
        with:
          name: artifact1

      - name: Download artifact2
        uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741 # tag=v2.1.0
        with:
          name: artifact2

      - name: Upload assets
        uses: softprops/action-gh-release@1e07f4398721186383de40550babbdf2b84acfc5 # v0.1.14
        with:
          files: |
            artifact1
            artifact2
```

### Referencing the SLSA generator

At present, the generator **MUST** be referenced
by tag of the form `@vX.Y.Z`, because the build will fail if you reference it via a shorter tag like `@vX.Y` or `@vX` or if you reference it by a hash.

For more information about the design decision and how to configure renovatebot,see the main repository [README.md](README.md).

### Supported Triggers

The following [GitHub trigger events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows) are fully supported and tested:

- `schedule`
- `push` (including new tags)
- `release`
- Manual run via `workflow_dispatch`

However, in practice, most triggers should work with the exception of
`pull_request`. If you would like support for `pull_request`, please tell us
about your use case on [issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue with any other triggers please submit a [new
issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The [generic workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_generic_slsa3.yml) accepts the following inputs:

| Name               | Required | Default                                                                                         | Description                                                                                                                                                                                                                                                      |
| ------------------ | -------- | ----------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `base64-subjects`  | yes      |                                                                                                 | Artifact(s) for which to generate provenance, formatted the same as the output of sha256sum (SHA256 NAME\n[...]) and base64 encoded. The encoded value should decode to, for example: `90f3f7d6c862883ab9d856563a81ea6466eb1123b55bff11198b4ed0030cac86 foo.zip` |
| `upload-assets`    | no       | false                                                                                           | If true provenance is uploaded to a GitHub release for new tags.                                                                                                                                                                                                 |
| `provenance-name`  | no       | "(subject name).intoto.jsonl" if a single subject. "multiple.intoto.json" if multiple subjects. | The artifact name of the signed provenance. The file must have the `intoto.jsonl` extension.                                                                                                                                                                     |
| `attestation-name` | no       | "(subject name).intoto.jsonl" if a single subject. "multiple.intoto.json" if multiple subjects. | The artifact name of the signed provenance. The file must have the `intoto.jsonl` extension. DEPRECATED: use `provenance-name` instead.                                                                                                                          |

### Workflow Outputs

The [generic workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_generic_slsa3.yml) produces the following outputs:

| Name               | Description                                                                            |
| ------------------ | -------------------------------------------------------------------------------------- |
| `provenance-name`  | The artifact name of the signed provenance.                                            |
| `attestation-name` | The artifact name of the signed provenance. DEPRECATED: use `provenance-name` instead. |

### Provenance Format

The project generates SLSA provenance with the following values.

| Name                         | Value                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/generic@v1"` | Identifies a generic GitHub Actions build.                                                                                                                                                                             |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                             | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

### Provenance Example

The following is an example of the generated provenance. Provenance is
generated as an [in-toto](https://in-toto.io/) statement with a SLSA predicate.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "ghcr.io/ianlewis/actions-test",
      "digest": {
        "sha256": "8ae83e5b11e4cc8257f5f4d1023081ba1c72e8e60e8ed6cacd0d53a4ca2d142b"
      }
    },
  ],
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.2.0"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator/generic@v1",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/heads/main.git",
        "digest": {
          "sha1": "e491e4b2ce5bc76fb103729b61b04d3c46d8a192"
        },
        "entryPoint": ".github/workflows/generic-container.yml"
      },
      "parameters": {},
      "environment": {
        "github_actor": "ianlewis",
        "github_actor_id": "49289",
        "github_base_ref": "",
        "github_event_name": "push",
        "github_event_payload": {...},
        "github_head_ref": "",
        "github_ref": "refs/tags/v0.0.9",
        "github_ref_type": "tag",
        "github_repository_id": "474793590",
        "github_repository_owner": "ianlewis",
        "github_repository_owner_id": "49289",
        "github_run_attempt": "1",
        "github_run_id": "2556669934",
        "github_run_number": "12",
        "github_sha1": "e491e4b2ce5bc76fb103729b61b04d3c46d8a192"
      }
    },
    "metadata": {
      "buildInvocationID": "2556669934-1",
      "completeness": {
        "parameters": true,
        "environment": false,
        "materials": false
      },
      "reproducible": false
    },
    "materials": [
      {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/tags/v0.0.9",
        "digest": {
          "sha1": "e491e4b2ce5bc76fb103729b61b04d3c46d8a192"
        }
      }
    ]
  }
}
```

## Integration With Other Build Systems

This section explains how to generate non-forgeable SLSA provenance with existing build systems.

### Provenance for GoReleaser

If you use [GoReleaser](https://github.com/goreleaser/goreleaser-action) to generate your build, you can easily
generate SLSA3 provenance by updating your existing workflow with the steps indicated in the workflow below:

1. Declare an `outputs` for the GoReleaser job:

```yaml
jobs:
  goreleaser:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Add an `id: run-goreleaser` field to your goreleaser step:

```yaml
    steps:
      [...]
      - name: Run GoReleaser
        id: run-goreleaser
        uses: goreleaser/goreleaser-action@b953231f81b8dfd023c58e0854a721e35037f28b # tag=v3

```

3. Add a step to generate the provenance subjects as shown below:

```yaml
- name: Generate subject
  id: hash
  env:
    ARTIFACTS: "${{ steps.run-goreleaser.outputs.artifacts }}"
  run: |
    set -euo pipefail

    checksum_file=$(echo "$ARTIFACTS" | jq -r '.[] | select (.type=="Checksum") | .path')
    echo "hashes=$(cat $checksum_file | base64 -w0)" >> "$GITHUB_OUTPUT"
```

4. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [goreleaser]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.goreleaser.outputs.hashes }}"
    upload-assets: true # upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  goreleaser:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3

      - name: Run GoReleaser
        id: run-goreleaser
        uses: goreleaser/goreleaser-action@b953231f81b8dfd023c58e0854a721e35037f28b # tag=v3

      - name: Generate subject
        id: hash
        env:
          ARTIFACTS: "${{ steps.run-goreleaser.outputs.artifacts }}"
        run: |
          set -euo pipefail

          checksum_file=$(echo "$ARTIFACTS" | jq -r '.[] | select (.type=="Checksum") | .path')
          echo "hashes=$(cat $checksum_file | base64 -w0)" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [goreleaser]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.goreleaser.outputs.hashes }}"
      upload-assets: true # upload to a new release
```

### Provenance for Bazel

If you use [Bazel](https://bazel.build/) to generate your artifacts, you can easily generate SLSA3 provenance by updating your existing workflow with steps indicated in the workflow below:

1. Declare an `outputs` for the hashes:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Build your project and copy the binaries from `bazel-bin` path (i.e., Bazel sandbox) to the root of the repository for easier reference (this makes it easier to upload these to the release too!):

```yaml
    steps:
      [...]
      - name: Build using bazel
        run: |
          # Your normal build workflow targets here
          bazel build //path/to/target_binary //path/to_another/binary

          # Copy the binaries.
          cp bazel-bin/path/to/target_binary .
          cp bazel-bin/path/to/another/binary .

```

3. Add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for:

```yaml
- name: Generate subject
  id: hash
  run: |
    set -euo pipefail

    sha256sum target_binary binary > checksums

    echo "hashes=$(cat checksums | base64 -w0)" >> "$GITHUB_OUTPUT"
```

4. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
    upload-assets: true # Optional: Upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3

      - name: Build using bazel
        run: |
          # Your normal build workflow targets here
          bazel build //path/to/target_binary //path/to_another/binary

          # Copy the binaries.
          cp bazel-bin/path/to/target_binary .
          cp bazel-bin/path/to/another/binary .

      - name: Generate subject
        id: hash
        run: |
          set -euo pipefail

          sha256sum target_binary binary > checksums

          echo "hashes=$(cat checksums | base64 -w0)" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [build]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

### Provenance for Java

If you develop with Java and use [Maven](#maven) or [Gradle](#gradle), you can
easily generate SLSA3 provenance by updating your existing workflow with the
steps indicated in the workflow below:

#### Maven

1. Declare an `outputs` for the artifacts generated by the build and their hashes:

```yaml
jobs:
  build:
    outputs:
      artifacts: ${{ steps.build.outputs.artifacts }}
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Add an `id: build` field to your maven build step; and save the location of the maven output files for easier reference:

```yaml
    steps:
      [...]
      - name: Build using maven
        id: build
        run: |
          # Your normal build workflow targets here
          mvn clean package

          # Save the location of the maven output files for easier reference
          ARTIFACT_PATTERN=./target/$(mvn help:evaluate -Dexpression=project.artifactId -q -DforceStdout)-$(mvn help:evaluate -Dexpression=project.version -q -DforceStdout)*.jar
          echo "artifact_pattern=$ARTIFACT_PATTERN" >> "$GITHUB_OUTPUT"

```

3. Add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for:

```yaml
- name: Generate subject
  id: hash
  run: |
    echo "hashes=$(sha256sum ${{ steps.build.outputs.artifact_pattern }} | base64 -w0)" >> "$GITHUB_OUTPUT"
```

4. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
    upload-assets: true # Optional: Upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  build:
    outputs:
      artifacts: ${{ steps.build.outputs.artifacts }}
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3

      - name: Build using maven
        id: build
        run: |
          # Your normal build workflow targets here
          mvn clean package

          # Save the location of the maven output files for easier reference
          ARTIFACT_PATTERN=./target/$(mvn help:evaluate -Dexpression=project.artifactId -q -DforceStdout)-$(mvn help:evaluate -Dexpression=project.version -q -DforceStdout)*.jar
          echo "artifact_pattern=$ARTIFACT_PATTERN" >> "$GITHUB_OUTPUT"

      - name: Generate subject
        id: hash
        run: |
          echo "hashes=$(sha256sum ${{ steps.build.outputs.artifact_pattern }} | base64 -w0)" >> "$GITHUB_OUTPUT"

      - name: Upload build artifacts
        uses: actions/upload-artifact@3cea5372237819ed00197afe530f5a7ea3e805c8 # tag=v3
        with:
          name: maven-build-outputs
          path: ${{ steps.build.outputs.artifact_pattern }}
          if-no-files-found: error

  provenance:
    needs: [build]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

#### Gradle

1. Declare an `outputs` for the artifacts generated by the build and their hashes:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Add an `id: build` field to your gradle build ste:

```yaml
    steps:
      [...]
      - name: Build using gradle
        id: build
        run: |
          # Your normal build workflow targets here
          ./gradlew clean build
```

3. Add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for. (This build assumes build artifacts are saved in ./build/libs).

```yaml
- name: Generate subject
  id: hash
  run: |
    echo "hashes=$(sha256sum ./build/libs/* | base64 -w0)" >> "$GITHUB_OUTPUT"
```

4. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read
    id-token: write
    contents: read
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
    upload-assets: true # Optional: Upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3

      - name: Build using gradle
        id: build
        run: |
          # Your normal build workflow targets here
          ./gradlew clean build

      - name: Generate subject
        id: hash
        run: |
          echo "hashes=$(sha256sum ./build/libs/* | base64 -w0)" >> "$GITHUB_OUTPUT"

      - name: Upload build artifacts
        uses: actions/upload-artifact@3cea5372237819ed00197afe530f5a7ea3e805c8 # tag=v3
        with:
          name: gradle-build-outputs
          path: ./build/libs/
          if-no-files-found: error

  provenance:
    needs: [build]
    permissions:
      actions: read
      id-token: write
      contents: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

### Provenance for Rust

If you use [Cargo](https://doc.rust-lang.org/cargo/) to generate your artifacts, you can
easily generate SLSA3 provenance by updating your existing workflow with the
steps indicated in the workflow below:

1. Declare an `outputs` for the hashes:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Build your binaries. Then add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for:

```yaml
    steps:
      [...]
      - name: Build using cargo
        run: |
          # Your normal build workflow targets here.
          cargo build --release

          cp target/release/target_binary .

      # Generate the subject.
      - name: Generate subject
        id: hash
        run: |
          set -euo pipefail

          echo "hashes=$(sha256sum target_binary | base64 -w0)" >> "$GITHUB_OUTPUT"

```

3. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
    upload-assets: true # Optional: Upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3

      - name: Build using cargo
        run: |
          # Your normal build workflow targets here.
          cargo build --release

          cp target/release/target_binary .

      # Generate the subject.
      - name: Generate subject
        id: hash
        run: |
          set -euo pipefail

          echo "hashes=$(sha256sum target_binary | base64 -w0)" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [build]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

### Provenance for Haskell

If you use [Haskell](https://www.haskell.org/) (either via
[`cabal`](https://www.haskell.org/cabal/) or
[`stack`](https://docs.haskellstack.org/en/stable/README/)) to generate your
artifacts, you can easily generate SLSA3 provenance by updating your existing
workflow with the steps indicated in the workflow below.

1. Declare an `outputs` for the hashes:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
```

2. Build your binaries. Then add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for:

```yaml
    steps:
      [...]
      - name: Build using Haskell
        run: |
          # Your normal build workflow targets here.
          cabal build  # or stack build

          # Copy the binary to the root directory for easier reference
          # For Cabal, use the following command
          cp $(cabal list-bin .) .
          # For Stack, use the following command instead
          # cp $(stack path --local-install-root)/bin/target_binary .

      # Generate the subject.
      - name: Generate subject
        id: hash
        run: |
          set -euo pipefail

          echo "hashes=$(sha256sum target_binary | base64 -w0)" >> "$GITHUB_OUTPUT"

```

3. Call the generic workflow to generate provenance by declaring the job below:

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
    upload-assets: true # Optional: Upload to a new release
```

All in all, it will look as the following:

```yaml
jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3
      - name: Setup Haskell
        uses: haskell/actions/setup@745062a754c3c4b70b87cb93937ad443096cc94d # tag=v1

      - name: Build using Haskell
        run: |
          # Your normal build workflow targets here.
          cabal build  # or stack build

          # Copy the binary to the root directory for easier reference
          # For Cabal, use the following command
          cp $(cabal list-bin .) .
          # For Stack, use the following command instead
          # cp $(stack path --local-install-root)/bin/target_binary .

      # Generate the subject.
      - name: Generate subject
        id: hash
        run: |
          set -euo pipefail

          echo "hashes=$(sha256sum target_binary | base64 -w0)" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [build]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.2.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```
