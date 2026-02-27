# Minimal root Terramate config for unit tests.
# required_version is required for Terramate to detect this as the root config.
terramate {
  required_version = ">= 0.1.0"
  config {
    disable_safeguards = ["git-untracked", "git-uncommitted"]
  }
}
