# Security Policy

## Supported Versions

We provide security updates for the current major release and the previous major release. See the [Releases](https://github.com/devopsfactory-io/neptune/releases) page for version history.

| Version | Supported          |
| ------- | ------------------ |
| Latest release (see [Releases](https://github.com/devopsfactory-io/neptune/releases)) | :white_check_mark: |
| Older releases | :x: |

## Reporting a Vulnerability

We take security seriously. If you believe you have found a security vulnerability, please report it responsibly.

**Preferred method:** Use [GitHub Security Advisories](https://github.com/devopsfactory-io/neptune/security/advisories/new) (private disclosure). This keeps the report confidential until a fix is ready.

**Alternative:** Email the maintainers listed in [MAINTAINERS.md](MAINTAINERS.md) (use the contact method they provide, if any). Do not open a public issue for security vulnerabilities.

### What to include

- Description of the problem
- Steps to reproduce (as precise as possible)
- Affected version(s)
- Possible mitigations, if known

### What to expect

- We will acknowledge receipt within a few business days.
- We may follow up to clarify or confirm the issue.
- We will work on a fix and coordinate disclosure (e.g. release + advisory) when appropriate.

We do not have a formal embargo policy; we handle disclosure in a reasonable, coordinated way.

## Security Considerations

- Neptune runs Terraform/OpenTofu and interacts with object storage (GCS, S3) and GitHub. Ensure credentials (e.g. `GITHUB_TOKEN`, cloud credentials) are scoped and stored securely.
- In CI, config is loaded from the repository default branch to limit impact of malicious PR changes to workflow steps.
- For deployment and hardening guidance, see [docs/installation.md](docs/installation.md) and [docs/configuration.md](docs/configuration.md).
