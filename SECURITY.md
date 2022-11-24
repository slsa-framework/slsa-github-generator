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
vulnerabilities to us. All reports are thoroughly investigated by the project
[security team](#security-team).

Vulnerabilities are reported privately via GitHub's
[Security Avisories](https://docs.github.com/en/code-security/security-advisories)
feature. Please use the following link to submit your vulnerability:
[Report a vulnerability](https://github.com/slsa-framework/slsa-github-generator/security/advisories/new)

Please see
[Privately reporting a security vulnerability](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability#privately-reporting-a-security-vulnerability)
for more information on how to submit a vulnerability using GitHub's interface.

### When Should I Report a Vulnerability?

- You think you discovered a potential security vulnerability in slsa-github-generator
- You are unsure how a vulnerability affects slsa-github-generator
- You think you discovered a vulnerability in another project that slsa-github-generator depends on
  - For projects with their own vulnerability reporting and disclosure process, please report it directly there

### When Should I NOT Report a Vulnerability?

- You need help tuning GitHub Actions for security
- You need help applying security related updates
- Your issue is not security related

### Vulnerability Response

Each report is acknowledged and analyzed by the [Security Team](#security-team)
within 3 working days. This will set off the
[Security Release Process](#security-release--disclosure-process).

Any vulnerability information shared with the Security Team stays within
slsa-github-generator project and will not be disseminated to other projects
unless it is necessary to get the issue fixed.

As the security issue moves from triage, to identified fix, to release planning
we will keep the reporter updated.

## Security Release & Disclosure Process

Security vulnerabilities should be handled quickly and sometimes privately. The
primary goal of this process is to reduce the total time users are vulnerable
to publicly known exploits.

### Private Disclosure

We ask that all suspected vulnerabilities be privately and responsibly
disclosed via the [private disclosure process](#reporting-a-vulnerability)
outlined above.

Fixes may be developed and tested by the [Security Team](#security-team) in
private repositories if deemed necessary.

### Public Disclosure

Vulnerabilities are disclosed as GitHub
[Security Advisories](https://github.com/slsa-framework/slsa-github-generator/security/advisories).

A public disclosure date is negotiated by the [Security Team](#security-team)
and the bug submitter. We prefer to fully disclose the bug as soon as possible
once a user mitigation is available. It is reasonable to delay disclosure when
the bug or the fix is not yet fully understood, the solution is not
well-tested, or for vendor coordination. The timeframe for disclosure is from
immediate (especially if it's already publicly known) to a few weeks. For a
vulnerability with a straightforward mitigation, we expect report date to
disclosure date to be on the order of 7 days.

If you know of a publicly disclosed security vulnerability please IMMEDIATELY
[report the vulnerability](#reporting-a-vulnerability) to inform the
[Security Team](#security-team) about the vulnerability so they may start the
patch, release, and communication process.

If possible the Security Team will ask the person making the public report if
the issue can be handled via a private disclosure process. If the reporter
denies the request, the Security Team will move swiftly with the fix and
release process. In extreme cases you can ask GitHub to delete the issue but
this generally isn't necessary and is unlikely to make a public disclosure less
damaging.

## Security Releases

Once a fix is available it will be released and announced via the project on
GitHub. Security releases will announced and clearly marked as a security
release. As much as possible this announcement should be actionable, and
include any mitigating steps users can take prior to upgrading to a fixed
version.

Fixes will be applied in patch releases to all [supported
versions](#supported-versions).

## Security Team

The Security Team is responsible for the overall security of the
project and for reviewing reported vulnerabilities. Each member is familiar
with designing secure software, security issues related to CI/CD, GitHub
actions and build provenance.

You can view current team membership here:
[slsa-framework/slsa-github-generator-security](https://github.com/orgs/slsa-framework/teams/slsa-github-generator-security/members)

Security Team membership is currently considered on a case-by-case basis.

## Severity

The [Security Team](#security-team) evaluates vulnerability severity on a
case-by-case basis, guided by [CVSS 3.1](https://www.first.org/cvss/v3.1/specification-document).
