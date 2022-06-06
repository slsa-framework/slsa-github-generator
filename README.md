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

Each of the projects utilize the same base GitHub workflow SLSA provenance. The common fields of the SLSA provenance predicate attested to are below.


_Added v1.0.0_

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

`Invocation.environment`: This describes the GitHub workflow builder-controlled environment variables, including the event information, required to reproduce the build. See `github` content [documentation](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) for more information.


  | Name       | Value    | Description    |
| ---------------------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `github_event_name`   | `workflow_dispatch` | Name of the [event](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#available-events) that initiated the workflow run. |
| `github_event_payload` | `"{"inputs": null,"ref": "refs/heads/main", "repository": { ... }}"`    | The full event payload, including workflow inputs and repository information. |
| `github_ref_type` | `"branch"`    | The type of ref that triggered the workflow run. |
| `github_ref` | `"refs/pull/169/merge"`    | The ref that triggered the workflow run. |
| `github_base_ref` | `"feat/feat-branch"`    | The ref or source branch of the pull request in a workflow run. Only populated on pull requests. |
| `github_head_ref` | `"feat/feat-branch"`    | The is ref or source branch of the pull request in a workflow run. |
| `github_actor` | `"laurentsimon"`    | The username of the user that initiated the workflow run. |
| `github_sha1` | `"b54fb2ec8807a93b58d5f298b7e6b785ea7078bb"`    | The is the commit SHA that triggered the workflow run. |
| `github_repository_owner` | `"slsa-framework"`    | The owner of the repository. |
| `github_repository_id` | `"8923542"`    | The is the unique ID of the repository. |
| `github_actor_id` | `"branch"`    | The is the unique ID of the actor that triggered the workflow run. |
| `github_repository_owner_id` | `"123456"`    | The is the unique ID of the owner of the repository. |
| `github_run_attempt` | `"1"`    | The is run attempt of the workflow run. |
| `github_run_id` | `"2436960022"`    | The is the run ID for the workflow run. |
| `github_run_number` | `"32"`    | The is the run number of the workflow run. |

```json
"environment": {
    "github_actor": "laurentsimon",
    "github_base_ref": "",
    "github_event_name": "workflow_dispatch",
    "github_event_payload": {
      "inputs": null,
      "ref": "refs/heads/main",
      "repository": { ... }
    },
    "github_head_ref": "add-field-docs",
    "github_ref": "refs/pull/169/merge",
    "github_ref_type": "branch",
    "github_repository_owner": "slsa-framework",
    "github_run_attempt": "1",
    "github_run_id": "2436960022",
    "github_run_number": "32",
    "github_sha1": "b54fb2ec8807a93b58d5f298b7e6b785ea7078bb",
    "github_repository_id": "8923542",
    "github_repository_owner": "ianlewis",
    "github_repository_owner_id": "123456",
}
```

`BuildConfig`: This contains information on the steps of the build. The default is nil, specific builders implement their own. See [BuildConfig Format](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/go/README.md#buildconfig-format) in the Golang builder for an example.

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



