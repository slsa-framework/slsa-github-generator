# detect-workflow

detect-workflow is a Github Action that detects the repository and ref (the
branch, tag, or commit SHA) for a reusable workflow. This is useful in order to
check out code from the reusable workflow's repository at the git ref that the
user workflow specifies.

detect-workflow creates an OIDC token from Github's OIDC provider and reads the
[`job_workflow_ref`](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token)
in order to determine the reusable workflow's repository.

## Why?

When a Github Actions reusable workflow is executed, the workflow only has
access to the executing yaml itself. If any other code is required to execute
the workflow it needs to be checked out from the resuable workflow's repository
at the ref that was specifed by the user. A call to a reusable workflow might
look like this:

```yaml
myjob:
  permissions:
    id-token: write
    contents: read
    actions: read
  # {owner}/{repository}{/path}@{ref}
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@main
  with:
    subjects: "${{ needs.build.outputs.digest }}"
```

However, it is not trivial to determine the repository and ref because the
[Github Actions
contexts](https://docs.github.com/en/actions/learn-github-actions/contexts)
refer to the user workflow that called the reusable workflow. The reusable
workflow should also ideally avoid hard-coding this information as the
repository could be a fork and the ref used is unpredictable.

## Usage

To use detect-workflow, add a job to your resuable workflow with `repository`
and `ref` as an output. You can then use the output from the job in later jobs
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
        uses: slsa-framework/slsa-github-generator/.github/actions/detect-workflow@<git hash>

  # example of using the output from detect-workflow
  print:
    shell: bash
    env:
      REPO: "${{ steps.detect-workflow.outputs.repository }}"
      REF: "${{ steps.detect-workflow.outputs.ref }}"
    run: |
      echo $REPO
      echo $REF
```

In the example above, `REPO` and `REF` will be equal to the repository and ref
from the user workflow's call to the reusable workflow.

## Outputs

| Name         | Description                                                           |
| ------------ | --------------------------------------------------------------------- |
| `repository` | The repository of the reusable workflow (`{owner}/{repository name}`) |
| `ref`        | The ref (branch, tag, or commit SHA) specified by the user.           |
