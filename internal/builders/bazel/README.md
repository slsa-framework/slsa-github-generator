# Generation of SLSA3+ provenance for Bazel builds

This document explains how to generate SLSA provenance for Bazel packages.

This can be done by adding a step to your Github Actions workflow to call a
[reusable
workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
to build the package and generate SLSA provenance. We'll call this
workflow the "Bazel builder" from now on.

---

<!-- markdown-toc --bullets="-" -i README.md -->

<!-- toc -->

- [Benefits of Provenance](#benefits-of-provenance)
- [Development status](#development-status)
- [Generating Provenance](#generating-provenance)
  - [Getting Started](#getting-started)
  - [Referencing the Bazel builder](#referencing-the-bazel-builder)
  - [Private Repositories](#private-repositories)
  - [Supported Triggers](#supported-triggers)
  - [Workflow Inputs](#workflow-inputs)
  - [Workflow Outputs](#workflow-outputs)
  - [Provenance Format](#provenance-format)
  - [Provenance Example](#provenance-example)
- [Verification](#verification)
  - [npm audit signatures](#npm-audit-signatures)
  - [slsa-verifier](#slsa-verifier)

<!-- tocstop -->

---

## Benefits of Provenance

Using the Bazel builder will generate a non-forgeable attestation to the
Bazel build artifacts produced using the identity of the GitHub workflow. This can be used to
create a positive attestation to a package coming from your repository.

That means that once your users verify the artifact they have downloaded they can
be sure that it was created by your repository's workflow and hasn't been
tampered with.

## Development status

The Bazel builder is currently in beta. The API could change while approaching
a Generally Available (GA) release.

Please try it out and
[create an issue](https://github.com/slsa-framework/slsa-github-generator/issues/new)
to send us feedback!

## Generating Provenance

The Bazel builder uses a Github Actions reusable workflow to build your
artifact(s) and generate the provenance.

### Getting Started

Let's assume you have a Bazel project includes a
[WORKSPACE file](https://bazel.build/reference/glossary#workspace-file)
at the root of your repository as well as one or more
[BUILD files](https://bazel.build/reference/glossary#build-file). Also included
at the root of the directory is the
[.bazelversion file](https://github.com/bazelbuild/bazelisk#how-does-bazelisk-know-which-bazel-version-to-run)
, [.bazelrc file](https://bazel.build/run/bazelrc)
, and a
[MODULE.bazel file](https://bazel.build/rules/lib/globals/module).

To get started, you will need to add some steps to your current workflow. We
will assume you have an existing Github Actions workflow to build your project.
This assumes that the files above are in the root directory of your
repository.

We also define the targets to be built as 'target1' and 'target2', and
we define the flag of the build to be '--strip=always'

The following reusable workflow call will build the targets and upload the artifacts into a .zip and
generate provenance attestations which will be uploaded as another .zip to the
workflow run.

```yaml
jobs:
  build:
    permissions:
      id-token: write # For signing
      contents: read # For repo checkout.
      actions: read # For getting workflow run info.
    if: startsWith(github.ref, 'refs/tags/')
    uses: slsa-framework/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@v1.6.0
    with:
      targets: "//src:target1 //src:target2"
      flags: "--strip=always"
```

The `targets` are a set of space separated build targets to be built.

Once the targets are built, the Bazel builder creates a folder for the artifacts
and another for the provenance attestations which are uploaded as artifacts to the workflow run.

### Referencing the Node.js builder

At present, the builder **MUST** be referenced by a tag of the form `@vX.Y.Z`,
because the build will fail if you reference it via a shorter tag like `@vX.Y`
or `@vX` or if you reference it by a hash.

For more information about this design decision and how to configure
renovatebot, see the main repository [README.md](../../../README.md).

### Private Repositories

Private repositories are supported with some caveats. Currently all builds
generate and post a new entry in the public
[Rekor](https://github.com/sigstore/rekor) API server instance at
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

The following [GitHub trigger
events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows)
are fully supported and tested:

- `schedule`
- `push` (including new tags)
- `release`
- Manual run via `workflow_dispatch`

However, in practice, most triggers should work with the exception of
`pull_request`. If you would like support for `pull_request`, please tell us
about your use case on
[issue #358](https://github.com/slsa-framework/slsa-github-generator/issues/358). If
you have an issue in all other triggers please submit a
[new issue](https://github.com/slsa-framework/slsa-github-generator/issues/new/choose).

### Workflow Inputs

The Bazel builder accepts the following inputs:

Inputs:

| Name              | Required | Default            | Description                                                                                                                                                                                                                                         |
| ----------------- | -------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| rekor-log-public  | No       | false              | Set to true to opt-in to posting to the public transparency log. Will generate an error if false for private repositories. This input has no effect for public repositories. See [Private Repositories](#private-repositories).<br>Default: `false` |
| targets           | Yes      |                    | A space separated list of targets to build and generate artifacts for. See [targets](https://bazel.build/concepts/build-ref#targets) for more information. \                                                                                        |
| flags             | No       | ""                 | A space separated list of flags to modify the build by. See [flags](https://bazel.build/docs/user-manual#build-options) for more information. \                                                                                                     |

### Workflow Outputs

The Bazel builder produces the following outputs:

| Name                       | Description                                                            |
| -------------------------- | ---------------------------------------------------------------------- |
| provenance-download-name   | The name of the provenance attestation uploaded to the workflow run.   |

### Provenance Format



Provenance is generated as an [in-toto](https://in-toto.io/) statement with a
SLSA v1 predicate.

| Name           | Value                                                          | Description                                                                                    |
| -------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `subject.name` | Artifact Name | The subject identifies the artifact built by the builder. The provenance attests the build process for the artifact listed.                     |

The project generates SLSA v1 provenance predicate with the following values.

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
      "name": "fib",
      "digest": {
        "sha256": "caaadba2846905ac477c777e96a636e1c2e067fdf6fed90ec9eeca4df18d6ed9"
      }
    }
  ],
  "predicateType": "https://slsa.dev/provenance/v1",
  "predicate": {
    "buildDefinition": {
      "buildType": "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0",
      "externalParameters": {
        "inputs": {
          "rekor-log-public": false,
          "targets": "//src:fib //src:hello",
          "flags": "--strip=always"
        },
        "vars": {}
      },
      "internalParameters": {
        "GITHUB_ACTOR_ID": "78953604",
        "GITHUB_EVENT_NAME": "workflow_dispatch",
        "GITHUB_REF": "refs/heads/main",
        "GITHUB_REF_TYPE": "branch",
        "GITHUB_REPOSITORY": "enteraga6/cc_user_test",
        "GITHUB_REPOSITORY_ID": "647514313",
        "GITHUB_REPOSITORY_OWNER_ID": "78953604",
        "GITHUB_RUN_ATTEMPT": "1",
        "GITHUB_RUN_ID": "5214546713",
        "GITHUB_RUN_NUMBER": "1",
        "GITHUB_SHA": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6",
        "GITHUB_TRIGGERING_ACTOR_ID": "78953604",
        "GITHUB_WORKFLOW_REF": "enteraga6/cc_user_test/.github/workflows/flag_test.yml@refs/heads/main",
        "GITHUB_WORKFLOW_SHA": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6",
        "GITHUB_EVENT_PAYLOAD": {
          "inputs": null,
          "ref": "refs/heads/main",
          "repository": {
            "allow_forking": true,
            "archive_url": "https://api.github.com/repos/enteraga6/cc_user_test/{archive_format}{/ref}",
            "archived": false,
            "assignees_url": "https://api.github.com/repos/enteraga6/cc_user_test/assignees{/user}",
            "blobs_url": "https://api.github.com/repos/enteraga6/cc_user_test/git/blobs{/sha}",
            "branches_url": "https://api.github.com/repos/enteraga6/cc_user_test/branches{/branch}",
            "clone_url": "https://github.com/enteraga6/cc_user_test.git",
            "collaborators_url": "https://api.github.com/repos/enteraga6/cc_user_test/collaborators{/collaborator}",
            "comments_url": "https://api.github.com/repos/enteraga6/cc_user_test/comments{/number}",
            "commits_url": "https://api.github.com/repos/enteraga6/cc_user_test/commits{/sha}",
            "compare_url": "https://api.github.com/repos/enteraga6/cc_user_test/compare/{base}...{head}",
            "contents_url": "https://api.github.com/repos/enteraga6/cc_user_test/contents/{+path}",
            "contributors_url": "https://api.github.com/repos/enteraga6/cc_user_test/contributors",
            "created_at": "2023-05-31T00:27:44Z",
            "default_branch": "main",
            "deployments_url": "https://api.github.com/repos/enteraga6/cc_user_test/deployments",
            "description": null,
            "disabled": false,
            "downloads_url": "https://api.github.com/repos/enteraga6/cc_user_test/downloads",
            "events_url": "https://api.github.com/repos/enteraga6/cc_user_test/events",
            "fork": false,
            "forks": 0,
            "forks_count": 0,
            "forks_url": "https://api.github.com/repos/enteraga6/cc_user_test/forks",
            "full_name": "enteraga6/cc_user_test",
            "git_commits_url": "https://api.github.com/repos/enteraga6/cc_user_test/git/commits{/sha}",
            "git_refs_url": "https://api.github.com/repos/enteraga6/cc_user_test/git/refs{/sha}",
            "git_tags_url": "https://api.github.com/repos/enteraga6/cc_user_test/git/tags{/sha}",
            "git_url": "git://github.com/enteraga6/cc_user_test.git",
            "has_discussions": false,
            "has_downloads": true,
            "has_issues": true,
            "has_pages": false,
            "has_projects": true,
            "has_wiki": true,
            "homepage": null,
            "hooks_url": "https://api.github.com/repos/enteraga6/cc_user_test/hooks",
            "html_url": "https://github.com/enteraga6/cc_user_test",
            "id": 647514313,
            "is_template": false,
            "issue_comment_url": "https://api.github.com/repos/enteraga6/cc_user_test/issues/comments{/number}",
            "issue_events_url": "https://api.github.com/repos/enteraga6/cc_user_test/issues/events{/number}",
            "issues_url": "https://api.github.com/repos/enteraga6/cc_user_test/issues{/number}",
            "keys_url": "https://api.github.com/repos/enteraga6/cc_user_test/keys{/key_id}",
            "labels_url": "https://api.github.com/repos/enteraga6/cc_user_test/labels{/name}",
            "language": "C++",
            "languages_url": "https://api.github.com/repos/enteraga6/cc_user_test/languages",
            "license": null,
            "merges_url": "https://api.github.com/repos/enteraga6/cc_user_test/merges",
            "milestones_url": "https://api.github.com/repos/enteraga6/cc_user_test/milestones{/number}",
            "mirror_url": null,
            "name": "cc_user_test",
            "node_id": "R_kgDOJphIyQ",
            "notifications_url": "https://api.github.com/repos/enteraga6/cc_user_test/notifications{?since,all,participating}",
            "open_issues": 0,
            "open_issues_count": 0,
            "owner": {
              "avatar_url": "https://avatars.githubusercontent.com/u/78953604?v=4",
              "events_url": "https://api.github.com/users/enteraga6/events{/privacy}",
              "followers_url": "https://api.github.com/users/enteraga6/followers",
              "following_url": "https://api.github.com/users/enteraga6/following{/other_user}",
              "gists_url": "https://api.github.com/users/enteraga6/gists{/gist_id}",
              "gravatar_id": "",
              "html_url": "https://github.com/enteraga6",
              "id": 78953604,
              "login": "enteraga6",
              "node_id": "MDQ6VXNlcjc4OTUzNjA0",
              "organizations_url": "https://api.github.com/users/enteraga6/orgs",
              "received_events_url": "https://api.github.com/users/enteraga6/received_events",
              "repos_url": "https://api.github.com/users/enteraga6/repos",
              "site_admin": false,
              "starred_url": "https://api.github.com/users/enteraga6/starred{/owner}{/repo}",
              "subscriptions_url": "https://api.github.com/users/enteraga6/subscriptions",
              "type": "User",
              "url": "https://api.github.com/users/enteraga6"
            },
            "private": false,
            "pulls_url": "https://api.github.com/repos/enteraga6/cc_user_test/pulls{/number}",
            "pushed_at": "2023-06-08T18:28:37Z",
            "releases_url": "https://api.github.com/repos/enteraga6/cc_user_test/releases{/id}",
            "size": 13,
            "ssh_url": "git@github.com:enteraga6/cc_user_test.git",
            "stargazers_count": 0,
            "stargazers_url": "https://api.github.com/repos/enteraga6/cc_user_test/stargazers",
            "statuses_url": "https://api.github.com/repos/enteraga6/cc_user_test/statuses/{sha}",
            "subscribers_url": "https://api.github.com/repos/enteraga6/cc_user_test/subscribers",
            "subscription_url": "https://api.github.com/repos/enteraga6/cc_user_test/subscription",
            "svn_url": "https://github.com/enteraga6/cc_user_test",
            "tags_url": "https://api.github.com/repos/enteraga6/cc_user_test/tags",
            "teams_url": "https://api.github.com/repos/enteraga6/cc_user_test/teams",
            "topics": [],
            "trees_url": "https://api.github.com/repos/enteraga6/cc_user_test/git/trees{/sha}",
            "updated_at": "2023-05-31T00:36:58Z",
            "url": "https://api.github.com/repos/enteraga6/cc_user_test",
            "visibility": "public",
            "watchers": 0,
            "watchers_count": 0,
            "web_commit_signoff_required": false
          },
          "sender": {
            "avatar_url": "https://avatars.githubusercontent.com/u/78953604?v=4",
            "events_url": "https://api.github.com/users/enteraga6/events{/privacy}",
            "followers_url": "https://api.github.com/users/enteraga6/followers",
            "following_url": "https://api.github.com/users/enteraga6/following{/other_user}",
            "gists_url": "https://api.github.com/users/enteraga6/gists{/gist_id}",
            "gravatar_id": "",
            "html_url": "https://github.com/enteraga6",
            "id": 78953604,
            "login": "enteraga6",
            "node_id": "MDQ6VXNlcjc4OTUzNjA0",
            "organizations_url": "https://api.github.com/users/enteraga6/orgs",
            "received_events_url": "https://api.github.com/users/enteraga6/received_events",
            "repos_url": "https://api.github.com/users/enteraga6/repos",
            "site_admin": false,
            "starred_url": "https://api.github.com/users/enteraga6/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/enteraga6/subscriptions",
            "type": "User",
            "url": "https://api.github.com/users/enteraga6"
          },
          "workflow": ".github/workflows/flag_test.yml"
        }
      },
      "resolvedDependencies": [
        {
          "uri": "git+https://github.com/enteraga6/cc_user_test@refs/heads/main",
          "digest": {
            "gitCommit": "5d50da9ec44535cafb2e1e8bc250484983c7d0b6"
          }
        }
      ]
    },
    "runDetails": {
      "builder": {
        "id": "https://github.com/enteraga6/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml@refs/heads/bazel-builder"
      },
      "metadata": {
        "invocationId": "https://github.com/enteraga6/cc_user_test/actions/runs/5214546713/attempts/1"
      }
    }
  }
}
```

## Verification
TODO
