# Contributing to Neptune

Thank you for your interest in contributing. We welcome contributions and encourage you to open an issue or pull request.

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting started

1. Fork the repository and clone your fork (infra examples are in the `examples/` directory).
2. Create a branch from `main` for your changes. Optionally follow the [branch naming conventions](#branch-naming-and-pr-labels) below so the labeler can auto-apply PR labels (and release-note categories).
3. Set up your environment and run the project locally. See [docs/development.md](docs/development.md) for build and test commands (e.g. `make build`, `make test-all`, `make check-fmt`). Use the Go version from [go.mod](go.mod).

## How to contribute

### Issues

- **Bug reports** and **feature requests**: Use the [issue templates](.github/ISSUE_TEMPLATE/) when opening a new issue. Choose "Bug report" or "Feature request" as appropriate.
- Search [existing issues](https://github.com/devopsfactory-io/neptune/issues) first to avoid duplicates.
- For **security vulnerabilities**, do not open a public issue. Please report them as described in our [Security Policy](SECURITY.md) (e.g. via [GitHub Security Advisories](https://github.com/devopsfactory-io/neptune/security/advisories/new)).

### Pull requests

- **Sign-off (DCO)**: Your commits must comply with the [Developer Certificate of Origin (DCO)](https://developercertificate.org/). The [DCO bot](https://github.com/apps/dco) is enabled on this repository and will check that every commit has a `Signed-off-by` line matching the commit author.
  - **Git config**: Set your name and email so the sign-off is valid: `git config user.name "Your Name"` and `git config user.email "you@example.com"` (use `--global` to set for all repos).
  - **Adding sign-off**: Use `git commit -s` to add the line automatically. If you forgot, run `git commit --amend -s --no-edit` on the last commit.
  - The DCO bot will comment on the PR with the status; fix any unsigned commits (e.g. amend and force-push) before merge.
- Use the [pull request template](.github/pull_request_template.md). Fill in **What is this feature?**, **Why do we need this feature?**, and **Who is this feature for?** so the PR correlates with issues.
- Link related issues in the "Related issues" section (e.g. `Fixes #123` or `Relates to #456`).
- Complete the checklist before requesting review: breaking changes (if any), documentation updated, tests added or updated.
- You can add an **AI Summary** at the end if you have one (e.g. from Cursor or GitHub Copilot) to help reviewers.
- The `neptune` label is applied automatically when you open a PR; you don't need to add it yourself.

### Branch naming and PR labels

The [labeler](.github/workflows/labeler.yml) uses the **head branch name** (and changed files) to auto-apply labels. Those labels drive release-note categories. To get the right label automatically, you can name your branch as follows (see [.github/labeler.yml](.github/labeler.yml) for the exact rules):

| Branch name pattern | Label applied |
|---------------------|---------------|
| Starts with `feat` (e.g. `feat/add-thing`) | `feature` |
| Starts with `enhance` (e.g. `enhance/improve-x`) | `enhancement` |
| Starts with `fix` and not `fix*dep*` (e.g. `fix/bug-123`) | `bug` |
| Contains `!` (e.g. `feat!/breaking`) | `breaking-change` |
| Starts with `ci` (e.g. `ci/update-workflow`) | `github-actions` |
| Contains `(deps)` (e.g. `(deps)/go-mod`) | `dependencies` |

Labels are also applied from **changed files** (e.g. `dependencies` for `go.mod`/`go.sum`, `documentation` for `docs/**` or `*.md`). Branch naming is optional but helps ensure your PR is categorized correctly in release notes.

## Code and documentation

- **Code style**: Format with `gofmt -s` and run `make check-fmt` before committing. Linting follows [.golangci.yml](.golangci.yml). For more detail, see [AGENTS.md](AGENTS.md).
- **Documentation**: When you change behavior, configuration, or setup, update the relevant docs: [README.md](README.md), [docs/](docs/), and/or [AGENTS.md](AGENTS.md) as appropriate. See the project's documentation guidelines (e.g. in [.cursor/rules](.cursor/rules) or [.cursor/skills/maintain-documentation](.cursor/skills/maintain-documentation/)).
- **Infra examples**: Live in the `examples/` directory.

## CI

Pull requests must pass tests and formatting. CI runs:

- `make test-all` and `make check-fmt` (see [.github/workflows/test.yml](.github/workflows/test.yml))
- Optionally `make lint` (see [.github/workflows/lint.yml](.github/workflows/lint.yml))

Run `make test-all` and `make check-fmt` locally before pushing. For more on workflows and release process, see [AGENTS.md](AGENTS.md#ci).
