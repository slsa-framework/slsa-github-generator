# Generation of SLSA3+ provenance for container images

This document explains how to generate SLSA provenance for container images.

This can be done by adding an additional step to your existing Github Actions
workflow to call a [reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to generate generic SLSA provenance. We'll call this workflow the "container
workflow" from now on.

The container workflow differs from ecosystem specific builders (like the [Go
builder](../go)) which build the artifacts as well as generate provenance. This
project simply generates provenance as a separate step in an existing workflow.

---

- [Project Status](#project-status)
- [Benefits of Provenance](#benefits-of-provenance)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Verification](#verification)
  - [Kyverno](#kyverno)

---

## Project Status

This workflow is currently under active development. The API could change while
approaching an initial release. You can track progress towards General
Availability via [this milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/3).

## Benefits of Provenance

Using the container workflow will generate a non-forgeable attestation to the
container image using the identity of the GitHub workflow. This can be used
to create a positive attestation to a container image coming from your
repository.

That means that once your users verify the image they have downloaded they
can be sure that the image was created by your repository's workflow and
hasn't been tampered with.

## Generating Provenance

The container workflow uses a Github Actions reusable workflow to generate the
provenance.

### Getting Started

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.

```yaml
provenance:
  needs: [build]
  permissions:
    actions: read # for detecting the Github Actions environment.
    id-token: write # for creating OIDC tokens for signing.
    packages: write # for uploading attestations.
  if: startsWith(github.ref, 'refs/tags/')
  # TODO(https://github.com/slsa-framework/slsa-github-generator/issues/492): Use a tagged release once we have one.
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@main
  with:
    image: ${{ needs.build.outputs.tag }}
    digest: ${{ needs.build.outputs.digest }}
    registry-username: ${{ github.actor }}
    # TODO(https://github.com/slsa-framework/slsa-github-generator/issues/492): Remove after GA release.
    compile-generator: true
  secrets:
    registry-password: ${{ secrets.GITHUB_TOKEN }}
```

Here's an example of what it might look like all together.

```yaml
env:
  IMAGE_REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # This step builds our image, pushes it, and outputs the repo hash digest.
  build:
    permissions:
      contents: read
      packages: write
    outputs:
      image: ${{ steps.image.outputs.image }}
      digest: ${{ steps.build.outputs.digest }}
    runs-on: ubuntu-latest
    steps:
      - name: Checkout the repository
        uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # v2.3.4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@dc7b9719a96d48369863986a06765841d7ea23f6 # v2.0.0

      - name: Authenticate Docker
        uses: docker/login-action@49ed152c8eca782a232dede0303416e8f356c37b # v2.0.0
        with:
          registry: ${{ env.IMAGE_REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata (tags, labels) for Docker
        id: meta
        uses: docker/metadata-action@69f6fc9d46f2f8bf0d5491e4aabe0bb8c6a4678a # v4.0.1
        with:
          images: ${{ env.IMAGE_REGISTRY }}/${{ env.IMAGE_NAME }}

      - name: Build and push Docker image
        uses: docker/build-push-action@e551b19e49efd4e98792db7592c17c09b89db8d8 # v3.0.0
        id: build
        with:
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}

      - name: Output image
        id: image
        run: |
          # NOTE: Set the image as an output because the `env` context is not
          # available to the inputs of a reusable workflow call.
          image_name="${IMAGE_REGISTRY}/${IMAGE_NAME}"
          echo "::set-output name=image::$image_name"

  # This step calls the container workflow to generate provenance and push it to
  # the container registry.
  provenance:
    needs: [build]
    permissions:
      actions: read # for detecting the Github Actions environment.
      id-token: write # for creating OIDC tokens for signing.
      packages: write # for uploading attestations.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@main
    with:
      image: ${{ needs.build.outputs.image }}
      digest: ${{ needs.build.outputs.digest }}
      registry-username: ${{ github.actor }}
      # TODO(https://github.com/slsa-framework/slsa-github-generator/issues/492): Remove after GA release.
      compile-generator: true
    secrets:
      registry-password: ${{ secrets.GITHUB_TOKEN }}
```

### Supported Triggers

The following [GitHub trigger events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows) are fully supported and tested:

- `schedule`
- `push` (including new tags)
- `release`
- Manual run via `workflow_dispatch`

However, in practice, most triggers should work with the exception of
`pull_request`. If you would like support for `pull_request`, please tell us
about your use case on [issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a [new
issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The [container workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_container_slsa3.yml) accepts the following inputs:

Inputs:

| Name                | Required | Description                                                                                         |
| ------------------- | -------- | --------------------------------------------------------------------------------------------------- |
| `image`             | yes      | The OCI image name. This must not include a tag or digest.                                          |
| `digest`            | yes      | The OCI image digest. The image digest of the form '<algorithm>:<digest>' (e.g. 'sha256:abcdef...') |
| `registry-username` | yes      | Username to log into the container registry.                                                        |
| `compile-generator` | false    | Whether to build the generator from source. This increases build time by ~2m.                       |

Secrets:

| Name                | Required | Description                                |
| ------------------- | -------- | ------------------------------------------ |
| `registry-password` | yes      | Password to log in the container registry. |

### Provenance Format

The project generates SLSA provenance with the following values.

| Name                         | Value                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/generic@v1"` | Identifies a generic GitHub Actions build.                                                                                                                                                                             |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                             | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

### Provenance Example

The following is an example of the generated proveanance. Provenance is
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
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v1.1.1"
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

## Verification

Verification of provenance attestations can be done via several different tools.
This section shows examples of several popular tools.

## Kyverno

[Kyverno](https://kyverno.io/) is a policy management controller that can be
used to write Kubernetes-native policies for SLSA provenance. The following
assumes you have [installed Kyverno](https://kyverno.io/docs/installation/) in
your Kubernetes cluster.

The following Kyverno `ClusterPolicy` can be used to
[verify container images](https://kyverno.io/docs/writing-policies/verify-images/)
as part of [Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: check-slsa-attestations
spec:
  validationFailureAction: enforce
  webhookTimeoutSeconds: 30
  rules:
    - name: check-all-keyless
      match:
        any:
          - resources:
              kinds:
                - Pod
      verifyImages:
        # imageReferences sets which images the policy will apply to.
        # Replace with your image. Wildcard values are supported.
        - imageReferences:
            - "ghcr.io/ianlewis/actions-test:*"
          attestors:
            # This section declares which attestors are accepted. The subject
            # below corresponds to the OIDC identity of the container workflow.
            # The issuer corresponds to the GitHub OIDC server that issues the
            # identity.
            - entries:
                - keyless:
                    subject: "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v1.2.0"
                    issuer: "https://token.actions.githubusercontent.com"
          # This section declares some policy conditions acting on the provenance itself.
          attestations:
            - predicateType: https://slsa.dev/provenance/v0.2
              conditions:
                - all:
                    # This condition verifies that the image was generated from
                    # the source repository we expect. Replace this with your
                    # repository.
                    - key: "{{ invocation.configSource.uri }}"
                      operator: Equals
                      value: "git+https://github.com/ianlewis/actions-test@refs/tags/v0.0.11"

                    # This condition verifies the entrypoint of the workflow.
                    # Replace with the relative path to your workflow in your
                    #  repository.
                    - key: "{{ invocation.configSource.entryPoint }}"
                      operator: Equals
                      value: ".github/workflows/generic-container.yaml"

                    # This condition verifies that the builder is the builder we
                    # expect and trust. The following condition can be used
                    # unmodified. It verifies that the builder is the container
                    # workflow.
                    - key: "{{ regex_match('^https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v[0-9].[0-9].[0-9]$','{{ builder.id}}') }}"
                      operator: Equals
                      value: true
```

When applied the `ClusterPolicy` will be evaluated when a `Pod` is created. If
the `Pod` fulfills the policy conditions then the `Pod` can be created. If the
`Pod` does not fulfill one or more of the policy conditions then you will see an
error like the following. For example, the following error will occur when no
attestation for the image can be found.

```shell
$ kubectl apply -f pod.yaml
Error from server: error when creating "pod.yaml": admission webhook "mutate.kyverno.svc-fail" denied the request:

resource Pod/default/actions-test was blocked due to the following policies

check-slsa-attestations:
  check-all-keyless: |-
    failed to verify signature for ghcr.io/ianlewis/actions-test:v0.0.11: .attestors[0].entries[0].keyless: no matching attestations:
    no certificate found on attestation
```
