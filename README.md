# Neptune

<p align="center">
  <a href="https://github.com/kaio6fellipe/neptune/releases"><img src="https://img.shields.io/github/v/release/kaio6fellipe/neptune?color=%239F50DA&display_name=tag&label=Version" alt="Latest Release" /></a>
  <a href="https://pkg.go.dev/github.com/kaio6fellipe/neptune"><img src="https://pkg.go.dev/badge/github.com/kaio6fellipe/neptune" alt="Go Docs" /></a>
  <a href="https://goreportcard.com/report/github.com/kaio6fellipe/neptune"><img src="https://goreportcard.com/badge/github.com/kaio6fellipe/neptune" alt="Go Report Card" /></a>
  <a href="https://github.com/kaio6fellipe/neptune/actions?query=branch%3Amain"><img src="https://github.com/kaio6fellipe/neptune/actions/workflows/test.yml/badge.svg" alt="CI Status" /></a>
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
- **Releases**: [github.com/kaio6fellipe/neptune/releases](https://github.com/kaio6fellipe/neptune/releases)
- **Contributing**: [docs/development.md](docs/development.md) and [AGENTS.md](AGENTS.md) for AI/contributor guidance

## What is Neptune?

A Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis). It uses [Terramate](https://github.com/terramate-io/terramate) for change detection, object storage (GCS or S3) for stack locking, and GitHub for PR requirements and comments.

## What does it do?

Runs Terraform or OpenTofu plan and apply on pull requests, locks stacks in object storage, checks PR requirements (e.g. approved, mergeable, undiverged), and posts results as PR comments.

## Why should you use it?

- Make Terraform/OpenTofu changes visible to your whole team
- Apply approved changes in a consistent way
- Standardize workflows with configurable plan/apply steps
- Type-safe CLI with auto-completion support
