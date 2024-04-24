# Publishing SLSA3+ provenance to Maven Central

This document explains how to publish SLSA3+ artifacts and provenance to Maven central.

The publish Action is in its early stages and is likely to develop over time. Future breaking changes may occur.

To get started with publishing artifacts to Maven Central Repository, see [this guide](https://maven.apache.org/repository/guide-central-repository-upload.html).

Before you use this publish Action, you will need to configure your Github project with the correct secrets. See [this guide](https://docs.github.com/en/actions/publishing-packages/publishing-java-packages-with-maven) for more.

## Using the Maven Publish action

To use the Maven action you need to add the step in your release workflow that invokes it.

Before using the Maven publish action, you should have a workflow that invokes the [Maven builder](https://github.com/slsa-framework/slsa-github-generator/blob/main/internal/builders/maven/README.md). It will look something like this:

```yaml
name: Release Maven project
on:
  - workflow_dispatch

permissions: read-all

jobs:
  build:
    permissions:
      id-token: write
      contents: read
      actions: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_maven_slsa3.yml@v2.0.0
    with:
      rekor-log-public: true
```

To use the Publish action, you need to add another job:

```yaml
publish:
  runs-on: ubuntu-latest
  needs: build
  permissions:
    id-token: write
    contents: read
    actions: read
  steps:
    - name: publish
      id: publish
      uses: slsa-framework/slsa-github-generator/actions/maven/publish@v2.0.0
      with:
        provenance-download-name: "${{ needs.build.outputs.provenance-download-name }}"
        provenance-download-sha256: "${{ needs.build.outputs.provenance-download-sha256 }}"
        target-download-name: "${{ needs.build.outputs.target-download-name }}"
        target-download-sha256: "${{ needs.build.outputs.target-download-sha256 }}"
        maven-username: ${{ secrets.OSSRH_USERNAME }}
        maven-password: ${{ secrets.OSSRH_PASSWORD }}
        gpg-key-pass: ${{ secrets.GPG_PASSPHRASE }}
        gpg-private-key: ${{ secrets.GPG_PRIVATE_KEY }}
```

Set the values of "maven-username", "maven-password", "gpg-key-pass" and " gpg-private-key" for your account. The parameters to `provenance-download-name`, `provenance-download-sha256`, `target-download-name`, and `target-download-sha256` should not be changed.

Once you trigger this workflow, your artifacts and provenance files will be added to a staging repository in Maven Central. You need to close the staging repository and then release:

Closing the staging repository:

![closing the staging repository](/actions/gradle/publish/images/gradle-publisher-staging-repository.png)

Releasing:

![releasing the Gradle artefacts](/actions/gradle/publish/images/gradle-publisher-release-closed-repository.png)
