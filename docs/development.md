# Development

Contributing workflow (issues, pull requests, checklist) is described in [CONTRIBUTING.md](../CONTRIBUTING.md). For pull requests, ensure your commits are signed off (DCO); see the [Sign-off (DCO)](../CONTRIBUTING.md#pull-requests) section for git config and `git commit -s`.

## End-to-end tests

E2E tests run Neptune against MinIO via Docker Compose and do not require a real GitHub PR. See [e2e/README.md](../e2e/README.md) for prerequisites and how to run `./e2e/scripts/run-terramate.sh`. PR emulation is done with an isolated git repo (main + pr-1 branch with changed stacks) and `NEPTUNE_E2E=1` so GitHub checks are skipped.

## Go

```bash
make build      # build binary
make test-all   # run tests
make check-fmt  # check formatting
make lint       # run golangci-lint (optional)
```

Use the Go version from `go.mod`. See [AGENTS.md](../AGENTS.md) for code style, testing, and CI.

## Release Process

Neptune uses semantic versioning and [GoReleaser](https://goreleaser.com/) to create releases. Only maintainers with push access can create releases.

### Semantic Versioning

Tags follow the pattern `vMAJOR.MINOR.PATCH` (e.g., `v0.2.0`, `v1.0.0`). The `v` prefix is required.

- **MAJOR**: Breaking changes or significant architectural changes
- **MINOR**: New features, backward-compatible
- **PATCH**: Bug fixes, backward-compatible

### Creating a Release

1. **Ensure CI is green**: All tests, linting, and checks must pass on `main` (or the branch you're releasing from).

2. **Create and push the tag**:

   ```bash
   git tag v0.3.0
   git push origin v0.3.0
   ```

3. **GoReleaser runs automatically**: The [`.github/workflows/release.yml`](../.github/workflows/release.yml) workflow triggers on tag push and runs GoReleaser with the configuration from [`.goreleaser.yml`](../.goreleaser.yml).

4. **Release artifacts**: GoReleaser creates a [GitHub Release](https://github.com/devopsfactory-io/neptune/releases) with:
   - `neptune` binaries for multiple OS/arch (Linux, macOS, Windows; amd64, arm64)
   - `neptune-webhook.zip` (AWS Lambda handler for the GitHub App)
   - `checksums.txt` (SHA256 checksums for all artifacts)
   - Auto-generated changelog from commit messages since the previous tag

The version, commit SHA, and build date are injected into the binary via ldflags at build time (see `main.version`, `main.commit`, `main.date`).

### Manual Release (if needed)

If you need to run GoReleaser locally (e.g., for testing or troubleshooting), ensure you have [GoReleaser installed](https://goreleaser.com/install/) and a `GITHUB_TOKEN` with `contents: write` permission:

```bash
export GITHUB_TOKEN=your_token_here
goreleaser release --clean
```

For a dry run without publishing:

```bash
goreleaser release --snapshot --clean
```
