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
		listenAddr   string
		otelEndpoint string
		otelService  string
		wantEnv      config.Env
		wantLogLevel config.LogLevel
		wantAddr     string
		wantService  string
	}{
		{
			name:         "development with defaults",
			appEnv:       "development",
			wantEnv:      config.EnvDevelopment,
			wantLogLevel: config.LogInfo,
			wantAddr:     ":8080",
			wantService:  "hslee",
		},
		{
			name:         "staging with debug log level",
			appEnv:       "staging",
			logLevel:     "debug",
			wantEnv:      config.EnvStaging,
			wantLogLevel: config.LogDebug,
			wantAddr:     ":8080",
			wantService:  "hslee",
		},
		{
			name:         "production with error log level",
			appEnv:       "production",
			logLevel:     "error",
			wantEnv:      config.EnvProduction,
			wantLogLevel: config.LogError,
			wantAddr:     ":8080",
			wantService:  "hslee",
		},
		{
			name:         "custom listen addr and otel",
			appEnv:       "staging",
			listenAddr:   ":9090",
			otelEndpoint: "http://otel:4317",
			otelService:  "my-service",
			wantEnv:      config.EnvStaging,
			wantLogLevel: config.LogInfo,
			wantAddr:     ":9090",
			wantService:  "my-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv("LOG_LEVEL", tt.logLevel)
			t.Setenv("LISTEN_ADDR", tt.listenAddr)
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tt.otelEndpoint)
			t.Setenv("OTEL_SERVICE_NAME", tt.otelService)

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
			if cfg.ListenAddr != tt.wantAddr {
				t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, tt.wantAddr)
			}
			if cfg.OTELEndpoint != tt.otelEndpoint {
				t.Errorf("OTELEndpoint = %q, want %q", cfg.OTELEndpoint, tt.otelEndpoint)
			}
			if cfg.OTELService != tt.wantService {
				t.Errorf("OTELService = %q, want %q", cfg.OTELService, tt.wantService)
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

func TestLoad_DefaultListenAddr(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LISTEN_ADDR", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":8080")
	}
}

func TestLoad_DefaultOTELService(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("OTEL_SERVICE_NAME", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.OTELService != "hslee" {
		t.Errorf("OTELService = %q, want %q", cfg.OTELService, "hslee")
	}
}
