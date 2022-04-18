# slsa-github-generator

This repository contains a reference implementation for generating non-forgeable
[SLSA provenance](https://slsa.dev/) that meets the requirements for the [SLSA
level 2 and above](https://slsa.dev/spec/v0.1/levels) using a [Github Actions
reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows).

---

- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)

---

## Generating Provenance

`slsa-github-generator` uses a Github Actions reusable workflow to generate the
provenance.

### Getting Started

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.

Add a step to your workflow after you have built your project to generate a
sha256 hash of your artifacts. The following assumes you have a binary called
`binary-linux-amd64`.

After that, add a new job to call the `slsa-github-generator` reusable workflow.

```yaml
jobs:
  build:
    outputs:
      digest: ${{ steps.hash.outputs.digest }}
    runs-on: ubuntu-latest
    steps:
      # Your build steps are here.
      - name: "generate hash"
        shell: bash
        id: hash
        run: |
          set -euo pipefail
          DIGEST=$(sha256sum binary-linux-amd64)
          DIGEST="${DIGEST//'%'/'%25'}"
          DIGEST="${DIGEST//$'\n'/'%0A'}"
          DIGEST="${DIGEST//$'\r'/'%0D'}"
          echo "::set-output name=digest::$DIGEST"
  provenance:
    needs: [build]
    permissions:
      id-token: write
      contents: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/slsa2_provenance.yml@main
    with:
      subjects: "${{ needs.build.outputs.digest }}"
```

### Workflow Inputs

The builder workflow
[.github/workflows/slsa2_provenance.yml](.github/workflows/slsa2_provenance.yml) accepts
the following inputs:

| Name       | Required | Description                                                                                                    |
| ---------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `subjects` | yes      | Artifacts for which to generate provenance, formatted the same as the output of sha256sum (SHA256 NAME\n[...]) |

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
