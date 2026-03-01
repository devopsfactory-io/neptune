# Development

Contributing workflow (issues, pull requests, checklist) is described in [CONTRIBUTING.md](../CONTRIBUTING.md). For pull requests, ensure your commits are signed off (DCO); see the [Sign-off (DCO)](../CONTRIBUTING.md#pull-requests) section for git config and `git commit -s`.

## End-to-end tests

E2E tests run Neptune against MinIO via Docker Compose and do not require a real GitHub PR. See [e2e/README.md](../e2e/README.md) for prerequisites and how to run `./e2e/run.sh`. PR emulation is done with an isolated git repo (main + pr-1 branch with changed stacks) and `NEPTUNE_E2E=1` so GitHub checks are skipped.

## Go

```bash
make build      # build binary
make test-all   # run tests
make check-fmt  # check formatting
make lint       # run golangci-lint (optional)
```

Use the Go version from `go.mod`. See [AGENTS.md](../AGENTS.md) for code style, testing, and CI.
