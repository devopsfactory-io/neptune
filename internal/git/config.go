package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// DefaultBranch returns the repository's default branch name using git CLI only.
// workDir is the repository working directory (e.g. from lock's repoRoot).
// It tries local origin/HEAD first, then git ls-remote --symref origin HEAD if needed.
func DefaultBranch(workDir string) (string, error) {
	// 1. Try local: git rev-parse --abbrev-ref origin/HEAD -> e.g. "origin/main"
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			if branch := strings.TrimPrefix(name, "origin/"); branch != name {
				return branch, nil
			}
		}
	}

	// 2. Try: git symbolic-ref refs/remotes/origin/HEAD -> refs/remotes/origin/main
	cmd2 := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd2.Dir = workDir
	out2, err2 := cmd2.Output()
	if err2 == nil {
		name := strings.TrimSpace(string(out2))
		if name != "" {
			// refs/remotes/origin/main -> main
			if strings.HasPrefix(name, "refs/remotes/origin/") {
				return name[len("refs/remotes/origin/"):], nil
			}
		}
	}

	// 3. Query remote: git ls-remote --symref origin HEAD
	cmd3 := exec.Command("git", "ls-remote", "--symref", "origin", "HEAD")
	cmd3.Dir = workDir
	out3, err3 := cmd3.Output()
	if err3 != nil {
		return "", fmt.Errorf("getting default branch: %w", err3)
	}
	// Parse "ref: refs/heads/main  HEAD" or similar
	branch, err := parseLsRemoteSymref(string(out3))
	if err != nil {
		return "", fmt.Errorf("parsing ls-remote output: %w", err)
	}
	return branch, nil
}

// parseLsRemoteSymref parses the output of "git ls-remote --symref origin HEAD".
// Expected format: "ref: refs/heads/<branch>  HEAD\n" or "ref: refs/heads/main\tHEAD\n".
var lsRemoteSymrefRe = regexp.MustCompile(`ref:\s+refs/heads/(\S+)\s+HEAD`)

func parseLsRemoteSymref(out string) (string, error) {
	line := strings.TrimSpace(strings.Split(out, "\n")[0])
	if line == "" {
		return "", fmt.Errorf("empty ls-remote output")
	}
	matches := lsRemoteSymrefRe.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", fmt.Errorf("unexpected ls-remote format: %q", line)
	}
	return matches[1], nil
}

// ShowFileFromRef runs "git show <ref>:<path>" in workDir and returns the file content.
// path is relative to workDir (e.g. ".neptune.yaml").
func ShowFileFromRef(workDir, ref, path string) ([]byte, error) {
	cmd := exec.Command("git", "show", ref+":"+path)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", ref, path, err)
	}
	return out, nil
}

// FetchBranch runs "git fetch origin <branch>" in workDir so that origin/<branch> exists.
func FetchBranch(workDir, branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch origin %s: %w: %s", branch, err, string(out))
	}
	return nil
}
