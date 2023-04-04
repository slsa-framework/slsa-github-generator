# Provenance Format

The provenance format for each workflow conforms to the [SLSA v0.2 provenance
format](https://slsa.dev/provenance/v0.2). Each of the projects populates the
provenance with the same base information based on GitHub Actions. The common
fields of the SLSA provenance predicate attested to are below.

## Common Fields

### Builder

`Builder.Id`: The builder ID refers to the fully qualified URI of the trusted builder's workflow and its reference.

```json
"builder": {
  "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@refs/tags/v0.0.1"
}
```

### Build Type

`buildType`: This is the URI for the particular provenance builder. For example, this is the `buildType` for the Go builder:

```json
"buildType": "https://github.com/slsa-framework/slsa-github-generator/go@v1"
```

In SLSA the [buildType](https://slsa.dev/provenance/v0.2#buildType) defines how
many other fields are formatted including all of those below.

### Invocation

`invocation`: Identifies the event that kicked off the build. This describes the workflow run and includes GitHub workflow event information, entrypoint, and parameters from trigger events.

`invocation.configSource`: This describes the calling workflow's source (`uri` and `digest`) and the `entryPoint` of the build.

```json
"configSource": {
    "uri": "git+https://github.com/laurentsimon/slsa-verifier-test-gen@refs/heads/main",
    "digest": {
      "sha1": "15bf79ea9c89fffbf5dd02c6b5b686b291bfcbd2"
    },
    "entryPoint": ".github/workflows/release.yml"
},
```

`invocation.parameters`: This describes any parameters from trigger events.

`invocation.environment`: This describes the GitHub workflow builder-controlled environment variables, including the event information, required to reproduce the build. See `github` content [documentation](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) for more information.

| Name                         | Value                                        | Description                                                                                                                                                                     |
| ---------------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `github_event_name`          | `workflow_dispatch`, `schedule`, `push`, etc | Name of the [event](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#available-events) that initiated the workflow run.                         |
| `github_event_payload`       | `"{"inputs": null, "repository": { ... }}"`  | The full [event payload](https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads), including workflow inputs and repository information. |
| `github_ref_type`            | `"branch"`                                   | The type of ref that triggered the workflow run.                                                                                                                                |
| `github_ref`                 | `"refs/heads/main"`                          | The ref that triggered the workflow run.                                                                                                                                        |
| `github_base_ref`            | `"feat/feat-branch"`                         | The ref or source branch of the pull request in a workflow run. Only populated on pull requests.                                                                                |
| `github_head_ref`            | `"feat/feat-branch"`                         | The is ref or source branch of the pull request in a workflow run.                                                                                                              |
| `github_actor`               | `"laurentsimon"`                             | The username of the user that initiated the workflow run.                                                                                                                       |
| `github_sha1`                | `"b54fb2ec8807a93b58d5f298b7e6b785ea7078bb"` | The is the commit SHA that triggered the workflow run.                                                                                                                          |
| `github_repository_owner`    | `"slsa-framework"`                           | The owner of the repository.                                                                                                                                                    |
| `github_repository_id`       | `"8923542"`                                  | The is the unique ID of the repository.                                                                                                                                         |
| `github_actor_id`            | `"973615"`                                   | The is the unique ID of the actor that triggered the workflow run.                                                                                                              |
| `github_repository_owner_id` | `"123456"`                                   | The is the unique ID of the owner of the repository.                                                                                                                            |
| `github_run_attempt`         | `"1"`                                        | The is run attempt of the workflow run.                                                                                                                                         |
| `github_run_id`              | `"2436960022"`                               | The is the run ID for the workflow run.                                                                                                                                         |
| `github_run_number`          | `"32"`                                       | The is the run number of the workflow run.                                                                                                                                      |

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

### Build Config

`buildConfig`: This contains information on the steps of the build. The default is nil, specific builders implement their own. See:

- [Golang BuildConfig Format](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/go/README.md#buildconfig-format).

### Materials

`materials`: List of materials that influenced the build, including the repository that triggered the GitHub Actions workflow.

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

### Metadata

`metadata`: Other properties of the build, including `BuildInvocationID`, which is the unique `RunID` and `RunAttempt` separated by a '-'.

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

Specific implementations of builders may define and customize their own `buildConfig`, `invocation`, `materials`, and `metadata`.
