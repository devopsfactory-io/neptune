package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"neptune/internal/config"
	"neptune/internal/domain"
	"neptune/internal/github"
	"neptune/internal/lock"
	githubnotify "neptune/internal/notifications/github"
	"neptune/internal/run"
)

// NewCommandCmd returns the command (workflow) subcommand.
func NewCommandCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "command [workflow]",
		Short: "Run a workflow phase (e.g. plan, apply)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCommand,
	}
	return c
}

func runCommand(_ *cobra.Command, args []string) error {
	workflow := args[0]
	ctx := context.Background()

	env, err := config.LoadEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	cfg, err := config.Load(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		notifyAndExit(cfg, err.Error(), 1)
	}

	wfs := cfg.Workflows.Workflows[cfg.Repository.AllowedWorkflow]
	if _, ok := wfs.Phases[workflow]; !ok {
		notifyAndExit(cfg, "Workflow "+workflow+" is not valid, check the allowed workflow in the Neptune config", 1)
	}

	var requirements []string
	switch workflow {
	case "plan":
		requirements = cfg.Repository.PlanRequirements
	case "apply":
		requirements = cfg.Repository.ApplyRequirements
	}

	ghClient := github.NewClient(cfg)
	if ghClient == nil {
		notifyAndExit(cfg, "GITHUB_TOKEN, GITHUB_REPOSITORY, and GITHUB_PULL_REQUEST_NUMBER are required", 1)
	}
	status := ghClient.CheckRequirements(ctx, requirements)
	if !status.IsCompliant {
		notifyAndExit(cfg, "Cannot run "+workflow+" workflow: "+status.ErrorMessage, 1)
	}

	isPROpen := func(ctx context.Context, prNumber string) (bool, error) {
		return ghClient.IsPROpen(ctx, prNumber)
	}
	lockIface, err := lock.NewInterface(ctx, cfg, isPROpen)
	if err != nil {
		notifyAndExit(cfg, "Failed to get changed Terraform stacks: "+err.Error(), 1)
	}
	defer lockIface.Close()

	if len(lockIface.TerraformStacks.Stacks) == 0 {
		notifyAndExit(cfg, "No Terraform stacks found", 0)
	}

	locked, err := lockIface.StacksLocked(ctx)
	if err != nil {
		notifyAndExit(cfg, "Failed to check lock status: "+err.Error(), 1)
	}
	if locked.Locked {
		msg := fmt.Sprintf("Some stacks (%v) are locked by other PRs: %v", locked.StackPath, locked.PRs)
		notifyAndExit(cfg, msg, 1)
	}

	if !lockIface.DependsOnCompleted(workflow) {
		notifyAndExit(cfg, "Dependency is not met for phase "+workflow, 1)
	}

	if err := lockIface.LockStacks(ctx, workflow, lockIface.TerraformStacks.Stacks, domain.WorkflowStatusPending); err != nil {
		notifyAndExit(cfg, "Failed to lock stacks: "+err.Error(), 1)
	}

	phase := wfs.Phases[workflow]
	runner := &run.Runner{
		Config: cfg,
		Phase:  workflow,
		Locks:  lockIface,
		Stacks: lockIface.TerraformStacks.Stacks,
		Steps:  phase.Steps,
	}
	stepsOut, err := runner.Execute(ctx)
	if err != nil {
		notifyAndExit(cfg, "Failed to run steps: "+err.Error(), 1)
	}

	comment := &domain.PullRequestComment{
		StepsOutput:   stepsOut,
		Stacks:        lockIface.TerraformStacks,
		OverallStatus: stepsOut.OverallStatus,
	}
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		_ = notifier.CreateComment(comment)
	}

	msg := fmt.Sprintf("Workflow %s completed with status %d", workflow, stepsOut.OverallStatus)
	if stepsOut.OverallStatus != 0 {
		fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(stepsOut.OverallStatus)
	}
	fmt.Println("Success:", msg)
	return nil
}

func notifyAndExit(cfg *domain.NeptuneConfig, output string, status int) {
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		_ = notifier.CreateComment(&domain.PullRequestComment{
			SimpleOutput:  output,
			OverallStatus: status,
		})
	}
	if status != 0 {
		fmt.Fprintln(os.Stderr, "Error:", output)
		os.Exit(status)
	}
	fmt.Println("Success:", output)
	os.Exit(0)
}
