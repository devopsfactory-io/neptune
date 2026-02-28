# Contributing to Neptune

Thank you for your interest in contributing. We welcome contributions and encourage you to open an issue or pull request.

## Getting started

1. Fork the repository and clone your fork.
2. Create a branch from `main` for your changes.
3. Set up your environment and run the project locally. See [docs/development.md](docs/development.md) for build and test commands (e.g. `make build`, `make test-all`, `make check-fmt`). Use the Go version from [go.mod](go.mod).

## How to contribute

### Issues

- **Bug reports** and **feature requests**: Use the [issue templates](.github/ISSUE_TEMPLATE/) when opening a new issue. Choose "Bug report" or "Feature request" as appropriate.
- Search [existing issues](https://github.com/devopsfactory-io/neptune/issues) first to avoid duplicates.
- For security vulnerabilities, please report them responsibly (e.g. via GitHub Security Advisories or as specified in the repository's security policy).

### Pull requests

- Use the [pull request template](.github/pull_request_template.md). Fill in **What is this feature?**, **Why do we need this feature?**, and **Who is this feature for?** so the PR correlates with issues.
- Link related issues in the "Related issues" section (e.g. `Fixes #123` or `Relates to #456`).
- Complete the checklist before requesting review: breaking changes (if any), documentation updated, tests added or updated.
- You can add an **AI Summary** at the end if you have one (e.g. from Cursor or GitHub Copilot) to help reviewers.
- The `neptune` label is applied automatically when you open a PR; you don't need to add it yourself.

## Code and documentation

- **Code style**: Format with `gofmt -s` and run `make check-fmt` before committing. Linting follows [.golangci.yml](.golangci.yml). For more detail, see [AGENTS.md](AGENTS.md).
- **Documentation**: When you change behavior, configuration, or setup, update the relevant docs: [README.md](README.md), [docs/](docs/), and/or [AGENTS.md](AGENTS.md) as appropriate. See the project's documentation guidelines (e.g. in [.cursor/rules](.cursor/rules) or [.cursor/skills/maintain-documentation](.cursor/skills/maintain-documentation/)).

## CI

Pull requests must pass tests and formatting. CI runs:

- `make test-all` and `make check-fmt` (see [.github/workflows/test.yml](.github/workflows/test.yml))
- Optionally `make lint` (see [.github/workflows/lint.yml](.github/workflows/lint.yml))

Run `make test-all` and `make check-fmt` locally before pushing. For more on workflows and release process, see [AGENTS.md](AGENTS.md#ci).
