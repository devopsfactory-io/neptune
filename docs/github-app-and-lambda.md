# GitHub App and Lambda (webhook trigger)

You can trigger Neptune from webhooks: when a PR is opened or updated, **neptune plan** runs; when someone comments e.g. `@neptune-bot apply` on a PR, **neptune apply** runs. The default is to **install the Neptune project's neptune-bot GitHub App** on your repos—no need to create an app or run infrastructure. Alternatively, you can self-host by creating your own GitHub App and deploying the [Lambda](../lambda/) in this repo.

You can also run Neptune from a workflow that triggers on `pull_request` and/or `workflow_dispatch` (see [Installation](installation.md)).

## Default: Install neptune-bot

1. **Install the neptune-bot GitHub App** on your organization or user account. (Install link to be added when the app is published.)
2. **Add the workflow below** to each repository where you want Neptune to run. The app will trigger it via `repository_dispatch` when a PR is opened/updated or when someone comments `@neptune-bot apply` or `@neptune-bot plan`.
3. Configure **object storage** (e.g. S3) and a **`.neptune.yaml`** in the repo as required by Neptune (see [Configuration](configuration.md) and [Object storage](object-storage.md)).
4. Optionally add **branch protection** so that **neptune apply** is a required status check (see [Branch protection (recommended)](#5-branch-protection-recommended)).

No Lambda or AWS setup is required; the Neptune project runs the webhook endpoint.

### Workflow on `repository_dispatch`

Each repository that has the Neptune GitHub App installed (neptune-bot or your own) must have a workflow that listens for the dispatch event. Example:

```yaml
name: neptune

on:
  repository_dispatch:
    types: [neptune-command]

concurrency:
  group: neptune-${{ github.event.client_payload.pull_request_number }}
  cancel-in-progress: false

permissions:
  contents: read
  pull-requests: write
  statuses: write

jobs:
  neptune:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          ref: refs/pull/${{ github.event.client_payload.pull_request_number }}/head

      - name: Install Neptune
        run: |
          NEPTUNE_VERSION="v0.2.0"
          curl -sSL "https://github.com/kaio6fellipe/neptune/releases/download/${NEPTUNE_VERSION}/neptune_${NEPTUNE_VERSION}_linux_amd64.tar.gz" | tar -xz -C /usr/local/bin neptune

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.5.0"

      - name: Run Neptune
        run: neptune command ${{ github.event.client_payload.command }}
        env:
          GITHUB_REPOSITORY: ${{ github.repository }}
          GITHUB_PULL_REQUEST_NUMBER: ${{ github.event.client_payload.pull_request_number }}
          GITHUB_PULL_REQUEST_BRANCH: ${{ github.event.client_payload.pull_request_branch }}
          GITHUB_RUN_ID: ${{ github.run_id }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_PULL_REQUEST_COMMENT_ID: ""
          # Object storage and .neptune.yaml must be configured (see docs/object-storage.md and docs/configuration.md)
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: ${{ vars.AWS_REGION }}
```

Notes:

- **concurrency**: For Terraform/Neptune workflows, use `cancel-in-progress: false` so in-progress runs are never cancelled when a new event arrives.
- **checkout**: Uses `refs/pull/<number>/head` so the job runs on the PR branch; the payload's `pull_request_branch` and `pull_request_sha` are available if you need them.
- **Neptune env**: `GITHUB_REPOSITORY`, `GITHUB_PULL_REQUEST_NUMBER`, `GITHUB_PULL_REQUEST_BRANCH`, `GITHUB_RUN_ID`, `GITHUB_TOKEN`, and `GITHUB_PULL_REQUEST_COMMENT_ID` (can be empty) are required by Neptune.

### 4. Comment format

To run **apply** or **plan** from a PR comment, write a comment that:

- Mentions the app. When using **neptune-bot**, use `@neptune-bot`. If you self-host with a different app slug, use that slug (e.g. `@your-app-slug`).
- Contains the word **apply** or **plan**.

Examples:

- `@neptune-bot apply`
- `@neptune-bot plan`
- `Please run @neptune-bot apply when ready`

Only comments on **pull requests** (not on issues) are handled.

### 5. Branch protection (recommended)

To block merging until apply has run successfully, add **neptune apply** as a [required status check](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-a-branch-protection-rule) in your branch protection rules. The workflow above uses `statuses: write` so Neptune can set the **neptune plan** and **neptune apply** commit statuses.

## Alternative: Self-hosted (your own GitHub App and Lambda)

If you want to run your own GitHub App and Lambda (e.g. in your AWS account), use the Lambda code and CloudFormation in this repo. Your repositories will use the **same** `repository_dispatch` workflow and payload as above; the only difference is who hosts the app and webhook endpoint.

### 1. Create your own GitHub App

- Create a [GitHub App](https://docs.github.com/en/apps/creating-github-apps) (e.g. under your user or org).
- **Webhook**: Leave "Active" checked; set **Payload URL** to your Lambda Function URL (you get this after deploying the stack; see [lambda/README.md](../lambda/README.md)).
- **Webhook secret**: Generate a secret and store it in AWS Secrets Manager (see below).
- **Permissions**: Repository → **Contents** (Read), **Pull requests** (Read), **Metadata** (Read).
- **Subscribe to events**: **Pull requests**, **Issue comments**.
- **Private key**: Generate and download the PEM; store it in AWS Secrets Manager.
- Install the App on the repositories where you want Neptune to run.

### 2. Deploy the Lambda

- Build the Lambda binary and zip it (see [lambda/README.md](../lambda/README.md#build)).
- Create two secrets in **AWS Secrets Manager**: one with the webhook secret (plain string), one with the private key (full PEM).
- Upload the zip to S3 and deploy the CloudFormation stack from [lambda/cloudformation/template.yaml](../lambda/cloudformation/template.yaml) (see [lambda/README.md](../lambda/README.md#deploy-with-cloudformation)).
- Copy the stack output **WebhookUrl** into your GitHub App's Payload URL.

### 3. Add the workflow to your repositories

Add the same workflow shown in [Workflow on `repository_dispatch`](#workflow-on-repository_dispatch) above to each repo that has your App installed.

## Payload sent by the Lambda

The webhook handler (neptune-bot or your Lambda) triggers `repository_dispatch` with:

- **event_type**: `neptune-command`
- **client_payload**:
  - `command`: `"plan"` or `"apply"`
  - `pull_request_number`: PR number
  - `pull_request_branch`: head ref (e.g. `feature/my-branch`)
  - `pull_request_sha`: head commit SHA (optional but useful)
  - `pull_request_repo_full`: `owner/repo` (optional)

Your workflow reads these from `github.event.client_payload` and passes the required env to the `neptune` CLI.
