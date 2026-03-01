# Local stacks (config) example

This example uses **stacks_management: local** with the root-level **local_stacks** key in `.neptune.yaml`: stack list and run order come from the config instead of scanning for `stack.hcl`. Neptune still runs each workflow step **once per changed stack**, with the working directory set to that stack.

- **Config**: **local_stacks** is a root-level key (sibling of `repository` and `workflows`). Use `local_stacks.source: config` and `local_stacks.stacks` to list stack paths (relative to repo root) and optional `depends_on` for run order. Run order is a topological sort of dependencies; stacks with no deps can appear in any order.
- **Layout**: `stack-a/` and `stack-b/` (with optional `stack.hcl` and Terraform). The paths in `local_stacks.stacks` must match existing directories. Here `stack-b` depends on `stack-a`, so Neptune runs steps in `stack-a` before `stack-b`.
- **When to use**: Prefer config when you want an explicit, version-controlled stack list and order without relying on filesystem discovery. Use [local-stacks](local-stacks/) when you prefer discovery via `stack.hcl`.

Replace `s3://your-bucket` in `.neptune.yaml` and set object-storage env vars as in the [s3-backend](s3-backend/) example.

See [Neptune configuration – Stacks management](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) for `local_stacks` and `stacks_management: local`.
