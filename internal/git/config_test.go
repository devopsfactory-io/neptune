package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseLsRemoteSymref(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{"main", "ref: refs/heads/main  HEAD\n", "main", false},
		{"master", "ref: refs/heads/master\tHEAD\n", "master", false},
		{"with newline", "ref: refs/heads/develop  HEAD\nabc", "develop", false},
		{"empty", "", "", true},
		{"bad format", "not a symref", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLsRemoteSymref(tt.out)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLsRemoteSymref() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseLsRemoteSymref() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShowFileFromRef(t *testing.T) {
	dir := t.TempDir()
	// Create a minimal git repo with one file on main
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@test")
	runGit(t, dir, "config", "user.name", "Test")
	cfg := filepath.Join(dir, ".neptune.yaml")
	if err := os.WriteFile(cfg, []byte("repository:\n  branch: main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".neptune.yaml")
	runGit(t, dir, "commit", "-m", "add config")
	runGit(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")

	content, err := ShowFileFromRef(dir, "HEAD", ".neptune.yaml")
	if err != nil {
		t.Fatalf("ShowFileFromRef: %v", err)
	}
	if string(content) != "repository:\n  branch: main\n" {
		t.Errorf("ShowFileFromRef content = %q", content)
	}

	// Missing file
	_, err = ShowFileFromRef(dir, "HEAD", "missing.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestShowFileFromRef_NoRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := ShowFileFromRef(dir, "HEAD", ".neptune.yaml")
	if err == nil {
		t.Error("expected error when not in a git repo")
	}
}

func TestDefaultBranch_Local(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@test")
	runGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "file.txt")
	runGit(t, dir, "commit", "-m", "initial")
	runGit(t, dir, "remote", "add", "origin", "https://github.com/owner/repo.git")
	runGit(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	branch, err := DefaultBranch(dir)
	if err != nil {
		t.Fatalf("DefaultBranch: %v", err)
	}
	if branch != "main" {
		t.Errorf("DefaultBranch = %q, want main", branch)
	}
}

func TestFetchBranch_NoRemote(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	err := FetchBranch(dir, "main")
	if err == nil {
		t.Error("expected error when remote has no URL or unreachable")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}
