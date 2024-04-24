# Generation of SLSA3+ provenance for Node.js packages

This document explains how to generate SLSA provenance for Node.js packages.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Node.js builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Publishing packages](#publishing-packages)
    - [Using the `nodejs/publish` action](#using-the-nodejspublish-action)
    - [Custom publishing](#custom-publishing)
  - [Referencing the Node.js builder](#referencing-the-nodejs-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Verification](#verification)
  - [npm audit signatures](#npm-audit-signatures)
  - [slsa-verifier](#slsa-verifier)
  - [Known issues](#known-issues)
    - [Workspaces are not supported](#workspaces-are-not-supported)
    - [Other package managers not supported](#other-package-managers-not-supported)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Node.js builder will generate a non-forgeable attestation to the
Node.js package using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the package they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Node.js builder is currently in beta. The API could change while approaching
a Generally Available (GA) release. You can track progress towards General
Availability via the
[Node.js Builder GA milestone](https://github.com/slsa-framework/slsa-github-generator/milestone/17).

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Generating Provenance

The Node.js builder uses a Github Actions reusable workflow to build your
package and generate the provenance.

### Getting Started

Let's assume you have a `package.json` for your TypeScript project that looks
something like this. Here we have a `build` script that runs `tsc` to compile
our package's TypeScript code.

We also define a `ci` script that runs `npm ci`. We will see later that define
a `ci` script so we can install our `devDependencies` before running our
`build` script.

In order to publish to `npmjs.com`, packages should define a `repository`
field. This should match the repository used to run the Node.js builder.

```json
{
  "name": "mypackage",
  "version": "0.0.1",
  "description": "My Package",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": ["/dist"],
  "scripts": {
    "ci": "npm ci",
    "build": "tsc",
    "test": "jest"
  },
  "devDependencies": {
    "@types/jest": "^29.4.0",
    "jest": "^29.4.3",
    "ts-jest": "^29.0.5",
    "typescript": "^4.8.4"
  },
  "repository": {
    "type": "git",
    "url": "git+https://github.com/ianlewis/mypackage.git"
  }
}
```

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.
This assumes that the `package.json` is in the root directory of your
repository.

The following reusable workflow call will build the package into a tarball and
generate provenance attestations which will be uploaded as artifacts to the
workflow run.

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@v2.0.0
    with:
      run-scripts: "ci, test, build"
```

The `run-scripts` are a set of comma separated build scripts that are run to
perform the build. This should include a step to install development
dependencies, compile any code, run tests, etc. The scripts are run in the order
they are listed.

Once the build scripts are run, the Node.js builder creates a package tarball
and provenance attestation which are uploaded as artifacts to the workflow run.

### Publishing packages

#### Using the `nodejs/publish` action

After creating the package you can publish the package using the provided
`nodejs/publish` action.

```yaml
publish:
  needs: [build]
  runs-on: ubuntu-latest
  steps:
    - name: Set up Node registry authentication
      uses: actions/setup-node@64ed1c7eab4cce3362f8c340dee64e5eaeef8f7c # v3.6.0
      with:
        node-version: 18
        registry-url: "https://registry.npmjs.org"

    - name: publish
      id: publish
      uses: slsa-framework/slsa-github-generator/actions/nodejs/publish@v2.0.0
      with:
        access: public
        node-auth-token: ${{ secrets.NPM_TOKEN }}
        package-name: ${{ needs.build.outputs.package-name }}
        package-download-name: ${{ needs.build.outputs.package-download-name }}
        package-download-sha256: ${{ needs.build.outputs.package-download-sha256 }}
        provenance-name: ${{ needs.build.outputs.provenance-name }}
        provenance-download-name: ${{ needs.build.outputs.provenance-download-name }}
        provenance-download-sha256: ${{ needs.build.outputs.provenance-download-sha256 }}
```

This action downloads the package tarball and provenance before running `npm
publish` to publish your package to the npm registry. We provide a
`node-auth-token` so that we can authenticate with `npmjs.com`.

See the full documentation for the
[`publish` action](../../../actions/nodejs/publish/README.md) for more
information.

#### Custom publishing

After the package has been built you can publish on your own by downloading the
package archive and provenance attestations and running your own custom
publishing command.

Here is an example:

```yaml
jobs:
  # build job etc. ...

  publish:
    needs: [build]
    runs-on: ubuntu-latest
    steps:
      - name: Setup Node
        uses: actions/setup-node@e33196f7422957bea03ed53f6fbb155025ffc7b8 # v3.7.0
        with:
          node-version: 18
          registry-url: "https://registry.npmjs.org"

      - name: Download tarball
        uses: slsa-framework/slsa-github-generator/actions/nodejs/secure-package-download@v2.0.0
        with:
          name: ${{ needs.build.outputs.package-download-name }}
          path: ${{ needs.build.outputs.package-name }}
          sha256: ${{ needs.build.outputs.package-download-sha256 }}

      - name: Download provenance
        uses: slsa-framework/slsa-github-generator/actions/nodejs/secure-attestations-download@v2.0.0
        with:
          name: ${{ needs.build.outputs.provenance-download-name }}
          path: "attestations"
          sha256: ${{ needs.build.outputs.provenance-download-sha256 }}

      - name: Publish the package
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
          TARBALL_PATH: "${{ needs.build.outputs.package-name }}"
          PROVENANCE_PATH: "./attestations/${{ needs.build.outputs.provenance-name }}"
        run: |
          npm publish "${TARBALL_PATH}" --access=public --provenance-file="${PROVENANCE_PATH}"
```

You will need a package management tool that supports providing the provenance
file. Currently [npm], [lerna] or [pnpm] can support this.

See the full documentation for the
[`secure-attestations-download` action](../../../actions/nodejs/secure-attestations-download/README.md)
and
[`secure-package-download` action](../../../actions/nodejs/secure-package-download/README.md)
for more information.

### Referencing the Node.js builder

At present, the builder **MUST** be referenced by a tag of the form `@vX.Y.Z`,
because the build will fail if you reference it via a shorter tag like `@vX.Y`
or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure
renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor] API server instance at
https://rekor.sigstore.dev/. This entry includes the repository name. This will cause the
private repository name to leak and be discoverable via the public Rekor API
server.

If this is ok with you, you can set the `rekor-log-public` flag in order to
opt in to publishing to the public Rekor instance from a private repository.

```yaml
with:
  rekor-log-public: true
```

If you do not set this flag then private repositories will generate an error in
order to prevent leaking repository name information.

Support for private transparency log instances that would not leak repository
name information is tracked on [issue #372](https://github.com/slsa-framework/slsa-github-generator/issues/372).

### Supported Triggers

Only the following [event types] are supported:

| Supported event type  | Event description                          |
| --------------------- | ------------------------------------------ |
| [`create`]            | Creation of a git tag or branch.           |
| [`release`]           | Creation or update of a GitHub release.    |
| [`push`]              | Creation or update of a git tag or branch. |
| [`workflow_dispatch`] | Manual trigger of a workflow.              |

`pull_request` events are currently not supported. If you would like support for
`pull_request`, please tell us about your use case on
[issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a
[new issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

[event types]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
[`create`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#create
[`release`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#release
[`push`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#push
[`workflow_dispatch`]: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch

### Workflow Inputs

The Node.js builder accepts the following inputs:

Inputs:

| Name              | Required | Default            | Description                                                                                                                                                                                                                                         |
| ----------------- | -------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| directory         | No       | `github.workspace` | The root directory of the package (i.e. where the `package.json` is located)                                                                                                                                                                        |
| node-version      | No       |                    | The version of Node.js to use. If no value is supplied, the `node` version from `$PATH` is used.                                                                                                                                                    |
| node-version-file | No       |                    | File containing the version Spec of the version to use. Examples: .nvmrc, .node-version, .tool-versions.                                                                                                                                            |
| rekor-log-public  | No       | false              | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false` |
| run-scripts       | No       |                    | A comma separated ordered list of npm scripts to run before running `npm publish`. See [scripts] for more information. \                                                                                                                            |

### Workflow Outputs

The Node.js builder produces the following outputs:

| Name                       | Description                                                            |
| -------------------------- | ---------------------------------------------------------------------- |
| package-name               | The file name of the package tarball in the upload artifact.           |
| package-download-name      | The name of the package artifact uploaded to the workflow run.         |
| package-download-sha256    | The sha256 of the package artifact uploaded to the workflow run.       |
| provenance-name            | The file name of the provenance attestation upload artifact.           |
| provenance-download-name   | The name of the provenance attestation uploaded to the workflow run.   |
| provenance-download-sha256 | The sha256 of the provenance attestation uploaded to the workflow run. |

### Provenance Format

Provenance is generated as an [in-toto] statement with a
SLSA v0.2 predicate.

| Name           | Value                | Description                                          |
| -------------- | -------------------- | ---------------------------------------------------- |
| `subject.name` | Package url ([purl]) | The subject identifies the package in [purl] format. |

The project generates SLSA v0.2 provenance predicate with the following values.

| Name                         | Value                                                                                                                  | Description                                                                                                                                                                                                            |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `builder.id`                 | `https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@refs/tags/v1.5.0"` | Identifies the Node.js builder                                                                                                                                                                                         |
| `buildType`                  | `"https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0"`                                       | Identifies a the GitHub Actions build.                                                                                                                                                                                 |
| `metadata.buildInvocationID` | `"[run_id]-[run_attempt]"`                                                                                             | The GitHub Actions [`run_id`](https://docs.github.com/en/actions/learn-github-actions/contexts#github-context) does not update when a workflow is re-run. Run attempt is added to make the build invocation ID unique. |

### Provenance Example

The following is an example of the generated provenance.

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "subject": [
    {
      "name": "pkg:npm/%40ianlewis/actions-test@0.1.77",
      "digest": {
        "sha512": "3bbabe6803f7f9fd8a22dd3e85ae015664a75f277459d3bc2fc6aacd7054eb8af432188d92fca653af49460f9209562c0d298b959b169db98d369c4dcfe8406e"
      }
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_nodejs_slsa3.yml@refs/tags/v1.5.0"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/tags/v0.1.77",
        "digest": {
          "sha1": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91"
        },
        "entryPoint": ".github/workflows/nodejs.yml"
      },
      "parameters": {
        "inputs": {
          "access": "",
          "directory": "nodejs",
          "node-version": "16",
          "node-version-file": "",
          "rekor-log-public": false,
          "run-scripts": "build",
          "dist-tag": "latest"
        }
      },
      "environment": {
        "GITHUB_ACTOR_ID": "49289",
        "GITHUB_EVENT_NAME": "push",
        "GITHUB_REF": "refs/tags/v0.1.77",
        "GITHUB_REF_TYPE": "tag",
        "GITHUB_REPOSITORY": "ianlewis/actions-test",
        "GITHUB_REPOSITORY_ID": "474793590",
        "GITHUB_REPOSITORY_OWNER_ID": "49289",
        "GITHUB_RUN_ATTEMPT": "1",
        "GITHUB_RUN_ID": "4527998016",
        "GITHUB_RUN_NUMBER": "64",
        "GITHUB_SHA": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91",
        "GITHUB_TRIGGERING_ACTOR_ID": "49289",
        "GITHUB_WORKFLOW_REF": "ianlewis/actions-test/.github/workflows/nodejs.yml@refs/tags/v0.1.77",
        "GITHUB_WORKFLOW_SHA": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91",
        "IMAGE_OS": "ubuntu22",
        "IMAGE_VERSION": "20230317.1",
        "RUNNER_ARCH": "X64",
        "RUNNER_NAME": "GitHub Actions 12",
        "RUNNER_OS": "Linux"
      }
    },
    "metadata": {
      "buildInvocationId": "4527998016-1",
      "completeness": {
        "parameters": true
      }
    },
    "materials": [
      {
        "uri": "git+https://github.com/ianlewis/actions-test@refs/tags/v0.1.77",
        "digest": {
          "sha1": "04c0efe26488a0c5ca8f4404c41c4b96b6e9cf91"
        }
      }
    ]
  }
}
```

## Verification

Verification of provenance generated by the Node.js builder can be done via two
methods.

### npm audit signatures

The `npm` CLI includes an `audit signatures` command which will print
information about the availability of registry signatures and SLSA attestations.
This prints the number of installed dependency packages that have verified SLSA
attestations that have passed the official npm registry's verification at
package upload.

```shell
$ npm audit signatures
audited 1 package in 3s

1 package has a verified registry signature

1 package has a verified attestation
```

Note that the `npm audit signatures` command will succeed regardless of the
presence of signatures or attestations. You will need to check that the output
matches your expectations.

### slsa-verifier

The `slsa-verifier` tool includes support for Node.js packages published on the
official npm registry and provides the most flexibility in verifying the
content against an expected `builder.id`, source repository, and source tag,
package name, and package version.

Please see the [documentation](https://github.com/slsa-framework/slsa-verifier)
for more information.

### Known issues

#### Workspaces are not supported

[Workspaces] are currently not supported but will be supported in a future
release. See
[#1789](https://github.com/slsa-framework/slsa-github-generator/issues/1789) for
more details.

#### Other package managers not supported

Currently the Node.js builder does not support using other package managers like
[yarn], [pnpm], or [lerna] for building.

Currently [lerna] and [pnpm] can support publishing. See
[Custom publishing](#custom-publishing) for more details.

[Yarn] implements publishing on it's own and requires support for the
`provenance` and `provenaceFile` config options. See
[yarnpkg/berry#5430](https://github.com/yarnpkg/berry/issues/5430).

[in-toto]: https://in-toto.io/
[rekor]: https://github.com/sigstore/rekor
[purl]: https://github.com/package-url/purl-spec
[scripts]: https://docs.npmjs.com/cli/v9/using-npm/scripts
[workspaces]: https://docs.npmjs.com/cli/v9/using-npm/workspaces
[npm]: https://www.npmjs.com/package/npm
[yarn]: https://yarnpkg.com/
[pnpm]: https://pnpm.io/
[lerna]: https://lerna.js.org/
