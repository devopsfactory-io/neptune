package lock

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"neptune/internal/domain"
)

func TestChangedStacks_NoTerramateConfig(t *testing.T) {
	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{Branch: "main"},
	}
	stacks, err := ChangedStacks(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error when no Terramate config present")
	}
	if stacks != nil {
		t.Fatalf("expected nil stacks, got %v", stacks)
	}
	if err != nil && !strings.Contains(err.Error(), "failed to get changed Terraform stacks") {
		t.Errorf("expected stacks/Terramate error message, got: %v", err)
	}
}

func TestChangedStacks_ReturnsChangedStacksInRunOrder(t *testing.T) {
	// Create temp dir and copy testdata Terramate project.
	_, filename, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(filename), "testdata")
	dest := t.TempDir()

	err := filepath.Walk(src, func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			return errWalk
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0644)
	})
	if err != nil {
		t.Fatalf("copy testdata: %v", err)
	}

	// Initialize git repo: main with all stacks, then pr-1 with change in stack-a.
	runGit(t, dest, "init", "-b", "main")
	runGit(t, dest, "config", "user.email", "test@neptune.test")
	runGit(t, dest, "config", "user.name", "Neptune Test")
	runGit(t, dest, "add", ".")
	runGit(t, dest, "commit", "-m", "main: all stacks")
	// Create origin/main so ListChanged(BaseRef: "origin/main") can compare.
	runGit(t, dest, "update-ref", "refs/remotes/origin/main", "HEAD")

	runGit(t, dest, "checkout", "-b", "pr-1")
	// Change stack-a so it is detected as changed.
	stackAMain := filepath.Join(dest, "stack-a", "stack.tm.hcl")
	content, err := os.ReadFile(stackAMain)
	if err != nil {
		t.Fatalf("read stack-a config: %v", err)
	}
	if err := os.WriteFile(stackAMain, append(content, '\n', '#', ' ', 'c', 'h', 'a', 'n', 'g', 'e', '\n'), 0644); err != nil {
		t.Fatalf("write stack-a change: %v", err)
	}
	runGit(t, dest, "add", "stack-a/stack.tm.hcl")
	runGit(t, dest, "commit", "-m", "pr-1: change stack-a")

	// Run ChangedStacks from the repo dir (base ref origin/main).
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dest); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{Branch: "main"},
	}
	stacks, err := ChangedStacks(context.Background(), cfg)
	if err != nil {
		t.Fatalf("ChangedStacks: %v", err)
	}
	if stacks == nil {
		t.Fatal("expected non-nil TerraformStacks")
	}
	if len(stacks.Stacks) == 0 {
		t.Fatal("expected at least one changed stack (stack-a)")
	}
	found := false
	for _, p := range stacks.Stacks {
		if p == "stack-a" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected stack-a in changed stacks, got %v", stacks.Stacks)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
