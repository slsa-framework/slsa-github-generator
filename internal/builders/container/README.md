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
  - [Ko](#ko)
  - [GoReleaser](#goreleaser)
- [Provenance for matrix strategy builds](#provenance-for-matrix-strategy-builds)
- [Verification](#verification)
  - [slsa-verifier](#slsa-verifier)
  - [Cosign](#cosign)
  - [Sigstore policy-controller](#sigstore-policy-controller)
  - [Kyverno](#kyverno)
- [Known Issues](#known-issues)
  - [`packages: write` permission required even if not using ghcr.io](#packages-write-permission-required-even-if-not-using-ghcrio)

<!-- tocstop -->

---

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
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.0.0
  with:
    image: ${{ needs.build.outputs.image }}
    digest: ${{ needs.build.outputs.digest }}
    registry-username: ${{ github.actor }}
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
          echo "image=$image_name" >> "$GITHUB_OUTPUT"

  # This step calls the container workflow to generate provenance and push it to
  # the container registry.
  provenance:
    needs: [build]
    permissions:
      actions: read # for detecting the Github Actions environment.
      id-token: write # for creating OIDC tokens for signing.
      packages: write # for uploading attestations.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.0.0
    with:
      image: ${{ needs.build.outputs.image }}
      digest: ${{ needs.build.outputs.digest }}
      registry-username: ${{ github.actor }}
    secrets:
      registry-password: ${{ secrets.GITHUB_TOKEN }}
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
you have an issue in all other triggers please submit a [new
issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The [container workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_container_slsa3.yml) accepts the following inputs:

Inputs:

| Name                             | Description                                                                                                                                                                                                                                                                             |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `image`                          | **(Required)** The OCI image name. This must not include a tag or digest.                                                                                                                                                                                                               |
| `digest`                         | **(Required)** The OCI image digest. The image digest of the form '<algorithm>:<digest>' (e.g. 'sha256:abcdef...')                                                                                                                                                                      |
| `registry-username`              | Username to log in the container registry. Either `registry-username` input or `registry-username` secret is required.                                                                                                                                                                  |
| `compile-generator`              | Whether to build the generator from source. This increases build time by ~2m.<br>Default: `false`.                                                                                                                                                                                      |
| `private-repository`             | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false`                                     |
| `continue-on-error`              | Set to true to ignore errors. This option is useful if you won't want a failure to fail your entire workflow.<br>Default: `false`                                                                                                                                                       |
| `gcp-workload-identity-provider` | The full identifier of the Workload Identity Provider, including the project number, pool name, and provider name. If provided, this must be the full identifier which includes all parts:<br>`projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider` |
| `gcp-service-account`            | Email address or unique identifier of the Google Cloud service account for which to generate credentials. For example:<br>`my-service-account@my-project.iam.gserviceaccount.com`                                                                                                       |
| `provenance-registry-username`   | Username when publishing to provenance registry (option 'provenance-registry') instead of image registry. Either `provenance-registry-username` input or `provenance-registry-username` secret is required.                                                                                                                                                                                                                 |
| `provenance-registry`            | If set, provenance is pushed to this registry instead of image registry. (e.g. `gcr.io/my-new-repo`)                                                                                          |

Secrets:

| Name                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `image`                        | The OCI image name. This must not include a tag or digest. Either `image` input or `image` secret is **required**. Secret `image` value takes precedence on `image` input value. Should be used in scenarios when the image name contains secret values, and therefore can't be provided directly. Use case - an undisclosed private registry use.                                                                                                                                             |
| `registry-username`            | Username to log in the container registry. Either `registry-username` input or `registry-username` secret is required. This should only be used for high entropy values such as AWS Access Key as described [here](https://github.com/docker/login-action#aws-elastic-container-registry-ecr). Normal username values could match other input values and cause them to be ignored by GitHub Actions and causing your build to fail. In those cases, use the `registry-username` input instead. |
| `registry-password`            | **(Required)** Password to log in the container registry.                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `provenance-registry-username` | Username when publishing to provenance registry (option 'provenance-registry') instead of image registry. Either `provenance-registry-username` input or `provenance-registry-username` secret is required. This should only be used for high entropy values such as AWS Access Key as described [here](https://github.com/docker/login-action#aws-elastic-container-registry-ecr). Normal username values could match other input values and cause them to be ignored by GitHub Actions and causing your build to fail. In those cases, use the registry-username input instead.                                                                                                                                                                                            |
| `provenance-registry-password` | Password when publishing to provenance registry instead of image registry.                                               |

### Workflow Outputs

The [container workflow](https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/workflows/generator_container_slsa3.yml) accepts the following outputs:

| Name      | Description                                                                                     |
| --------- | ----------------------------------------------------------------------------------------------- |
| `outcome` | If `continue-on-error` is `true`, will contain the outcome of the run (`success` or `failure`). |

### Provenance Format

The project generates SLSA provenance with the following values.

| Name                         | Value                                                                    | Description                                                                                                                                                                                                            |
| ---------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/container@v1"` | Identifies a the GitHub Actions build.                                                                                                                                                                                 |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                               | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

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
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v1.4.0"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator/container@v1",
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

### Ko

[ko](https://github.com/ko-build/ko) is a simple, fast container image builder for Go applications. If you want to use `ko` you can generate SLSA3 provenance by updating your workflow withe following steps:

1. Declare an `outputs` for the build job:

   ```yaml
   jobs:
     build:
       outputs:
         image: ${{ steps.build.outputs.image }}
         digest: ${{ steps.build.outputs.digest }}
   ```

2. Add an `id: build` field to your ko step. Update the step to output the image
   and digest.

   ```yaml
   steps:
     [...]
     - name: Run ko
       id: build
       env:
         KO_DOCKER_REPO: "${{ env.IMAGE_REGISTRY }}/${{ env.IMAGE_NAME }}"
         KO_USER: ${{ github.actor }}
         KO_PASSWORD: ${{ secrets.GITHUB_TOKEN }}
         GIT_REF: ${{ github.ref }}
       run: |
         # get tag name without tags/refs/ prefix.
         tag=$(echo ${GIT_REF} | cut -d'/' -f3)

         # Log into regisry
         echo "${KO_PASSWORD}" | ko login ghcr.io --username "$KO_USER" --password-stdin

         # Build & push the image. Save the image name.
         ko build --bare --tags="${tag}" --image-refs .digest

         # Output the image name and digest so we can generate provenance.
         image=$(cat .digest | cut -d'@' -f1 | cut -d':' -f1)
         digest=$(cat .digest| cut -d'@' -f2)
         echo "image=$image" >> "$GITHUB_OUTPUT"
         echo "digest=$digest" >> "$GITHUB_OUTPUT"
   ```

3. Call the generic container workflow to generate provenance by declaring the job below:

   ```yaml
   provenance:
     needs: [build]
     permissions:
       actions: read
       id-token: write
       # contents: read
       packages: write
     if: startsWith(github.ref, 'refs/tags/')
     uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.0.0
     with:
       image: ${{ needs.build.outputs.image }}
       digest: ${{ needs.build.outputs.digest }}
       registry-username: ${{ github.actor }}
       compile-generator: true
     secrets:
       registry-password: ${{ secrets.GITHUB_TOKEN }}
   ```

   All together, it will look as the following:

   ```yaml
   jobs:
     build:
       permissions:
         contents: read
         packages: write
       outputs:
         image: ${{ steps.build.outputs.image }}
         digest: ${{ steps.build.outputs.digest }}
       runs-on: ubuntu-latest
       steps:
         - name: Checkout the repository
           uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # v2.3.4

         - uses: actions/setup-go@v5.0.1
           with:
             go-version: 1.19

         - name: Set up ko
           uses: imjasonh/setup-ko@v0.6

         - name: Run ko
           id: build
           env:
             KO_DOCKER_REPO: "${{ env.IMAGE_REGISTRY }}/${{ env.IMAGE_NAME }}"
             KO_USER: ${{ github.actor }}
             KO_PASSWORD: ${{ secrets.GITHUB_TOKEN }}
             GIT_REF: ${{ github.ref }}
           run: |
             # get tag name without tags/refs/ prefix.
             tag=$(echo ${GIT_REF} | cut -d'/' -f3)

             # Log into regisry
             echo "${KO_PASSWORD}" | ko login ghcr.io --username "$KO_USER" --password-stdin

             # Build & push the image. Save the image name.
             image_and_digest=$(ko build --tags="${tag}" .)

             # Output the image name and digest so we can generate provenance.
             image=$(echo "${image_and_digest}" | cut -d':' -f1)
             digest=$(echo "${image_and_digest}" | cut -d'@' -f2)
             echo "image=$image" >> "$GITHUB_OUTPUT"
             echo "digest=$digest" >> "$GITHUB_OUTPUT"

     # This step calls the generic workflow to generate provenance.
     provenance:
       needs: [build]
       permissions:
         actions: read
         id-token: write
         # contents: read
         packages: write
       if: startsWith(github.ref, 'refs/tags/')
       uses: slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@v2.0.0
       with:
         image: ${{ needs.build.outputs.image }}
         digest: ${{ needs.build.outputs.digest }}
         registry-username: ${{ github.actor }}
         compile-generator: true
       secrets:
         registry-password: ${{ secrets.GITHUB_TOKEN }}
   ```

### [GoReleaser](#goreleaser)

Follow the great blog post of [goreleaser.com](https://goreleaser.com/blog/slsa-generation-for-your-artifacts).

## Provenance for matrix strategy builds

See the
[equivalent section](../generic/README.md#provenance-for-matrix-strategy-builds)
for the generic generator.

## Verification

Verification of provenance attestations can be done via several different tools. This section shows examples of several popular tools.

### slsa-verifier

`slsa-verifier` can be used to verify the provenance attestation for the image. Please see the [documentation](https://github.com/slsa-framework/slsa-verifier#containers) in the slsa-verifier repository.

### Cosign

[Cosign](https://docs.sigstore.dev/cosign/overview/) can be used to verify the provenance attestation for the image. A [CUE](https://cuelang.org/) policy can also be used to verify parts of the SLSA attestation.

Here is an example policy stored in `policy.cue`:

```text
// The predicateType field must match this string
predicateType: "https://slsa.dev/provenance/v0.2"

predicate: {
  // This condition verifies that the builder is the builder we
  // expect and trust. The following condition can be used
  // unmodified. It verifies that the builder is the container
  // workflow.
  builder: {
    id: =~"^https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v[0-9]+.[0-9]+.[0-9]+$"
  }
  invocation: {
    configSource: {
      // This condition verifies the entrypoint of the workflow.
      // Replace with the relative path to your workflow in your
      // repository.
      entryPoint: ".github/workflows/generic-container.yml"

      // This condition verifies that the image was generated from
      // the source repository we expect. Replace this with your
      // repository.
      uri: =~"^git\\+https://github.com/ianlewis/actions-test@refs/tags/v[0-9]+.[0-9]+.[0-9]+$"
    }
  }
}
```

We can then use `cosign` to verify the attestation using the policy.

```shell
COSIGN_EXPERIMENTAL=1 cosign verify-attestation \
  --type slsaprovenance \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v[0-9]+.[0-9]+.[0-9]+$' \
  --policy policy.cue \
  ghcr.io/ianlewis/actions-test:v0.0.79
```

This should result in output like the following:

```text
will be validating against CUE policies: [policy.cue]

Verification for ghcr.io/ianlewis/actions-test:v0.0.79 --
The following checks were performed on each of these signatures:
  - The cosign claims were validated
  - Existence of the claims in the transparency log was verified offline
  - The code-signing certificate was verified using trusted certificate authority certificates
Certificate subject:  https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v1.4.0
Certificate issuer URL:  https://token.actions.githubusercontent.com
GitHub Workflow Trigger: push
GitHub Workflow SHA: 3f938aae461d2a8bc7897ff975e77a876e3d9123
GitHub Workflow Name: Generic container
GitHub Workflow Trigger ianlewis/actions-test
GitHub Workflow Ref: refs/tags/v0.0.79
{"payloadType":"application/vnd.in-toto+json","payload":"...","signatures":[{"keyid":"","sig":"..."}]}
```

You can read more in the [cosign documentation](https://docs.sigstore.dev/cosign/attestation/).

### Sigstore policy-controller

Sigstore [policy-controller](https://docs.sigstore.dev/policy-controller/overview)
is a policy management controller that can be used to write Kubernetes-native
policies for SLSA provenance. The following assumes you have
[installed policy-controller](https://docs.sigstore.dev/policy-controller/installation)
in your Kubernetes cluster.

The following `ClusterImagePolicy` can be used to verify container images
as part of [Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

```yaml
apiVersion: policy.sigstore.dev/v1alpha1
kind: ClusterImagePolicy
metadata:
  name: image-is-signed-by-github-actions
spec:
  images:
    # Matches all versions of the actions-test image.
    # NOTE: policy-controller mutates pods to use a digest even if originally
    # specified by tag.
    - glob: "ghcr.io/ianlewis/actions-test@*"
  authorities:
    - keyless:
        # Signed by the public Fulcio certificate authority
        url: https://fulcio.sigstore.dev
        identities:
          # Matches the Github Actions OIDC issuer
          - issuer: https://token.actions.githubusercontent.com
            # Matches the reusable workflow's signing identity.
            subjectRegExp: "^https://github.com/ianlewis/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v[0-9]+.[0-9]+.[0-9]+$"
      attestations:
        - name: must-have-slsa
          predicateType: slsaprovenance
          policy:
            type: cue
            data: |
              // The predicateType field must match this string
              predicateType: "https://slsa.dev/provenance/v0.2"

              predicate: {
                // This condition verifies that the builder is the builder we
                // expect and trust. The following condition can be used
                // unmodified. It verifies that the builder is the container
                // workflow.
                builder: {
                  id: =~"^https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v[0-9]+.[0-9]+.[0-9]+$"
                }
                invocation: {
                  configSource: {
                    // This condition verifies the entrypoint of the workflow.
                    // Replace with the relative path to your workflow in your
                    // repository.
                    entryPoint: ".github/workflows/generic-container.yml"

                    // This condition verifies that the image was generated from
                    // the source repository we expect. Replace this with your
                    // repository.
                    uri: =~"^git\\+https://github.com/ianlewis/actions-test@refs/tags/v[0-9]+.[0-9]+.[0-9]+$"
                  }
                }
              }
```

When applied the `ClusterImagePolicy` will be evaluated when a `Pod` is
created. If the `Pod` fulfills the policy conditions then the `Pod` can be
created. If the `Pod` does not fulfill one or more of the policy conditions
then you will see an error like the following. For example, the following error
will occur when issuer does not match.

```shell
$ kubectl run actions-test --image=ghcr.io/ianlewis/actions-test:v0.0.38 --port=8080
Error from server (BadRequest): admission webhook "policy.sigstore.dev" denied the request: validation failed: failed policy: image-is-signed-by-github-actions: spec.containers[0].image
ghcr.io/ianlewis/actions-test@sha256:7c01e1c050f6b7a9b38a53da1be0835288da538d506de571f654417ae89aea4e attestation keyless validation failed for authority authority-0 for ghcr.io/ianlewis/actions-test@sha256:7c01e1c050f6b7a9b38a53da1be0835288da538d506de571f654417ae89aea4e: no matching attestations:
none of the expected identities matched what was in the certificate
```

This behavior can be configured to `allow`, `deny`, or `warn` depending on your
use case. See the [sigstore
docs](https://docs.sigstore.dev/policy-controller/overview/#admission-of-images)
for more info.

### Kyverno

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
                    subject: "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/v1.4.0"
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

## Known Issues

### `packages: write` permission required even if not using ghcr.io

Due to limitations in how GitHub actions manages permissions on ephemeral tokens
in reusable workflows, and how cosign uses available credentials, the container
workflow always requires `packages: write`.

Please see #1257 for details.
