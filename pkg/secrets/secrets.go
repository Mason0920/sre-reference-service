// Package secrets provides a client for AWS Secrets Manager.
package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Provider is the interface for fetching secrets.
type Provider interface {
	GetString(ctx context.Context, name string) (string, error)
}

// Client fetches secrets from AWS Secrets Manager.
type Client struct {
	sm *secretsmanager.Client
}

// EnvProvider reads secrets from environment variables (for local development).
type EnvProvider struct{}

// New creates a Client using the default AWS credential chain (IRSA, instance profile, etc.).
func New(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("secrets: load aws config: %w", err)
	}
	return &Client{sm: secretsmanager.NewFromConfig(cfg)}, nil
}

// GetString fetches a secret string by ARN or name.
func (c *Client) GetString(ctx context.Context, name string) (string, error) {
	out, err := c.sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("secrets: get %q: %w", name, err)
	}
	if out.SecretString == nil {
		return "", fmt.Errorf("secrets: %q has no string value", name)
	}
	return *out.SecretString, nil
}

// GetJSON fetches a secret and JSON-unmarshals it into v.
func (c *Client) GetJSON(ctx context.Context, name string, v any) error {
	s, err := c.GetString(ctx, name)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), v); err != nil {
		return fmt.Errorf("secrets: unmarshal %q: %w", name, err)
	}
	return nil
}

// GetString reads the secret value from the environment variable matching name.
// Env var name is derived by upper-casing and replacing '/' with '_'.
func (e *EnvProvider) GetString(_ context.Context, name string) (string, error) {
	key := envKey(name)
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("secrets: env var %q not set (secret name: %q)", key, name)
	}
	return val, nil
}

// envKey converts a secret name like "hslee/prod/db-password" to "HSLEE_PROD_DB_PASSWORD".
func envKey(name string) string {
	b := make([]byte, len(name))
	for i := range len(name) {
		switch {
		case name[i] == '/' || name[i] == '-':
			b[i] = '_'
		case name[i] >= 'a' && name[i] <= 'z':
			b[i] = name[i] - 32
		default:
			b[i] = name[i]
		}
	}
	return string(b)
}
