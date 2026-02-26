package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"neptune/internal/config"
	"neptune/internal/domain"
	"neptune/internal/github"
	"neptune/internal/lock"
	"neptune/internal/log"
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
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	cfg, err := config.Load(env)
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		notifyAndExit(cfg, err.Error(), 1)
	}
	log.Init(cfg.LogLevel)

	// Config summary banner (after validate, like Python)
	configSummaryLines := configSummaryLines(cfg)
	log.Banner("Neptune Config Summary", configSummaryLines)

	log.Banner("Neptune Command", []string{"Neptune is running: " + workflow})

	wfs := cfg.Workflows.Workflows[cfg.Repository.AllowedWorkflow]
	if _, ok := wfs.Phases[workflow]; !ok {
		notifyAndExit(cfg, "Workflow "+workflow+" is not valid, check the allowed workflow in the Neptune config", 1)
	}
	log.For("cli").Info("Workflow " + workflow + " is valid")

	var requirements []string
	switch workflow {
	case "plan":
		requirements = cfg.Repository.PlanRequirements
	case "apply":
		requirements = cfg.Repository.ApplyRequirements
	}

	e2eMode := os.Getenv("NEPTUNE_E2E") == "1"
	var isPROpen func(context.Context, string) (bool, error)
	if e2eMode {
		isPROpen = func(_ context.Context, _ string) (bool, error) { return true, nil }
	} else {
		ghClient := github.NewClient(cfg)
		if ghClient == nil {
			notifyAndExit(cfg, "GITHUB_TOKEN, GITHUB_REPOSITORY, and GITHUB_PULL_REQUEST_NUMBER are required", 1)
		}
		status := ghClient.CheckRequirements(ctx, requirements)
		if !status.IsCompliant {
			notifyAndExit(cfg, "Cannot run "+workflow+" workflow: "+status.ErrorMessage, 1)
		}
		log.For("cli").Info("PR requirements check passed for workflow " + workflow)
		reqLine := "PR requirements (" + strings.Join(requirements, ", ") + ") check passed for workflow: " + workflow
		log.Banner("Neptune Plan/Apply Requirements Check", []string{reqLine})
		isPROpen = func(ctx context.Context, prNumber string) (bool, error) {
			return ghClient.IsPROpen(ctx, prNumber)
		}
	}
	lockIface, err := lock.NewInterface(ctx, cfg, isPROpen)
	if err != nil {
		notifyAndExit(cfg, "Failed to get changed Terraform stacks: "+err.Error(), 1)
	}
	defer func() {
		if err := lockIface.Close(); err != nil {
			log.For("cli").Error("close lock", "err", err)
		}
	}()

	if len(lockIface.TerraformStacks.Stacks) == 0 {
		notifyAndExit(cfg, "No Terraform stacks found", 0)
	}

	lockBannerLines := []string{"Neptune is considering the following stacks in the current PR:"}
	lockBannerLines = append(lockBannerLines, strings.Join(lockIface.TerraformStacks.Stacks, ", "))
	log.Banner("Neptune Lock", lockBannerLines)

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

	stepsSummaryLines := stepsSummaryLines(workflow, stepsOut)
	log.Banner("Neptune Steps Summary", stepsSummaryLines)

	comment := &domain.PullRequestComment{
		StepsOutput:   stepsOut,
		Stacks:        lockIface.TerraformStacks,
		OverallStatus: stepsOut.OverallStatus,
	}
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		if err := notifier.CreateComment(comment); err != nil {
			log.For("cli").Error("failed to post comment", "err", err)
		}
	}

	msg := fmt.Sprintf("Workflow %s completed with status %d", workflow, stepsOut.OverallStatus)
	if stepsOut.OverallStatus != 0 {
		log.For("cli").Error("workflow failed", "msg", msg)
		for _, o := range stepsOut.Outputs {
			if o.Status != 0 {
				log.For("cli").Error("Failed command", "command", o.Command, "stdout", o.Output, "stderr", o.Error)
				break
			}
		}
		os.Exit(stepsOut.OverallStatus)
	}
	log.For("cli").Info("Success", "msg", msg)
	return nil
}

func configSummaryLines(cfg *domain.NeptuneConfig) []string {
	repo := cfg.Repository
	if repo == nil {
		return nil
	}
	var lines []string
	if repo.GitHub != nil {
		lines = append(lines,
			"- Repository: "+repo.GitHub.Repository,
			"- Pull request branch: "+repo.GitHub.PullRequestBranch,
			"- Pull request number: "+repo.GitHub.PullRequestNumber,
		)
	}
	lines = append(lines,
		"- Object storage:",
		"  "+repo.ObjectStorage,
		"- Plan requirements: "+strings.Join(repo.PlanRequirements, ", "),
		"- Apply requirements: "+strings.Join(repo.ApplyRequirements, ", "),
		"- Allowed workflow: "+repo.AllowedWorkflow,
	)
	wf, ok := cfg.Workflows.Workflows[repo.AllowedWorkflow]
	if ok {
		var phases []string
		for p := range wf.Phases {
			phases = append(phases, p)
		}
		lines = append(lines, "- Workflow phases: "+strings.Join(phases, ", "))
	}
	return lines
}

func stepsSummaryLines(phase string, out *domain.StepsOutput) []string {
	if out == nil {
		return []string{"- Phase: " + phase}
	}
	lines := []string{"", "- Phase: " + phase, "- Steps:"}
	for _, o := range out.Outputs {
		statusStr := fmt.Sprint(o.Status)
		lines = append(lines, "  - Command: "+o.Command)
		lines = append(lines, "    - Status: "+statusStr)
	}
	return lines
}

func notifyAndExit(cfg *domain.NeptuneConfig, output string, status int) {
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		if err := notifier.CreateComment(&domain.PullRequestComment{
			SimpleOutput:  output,
			OverallStatus: status,
		}); err != nil {
			log.For("cli").Error("failed to post comment", "err", err)
		}
	}
	if status != 0 {
		log.For("cli").Error("Error", "msg", output)
		os.Exit(status)
	}
	log.For("cli").Info("Success", "msg", output)
	os.Exit(0)
}
