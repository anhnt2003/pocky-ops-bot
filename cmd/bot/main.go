// Package main provides the entry point for the Telegram bot application.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pocky-ops-bot/internal/bot/types"
	"github.com/pocky-ops-bot/internal/clients/telegram"
	"github.com/pocky-ops-bot/internal/config"
)

func main() {
	// Load configuration from environment/.env file
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.TelegramToken == "" {
		slog.Error("TELEGRAM_BOT_TOKEN is required. Set it in .env file or environment variable.")
		os.Exit(1)
	}

	// Setup structured logger
	var level slog.Level
	_ = level.UnmarshalText([]byte(cfg.LogLevel))
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Create Telegram poller
	poller, err := telegram.NewPollerWithOptions(
		cfg.TelegramToken,
		telegram.WithTimeout(cfg.Timeout),
		telegram.WithPollInterval(cfg.PollInterval),
		telegram.WithMaxRetries(cfg.MaxRetries),
		telegram.WithLogger(logger),
	)
	if err != nil {
		slog.Error("Failed to create poller", "error", err)
		os.Exit(1)
	}

	// Test connection to Telegram API
	ctx := context.Background()
	bot, err := poller.GetMe(ctx)
	if err != nil {
		slog.Error("Failed to connect to Telegram API", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot connected successfully",
		"username", "@"+bot.Username,
		"id", bot.ID,
		"name", bot.FirstName,
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)

	// Define update handler
	handler := func(ctx context.Context, update types.Update) error {
		if update.Message != nil {
			slog.Info("Message received",
				"update_id", update.UpdateID,
				"from", update.Message.From.Username,
				"chat_id", update.Message.Chat.ID,
				"text", update.Message.Text,
			)
		}

		if update.CallbackQuery != nil {
			slog.Info("Callback query received",
				"update_id", update.UpdateID,
				"data", update.CallbackQuery.Data,
			)
		}

		if update.EditedMessage != nil {
			slog.Info("Message edited",
				"update_id", update.UpdateID,
				"text", update.EditedMessage.Text,
			)
		}

		return nil
	}

	// Start polling
	if err := poller.StartWithHandler(ctx, handler); err != nil {
		slog.Error("Failed to start poller", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("Shutting down gracefully...")
	cancel()
	poller.Stop()
	slog.Info("Bot stopped successfully.")
}
