# Getting started with Neptune (Local stacks)

This guide walks you through setting up Neptune with **GitHub Actions** and the **neptbot GitHub App** when your repository uses **local** stack management (no Terramate). Neptune runs plan and apply in GitHub Actions; neptbot triggers **plan** when a PR is opened or updated, and **apply** when someone comments `@neptbot apply` on the PR. For why apply-before-merge matters, see [Workflow comparison](workflow-comparison.md).

## Requirements

- **Git host**: GitHub (GitHub.com or GitHub Enterprise).
- **Repository**: Terraform or OpenTofu stacks with **remote state** (e.g. S3 or GCS backend). Plan and apply must be reproducible in CI; avoid local state only.
- **Object storage for Neptune**: A bucket in Google Cloud Storage (GCS) or AWS S3 (or S3-compatible storage such as MinIO) for Neptune's stack lock files. See [Object storage](object-storage.md) for credentials and environment variables.
- **Local stacks**: Stacks are defined either by a **local_stacks** block in `.neptune.yaml` (source `config` with an explicit list of stack paths and optional `depends_on`) or by **discovery** (directories that contain a `stack.hcl` file). See [Configuration — Stacks management: terramate vs local](configuration.md#stacks-management-terramate-vs-local).

## Repository structure

Two common layouts:

**Option A — Discovery (each stack has `stack.hcl`):**

```
your-infra-repo/
├── .neptune.yaml         # Neptune config with stacks_management: local (no local_stacks, or source: discovery)
├── stack-a/
│   ├── stack.hcl         # Stack definition
│   └── main.tf           # Terraform/OpenTofu (with backend block for remote state)
├── stack-b/
│   ├── stack.hcl
│   └── main.tf
└── .github/
    └── workflows/
        └── neptune.yml   # repository_dispatch workflow (see Step 4)
```

**Option B — Config (explicit stack list in `.neptune.yaml`):**

Same tree, but `.neptune.yaml` includes a `local_stacks` block with `source: config` and a `stacks` list so you control run order and dependencies (e.g. `depends_on`).

You can copy structure and config from the [local-stacks](../examples/local-stacks/) or [local-stacks-config](../examples/local-stacks-config/) examples in this repo.

## Step 1: Install neptbot

Install the **neptbot GitHub App** on your organization or user account so that PR events and comments (e.g. `@neptbot apply`) can trigger your workflow. No Lambda or AWS setup is required; the Neptune project runs the webhook endpoint.

**PR label required:** Pull requests must have the label **neptune** to trigger the `repository_dispatch` workflow. You can auto-apply the label using the [labeler](https://github.com/actions/labeler) GitHub Action: add a workflow (e.g. `.github/workflows/labeler.yml`) and a config file (e.g. `.github/labeler.yml`) so that PRs touching your infra paths get the `neptune` label. See the [terramate-stacks](../examples/terramate-stacks/) or [automerge](../examples/automerge/) examples (`.github/workflows/labeler.yml` and `.github/labeler.yml`) for reference.

See [GitHub App and Lambda — Default: Install neptbot](github-app-and-lambda.md#default-install-neptbot) for the install link and details.

## Step 2: Object storage for Neptune

Create a bucket (or use an existing one) in GCS or S3 for Neptune's stack lock files. Configure credentials so that your GitHub Actions workflow can access it:

- **AWS S3**: Set `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` (e.g. as repository secrets/variables). Alternatively use OIDC or another credential method.
- **GCS**: Set `GOOGLE_APPLICATION_CREDENTIALS` to a service account key path, or use Application Default Credentials.

Use the same bucket (and optional prefix) in `.neptune.yaml` as the `object_storage` URL (e.g. `s3://your-bucket` or `gs://your-bucket`). Full details: [Object storage](object-storage.md).

## Step 3: Add `.neptune.yaml`

Add a `.neptune.yaml` file in the root of your repository.

**Example with discovery (no explicit stack list):** Neptune discovers stacks by scanning for directories that contain `stack.hcl`.

```yaml
repository:
  object_storage: s3://your-bucket
  stacks_management: local
  branch: main
  plan_requirements:
    - rebased
    - undiverged
  apply_requirements:
    - rebased
    - approved
    - mergeable
    - undiverged
  allowed_workflow: default

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

**Example with explicit stack list (config source):** Use `local_stacks` to define stacks and run order:

```yaml
repository:
  object_storage: s3://your-bucket
  stacks_management: local
  branch: main
  plan_requirements:
    - rebased
    - undiverged
  apply_requirements:
    - rebased
    - approved
    - mergeable
    - undiverged
  allowed_workflow: default

local_stacks:
  source: config
  stacks:
    - path: stack-a
    - path: stack-b
      depends_on:
        - stack-a

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

Neptune runs each step **once per changed stack**. For the full schema and options, see [Configuration](configuration.md).

## Step 4: Add the `repository_dispatch` workflow

Add the same **repository_dispatch** workflow as in [Getting started (Terramate) — Step 4](getting-started-terramate.md#step-4-add-the-repository_dispatch-workflow): create `.github/workflows/neptune.yml` that listens for `repository_dispatch` with type `neptune-command`, installs Neptune, checks out the PR ref, sets the required env vars (including object storage credentials), and runs `neptune command ${{ github.event.client_payload.command }}`. Use Terraform or OpenTofu setup as appropriate for your repo.

## Step 5: Branch protection (recommended)

To block merging until apply has run successfully, add **neptune apply** as a [required status check](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-a-branch-protection-rule) in your branch protection rules. See [Branch protection (recommended)](github-app-and-lambda.md#5-branch-protection-recommended).

## Step 6: End-to-end flow

Once the steps above are in place, the flow is:

1. **Open a PR** — Pushing a branch and opening a pull request triggers neptbot, which dispatches the workflow with `command: plan`. Neptune runs plan for each changed stack and posts a plan comment on the PR.
2. **Plan comment** — Review the plan output in the PR comment and the **neptune plan** commit status.
3. **Comment `@neptbot apply`** — When the plan looks good and the PR is approved, comment `@neptbot apply` on the PR. neptbot triggers the workflow with `command: apply`. Neptune runs apply for each changed stack and posts an apply comment.
4. **Apply comment** — After a successful apply, the **neptune apply** status is set to success. If you added **neptune apply** as a required check, the PR becomes mergeable.
5. **Auto-merge (optional)** — If you set `repository.automerge: true` in `.neptune.yaml`, Neptune enables PR auto-merge after a successful apply; the PR will merge when all required checks pass.

You can add screenshots here to illustrate each step (e.g. PR opened, plan comment, `@neptbot apply` comment, apply comment, and auto-merge).
