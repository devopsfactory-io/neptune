# Usage

## CLI commands

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

Typically you run `neptune command plan` or `neptune command apply` from CI (e.g. GitHub Actions) with the required environment variables set; see [Installation](installation.md).

## Commit statuses

Neptune sets GitHub commit statuses on the PR head commit:

- **neptune plan** – Set to pending when plan starts, then success or failure when plan finishes.
- **neptune apply** – After a successful plan, set to pending with a message that the PR cannot be merged until apply is run; set to pending when apply starts, then success or failure when apply finishes.

To block merging until apply has run successfully, add **neptune apply** as a [required status check](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-a-branch-protection-rule) in your branch protection rules. Your CI workflow must grant `statuses: write` to the token (e.g. `GITHUB_TOKEN`) so Neptune can create these statuses.
