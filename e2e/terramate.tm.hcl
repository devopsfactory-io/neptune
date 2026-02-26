# Root Terramate config for e2e tests.
# Disable run safeguards so terraform init-generated .terraform/ and tfplan do not fail the run.
terramate {
  config {
    disable_safeguards = ["git-untracked", "git-uncommitted"]
  }
}
