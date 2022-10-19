# Renovate Best Practices and SLSA-GitHub-Generator

Renovate helps users to enforce security best practices when continuously upgrading GitHub actions.

Renovate provides a configuration snippet, which is used by most GitHub projects, to [automatically pin dependencies using the digest](https://docs.renovatebot.com/presets-helpers/#helperspingithubactiondigests) instead of git tags: `helpers:pinGitHubActionDigests`.

To add an exception to this rule for slsa-github-generator add the following package rule to your `renovate.json` config.

```json
"packageRules": [
    {
      "matchManagers": ["github-actions"],
      "matchPackageNames": ["slsa-framework/slsa-github-generator"],
      "pinDigests": false
    }
  ]
```

This will enable you to receive upgrades for the generator and keep the tagged version.
