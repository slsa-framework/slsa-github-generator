# Generation of SLSA3+ provenance for native GitHub projects

This repository contains the implementation for generating non-forgeable [SLSA provenance](https://slsa.dev/) that meets the requirement for the [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels). The SLSA generation makes use of a [Github Actions reusable workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows).

This repository contains the code, examples and technical design for the blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

---

- [Golang projects](#golang-projects)
- [Other projects](#other-projects)
- [Technical design](#technical-design)
  - [Blog post](#blog-post)
  - [Specifications](#specifications)

---

## Golang projects

To generate SLSA provenance for your Golang project, follow [internal/builders/go/README.md](internal/builders/go/README.md).

## Other projects

To generate SLSA provenance for other programming languages, follow [internal/builders/generic/README.md](internal/builders/generic/README.md).

## Technical design

### Blog post

Find our blog post series [here](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

### Specifications

For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).

### Format

Each of the projects utilize the same base GitHub workflow SLSA provenance. The common fields of the SLSA provenance predicate attested to are:



`BuildType`: This is the URI for the particular provenance builder, for example, the go or generic builder. 
```json
  "buildType": "https://github.com/slsa-framework/slsa-github-generator-go@v1"
```
`Builder.Id`: The builder ID refers to the fully qualified URI of the trusted builder's workflow and its reference. 
```json
 "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@refs/tags/v0.0.1"
  }
```
`Invocation`: Identifies the event that kicked off the build. This describes the workflow run and includes GitHub workflow event information, entrypoint, and parameters from trigger events. 
`Invocation.configSource`: This describes the calling workflow's source and the entrypoint of the build.
```json
      "configSource": {
        "uri": "git+https://github.com/laurentsimon/slsa-verifier-test-gen@refs/heads/main",
        "digest": {
          "sha1": "15bf79ea9c89fffbf5dd02c6b5b686b291bfcbd2"
        },
        "entryPoint": "Go SLSA Release"
      },
```
`Invocation.parameters`: This describes any parameters from trigger events.

`Invocation.environment`: This describes the GitHub workflow environment, including the event information.
```json
"environment": {
        "arch": "X64",
        "github_actor": "laurentsimon",
        "github_base_ref": "",
        "github_event_name": "workflow_dispatch",
        "github_event_payload": {
          "inputs": null,
          "ref": "refs/heads/main",
          "repository": { ... }
        },
}
```

`BuildConfig`: This contains information on the steps of the build. The default is nil, specific builders implement their own.

`Materials`: List of materials that influenced the build, including the repository that triggered the GitHub Actions workflow.
```json
    "materials": [
      {
        "uri": "git+https://github.com/laurentsimon/slsa-verifier-test-gen@refs/heads/main",
        "digest": {
          "sha1": "15bf79ea9c89fffbf5dd02c6b5b686b291bfcbd2"
        }
      },
      {
        "uri": "https://github.com/actions/virtual-environments/releases/tag/ubuntu20/20220515.1"
      }
    ]
```
`Metadata`: Other properties of the build, including `BuildInvocationID` as the unique `RunID` and `RunAttempt`. 
```json
"metadata": {
      "buildInvocationID": "2387611653-1",
      "completeness": {
        "parameters": true,
        "environment": false,
        "materials": false
      },
      "reproducible": false
    },

```


Specific implementations of builders may define and customize their own `BuildConfig`, `Invocation`, `Materials`, and `Metadata`.



