// Package main provides the entry point for the Telegram bot application.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pocky-ops-bot/internal/bot"
	"github.com/pocky-ops-bot/internal/bot/handlers"
	"github.com/pocky-ops-bot/internal/clients/ai"
	"github.com/pocky-ops-bot/internal/clients/binance"
	"github.com/pocky-ops-bot/internal/clients/telegram"
	"github.com/pocky-ops-bot/internal/config"
	"github.com/pocky-ops-bot/internal/services"
	"github.com/pocky-ops-bot/internal/tools"
	binancetools "github.com/pocky-ops-bot/internal/tools/binance"
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
	botUser, err := poller.GetMe(ctx)
	if err != nil {
		slog.Error("Failed to connect to Telegram API", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot connected successfully",
		"username", "@"+botUser.Username,
		"id", botUser.ID,
		"name", botUser.FirstName,
	)

	// Create Telegram sender
	sender, err := telegram.NewSender(
		cfg.TelegramToken,
		telegram.WithSenderLogger(logger),
	)
	if err != nil {
		slog.Error("Failed to create sender", "error", err)
		os.Exit(1)
	}

	// Register command menu with Telegram (appears when user types "/")
	if err := sender.SetMyCommands(ctx, []telegram.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "balance", Description: "View Binance portfolio & today's P&L"},
		{Command: "clear", Description: "Clear conversation history"},
		{Command: "help", Description: "Show available commands"},
	}); err != nil {
		slog.Warn("Failed to set bot commands", "error", err)
	}

	// Create AI client
	aiOpts := []ai.ClientOption{
		ai.WithProvider(ai.Provider(cfg.AIProvider)),
		ai.WithModel(cfg.AIModel),
		ai.WithMaxTokens(cfg.AIMaxTokens),
		ai.WithAITimeout(cfg.AITimeout),
		ai.WithAILogger(logger),
	}
	if cfg.AIBaseURL != "" {
		aiOpts = append(aiOpts, ai.WithBaseURL(cfg.AIBaseURL))
	}
	aiClient, err := ai.NewClient(cfg.AIAPIKey, aiOpts...)
	if err != nil {
		slog.Error("Failed to create AI client", "error", err)
		os.Exit(1)
	}

	// Create tool registry with Binance tools (if configured)
	var chatOpts []services.ChatServiceOption
	if cfg.BinanceAPIKey != "" && cfg.BinanceSecretKey != "" {
		bnOpts := []binance.ClientOption{
			binance.WithLogger(logger),
		}
		if cfg.BinanceBaseURL != "" {
			bnOpts = append(bnOpts, binance.WithBaseURL(cfg.BinanceBaseURL))
		}

		bnClient, err := binance.NewClient(cfg.BinanceAPIKey, cfg.BinanceSecretKey, bnOpts...)
		if err != nil {
			slog.Error("Failed to create Binance client", "error", err)
			os.Exit(1)
		}

		registry := tools.NewRegistry(logger)
		registry.Register(binancetools.NewGetBalancesTool(bnClient, logger))
		registry.Register(binancetools.NewGetPricesTool(bnClient, logger))
		registry.Register(binancetools.NewGet24hrStatsTool(bnClient, logger))

		chatOpts = append(chatOpts, services.WithTools(registry))
		slog.Info("Binance tools registered", "tools", 3)
	}

	// Create stateless chat service
	chatService := services.NewChatService(aiClient, cfg.AISystemPrompt, logger, chatOpts...)

	// Build router for stateless commands (/start, /help)
	router := bot.NewRouter(logger)
	cmdHandler := handlers.NewCommandHandler(sender, logger)
	router.RegisterCommand("start", cmdHandler.Start)
	router.RegisterCommand("help", cmdHandler.Help)

	// Create dispatcher — channel per-chat, zero shared state
	dispatcher := bot.NewDispatcher(router, chatService, sender, logger,
		bot.WithBufferSize(5),
		bot.WithIdleTTL(cfg.ConversationTTL),
		bot.WithMaxTurns(cfg.ConversationMaxTurns),
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)

	// Start polling with dispatcher
	if err := poller.StartWithHandler(ctx, dispatcher.Dispatch); err != nil {
		slog.Error("Failed to start poller", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot is running. Press Ctrl+C to stop.",
		"ai_provider", cfg.AIProvider,
		"ai_model", cfg.AIModel,
	)

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("Shutting down gracefully...")
	cancel()
	poller.Stop()
	dispatcher.Shutdown()
	slog.Info("Bot stopped successfully.")
}
