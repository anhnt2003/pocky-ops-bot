// Package config provides configuration loading for the bot application.
package config

import (
	"os"
	"strconv"
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

	// AIProvider is the AI service provider (gemini, claude, openai).
	AIProvider string

	// AIAPIKey is the API key for the AI provider.
	AIAPIKey string

	// AIModel is the model name to use (e.g. gemini-2.0-flash, claude-sonnet-4-20250514).
	AIModel string

	// AIBaseURL overrides the default API base URL for the provider (optional).
	AIBaseURL string

	// AIMaxTokens is the maximum number of tokens in AI responses.
	AIMaxTokens int

	// AITimeout is the timeout for AI API requests.
	AITimeout time.Duration

	// AISystemPrompt is the system prompt for AI conversations.
	AISystemPrompt string

	// ConversationMaxTurns is the maximum number of message pairs to keep in history.
	ConversationMaxTurns int

	// ConversationTTL is the time-to-live for conversation history.
	ConversationTTL time.Duration

	// BinanceAPIKey is the Binance API key for portfolio tracking.
	BinanceAPIKey string

	// BinanceSecretKey is the Binance secret key for HMAC request signing.
	BinanceSecretKey string

	// BinanceBaseURL overrides the default Binance API base URL (optional, for testnet).
	BinanceBaseURL string
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

		AIProvider:     getEnvOrDefault("AI_PROVIDER", "gemini"),
		AIAPIKey:       os.Getenv("AI_API_KEY"),
		AIModel:        getEnvOrDefault("AI_MODEL", "gemini-2.0-flash"),
		AIBaseURL:      os.Getenv("AI_BASE_URL"),
		AIMaxTokens:    parseInt("AI_MAX_TOKENS", 1024),
		AITimeout:      parseDuration("AI_TIMEOUT", 60*time.Second),
		AISystemPrompt: getEnvOrDefault("AI_SYSTEM_PROMPT", "You are Pocky, a helpful and friendly assistant."),

		ConversationMaxTurns: parseInt("CONVERSATION_MAX_TURNS", 20),
		ConversationTTL:      parseDuration("CONVERSATION_TTL", 30*time.Minute),

		BinanceAPIKey:    os.Getenv("BINANCE_API_KEY"),
		BinanceSecretKey: os.Getenv("BINANCE_SECRET_KEY"),
		BinanceBaseURL:   os.Getenv("BINANCE_BASE_URL"),
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

// parseInt parses an integer from an environment variable.
func parseInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
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
