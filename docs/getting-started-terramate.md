# Getting started with Neptune (Terramate stacks)

This guide walks you through setting up Neptune with **GitHub Actions** and the **neptbot GitHub App** when your repository uses **Terramate** for stack layout and change detection. Neptune runs plan and apply in GitHub Actions; neptbot triggers **plan** when a PR is opened or updated, and **apply** when someone comments `@neptbot apply` on the PR. For why apply-before-merge matters, see [Workflow comparison](workflow-comparison.md).

## Requirements

- **Git host**: GitHub (GitHub.com or GitHub Enterprise).
- **Repository**: Terraform or OpenTofu stacks with **remote state** (e.g. S3 or GCS backend). Plan and apply must be reproducible in CI; avoid local state only.
- **Object storage for Neptune**: A bucket in Google Cloud Storage (GCS) or AWS S3 (or S3-compatible storage such as MinIO) for Neptune's stack lock files. See [Object storage](object-storage.md) for credentials and environment variables.
- **Terramate project**: Your repo must be a [Terramate](https://github.com/terramate-io/terramate) project: a root `terramate.tm.hcl` and each stack directory with a `stack.tm.hcl` file.

## Repository structure

A typical layout for Terramate stacks:

```
your-infra-repo/
├── terramate.tm.hcl      # Root Terramate config
├── .neptune.yaml         # Neptune config (see Step 3)
├── stack-a/
│   ├── stack.tm.hcl      # Terramate stack definition
│   └── main.tf           # Terraform/OpenTofu (with backend block for remote state)
├── stack-b/
│   ├── stack.tm.hcl
│   └── main.tf
└── .github/
    └── workflows/
        └── neptune.yml   # repository_dispatch workflow (see Step 4)
```

You can copy structure and config from the [terramate-stacks example](../examples/terramate-stacks/) in this repo.

## Step 1: Install neptbot

Install the **neptbot GitHub App** on your organization or user account so that PR events and comments (e.g. `@neptbot apply`) can trigger your workflow. Install from [github.com/apps/neptbot](https://github.com/apps/neptbot). No Lambda or AWS setup is required; the Neptune project runs the webhook endpoint.

**PR label required:** Pull requests must have the label **neptune** to trigger the `repository_dispatch` workflow. You can auto-apply the label using the [labeler](https://github.com/actions/labeler) GitHub Action: add a workflow (e.g. `.github/workflows/labeler.yml`) and a config file (e.g. `.github/labeler.yml`) so that PRs touching your infra paths get the `neptune` label. See the [terramate-stacks example](../examples/terramate-stacks/) (`.github/workflows/labeler.yml` and `.github/labeler.yml`) for reference.

See [GitHub App and Lambda — Default: Install neptbot](github-app-and-lambda.md#default-install-neptbot) for details.

## Step 2: Object storage for Neptune

Create a bucket (or use an existing one) in GCS or S3 for Neptune's stack lock files. Configure credentials so that your GitHub Actions workflow can access it:

- **AWS S3**: Set `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` (e.g. as repository secrets/variables). Alternatively use OIDC or another credential method.
- **GCS**: Set `GOOGLE_APPLICATION_CREDENTIALS` to a service account key path, or use Application Default Credentials.

Use the same bucket (and optional prefix) in `.neptune.yaml` as the `object_storage` URL (e.g. `s3://your-bucket` or `gs://your-bucket`). Full details: [Object storage](object-storage.md).

## Step 3: Add `.neptune.yaml`

Add a `.neptune.yaml` file in the root of your repository. Example for Terramate with Terraform:

```yaml
repository:
  object_storage: s3://your-bucket   # or gs://your-bucket for GCS
  stacks_management: terramate
  branch: main
  plan_requirements:
    - undiverged
  apply_requirements:
    - approved
    - mergeable
    - undiverged
  allowed_workflow: default
  # automerge: true   # optional: enable PR auto-merge after successful apply

workflows:
  default:
    plan:
      steps:
        - run: terraform init -input=false
        - run: terraform plan -input=false -out=tfplan
    apply:
      depends_on:
        - plan
      steps:
        - run: terraform apply -input=false tfplan
```

Neptune runs each step **once per changed stack** (with the working directory set to that stack). For the full schema and options (e.g. `log_level`, `once: true` for global steps), see [Configuration](configuration.md).

## Step 4: Add the `repository_dispatch` workflow

Each repository that has neptbot (or your own GitHub App) installed must have a workflow that listens for `repository_dispatch` with type `neptune-command`. Create `.github/workflows/neptune.yml`:

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
          NEPTUNE_VERSION="v1.0.0"
          curl -sSL "https://github.com/devopsfactory-io/neptune/releases/download/${NEPTUNE_VERSION}/neptune_${NEPTUNE_VERSION}_linux_amd64.tar.gz" | tar -xz -C /usr/local/bin neptune

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
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: ${{ vars.AWS_REGION }}
```

If you use **OpenTofu** instead of Terraform, replace the "Setup Terraform" step with an OpenTofu setup action (e.g. `opentofu/setup-opentofu` or install OpenTofu manually). For GCS, set `GOOGLE_APPLICATION_CREDENTIALS` and ensure the credentials file is available in the runner. If you enable `repository.automerge`, add `pull_requests: write` to the workflow `permissions`. See [GitHub App and Lambda — Workflow on `repository_dispatch`](github-app-and-lambda.md#workflow-on-repository_dispatch) for notes on concurrency, checkout, and env vars.

## Step 5: Branch protection (recommended)

To block merging until apply has run successfully, add **neptune apply** as a [required status check](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-a-branch-protection-rule) in your branch protection rules. The workflow above uses `statuses: write` so Neptune can set the **neptune plan** and **neptune apply** commit statuses. See [Branch protection (recommended)](github-app-and-lambda.md#5-branch-protection-recommended).

## Step 6: End-to-end flow

Once the steps above are in place, the flow is:

1. **Open a PR** — Pushing a branch and opening a pull request triggers neptbot, which dispatches the workflow with `command: plan`. Neptune runs plan for each changed stack and posts a plan comment on the PR.
2. **Plan comment** — Review the plan output in the PR comment and the **neptune plan** commit status.
3. **Comment `@neptbot apply`** — When the plan looks good and the PR is approved, comment `@neptbot apply` on the PR. neptbot triggers the workflow with `command: apply`. Neptune runs apply for each changed stack and posts an apply comment.
4. **Apply comment** — After a successful apply, the **neptune apply** status is set to success. If you added **neptune apply** as a required check, the PR becomes mergeable.
5. **Auto-merge (optional)** — If you set `repository.automerge: true` in `.neptune.yaml`, Neptune enables PR auto-merge after a successful apply; the PR will merge when all required checks pass.

You can add screenshots here to illustrate each step (e.g. PR opened, plan comment, `@neptbot apply` comment, apply comment, and auto-merge).
