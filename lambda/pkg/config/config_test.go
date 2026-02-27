package config

import (
	"context"
	"testing"
)

func TestLoad_EnvOnly_Success(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET", "whsec")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET_ARN", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN", "")

	cfg, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppID != "123" {
		t.Errorf("AppID: got %q", cfg.AppID)
	}
	if cfg.WebhookSecret != "whsec" {
		t.Errorf("WebhookSecret: got %q", cfg.WebhookSecret)
	}
	if cfg.PrivateKey == "" {
		t.Error("PrivateKey should be set")
	}
}

func TestLoad_EnvOnly_WithAppSlug(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GITHUB_APP_ID", "1")
	t.Setenv("GITHUB_APP_SLUG", "my-bot")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET", "s")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\nx\n-----END RSA PRIVATE KEY-----")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET_ARN", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN", "")

	cfg, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppSlug != "my-bot" {
		t.Errorf("AppSlug: got %q", cfg.AppSlug)
	}
}

func TestLoad_MissingAppID(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GITHUB_APP_ID", "")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET", "s")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "key")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET_ARN", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN", "")

	_, err := Load(ctx)
	if err == nil {
		t.Fatal("expected error when GITHUB_APP_ID is missing")
	}
	if err.Error() != "GITHUB_APP_ID is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_MissingWebhookSecret(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GITHUB_APP_ID", "1")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\nx\n-----END RSA PRIVATE KEY-----")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET_ARN", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN", "")

	_, err := Load(ctx)
	if err == nil {
		t.Fatal("expected error when webhook secret is missing")
	}
	if err.Error() != "webhook secret is required (GITHUB_APP_WEBHOOK_SECRET or GITHUB_APP_WEBHOOK_SECRET_ARN)" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_MissingPrivateKey(t *testing.T) {
	ctx := context.Background()
	t.Setenv("GITHUB_APP_ID", "1")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET", "s")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "")
	t.Setenv("GITHUB_APP_WEBHOOK_SECRET_ARN", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN", "")

	_, err := Load(ctx)
	if err == nil {
		t.Fatal("expected error when private key is missing")
	}
	if err.Error() != "private key is required (GITHUB_APP_PRIVATE_KEY or GITHUB_APP_PRIVATE_KEY_SECRET_ARN)" {
		t.Errorf("unexpected error: %v", err)
	}
}
