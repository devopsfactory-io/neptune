package github

import (
	"strings"
	"testing"

	"neptune/internal/domain"
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
	if !strings.Contains(body, "@neptune-bot apply") {
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
