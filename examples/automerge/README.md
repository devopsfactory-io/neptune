# Automerge example

After a successful **apply**, Neptune can enable GitHub **auto-merge** on the pull request so it merges when all required checks pass. This example sets `repository.automerge: true` in `.neptune.yaml`.

## Requirements

- The GitHub token used by Neptune must have permission to enable auto-merge on pull requests. In GitHub Actions, grant `pull_requests: write` to the job that runs Neptune.
- Auto-merge is only triggered after a successful **apply**, not after plan.

See [Neptune configuration](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) for the `automerge` option and repository fields.

## Setup

Replace `s3://your-bucket` (or use `gs://your-bucket`) and set the same object-storage environment variables as in the [s3-backend](s3-backend/) or [gcs-backend](gcs-backend/) examples.
