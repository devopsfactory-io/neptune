package run

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"neptune/internal/domain"
)

func TestRunner_Execute_SimpleStep(t *testing.T) {
	cmd := "echo hello"
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	}
	onceTrue := true
	cfg := &domain.NeptuneConfig{}
	runner := &Runner{
		Config: cfg,
		Phase:  "plan",
		Locks:  nil,
		Stacks: nil,
		Steps:  []domain.WorkflowStep{{Run: cmd, Once: &onceTrue}},
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
	onceTrue := true
	runner := &Runner{
		Config: &domain.NeptuneConfig{},
		Phase:  "plan",
		Steps:  []domain.WorkflowStep{{Run: cmd, Once: &onceTrue}},
	}
	out, err := runner.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.OverallStatus != 1 {
		t.Errorf("expected overall status 1, got %d", out.OverallStatus)
	}
}

func TestRunner_Execute_OnceUnset_PerStack(t *testing.T) {
	dir := t.TempDir()
	stackDir := filepath.Join(dir, "stack-a")
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Run a step with once unset (default per-stack); command writes a file so we can assert it ran in the stack dir.
	writeCmd := "echo done > neptune_test_out.txt"
	if runtime.GOOS == "windows" {
		writeCmd = "echo done > neptune_test_out.txt"
	}
	runner := &Runner{
		Config: &domain.NeptuneConfig{},
		Phase:  "plan",
		Stacks: []string{"stack-a"},
		Steps:  []domain.WorkflowStep{{Run: writeCmd}},
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	out, err := runner.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.OverallStatus != 0 {
		t.Errorf("overall status %d", out.OverallStatus)
	}
	if len(out.Outputs) != 1 {
		t.Fatalf("expected 1 output (one per stack), got %d", len(out.Outputs))
	}
	// File should exist in stack-a dir.
	testFile := filepath.Join(stackDir, "neptune_test_out.txt")
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("expected command to run in stack dir, file not found: %v", err)
	}
}
