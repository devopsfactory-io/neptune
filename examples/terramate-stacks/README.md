# Terramate stacks example

This example shows **Terramate stacks** with plain Terraform. Neptune runs each workflow step **once per changed stack**, with the working directory set to that stack. The Terramate CLI is **not** required for this: Neptune uses the Terramate Go SDK for change detection and run order.

- **terramate: true** (default): Neptune executes the step’s `run` command in each changed stack directory.
- **terramate: false**: Run the command once (e.g. at repo root); use this if you invoke the Terramate CLI yourself, e.g. `run: terramate run --changed -- some-command`.

See [Neptune configuration – Step options: terramate and changed](https://github.com/devopsfactory-io/neptune/blob/main/docs/configuration.md) for details.

Replace `s3://your-bucket` in `.neptune.yaml` and set object-storage env vars as in the [s3-backend](s3-backend/) example.
