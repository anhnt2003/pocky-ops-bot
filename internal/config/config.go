// Package config provides configuration loading for the bot application.
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
type Config struct {
	// TelegramToken is the bot token from BotFather.
	TelegramToken string

	// LogLevel is the logging level (debug, info, warn, error).
	LogLevel string

	// PollInterval is the minimum time between polling requests.
	PollInterval time.Duration

	// Timeout is the long-polling timeout.
	Timeout time.Duration

	// MaxRetries is the maximum retry attempts for transient failures.
	MaxRetries int
}

// Load reads configuration from environment variables and .env file.
// Environment variables take precedence over .env file values.
func Load() (*Config, error) {
	// Load .env file if it exists (optional, won't error if missing)
	_ = godotenv.Load()

	cfg := &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
		PollInterval:  parseDuration("POLL_INTERVAL", time.Second),
		Timeout:       parseDuration("TIMEOUT", 30*time.Second),
		MaxRetries:    3,
	}

	return cfg, nil
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// parseDuration parses a duration from an environment variable.
func parseDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
