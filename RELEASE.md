# How to cut a release

This is a  document to describe the release process for the Go builder. Since all builders are in this repository, we will expand this doc to cover other builders in the future.

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
```
$ export GH_TOKEN=<PAT-token>
$ export GITHUB_USERNAME="laurentsimon"
# This is the existing slsa-verifier version used by the builder. (https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/actions/generate-builder/action.yml#L55)
$ export VERIFIER_TAG="v1.1.1"
$ export VERIFIER_REPOSITORY="$GITHUB_USERNAME/slsa-verifier"
# Release tag of the builder we want to release
$ export BUILDER_TAG="v1.2.0"
# Branch name for our test
$ export BUILDER_REF="release/bad-verifier-$BUILDER_TAG"
$ export BUILDER_REPOSITORY="$GITHUB_USERNAME/slsa-github-generator"
$ export GH=/path/to/gh
```

## Pre-release tests

Needless to say, only think about a release when all the e2e tests in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/) are passing. (They run daily).

There is one integration test we cannot easily test "live", so we need to simulate it by changing the code: malicious verifier binary in assets. We want to be sure the builder fails if the verifier's binary is tampered with. For this:

1. Create a new release for your fork of the slsa-verifier repository with a malicious binary. 
```
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
1. For the Go builder, update the file `$BUILDER_REPOSITORY/main/.github/workflows/builder_go_slsa3.yml#L98` by replacing the strings `BUILDER_REPOSITORY` with your own username (value of `$GITHUB_USERNAME`). Then push the changes.
1. For the Generic generator, update the file `$BUILDER_REPOSITORY/main/.github/workflows/generic_generator_slsa3.yml#L98` by replacing the strings `BUILDER_REPOSITORY` with your own username (value of `$GITHUB_USERNAME`). Then push the changes.
1. Create a release for the builders for this branch:
```
$ "$GH" release -R "$BUILDER_REPOSITORY" create "$BUILDER_TAG" --title "$BUILDER_TAG" --notes "pre-release tests for $BUILDER_TAG $(date)" --target "$BUILDER_REF"
```
This will trigger a workflow release, let it complete and generate the release assets.

### Go builder
1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:
```
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
```
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
The tag *MUST* be a "canonical" semantic version without metadata (`$BUILDER_TAG`). Shorter versions are not accepted by the builder's and verifier's code. 

Set the title to `$BUILDER_TAG`.

Tick the `This is a pre-release` option.

Click `Publish release`.

Download the generated binary `slsa-builder-go-linux-amd64` locally on your machine:
```
$ "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-go-linux-amd64"
```

## Post-release tests

End-to-end tests run daily in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/), and contain adversarial tests (developer tampers with the artifacts used by the builders). All these adversarial tests compile the builder from source (`compile-builder: true`). But we need to verify that the builder detects malicious builder's binary when `compile-builder: false` (default).

For this:
1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-builder-go-linux-amd64`, either via the web UI or via:
```
$ "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-go-linux-amd64"
$ mv slsa-builder-go-linux-amd64 slsa-builder-go-linux-amd64-"$BUILDER_TAG".original
```
1. Upload a different binary to the assets:
```
$ echo hello > slsa-builder-go-linux-amd64
$ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
```
1. Update the version of the workflow [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L14](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml#L14) with the `$BUILDER_TAG` to test.
1. Trigger the test in [slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml) by cliking `Run workflow`. Verify that it fails, with a message:
```
verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
verifier hash verification has passed
...
FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
Error: Process completed with exit code 6.
```

If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

1. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:
```
$ mv slsa-builder-go-linux-amd64-"$BUILDER_TAG".original slsa-builder-go-linux-amd64
$ "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
```

1. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

## Update verifier

The next step is to update the verifier's e2e tests. For this, you need to:

1. Generate binaries and provenance for a project, using the [example-package](https://github.com/slsa-framework/example-package) builder. You will need to create provenance via a `workflow_dispatch` from this [workflow](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.config-noldflags.slsa3.yml) of the e2e test repository, and push a tag of the form vX and vX.Y to trigger this [workflow](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.tag.main.config-ldflags-assets.major.slsa3.yml).

1. Place the files in a new directory `slsa-framework/slsa-verifier/tree/main/testdata/$BUILDER_TAG`.

1. Add the new release to the list defined in [slsa-framework/slsa-verifier/blob/main/main_test.go](https://github.com/slsa-framework/slsa-verifier/blob/main/main_test.go).

Send a pull request to merge the changes into the verifier's repository. The pre-submits will validate that the verifier is able to verify provenance from the `$BUILDER_TAG` builder.

## Finalize release

Untick the `This is a pre-release` option. 

Update the documentation to recommend using the new version:

```shell
$ find . -name "*.md" -exec sed -i "s/v1.0.0/v1.1.1/g" {} +
```

## Update the starter workflows

Update:

1. [Go builder's workflow](https://github.com/actions/starter-workflows/blob/main/ci/go-ossf-slsa3-publish.yml)
1. [Generic generator's workflow](https://github.com/actions/starter-workflows/blob/main/ci/generic-generator-ossf-slsa3-publish.yml)

## Announce

<!-- TODO(release): Provide details -->
