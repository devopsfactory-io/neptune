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
	if !strings.Contains(body, "/neptune apply") {
		t.Error("missing apply hint")
	}
}
