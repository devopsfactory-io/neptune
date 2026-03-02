# AGENTS.md

Guidance for AI coding agents working on the Neptune project.

---

## Project Overview

**Neptune** is a Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis). It runs plan/apply (Terraform or OpenTofu) on pull requests using the [Terramate](https://github.com/terramate-io/terramate) Go SDK for change detection and run order. When a step has `once` false or unset (default), Neptune runs the step’s command in each changed stack (Terramate SDK or local stacks; no Terramate CLI needed when using Terramate); object storage (GCS or S3) is used for stack locking, and GitHub for PR requirements and comments.

**Main capabilities**: Load config from `.neptune.yaml` and env; in CI (non-E2E), config is loaded from the repository’s default branch via git (fallback to PR branch) so PR authors cannot change workflow steps; check PR requirements (approved, mergeable, undiverged, rebased); lock stacks in object storage (GCS, AWS S3, or S3-compatible e.g. MinIO); run workflow steps (per-stack by default, or once in root when `once: true`); for stacks_management: local, **neptune stacks** provides list (--changed) and create; post results as PR comments. Optional `repository.automerge: true` enables PR auto-merge after a successful apply (GitHub GraphQL). Log level is configurable via `log_level` (config) or `NEPTUNE_LOG_LEVEL` (DEBUG, INFO, ERROR).

**Language**: Go (see `go.mod`).

---

## Repository Structure

- **`main.go`** – Entry point; version/commit/date via ldflags.
- **`cmd/`** – CLI (Cobra): `root.go`, `version.go`, `command.go`, `unlock.go`, `stacks.go` (stacks list, create).
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
- **`.github/workflows/release.yml`** – On push of tags `v*.*.*` (and workflow_dispatch); runs GoReleaser to create GitHub Release and binaries.
- **Renovate** – Dependency-update PRs (Go modules and GitHub Actions) are opened by [Renovate](https://docs.renovatebot.com/) from [.github/renovate.json5](.github/renovate.json5). Do not remove or override this config without reason.

Semantic versioning: use tags like `v0.2.0`. GoReleaser injects version/commit/date into the binary via ldflags.

---

## Documentation and AI Context (Mandatory)

After any change that affects behavior, APIs, config, or CI:

1. **Consider human docs**: **README.md** is the high-level entry point; detailed user docs (configuration, object storage, installation, usage, development) live in **docs/** (see [docs/README.md](docs/README.md)); **examples/** is part of the documentation (copy-pasteable usage examples). Update README, **docs/*.md**, **examples/** (add or update examples when behavior or config changes), and **`.neptune.example.yaml`** (or comments) if install, usage, or config schema changed.
2. **Consider AI docs**: Update **AGENTS.md** if project structure, setup, or conventions changed. Update **`.cursor/rules/*.mdc`** if coding or workflow rules changed. Update **`.cursor/skills/*/SKILL.md`** if a documented workflow or checklist changed.

If you add a feature, change a command, or modify workflows: check README, docs/, examples/, and AGENTS.md; if rules or skills are affected, update the corresponding file. When in doubt, update. See the project skill **maintain-documentation** (`.cursor/skills/maintain-documentation/`) for a detailed checklist.

Do not edit plan files (e.g. `neptune_go_rewrite*.plan.md` or `ai_agent_config*.plan.md`) unless the user explicitly asks.

---

## PR Guidance

Before submitting:

1. **Commits must be signed off (DCO).** Use `git commit -s` when creating commits. If you already committed without sign-off, run `git commit --amend -s --no-edit` then force-push. See [CONTRIBUTING.md](CONTRIBUTING.md) and the `.cursor/rules/commits-dco.mdc` rule.
2. Run `make test-all` and `make check-fmt`.
3. Ensure no new linter errors (`make lint` if available).
4. If behavior or setup changed, update README, docs/, examples/, and/or AGENTS.md and rules/skills as above.

PR titles may follow a conventional style (e.g. `feat(cmd): ...`, `fix(lock): ...`, `docs: ...`) but this is not enforced.

---

## References

- **Contributing (human)**: [CONTRIBUTING.md](CONTRIBUTING.md) – main entry for contributors; [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), [SECURITY.md](SECURITY.md); issue and PR templates in [.github/ISSUE_TEMPLATE/](.github/ISSUE_TEMPLATE/) and [.github/pull_request_template.md](.github/pull_request_template.md).
- **Governance**: [GOVERNANCE.md](GOVERNANCE.md), [MAINTAINERS.md](MAINTAINERS.md), [ROADMAP.md](ROADMAP.md).
- **Cursor rules**: `.cursor/rules/` – file-specific and always-applied rules.
- **Cursor skills**: `.cursor/skills/` – workflows for documentation maintenance, releases, testing, and open-pull-request (open a PR from current changes via gh CLI).
- **Getting started**: [docs/getting-started-terramate.md](docs/getting-started-terramate.md) and [docs/getting-started-local-stacks.md](docs/getting-started-local-stacks.md) – onboarding with GitHub Actions and neptbot (Terramate or local stacks).
- **Neptune config**: [docs/configuration.md](docs/configuration.md) and [.neptune.example.yaml](.neptune.example.yaml) for `.neptune.yaml` schema; [docs/object-storage.md](docs/object-storage.md) for backend env vars.
- **Why Neptune / workflow comparison**: [docs/workflow-comparison.md](docs/workflow-comparison.md) – comparison of normal Terraform + GitHub Actions, Neptune, and Atlantis; use when explaining rationale for apply-before-merge.
