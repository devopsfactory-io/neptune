package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"neptune/internal/domain"
	"neptune/internal/lock"
	"neptune/internal/log"
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
	log.For("run").Info("Executing workflow phase: " + r.Phase)
	if r.Locks != nil && len(r.Stacks) > 0 {
		if err := r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusInProgress); err != nil {
			log.For("run").Error("update stacks", "err", err)
		}
	}
	for _, step := range r.Steps {
		if step.Run == "" {
			continue
		}
		runOnceInRoot := step.Once != nil && *step.Once
		if runOnceInRoot {
			log.Banner("Neptune Runner", []string{"Neptune is running the following command: " + step.Run})
			log.For("run").Info("Running command: " + step.Run)
			runOut := r.runCommand(ctx, step.Run)
			out.Outputs = append(out.Outputs, runOut)
			log.For("run").Info("Command completed with return code " + fmt.Sprint(runOut.Status))
			if runOut.Status != 0 {
				out.OverallStatus = 1
				if r.Locks != nil && len(r.Stacks) > 0 {
					if err := r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusPending); err != nil {
						log.For("run").Error("update stacks", "err", err)
					}
				}
				return out, nil
			}
			continue
		}
		// Run step in each changed stack (once: false or unset).
		repoRoot, err := os.Getwd()
		if err != nil {
			log.For("run").Error("get working directory", "err", err)
			out.OverallStatus = 1
			return out, nil
		}
		if len(r.Stacks) == 0 {
			log.For("run").Info("No changed stacks, skipping step: " + step.Run)
			continue
		}
		for _, stack := range r.Stacks {
			stackDir := filepath.Join(repoRoot, stack)
			log.Banner("Neptune Runner", []string{"Stack " + stack + ": " + step.Run})
			log.For("run").Info("Running command in stack", "stack", stack, "command", step.Run)
			runOut := r.runCommandInDir(ctx, stackDir, step.Run)
			runOut.Command = step.Run
			runOut.Stack = stack
			if runOut.Output != "" || runOut.Error != "" {
				runOut.Output = "[" + stack + "] " + runOut.Output
				if runOut.Error != "" {
					runOut.Error = "[" + stack + "] " + runOut.Error
				}
			}
			out.Outputs = append(out.Outputs, runOut)
			log.For("run").Info("Command completed", "stack", stack, "status", runOut.Status)
			if runOut.Status != 0 {
				out.OverallStatus = 1
				if r.Locks != nil && len(r.Stacks) > 0 {
					if err := r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusPending); err != nil {
						log.For("run").Error("update stacks", "err", err)
					}
				}
				return out, nil
			}
		}
	}
	if r.Locks != nil && len(r.Stacks) > 0 {
		if err := r.Locks.UpdateStacks(ctx, r.Phase, r.Stacks, domain.WorkflowStatusCompleted); err != nil {
			log.For("run").Error("update stacks", "err", err)
		}
	}
	log.For("run").Info("Workflow phase completed")
	return out, nil
}

func (r *Runner) runCommand(ctx context.Context, command string) domain.RunOutput {
	return r.runCommandInDir(ctx, "", command)
}

func (r *Runner) runCommandInDir(ctx context.Context, dir, command string) domain.RunOutput {
	var stdout, stderr bytes.Buffer
	shell := "sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/c"
	}
	cmd := exec.CommandContext(ctx, shell, flag, command)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
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
