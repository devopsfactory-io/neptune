package lock

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"neptune/internal/domain"
)

// ChangedStacks runs `terramate list --changed --run-order` and returns the list of stack paths.
func ChangedStacks(ctx context.Context, _ *domain.NeptuneConfig) (*domain.TerraformStacks, error) {
	cmd := exec.CommandContext(ctx, "terramate", "list", "--changed", "--run-order")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, &TerramateError{Err: err, Stderr: stderr.String()}
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var stacks []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			stacks = append(stacks, line)
		}
	}
	return &domain.TerraformStacks{Stacks: stacks}, nil
}

// TerramateError is returned when terramate list fails.
type TerramateError struct {
	Err    error
	Stderr string
}

func (e *TerramateError) Error() string {
	if e.Stderr != "" {
		return "failed to get changed Terraform stacks: " + e.Stderr
	}
	return "failed to get changed Terraform stacks: " + e.Err.Error()
}
