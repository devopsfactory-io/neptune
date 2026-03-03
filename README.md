# Neptune 🌊

<p align="center">
  <a href="https://github.com/devopsfactory-io/neptune/releases"><img src="https://img.shields.io/github/v/release/devopsfactory-io/neptune?color=%239F50DA&display_name=tag&label=Version" alt="Latest Release" /></a>
  <a href="https://pkg.go.dev/github.com/devopsfactory-io/neptune"><img src="https://pkg.go.dev/badge/github.com/devopsfactory-io/neptune" alt="Go Docs" /></a>
  <a href="https://github.com/devopsfactory-io/neptune/actions?query=branch%3Amain"><img src="https://github.com/devopsfactory-io/neptune/actions/workflows/test.yml/badge.svg" alt="CI Status" /></a>
</p>

<p align="center">
  <img src="./img/neptune-logo.png" alt="Neptune Logo" width="175"/><br><br>
  <b>Terraform and OpenTofu Pull Request Automation with Github Actions</b>
</p>

- [Neptune 🌊](#neptune-)
  - [Resources](#resources)
  - [What is Neptune?](#what-is-neptune)
  - [What does it do?](#what-does-it-do)
  - [Why should you use it?](#why-should-you-use-it)

## Resources

- **Code of Conduct**: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) – we adopt the CNCF Community Code of Conduct.
- **Documentation**: [docs/](docs/README.md) – Configuration, object storage, installation, usage, and development. Log level can be set via `log_level` in config or `NEPTUNE_LOG_LEVEL` (DEBUG, INFO, ERROR).
- **E2E tests**: [e2e/README.md](e2e/README.md) – Run against MinIO with `./e2e/scripts/run-terramate.sh` or `make e2e`
- **Releases**: [github.com/devopsfactory-io/neptune/releases](https://github.com/devopsfactory-io/neptune/releases)
- **Infra examples**: [examples/](examples/) – S3/GCS backend, automerge, Terramate stacks, Terragrunt.
- **neptbot**: Trigger Neptune from PR open and @-mention comments by [installing the neptbot GitHub App](https://github.com/apps/neptbot) and [adding the workflow](docs/github-app-and-lambda.md) (recommended). To self-host, see [lambda/](lambda/) and [lambda/README.md](lambda/README.md).
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md) – how to contribute; [docs/development.md](docs/development.md) and [AGENTS.md](AGENTS.md) for setup and AI/contributor guidance
- **Dependencies**: [Renovate](https://docs.renovatebot.com/) opens PRs for Go modules and GitHub Actions updates (see [.github/renovate.json5](.github/renovate.json5)).

## What is Neptune?

A Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis), but runs entirely in GitHub Actions. It supports two modes for stack management: **Terramate** (using the [Terramate](https://github.com/terramate-io/terramate) Go SDK for change detection and run order) or **local** (config or `stack.hcl` discovery with git-based change detection). Object storage (GCS or S3) is used for stack locking (we make sure that an stack can not be changed by multiple PRs at the same time), and GitHub for PR requirements and comments.

## What does it do?

Runs Terraform or OpenTofu plan and apply on pull requests safely with github actions. Locks stacks in object storage, checks PR requirements (e.g. approved, mergeable, undiverged), posts results as PR comments, and sets GitHub commit statuses for **neptune plan** and **neptune apply** (so you can require **neptune apply** in branch protection to block merge until apply has run).

## Why should you use it?

With the typical Terraform + GitHub Actions flow, apply often runs *after* merge. Code on `main` can end up broken, and you fix it with follow-up PRs. **Apply-before-merge** (plan on PR → approve → apply on the PR → merge only when apply succeeds) keeps `main` fully executable. Neptune and [Atlantis](https://github.com/runatlantis/atlantis) both support this; Neptune runs **entirely in GitHub Actions**—no separate servers or self-hosted runners.

- Make Terraform/OpenTofu changes visible to your whole team
- Apply approved changes in a consistent way
- Standardize workflows with configurable plan/apply steps

For a detailed comparison of the normal Terraform workflow, Neptune, and Atlantis, see [Workflow comparison](docs/workflow-comparison.md).
