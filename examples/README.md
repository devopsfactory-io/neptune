# neptune-infra-examples

Infrastructure examples for [Neptune](https://github.com/devopsfactory-io/neptune) (Terraform/OpenTofu PR automation with Terramate).

This repository is intended to be used as the **examples** submodule of the Neptune repo; from Neptune root run `git submodule update --init --recursive` to fetch it.

These examples show different ways to use [Neptune](https://github.com/devopsfactory-io/neptune) for Terraform/OpenTofu pull request automation: object storage backends (S3, GCS), PR automerge, Terramate stacks with Terraform, and Terragrunt with Terramate.

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

- **As a submodule from Neptune**: From the Neptune repo root, run `git submodule update --init --recursive` to fetch this repo into `examples/`. The examples listed above then live under `examples/examples/`.
- **Standalone**: Clone [neptune-infra-examples](https://github.com/devopsfactory-io/neptune-infra-examples) and copy or adapt the example you need; set the required env vars and replace placeholder bucket names in `.neptune.yaml`.
