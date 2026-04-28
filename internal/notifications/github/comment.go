package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/log"
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
	log.For("notifications.github").Info("Initializing GitHub API")
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
	log.For("notifications.github").Info("Creating comment on PR " + n.prNum)
	log.For("notifications.github").Info("Formatting plan output for PR " + n.prNum)
	body := n.formatCommentBody(comment)
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", n.repo, n.prNum)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"body":`+escapeJSON(body)+`}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+n.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req) //nolint:gosec // G704: URL from controlled repo and PR number
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.For("notifications.github").Error("close response body", "err", err)
		}
	}()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create comment: status %d", resp.StatusCode)
	}
	log.For("notifications.github").Info("Comment created on PR " + n.prNum)
	return nil
}

// formatCommentBody builds the comment body from the comment; normalizes phase so "Plan" / "APPLY" use plan/apply format.
func (n *Notifier) formatCommentBody(comment *domain.PullRequestComment) string {
	if comment.StepsOutput == nil {
		comment.StepsOutput = &domain.StepsOutput{Phase: "custom", OverallStatus: comment.OverallStatus}
	}
	phase := strings.ToLower(strings.TrimSpace(comment.StepsOutput.Phase))
	switch phase {
	case "plan":
		return n.formatPlan(comment)
	case "apply":
		return n.formatApply(comment)
	default:
		return n.formatCustom(comment)
	}
}

func escapeJSON(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return s
	}
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
	if comment.SimpleOutput != "" {
		if comment.OverallStatus != 0 {
			b.WriteString("An error occurred:\n- ")
		} else {
			b.WriteString("**Neptune:** ")
		}
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
	b.WriteString("\nTo apply these changes, comment:\n```\n@neptbot apply\n```\n")
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
	if n.cfg.Repository != nil && n.cfg.Repository.Automerge && comment.OverallStatus == 0 {
		b.WriteString("\n\nAutomatically merging because all changed stacks have been successfully applied.\n")
	}
	return b.String()
}

func formatStep(b *strings.Builder, out domain.RunOutput) {
	limit := bodyLimit / 4
	cleanedErr := truncate(stripANSI(out.Error), limit)
	cleanedOut := truncate(stripANSI(out.Output), limit)
	b.WriteString("**Command ")
	if out.Status == 0 {
		b.WriteString("✅")
	} else {
		b.WriteString("❌")
	}
	b.WriteString("** `")
	b.WriteString(out.Command)
	b.WriteString("`")
	if out.Stack != "" {
		b.WriteString(" (stack: ")
		b.WriteString(out.Stack)
		b.WriteString(")")
	}
	b.WriteString("\n<details>\n<summary>Click to see the command output</summary>\n\n```\nstderr:\n")
	b.WriteString(cleanedErr)
	b.WriteString("\n\nstdout:\n")
	b.WriteString(cleanedOut)
	b.WriteString("\n```\n\n</details>\n\n")
}
