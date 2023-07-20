# Maven verification plugin

The Maven verification plugin can be used to verify the provenance of the dependencies of a Java project.

It is meant to make it easy for project owners and consumers to:
1: Check how many and which dependencies of a Maven-based project are released with provenance files.
2: Verify the provenance files of the dependencies of a given Maven-based project.

The plugin wraps the [the slsa verifier](https://github.com/slsa-framework/slsa-verifier) and invokes it for all the dependencies in a `pom.xml`.

## Prerequisites

To use the plugin you must have Go, Java and Maven installed. It has currently only been tested on Ubuntu.

The plugin requires that the slsa-verifier is already installed on the machine.

## Development status

The plugin is in its early stages and is not ready for production.

Things that work well are:
1: Resolving dependencies and checking whether they have provenance files in the remote repository.
2: Running the slsa-verifier against dependencies with provenance files.
3: Outputting the result from the slsa-verifier.

Things that are unfinished:
1: What to do with the results from the verifier. Currently we have not taken a stand on what the Maven verification plugin should do with the output from the slsa-verifier. This is a UX decision more than it is a technical decision.

## Using the Maven verification plugin

### Invoking it directly

It can be run from the root of a given project file.
A pseudo-workflow looks like this:
  1: `git clone --depth=1 https://github.com/slsa-framework/slsa-github-generator`
  2: `cd slsa-github-generator/internal/builders/maven/plugins/verification-plugin`
  3: `mvn clean install`
  4: `cd /tmp`
  5: `git clone your repository to text`
  6: `cd into your repository`
  7: `mvn io.github.adamkorcz:slsa-verification-plugin:0.0.1:verify`

The plugin will now go through all the dependencies in the `pom.xml` file and check if they have a provenance statement attached to their release. If a dependency has a SLSA provenance file, the Maven verification plugin will fetch it from the remote repository and invoke the `slsa-verifier` binary against the dependency and the provenance file.

### Integrating it into your Maven build cycle

The plugin can also live in your Maven build cycle. If you add it to your own `pom.xml`, the plugin will execute during the validation phase of the Maven build cycle.
