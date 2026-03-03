package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

func TestIsBranchRebased_NoRepo(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{Branch: "main", GitHub: &domain.GitHubConfig{}},
	}
	// Run from a temp dir with no git repo
	dir := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(orig) }()
	got := IsBranchRebased(cfg)
	if got {
		t.Error("expected false when not in a git repo")
	}
}

func TestIsBranchRebased_NilConfig(t *testing.T) {
	if IsBranchRebased(nil) {
		t.Error("expected false for nil config")
	}
}

func TestIsBranchRebased_EmptyBranch(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{Branch: "", GitHub: &domain.GitHubConfig{}},
	}
	dir := t.TempDir()
	_ = os.Chdir(dir)
	defer os.Chdir(".")
	// No git repo - should return false
	if IsBranchRebased(cfg) {
		t.Error("expected false")
	}
}

func TestIsBranchRebased_RealRepo(t *testing.T) {
	// Only run if we're inside the neptune repo (which has .git)
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()
	if _, err := os.Stat(filepath.Join(orig, ".git")); os.IsNotExist(err) {
		t.Skip("not in a git repo")
	}
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{Branch: "main", GitHub: &domain.GitHubConfig{}},
	}
	_ = IsBranchRebased(cfg)
	// Just ensure we don't panic; result depends on repo state
}
