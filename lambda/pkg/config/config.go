package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Config holds the Lambda runtime configuration (GitHub App and secrets).
type Config struct {
	AppID         string
	PrivateKey    string
	WebhookSecret string
	AppSlug       string // optional, for comment @-mention matching
}

// Load reads config from environment and optionally Secrets Manager.
// Env: GITHUB_APP_ID, GITHUB_APP_SLUG (optional).
// If GITHUB_APP_WEBHOOK_SECRET_ARN and/or GITHUB_APP_PRIVATE_KEY_SECRET_ARN are set, those values are fetched from Secrets Manager.
// Otherwise GITHUB_APP_WEBHOOK_SECRET and GITHUB_APP_PRIVATE_KEY can be set directly (e.g. for local dev).
func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{
		AppID:   os.Getenv("GITHUB_APP_ID"),
		AppSlug: os.Getenv("GITHUB_APP_SLUG"),
	}
	if cfg.AppID == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID is required")
	}

	webhookSecretARN := os.Getenv("GITHUB_APP_WEBHOOK_SECRET_ARN")
	privateKeyARN := os.Getenv("GITHUB_APP_PRIVATE_KEY_SECRET_ARN")

	if webhookSecretARN != "" || privateKeyARN != "" {
		awsCfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("load AWS config: %w", err)
		}
		client := secretsmanager.NewFromConfig(awsCfg)
		if webhookSecretARN != "" {
			secret, err := getSecret(ctx, client, webhookSecretARN)
			if err != nil {
				return nil, fmt.Errorf("webhook secret: %w", err)
			}
			cfg.WebhookSecret = secret
		} else {
			cfg.WebhookSecret = os.Getenv("GITHUB_APP_WEBHOOK_SECRET")
		}
		if privateKeyARN != "" {
			secret, err := getSecret(ctx, client, privateKeyARN)
			if err != nil {
				return nil, fmt.Errorf("private key: %w", err)
			}
			cfg.PrivateKey = secret
		} else {
			cfg.PrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
		}
	} else {
		cfg.WebhookSecret = os.Getenv("GITHUB_APP_WEBHOOK_SECRET")
		cfg.PrivateKey = os.Getenv("GITHUB_APP_PRIVATE_KEY")
	}

	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("webhook secret is required (GITHUB_APP_WEBHOOK_SECRET or GITHUB_APP_WEBHOOK_SECRET_ARN)")
	}
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required (GITHUB_APP_PRIVATE_KEY or GITHUB_APP_PRIVATE_KEY_SECRET_ARN)")
	}
	return cfg, nil
}

// getSecret returns the secret string value from Secrets Manager (SecretString).
func getSecret(ctx context.Context, client *secretsmanager.Client, arn string) (string, error) {
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
	if err != nil {
		return "", err
	}
	if out.SecretString != nil {
		return *out.SecretString, nil
	}
	return "", fmt.Errorf("secret %s has no SecretString", arn)
}

// SecretJSON is used when a single secret holds JSON with webhook_secret and private_key.
func (c *Config) LoadFromJSONSecret(jsonBody string) error {
	var s struct {
		WebhookSecret string `json:"webhook_secret"`
		PrivateKey    string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(jsonBody), &s); err != nil {
		return fmt.Errorf("parse secret JSON: %w", err)
	}
	if s.WebhookSecret != "" {
		c.WebhookSecret = s.WebhookSecret
	}
	if s.PrivateKey != "" {
		c.PrivateKey = s.PrivateKey
	}
	return nil
}
