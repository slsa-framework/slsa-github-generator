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

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the SLSA generator](#referencing-the-slsa-generator)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Integration With Other Build Systems](#integration-with-other-build-systems)
  - [Provenance for GoReleaser](#provenance-for-goreleaser)
  - [Provenance for JReleaser](#provenance-for-jreleaser)
  - [Provenance for Bazel](#provenance-for-bazel)
  - [Provenance for Java](#provenance-for-java)
    - [Maven](#maven)
    - [Gradle](#gradle)
  - [Provenance for Rust](#provenance-for-rust)
  - [Provenance for Haskell](#provenance-for-haskell)
  - [Provenance for Python](#provenance-for-python)
- [Provenance for matrix strategy builds](#provenance-for-matrix-strategy-builds)
  - [A single provenance attestation for all artifacts](#a-single-provenance-attestation-for-all-artifacts)
  - [A different attestation for each iteration](#a-different-attestation-for-each-iteration)
- [Provenance for artifacts built across multiple operating systems](#provenance-for-artifacts-built-across-multiple-operating-systems)
- [Known Issues](#known-issues)
  - [Skip output 'hashes' since it may contain secret](#skip-output-hashes-since-it-may-contain-secret)
  - ['internal error' when using `upload-assets`](#internal-error-when-using-upload-assets)
  - [error updating to TUF remote mirror: tuf: invalid key](#error-updating-to-tuf-remote-mirror-tuf-invalid-key)
  - [Compatibility with `actions/download-artifact`](#compatibility-with-actionsdownload-artifact)

<!-- tocstop -->

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
sha256sum artifact1 artifact2 ... | base64 -w0
```

This workflow expects the `base64-subjects` input to decode to a string conforming to the expected output of the `sha256sum` command. Specifically, the decoded output is expected to be comprised of a hash value followed by a space followed by the artifact name.

After you have encoded your digest, add a new job to call the reusable workflow.

```yaml
provenance:
  permissions:
    actions: read # Needed for detection of GitHub Actions environment.
    id-token: write # Needed for provenance signing and ID
    contents: write # Needed for release uploads
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
  with:
    base64-subjects: "${{ needs.build.outputs.hashes }}"
```

The `base64-subjects` input has a maximum length as defined by [ARG_MAX](https://www.in-ulm.de/~mascheck/various/argmax/) on the runner. If you need to attest to a large number of files that exceeds the maximum length, use the `base64-subjects-as-file` input option instead. Another use case for this option is when GitHub Actions runner masks the job output because it detects a secret (see the discussion [here](https://github.com/orgs/community/discussions/37942)). This option requires that you save the output of the sha256sum command into a file:

```shell
sha256sum artifact1 artifact2 ... | base64 -w0 > large_digests_file.text
```

The you must then share this file with the generator using the [actions/generator/generic/create-base64-subjects-from-file Action](https://github.com/slsa-framework/slsa-github-generator/tree/main/actions/generator/generic/create-base64-subjects-from-file):

```yaml
build:
  outputs:
    subjects-as-file: ${{ steps.hashes.outputs.handle }}
  ...
    uses: slsa-framework/slsa-github-generator/actions/generator/generic/create-base64-subjects-from-file@v2.0.0
    id: hashes
    with:
      path: large_digests_file.text
provenance:
  permissions:
    actions: read # Needed for detection of GitHub Actions environment.
    id-token: write # Needed for provenance signing and ID
    contents: write # Needed for release uploads
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
  with:
    base64-subjects-as-file: "${{ needs.build.outputs.subjects-as-file }}"
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
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
        uses: actions/download-artifact@c850b930e6ba138125429b7e5c93fc707a7f8427 # v4.1.4
        with:
          name: artifact1

      - name: Download artifact2
        uses: actions/download-artifact@c850b930e6ba138125429b7e5c93fc707a7f8427 # v4.1.4
        with:
          name: artifact2

      - name: Upload assets
        uses: softprops/action-gh-release@69320dbe05506a9a39fc8ae11030b214ec2d1f87 # v2.0.5
        with:
          files: |
            artifact1
            artifact2
```

### Referencing the SLSA generator

At present, the generator **MUST** be referenced
by a tag of the form `@vX.Y.Z`, because the build will fail if you reference it via a shorter tag like `@vX.Y` or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
https://rekor.sigstore.dev/. This entry includes the repository name. This will cause the
private repository name to leak and be discoverable via the public Rekor API
server.

If this is ok with you, you can set the `private-repository` flag in order to
opt in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  private-repository: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue #372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

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

| Name                      | Required                                                           | Default                                                                                         | Description                                                                                                                                                                                                                                                                        |
| ------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `base64-subjects`         | One of `base64-subjects` or `base64-subjects-as-file` is required. |                                                                                                 | Artifact(s) for which to generate provenance, formatted the same as the output of sha256sum (SHA256 NAME\n[...]) and base64 encoded. The encoded value should decode to, for example: `90f3f7d6c862883ab9d856563a81ea6466eb1123b55bff11198b4ed0030cac86 foo.zip`                   |
| `base64-subjects-as-file` | One of `base64-subjects` or `base64-subjects-as-file` is required. |                                                                                                 | The name of a artifacts containing formatted subjects as uploaded by the [actions/generator/generic/create-base64-subjects-from-file Action](https://github.com/slsa-framework/slsa-github-generator/tree/main/actions/generator/generic/create-base64-subjects-from-file) action. |
| `upload-assets`           | no                                                                 | false                                                                                           | If true provenance is uploaded to a GitHub release for new tags.                                                                                                                                                                                                                   |
| `upload-tag-name`         | no                                                                 |                                                                                                 | If specified and `upload-assets` is set to true, the provenance will be uploaded to a Github release identified by the tag-name regardless of the triggering event.                                                                                                                |
| `provenance-name`         | no                                                                 | "(subject name).intoto.jsonl" if a single subject. "multiple.intoto.json" if multiple subjects. | The artifact name of the signed provenance. The file must have the `intoto.jsonl` extension.                                                                                                                                                                                       |
| `private-repository`      | no                                                                 | false                                                                                           | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).                                                    |
| `continue-on-error`       | no                                                                 | false                                                                                           | Set to true to ignore errors. This option is useful if you won't want a failure to fail your entire workflow.                                                                                                                                                                      |
| `draft-release`           | no                                                                 | false                                                                                           | If true, the release is created as a draft                                                                                                                                                                                                                                         |

### Workflow Outputs

The [generic workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_generic_slsa3.yml) produces the following outputs:

| Name               | Description                                                                                     |
| ------------------ | ----------------------------------------------------------------------------------------------- |
| `provenance-name`  | The artifact name of the signed provenance.                                                     |
| `outcome`          | If `continue-on-error` is `true`, will contain the outcome of the run (`success` or `failure`). |

### Provenance Format

The project generates SLSA provenance with the following values.

| Name                         | Value                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/generic@v1"` | Identifies a generic GitHub Actions build.                                                                                                                                                                             |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                             | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

**Note**: The generated provenance will probably be wrapped in a [DSSE](https://github.com/secure-systems-lab/dsse) envelope and encoded in base64. Check the human-readable result running `cat encoded-artifact.intoto.jsonl | jq -r '.payload' | base64 -d | jq`.

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
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.2.2"
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

**Notes**:

- Make sure you did not disable checksum generation in the goreleaser yml.
- Make sure you specified sha256 as the algorithm for the checksum or left it empty (sha256 is the default).
- To enable provenance generation for dockers (as well as artifacts), use [goreleaser version >= v1.13.0](https://github.com/goreleaser/goreleaser/releases/tag/v1.13.0).

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
           uses: goreleaser/goreleaser-action@b508e2e3ef3b19d4e4146d4f8fb3ba9db644a757 # tag=v3.2.0
   ```

3. Add a step to generate the provenance subjects as shown below:

   ```yaml
   - name: Generate subject
     id: hash
     env:
       ARTIFACTS: "${{ steps.run-goreleaser.outputs.artifacts }}"
     run: |
       set -euo pipefail
       hashes=$(echo $ARTIFACTS | jq --raw-output '.[] | {name, "digest": (.extra.Digest // .extra.Checksum)} | select(.digest) | {digest} + {name} | join("  ") | sub("^sha256:";"")' | base64 -w0)
       if test "$hashes" = ""; then # goreleaser < v1.13.0
         checksum_file=$(echo "$ARTIFACTS" | jq -r '.[] | select (.type=="Checksum") | .path')
         hashes=$(cat $checksum_file | base64 -w0)
       fi
       echo "hashes=$hashes" >> $GITHUB_OUTPUT
   ```

4. Call the generic workflow to generate provenance by declaring the job below:

   ```yaml
   provenance:
     needs: [goreleaser]
     permissions:
       actions: read # To read the workflow path.
       id-token: write # To sign the provenance.
       contents: write # To add assets to a release.
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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

          hashes=$(echo $ARTIFACTS | jq --raw-output '.[] | {name, "digest": (.extra.Digest // .extra.Checksum)} | select(.digest) | {digest} + {name} | join("  ") | sub("^sha256:";"")' | base64 -w0)
          if test "$hashes" = ""; then # goreleaser < v1.13.0
            checksum_file=$(echo "$ARTIFACTS" | jq -r '.[] | select (.type=="Checksum") | .path')
            hashes=$(cat $checksum_file | base64 -w0)
          fi
          echo "hashes=$hashes" >> $GITHUB_OUTPUT

  provenance:
    needs: [goreleaser]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.goreleaser.outputs.hashes }}"
      upload-assets: true # upload to a new release
```

For more details, follow the great blog post on [goreleaser.com](https://goreleaser.com/blog/slsa-generation-for-your-artifacts).

### Provenance for JReleaser

If you use [JReleaser](https://github.com/jreleaser/release-action) to generate your build, you can easily
generate SLSA3 provenance by updating your existing workflow with the steps indicated in the workflow below:

**Notes**:

1. Declare an `outputs` for the job that runs JReleaser:

   ```yaml
   jobs:
     release:
       outputs:
         hashes: ${{ steps.slsa.outputs.hashes }}
         tagname: ${{ steps.slsa.outputs.tagname }}
   ```

2. Build your project and release it:

   ```yaml
   # project specific build instructions
   - name: Build
     run: |
       make build

   - name: Run JReleaser
     uses: jreleaser/release-action@f2226e009ec9445383677f56482ca3181d649bcc # branch=v2
     with:
       arguments: full-release
     env:
       JRELEASER_PROJECT_VERSION: 1.2.3 # value supplied as input or read from sources
       JRELEASER_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```

3. Add a step to generate the provenance subjects as shown below:

   ```yaml
   - name: Generate subject
     shell: bash
     id: slsa
     run: |
       echo "hashes=$(cat out/jreleaser/checksums/checksums_sha256.txt | base64 -w0)" >> "$GITHUB_OUTPUT"
       echo "tagname=$(grep tagName out/jreleaser/output.properties | awk -F'=' '{print $2}')" >> "$GITHUB_OUTPUT"
   ```

4. Call the generic workflow to generate provenance by declaring the job below:

   ```yaml
   provenance:
     needs: [release]
     permissions:
       actions: read # To read the workflow path.
       id-token: write # To sign the provenance.
       contents: write # To add assets to a release.
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
     with:
       base64-subjects: ${{ needs.release.outputs.hashes }}
       upload-assets: true # upload to a new release
       upload-tag-name: ${{ needs.release.outputs.tagname }}
   ```

All in all, it will look as the following:

```yaml
jobs:
  release:
    outputs:
      hashes: ${{ steps.slsa.outputs.hashes }}
      tagname: ${{ steps.slsa.outputs.tagname }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # tag=v3
        with:
          fetch-depth: 0

        # project specific build instructions
      - name: Build
        run: |
          make build

      - name: Run JReleaser
        uses: jreleaser/release-action@f2226e009ec9445383677f56482ca3181d649bcc # branch=v2
        with:
          arguments: full-release
        env:
          JRELEASER_PROJECT_VERSION: 1.2.3 # value supplied as input or read from sources
          JRELEASER_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Generate subject
        shell: bash
        id: slsa
        run: |
          echo "hashes=$(cat out/jreleaser/checksums/checksums_sha256.txt | base64 -w0)" >> "$GITHUB_OUTPUT"
          echo "tagname=$(grep tagName out/jreleaser/output.properties | awk -F'=' '{print $2}')" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [release]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: ${{ needs.release.outputs.hashes }}
      upload-assets: true # upload to a new release
      upload-tag-name: ${{ needs.release.outputs.tagname }}
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
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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

2. Add an `id: build` field to your maven build step and save the location of the maven output files for easier reference:

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
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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

2. Add an `id: build` field to your gradle build step:

   ```yaml
   steps:
     [...]
     - name: Build using gradle
       id: build
       run: |
         # Your normal build workflow targets here
         ./gradlew clean build
   ```

3. Add a step to generate the provenance subjects as shown below. Update the
   sha256 sum arguments to include all binaries that you generate provenance
   for. (This build assumes build artifacts are saved in `./build/libs`).

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
       contents: write
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
     with:
       base64-subjects: "${{ needs.build.outputs.hashes }}"
       upload-assets: true # Optional: Upload to a new release
   ```

All in all, it will look as the following:

```yaml
Jobs:
  build:
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}

    runs-on: ubuntu-latest
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
      contents: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
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
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

### Provenance for Python

If you develop with Python you can
easily generate SLSA3 provenance by updating your existing workflow with the
steps indicated in the workflow below:

1. Declare an outputs for the artifacts generated by the build and their hashes:

   ```yaml
   jobs:
     build:
       name: "Build dists"
       runs-on: "ubuntu-latest"
       environment:
         name: "publish"
       outputs:
         hashes: ${{ steps.hash.outputs.hashes }}
   ```

2. Add an id: build field to your python build step

   ```yaml
   steps:
     - name: "Checkout repository"
       uses: "actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b" # tag=v3

     - name: "Setup Python"
       uses: "actions/setup-python@13ae5bb136fac2878aff31522b9efb785519f984" # tag=v4
       with:
         python-version: "3.x"

     - name: "Install dependencies"
       run: python -m pip install build

     - name: Build using python
       id: build
       run: python -m build
   ```

3. Add a step to generate the provenance subjects as shown below. Update the sha256 sum arguments to include all binaries that you generate provenance for:

   ```yaml
   - name: Generate subject
     id: hash
     run: |
       cd dist
       HASHES=$(sha256sum * | base64 -w0)
       echo "hashes=$HASHES" >> "$GITHUB_OUTPUT"
   ```

4. Call the generic workflow to generate provenance by declaring the job below:

   ```yaml
   provenance:
     needs: [build]
     permissions:
       actions: read # To read the workflow path.
       id-token: write # To sign the provenance.
       contents: write # To add assets to a release.
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
     with:
       base64-subjects: "${{ needs.build.outputs.hashes }}"
       upload-assets: true # Optional: Upload to a new release
   ```

All in all, it will look as the following:

```yaml
jobs:
  build:
    name: "Build dists"
    runs-on: "ubuntu-latest"
    environment:
      name: "publish"
    outputs:
      hashes: ${{ steps.hash.outputs.hashes }}
  steps:
    - name: "Checkout repository"
      uses: "actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b" # tag=v3

    - name: "Setup Python"
      uses: "actions/setup-python@13ae5bb136fac2878aff31522b9efb785519f984" # tag=v4
      with:
        python-version: "3.x"

    - name: "Install dependencies"
      run: python -m pip install build

    - name: Build using Python
      id: build
      run: |
        python -m build

    - name: Generate subject
      id: hash
      run: |
        cd dist
        HASHES=$(sha256sum * | base64 -w0)
        echo "hashes=$HASHES" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [build]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.build.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

## Provenance for matrix strategy builds

There are a few ways to handle provenance for matrix builds. You can create a
single provenance file describing all the artifacts from the different
iterations or a different file for each iteration's artifact(s).

Regardless of your choice, there's unfortunately a bit of necessary boilerplate.

### A single provenance attestation for all artifacts

1. As with the examples above, the first thing to do is define the build job,
   with its outputs and its matrix strategy.

   GitHub currently doesn't support different outputs for matrix builds. We must
   therefore declare a different hash output for each iteration. A follow-up job
   will collate all the hashes into a single string.

   ```yaml
   jobs:
     build:
       strategy:
         matrix:
           color: ["red", "blue", "green"]
           flavor: ["mint", "vanilla"]
       outputs:
         # The key-names are actually irrelevant, but keep them descriptive
         hash-red-mint: ${{ steps.hash.outputs.hash-red-mint }}
         hash-red-vanilla: ${{ steps.hash.outputs.hash-red-vanilla }}
         hash-blue-mint: ${{ steps.hash.outputs.hash-blue-mint }}
         hash-blue-vanilla: ${{ steps.hash.outputs.hash-blue-vanilla }}
         hash-green-mint: ${{ steps.hash.outputs.hash-green-mint }}
         hash-green-vanilla: ${{ steps.hash.outputs.hash-green-vanilla }}
   ```

2. You'll now have to build your project as usual:

   ```yaml
   steps:
     # whatever you need to do to build (checkout, setup the environment,
     # get dependencies, compile...)
     - ...
     - ...
     - ...
   ```

3. As with the other examples, you'll then have to generate the hashes that
   represent your build. This step is effectively identical to all the examples
   above, except each iteration must store its hash in a different output
   variable.

   ```yaml
   - name: Generate subject
     id: hash
     run: |
       echo "hash-${{ matrix.color }}-${{ matrix.flavor }}=$( \
         sha256sum ... | base64 -w0 \
       )" >> "$GITHUB_OUTPUT"
   ```

4. Now you'll collate all the individual hashes into a single bas64 string.

   ```yaml
   combine_hashes:
     needs: [build]
     outputs:
       hashes: ${{ steps.hashes.outputs.hashes }}
     env:
       HASHES: ${{ toJSON(needs.build.outputs) }}
     steps:
       - id: hashes
         run: |
           echo "$HASHES" | jq -r '.[] | @base64d' | sed "/^$/d" > hashes.txt
           echo "hashes=$(cat hashes.txt | base64 -w0)" >> "$GITHUB_OUTPUT"
   ```

5. The provenance job is also effectively identical to the examples above,
   except that it relies on `combine_hashes` instead of the `build` job.

   ```yaml
   provenance:
     needs: [combine_hashes]
     permissions:
       actions: read # To read the workflow path.
       id-token: write # To sign the provenance.
       contents: write # To add assets to a release.
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
     with:
       base64-subjects: "${{ needs.combine_hashes.outputs.hashes }}"
       upload-assets: true # Optional: Upload to a new release
   ```

Now all together:

```yaml
jobs:
  build:
    strategy:
      matrix:
        color: ["red", "blue", "green"]
        flavor: ["mint", "vanilla"]
    outputs:
      hash-red-mint: ${{ steps.hash.outputs.hash-red-mint }}
      hash-red-vanilla: ${{ steps.hash.outputs.hash-red-vanilla }}
      hash-blue-mint: ${{ steps.hash.outputs.hash-blue-mint }}
      hash-blue-vanilla: ${{ steps.hash.outputs.hash-blue-vanilla }}
      hash-green-mint: ${{ steps.hash.outputs.hash-green-mint }}
      hash-green-vanilla: ${{ steps.hash.outputs.hash-green-vanilla }}
    steps:
      # all your build steps
      - ...
      - ...
      - ...

      - name: Generate subject
        id: hash
        run: |
          echo "hash-${{ matrix.color }}-${{ matrix.flavor }}=$( \
            sha256sum ... | base64 -w0 \
          )" >> "$GITHUB_OUTPUT"

  combine_hashes:
    needs: [build]
    outputs:
      hashes: ${{ steps.hashes.outputs.hashes }}
    env:
      HASHES: ${{ toJSON(needs.build.outputs) }}
    steps:
      - id: hashes
        run: |
          echo "$HASHES" | jq -r '.[] | @base64d' | sed "/^$/d" > hashes.txt
          echo "hashes=$(cat hashes.txt | base64 -w0)" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [combine_hashes]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.combine_hashes.outputs.hashes }}"
      upload-assets: true # Optional: Upload to a new release
```

This will create a single provenance file that describes all of the built
artifacts. The default name will be `multiple.intoto.jsonl`, but can be modified
by passing the `provenance-name` argument to the generator.

Should your build job require other outputs (not just the hashes, but other
values such as the artifact name, for example), you'll need to change the `jq`
command within the `combine_hashes` job. You'll use a regex to only use the hash
values. Assuming your hashes are stored in
`needs.build.outputs.[hash-$color-$flavor]`:

```bash
... | jq -r 'with_entries(select(.key | match("hash-.*-.*")))[] | @base64d' | ...
```

### A different attestation for each iteration

This case is simpler. We can copy the single-attestation version's steps 1-3 and
ignore step 4's `combine_hashes` job entirely. The changes are entirely within
the provenance job, where we'll have to repeat the build's matrix strategy and
use its values to define unique names for each provenance attestation using the
[format](https://docs.github.com/en/actions/learn-github-actions/expressions#format)
function.

```yaml
provenance:
  needs: [build]
  strategy:
    matrix:
      color: ["red", "blue", "green"]
      flavor: ["mint", "vanilla"]
  permissions:
    actions: read # To read the workflow path.
    id-token: write # To sign the provenance.
    contents: write # To add assets to a release.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
  with:
    base64-subjects: "${{ needs.build.outputs[format('hash-{0}-{1}', matrix.color, matrix.flavor)] }}"
    upload-assets: true # Optional: Upload to a new release
```

So, all together, this version becomes:

```yaml
jobs:
  build:
    strategy:
      matrix:
        color: ["red", "blue", "green"]
        flavor: ["mint", "vanilla"]
    outputs:
      hash-red-mint: ${{ steps.hash.outputs.hash-red-mint }}
      hash-red-vanilla: ${{ steps.hash.outputs.hash-red-vanilla }}
      hash-blue-mint: ${{ steps.hash.outputs.hash-blue-mint }}
      hash-blue-vanilla: ${{ steps.hash.outputs.hash-blue-vanilla }}
      hash-green-mint: ${{ steps.hash.outputs.hash-green-mint }}
      hash-green-vanilla: ${{ steps.hash.outputs.hash-green-vanilla }}
    steps:
      # all your build steps
      - ...
      - ...
      - ...

      - name: Generate subject
        id: hash
        run: |
          echo "hash-${{ matrix.color }}-${{ matrix.flavor }}=$( \
            sha256sum ... | base64 -w0 \
          )" >> "$GITHUB_OUTPUT"

  provenance:
    needs: [build]
    strategy:
      matrix:
        color: ["red", "blue", "green"]
        flavor: ["mint", "vanilla"]
    permissions:
      actions: read # To read the workflow path.
      id-token: write # To sign the provenance.
      contents: write # To add assets to a release.
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.build.outputs[format('hash-{0}-{1}', matrix.color, matrix.flavor)] }}"
      upload-assets: true # Optional: Upload to a new release
```

## Provenance for artifacts built across multiple operating systems

If a single release job produces artifacts for multiple operating systems (using
matrix strategy), there are a few more caveats to consider:

1. First, it is ideal to make Windows behave the same as Linux and MacOS by
   making the runner use `bash` as the shell to execute commands in:

   ```yaml
   defaults:
     run:
       shell: bash
   ```

2. Optionally, you might also want to make the workflow use LF as line
   terminator even on Windows:

   ```yaml
       - run: git config --global core.autocrlf input
       # Alternatively, also force line endings for every file
       # - run: |
       #     git config --global core.eol lf
       #     git config --global core.autocrlf input
   ```

The other complexity arises from the fact that the utilities used to compute the
digest (`sha256sum`) and the base 64 encoding (`base64`) have different
behaviors across the operating systems:

- On MacOS, the utlity to compute the digest is called `shasum` and the
  algorithm is passed via the `-a 256` algorithm
- On Windows, we need to tell `sha256sum` to treat all files as text by using
  the `-t` switch, otherwise the output will contain an extra `*` in front of
  the filename. This will later be wrongly interpretted as a glob pattern, so we
  should avoid it.
- On MacOS, `base64` does not have a `-w0` option, the line wrapping is
  implicit.

One way to merge all these differences is to use the bash `||` operator:

```yaml
      - id: hash
        run: |
          set -euo pipefail
          (sha256sum -t release_artifact_${{ runner.os }} || shasum -a 256 release_artifact_${{ runner.os }}) > checksum
          echo "hash-${{ matrix.os }}=$(base64 -w0 checksum || base64 checksum)" >> "${GITHUB_OUTPUT}"
```

Thus, to generate a single provenance for artifacts built on all 3 operating
systems, you would use the following example:

```yaml
defaults:
  run:
    shell: bash

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false # Don't cancel other jobs if one fails
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    outputs:
      hash-ubuntu-latest: ${{ steps.hash.outputs.hash-ubuntu-latest }}
      hash-macos-latest: ${{ steps.hash.outputs.hash-macos-latest }}
      hash-windows-latest: ${{ steps.hash.outputs.hash-windows-latest }}
    steps:
      - run: git config --global core.autocrlf input
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          persist-credentials: false

      # Do the build to create release_artifact_${{ runner.os }}
      - run: ...

      - uses: actions/upload-artifact@5d5d22a31266ced268874388b861e4b58bb5c2f3 # v4.3.1
        with:
          path: release_artifact_${{ runner.os }}
          name: release_artifact_${{ runner.os }}
          if-no-files-found: error
      - id: hash
        run: |
          set -euo pipefail
          (sha256sum -t release_artifact_${{ runner.os }} || shasum -a 256 release_artifact_${{ runner.os }}) > checksum
          echo "hash-${{ matrix.os }}=$(base64 -w0 checksum || base64 checksum)" >> "${GITHUB_OUTPUT}"

  provenance:
    needs: [build]
    strategy:
      fail-fast: false # Don't cancel other jobs if one fails
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    permissions:
      actions: read
      id-token: write
      contents: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.build.outputs[format('hash-{0}', matrix.os)] }}"
      upload-assets: true # NOTE: This does nothing unless 'upload-tag-name' parameter is also set to an existing tag
```

Alternatively, to generate 3 different provenance statements, one per each
artifact, you would use the following example:

```yaml
defaults:
  run:
    shell: bash

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false # Don't cancel other jobs if one fails
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    outputs:
      hash-ubuntu-latest: ${{ steps.hash.outputs.hash-ubuntu-latest }}
      hash-macos-latest: ${{ steps.hash.outputs.hash-macos-latest }}
      hash-windows-latest: ${{ steps.hash.outputs.hash-windows-latest }}
    steps:
      - run: git config --global core.autocrlf input
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          persist-credentials: false

      # Do the build to create release_artifact_${{ runner.os }}
      - run: ...

      - uses: actions/upload-artifact@5d5d22a31266ced268874388b861e4b58bb5c2f3 # v4.3.1
        with:
          path: release_artifact_${{ runner.os }}
          name: release_artifact_${{ runner.os }}
          if-no-files-found: error
      - id: hash
        run: |
          set -euo pipefail
          (sha256sum -t release_artifact_${{ runner.os }} || shasum -a 256 release_artifact_${{ runner.os }}) > checksum
          echo "hash-${{ matrix.os }}=$(base64 -w0 checksum || base64 checksum)" >> "${GITHUB_OUTPUT}"

  combine_hashes:
    needs: [build]
    runs-on: ubuntu-latest
    outputs:
      hashes: ${{ steps.combine_hashes.outputs.hashes }}
    env:
      HASHES: ${{ toJSON(needs.build.outputs) }}
    steps:
      - id: combine_hashes
        run: |
          set -euo pipefail
          echo "${HASHES}" | jq -r '.[] | @base64d' | sed "/^$/d" > hashes
          echo "hashes=$(base64 -w0 hashes)" >> "${GITHUB_OUTPUT}"

  provenance:
    needs: [combine_hashes]
    permissions:
      actions: read
      id-token: write
      contents: write
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
    with:
      base64-subjects: "${{ needs.combine_hashes.outputs.hashes }}"
      upload-assets: true # NOTE: This does nothing unless 'upload-tag-name' parameter is also set to an existing tag

```

## Known Issues

### Skip output 'hashes' since it may contain secret

The GitHub Actions runner sometimes masks the job output if it potentially contains a secret. One solution is to use the ``base64-subjects-as-file`` option to pass the artifact hashes using an existing file instead. See the instructions [here](#getting-started) to use the ``base64-subjects-as-file`` option.

### 'internal error' when using `upload-assets`

**Affected versions:** v1.5.0

When setting `upload-assets` to `true` when the trigger occurred on a git ref
that is a not a tag (e.g. a push to a branch), the workflow would fail with
`'internal error'` if `upload-tag-name` input was not specified.

Prior to 1.5.0 `upload-assets` was ignored if a tag name could not be determined
by the git ref.

Please set the following to `upload-assets`:

```yaml
upload-assets: ${{ startsWith(github.ref, 'refs/tags/') }}
```

or specify a value for `upload-tag-name`:

```yaml
upload-assets: true
upload-tag-name: "v1.0.0"
```

See issue
[#1702](https://github.com/slsa-framework/slsa-github-generator/issues/1702) for
more details.

### error updating to TUF remote mirror: tuf: invalid key

**Affected versions:** v1.2.x

Workflows are currently failing with the error:

```text
validating log entry: unable to fetch Rekor public keys from TUF repository, and not trusting the Rekor API for fetching public keys: updating local metadata and targets: error updating to TUF remote mirror: tuf: invalid key
```

This issue is currently tracked by [issue #1163](https://github.com/slsa-framework/slsa-github-generator/issues/1163)

You can work around this error by setting `compile-generator` input flag.

```yaml
with:
  compile-generator: true
```

This will compile the generator binary used by the workflow instead of
downloading the latest release. Make sure you continue to reference the workflow
using a release tag in order to allow verification by `slsa-verifier`.

```yaml
uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.0.0
```

### Compatibility with `actions/download-artifact`

`slsa-github-generator@v1.9.0` and prior use [`actions/upload-artifact@v3`](https://github.com/actions/upload-artifact) and [`actions/download-artifact@v3`](https://github.com/actions/download-artifact) which are not backwards compatible the `@v4`s used in current versions of `slsa-github-generator`.
The interface remains the same, however. If your own workflows want to download artifacts produced by our workflows, they must begin using `actions/download-artifact@v4`. For your other dependent workflows, you may find that you need to upgrade all of your uses of both of the actions to `@v4` to maintain compatibility.

See more migration guidance

- https://github.com/actions/upload-artifact/blob/main/docs/MIGRATION.md
- https://github.com/actions/download-artifact/blob/main/docs/MIGRATION.md

This is part of our effort to upgrade from the now-deprecated node16 that the `@v3`s used. `@v4s` use node20.
