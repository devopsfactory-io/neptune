package run

import (
	"context"
	"runtime"
	"testing"

	"neptune/internal/domain"
)

func TestRunner_Execute_SimpleStep(t *testing.T) {
	cmd := "echo hello"
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	}
	cfg := &domain.NeptuneConfig{}
	runner := &Runner{
		Config: cfg,
		Phase:  "plan",
		Locks:  nil,
		Stacks: nil,
		Steps:  []domain.WorkflowStep{{Run: cmd}},
	}
	out, err := runner.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.OverallStatus != 0 {
		t.Errorf("overall status %d", out.OverallStatus)
	}
	if len(out.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(out.Outputs))
	}
	if out.Outputs[0].Status != 0 {
		t.Errorf("step status %d", out.Outputs[0].Status)
	}
}

func TestRunner_Execute_FailingStep(t *testing.T) {
	cmd := "exit 2"
	if runtime.GOOS == "windows" {
		cmd = "exit /b 2"
	}
	runner := &Runner{
		Config: &domain.NeptuneConfig{},
		Phase:  "plan",
		Steps:  []domain.WorkflowStep{{Run: cmd}},
	}
	out, err := runner.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.OverallStatus != 1 {
		t.Errorf("expected overall status 1, got %d", out.OverallStatus)
	}
}
