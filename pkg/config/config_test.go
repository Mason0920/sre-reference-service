package config_test

import (
	"testing"

	"github.com/hslee/hslee/pkg/config"
)

func TestLoad_Success(t *testing.T) {
	tests := []struct {
		name         string
		appEnv       string
		logLevel     string
		wantEnv      config.Env
		wantLogLevel config.LogLevel
	}{
		{
			name:         "development with default log level",
			appEnv:       "development",
			wantEnv:      config.EnvDevelopment,
			wantLogLevel: config.LogInfo,
		},
		{
			name:         "staging with debug log level",
			appEnv:       "staging",
			logLevel:     "debug",
			wantEnv:      config.EnvStaging,
			wantLogLevel: config.LogDebug,
		},
		{
			name:         "production with error log level",
			appEnv:       "production",
			logLevel:     "error",
			wantEnv:      config.EnvProduction,
			wantLogLevel: config.LogError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			if tt.logLevel != "" {
				t.Setenv("LOG_LEVEL", tt.logLevel)
			} else {
				t.Setenv("LOG_LEVEL", "")
			}

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.AppEnv != tt.wantEnv {
				t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, tt.wantEnv)
			}
			if cfg.LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, tt.wantLogLevel)
			}
		})
	}
}

func TestLoad_MissingAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing APP_ENV, got nil")
	}
}

func TestLoad_InvalidAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "invalid-env")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid APP_ENV, got nil")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "verbose")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid LOG_LEVEL, got nil")
	}
}
