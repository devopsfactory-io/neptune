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
