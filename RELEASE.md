# How to cut a release

This is a document to describe the release process for the slsa-github-generator repository.

---

<!-- markdown-toc --bullets="-" -i RELEASE.md -->

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Release candidate](#release-candidate)
  - [Update CHANGELOG](#update-changelog)
  - [Update dependencies](#update-dependencies)
  - [RC tagging](#rc-tagging)
  - [Verify RC version references](#verify-rc-version-references)
  - [Adversarial verifier tests](#adversarial-verifier-tests)
    - [Go builder verifier test](#go-builder-verifier-test)
    - [Generic generator verifier test](#generic-generator-verifier-test)
    - [Container generator verifier test](#container-generator-verifier-test)
    - [Container-based builder verifier test](#container-based-builder-verifier-test)
  - [Adversarial builder tests](#adversarial-builder-tests)
    - [Adversarial Go builder](#adversarial-go-builder)
    - [Adversarial generic generator](#adversarial-generic-generator)
    - [Adversarial container generator](#adversarial-container-generator)
    - [Adversarial container-based builder](#adversarial-container-based-builder)
  - [Finalize release candidate](#finalize-release-candidate)
  - [Code Freeze](#code-freeze)
- [Finalize release](#finalize-release)
  - [Update CHANGELOG](#update-changelog-1)
  - [Release tagging](#release-tagging)
  - [Verify final version references](#verify-final-version-references)
  - [Final adversarial tests](#final-adversarial-tests)
  - [Reference Actions at main](#reference-actions-at-main)
  - [Update verifier](#update-verifier)
  - [Finish the release](#finish-the-release)
  - [Update SECURITY.md](#update-securitymd)
  - [Update the starter workflows](#update-the-starter-workflows)
  - [Announce](#announce)

<!-- tocstop -->

---

## Prerequisites

Set up env variables:

```shell
export GITHUB_USERNAME="laurentsimon"
# This is the existing slsa-verifier version used by the builder. (https://github.com/slsa-framework/slsa-github-generator/blob/main/.github/actions/generate-builder/action.yml#L55)
export VERIFIER_TAG="v2.1.0"
export VERIFIER_REPOSITORY="$GITHUB_USERNAME/slsa-verifier"
# Release tag of the builder we want to release. Release Candidates end with "-rc.#"
export BUILDER_TAG="v1.5.0-rc.0"
# Branch name for our test
export BUILDER_REF="release/bad-verifier-$BUILDER_TAG"
export BUILDER_REPOSITORY="$GITHUB_USERNAME/slsa-github-generator"
export GH=/path/to/gh
GH_TOKEN=$(${GH} auth token)
export GH_TOKEN
```

## Release candidate

The first step in creating a release is to create a release candidate. Release candidates exercise the release and testing process.

If any tests fail for a release candidate you can address the issues and create a new release candidate after incrementing the release candidate number.

### Update CHANGELOG

Finalize the [CHANGELOG](./CHANGELOG.md) entry for the release candidate noting changes since the last release or release candidate.

### Update dependencies

In order to minimize vulnerabilities in releases, merge
[outstanding PRs from Renovate](https://github.com/slsa-framework/slsa-github-generator/pulls/renovate-bot)
as best as possible. Renovate PRs that update major versions can be skipped.

### RC tagging

Create a new tag for the Release Candidate via [slsa-framework/slsa-github-generator/releases/new](https://github.com/slsa-framework/slsa-github-generator/releases/new). The tag _MUST_ be a "canonical" [semantic version](https://semver.org/) without metadata (`$BUILDER_TAG`). Shorter versions are not accepted by the builder's and verifier's code.

Release candidates should include a suffix indicating the release candidate number of the form `-rc.#` where `#` is a number starting from `0`.

1. Set the title to `$BUILDER_TAG`
2. Add the following description.

   ```text
   **This is an un-finalized pre-release.**

   See the [CHANGELOG](./CHANGELOG.md) for details.
   ```

3. Tick the `This is a pre-release` option.
4. Click `Publish release`.

This will trigger the [release workflow](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml). Cancel this in the [UI](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml).

### Verify RC version references

Update version references with the following command:

```shell
find .github/workflows/ .github/actions/ actions/ internal/builders/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/uses: slsa-framework\/slsa-github-generator\/\(.*\)@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: slsa-framework\/slsa-github-generator\/\1@$BUILDER_TAG/"
find actions/maven/ internal/builders/maven/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/\(ref:[ ]*\)\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/\1$BUILDER_TAG/"
```

Send a PR with this update and add the following to the PR description.

```text
#label:release ${BUILDER_TAG}
```

Once the PR is merged, immediately update the tag to point to HEAD.

```shell
git tag $BUILDER_TAG -f
git push origin $BUILDER_TAG -f
```

This will trigger the [release workflow](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml). Ensure this workflow succeeds and that the release assets are updated.

### Adversarial verifier tests

There is one integration test we cannot easily test "live", so we need to simulate it by changing the code: malicious verifier binary in assets. We want to be sure the builder fails if the verifier's binary is tampered with. For this:

1. Create a new release for your fork of the slsa-verifier repository with a malicious binary.

   Create a release. Note that this will create a release workflow: cancel it in the GitHub UI.

   ```shell
   "$GH" release -R "$VERIFIER_REPOSITORY" create "$VERIFIER_TAG" --title "$VERIFIER_TAG" --notes "pre-release tests for builder $BUILDER_TAG $(date)"
   ```

   Simulate uploading a malicious binary.

   ```shell
   echo hello > slsa-verifier-linux-amd64
   "$GH" release -R "$VERIFIER_REPOSITORY" upload "$VERIFIER_TAG" slsa-verifier-linux-amd64
   ```

2. Ensure your fork of the builder is at the same commit hash as the official builder's `$BUILDER_TAG` release.
3. Create a new branch

   ```shell
   git checkout -b "$BUILDER_REF"
   ```

4. Update the file `.github/actions/generate-builder/action.yml` by replacing the strings `BUILDER_REPOSITORY` and `VERIFIER_REPOSITORY` with your own username (value of `$GITHUB_USERNAME`). Then push the changes.

   ```shell
   sed -i "s/BUILDER_REPOSITORY: slsa-framework\/slsa-github-generator/BUILDER_REPOSITORY: ${GITHUB_USERNAME}\/slsa-github-generator/" .github/actions/generate-builder/action.yml
   sed -i "s/VERIFIER_REPOSITORY: slsa-framework\/slsa-verifier/VERIFIER_REPOSITORY: ${GITHUB_USERNAME}\/slsa-verifier/" .github/actions/generate-builder/action.yml
   ```

5. For the Go builder, update the file `.github/workflows/builder_go_slsa3.yml` to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

   using this command:

   ```shell
   sed -i "s/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/generate-builder@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: ${BUILDER_REPOSITORY/\//\/}\/.github\/actions\/generate-builder@${BUILDER_TAG}/" .github/workflows/builder_go_slsa3.yml
   ```

   Add `testing: true` as an input.

6. For the Generic generator, update the file `.github/workflows/generator_generic_slsa3.yml` to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

   using this command:

   ```shell
   sed -i "s/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/generate-builder@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: ${BUILDER_REPOSITORY/\//\/}\/.github\/actions\/generate-builder@${BUILDER_TAG}/" .github/workflows/generator_generic_slsa3.yml
   ```

   Add `testing: true` as an input.

7. For the Container generator, update the file `.github/workflows/generator_container_slsa3.yml` to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

   using this command:

   ```shell
   sed -i "s/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/generate-builder@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: ${BUILDER_REPOSITORY/\//\/}\/.github\/actions\/generate-builder@${BUILDER_TAG}/" .github/workflows/generator_container_slsa3.yml
   ```

   Add `testing: true` as an input.

8. For the Container-based generator, update the file `.github/workflows/builder_container-based_slsa3.yml` to:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/actions/generate-builder@$BUILDER_TAG
   ```

   using this command:

   ```shell
   sed -i "s/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/generate-builder@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: ${BUILDER_REPOSITORY/\//\/}\/.github\/actions\/generate-builder@${BUILDER_TAG}/" .github/workflows/builder_container-based_slsa3.yml
   ```

   Add `testing: true` as an input.

9. Commit and push the changes

10. Create a release for the builders for this branch:

    ```shell
    "$GH" release -R "$BUILDER_REPOSITORY" create "$BUILDER_TAG" --title "$BUILDER_TAG" --notes "pre-release tests for $BUILDER_TAG $(date)" --target "$BUILDER_REF"
    ```

    This will trigger a workflow release, let it complete and generate the release assets.

#### Go builder verifier test

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/builder_go_slsa3.yml@$BUILDER_TAG
   ```

2. Run the test manually via the GitHub UX in [https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by cliking `Run Workflow`.
3. Verify the run fails with log message:

   ```text
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

#### Generic generator verifier test

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/generator_generic_slsa3.yml@$BUILDER_TAG
   ```

2. Run the test manually via the GitHub UX in [https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by cliking `Run Workflow`.
3. Verify the run fails with log message:

   ```text
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

#### Container generator verifier test

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.container.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.container.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/generator_container_slsa3.yml@$BUILDER_TAG
   ```

2. Run the test manually via the GitHub UX in
   [https://github.com/slsa-framework/example-package/actions/workflows/e2e.container.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.container.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml)
   by cliking `Run Workflow`.

3. Verify the run fails with log message:

   ```text
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

#### Container-based builder verifier test

1. Edit the file [slsa-framework/example-package/.github/workflows/e2e.container-based.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.container-based.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml) by using `$BUILDER_REPOSITORY` and `$BUILDER_TAG`:

   ```yaml
   uses: $BUILDER_REPOSITORY/.github/workflows/builder_container-based_slsa3.yml@$BUILDER_TAG
   ```

2. Run the test manually via the GitHub UX in
   [https://github.com/slsa-framework/example-package/actions/workflows/e2e.container-based.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.container-based.workflow_dispatch.main.adversarial-verifier-binary.slsa3.yml)
   by cliking `Run Workflow`.

3. Verify the run fails with log message:

   ```text
   verifier hash computed is 5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
   Error: Process completed with exit code 4.
   ```

### Adversarial builder tests

End-to-end tests run daily in [github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/), and contain adversarial tests (developer tampers with the artifacts used by the builders). All these adversarial tests compile the builder from source (`compile-builder: true`). But we need to verify that the builder detects malicious builder's binary when `compile-builder: false` (default).

#### Adversarial Go builder

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-builder-go-linux-amd64`, either via the web UI or via:

   ```shell
   "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-go-linux-amd64"
   mv slsa-builder-go-linux-amd64 slsa-builder-go-linux-amd64-"$BUILDER_TAG".original
   ```

2. Upload a different binary to the assets:

   ```shell
   echo hello > slsa-builder-go-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
   ```

3. Update the version of the workflow
   [slsa-framework/example-package/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   with the `$BUILDER_TAG` to test.

4. Trigger the test in [slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.go.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml) by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

5. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   mv slsa-builder-go-linux-amd64-"$BUILDER_TAG".original slsa-builder-go-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-go-linux-amd64  --clobber
   ```

6. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

#### Adversarial generic generator

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-generator-generic-linux-amd64`, either via the web UI or via:

   ```shell
   "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-generator-generic-linux-amd64"
   mv slsa-generator-generic-linux-amd64 slsa-generator-generic-linux-amd64-"$BUILDER_TAG".original
   ```

2. Upload a different binary to the assets:

   ```shell
   echo hello > slsa-generator-generic-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-generic-linux-amd64  --clobber
   ```

3. Update the version of the workflow
   [slsa-framework/example-package/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   with the `$BUILDER_TAG` to test.

4. Trigger the test in
   [slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.generic.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

5. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   mv slsa-generator-generic-linux-amd64-"$BUILDER_TAG".original slsa-generator-generic-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-generic-linux-amd64  --clobber
   ```

6. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

#### Adversarial container generator

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-generator-container-linux-amd64`, either via the web UI or via:

   ```shell
   "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-generator-container-linux-amd64"
   mv slsa-generator-container-linux-amd64 slsa-generator-container-linux-amd64-"$BUILDER_TAG".original
   ```

2. Upload a different binary to the assets:

   ```shell
   echo hello > slsa-generator-container-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-container-linux-amd64  --clobber
   ```

3. Update the version of the workflow
   [slsa-framework/example-package/.github/workflows/e2e.container.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.container.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   with the `$BUILDER_TAG` to test.

4. Trigger the test in
   [slsa-framework/example-package/actions/workflows/e2e.container.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.container.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

5. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   mv slsa-generator-container-linux-amd64-"$BUILDER_TAG".original slsa-generator-container-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-generator-container-linux-amd64  --clobber
   ```

6. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

#### Adversarial container-based builder

1. Make sure you have downloaded the `$BUILDER_TAG` builder's binary locally `slsa-builder-docker-linux-amd64`, either via the web UI or via:

   ```shell
   "$GH" release -R slsa-framework/slsa-github-generator download "$BUILDER_TAG" -p "slsa-builder-docker-linux-amd64"
   mv slsa-builder-docker-linux-amd64 slsa-builder-docker-linux-amd64-"$BUILDER_TAG".original
   ```

2. Upload a different binary to the assets:

   ```shell
   echo hello > slsa-builder-docker-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-docker-linux-amd64  --clobber
   ```

3. Update the version of the workflow
   [slsa-framework/example-package/.github/workflows/e2e.container-based.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/e2e.container-based.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   with the `$BUILDER_TAG` to test.

4. Trigger the test in
   [slsa-framework/example-package/actions/workflows/e2e.container-based.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/e2e.container-based.workflow_dispatch.main.adversarial-builder-binary.slsa3.yml)
   by cliking `Run workflow`. Verify that it fails, with a message:

   ```shell
   verifier hash computed is 60c91c9d5b9a059e37ac46da316f20c81da335b5d00e1f74d03dd50f819694bd
   verifier hash verification has passed
   ...
   FAILED: SLSA verification failed: expected hash '5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03', got 'e8af48495ca3c5a7737b4a34322afc7e95a85cf1457a37473fb81cff9b4f0d05': binary artifact hash does not match provenance subject
   Error: Process completed with exit code 6.
   ```

5. If the test above failed with the expected message, re-upload the original binary back to the assets, e.g. via:

   ```shell
   mv slsa-builder-docker-linux-amd64-"$BUILDER_TAG".original slsa-builder-docker-linux-amd64
   "$GH" release -R slsa-framework/slsa-github-generator upload "$BUILDER_TAG" slsa-builder-docker-linux-amd64  --clobber
   ```

6. Re-run the workflow above and verify that it succeeds. (TODO: https://github.com/slsa-framework/slsa-github-generator/issues/116).

   If it does not, delete the release, fix the bug and re-start the release process at the top of this page.

### Finalize release candidate

Remove the "This is an un-finalized pre-release." note from the release description.

### Code Freeze

Code freeze the repository for 1-2 days.

After the code freeze, verify all the e2e tests in
[github.com/slsa-framework/example-package/.github/workflows/](github.com/slsa-framework/example-package/.github/workflows/)
are passing. (They run daily).

## Finalize release

Once the code release is complete you may create a final release.

Update your `BUILDER_TAG` environment variable to the final release tag.

```shell
export BUILDER_TAG="vX.Y.Z"
```

### Update CHANGELOG

Finalize the [CHANGELOG](./CHANGELOG.md) entry for the release candidate noting changes since the last major release (not including release candidates).

### Release tagging

Create a new tag for the final release via [slsa-framework/slsa-github-generator/releases/new](https://github.com/slsa-framework/slsa-github-generator/releases/new). The tag _MUST_ be a "canonical" [semantic version](https://semver.org/) without metadata (`$BUILDER_TAG`). Shorter versions are not accepted by the builder's
and verifier's code.

1. Set the title to `$BUILDER_TAG`
2. Add the following description.

   ```text
   **This is an un-finalized release.**

   See the [CHANGELOG](./CHANGELOG.md) for details.
   ```

3. Tick the `This is a pre-release` option.
4. Click `Publish release`.

This will trigger the [release workflow](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml). Cancel this in the [UI](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml).

### Verify final version references

Update version references with the following command:

```shell
find .github/workflows/ .github/actions/ actions/ internal/builders/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/uses: slsa-framework\/slsa-github-generator\/\(.*\)@\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/uses: slsa-framework\/slsa-github-generator\/\1@$BUILDER_TAG/"
# NOTE: The gradle builder does not need this update.
find actions/maven/ internal/builders/maven/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/\(ref:[ ]*\)\(main\|v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?\)/\1$BUILDER_TAG/"
```

Likewise, update documentation with the following command:

```shell
find . -name "*.md" -type f -exec sed -i "s~\(uses: .*/slsa-github-generator/.*@\)v[0-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?~\1$BUILDER_TAG~g" {} +
```

Send a PR with this update and add the following to the PR description.

```text
#label:release ${BUILDER_TAG}
```

Once the PR is merged, immediately update the tag to point to HEAD.

```shell
git tag $BUILDER_TAG -f
git push origin $BUILDER_TAG -f
```

This will trigger the [release workflow](https://github.com/slsa-framework/slsa-github-generator/actions/workflows/release.yml). Ensure this workflow succeeds and that the release assets are updated.

### Final adversarial tests

Re-run the [adversarial builder tests](#adversarial-builder-tests) using the final `$BUILDER_TAG` for the release. If any tests fail you will need to delete the release and address the issues.

### Reference Actions at main

Send a PR to reference the Actions at `@main`. You can use:

```shell
find .github/workflows/ .github/actions/ actions/ internal/builders/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/uses: slsa-framework\/slsa-github-generator\/\(.*\)@${BUILDER_TAG}/uses: slsa-framework\/slsa-github-generator\/\1@main/"
find actions/maven/ internal/builders/maven/ -name '*.yaml' -o -name '*.yml' -type f | xargs sed -i "s/\(ref:[ ]*\)$BUILDER_TAG/\1main/"
```

### Update verifier

The next step is to update the verifier's GitHub Actions e2e tests. There are GitHub actions Go and generic actions.

<!-- TODO(github.com/slsa-framework/slsa-github-generator/issues/1110): Describe GHA generic container e2e tests. -->

For the BYOB (a.k.a delegator) workflows, you will need to update the tag of the [slsa-framework/example-trw](https://github.com/slsa-framework/example-trw/) repository:

1. Update the references to the tag and send a PR:

   ```bash
   bash update-main-to-tag.sh "${BUILDER_TAG}"
   ```

2. Cut a release with tag `${BUILDER_TAG}`.

3. Update the references back to main and send a PR:

   ```bash
   bash update-tag-to-main.sh "${BUILDER_TAG}"
   ```

Then, for each of the GHA builders, you will need to:

1. Generate binaries and provenance in
   [example-package](https://github.com/slsa-framework/example-package) using
   the GHA action builder.

   These require using the updated builders, so the first step is to update
   [verifier-e2e.all.workflow_dispatch.main.all.slsa3.yml](https://github.com/slsa-framework/example-package/blob/main/.github/workflows/verifier-e2e.all.workflow_dispatch.main.all.slsa3.yml)
   to reference actions at `$BUILDER_TAG`.

   For example:

   ```yaml
   uses: slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@<BUILDER_TAG>
   ```

   Next run the
   [verifier-e2e.all.workflow_dispatch.main.all.slsa3.yml](https://github.com/slsa-framework/example-package/actions/workflows/verifier-e2e.all.workflow_dispatch.main.all.slsa3.yml).
   This will dispatch the workflow and create provenance for the workflow
   dispatch event, and then trigger subsequent runs on the following fixed
   release tags.

   - [v14](https://github.com/slsa-framework/example-package/releases/tag/v14)
   - [v14.2](https://github.com/slsa-framework/example-package/releases/tag/v14.2)
   - [v13.0.30](https://github.com/slsa-framework/example-package/releases/tag/v13.0.30)

   Wait for the runs to complete

2. Download the uploaded artifacts of each of the created releases. You can make use of [download-artifacts.sh](https://github.com/slsa-framework/slsa-verifier/blob/main/download-artifacts.sh) (not fully automated).

3. Move these files to
   `./cli/slsa-verifier/testdata/gha_$BUILDER_TYPE/$BUILDER_TAG/` in the
   slsa-verifier repository. You may need to rename the files to match the
   directory structure.

4. Determine the digest that was uploaded by the build to the
   [example-package.verifier-e2e.all.tag.main.default.slsa3](https://github.com/slsa-framework/example-package/pkgs/container/example-package.verifier-e2e.all.tag.main.default.slsa3)
   package. Create the file `./cli/slsa-verifier/testdata/gha_generic_container/$BUILDER_TAG/container_workflow_dispatch.digest` with the content `sha256:digest` (no new line).

5. Export the image to the
   `./cli/slsa-verifier/testdata/gha_generic_container/$BUILDER_TAG/` directory
   in the slsa-verifier repository.

   Make sure you are authenticated to GitHub's Container registry:

   ```shell
   gh auth login --scopes=read:packages
   echo `gh auth token` | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
   ```

   Export the image:

   ```shell
   cosign save \
       --dir ./cli/slsa-verifier/testdata/gha_generic_container/$BUILDER_TAG/container_workflow_dispatch \
       ghcr.io/slsa-framework/example-package.verifier-e2e.all.tag.main.default.slsa3@sha256:<digest>
   ```

6. Send a pull request to merge the changes into the verifier's repository. The
   pre-submits will validate that the verifier is able to verify provenance from
   the `$BUILDER_TAG` builder.

### Finish the release

1. Remove the "This is an un-finalized release." note from the release description.
2. Un-tick the `This is a pre-release` option.
3. If it's the latest release, tick the `Set as the latest release` option.

### Update SECURITY.md

Update the `Supported Versions` section in [SECURITY.md](./SECURITY.md).

### Update the starter workflows

Update:

1. [Go builder's workflow](https://github.com/actions/starter-workflows/blob/main/ci/go-ossf-slsa3-publish.yml)
2. [Generic generator's workflow](https://github.com/actions/starter-workflows/blob/main/ci/generator-generic-ossf-slsa3-publish.yml)

### Announce

<!-- TODO(release): Provide details  -->

Announce the release to users.
