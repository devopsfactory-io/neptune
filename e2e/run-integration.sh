#!/usr/bin/env bash
set -euo pipefail

# Integration test: run Neptune plan/apply on the current checkout (real PR context)
# with real GitHub (requirements check, PR comments). MinIO is used for locks.
# Caller must set: GITHUB_REPOSITORY, GITHUB_PULL_REQUEST_NUMBER, GITHUB_PULL_REQUEST_BRANCH,
# GITHUB_RUN_ID, GITHUB_TOKEN; GITHUB_PULL_REQUEST_COMMENT_ID may be empty.
# Caller must ensure the base ref (e.g. origin/main) is available for Terramate change detection.

function compose_teardown() {
  cd "$E2E_DIR"
  echo "Tearing down MinIO..."
  docker compose down
  echo "...MinIO torn down"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
E2E_DIR="$SCRIPT_DIR"

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
export NEPTUNE_CONFIG_PATH="e2e/.neptune.yaml"
export GITHUB_PULL_REQUEST_COMMENT_ID="${GITHUB_PULL_REQUEST_COMMENT_ID:-}"
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-minioadmin}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-minioadmin}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL_S3:-http://localhost:9000}"

echo "Running neptune command plan..."
neptune command plan
echo "...Finished neptune command plan"
echo "--------------------------------"

echo "Running neptune command apply..."
neptune command apply
echo "...Finished neptune command apply"
echo "--------------------------------"
echo "Integration test completed successfully."
