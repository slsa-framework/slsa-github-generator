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

## Pre-release tests

Needless to say, only think about a release when all the e2e tests in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/) are passing. (They run daily).

There is one integration test we cannot easily test "live", so we need to simulate it by changing the code: malicious verifier binary in assets. We want to be sure the builder fails if the verifier's binary is tampered with. For this:

1. Create a branch and change the verifier repository [TODO:file]
2. Edit the file [slsa-framework/example-package](TODO) by changing the name/branch of the builder to point to your builder.
3. Run the test manually via the GitHub UX in [TODO](TODO)
4. Verify the run fails with log message `TODO`.

## Tagging

In the rest of the document, we will assume we want to release `vX.Y.Z` of the builder.

A new tag should be created via [slsa-framework/slsa-github-generator/releases/new](https://github.com/slsa-framework/slsa-github-generator/releases/new). 

The tag *MUST* be a "canonical" semnatic version without metadata (`vX.Y.Z`). Shorter versions are not accepted by the builder's and verifier's code. 

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