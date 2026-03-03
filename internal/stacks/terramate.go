package stacks

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/terramate-io/terramate/config"
	"github.com/terramate-io/terramate/git"
	"github.com/terramate-io/terramate/run"
	"github.com/terramate-io/terramate/stack"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

// TerramateProvider returns changed stacks using the Terramate SDK.
type TerramateProvider struct{}

// ChangedStacks returns the list of changed stack paths in run order using the Terramate SDK.
// Repo root is the current working directory (or the directory of NEPTUNE_CONFIG_PATH if set).
// Base ref for change detection is origin/<cfg.Repository.Branch> (default origin/master).
func (p *TerramateProvider) ChangedStacks(ctx context.Context, cfg *domain.NeptuneConfig) (*domain.TerraformStacks, error) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	_ = ctx // Terramate SDK ListChanged does not accept context
	rootdir, err := GetRepoRoot()
	if err != nil {
		return nil, &TerramateError{Err: err, Stderr: ""}
	}

	root, configPath, found, err := config.TryLoadConfig(rootdir, true)
	if err != nil {
		return nil, &TerramateError{Err: fmt.Errorf("loading Terramate config: %w", err), Stderr: ""}
	}
	if !found || configPath == "" {
		return nil, &TerramateError{Err: fmt.Errorf("no Terramate config found in %s or parent directories", rootdir), Stderr: ""}
	}

	g, err := git.WithConfig(git.Config{
		WorkingDir: rootdir,
		Env:        os.Environ(),
	})
	if err != nil {
		return nil, &TerramateError{Err: fmt.Errorf("initializing git: %w", err), Stderr: ""}
	}

	mgr := stack.NewGitAwareManager(root, g)
	baseBranch := cfg.Repository.Branch
	if baseBranch == "" {
		baseBranch = "master"
	}
	baseRef := "origin/" + baseBranch

	report, err := mgr.ListChanged(stack.ChangeConfig{BaseRef: baseRef})
	if err != nil {
		return nil, &TerramateError{Err: fmt.Errorf("listing changed stacks: %w", err), Stderr: ""}
	}

	if len(report.Stacks) == 0 {
		return &domain.TerraformStacks{Stacks: nil}, nil
	}

	entries := make([]stack.Entry, len(report.Stacks))
	copy(entries, report.Stacks)

	_, err = run.Sort(root, entries, func(e stack.Entry) *config.Stack { return e.Stack })
	if err != nil {
		return nil, &TerramateError{Err: fmt.Errorf("computing run order: %w", err), Stderr: ""}
	}

	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		rel := e.Stack.Dir.String()
		rel = strings.TrimPrefix(rel, "/")
		if rel != "" {
			paths = append(paths, rel)
		}
	}

	return &domain.TerraformStacks{Stacks: paths}, nil
}

// TerramateError is returned when stack listing fails (Terramate SDK or config).
type TerramateError struct {
	Err    error
	Stderr string
}

func (e *TerramateError) Error() string {
	if e.Stderr != "" {
		return "failed to get changed Terraform stacks: " + e.Stderr
	}
	if e.Err != nil {
		return "failed to get changed Terraform stacks: " + e.Err.Error()
	}
	return "failed to get changed Terraform stacks"
}
