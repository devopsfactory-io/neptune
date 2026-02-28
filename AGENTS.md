# AGENTS.md

Guidance for AI coding agents working on the Neptune project.

---

## Project Overview

**Neptune** is a Terraform and OpenTofu pull request automation tool inspired by [Atlantis](https://github.com/runatlantis/atlantis). It runs plan/apply (Terraform or OpenTofu) on pull requests using the [Terramate](https://github.com/terramate-io/terramate) Go SDK for change detection and run order. When a step has `terramate: true` (default), Neptune runs the step’s command in each changed stack via the SDK (no Terramate CLI needed for that step); object storage (GCS or S3) is used for stack locking, and GitHub for PR requirements and comments.

**Main capabilities**: Load config from `.neptune.yaml` and env; in CI (non-E2E), config is loaded from the repository’s default branch via git (fallback to PR branch) so PR authors cannot change workflow steps; check PR requirements (approved, mergeable, undiverged, rebased); lock stacks in object storage (GCS, AWS S3, or S3-compatible e.g. MinIO); run workflow steps (per-stack by default, or once when `terramate: false`); post results as PR comments. Log level is configurable via `log_level` (config) or `NEPTUNE_LOG_LEVEL` (DEBUG, INFO, ERROR).

**Language**: Go (see `go.mod`). Legacy Python code exists under `neptune/` and `tests/`; primary codebase is Go.

---

## Repository Structure

- **`main.go`** – Entry point; version/commit/date via ldflags.
- **`cmd/`** – CLI (Cobra): `root.go`, `version.go`, `command.go`, `unlock.go`.
- **`internal/config`** – Load env + YAML, validate `.neptune.yaml` (including optional `log_level`).
- **`internal/domain`** – Config, lock, run, and GitHub domain structs.
- **`internal/log`** – Structured logging (DEBUG, INFO, ERROR) via `log/slog`; level from `NEPTUNE_LOG_LEVEL` or config `log_level`.
- **`internal/lock`** – Changed stacks via Terramate SDK (list + run order), object-storage lock files (GCS, S3), lock interface.
- **`internal/run`** – Execute workflow phase steps (shell).
- **`internal/github`** – GitHub API client, PR requirements (approved, mergeable, undiverged), commit statuses (GetHeadSHA, CreateCommitStatus) for **neptune plan** / **neptune apply**.
- **`internal/git`** – Rebased check; DefaultBranch (git CLI), ShowFileFromRef, FetchBranch for loading config from default branch.
- **`internal/notifications/github`** – Format and post PR comments.
- **`e2e/`** – End-to-end tests: three Terramate stacks (null_resource/local_file), MinIO via Docker Compose, and `run.sh` that runs Neptune plan/apply with `NEPTUNE_E2E=1` (skips GitHub; see [e2e/README.md](e2e/README.md)).
- **`lambda/`** – AWS Lambda handler for Neptune GitHub App webhooks (verify signature, parse `pull_request`/`issue_comment`, trigger `repository_dispatch`). See [lambda/README.md](lambda/README.md).
- **`lambda/cloudformation/`** – CloudFormation template to deploy the Lambda (Function URL, IAM, Secrets Manager). See [lambda/README.md](lambda/README.md#deploy-with-cloudformation).
- **`Makefile`**, **`.golangci.yml`**, **`.goreleaser.yml`**, **`.github/workflows/`** – Build, test, lint, release.

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
- **E2E**: Run `make e2e` or `./e2e/run.sh` (requires Docker, Terraform). Uses MinIO and `NEPTUNE_E2E=1` to skip GitHub. E2E config uses steps with default `terramate: true` (Neptune runs commands per stack via SDK; Terramate CLI not required for steps).

---

## CI

- **`.github/workflows/test.yml`** – On push to `main`/`release-*` and on PRs; path filter for Go files; runs `make test-all` and `make check-fmt`.
- **`.github/workflows/e2e.yml`** – On push/PR when e2e-related paths change; runs `./e2e/run.sh` with MinIO (Docker Compose). See [e2e/README.md](e2e/README.md).
- **`.github/workflows/integration.yml`** – On PRs when integration-relevant paths change; runs Neptune plan/apply on the same PR with real GitHub (requirements check, PR comments, commit statuses) and MinIO for locks. Needs `statuses: write` for commit status API. See [e2e/README.md](e2e/README.md#integration-tests).
- **`.github/workflows/lint.yml`** – On PRs; path filter for Go; runs golangci-lint.
- **`.github/workflows/release.yml`** – On push of tags `v*.*.*` (and workflow_dispatch); runs GoReleaser to create GitHub Release and binaries.

Semantic versioning: use tags like `v0.2.0`. GoReleaser injects version/commit/date into the binary via ldflags.

---

## Documentation and AI Context (Mandatory)

After any change that affects behavior, APIs, config, or CI:

1. **Consider human docs**: **README.md** is the high-level entry point; detailed user docs (configuration, object storage, installation, usage, development) live in **docs/** (see [docs/README.md](docs/README.md)). Update README, **docs/*.md**, and **`.neptune.example.yaml`** (or comments) if install, usage, or config schema changed.
2. **Consider AI docs**: Update **AGENTS.md** if project structure, setup, or conventions changed. Update **`.cursor/rules/*.mdc`** if coding or workflow rules changed. Update **`.cursor/skills/*/SKILL.md`** if a documented workflow or checklist changed.

If you add a feature, change a command, or modify workflows: check README, docs/, and AGENTS.md; if rules or skills are affected, update the corresponding file. When in doubt, update. See the project skill **maintain-documentation** (`.cursor/skills/maintain-documentation/`) for a detailed checklist.

Do not edit plan files (e.g. `neptune_go_rewrite*.plan.md` or `ai_agent_config*.plan.md`) unless the user explicitly asks.

---

## PR Guidance

Before submitting:

1. Run `make test-all` and `make check-fmt`.
2. Ensure no new linter errors (`make lint` if available).
3. If behavior or setup changed, update README, docs/, and/or AGENTS.md and rules/skills as above.

PR titles may follow a conventional style (e.g. `feat(cmd): ...`, `fix(lock): ...`, `docs: ...`) but this is not enforced.

---

## References

- **Cursor rules**: `.cursor/rules/` – file-specific and always-applied rules.
- **Cursor skills**: `.cursor/skills/` – workflows for documentation maintenance, releases, testing, and open-pull-request (open a PR from current changes via gh CLI).
- **Neptune config**: [docs/configuration.md](docs/configuration.md) and [.neptune.example.yaml](.neptune.example.yaml) for `.neptune.yaml` schema; [docs/object-storage.md](docs/object-storage.md) for backend env vars.
