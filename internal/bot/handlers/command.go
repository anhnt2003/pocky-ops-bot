// Package handlers provides Telegram update handler implementations.
package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pocky-ops-bot/internal/bot/types"
)

// MessageSender is the interface for sending text messages back to Telegram.
// Defined at the consumer side for testability.
type MessageSender interface {
	SendText(ctx context.Context, chatID int64, text string) error
}

// CommandHandler handles bot commands like /start, /help, /clear.
type CommandHandler struct {
	sender MessageSender
	logger *slog.Logger
}

// NewCommandHandler creates a new CommandHandler.
func NewCommandHandler(sender MessageSender, logger *slog.Logger) *CommandHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &CommandHandler{
		sender: sender,
		logger: logger,
	}
}

// Start handles the /start command.
func (h *CommandHandler) Start(ctx context.Context, msg *types.Message) error {
	name := "there"
	if msg.From != nil && msg.From.FirstName != "" {
		name = msg.From.FirstName
	}

	text := fmt.Sprintf("Hi %s! I'm Pocky Bot 🤖\n\nSend me a message and I'll reply using AI.\n\nCommands:\n/balance - View Binance portfolio & today's P&L\n/help - Show available commands\n/clear - Clear conversation history", name)

	return h.sender.SendText(ctx, msg.Chat.ID, text)
}

// Help handles the /help command.
func (h *CommandHandler) Help(ctx context.Context, msg *types.Message) error {
	text := "Available commands:\n\n/start - Start the bot\n/balance - View Binance portfolio & today's P&L\n/help - Show this help message\n/clear - Clear conversation history"

	return h.sender.SendText(ctx, msg.Chat.ID, text)
}

