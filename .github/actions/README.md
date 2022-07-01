# Internal Action Development

## Internal Actions
Although the Actions are hosted on the same repository, we consider them "external": they are not called via:

```././github/actions/name```

but instead via their "fully-qualified" name:

```slsa-framework/slsa-github-generator/.github/actions/name@hash```. 

We do this because the Actions are part of the builder, whereas the workflow runs in the "context" of the calling repository.

## Checkout Rules
Actions that are called with a copy of the calling repository on disk (`actions/checkout` for the calling repository)
should *NEVER* "checkout" the builder's repository, because it creates interference with the calling repository
and is difficult to get right.
    
In particular, *composite actions* need to use inline scripts and cannot invoke script files stored in the builder's repository.

In general, Actions that need to "checkout" their code should use Dockerfile or nodejs-type projects. An example of such an Action
is the `./github/actions/detect-workflow` Action.

There is one exception today: the `./github/actions/generate-builder` Action. It "checkouts" its own code and is allowed to do it
because it does so in a job that never "checkouts" the calling repository. (Note: the code will be migrated to 
a Dockerfile or nodejs-type projects in the futue).

## Development

To create or update an internal Action, follow the 2 following steps:

1. Create / modify the Action under `./github/actions/<your-action>` and get the changes merged. Let's call the resulting
commit hash after merge `CH`. (his won't affect the existing code even if the behavior of the Action has changed, since
the existing code will still be calling the Action at an older commit hash).

1. Update the re-usable workflow / Actions to use them in a follow-up PR:
```yaml
uses: slsa-framework/slsa-github-generator/.github/actions/<your-action>@CH
```

  You can update using the following command:

```shell
find .github/ -name '*.yaml' -o -name '*.yml' | xargs sed -i 's/uses: slsa-framework\/slsa-github-generator\/\.github\/actions\/\(.*\)@[a-f0-9]*/uses: slsa-framework\/slsa-github-generator\/.github\/actions\/\1@YOUR_HASH/'
```