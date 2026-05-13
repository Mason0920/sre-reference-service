package secrets

import (
	"context"
	"testing"
)

func TestEnvProvider_GetString(t *testing.T) {
	p := &EnvProvider{}
	ctx := context.Background()

	t.Run("returns value when env var set", func(t *testing.T) {
		t.Setenv("HSLEE_PROD_DB_PASSWORD", "supersecret")
		got, err := p.GetString(ctx, "hslee/prod/db-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "supersecret" {
			t.Errorf("got %q, want %q", got, "supersecret")
		}
	})

	t.Run("errors when env var missing", func(t *testing.T) {
		_, err := p.GetString(ctx, "hslee/prod/missing-secret")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestEnvKey(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"hslee/prod/db-password", "HSLEE_PROD_DB_PASSWORD"},
		{"hslee/staging/api-key", "HSLEE_STAGING_API_KEY"},
		{"simple", "SIMPLE"},
		{"MY_SECRET", "MY_SECRET"},
	}
	for _, tc := range cases {
		got := envKey(tc.name)
		if got != tc.want {
			t.Errorf("envKey(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// mockProvider implements Provider for testing callers.
type mockProvider struct {
	secrets map[string]string
	err     error
}

func (m *mockProvider) GetString(_ context.Context, name string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	v, ok := m.secrets[name]
	if !ok {
		return "", &notFoundError{name: name}
	}
	return v, nil
}

type notFoundError struct{ name string }

func (e *notFoundError) Error() string { return "secret not found: " + e.name }

func TestMockProvider(t *testing.T) {
	p := &mockProvider{secrets: map[string]string{"key": "value"}}
	ctx := context.Background()

	got, err := p.GetString(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}

	_, err = p.GetString(ctx, "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}
