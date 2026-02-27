# Configuration

Neptune reads repository configuration from a `.neptune.yaml` file in the root of your Infrastructure as Code repository.

## Log level

You can control how verbose Neptune's logs are:

- **Config file**: Set `log_level` at the top level of `.neptune.yaml` to one of `DEBUG`, `INFO`, or `ERROR` (case-insensitive). Default is `INFO`.
- **Environment**: Set `NEPTUNE_LOG_LEVEL` to one of `DEBUG`, `INFO`, or `ERROR`. This overrides the config file value.

Use `DEBUG` for detailed output (e.g. each step and lock operation); use `ERROR` to only see errors. Log lines include a source (e.g. `neptune.config`, `neptune.lock`, `neptune.run`), and bordered banners are printed for the command, config summary, requirements check, lock, runner, and steps summary.

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

### Top-level optional fields

- **log_level**: One of `DEBUG`, `INFO`, or `ERROR`. Default `INFO`. Overridden by the `NEPTUNE_LOG_LEVEL` environment variable.

### Workflows

Workflows define `plan` and `apply` phases. Each phase has `steps`:

- **steps**: List of steps. Each step has `run: <shell command>` and optional **terramate** and **changed** (see below).
- **apply.depends_on**: Optional list of phases that must have run first (e.g. `plan`).

#### Step options: terramate and changed

- **terramate** (optional, default: `true`): When `true` (or omitted), Neptune runs the step’s `run` command **once per changed stack**, with the working directory set to each stack (using the Terramate SDK; no Terramate CLI needed for this). You can write e.g. `run: terragrunt plan` and Neptune will execute it in each changed stack.
- **terramate: false**: Run the step’s command **once** in the process current directory (e.g. repo root). Use this if you still invoke the Terramate CLI yourself, e.g. `run: terramate run --changed -- terragrunt plan`.
- **changed** (optional): When `terramate` is true, Neptune already runs only in changed stacks. Use `changed: true` only for clarity in config; no extra logic.

Example with per-stack execution (default):

```yaml
steps:
  - run: terragrunt init -upgrade
  - run: terragrunt plan
  - run: terragrunt apply -auto-approve
```

Example with a single global command:

```yaml
steps:
  - run: terramate run --changed -- some-global-script.sh
    terramate: false
```

See [.neptune.example.yaml](../.neptune.example.yaml) in the repo root for a full example.

## Terramate requirement

The repository must be a [Terramate](https://github.com/terramate-io/terramate) project (root and stack config) so Neptune can detect changed stacks and their run order. Neptune uses the Terramate Go SDK for this; the Terramate CLI is **not** required for listing changed stacks or for running steps when `terramate` is true (default). When a step has `terramate: false` and your `run` string invokes `terramate run ...`, the Terramate CLI must be installed in the environment where Neptune runs (e.g. CI).
