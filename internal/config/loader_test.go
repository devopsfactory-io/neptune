package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv_Missing(t *testing.T) {
	os.Unsetenv("NEPTUNE_E2E")
	for _, k := range requiredEnvVars {
		os.Unsetenv(k)
	}
	_, err := LoadEnv()
	if err == nil {
		t.Fatal("expected error when required env vars missing")
	}
	if !isLoadError(err) {
		t.Errorf("expected LoadError, got %T", err)
	}
}

func TestLoadEnv_E2EMode_AllowsEmptyToken(t *testing.T) {
	os.Setenv("NEPTUNE_E2E", "1")
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GITHUB_PULL_REQUEST_BRANCH")
	os.Unsetenv("GITHUB_PULL_REQUEST_NUMBER")
	os.Unsetenv("GITHUB_RUN_ID")
	defer func() {
		os.Unsetenv("NEPTUNE_E2E")
	}()
	env, err := LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env["GITHUB_TOKEN"] != "" {
		t.Errorf("e2e mode should allow empty token, got %q", env["GITHUB_TOKEN"])
	}
	if env["GITHUB_REPOSITORY"] != "e2e/neptune-test" {
		t.Errorf("e2e default repo: got %q", env["GITHUB_REPOSITORY"])
	}
	if env["GITHUB_PULL_REQUEST_NUMBER"] != "1" {
		t.Errorf("e2e default PR number: got %q", env["GITHUB_PULL_REQUEST_NUMBER"])
	}
}

func TestLoadEnv_Success(t *testing.T) {
	os.Setenv("NEPTUNE_CONFIG_PATH", ".neptune.yaml")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GITHUB_PULL_REQUEST_BRANCH", "feature")
	os.Setenv("GITHUB_PULL_REQUEST_NUMBER", "1")
	os.Setenv("GITHUB_RUN_ID", "3")
	os.Setenv("GITHUB_TOKEN", "token")
	defer func() {
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GITHUB_PULL_REQUEST_BRANCH")
		os.Unsetenv("GITHUB_PULL_REQUEST_NUMBER")
		os.Unsetenv("GITHUB_RUN_ID")
		os.Unsetenv("GITHUB_TOKEN")
	}()
	env, err := LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env["GITHUB_REPOSITORY"] != "owner/repo" {
		t.Errorf("got %q", env["GITHUB_REPOSITORY"])
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	env := map[string]string{
		"NEPTUNE_CONFIG_PATH":        "/nonexistent/.neptune.yaml",
		"GITHUB_REPOSITORY":          "o/r",
		"GITHUB_PULL_REQUEST_BRANCH": "b",
		"GITHUB_PULL_REQUEST_NUMBER": "1",
		"GITHUB_RUN_ID":              "3",
		"GITHUB_TOKEN":               "t",
	}
	_, err := Load(env)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".neptune.yaml")
	content := `
repository:
  object_storage: gs://bucket
  branch: main
  plan_requirements: []
  apply_requirements: []
  allowed_workflow: default
workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on: [plan]
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"NEPTUNE_CONFIG_PATH":        path,
		"GITHUB_REPOSITORY":          "owner/repo",
		"GITHUB_PULL_REQUEST_BRANCH": "feature",
		"GITHUB_PULL_REQUEST_NUMBER": "1",
		"GITHUB_RUN_ID":              "3",
		"GITHUB_TOKEN":               "token",
	}
	cfg, err := Load(env)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Repository.ObjectStorage != "gs://bucket" {
		t.Errorf("got object_storage %q", cfg.Repository.ObjectStorage)
	}
	if cfg.Repository.Branch != "main" {
		t.Errorf("got branch %q", cfg.Repository.Branch)
	}
	wf, ok := cfg.Workflows.Workflows["default"]
	if !ok {
		t.Fatal("workflow default not found")
	}
	if _, ok := wf.Phases["plan"]; !ok {
		t.Fatal("phase plan not found")
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("default log_level should be INFO, got %q", cfg.LogLevel)
	}
}

func TestLoad_LogLevel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".neptune.yaml")
	content := `
log_level: debug
repository:
  object_storage: gs://bucket
  branch: main
  plan_requirements: []
  apply_requirements: []
  allowed_workflow: default
workflows:
  default:
    plan:
      steps:
        - run: terramate run --changed -- terragrunt plan
    apply:
      depends_on: [plan]
      steps:
        - run: terramate run --changed -- terragrunt apply -auto-approve
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"NEPTUNE_CONFIG_PATH":        path,
		"GITHUB_REPOSITORY":          "owner/repo",
		"GITHUB_PULL_REQUEST_BRANCH": "feature",
		"GITHUB_PULL_REQUEST_NUMBER": "1",
		"GITHUB_RUN_ID":              "3",
		"GITHUB_TOKEN":               "token",
	}
	os.Unsetenv("NEPTUNE_LOG_LEVEL")
	t.Cleanup(func() { os.Unsetenv("NEPTUNE_LOG_LEVEL") })
	cfg, err := Load(env)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log_level from YAML: got %q, want debug", cfg.LogLevel)
	}
	os.Setenv("NEPTUNE_LOG_LEVEL", "ERROR")
	cfg2, err := Load(env)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.LogLevel != "ERROR" {
		t.Errorf("NEPTUNE_LOG_LEVEL should override: got %q, want ERROR", cfg2.LogLevel)
	}
}

func isLoadError(err error) bool {
	_, ok := err.(*LoadError)
	return ok
}
