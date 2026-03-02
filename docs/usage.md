# Usage

The primary way to use Neptune is **GitHub Actions and PR interaction**: open a PR and Neptune runs **plan** (triggered by the neptbot GitHub App); when the plan looks good, comment **`@neptbot apply`** on the PR to run **apply**. Merge only when apply has succeeded. For full setup, see [Getting started (Terramate)](getting-started-terramate.md) or [Getting started (Local stacks)](getting-started-local-stacks.md).

## Flow

- **Plan** runs when a PR is opened or updated (neptbot sends `repository_dispatch` with `command: plan`). Neptune posts a plan comment and sets the **neptune plan** commit status.
- **Apply** runs when someone comments `@neptbot apply` (or `@neptbot plan` to re-run plan) on the PR. Neptune runs apply, posts an apply comment, and sets the **neptune apply** commit status.
- To block merging until apply has run, add **neptune apply** as a [required status check](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-a-branch-protection-rule) in branch protection. Your workflow must grant `statuses: write` so Neptune can create these statuses.

## Commit statuses

Neptune sets GitHub commit statuses on the PR head commit:

- **neptune plan** – Set to pending when plan starts, then success or failure when plan finishes.
- **neptune apply** – After a successful plan, set to pending with a message that the PR cannot be merged until apply is run; set to pending when apply starts, then success or failure when apply finishes.

## CLI reference

When running in CI, the workflow runs `neptune command ${{ github.event.client_payload.command }}`. For local runs or debugging you can use:

```bash
# Show help
neptune --help

# Print version
neptune version

# Run a workflow phase (plan or apply)
neptune command plan
neptune command apply

# Unlock all stacks (requires --all)
neptune unlock --all
```

Set the same environment variables as in the workflow (`GITHUB_REPOSITORY`, `GITHUB_PULL_REQUEST_NUMBER`, `GITHUB_PULL_REQUEST_BRANCH`, `GITHUB_RUN_ID`, `GITHUB_TOKEN`, and object storage vars); see [Installation](installation.md) and the getting-started guides.
