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
2. Add the token as a Repository Secret, `my-slsa-gh-token`, for example
3. Supply token to the action

The generic generator workflows will expect this token when they pass
along to this action, so you can also use them like

```yaml
jobs:
  detect-workflow:
    runs-on: ubuntu-latest
      uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v1.9.0@<git hash>
      secrets:
        token: ${{ secrets.my-slsa-gh-token }}
```

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

## Usage

First create a Token and Secret described [above](#administrationread-permissions)

```yaml
jobs:
  detect-workflow:
    runs-on: ubuntu-latest
    steps:
      - name: Detect the repository and ref
        id: detect
        uses: slsa-framework/slsa-github-generator/.github/actions/ensure-github-hosted-runners@<git hash>
        with:
          token: ${{ secrets.my-slsa-gh-token }}
```
