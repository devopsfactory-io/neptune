# Neptune GitHub App webhook Lambda (self-hosted)

This directory is for **self-hosting** the webhook handler. The default way to use Neptune with webhooks is to install the Neptune project's **neptbot** GitHub App; use this Lambda code and CloudFormation only if you want to run your own GitHub App and Lambda (e.g. in your AWS account).

The Lambda receives webhooks (pull request and issue comment events), verifies the signature, and triggers a GitHub Actions workflow in the target repository via `repository_dispatch` so that Neptune runs `plan` (on PR open/sync) or `apply` (when someone comments e.g. `@neptune apply`). It also adds a 👀 (eyes) reaction to the PR and to the comment that triggered the command for visibility. This repo provides the Lambda code and a CloudFormation template to deploy it.

## Prerequisites (if self-hosting)

If you are self-hosting, you need:

- Go 1.22+ (to build the binary)
- AWS CLI configured (to deploy and to create secrets)
- A [GitHub App](https://docs.github.com/en/apps/creating-github-apps) with:
  - Webhook URL set to your Lambda Function URL (after deploy)
  - Webhook secret (stored in AWS Secrets Manager)
  - Permissions: Repository permissions → **Contents** (read and write), **Pull requests** (read and write), **Issues** (read and write), **Metadata** (read). The `repository_dispatch` API requires write access to the repository. **Issues** (read and write) is required for the Lambda to add a 👀 reaction to the PR and to the comment; **Pull requests** (read and write) is also recommended for reactions on pull requests—see [GitHub App permissions](https://docs.github.com/en/rest/authentication/permissions-required-for-github-apps#repository-permissions-for-pull-requests).
  - Subscribe to events: **Pull requests**, **Issue comments**
  - Private key (stored in AWS Secrets Manager)
- Optionally set **NEPTUNE_PR_LABEL** (e.g. `neptune`) so that only PRs with that label trigger the workflow and get the eyes reaction; add that label to infrastructure-related PRs.

## Build

From this directory (`lambda/`):

```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap .
zip neptune-webhook.zip bootstrap
```

From the repository root you can run `make lambda.build` and `make lambda.zip`. The binary must be named `bootstrap` for the `provided.al2023` runtime. Run `make lambda.test` to execute the Lambda unit tests.

## Deploy with CloudFormation

1. **Create two secrets in AWS Secrets Manager** (in the same region/account where you will deploy the stack):

   - **Webhook secret**: Create a secret (e.g. `neptune-github-app/webhook-secret`) with a **plain text** value: the webhook secret from your GitHub App settings.
   - **Private key**: Create a secret (e.g. `neptune-github-app/private-key`) with a **plain text** value: the full PEM content of your GitHub App private key (including `-----BEGIN RSA PRIVATE KEY-----` and `-----END RSA PRIVATE KEY-----`).

2. **Upload the Lambda zip to S3**:

   ```bash
   aws s3 cp neptune-webhook.zip s3://YOUR_BUCKET/neptune-webhook.zip
   ```

3. **Deploy the stack**:

   ```bash
   aws cloudformation deploy \
     --template-file cloudformation/template.yaml \
     --stack-name neptune-webhook \
     --parameter-overrides \
       GitHubAppId=YOUR_APP_ID \
       LambdaS3Bucket=YOUR_BUCKET \
       LambdaS3Key=neptune-webhook.zip \
       WebhookSecretArn=arn:aws:secretsmanager:REGION:ACCOUNT:secret:neptune-github-app/webhook-secret \
       PrivateKeySecretArn=arn:aws:secretsmanager:REGION:ACCOUNT:secret:neptune-github-app/private-key \
       GitHubAppSlug=neptbot \
     --capabilities CAPABILITY_NAMED_IAM
   ```

   Replace `YOUR_APP_ID`, `YOUR_BUCKET`, the secret ARNs, and optionally `GitHubAppSlug` (the app’s login/slug for @-mention matching; omit or leave empty to use the default `neptbot`).

4. **Get the webhook URL**:

   ```bash
   aws cloudformation describe-stacks --stack-name neptune-webhook \
     --query "Stacks[0].Outputs[?OutputKey=='WebhookUrl'].OutputValue" --output text
   ```

   Set this URL as the **Payload URL** in your GitHub App’s webhook settings.

## Verify deployment

After deploying, confirm the Lambda is reachable and the handler runs:

1. **Smoke tests** (replace `YOUR_FUNCTION_URL` with the URL from the stack output):

   - **GET** (expect **405** Method Not Allowed):
     ```bash
     curl -s -o /dev/null -w "%{http_code}" YOUR_FUNCTION_URL
     ```
   - **POST without signature** (expect **401** Invalid signature):
     ```bash
     curl -s -o /dev/null -w "%{http_code}" -X POST YOUR_FUNCTION_URL -d '{}'
     ```

   If you see **403 Forbidden**, either (1) the Function URL is using IAM auth — set **Auth type** to **NONE** in the Lambda console under **Configuration** → **Function URL** — or (2) the console reports that Auth is NONE but "permissions for public access" are missing. For (2), redeploy the stack so the template’s resource-based permission (`lambda:InvokeFunctionUrl` for principal `*`) is applied, or in the Lambda console go to **Configuration** → **Permissions** and add a resource-based policy that allows public invocation of the Function URL (e.g. use the console’s option to add permissions for **Function URL** when Auth type is NONE).

2. **Full verification**: Set the GitHub App **Payload URL** to this Lambda URL, then open **Settings → Developer settings → GitHub Apps → your app → Advanced → Recent Deliveries**. Use **Redeliver** on the initial **ping**. A **200** response means the Lambda accepted GitHub’s signed payload and is working end-to-end.

3. **Logs**: If you get **500** or unexpected behavior, check CloudWatch logs for the function (Lambda → Monitor → View CloudWatch logs) for messages like `load config: ...`, `verify signature: ...`, or `repository_dispatch: ...`. A 500 with body "Dispatch error" often means GitHub returned 403 (e.g. "Resource not accessible by integration"); ensure the App has **Contents: Read and write** so `repository_dispatch` is allowed.

4. **Eyes reaction not appearing on PR or comment**: The Lambda adds a 👀 reaction after a successful dispatch. If dispatch works but no reaction appears, check CloudWatch for `eyes reaction on PR` or `eyes reaction on comment` — a failure there (e.g. status 403) usually means the App lacks the right permissions. In GitHub, go to the App → **Permissions and events** → set **Issues** to **Read and write** and **Pull requests** to **Read and write** (reactions on PRs can require both; see [GitHub App permissions](https://docs.github.com/en/rest/authentication/permissions-required-for-github-apps#repository-permissions-for-pull-requests)), save, and have the installation accept the new permissions if prompted. You can verify reactions with the [check-reactions script](scripts/check-reactions.sh): `./lambda/scripts/check-reactions.sh PR_NUMBER` (requires `gh` and `jq`).

## Environment variables (Lambda)

The CloudFormation template sets these from parameters and secret ARNs:

| Variable | Source | Description |
|----------|--------|-------------|
| `GITHUB_APP_ID` | Parameter | GitHub App ID (numeric). |
| `GITHUB_APP_WEBHOOK_SECRET_ARN` | Parameter | ARN of the Secrets Manager secret with the webhook secret (plain string). |
| `GITHUB_APP_PRIVATE_KEY_SECRET_ARN` | Parameter | ARN of the Secrets Manager secret with the App private key (PEM). |
| `GITHUB_APP_SLUG` | Parameter (optional) | App slug for @-mention matching in comments (e.g. `neptbot`). Default in code: `neptbot`. |
| `NEPTUNE_PR_LABEL` | Parameter (optional) | When set, only PRs with this label trigger dispatch and eyes (e.g. `neptune`). Leave empty to trigger on all PRs. |

At runtime the Lambda fetches the webhook secret and private key from Secrets Manager using these ARNs.

## User repositories

Repositories that have the Neptune GitHub App installed must add a workflow that runs on `repository_dispatch` with event type `neptune-command`, and run `neptune command plan` or `neptune command apply` using the payload. See [GitHub App and Lambda](../docs/github-app-and-lambda.md) in the main docs for the workflow example and required environment variables.

## Events handled

- **pull_request** (`opened`, `reopened`, `synchronize`, `ready_for_review`): triggers `repository_dispatch` with `command: plan`, and adds a 👀 reaction to the PR.
- **issue_comment** (created, on a PR): if the comment body mentions the app (e.g. `@neptbot`) and contains the word `apply` or `plan`, triggers `repository_dispatch` with that command, and adds a 👀 reaction to the comment.

If **NEPTUNE_PR_LABEL** is set, only PRs that have that label (e.g. `neptune`) trigger the workflow and get the eyes reaction; other PRs receive a silent `200 OK` with no dispatch and no reaction.
