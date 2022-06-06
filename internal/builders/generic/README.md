# Generation of SLSA3+ provenance for arbitrary projects

This document explains how to use the builder for projects for which there is no language-specific builder available.

---

- [Project Status](#project-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)

---

## Project Status

This project is currently under active development. The API could change while
approaching an initial release.

## Generating Provenance

`slsa-github-generator` uses a Github Actions reusable workflow to generate the
provenance.

### Getting Started

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.

Add a step to your workflow after you have built your project to generate a
sha256 hash of your artifacts and base64 encode it.

Assuming you have a binary called `binary-linux-amd64` you can use the
`sha256sum` and `base64` commands to create the digest. Here we use the `-w0` to
output the encoded data on one line and make it easier to use as a Github Actions
output:

```shell
$ sha256sum artifact1 artifact2 ... | base64 -w0
```

After you have encoded your digest, add a new job to call the
`slsa-github-generator` reusable workflow. Here's an example of what it might
look like all together.

```yaml
jobs:
  build:
    outputs:
      digest: ${{ steps.hash.outputs.digest }}
    runs-on: ubuntu-latest
    steps:
      - name: "build artifacts"
        run: |
          # Build build artifacts here.
      - name: "generate hash"
        shell: bash
        id: hash
        run: |
          set -euo pipefail
          # sha256sum generates sha256 hash for all artifacts.
          # base64 -w0 encodes to base64 and outputs on a single line.
          # sha256sum artifact1 artifact2 ... | base64 -w0
          echo "::set-output name=digest::$(sha256sum artifact1 artifact2 | base64 -w0)"
  provenance:
    needs: [build]
    permissions:
      id-token: write
      contents: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/slsa2_provenance.yml@main
    with:
      base64-subjects: "${{ needs.build.outputs.digest }}"
```

### Workflow Inputs

The builder workflow
[.github/workflows/slsa2_provenance.yml](.github/workflows/slsa2_provenance.yml) accepts
the following inputs:

| Name              | Required | Description                                           |
| ----------------- | -------- | ----------------------------------------------------- |
| `base64-subjects` | yes      | Artifacts for which to generate provenance, formatted |
|                   |          | the same as the output of sha256sum                   |
|                   |          | (SHA256 NAME\n[...]) and base64 encoded.              |

### Workflow Outputs

The builder workflow
[.github/workflows/slsa2_provenance.yml](.github/workflows/slsa2_provenance.yml)
produces the following outputs:

| Name               | Description                                |
| ------------------ | ------------------------------------------ |
| `attestation-name` | The artifact name of the signed provenance |

### Provenance Format

The project generates SLSA provenance with the following values.

| Name                         | Value                                                          | Description                                                                                                                                                                                                            |
| ---------------------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator@v1"` | Identifies a generic GitHub Actions build.                                                                                                                                                                             |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                     | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

### Provenance Example

The following is an example of the generated proveanance. Provenance is
generated as an [in-toto](https://in-toto.io/) statement with a SLSA predecate.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "binary-linux-amd64",
      "digest": {
        "sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2"
      }
    },
  ],
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/slsa2_provenance.yml@refs/heads/main"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator@v1",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/heads/main.git",
        "digest": {
          "sha1": "3b5dc7cf5c0fd71c5a74c6b16cae78d49e03857c"
        },
        "entryPoint": "SLSA provenance"
      },
      "parameters": {},
      "environment": {
        "github_actor": "ianlewis",
        "github_base_ref": "",
        "github_event_name": "workflow_dispatch",
        "github_event_payload": ...,
        "github_head_ref": "",
        "github_ref": "refs/heads/main",
        "github_ref_type": "branch",
        "github_run_attempt": "1",
        "github_run_id": "2093917134",
        "github_run_number": "19",
        "github_sha1": "3b5dc7cf5c0fd71c5a74c6b16cae78d49e03857c"
      }
    },
    "metadata": {
      "buildInvocationID": "2182400786-1",
      "completeness": {
        "parameters": true,
        "environment": false,
        "materials": false
      },
      "reproducible": false
    },
    "materials": [
      {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/heads/main.git",
        "digest": {
          "sha1": "3b5dc7cf5c0fd71c5a74c6b16cae78d49e03857c"
        }
      }
    ]
  }
}
```
