#!/usr/bin/env bash
# E2E for stacks_management: local with root-level local_stacks (source: config). Same flow as run-local-stacks-files.sh but uses .neptune.localdeclaredstacks.yaml.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$E2E_DIR/.." && pwd)"

cd "$REPO_ROOT"
go build -o neptune .
export PATH="$REPO_ROOT:$PATH"

cd "$E2E_DIR"
docker compose up -d minio
docker compose run --rm minio-init
cd "$REPO_ROOT"

E2E_TMP=$(mktemp -d)
trap 'cd "$E2E_DIR" && docker compose down; rm -rf "$E2E_TMP"' EXIT
cp -r "$E2E_DIR"/stack-a "$E2E_TMP/"
cp -r "$E2E_DIR"/stack-b "$E2E_TMP/"
cp -r "$E2E_DIR"/stack-c "$E2E_TMP/"
cp "$E2E_DIR"/.neptune.localdeclaredstacks.yaml "$E2E_TMP/.neptune.yaml"

cd "$E2E_TMP"
git init -b main
git config user.email "e2e@neptune.test"
git config user.name "E2E Test"
git add .
git commit -m "main: all stacks"
git update-ref refs/remotes/origin/main HEAD

git checkout -b pr-1
echo "# e2e change" >> stack-a/main.tf
echo "# e2e change" >> stack-b/main.tf
echo "# e2e change" >> stack-c/main.tf
git add stack-a/main.tf stack-b/main.tf stack-c/main.tf
git commit -m "pr-1: change all stacks"

export NEPTUNE_E2E=1
export NEPTUNE_CONFIG_PATH=".neptune.yaml"
export GITHUB_REPOSITORY="e2e/neptune-test"
export GITHUB_PULL_REQUEST_NUMBER="1"
export GITHUB_PULL_REQUEST_BRANCH="pr-1"
export GITHUB_RUN_ID="1"
export GITHUB_TOKEN=""
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-minioadmin}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-minioadmin}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL_S3:-http://localhost:9000}"

echo "Running neptune command plan (stacks_management: local, local_stacks config)..."
neptune command plan
echo "Running neptune command apply (stacks_management: local, local_stacks config)..."
neptune command apply
echo "E2E (local declared stacks) completed successfully."
