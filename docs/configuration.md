# Configuration

Neptune reads repository configuration from a `.neptune.yaml` file in the root of your Infrastructure as Code repository.

## Where Neptune loads the config (security)

In CI (when not in E2E mode), Neptune loads `.neptune.yaml` from the **default branch** of the repository using git (e.g. `git show origin/main:.neptune.yaml`). If the file is not present on the default branch (e.g. first-time setup), Neptune tries the **PR branch** (HEAD). This ensures that workflow steps cannot be changed by a PR author: only the version of the config on the default branch (or, as fallback, on the PR branch) is used. In **E2E mode** (`NEPTUNE_E2E=1`) or when the repo is not a git repository, Neptune reads the config from the local filesystem.

## Log level

You can control how verbose Neptune's logs are:

- **Config file**: Set `log_level` at the top level of `.neptune.yaml` to one of `DEBUG`, `INFO`, or `ERROR` (case-insensitive). Default is `ERROR`.
- **Environment**: Set `NEPTUNE_LOG_LEVEL` to one of `DEBUG`, `INFO`, or `ERROR`. This overrides the config file value.
- **CLI**: Pass **--log-level** to any command (e.g. `neptune command plan --log-level DEBUG`, `neptune stacks list --log-level INFO`). This overrides both the config file and `NEPTUNE_LOG_LEVEL`. The flag is applied before config is loaded, so it affects all logging (including config load messages).

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
  # automerge: false  # optional; when true, enable PR auto-merge after successful apply

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
- **stacks_management** (optional, default: `terramate`): How Neptune discovers stacks. Use **`terramate`** to use the Terramate SDK (repository must be a Terramate project). Use **`local`** to discover stacks via the root-level **local_stacks** key (config or discovery) and git-based change detection; then use **neptune stacks list** (with optional **--changed**) and **neptune stacks create**.
- **branch**: Base branch (e.g. `master` or `main`).
- **plan_requirements**: Requirements that must be met before running plan (e.g. `undiverged`, `rebased`).
- **apply_requirements**: Requirements that must be met before apply (e.g. `approved`, `mergeable`, `undiverged`, `rebased`).
- **allowed_workflow**: Name of the workflow to run (e.g. `default`).
- **automerge** (optional, default: `false`): When `true`, after a successful apply Neptune adds a comment that the PR will be auto-merged and calls GitHub to enable auto-merge on the PR (the PR will merge when all required checks pass). The GitHub token must have permission to enable auto-merge (e.g. `pull_requests: write`). Auto-merge is only triggered on successful **apply**, not on plan.

### Top-level optional fields

- **log_level**: One of `DEBUG`, `INFO`, or `ERROR`. Default `ERROR`. Overridden by the `NEPTUNE_LOG_LEVEL` environment variable and by the global **--log-level** CLI flag.
- **local_stacks** (optional, only when `stacks_management: local`): Root-level key (sibling of `repository` and `workflows`) for local stack discovery. **source**: `config` (use the **stacks** list) or `discovery` (scan the repo for directories containing **stack.hcl**). **stacks**: list of `{ path: "<dir>", depends_on: ["<dir>"] }` for run order; used when source is `config`. When source is **discovery**, each **stack.hcl** can optionally declare `depends_on = ["<path>", ...]` inside its `stack` block to control run order (see below).

### Workflows

Workflows define `plan` and `apply` phases. Each phase has `steps`:

- **steps**: List of steps. Each step has `run: <shell command>` and optional **once** (see below).
- **apply.depends_on**: Optional list of phases that must have run first (e.g. `plan`).

#### Step options: once

- **once** (optional, default: `false`): When `false` (or omitted), Neptune runs the step’s `run` command **once per changed stack**, with the working directory set to each stack. When `true`, Neptune runs the command **once** in the repo root. Use `once: true` if you invoke the Terramate CLI yourself, e.g. `run: terramate run --changed -- terragrunt plan`.

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
    once: true
```

See [.neptune.example.yaml](../.neptune.example.yaml) in the repo root for a full example.

## Stacks management: terramate vs local

- **terramate** (default): The repository must be a [Terramate](https://github.com/terramate-io/terramate) project (root and stack config) so Neptune can detect changed stacks and their run order. Neptune uses the Terramate Go SDK for this; the Terramate CLI is **not** required for listing changed stacks or for running steps when `once` is false. When a step has `once: true` and your `run` string invokes `terramate run ...`, the Terramate CLI must be installed (e.g. in CI).
- **local**: Set `stacks_management: local` and optionally `local_stacks` (source **config** or **discovery**). Neptune discovers stacks and filters by git changes. Use **neptune stacks list** and **neptune stacks list --changed**; use **neptune stacks create &lt;name&gt;** to scaffold a new stack with **stack.hcl**. When using **discovery** (directories containing **stack.hcl**), you can set run order via an optional **depends_on** attribute in the `stack` block: a list of **paths** (repo-root-relative or relative to the stack’s directory, e.g. `../foundation`). Each path can be a single stack directory or a directory that contains multiple stacks; in the latter case, this stack runs after **all** stacks inside that folder. Paths are path-only (no stack name resolution). Same run-order semantics as config-based `depends_on`.
