# Terragrunt with Terramate stacks example

This example uses **Terragrunt** with **Terramate** stacks. Neptune runs each workflow step (e.g. `terragrunt init`, `terragrunt plan`, `terragrunt apply`) **once per changed stack**, with the working directory set to that stack. No Terramate CLI is required for that: Neptune uses the Terramate SDK to list changed stacks and run in order.

- **Layout**: Each stack (`stack-a/`, `stack-b/`) has a `terragrunt.hcl` that sources a shared module (`_modules/minimal`). The Terraform backend in the module is `local` so no cloud state backend is required for this example.
- **Alternative**: You can set `once: true` on a step and run `terramate run --changed -- terragrunt plan` (and similar for apply); in that case the Terramate CLI must be installed, and Neptune runs the command once at repo root.

Replace `s3://your-bucket` in `.neptune.yaml` and set object-storage env vars as in the [s3-backend](s3-backend/) example. Ensure Terragrunt is installed in the environment where Neptune runs (e.g. CI).
