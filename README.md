# Neptune

<p align="center">
  <a href="https://github.com/devopsfactory-io/neptune/releases"><img src="https://img.shields.io/github/v/release/devopsfactory-io/neptune?color=%239F50DA&display_name=tag&label=Version" alt="Latest Release" /></a>
  <a href="https://pkg.go.dev/github.com/devopsfactory-io/neptune"><img src="https://pkg.go.dev/badge/github.com/devopsfactory-io/neptune" alt="Go Docs" /></a>
  <a href="https://goreportcard.com/report/github.com/devopsfactory-io/neptune"><img src="https://goreportcard.com/badge/github.com/devopsfactory-io/neptune" alt="Go Report Card" /></a>
  <a href="https://github.com/devopsfactory-io/neptune/actions?query=branch%3Amain"><img src="https://github.com/devopsfactory-io/neptune/actions/workflows/test.yml/badge.svg" alt="CI Status" /></a>
</p>

<p align="center">
  <b>Terraform and OpenTofu Pull Request Automation</b>
</p>

- [Resources](#resources)
- [What is Neptune?](#what-is-neptune)
- [What does it do?](#what-does-it-do)
- [Why should you use it?](#why-should-you-use-it)

## Resources

- **Documentation**: [docs/](docs/README.md) – Configuration, object storage, installation, usage, and development. Log level can be set via `log_level` in config or `NEPTUNE_LOG_LEVEL` (DEBUG, INFO, ERROR).
- **E2E tests**: [e2e/README.md](e2e/README.md) – Run against MinIO with `./e2e/run.sh` or `make e2e`
- **Releases**: [github.com/devopsfactory-io/neptune/releases](https://github.com/devopsfactory-io/neptune/releases)
- **neptbot**: Trigger Neptune from PR open and @-mention comments by [installing the neptbot GitHub App](docs/github-app-and-lambda.md) and adding the workflow (recommended). To self-host, see [lambda/](lambda/) and [lambda/README.md](lambda/README.md).
- **Contributing**: [docs/development.md](docs/development.md) and [AGENTS.md](AGENTS.md) for AI/contributor guidance

## What is Neptune?

A Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis). It uses the [Terramate](https://github.com/terramate-io/terramate) Go SDK for change detection and run order. By default, workflow steps run in each changed stack (no Terramate CLI required for that); object storage (GCS or S3) is used for stack locking, and GitHub for PR requirements and comments.

## What does it do?

Runs Terraform or OpenTofu plan and apply on pull requests, locks stacks in object storage, checks PR requirements (e.g. approved, mergeable, undiverged), posts results as PR comments, and sets GitHub commit statuses for **neptune plan** and **neptune apply** (so you can require **neptune apply** in branch protection to block merge until apply has run).

## Why should you use it?

- Make Terraform/OpenTofu changes visible to your whole team
- Apply approved changes in a consistent way
- Standardize workflows with configurable plan/apply steps
- Type-safe CLI with auto-completion support
