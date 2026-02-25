package run

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"

	"neptune/internal/domain"
	"neptune/internal/lock"
)

// Runner runs workflow phase steps and updates lock status.
type Runner struct {
	Config *domain.NeptuneConfig
	Phase  string
	Locks  *lock.Interface
	Stacks []string
	Steps  []domain.WorkflowStep
}

// Execute runs all steps; streams stdout/stderr to the process; updates lock to IN_PROGRESS then COMPLETED or PENDING.
func (r *Runner) Execute(ctx context.Context) (*domain.StepsOutput, error) {
	out := &domain.StepsOutput{Phase: r.Phase, OverallStatus: 0}
	if r.Locks != nil && len(r.Stacks) > 0 {
		_ = r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusInProgress)
	}
	for _, step := range r.Steps {
		if step.Run == "" {
			continue
		}
		runOut := r.runCommand(ctx, step.Run)
		out.Outputs = append(out.Outputs, runOut)
		if runOut.Status != 0 {
			out.OverallStatus = 1
			if r.Locks != nil && len(r.Stacks) > 0 {
				_ = r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusPending)
			}
			return out, nil
		}
	}
	if r.Locks != nil && len(r.Stacks) > 0 {
		_ = r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusCompleted)
	}
	return out, nil
}

func (r *Runner) runCommand(ctx context.Context, command string) domain.RunOutput {
	var stdout, stderr bytes.Buffer
	shell := "sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/c"
	}
	cmd := exec.CommandContext(ctx, shell, flag, command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	status := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status = exitErr.ExitCode()
		} else {
			status = 1
		}
	}
	return domain.RunOutput{
		Command: command,
		Output:  strings.TrimSpace(stdout.String()),
		Error:   strings.TrimSpace(stderr.String()),
		Status:  status,
	}
}
