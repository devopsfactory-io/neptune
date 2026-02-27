# Installation

## Go (recommended)

```bash
# Build from source
go build -o neptune .

# Or install into $GOPATH/bin
go install .
```

Binaries for Linux, macOS, and Windows are published to the [GitHub Releases](https://github.com/kaio6fellipe/neptune/releases) page when you push a version tag (e.g. `v0.2.0`) via GoReleaser.

## Python (legacy)

```bash
# From the repository root
pip install -e .

# Enable shell completion (optional)
neptune --install-completion
```

## Using with GitHub Actions

Set the same environment variables in your workflow (`GITHUB_REPOSITORY`, `GITHUB_PULL_REQUEST_BRANCH`, `GITHUB_PULL_REQUEST_NUMBER`, `GITHUB_PULL_REQUEST_COMMENT_ID`, `GITHUB_RUN_ID`, `GITHUB_TOKEN`), then run `neptune command plan` or `neptune command apply` as needed. See [Usage](usage.md) for CLI commands.

## Triggering Neptune via GitHub App

The default way to trigger Neptune from PR open and comments (e.g. `@neptbot apply`) is to **install the Neptune project's neptbot GitHub App** on your repos and add the workflow described in [GitHub App and Lambda](github-app-and-lambda.md). You can also self-host by creating your own GitHub App and deploying the [Lambda](lambda/README.md) in this repo.
