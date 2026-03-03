package github

import (
	"strings"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

func TestStripANSI(t *testing.T) {
	in := "\x1B[31mred\x1B[0m normal"
	got := stripANSI(in)
	if got != "red normal" {
		t.Errorf("got %q", got)
	}
}

func TestTruncate(t *testing.T) {
	s := strings.Repeat("a", 100)
	got := truncate(s, 20)
	if len(got) > 20+len("...[truncated]") {
		t.Errorf("truncate too long: %d", len(got))
	}
	if !strings.HasSuffix(got, "...[truncated]") {
		t.Errorf("expected suffix ...[truncated], got %q", got)
	}
}

func TestFormatPlan(t *testing.T) {
	n := &Notifier{repo: "owner/repo", runID: "123"}
	c := &domain.PullRequestComment{
		Stacks:        &domain.TerraformStacks{Stacks: []string{"stack1"}},
		OverallStatus: 0,
		StepsOutput: &domain.StepsOutput{
			Phase: "plan",
			Outputs: []domain.RunOutput{
				{Command: "terramate run --changed -- plan", Status: 0, Output: "ok", Error: ""},
			},
		},
	}
	body := n.formatPlan(c)
	if !strings.Contains(body, "Neptune Plan Results") {
		t.Error("missing plan header")
	}
	if !strings.Contains(body, "stack1") {
		t.Error("missing stacks")
	}
	if !strings.Contains(body, "@neptbot apply") {
		t.Error("missing apply hint")
	}
}

func TestFormatCommentBody_PhaseNormalization(t *testing.T) {
	n := &Notifier{repo: "owner/repo", runID: "123"}
	c := &domain.PullRequestComment{
		Stacks:        &domain.TerraformStacks{Stacks: []string{"stack-a"}},
		OverallStatus: 0,
		StepsOutput: &domain.StepsOutput{
			Phase:   "Plan",
			Outputs: []domain.RunOutput{{Command: "terraform plan", Status: 0}},
		},
	}
	body := n.formatCommentBody(c)
	if !strings.Contains(body, "Neptune Plan Results") {
		t.Errorf("Phase \"Plan\" should use plan format; got body (excerpt): %s", body[:min(200, len(body))])
	}
	if !strings.Contains(body, "stack-a") {
		t.Error("missing stacks in plan body")
	}
}

func TestFormatApply_AutomergeMessage(t *testing.T) {
	automergeMsg := "Automatically merging because all changed stacks have been successfully applied."

	// automerge true, success -> message present
	n := &Notifier{
		repo:  "owner/repo",
		runID: "123",
		cfg: &domain.NeptuneConfig{
			Repository: &domain.RepositoryConfig{Automerge: true},
		},
	}
	c := &domain.PullRequestComment{
		Stacks:        &domain.TerraformStacks{Stacks: []string{"stack1"}},
		OverallStatus: 0,
		StepsOutput:   &domain.StepsOutput{Phase: "apply", Outputs: []domain.RunOutput{{Command: "apply", Status: 0}}},
	}
	body := n.formatApply(c)
	if !strings.Contains(body, automergeMsg) {
		t.Errorf("formatApply with automerge true and success should contain automerge message; got: %s", body)
	}

	// automerge false, success -> message absent
	n.cfg.Repository.Automerge = false
	body = n.formatApply(c)
	if strings.Contains(body, automergeMsg) {
		t.Error("formatApply with automerge false should not contain automerge message")
	}

	// automerge true, failure -> message absent
	n.cfg.Repository.Automerge = true
	c.OverallStatus = 1
	body = n.formatApply(c)
	if strings.Contains(body, automergeMsg) {
		t.Error("formatApply with apply failure should not contain automerge message")
	}
}
