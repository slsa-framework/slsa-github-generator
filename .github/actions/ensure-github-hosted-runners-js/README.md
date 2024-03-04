# ensure-github-hosted-runners-js

Github action runners can be either github-hosted or self-hosted.

`ensure-github-hosted-runners-js` is a Github Action that ensures that no self-hosted 
runners are used for any adjacent jobs in the workflow.

It does this by comparing the `runs-on` labels in the job specs with the 
labels of the repository's known self-hosted runner labels.

## Why?

SLSA Build 3 is intends for builds to be performed on trusted build platforms, which
means only using github-hosted runners.

All of our language-specific builder workflows both build code and produce provenance.
They are run using Github Reusable Workflows, which we know will always run on github-hoastwed runners.
Therefore we can trust that both the build artifacts and provenance were produce on
github-hosted runners.

But our generic workflows do not perform actual builds. Instead they accept the build artifacts
produced by the the user's own jobs, which the user could specify to run on self-hosted runners.

## Caveats

### Administration:read permissions

This action requires the user to supply their own token with extra permissions, rather than the 
default token that Github will automatically supply to workflows. We will have our calling workflows
pass the token along.

1. Create a new Token for your repository, with:
  1. actions:read
  2. administration:read
2. Add the token as a Repository Secret, `my-slsa-gh-token`
3. Supply token to the actio

### Race conditon

We acknowledge that a malicious repository woner could still workaround this mechanism.
At the start of a workflow, they could have one Job run with `runs-on: my-vps` that targets one of 
their self-hosted runbners. Then, after that Job completes, but before the slsa-framework job starts,
they could use ther github UI to remove that particual `my-vps` label from their self-hosted runner.
This action would not detect that the first job did actually run on a slef-hosted runners.

### Considered mitigations

#### What about using a list of known github-hosted runner labels?

Github does not publish a difnitive list. But an even greater problem is that the user could assign their
runner any of Github's common labels. They could label their self-hosted runner with `ubuntu-latest`, for example.


#### What about using the Job's runner ID?

Github's List Jobs for Workflow Run also returns the assigned runner IDs for the job. We had considered
using these IDs against the known self-hsoted runner IDs, but it still does not solve the race codnition problem
because a user may delete their self-hosted runner before this action executes. And Github does not publish
a list of github-hosted runner IDs.


TODO: usage docs


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

In the case of a pull request to the slsa-github-generator, we use the head
SHA as the ref and the head repository as the repository.

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
