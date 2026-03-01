# Neptune infrastructure examples

Infrastructure examples for [Neptune](https://github.com/devopsfactory-io/neptune) (Terraform/OpenTofu PR automation with Terramate). These files live in the Neptune repo under `examples/`.

They show different ways to use [Neptune](https://github.com/devopsfactory-io/neptune) for Terraform/OpenTofu pull request automation: object storage backends (S3, GCS), PR automerge, Terramate stacks with Terraform, and Terragrunt with Terramate.

| Example | Description |
|---------|-------------|
| [s3-backend](s3-backend/) | Neptune with **S3** (or MinIO) for stack lock storage |
| [gcs-backend](gcs-backend/) | Neptune with **GCS** for stack lock storage |
| [automerge](automerge/) | **PR automerge** after successful apply (`repository.automerge: true`) |
| [terramate-stacks](terramate-stacks/) | **Terramate stacks** with plain Terraform; Neptune runs steps per changed stack |
| [terragrunt-terramate](terragrunt-terramate/) | **Terragrunt** with Terramate stacks; Neptune runs Terragrunt per stack |

Each example is self-contained: Terramate root, stacks, `.neptune.yaml`, and a README with required environment variables and usage.

## Neptune config in CI

Neptune loads `.neptune.yaml` from the **default branch** of your repository in CI (so PR authors cannot change workflow steps). See [Neptune documentation](https://github.com/devopsfactory-io/neptune/tree/main/docs) for [configuration](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) and [object storage](https://github.com/devopsfactory-io/neptune/blob/main/docs/object-storage.md).

## Using these examples

The examples are in this repo under `examples/` (e.g. `examples/s3-backend/`, `examples/automerge/`). Copy or adapt the example you need; set the required env vars and replace placeholder bucket names in `.neptune.yaml`.
