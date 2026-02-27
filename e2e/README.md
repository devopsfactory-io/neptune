# End-to-end tests

E2E tests run Neptune against MinIO (S3-compatible) using Docker Compose. They verify lock storage, workflow execution (plan/apply), and Terramate change detection without a real GitHub PR.

## Prerequisites

- **Docker** and **Docker Compose** – for MinIO
- **Go** – to build Neptune (see `go.mod`)
- **Terramate** – [installation](https://terramate.io/docs/cli/installation) (used by the e2e workflow steps; Neptune itself uses the Terramate Go library for listing changed stacks)
- **Terraform** (or OpenTofu) – for the test stacks

## Running e2e

From the repository root:

```bash
./e2e/run.sh
```

The script will:

1. Build the `neptune` binary
2. Start MinIO with Docker Compose and create the `neptune-e2e` bucket
3. Prepare an isolated git repo with a `main` branch and a `pr-1` branch that has changes in `stack-a` (so Neptune’s Terramate SDK reports that stack as changed)
4. Run `neptune command plan` and `neptune command apply` with `NEPTUNE_E2E=1`
5. Tear down MinIO

## E2E mode (`NEPTUNE_E2E=1`)

When `NEPTUNE_E2E=1` is set:

- GitHub environment variables may be empty; defaults are used for repo and PR number so lock file metadata works.
- Neptune skips GitHub API calls (no requirement checks, no comment posting).
- A stub treats the current “PR” as open so locks are not removed as stale.

This allows e2e to run in CI and locally without a GitHub token or real pull request.

## Structure

- **stack-a**, **stack-b**, **stack-c** – Terramate stacks with Terraform `null_resource` and `local_file` (no cloud providers).
- **.neptune.yaml** – E2E config pointing at `s3://neptune-e2e` and minimal plan/apply requirements.
- **docker-compose.yaml** – MinIO plus an init container that creates the bucket.
- **run.sh** – Orchestrates MinIO, git fixture, and Neptune plan/apply.

## CI

The [e2e workflow](../.github/workflows/e2e.yml) runs on pushes to `main`/`release-*` and on pull requests when relevant paths change (e2e/, internal/lock, cmd/, config, compose files). It installs Terramate and Terraform, then runs `./e2e/run.sh`.
