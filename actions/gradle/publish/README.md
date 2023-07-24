# Publishing SLSA3+ provenance to Maven Central

This document explains how to publish SLSA3+ artifacts and provenance to Maven central.

The publisher is in its early stages and is likely to develop over time. Future breaking changes will occur.

To get started with publishing artifacts to Maven Central Repository, see [this guide](https://maven.apache.org/repository/guide-central-repository-upload.html).

Before you use the SLSA Gradle publisher, you will need to configure your Github project with the correct secrets. See [this guide](https://docs.github.com/en/actions/publishing-packages/publishing-java-packages-with-gradle) for more.

Your project needs to be already set up with Gradle and must have a gradle wrapper file in order to use the Gradle publisher.

The Gradle publisher expects you to have built the artifacts using the SLSA Gradle builder and that the provenance is available in `./build/libs/slsa-attestations/`.
