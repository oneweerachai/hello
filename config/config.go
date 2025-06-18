package config

import (
	"os"
	"user-api/tracing"
)

// Config holds application configuration
type Config struct {
	Port        string
	Environment string
	Tracing     tracing.TracingConfig
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	environment := getEnv("ENVIRONMENT", "development")

	config := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: environment,
		Tracing:     tracing.LoadTracingConfigFromEnv(environment),
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
