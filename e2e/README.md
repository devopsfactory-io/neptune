# End-to-end tests

E2E tests run Neptune against MinIO (S3-compatible) using Docker Compose. They verify lock storage, workflow execution (plan/apply), and Terramate change detection without a real GitHub PR.

## Prerequisites

- **Docker** and **Docker Compose** – for MinIO
- **Go** – to build Neptune (see `go.mod`)
- **Terraform** (or OpenTofu) – for the test stacks

Neptune uses the Terramate Go SDK for change detection and run order; the Terramate CLI is not required for e2e or integration.

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
- **run-integration.sh** – Orchestrates MinIO and Neptune plan/apply with real GitHub (used by the integration workflow).

## CI

The [e2e workflow](../.github/workflows/e2e.yml) runs on pushes to `main`/`release-*` and on pull requests when relevant paths change (e2e/, internal/lock, cmd/, config, compose files). It sets up Terraform and runs `./e2e/run.sh`.

## Integration tests

Integration tests run Neptune (plan/apply) on a **real pull request** with a **real GitHub connection**: PR requirement checks and comment posting. Lock storage still uses MinIO in the same job (no real GCS/S3).

- **When they run**: The [integration workflow](../.github/workflows/integration.yml) runs only on **pull requests** when integration-relevant paths change (e2e/, internal/github/, internal/notifications/github/, cmd/, internal/config/, internal/lock/, internal/run/, internal/domain/, compose files, and the workflow file).
- **What they do**: Checkout (with full history so the base ref exists for Terramate), start MinIO and create the bucket, then run `./e2e/run-integration.sh`. The script injects a trivial change in `e2e/stack-a/main.tf` and `e2e/stack-b/main.tf` and commits it locally so Terramate always sees changed stacks (HEAD vs `origin/main`), then runs `neptune command plan` and `neptune command apply` from the `e2e/` directory with `NEPTUNE_CONFIG_PATH=.neptune.yaml`. GitHub env is set from the workflow so Neptune posts result comments on the PR.
- **Full plan/apply**: Because the script commits a change under `e2e/`, Terramate reports changed stacks and plan/apply run on them every time; PR comments show the full plan/apply format with stacks and command output.
