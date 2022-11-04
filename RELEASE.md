# How to cut a release

This is a document to describe the release process for the Go builder. Since all builders are in this repository, we will expand this doc to cover other builders in the future.

---

- [Pre-release](#pre-release-tests)
- [Tagging](#tagging)
- [Post-release tests](#post-release-tests)
- [Update Verifier](#update-verifier)
- [Finalize release](#finalize-release)
- [Announce](#announce)

---

## Prerequisites

Set up env variables:

```shell
$ export GH_TOKEN=<PAT-token>
$ export GITHUB_USERNAME="laurentsimon"
# This is the existing slsa-verifier version used by the builder. (https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/actions/generate-builder/action.yml#L55)
$ export VERIFIER_TAG="v1.3.2"
$ export VERIFIER_REPOSITORY="$GITHUB_USERNAME/slsa-verifier"
# Release tag of the builder we want to release
$ export BUILDER_TAG="v1.2.0"
# Branch name for our test
$ export BUILDER_REF="release/bad-verifier-$BUILDER_TAG"
$ export BUILDER_REPOSITORY="$GITHUB_USERNAME/slsa-github-generator"
$ export GH=/path/to/gh
```

## Pre-release tests

Verify the references to the internal Actions by manually running the [release workflow](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml). Ensure this workflow succeeds. If the references are still `@main`, update them with the following command:

```shell
find .github/workflows/ -name '*.yaml' -o -name '*.yml' | xargs sed -i 's/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/\(.*\)@main*/uses: slsa-framework\/slsa-github-generator\/.github\/actions\/\1@_YOUR_RELEASE_TAG_/'
```

Send a PR with this update and add `#label:release` in the PR description.

Code freeze the repository for 1-2 days.

Verify all the e2e tests in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/) are passing. (They run daily).

There is one integration test we cannot easily test "live", so we need to simulate it by changing the code: malicious verifier binary in assets. We want to be sure the builder fails if the verifier's binary is tampered with. For this:

1. Create a new release for your fork of the slsa-verifier repository with a malicious binary.

   ```shell
   # Create a release

   $ "$GH" release -R "$VERIFIER_REPOSITORY" create "$VERIFIER_TAG" --title "$VERIFIER_TAG" --notes "pre-release tests for builder $BUILDER_TAG $(date)"
   $ # Note: this will create a release workflow: cancel it in the GitHub UI.

   # Upload a malicious binary.

   $ echo hello > slsa-verifier-linux-amd64
   $ "$GH" release -R "$VERIFIER_REPOSITORY" upload "$VERIFIER_TAG" slsa-verifier-linux-amd64
   ```

1. Ensure your fork of the builder is at the same commit hash as the offical builder's `$BUILDER_TAG` release.
1. Create a new branch `git checkout -b "$BUILDER_REF"`
1. Update the file `$BUILDER_REPOSITORY/main/.github/actions/generate-builder/action.yml` by replacing the strings `BUILDER_REPOSITORY` and `VERIFIER_REPOSITORY` with your own username (value of `$GITHUB_USERNAME`). Then push the changes.
1. For the Go builder, update the file `$BUILDER_REPOSITORY/main/.github/workflows/builder_go_slsa3.yml` to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

1. For the Generic generator, update the file `$BUILDER_REPOSITORY/main/.github/workflows/generator_generic_slsa3.yml`to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

1. Create a release for the builders for this branch:

   ```shell
   $ "$GH" release -R "$BUILDER_REPOSITORY" create "$BUILDER_TAG" --title "$BUILDER_TAG" --notes "pre-release tests for $BUILDER_TAG $(date)" --target "$BUILDER_REF"
   ```

   This will trigger a workflow release, let it complete and generate the release assets.

### Go builder

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/builder_go_slsa3.yml@$BUILDER_TAG
   ```

1. Run the test manually via the GitHub UX in [https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by cliking `Run Workflow`.
1. Verify the run fails with log message:

   ```
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

### Generic generator

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L36](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L36) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/generator_generic_slsa3.yml@$BUILDER_TAG
   ```

1. Run the test manually via the GitHub UX in [https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by cliking `Run Workflow`.
1. Verify the run fails with log message:

   ```
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

## Tagging

Create a new tag for the official generator via [slsa-framework/slsa-github-generator/releases/new](https://github.com/slsa-framework/slsa-github-generator/releases/new).
The tag _MUST_ be a "canonical" semantic version without metadata (`$BUILDER_TAG`). Shorter versions are not accepted by the builder's and verifier's code.

Set the title to `$BUILDER_TAG`.

Tick the `This is a pre-release` option.

Click `Publish release`.

Download the generated binary `slsa-builder-go-linux-amd64` locally on your machine:

```
$ "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-go-linux-amd64"
```

## Post-release tests

End-to-end tests run daily in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/), and contain adversarial tests (developer tampers with the artifacts used by the builders). All these adversarial tests compile the builder from source (`compile-builder: true`). But we need to verify that the builder detects malicious builder's binary when `compile-builder: false` (default).

### Go builder

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-builder-go-linux-amd64`, either via the web UI or via:

   ```shell
   $ "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-go-linux-amd64"
   $ mv slsa-builder-go-linux-amd64 slsa-builder-go-linux-amd64-"$BUILDER_TAG".original
   ```

1. Upload a different binary to the assets:

   ```shell
   $ echo hello > slsa-builder-go-linux-amd64
   $ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
   ```

1. Update the version of the workflow [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L14](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L14) with the `$BUILDER_TAG` to test.

1. Trigger the test in [slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml) by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

1. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   $ mv slsa-builder-go-linux-amd64-"$BUILDER_TAG".original slsa-builder-go-linux-amd64
   $ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
   ```

1. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

### Generic generator

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-generator-generic-linux-amd64`, either via the web UI or via:

   ```shell
   $ "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-generator-generic-linux-amd64"
   $ mv slsa-generator-generic-linux-amd64 slsa-generator-generic-linux-amd64-"$BUILDER_TAG".original
   ```

1. Upload a different binary to the assets:

   ```shell
   $ echo hello > slsa-generator-generic-linux-amd64
   $ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-generic-linux-amd64  --clobber
   ```

1. Update the version of the workflow [slsa-framework/example-package/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L35](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L35) with the `$BUILDER_TAG` to test.

1. Trigger the test in [slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml) by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

1. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   $ mv slsa-generator-generic-linux-amd64-"$BUILDER_TAG".original slsa-generator-generic-linux-amd64
   $ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-generic-linux-amd64  --clobber
   ```

1. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

## Update verifier

The next step is to update the verifier's GitHub Actions e2e tests. There are GitHub actions Go and generic actions.

<!-- TODO(https://github.com/slsa-framework/slsa-github-generator/issues/1110): Describe GHA generic container e2e tests. -->

For each of the GHA builders, you will need to:

1. Generate binaries and provenance in [example-package](https://github.com/slsa-framework/example-package) using the GHA action builder. These require using the updated builders, so validate that the workflows you use below are pinned at `$BUILDER_TAG`.

You will need the following trigger types:
* A workflow dispatch event.
* A tag of the form `vX.Y.Z`.
* Tags of the form `vX` and `vX.Y`.

To do this, trigger the [Go workflow dispatch](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/verifier-e2e.go.workflow_dispatch.main.all.slsa3.yml) and [Generic workflow dispatch](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/verifier-e2e.generic.workflow_dispatch.main.all.slsa3.yml). These will dispatch the workflow and create provenance for the workflow dispatch event, and then trigger subsequent runs on fixed tags.

Download the uploaded artifacts of each of these, labelling the workflow dispatch artifacts by `binary-linux-amd64-workflow_dispatch(.intoto.jsonl)` and the tags by `binary-linux-amd64-push-v$TAG(.intoto.jsonl)`. 

2. Move these files to `./cli/slsa-verifier/testdata/gha_$BUILDER_TYPE/$BUILDER_TAG/`. Send a pull request to merge the changes into the verifier's repository. The pre-submits will validate that the verifier is able to verify provenance from the `$BUILDER_TAG` builder.

## Finalize release

Untick the `This is a pre-release` option.

Update the documentation to recommend using the new version:

```shell
$ find . -name "*.md" -exec sed -i "s/v1.0.0/v1.1.1/g" {} +
```

## Send a PR to reference Actions at main

Send a PR to reference the internal Actions at `@main`. You can use:

```shell
find .github/ -name '*.yaml' -o -name '*.yml' | xargs sed -i 's/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/\(.*\)@_YOUR_RELEASE_TAG_*/uses: slsa-framework\/slsa-github-generator\/.github\/actions\/\1@main/'
```

## Update the starter workflows

Update:

1. [Go builder's workflow](https://github.com/actions/starter-workflows/blob/main/ci/go-ossf-slsa3-publish.yml)
1. [Generic generator's workflow](https://github.com/actions/starter-workflows/blob/main/ci/generic-generator-ossf-slsa3-publish.yml)

## Announce

<!-- TODO(release): Provide details -->
