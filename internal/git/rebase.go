package git

import (
	"os/exec"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

// IsBranchRebased returns true if the current HEAD has no commits behind origin/<defaultBranch>.
// Same semantics as Python: commits in default branch but not in current branch = 0.
func IsBranchRebased(cfg *domain.NeptuneConfig) bool {
	if cfg == nil || cfg.Repository == nil {
		return false
	}
	branch := cfg.Repository.Branch
	if branch == "" {
		branch = "master"
	}
	ref := "origin/" + branch
	// rev-list HEAD..origin/branch = commits reachable from origin/branch but not from HEAD (we're behind)
	cmd := exec.Command("git", "rev-list", "--count", "HEAD.."+ref) //nolint:gosec // G204: controlled ref from config
	cmd.Dir = "."
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	count := strings.TrimSpace(string(out))
	return count == "0"
}
