# Root Terramate config for automerge example.
terramate {
  required_version = ">= 0.1.0"
  config {
    disable_safeguards = ["git-untracked", "git-uncommitted"]
  }
}
