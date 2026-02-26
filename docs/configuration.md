# Configuration

Neptune reads repository configuration from a `.neptune.yaml` file in the root of your Infrastructure as Code repository.

## Repository configuration

Your repository must have a `.neptune.yaml` file with the following structure:

```yaml
repository:
  # Object storage URL: gs://bucket or gs://bucket/prefix (GCS), s3://bucket or s3://bucket/prefix (S3 or S3-compatible e.g. MinIO)
  # See docs/object-storage.md for backend-specific env vars.
  object_storage: gs://object_storage_url
  branch: master
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
    - mergeable
    - undiverged
  allowed_workflow: default

workflows:
  default:
    plan:
      steps:
        - run: echo "Custom command"
        - run: terramate run --parallel $(nproc --all) --changed -- terragrunt init -upgrade
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on:
        - plan
      steps:
        - run: echo "Custom command"
        - run: terramate run --changed -- terragrunt apply -auto-approve
```

### Repository fields

- **object_storage**: URL for stack lock files. Use `gs://bucket` or `gs://bucket/prefix` for GCS, `s3://bucket` or `s3://bucket/prefix` for AWS S3 or S3-compatible storage (e.g. MinIO). See [Object storage](object-storage.md) for credentials and env vars.
- **branch**: Base branch (e.g. `master` or `main`).
- **plan_requirements**: Requirements that must be met before running plan (e.g. `undiverged`, `rebased`).
- **apply_requirements**: Requirements that must be met before apply (e.g. `approved`, `mergeable`, `undiverged`, `rebased`).
- **allowed_workflow**: Name of the workflow to run (e.g. `default`).

### Workflows

Workflows define `plan` and `apply` phases. Each phase has `steps`:

- **steps**: List of `run: <shell command>` steps. Neptune runs them in order.
- **apply.depends_on**: Optional list of phases that must have run first (e.g. `plan`).

You can add custom commands before or after Terramate/Terragrunt; the example shows a placeholder `echo "Custom command"` and the typical `terramate run --changed -- ...` pattern.

## Terramate requirement

The repository must use [Terramate](https://github.com/terramate-io/terramate) to orchestrate Terraform stacks and support the `--changed` flag so Neptune can run commands only for stacks changed in the PR.

Example usage in steps:

- `terramate run --parallel $(nproc --all) --changed -- terragrunt init -upgrade`
- `terramate run --changed -- terragrunt plan`
- `terramate run --changed -- terragrunt apply -auto-approve`

See [.neptune.example.yaml](../.neptune.example.yaml) in the repo root for a full example.
