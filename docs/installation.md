# Installation

For the recommended setup (GitHub Actions + neptbot), see [Getting started (Terramate)](getting-started-terramate.md) or [Getting started (Local stacks)](getting-started-local-stacks.md). For why apply-before-merge matters and how Neptune compares to other Terraform workflows, see [Workflow comparison](workflow-comparison.md).

## Install Neptune

```bash
# Build from source
go build -o neptune .

# Or install into $GOPATH/bin
go install .
```

Binaries for Linux, macOS, and Windows are published to the [GitHub Releases](https://github.com/devopsfactory-io/neptune/releases) page when you push a version tag (e.g. `v0.2.0`) via GoReleaser.

## Triggering Neptune via GitHub App

The default way to trigger Neptune from PR open and comments (e.g. `@neptbot apply`) is to **install the Neptune project's [neptbot GitHub App](https://github.com/apps/neptbot)** on your repos and add the `repository_dispatch` workflow. The getting-started guides and [GitHub App and Lambda](github-app-and-lambda.md) describe the workflow and required environment variables (including object storage). You can also self-host by creating your own GitHub App and deploying the [Lambda](https://github.com/devopsfactory-io/neptune/blob/main/lambda/README.md) in this repo.
