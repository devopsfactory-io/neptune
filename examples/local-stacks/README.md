# Local stacks example

This example uses **stacks_management: local** so Neptune discovers stacks by scanning the repository for directories that contain a **stack.hcl** file (no Terramate project required). Neptune runs each workflow step **once per changed stack**, with the working directory set to that stack—same execution model as Terramate, but stack list and run order come from Neptune.

- **Layout**: `stack-a/` and `stack-b/` each have a `stack.hcl` (minimal `stack { name = "..." }`) and Terraform (null_resource, local_file). No Terramate root or stack.tm.hcl.
- **Discovery**: With `stacks_management: local` and no root-level **local_stacks** (or `local_stacks.source: discovery`), Neptune finds all directories containing `stack.hcl` and filters by git changes for plan/apply.
- **CLI**: Use **neptune stacks list** to list all stacks, **neptune stacks list --changed** to list only stacks with changes, and **neptune stacks create &lt;name&gt;** to scaffold a new stack with `stack.hcl`.

Replace `s3://your-bucket` in `.neptune.yaml` and set object-storage env vars as in the [s3-backend](s3-backend/) example.

See [Neptune configuration – Stacks management](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) for `stacks_management: terramate` vs `local` and `stack.hcl` format.
