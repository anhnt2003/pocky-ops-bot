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

	text := fmt.Sprintf("Xin chào %s! Mình là Pocky Bot 🤖\n\nGửi tin nhắn cho mình, mình sẽ trả lời bằng AI nhé.\n\n📋 Lệnh:\n/dautu - 💰 Xem danh mục đầu tư Spot & Futures\n/trogiup - ❓ Hướng dẫn sử dụng\n/xoa - 🗑️ Xoá lịch sử trò chuyện", name)

	return h.sender.SendText(ctx, msg.Chat.ID, text)
}

// Help handles the /help command.
func (h *CommandHandler) Help(ctx context.Context, msg *types.Message) error {
	text := "📋 Các lệnh có sẵn:\n\n🚀 /start - Bắt đầu sử dụng bot\n💰 /dautu - Xem danh mục đầu tư Spot & Futures\n❓ /trogiup - Hướng dẫn sử dụng\n🗑️ /xoa - Xoá lịch sử trò chuyện"

	return h.sender.SendText(ctx, msg.Chat.ID, text)
}

