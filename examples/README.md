# Neptune infrastructure examples

Infrastructure examples for [Neptune](https://github.com/devopsfactory-io/neptune) (Terraform/OpenTofu PR automation with Terramate). These files live in the Neptune repo under `examples/`.

They show different ways to use [Neptune](https://github.com/devopsfactory-io/neptune) for Terraform/OpenTofu pull request automation: object storage backends (S3, GCS), PR automerge, stack management (Terramate or local), Terragrunt, and Terraform.

| Example | Description |
|---------|-------------|
| [s3-backend](s3-backend/) | Neptune with **S3** (or MinIO) for stack lock storage |
| [gcs-backend](gcs-backend/) | Neptune with **GCS** for stack lock storage |
| [automerge](automerge/) | **PR automerge** after successful apply (`repository.automerge: true`) |
| [terramate-stacks](terramate-stacks/) | **stacks_management: terramate** with plain Terraform; Neptune runs steps per changed stack |
| [terragrunt-terramate](terragrunt-terramate/) | **Terragrunt** with Terramate stacks; Neptune runs Terragrunt per stack |
| [local-stacks](local-stacks/) | **stacks_management: local**; Neptune discovers stacks via `stack.hcl` (no Terramate); use **neptune stacks list** / **create** |
| [local-stacks-config](local-stacks-config/) | **stacks_management: local** with root-level **local_stacks** in `.neptune.yaml` (source: config, explicit paths and `depends_on` for run order) |

Each example is self-contained: `.neptune.yaml` (with explicit `stacks_management: terramate` or `local`), stacks, and a README with required environment variables and usage.

## Neptune config in CI

Neptune loads `.neptune.yaml` from the **default branch** of your repository in CI (so PR authors cannot change workflow steps). See [Neptune documentation](https://github.com/devopsfactory-io/neptune/tree/main/docs) for [configuration](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) and [object storage](https://github.com/devopsfactory-io/neptune/blob/main/docs/object-storage.md).

## Using these examples

The examples are in this repo under `examples/` (e.g. `examples/s3-backend/`, `examples/automerge/`). Copy or adapt the example you need; set the required env vars and replace placeholder bucket names in `.neptune.yaml`.
