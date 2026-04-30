// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Env represents the deployment environment.
type Env string

const (
	EnvDevelopment Env = "development"
	EnvStaging     Env = "staging"
	EnvProduction  Env = "production"
)

// LogLevel represents the logging verbosity.
type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// Config holds all application configuration.
type Config struct {
	AppEnv   Env
	LogLevel LogLevel
}

var validEnvs = map[Env]bool{
	EnvDevelopment: true,
	EnvStaging:     true,
	EnvProduction:  true,
}

var validLogLevels = map[LogLevel]bool{
	LogDebug: true,
	LogInfo:  true,
	LogWarn:  true,
	LogError: true,
}

// Load reads configuration from environment variables.
// APP_ENV is required; LOG_LEVEL defaults to "info".
func Load() (*Config, error) {
	appEnv := Env(os.Getenv("APP_ENV"))
	if appEnv == "" {
		return nil, fmt.Errorf("APP_ENV is required (development|staging|production)")
	}
	if !validEnvs[appEnv] {
		return nil, fmt.Errorf("APP_ENV %q is invalid: must be development, staging, or production", appEnv)
	}

	logLevel := LogLevel(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = LogInfo
	}
	if !validLogLevels[logLevel] {
		return nil, fmt.Errorf("LOG_LEVEL %q is invalid: must be debug, info, warn, or error", logLevel)
	}

	return &Config{
		AppEnv:   appEnv,
		LogLevel: logLevel,
	}, nil
}
