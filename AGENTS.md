# AGENTS.md

Guidance for AI coding agents working on the Neptune project.

---

## Project Overview

**Neptune** is a Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis). It runs plan/apply (Terraform or OpenTofu) on pull requests using the [Terramate](https://github.com/terramate-io/terramate) Go SDK for change detection and run order. When a step has `once` false or unset (default), Neptune runs the step’s command in each changed stack (Terramate SDK or local stacks; no Terramate CLI needed when using Terramate); object storage (GCS or S3) is used for stack locking, and GitHub for PR requirements and comments.

**Main capabilities**: Load config from `.neptune.yaml` and env; in CI (non-E2E), config is loaded from the repository’s default branch via git (fallback to PR branch) so PR authors cannot change workflow steps; check PR requirements (approved, mergeable, undiverged, rebased); lock stacks in object storage (GCS, AWS S3, or S3-compatible e.g. MinIO); run workflow steps (per-stack by default, or once in root when `once: true`); for stacks_management: local, **neptune stacks** provides list (--changed) and create, with **--format** (json, yaml, text, formatted; default formatted); when using discovery, run order can be set via **depends_on** in each **stack.hcl** (path-only; relative paths and directory-of-stacks supported); post results as PR comments. Optional `repository.automerge: true` enables PR auto-merge after a successful apply (GitHub GraphQL). Log level is configurable via `log_level` (config), `NEPTUNE_LOG_LEVEL`, or the global **--log-level** CLI flag (DEBUG, INFO, ERROR).

**Language**: Go (see `go.mod`).

---

## Repository Structure

- **`main.go`** – Entry point; version/commit/date via ldflags.
- **`cmd/`** – CLI (Cobra): `root.go` (global **--log-level** flag), `version.go`, `command.go`, `unlock.go`, `stacks.go` (stacks list, create; **--format** json|yaml|text|formatted, default formatted; **neptune stacks create** supports optional **--depends-on** comma-separated paths). **neptune stacks list** and **neptune stacks create** are for local use and do not require `GITHUB_TOKEN` or other CI env vars (unlike **neptune command** and **neptune unlock**).
- **`internal/config`** – Load env + YAML, validate `.neptune.yaml` (including optional `log_level`, `stacks_management`, root-level `local_stacks`).
- **`internal/domain`** – Config, lock, run, and GitHub domain structs (WorkflowStep uses `once`; RepositoryConfig has `StacksManagement`, `LocalStacks`).
- **`internal/log`** – Structured logging (DEBUG, INFO, ERROR) via `log/slog`; level from `NEPTUNE_LOG_LEVEL` or config `log_level`.
- **`internal/stacks`** – Stacks provider interface (terramate, local); list/changed stacks for locking and runner.
- **`internal/lock`** – Lock interface: gets stack list from stacks provider, object-storage lock files (GCS, S3).
- **`internal/run`** – Execute workflow phase steps (shell).
- **`internal/github`** – GitHub API client, PR requirements (approved, mergeable, undiverged), commit statuses (GetHeadSHA, CreateCommitStatus) for **neptune plan** / **neptune apply**; GraphQL EnablePullRequestAutoMerge when `repository.automerge` is true.
- **`internal/git`** – Rebased check; DefaultBranch (git CLI), ShowFileFromRef, FetchBranch for loading config from default branch.
- **`internal/notifications/github`** – Format and post PR comments.
- **`examples/`** – Infra examples (S3/GCS backend, automerge, Terramate stacks, Terragrunt).
- **`scripts/`** – Maintainer scripts (if any).
- **`e2e/`** – End-to-end tests: three Terramate stacks (null_resource/local_file), MinIO via Docker Compose, and `scripts/run-terramate.sh` that runs Neptune plan/apply with `NEPTUNE_E2E=1` (skips GitHub; see [e2e/README.md](e2e/README.md)).
- **`lambda/`** – AWS Lambda handler for Neptune GitHub App webhooks (verify signature, parse `pull_request`—including `labeled` when the added label is `NEPTUNE_PR_LABEL`—and `issue_comment`, trigger `repository_dispatch`; optional `NEPTUNE_PR_LABEL` gates on PR label). See [lambda/README.md](lambda/README.md).
- **`lambda/cloudformation/`** – CloudFormation template to deploy the Lambda (Function URL, IAM, Secrets Manager). See [lambda/README.md](lambda/README.md#deploy-with-cloudformation).
- **`Makefile`**, **`.golangci.yml`**, **`.goreleaser.yml`**, **`.github/workflows/`** – Build, test, lint, release.
- **`.cursor/agents/`** – Cursor/Task subagents: documentation-maintainer (runs the doc checklist after code/config/CI changes; delegate to it for README, docs/, examples/, AGENTS.md, rules, commands, skills), issue-reviewer, pr-reviewer (discoverable for triage and PR review), issue-writer (opens feature requests and bug reports from `/feature` and `/bug` using [.github/ISSUE_TEMPLATE/](.github/ISSUE_TEMPLATE/); drafts are validated by issue-reviewer before upload).
- **`.cursor/commands/`** – Cursor slash commands: `/feature`, `/bug` (invoke the issue-writer workflow to create issues from the repo’s issue templates; the draft is validated by issue-reviewer before `gh issue create`).
- **Root community docs**: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), [SECURITY.md](SECURITY.md), [GOVERNANCE.md](GOVERNANCE.md), [MAINTAINERS.md](MAINTAINERS.md), [ROADMAP.md](ROADMAP.md), [LICENSE](LICENSE).

---

## Setup Commands

```bash
# Build binary
make build

# Run all tests (main module)
make test-all

# Lambda (separate module under lambda/): build, package, test
make lambda.build
make lambda.zip
make lambda.test

# Check Go formatting (includes lambda/)
make check-fmt

# Lint (optional; requires golangci-lint)
make lint
```

Use Go version from `go.mod`. No other prerequisites for building or testing the Go CLI. CI runs `make test-all`, `make lambda.test`, and `make check-fmt`.

---

## Code Style

- **Format**: Use `gofmt -s`; run `make check-fmt` before committing.
- **Linting**: `.golangci.yml` is authoritative; do not introduce new linter violations.
- **Packages**: Code under `internal/` must not be imported from outside this module.
- **Errors**: Return errors with context (e.g. `fmt.Errorf("...: %w", err)`); avoid naked returns.
- **Exports**: Public functions and types should have doc comments starting with the name.

---

## Testing

- **Run**: `go test ./...` or `make test-all`.
- **Location**: Place `*_test.go` next to the code under test (same package).
- **Coverage**: Existing tests cover `internal/config`, `internal/git`, `internal/github`, `internal/run`, `internal/notifications/github`; add tests for new behavior and keep coverage for touched code.
- **No external services**: Unit tests should not require live GitHub or GCS; mock or stub as needed.
- **E2E**: Run `make e2e` or `./e2e/scripts/run-terramate.sh` (requires Docker, Terraform). Uses MinIO and `NEPTUNE_E2E=1` to skip GitHub. For **stacks_management: local**, run `./e2e/scripts/run-local-stacks-files.sh` or `./e2e/scripts/run-local-declared-stacks.sh`. E2E config uses steps with default `once: false` (Neptune runs commands per stack).
- **Automerge**: E2E and integration tests in this repo do not exercise the automerge feature. A separate repository (e.g. a fork or copy of [examples/](examples/)) can be used to test automerge end-to-end (e.g. open a PR with changes in two stacks, comment `@neptbot apply`, then verify the apply comment and that the PR is set to auto-merge after checks pass).

---

## CI

- **`.github/workflows/test.yml`** – On push to `main`/`release-*` and on PRs; path filter for Go files; runs `make test-all` and `make check-fmt`.
- **`.github/workflows/e2e.yml`** – On push/PR when e2e-related paths change; runs `./e2e/scripts/run-terramate.sh` with MinIO (Docker Compose). See [e2e/README.md](e2e/README.md).
- **`.github/workflows/integration.yml`** – On PRs when integration-relevant paths change; runs Neptune plan/apply on the same PR with real GitHub (requirements check, PR comments, commit statuses) and MinIO for locks. Needs `statuses: write` for commit status API. See [e2e/README.md](e2e/README.md#integration-tests).
- **`.github/workflows/lint.yml`** – On PRs; path filter for Go; runs golangci-lint.
- **`.github/workflows/labeler.yml`** – On pull_request (opened, synchronize, reopened); runs [actions/labeler](https://github.com/actions/labeler) with [.github/labeler.yml](.github/labeler.yml). Path-based: neptune, dependencies, documentation. Head-branch (branch name): `feat*`→feature, `enhance*`→enhancement, `fix*` (not fix*dep*)→bug, branch containing `!`→breaking-change, `ci*`→github-actions, `(deps)`→dependencies. See CONTRIBUTING.md for contributor-facing branch naming.
- **`.github/workflows/label-old-prs.yml`** – workflow_dispatch; applies the labeler to existing PRs (inputs: state e.g. merged/closed/all, limit). Use to backfill labels on old or merged PRs (Actions → "Label old PRs" → Run workflow). Because the labeler does not receive branch context on workflow_dispatch, a separate step applies the same rules as `.github/labeler.yml` to both head branch and PR title (e.g. `feat*`→feature if either branch or title matches), then the labeler runs for path-based labels.
- **`.github/workflows/release.yml`** – On push of tags `v*.*.*` (and workflow_dispatch); runs GoReleaser to create GitHub Release with neptune binaries (archives e.g. `neptune_linux_amd64.tar.gz` and raw binaries e.g. `neptune_linux_amd64`), Lambda zip (`neptune-webhook.zip`) and raw binary `neptune-webhook_linux_amd64` (Lambda binary is `neptune-webhook`; zip contains it), checksums, and release notes. The release body is generated by **GitHub** (GoReleaser uses `changelog.use: github-native`) and categorized using [.github/release.yml](.github/release.yml) and PR labels (Breaking Changes, Features, Bug fixes, Documentation, Dependency updates, Other work). Release footer includes Full Changelog link.
- **Renovate** – Dependency-update PRs (Go modules and GitHub Actions) are opened by [Renovate](https://docs.renovatebot.com/) from [.github/renovate.json5](.github/renovate.json5). To enable Renovate, install the [Renovate GitHub App](https://github.com/apps/renovate) and select the repo. Do not remove or override this config without reason.

Semantic versioning: use tags like `v0.2.0`. GoReleaser injects version/commit/date into the binary via ldflags.

**Changelog and breaking changes**: The release body is generated by GitHub (github-native) and categorized by [.github/release.yml](.github/release.yml) and **PR labels**. For breaking changes to appear under "Breaking Changes", apply the `breaking-change` label to the PR before merge. The commit subject convention (`!:`) is still recommended for semver (e.g. `feat!: remove deprecated flag`) but does not drive release-note sections; labels do.

---

## Documentation and AI Context (Mandatory)

After any change that affects behavior, APIs, config, or CI:

1. **Delegate**: Delegate documentation updates to the **documentation-maintainer** subagent (`.cursor/agents/documentation-maintainer.md`) so it runs the full maintain-documentation checklist (README, docs/, examples/, AGENTS.md, .cursor/rules, .cursor/commands, .cursor/skills).
2. **Do not edit plan files** (e.g. `neptune_go_rewrite*.plan.md` or `ai_agent_config*.plan.md`) unless the user explicitly asks.

When in doubt, update. See `.cursor/rules/docs-and-ai-context.mdc` (always-applied) and the **maintain-documentation** skill (`.cursor/skills/maintain-documentation/`); the subagent holds the detailed checklist.

---

## PR Guidance

Before submitting:

1. **Commits must be signed off (DCO).** Use `git commit -s` when creating commits. Do not add a `Made-with: Cursor` (or similar) trailer to commit messages. If you already committed without sign-off, run `git commit --amend -s --no-edit` then force-push. See [CONTRIBUTING.md](CONTRIBUTING.md) and the `.cursor/rules/commits-dco.mdc` rule.
2. Run `make test-all` and `make check-fmt`.
3. Ensure no new linter errors (`make lint` if available).
4. If behavior or setup changed, delegate to the **documentation-maintainer** subagent.
5. **Branch naming**: Branch names matching [.github/labeler.yml](.github/labeler.yml) (e.g. `feat/...`, `fix/...`, `enhance/...`, `(deps)/...`, `ci/...`, or branch containing `!` for breaking) get PR labels applied automatically, which drive release-note categories. See [CONTRIBUTING.md](CONTRIBUTING.md#branch-naming-and-pr-labels).

PR titles may follow a conventional style (e.g. `feat(cmd): ...`, `fix(lock): ...`, `docs: ...`) but this is not enforced.

---

## References

- **Contributing (human)**: [CONTRIBUTING.md](CONTRIBUTING.md) – main entry for contributors; [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), [SECURITY.md](SECURITY.md); issue and PR templates in [.github/ISSUE_TEMPLATE/](.github/ISSUE_TEMPLATE/) and [.github/pull_request_template.md](.github/pull_request_template.md).
- **Governance**: [GOVERNANCE.md](GOVERNANCE.md), [MAINTAINERS.md](MAINTAINERS.md), [ROADMAP.md](ROADMAP.md).
- **Cursor rules**: `.cursor/rules/` – file-specific and always-applied rules.
- **Cursor commands**: `.cursor/commands/` – slash commands (e.g. `/feature`, `/bug`) that trigger the issue-writer workflow. Drafts are validated by issue-reviewer before creation.
- **Cursor skills**: `.cursor/skills/` – workflows for documentation maintenance, releases, testing, and open-pull-request (open a PR from current changes via gh CLI).
- **Getting started**: [docs/getting-started-terramate.md](docs/getting-started-terramate.md) and [docs/getting-started-local-stacks.md](docs/getting-started-local-stacks.md) – onboarding with GitHub Actions and neptbot (Terramate or local stacks).
- **Neptune config**: [docs/configuration.md](docs/configuration.md) and [.neptune.example.yaml](.neptune.example.yaml) for `.neptune.yaml` schema; [docs/object-storage.md](docs/object-storage.md) for backend env vars.
- **Why Neptune / workflow comparison**: [docs/workflow-comparison.md](docs/workflow-comparison.md) – comparison of normal Terraform + GitHub Actions, Neptune, and Atlantis; use when explaining rationale for apply-before-merge.
