# Internal Action Development

## External Actions
The following Actions:
- detect-workflow
- privacy-check
- rng
- secure-builder-checkout
- generate-builder

are considered "external" even though they are hosted on the same repository: they are not called via:

```././github/actions/name```

but instead via their "fully-qualified" name:

```slsa-framework/slsa-github-generator/.github/actions/name@vX.Y.Z```. 

We do this because the Actions are part of the builder, whereas the workflow runs in the "context" of the calling repository.

These Action *MUST* be pinned with the release tag for consistency.

## Internal Actions

Other Actions are called via:

```././github/actions/name```

and always require a checkout of the builder repository before being called.
The `secure-builder-checkout` is always used to checkout the builder repository
at `__BUILDER_CHECKOUT_DIR__` location. The `secure-project-checkout-*` checkout
the project to build at the location `__PROJECT_CHECKOUT_DIR__`.

These Actions are *composite actions*. They invoke scripts and also call other Actions.

## Development

To create or update an internal Action, follow the 2 following steps:

1. Create / modify the Action under `./github/actions/<your-action>` and get the changes merged. Let's call the resulting
commit hash after merge `CH`. (Note: This won't affect any workflow's behavior since
the existing code will still be calling the Action at an older commit hash).

1. Update the re-usable workflow / Actions to use them in a follow-up PR:
```yaml
uses: slsa-framework/slsa-github-generator/.github/actions/<your-action>@CH
```

  You can update using the following command:

```shell
find .github/ -name '*.yaml' -o -name '*.yml' | xargs sed -i 's/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/\(.*\)@[a-f0-9]*/uses: slsa-framework\/slsa-github-generator\/.github\/actions\/\1@_YOUR_CH__/'
```