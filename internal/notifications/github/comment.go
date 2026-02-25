package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"neptune/internal/domain"
)

const (
	bodyLimit = 65536 - 2048
)

var ansiRE = regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)

// Notifier posts PR comments to GitHub.
type Notifier struct {
	cfg    *domain.NeptuneConfig
	repo   string
	prNum  string
	token  string
	runID  string
	client *http.Client
}

// NewNotifier creates a notifier from config.
func NewNotifier(cfg *domain.NeptuneConfig) *Notifier {
	if cfg == nil || cfg.Repository == nil || cfg.Repository.GitHub == nil {
		return nil
	}
	repo := cfg.Repository.GitHub.Repository
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/")
	return &Notifier{
		cfg:    cfg,
		repo:   repo,
		prNum:  cfg.Repository.GitHub.PullRequestNumber,
		token:  cfg.Repository.GitHub.Token,
		runID:  cfg.Repository.GitHub.RunID,
		client: &http.Client{},
	}
}

// CreateComment posts the comment to the PR.
func (n *Notifier) CreateComment(comment *domain.PullRequestComment) error {
	if n == nil || n.token == "" || n.repo == "" || n.prNum == "" {
		return nil
	}
	if comment == nil {
		return nil
	}
	if comment.StepsOutput == nil {
		comment.StepsOutput = &domain.StepsOutput{Phase: "custom", OverallStatus: comment.OverallStatus}
	}
	var body string
	switch comment.StepsOutput.Phase {
	case "plan":
		body = n.formatPlan(comment)
	case "apply":
		body = n.formatApply(comment)
	default:
		body = n.formatCustom(comment)
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", n.repo, n.prNum)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"body":`+escapeJSON(body)+`}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+n.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create comment: status %d", resp.StatusCode)
	}
	return nil
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func truncate(s string, limit int) string {
	const suffix = "...[truncated]"
	if len(s) > limit {
		return s[:limit-len(suffix)] + suffix
	}
	return s
}

func (n *Notifier) formatCustom(comment *domain.PullRequestComment) string {
	var b strings.Builder
	b.WriteString("### 🌊 Neptune Execution Results\n\n")
	if comment.OverallStatus != 0 && comment.SimpleOutput != "" {
		b.WriteString("An error occurred:\n- ")
		b.WriteString(comment.SimpleOutput)
		b.WriteString("\n\n")
	}
	if comment.Stacks != nil && len(comment.Stacks.Stacks) > 0 {
		b.WriteString("**Terraform Stacks:** `")
		b.WriteString(strings.Join(comment.Stacks.Stacks, ", "))
		b.WriteString("`\n\n")
	}
	if comment.StepsOutput != nil {
		b.WriteString("**Neptune completed the ")
		b.WriteString(comment.StepsOutput.Phase)
		b.WriteString(" workflow with status:** ")
		if comment.OverallStatus == 0 {
			b.WriteString("✅")
		} else {
			b.WriteString("❌")
		}
		b.WriteString("\n> For more details, see the [GitHub Actions run](https://github.com/")
		b.WriteString(n.repo)
		b.WriteString("/actions/runs/")
		b.WriteString(n.runID)
		b.WriteString(")\n\n")
		for _, out := range comment.StepsOutput.Outputs {
			formatStep(&b, out)
		}
	}
	return b.String()
}

func (n *Notifier) formatPlan(comment *domain.PullRequestComment) string {
	var b strings.Builder
	b.WriteString("### 🌊 Neptune Plan Results\n\n")
	if comment.Stacks != nil && len(comment.Stacks.Stacks) > 0 {
		b.WriteString("**Terraform Stacks:** `")
		b.WriteString(strings.Join(comment.Stacks.Stacks, ", "))
		b.WriteString("`\n\n")
	}
	b.WriteString("**Neptune completed the plan with status:** ")
	if comment.OverallStatus == 0 {
		b.WriteString("✅")
	} else {
		b.WriteString("❌")
	}
	b.WriteString("\n> For more details, see the [GitHub Actions run](https://github.com/")
	b.WriteString(n.repo)
	b.WriteString("/actions/runs/")
	b.WriteString(n.runID)
	b.WriteString(")\n\n")
	if comment.StepsOutput != nil {
		for _, out := range comment.StepsOutput.Outputs {
			formatStep(&b, out)
		}
	}
	b.WriteString("\nTo apply these changes, comment:\n```\n/neptune apply\n```\n")
	return b.String()
}

func (n *Notifier) formatApply(comment *domain.PullRequestComment) string {
	var b strings.Builder
	b.WriteString("### 🌊 Neptune Apply Results\n\n")
	if comment.Stacks != nil && len(comment.Stacks.Stacks) > 0 {
		b.WriteString("**Terraform Stacks:** `")
		b.WriteString(strings.Join(comment.Stacks.Stacks, ", "))
		b.WriteString("`\n\n")
	}
	b.WriteString("**Neptune completed the apply with status:** ")
	if comment.OverallStatus == 0 {
		b.WriteString("✅")
	} else {
		b.WriteString("❌")
	}
	b.WriteString("\n> For more details, see the [GitHub Actions run](https://github.com/")
	b.WriteString(n.repo)
	b.WriteString("/actions/runs/")
	b.WriteString(n.runID)
	b.WriteString(")\n\n")
	if comment.StepsOutput != nil {
		for _, out := range comment.StepsOutput.Outputs {
			formatStep(&b, out)
		}
	}
	return b.String()
}

func formatStep(b *strings.Builder, out domain.RunOutput) {
	limit := bodyLimit / 4
	cleanedErr := truncate(stripANSI(out.Error), limit)
	cleanedOut := truncate(stripANSI(out.Output), limit)
	b.WriteString("- **Command ")
	if out.Status == 0 {
		b.WriteString("✅")
	} else {
		b.WriteString("❌")
	}
	b.WriteString("** `")
	b.WriteString(out.Command)
	b.WriteString("`\n<details>\n<summary>Click to see the command output</summary>\n\n```\nstderr:\n")
	b.WriteString(cleanedErr)
	b.WriteString("\n\nstdout:\n")
	b.WriteString(cleanedOut)
	b.WriteString("\n```\n\n</details>\n\n")
}
