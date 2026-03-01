---
description: .neptune.yaml schema and env vars; keep example and docs in sync
globs: "**/*.yaml,**/*.yml,**/.neptune*"
alwaysApply: false
---

# Config and YAML

- **.neptune.yaml**: Top-level keys are `repository` and `workflows`. `repository` has `object_storage` (must be `gs://...`), `branch`, `plan_requirements`, `apply_requirements`, `allowed_workflow`. `workflows` is a map of workflow name to phases (e.g. `plan`, `apply`) with `steps` (list of `run: "..."`) and optional `depends_on`.
- **Required env vars**: `NEPTUNE_CONFIG_PATH`, `GITHUB_REPOSITORY`, `GITHUB_PULL_REQUEST_BRANCH`, `GITHUB_PULL_REQUEST_NUMBER`, `GITHUB_RUN_ID`, `GITHUB_TOKEN`.
- **Validation**: Enforced in `internal/config` (object_storage prefix, allowed requirements, plan+apply phases, terramate+`--changed` for terraform/terragrunt steps).
- When changing schema or env: update README, `.neptune.example.yaml`, **examples/** (if an example uses the changed config), and AGENTS.md or rules so human and AI docs stay in sync.
