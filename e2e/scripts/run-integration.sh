#!/usr/bin/env bash
set -euo pipefail

# Integration test: run Neptune plan/apply on the current checkout (real PR context)
# with real GitHub (requirements check, PR comments). MinIO is used for locks.
# Caller must set: GITHUB_REPOSITORY, GITHUB_PULL_REQUEST_NUMBER, GITHUB_PULL_REQUEST_BRANCH,
# GITHUB_RUN_ID, GITHUB_TOKEN.
# Caller must ensure the base ref (e.g. origin/main) is available for Terramate change detection.

function compose_teardown() {
  cd "$E2E_DIR"
  echo "Tearing down MinIO..."
  docker compose down
  echo "...MinIO torn down"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$E2E_DIR/.." && pwd)"

cd "$REPO_ROOT"

# Build neptune binary
echo "Building neptune..."
go build -o neptune .
export PATH="$REPO_ROOT:$PATH"

# Start MinIO and create bucket
echo "Starting MinIO..."
cd "$E2E_DIR"
docker compose up -d minio
docker compose run --rm minio-init
cd "$REPO_ROOT"

trap compose_teardown EXIT

# Use current checkout; GitHub env is set by caller (e.g. GHA). No NEPTUNE_E2E.
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-minioadmin}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-minioadmin}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL_S3:-http://localhost:9000}"

# Ensure at least one stack is "changed" so Terramate ListChanged (HEAD vs origin/main) finds it.
# Otherwise integration runs that don't touch e2e/ would see no stacks and skip plan/apply.
echo "Injecting trivial change in e2e/stack-a and stack-b for change detection..."
git config user.email "neptune@integration.test"
git config user.name "Neptune Integration"
echo "# integration: ensure changed stack for Terramate" >> "$E2E_DIR/stack-a/main.tf"
echo "# integration: ensure changed stack for Terramate" >> "$E2E_DIR/stack-b/main.tf"
git add "$E2E_DIR/stack-a/main.tf" "$E2E_DIR/stack-b/main.tf"
git commit -m "chore(e2e): trigger integration changed stacks"

# Run Neptune from e2e/ so the runner's cwd matches the Terramate root (stack paths resolve correctly).
cd "$E2E_DIR"
export NEPTUNE_CONFIG_PATH=".neptune.yaml"

echo "Running neptune command plan..."
neptune command plan
echo "...Finished neptune command plan"
echo "--------------------------------"

echo "Running neptune command apply..."
neptune command apply
echo "...Finished neptune command apply"
echo "--------------------------------"
echo "Integration test completed successfully."
