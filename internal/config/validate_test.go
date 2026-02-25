package config

import (
	"testing"

	"neptune/internal/domain"
)

func TestValidate_ObjectStorage(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:   "http://bad",
			Branch:          "main",
			AllowedWorkflow: "default",
			GitHub:          &domain.GitHubConfig{PullRequestBranch: "feature"},
		},
		Workflows: &domain.Workflows{
			Workflows: map[string]domain.WorkflowStatement{
				"default": {
					Phases: map[string]domain.WorkflowPhase{
						"plan":  {Steps: []domain.WorkflowStep{{Run: "echo 1"}}},
						"apply": {Steps: []domain.WorkflowStep{{Run: "echo 2"}}},
					},
				},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for non-gs URL")
	}
}

func TestValidate_PlanApplyPhases(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:   "gs://bucket",
			Branch:          "main",
			AllowedWorkflow: "default",
			GitHub:          &domain.GitHubConfig{PullRequestBranch: "feature"},
		},
		Workflows: &domain.Workflows{
			Workflows: map[string]domain.WorkflowStatement{
				"default": {
					Phases: map[string]domain.WorkflowPhase{
						"plan": {Steps: []domain.WorkflowStep{{Run: "echo 1"}}},
						// missing apply
					},
				},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for missing apply phase")
	}
}

func TestValidate_TerragruntWithoutTerramate(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:   "gs://bucket",
			Branch:          "main",
			AllowedWorkflow: "default",
			GitHub:          &domain.GitHubConfig{PullRequestBranch: "feature"},
		},
		Workflows: &domain.Workflows{
			Workflows: map[string]domain.WorkflowStatement{
				"default": {
					Phases: map[string]domain.WorkflowPhase{
						"plan":  {Steps: []domain.WorkflowStep{{Run: "terragrunt plan"}}},
						"apply": {Steps: []domain.WorkflowStep{{Run: "terragrunt apply"}}},
					},
				},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error when terragrunt used without terramate and --changed")
	}
}

func TestValidate_Ok(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:   "gs://bucket",
			Branch:          "main",
			AllowedWorkflow: "default",
			GitHub:          &domain.GitHubConfig{PullRequestBranch: "feature"},
		},
		Workflows: &domain.Workflows{
			Workflows: map[string]domain.WorkflowStatement{
				"default": {
					Phases: map[string]domain.WorkflowPhase{
						"plan":  {Steps: []domain.WorkflowStep{{Run: "terramate run --changed -- terragrunt plan"}}},
						"apply": {Steps: []domain.WorkflowStep{{Run: "terramate run --changed -- terragrunt apply"}}},
					},
				},
			},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Fatal(err)
	}
}
