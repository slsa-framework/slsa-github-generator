# Releasing the Go builder

This is a  document to describe the release process for the Go builder. Since all builders are in this repository, we will expand this doc to cover other builders in the future.

---

- [Pre-release](#pre-release-tests)
- [Tagging](#tagging)
- [Post-release tests](#post-release-tests)
- [Update Verifier](#update-verifier)
- [Finalize release](#finalize-release)
- [Announce](#announce)

---

## Pre-requesites

Set up env variables:
```
```
$ export GH_TOKEN=<PAT-token>
$ export GITHUB_USERNAME="laurentsimon"
$ export VERIFIER_TAG="v2.3.4" # change accordingly.
$ export VERIFIER_REPOSITORY="$GITHUB_USERNAME/slsa-verifier"
$ export BUILDER_TAG="v0.0.2"
$ export BUILDER_REF="feat/bad-verifier"
$ export GH=/path/to/gh
```

## Pre-release tests

Needless to say, only think about a release when all the e2e tests in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/) are passing. (They run daily).

There is one integration test we cannot easily test "live", so we need to simulate it by changing the code: malicious verifier binary in assets. We want to be sure the builder fails if the verifier's binary is tampered with. For this:

1. Create a new release for your fork of the slsa-verifier repository, e.g. `v2.3.4` with a malicious binary (We need a release because the builder only accepts reference tags). 

```
$ echo hello > slsa-verifier-linux-amd64
$ "$GH" release -R "$VERIFIER_REPOSITORY" create "$VERIFIER_TAG" --title "$VERIFIER_TAG" --notes "pre-release tests for builder $BUILDER_TAG $(date)"
$ # Note: this will create a release workflow: cancel it in the GitHub UI.
$ "$GH" release -R "$VERIFIER_REPOSITORY" upload "$VERIFIER_TAG" slsa-verifier-linux-amd64
```
2. Ensure your fork of the builder is at the same commit hash as the offical builder's `$BUILDER_TAG` release.
3. Create a new branch `git checkout -b "$BUILDER_REF"`
4. Update the file `$VERIFIER_REPOSITORY/main/.github/workflows/builder_go_slsa3.yml#L28` and push the changes.
3. Create a release for your builder:
```

```
3. Edit the file [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml#L14) by using your own repo/tag.
3. Run the test manually via the GitHub UX in [https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) and click `Run Workflow`.
4. Verify the run fails with log message `TODO`.

## Tagging

A new tag should be created via [slsa-framework/slsa-github-generator/releases/new](https://github.com/slsa-framework/slsa-github-generator/releases/new). 

The tag *MUST* be a "canonical" semantic version without metadata (`vX.Y.Z`). Shorter versions are not accepted by the builder's and verifier's code. 

Set the title to `vX.Y.Z`.

Tick the `This is a pre-release` option.

Click `Publish release`.

Download the generated binary `slsa-builder-go-linux-amd64` locally on your machine.

## Post-release tests

End-to-end tests run daily in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/), and contain adversarial tests (developer tampers with the artifacts used by the builders). All these adversarial tests compile the builder from source (`compile-builder: true`). But we need to verify that the builder detects malicious builder's binary when `compile-builder: false` (default).

For this:
1. Make sure you have downloaded the `vX.Y.Z` builder's binary locally `slsa-builder-go-linux-amd64`, either via the web UI or via `TODO:command`.
2. Upload a different binary to the assets. You can use the [GitHub CLI](TODO) command as follows:
3. Trigger the test in [TODO](TODO) and verifies that it fails, with a message `TODO`. If it does not, delete the release, fix the bug and re-start the release process at the top of this page.
4. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via `TODO:command`.

## Update verifier

The next step is to update the verifier's e2e tests. For this, you need to:

1. Generate binaries and provenance for a project, using the `vX.Y.Z` builder. 

2. Place the files in a new directory [slsa-framework/slsa-verifier/tree/main/testdata/vX.Y.Z](https://github.com/slsa-framework/slsa-verifier/tree/main/testdata/vX.Y.Z).

3. Add the new release to the list defined in [slsa-framework/slsa-verifier/blob/main/main_test.go](https://github.com/slsa-framework/slsa-verifier/blob/main/main_test.go).

Send a pull request to merge the changes into the verifier's repository. The pre-submit will validate that the verifier is able to verify provenance from the `vX.Y.Z` builder.

## Finalize release

Untick the `This is a pre-release` option. That's it!

## Update the starter workflow

TODO

## Announce

<!-- TODO(release): Provide details -->
