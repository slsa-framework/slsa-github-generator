# Docker-based builder

This folder contains a command line tool for building artifacts using a Docker image.

It is meant to be used as part of a GitHub Actions reusable workflow for
generating SLSA provenances. However, users can also run the command locally.

The command line tool provides two sub-commands, namely `dry-run` and `build`.

## The `dry-run` subcommand

The `dry-run` subcommand can be used to validate the inputs. If the inputs are
valid, then the tool creates a `BuildDefinition` and stores that as a JSON
document in the output path that must be provided as one of the flags to the
command. The following is an example, which assumes you are running the code in
`internal/builders/docker`:

```bash
go run *.go  dry-run \
  --build-config-path testdata/config.toml \
  --build-definition-path bd.json \
  --builder-image bash@sha256:9e2ba52487d \
  --git-commit-hash sha1:9b5f98310dbbad675834474fa68c37d880687cb9 \
  --source-repo git+https://github.com/project-oak/transparent-release
```

The output of this is a JSON document stored in `bd.json`.

## The `build` subcommand
 
The `build` subcommand takes more or less the same inputs as the `dry-run`
subcommand, but actually builds the artifacts. To successfully run this
command, you need to have [rootless Docker installed](https://docs.docker.com/engine/security/rootless/).

The following is an example:

```bash
go run *.go build \
  --build-config-path internal/builders/docker/testdata/config.toml \
  --builder-image bash@sha256:9e2ba52487d945504d250de186cb4fe2e3ba023ed2921dd6ac8b97ed43e76af9 \
  --git-commit-hash sha1:cf5804b5c6f1a4b2a0b03401a487dfdfbe3a5f00 \
  --source-repo git+https://github.com/slsa-framework/slsa-github-generator \
  --force-checkout
```

If the build is successful, this command will generate and output a list of
generated artifacts and their SHA256 digests.