# detect-workflow-js

detect-workflow-js is a GitHub Action that detects the repository and ref (the
branch, tag, or commit SHA) for a reusable workflow. This is useful in order to
check out code from the reusable workflow's repository at the git ref that the
user workflow specifies.

detect-workflow-js retrieves the [current workflow run](https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#get-a-workflow-run) and inspects the `referenced_workflows` to determine the reusable workflow's repository.

## Why?

When a GitHub Actions reusable workflow is executed, the workflow only has
access to the executing YAML itself. If any other code is required to execute
the workflow it needs to be checked out from the reusable workflow's repository
at the ref that was specified by the user. A call to a reusable workflow might
look like this:

```yaml
myjob:
  permissions:
    id-token: write
    contents: read
    actions: read
  # {owner}/{repository}{/path}@{ref}
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@<tag>
  with:
    subjects: "${{ needs.build.outputs.digest }}"
```

However, it is not trivial to determine the repository and ref because the
[GitHub Actions
contexts](https://docs.github.com/en/actions/learn-github-actions/contexts)
refer to the user workflow that called the reusable workflow. The reusable
workflow should also ideally avoid hard-coding this information as the
repository could be a fork and the ref used is unpredictable.

## Usage

To use detect-workflow-js, add a job to your reusable workflow with `repository`, `ref`, and `workflow` as an output. You can then use the output from the job in later jobs
in the workflow.

```yaml
jobs:
  detect-workflow:
    runs-on: ubuntu-latest
    permissions:
      id-token: write # Needed to detect the current reusable repository and ref.
      contents: read
    outputs:
      repository: ${{ steps.detect.outputs.repository }}
      ref: ${{ steps.detect.outputs.ref }}
    steps:
      - name: Detect the repository and ref
        id: detect
        uses: slsa-framework/slsa-github-generator/.github/actions/detect-workflow-js@<git hash>

  # example of using the output from detect-workflow
  print:
    shell: bash
    env:
      REPO: "${{ steps.detect-workflow.outputs.repository }}"
      REF: "${{ steps.detect-workflow.outputs.ref }}"
      WORKFLOW: "${{ steps.detect-workflow.outputs.workflow }}"
    run: |
      echo $REPO
      echo $REF
      echo $WORKFLOW
```

In the example above, `REPO`, `WORKFLOW` and `REF` will be equal to the
repository, workflow path, and ref from the user workflow's call to the
reusable workflow.

## Outputs

| Name         | Description                                                                    |
| ------------ | ------------------------------------------------------------------------------ |
| `repository` | The repository of the reusable workflow (`{owner}/{repository name}`)          |
| `ref`        | The ref (branch, tag, or commit SHA) specified by the user.                    |
| `workflow`   | The workflow path, relative to the `repository` (`.github/workflows/test.yml`) |
