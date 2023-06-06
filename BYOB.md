# Build Your Own Builder (BYOB) Framework

<!-- toc -->

- [Design Overview](#design-overview)
  - [Project Workflow (PW)](#project-workflow-pw)
  - [Tool Repository](#tool-repository)
    - [Tool Reusable Workflow (TRW)](#tool-reusable-workflow-trw)
    - [Tool Callback Action (TCA)](#tool-callback-action-tca)
  - [SLSA Repository](#slsa-repository)
    - [SLSA Setup Action (SSA)](#slsa-repository-ssa)
    - [SLSA Reusable Workflow (SRW)](#slsa-reusable-workflow-srw)
- [Integration Steps](#integration-steps)
  - [TRW inputs](#trw-inputs)
    - [Inputs](#inputs)
    - [Secrets](#secrets)
    - [Outputs](#outputs)
    - [Important Notes](#important-notes)
  - [SRW Setup](#srw-setup)
  - [SRW Invocation](@srw-invocation)
  - [TRW Callback Action](#trw-callback-action)
    - [Inputs](#inputs-1)
    - [Outputs](#outputs-1)
    - [Invocation of Existing Action](#invocation-of-existing-action)
    - [Generation of Metadata Layout File](#generation-of-metadata-layout-file)
  - [Upload Attestations](#upload-attestations)
- [Other Delegators](#other-delegators)
- [Hardening](#hardening)

<!-- tocstop -->

# Design Overview

TODO(add diagram)

The diagram above depicts the different components of the BYOB framework.

## Project Workflow (PW)
On the left, the end-user project workflow (PW) is depicted. The PW is hosted in the repository that wants to build an artifact. As part of a build, the PW invokes the SLSA compliant builder defined by the TRW:

```yaml
- uses: npm/builder/.github/workflows/slsa.3.yml@v1.7.0
```

The example snippet shows the invocation of a builder with path '/.github/workflows/slsa.3.yml' from the GitHub's org "npm/builder" repository.

## Tool Repository

This is the tool repository hosting the builder invoked by PWs. The repository contains two components:

### Tool Reusable Workflow (TRW)
The "Tool Reusable Workflow" (TRW) is the SLSA compliant builder that wraps the original TA. End users' PWs invoke the TRW to build their artifacts. The TRW file must be created as part of the integration.

### Tool Callback Action (TCA)
The "Tool Callback Action" (TCA) is the Action that is invoked by the integration in an isolated GitHub job. The TCA is also hosted in the tool repository. The TCA's role is threefold:
Set the environment. For example, if the builder wants to build Go projects, the TCA would install the Go compiler. 
Call an Action that already exists in the repository. For example, if the builder wants to make the GoReleaser Action SLSA compliant, the TCA would call the existing GoReleaser Action after it has set up the environment.
Output attestation metadata (name, binaries and hashes) that are used by the framework.

## SLSA GitHub Repository
The [slsa-github-generator](todo:link) repository hosts the code for the BYOB framework maintained by the OpenSSF SLSA tooling team.

### SLSA Setup Action (SSA)
This [Action](todo:link-to-action) is used to initialize the BYOB framework. It returns a so-called "SLSA token" which is a required input to the SRW invocation:
```yaml
- uses: slsa-framework/slsa-github-generator/actions/delegator/setup-generic@v1.7.0
...
- uses: slsa-framework/slsa-github-generator/.github/workflow/delegator_generic_slsa3.yml@v1.7.0
 with:
   - slsa-token: <output-of-setup>
```

### SLSA Reusable Workflow (SRW)
The SRW acts as the build's orchestrator. It calls the TCA, invokes internal components to generate provenance, and returns the provenance to its TRW caller. A TRW would typically call the SRW as follows:

```yaml
- uses: slsa-framework/slsa-github-generator/.github/workflow/delegator_generic_slsa3.yml@v1.2.3
```

# Integration Steps

In this example, we will assume there is an existing [GitHub Action](https://github.com/laurentsimon/byob-doc/blob/main/action.yml) which builds an artifact. The builder is fairly simple: it just echos the parameters into the artifact. It also takes a username and password to retrieve / push information from a remote registry. The full action.yml is available [here](https://github.com/laurentsimon/byob-doc/blob/main/action.yml).

## TRW inputs
The first step for our integration is to create our TRW file and define its inputs. The inputs should mirror those of the existing Action.

### Inputs
Inputs that have low entropy are defined under the [inputs section](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L25-L39). Unless Action inputs, you must define the type ([boolean, number, or string](https://docs.github.com/en/actions/using-workflows/reusing-workflows)) of each input. You may also provide a default value. The inputs will be attested to in the generated provenance. We will in [Section xxx](todo:link) how to redact certain inputs from the provenance, such as the username that may be considered sensitive. 

We also declare an additional [rekor-log-public](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L43-L47) boolean input. Given that the name of the repository will be available in the provenance and will be uploaded to the public transparency log, we need users to acknowledge that they are aware that private repository names will be made public. We encourage all TRWs to define this option. For public repositories, the value of the input is set to true by default by the SRW. For private repositories, users must set if to true when calling the TRW.

### Secrets
Unlike Actions, secrets are defined under a separate [secrets section](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L49-L57). Secret should only be high-entropy values. Do not set username or other low-entropy PII as secrets, as it may intermittently fail due to this [unresolved GitHub issues](https://github.com/orgs/community/discussions/37942). Secrets may be marked as optional. Unlike for Actions, secrets cannot have default values: inn our example, the 'token' secret has no default value, whereas the original Action [has one](https://github.com/laurentsimon/byob-doc/blob/main/action.yml#L33). We will see in [Section xxx](todo:link) how to set default values in the TCA. 

### Outputs
The outputs from the TCA may be returned to the PW as well. To do this, use the [outputs section](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L61-L67). There are [other outputs set](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L69-L88) as well: Those provide metadata about the built artifacts and their provenance, and we will discuss them in [Section xxx](todo:link).

### Important Notes
One key difference between the Action and reusable workflow is isolation. The SRW runs on a different VM than the TRW; and the TRW runs on a different VM from the PW. This means that the artifact built by the TCA (which is managed by the SRW) is not accessible directly by the TRW. The SRW needs to share these files with the TRW; which may also share them with the PW. We will see in the following sections how to do that. The [TRW outputs](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L69-L88) provides the metadata necessary to download these files, and we will discuss them in [Section xxx](todo:link).

## SRW Setup
To initialize the SRW framework, you need to invoke a SLSA Setup Action (SSA). These Actions are declared under the [SLSA repo's actions/delegator/setup-*](https://github.com/slsa-framework/slsa-github-generator/tree/main/actions/delegator) folder. For our example, we will use the [setup-generic Action](https://github.com/slsa-framework/slsa-github-generator/tree/main/actions/delegator/setup-generic). The [relevant code](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L91-L107) calls the SSA as follows:

```yaml
uses: slsa-framework/slsa-github-generator/actions/delegator/setup-generic@v1.7.0
  with:
    slsa-workflow-recipient: "delegator_generic_slsa3.yml"
    slsa-rekor-log-public: ${{ inputs.rekor-log-public }}
    slsa-runner-label: "ubuntu-latest"
    slsa-build-action-path: "./internal/callback_action"
    slsa-workflow-inputs: ${{ toJson(inputs) }}
    slsa-workflow-masked-inputs: username
```

Let's go through the parameters:
 - slsa-workflow-recipient is the name of the SRW we are initializing for. This is the workflow that we will call to run the build in our example.
- slsa-rekor-log-public is simply the same as the TRW 'slsa-rekor-log-public' input ,so we just set the value with the TW's corresponding value.
- slsa-runner-label is the runner label to run the build on. We currently only support ubuntu runners, but we will add support for other runners in the future.
- slsa-build-action-path is the path to the TRW Callback Action (TCA), relative to the root of the repository.
- slsa-workflow-inputs are the inputs to the TRW, which the provenance will attest to. These inputs are also relayed to the TCA.
- slsa-workflow-masked-inputs is a list of comma separated field names that are redacted from the generated SLSA provenance. In this example, we're telling the TRW that the username input should be redacted.

## SRW Invocation
Once we have initialize the SRW, we [call the SRW](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L109-L122):

```yaml
slsa-run:
    needs: [slsa-setup]
    permissions:
      id-token: write # For signing.
      contents: read # For asset uploads.
      actions: read # For the entrypoint.
    uses: slsa-framework/slsa-github-generator/.github/workflows/delegator_generic_slsa3.yml@v1.7.0
    with:
      slsa-token: ${{ needs.slsa-setup.outputs.slsa-token }}
    secrets:
      secret1: ${{ inputs.password }}
      secret2: ${{ inputs.token }}
```

In addition to the token, we also [provide the secrets](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L118-L120). Up to 15 secrets are supported.

## TRW Callback Action
The call above will run the SRW and invoke the callback Action, so let's see how to define it now. The Action ode is available under [internal/callback_action](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/).

### Inputs
The inputs to the [TCA are pre-defined](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/action.yml#L6-L14), so you just have to follow their definition:
slsa-workflow-inputs contains a JSON object with a list of key-value pairs for the inputs provided by the [TRW to the SSA during initialization ](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L106). We will see shortly how to use these values.
slsa-layout-file: a path to a file containing the layout to generate the attestation. This is an expected output. We will see shortly how the format for this file.
slsa-workflow-secretX, where X from '1' to '15'. These contain the secrets that the TRW [provided to the SRW during invocation](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L118-L120). Unused secrets [should be defined as unused](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/action.yml#L26-L39). 

### Outputs
We [declare the same outputs](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/action.yml#L41-L47) as the existing Actions. These outputs are available to the TRW and [may be returned by the TRW to the PW](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L61-L67), as is the case in our TRW.

### Invocation of Existing Action
We [invoke the existing Action](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/action.yml#L57-L65) as a local Action and pass it the inputs by extracting them from the slsa-workflow-inputs argument:

```yaml
uses: ./../__TOOL_CHECKOUT_DIR__
id: build
  with:
    artifact: ${{ fromJson(inputs.slsa-workflow-inputs).artifact }}
    content: ${{ fromJson(inputs.slsa-workflow-inputs).content }}
    username: ${{ fromJson(inputs.slsa-workflow-inputs).username }}
    password: ${{ inputs.slsa-workflow-secret1 }}
    token: ${{ inputs.slsa-workflow-secret2 || github.token }}
```

Note that the `./../__TOOL_CHECKOUT_DIR__` is the path where the TRW repository is checked out by the framework, so it's accessible locally.
Notice how we populate the token field: If the user has not passed a value to `inputs.slsa-workflow-secret2`, we default to using the GitHub token `github.token`.

### Generation of Metadata Layout File
The last thing to do in the TCA is to [generate the metadata file](https://github.com/laurentsimon/byob-doc/blob/main/internal/callback_action/action.yml#L67-L73) to indicate to the BYOB platform which files to attest to, and which attestations to generate. You can ask the platform to generate several attestations, each attestation to one or more generated files. The snippet below shows the generation of a single attestation attestation to a single file:

```json
{
  "version": 1,
  "attestations": [
    {
      "name": "<artifact>.intoto",
      "subjects": [
        {
          "name": "<artifact>",
          "digest": {
            "sha256": "c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4"
          }
        }
      ]
    }
  ]
}
```

## Upload Attestations
In a final ["publish" job](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L110-L141), we may [download the attestations](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L117C19-L121) and do whatever we'd like with them. We may [upload them as release assets to a GitHub release](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L123-L141), or upload them to a registry, etc. 

Also think about [returning the attestation to the PW](https://github.com/laurentsimon/byob-doc/blob/main/.github/workflows/builder_example_slsa3.yml#L69-L75) in case they want to use them themselves. If you return the attestations as output, we encourage you to create a 'secure-download-attestation' Action for your users, e.g. under a download folder in your repository. See an example [here](https://github.com/laurentsimon/byob-doc/blob/main/download/attestation/action.yml).

# Other Delegators
TODO(Non-low perms)

# Hardening
TODO(Scorecard, pin actions and dependencies, verify provenance when downloading binary, etc, Use low-perm)
