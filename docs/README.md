# Neptune documentation

This folder contains detailed documentation for configuration, object storage, installation, usage, and development.

## Getting started

| Topic | Description |
|-------|-------------|
| [Getting started (Terramate)](getting-started-terramate.md) | Set up Neptune with GitHub Actions and neptbot using Terramate stacks |
| [Getting started (Local stacks)](getting-started-local-stacks.md) | Set up Neptune with GitHub Actions and neptbot using local stack discovery or config |

## Reference

| Topic | Description |
|-------|-------------|
| [Workflow comparison](workflow-comparison.md) | Comparison of normal Terraform + GitHub Actions, Neptune, and Atlantis workflows; apply-before-merge and where runs execute |
| [Configuration](configuration.md) | `.neptune.yaml` schema, plan/apply requirements, and Terramate |
| [Object storage](object-storage.md) | GCS, S3, and MinIO setup and environment variables |
| [Installation](installation.md) | Install Neptune binary and triggering via GitHub App |
| [Usage](usage.md) | GitHub Actions/PR flow and CLI reference |
| [GitHub App and Lambda](github-app-and-lambda.md) | Trigger Neptune via GitHub App webhooks and AWS Lambda (`repository_dispatch`) |
| [Development](development.md) | Building, testing, and contributing |
| [Examples](../examples/) | Sample configs and backends (S3, GCS, automerge, Terramate, Terragrunt) in `examples/` |
