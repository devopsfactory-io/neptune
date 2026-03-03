package config

import (
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
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
		t.Fatal("expected validation error for non-gs/non-s3 URL")
	}
}

func TestValidate_StacksManagement_Invalid(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:    "gs://bucket",
			StacksManagement: "other",
			Branch:           "main",
			AllowedWorkflow:  "default",
			GitHub:           &domain.GitHubConfig{PullRequestBranch: "feature"},
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
		t.Fatal("expected validation error for invalid stacks_management")
	}
	if err.Error() != "repository stacks_management must be terramate or local" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ObjectStorageS3(t *testing.T) {
	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			ObjectStorage:   "s3://my-bucket",
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
		t.Fatalf("s3:// URL should be valid: %v", err)
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
	onceTrue := true
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
						"plan":  {Steps: []domain.WorkflowStep{{Run: "terragrunt plan", Once: &onceTrue}}},
						"apply": {Steps: []domain.WorkflowStep{{Run: "terragrunt apply", Once: &onceTrue}}},
					},
				},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error when once is true and run uses terragrunt without terramate and --changed in run string")
	}
}

func TestValidate_TerragruntWithOnceUnset(t *testing.T) {
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
	if err := Validate(cfg); err != nil {
		t.Fatalf("terragrunt with once unset (per-stack) should be valid: %v", err)
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

func TestValidate_LogLevel_Invalid(t *testing.T) {
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
		LogLevel: "TRACE",
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for invalid log_level")
	}
	if err.Error() != "log_level must be one of: DEBUG, INFO, ERROR" {
		t.Errorf("unexpected error: %v", err)
	}
}
