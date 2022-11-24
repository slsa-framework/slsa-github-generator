# Security Policy

## Supported Versions

The following versions are currently supported and receive security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.2.x   | :white_check_mark: |
| 1.1.x   | :x:                |
| 0.0.x   | :x:                |

## Reporting a Vulnerability

We're extremely grateful for security researchers and users that report
vulnerabilities to the Kubernetes Open Source Community. All reports are
thoroughly investigated by the project [security team](#security-team).

Vulnerabilities are reported privately via GitHub's [Security Avisories](https://docs.github.com/en/code-security/security-advisories) feature. Please use the following link to submit your vulnerability:
[Report a vulnerability](https://github.com/slsa-framework/slsa-github-generator/security/advisories/new)

Please see
[Privately reporting a security vulnerability](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability#privately-reporting-a-security-vulnerability)
for more information on how to submit a vulnerability using GitHub's interface.

### When Should I Report a Vulnerability?

- You think you discovered a potential security vulnerability in Kubernetes
- You are unsure how a vulnerability affects slsa-github-generator
- You think you discovered a vulnerability in another project that slsa-github-generator depends on
  - For projects with their own vulnerability reporting and disclosure process, please report it directly there

### When Should I NOT Report a Vulnerability?

- You need help tuning GitHub Actions for security
- You need help applying security related updates
- Your issue is not security related

### Vulnerability Response

Each report is acknowledged and analyzed by the [Security Team](#security-team)
within 3 working days. This will set off the Security Release Process.

Any vulnerability information shared with Security Response Committee stays
within slsa-github-generator project and will not be disseminated to other
projects unless it is necessary to get the issue fixed.

As the security issue moves from triage, to identified fix, to release planning
we will keep the reporter updated.

## Disclosure Process

### Private Disclosure

We ask that all suspected vulnerabilities be privately and responsibly
disclosed via the [private disclosure process](#reporting-a-vulnerability)
outlined above.

Fixes may be developed and tested by the [Security Team](#security-team) in
private repositories if deemed necessary.

### Public Disclosure

If you know of a publicly disclosed security vulnerability please IMMEDIATELY
[report the vulnerability](#reporting-a-vulnerability) to inform the [Security
Team](#security-team) about the vulnerability so they may start the patch,
release, and communication process.

If possible the Security Team will ask the person making the public report if
the issue can be handled via a private disclosure process. If the reporter
denies the request, the Security Teamgwill move swiftly with the fix and
release process. In extreme cases you can ask GitHub to delete the issue but
this generally isn't necessary and is unlikely to make a public disclosure less
damaging.

### Fix Disclosure Process

Disclosure of vulnerabilities and their fixes will be made public when the fix
is released.

## Security Team

The Security Team is responsible for the overall security of the
project and for reviewing reported vulnerabilities. Each member is familiar
with designing secure software, security issues related to CI/CD, GitHub
actions and build provenance.

You can view current team membership here:
[slsa-framework/slsa-github-generator-security](https://github.com/orgs/slsa-framework/teams/slsa-github-generator-security)

## Severity

The [Security Team](#security-team) evaluates vulnerability severity on a
case-by-case basis, guided by [CVSS 3.1](https://www.first.org/cvss/v3.1/specification-document).
