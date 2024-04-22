# Publishing SLSA3+ provenance to Maven Central

This document explains how to publish SLSA3+ artifacts and provenance to Maven central.

The publish Action is in its early stages and is likely to develop over time. Future breaking changes may occur.

To get started with publishing artifacts to Maven Central Repository, see [this guide](https://maven.apache.org/repository/guide-central-repository-upload.html).

Before you use this publish Action, you will need to configure your Github project with the correct secrets. See [this guide](https://docs.github.com/en/actions/publishing-packages/publishing-java-packages-with-gradle) for more.

Your project needs to be already set up with Gradle and must have a gradle wrapper file in order to use the Action.

The Action expects you to have built the artifacts using the SLSA Gradle builder and that the provenance is available in `./build/libs/slsa-attestations/`.

## Using the Gradle Publish action

To use the Gradle action you need to:

1. Modify your `build.gradle.kts` file.
2. Add the step in your release workflow that invokes it.

### Modify your `build.gradle.kts` file

Assuming you have already configured your Gradle repository to release to Maven Central, your `build.gradle.kts` looks something like this:

```kotlin
import java.io.File

plugins {
    `java-library`
    `maven-publish`
    `signing`
}

repositories {
    mavenLocal()
    maven {
        url = uri("https://repo.maven.apache.org/maven2/")
    }
}

group = "io.github.adamkorcz"
version = "0.1.18"
description = "Adam's test java project"
java.sourceCompatibility = JavaVersion.VERSION_1_8

java {
    withSourcesJar()
    withJavadocJar()
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            artifactId = "test-java-project"
            from(components["java"])
            
            pom {
                name.set("test-java-project")
                description.set("Adam's test java project")
                url.set("https://github.com/AdamKorcz/test-java-project")
                licenses {
                    license {
                        name.set("MIT License")
                        url.set("http://www.opensource.org/licenses/mit-license.php")
                    }
                }
                developers {
                    developer {
                        id.set("adamkrocz")
                        name.set("Adam K")
                        email.set("Adam@adalogics.com")
                    }
                }
                scm {
                    connection.set("scm:git:git://github.com/adamkorcz/test-java-project.git")
                    developerConnection.set("scm:git:ssh://github.com:simpligility/test-java-project.git")
                    url.set("http://github.com/adamkorcz/test-java-project/tree/main")
                }
            }
        }
    }
    repositories {
        maven {
            credentials {
                username = System.getenv("MAVEN_USERNAME")
                password = System.getenv("MAVEN_PASSWORD")
            }
            name = "test-java-project"
            url = uri("https://s01.oss.sonatype.org/service/local/staging/deploy/maven2/")
        }
    }
}

signing {
    useGpgCmd()
    sign(publishing.publications["maven"])
}
```

You need to add the following lines to your `build.gradle.kts` at the top inside of `create<MavenPublication>("maven")`:

```kotlin
val base_dir = "build/libs/slsa-attestations"
File(base_dir).walkTopDown().forEach {
    if (it.isFile()) {
        var path = it.getName()
        val name = path.replace(project.name + "-" + project.version, "").split(".", limit=2)
        if (name.size != 2) {
            throw StopExecutionException("Found incorrect file name: " + path)
        }
        var cls = name[0]
        var ext = name[1]
        if (cls.startsWith("-")) {
            cls = cls.substring(1)
        }
        artifact (base_dir + "/" + path) {
            classifier = cls
            extension = ext
        }
    }
}
```

Your final `build.gradle.kts` file should look like this:

```kotlin
import java.io.File

plugins {
    `java-library`
    `maven-publish`
    `signing`
}

repositories {
    mavenLocal()
    maven {
        url = uri("https://repo.maven.apache.org/maven2/")
    }
}

group = "io.github.adamkorcz"
version = "0.1.18"
description = "Adams test java project"
java.sourceCompatibility = JavaVersion.VERSION_1_8

java {
    withSourcesJar()
    withJavadocJar()
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            artifactId = "test-java-project"
            from(components["java"])
            val base_dir = "build/libs/slsa-attestations"
            File(base_dir).walkTopDown().forEach {
                if (it.isFile()) {
                    var path = it.getName()
                    val name = path.replace(project.name + "-" + project.version, "").split(".", limit=2)
                    if (name.size != 2) {
                        throw StopExecutionException("Found incorrect file name: " + path)
                    }
                    var cls = name[0]
                    var ext = name[1]
                    if (cls.startsWith("-")) {
                        cls = cls.substring(1)
                    }
                    artifact (base_dir + "/" + path) {
                        classifier = cls
                        extension = ext
                    }
                }
            }            
            pom {
                name.set("test-java-project")
                description.set("Adams test java project")
                url.set("https://github.com/AdamKorcz/test-java-project")
                licenses {
                    license {
                        name.set("MIT License")
                        url.set("http://www.opensource.org/licenses/mit-license.php")
                    }
                }
                developers {
                    developer {
                        id.set("adamkrocz")
                        name.set("Adam K")
                        email.set("Adam@adalogics.com")
                    }
                }
                scm {
                    connection.set("scm:git:git://github.com/adamkorcz/test-java-project.git")
                    developerConnection.set("scm:git:ssh://github.com:simpligility/test-java-project.git")
                    url.set("http://github.com/adamkorcz/test-java-project/tree/main")
                }
            }
        }
    }
    repositories {
        maven {
            credentials {
                username = System.getenv("MAVEN_USERNAME")
                password = System.getenv("MAVEN_PASSWORD")
            }
            name = "test-java-project"
            url = uri("https://s01.oss.sonatype.org/service/local/staging/deploy/maven2/")
        }
    }
}

signing {
    useGpgCmd()
    sign(publishing.publications["maven"])
}
```

You don't need to configure anything inside that code snippet; Adding them to your `build.gradle.kts` file is enough.

### Add the publish action to your release workflow

Before using the Gradle publish action, you should have a workflow that invokes the Gradle builder. It will look something like this:

```yaml
name: Publish Gradle with action
on:
  - workflow_dispatch

permissions: read-all

jobs:
  build:
    permissions:
      id-token: write
      contents: read
      actions: read
      packages: read
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_gradle_slsa3.yml@v2.0.0
    with:
      rekor-log-public: true
      artifact-list: build/libs/artifact1-0.1.18.jar,build/libs/artifact-0.1.18-javadoc.jar,build/libs/artifact-0.1.18-sources.jar
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
      uses: slsa-framework/slsa-github-generator/actions/gradle/publish@v2.0.0
      with:
        provenance-download-name: "${{ needs.build.outputs.provenance-download-name }}"
        provenance-download-sha256: "${{ needs.build.outputs.provenance-download-sha256 }}"
        build-download-name: "${{ needs.build.outputs.build-download-name }}"
        build-download-sha256: "${{ needs.build.outputs.build-download-sha256 }}"
        maven-username: ${{ secrets.OSSRH_USERNAME }}
        maven-password: ${{ secrets.OSSRH_PASSWORD }}
        gpg-key-pass: ${{ secrets.GPG_PASSPHRASE }}
        gpg-private-key: ${{ secrets.GPG_PRIVATE_KEY }}
        jdk-version: "17"
```

Set the values of "maven-username", "maven-password", "gpg-key-pass" and " gpg-private-key" for your account. The parameters to `provenance-download-name`, `provenance-download-sha256`, `target-download-name`, and `target-download-sha256` should not be changed.

Once you trigger this workflow, your artifacts and provenance files will be added to a staging repository in Maven Central. You need to close the staging repository and then release:

Closing the staging repository:

![closing the staging repository](/actions/gradle/publish/images/gradle-publisher-staging-repository.png)

Releasing:

![releasing the Gradle artefacts](/actions/gradle/publish/images/gradle-publisher-release-closed-repository.png)

### Multi-Project Builds

See the same guidance in the [build docs](../../../internal/builders/gradle/README.md#multi-project-builds) for consolidating files from multi-project builds.
