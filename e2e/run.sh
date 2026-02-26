#!/usr/bin/env bash
set -euo pipefail

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

# Prepare isolated git repo with main + pr-1 (changed stacks) so terramate --changed works
E2E_TMP=$(mktemp -d)
trap 'rm -rf "$E2E_TMP"' EXIT
cp -r "$E2E_DIR"/stack-a "$E2E_TMP/"
cp -r "$E2E_DIR"/stack-b "$E2E_TMP/"
cp -r "$E2E_DIR"/stack-c "$E2E_TMP/"
cp "$E2E_DIR"/.neptune.yaml "$E2E_TMP/"
cp "$E2E_DIR"/terramate.tm.hcl "$E2E_TMP/" 2>/dev/null || true

cd "$E2E_TMP"
git init -b main
git config user.email "e2e@neptune.test"
git config user.name "E2E Test"
git add .
git commit -m "main: all stacks"

git checkout -b pr-1
# Change stack-a so terramate list --changed returns at least stack-a
echo "# e2e change" >> stack-a/main.tf
git add stack-a/main.tf
git commit -m "pr-1: change stack-a"

# Run Neptune plan and apply against MinIO (e2e mode: no real GitHub)
export NEPTUNE_E2E=1
export NEPTUNE_CONFIG_PATH=".neptune.yaml"
export GITHUB_REPOSITORY="e2e/neptune-test"
export GITHUB_PULL_REQUEST_NUMBER="1"
export GITHUB_PULL_REQUEST_BRANCH="pr-1"
export GITHUB_PULL_REQUEST_COMMENT_ID=""
export GITHUB_RUN_ID="1"
export GITHUB_TOKEN=""
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-minioadmin}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-minioadmin}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL_S3:-http://localhost:9000}"

echo "Running neptune command plan..."
neptune command plan
echo "Running neptune command apply..."
neptune command apply

echo "E2E completed successfully."

# Tear down MinIO so CI and repeated runs start clean
cd "$E2E_DIR"
docker compose down
