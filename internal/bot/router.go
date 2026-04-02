// Package bot provides the core bot routing and handler infrastructure.
package bot

import (
	"context"
	"log/slog"
	"strings"

	"github.com/pocky-ops-bot/internal/bot/types"
)

// CommandHandler processes a specific bot command.
type CommandHandler func(ctx context.Context, msg *types.Message) error

// UpdateHandler is a function type for processing updates.
type UpdateHandler func(ctx context.Context, update types.Update) error

// Router dispatches Telegram updates to the appropriate handler.
type Router struct {
	commands    map[string]CommandHandler
	chatHandler UpdateHandler
	logger      *slog.Logger
}

// NewRouter creates a new Router with the given logger.
func NewRouter(logger *slog.Logger) *Router {
	if logger == nil {
		logger = slog.Default()
	}
	return &Router{
		commands: make(map[string]CommandHandler),
		logger:   logger,
	}
}

// RegisterCommand registers a handler for a /command.
// The command should be without the leading slash (e.g. "start", "help").
func (r *Router) RegisterCommand(cmd string, handler CommandHandler) {
	r.commands[cmd] = handler
}

// SetChatHandler sets the fallback handler for non-command text messages.
func (r *Router) SetChatHandler(handler UpdateHandler) {
	r.chatHandler = handler
}

// Handle dispatches an incoming update to the appropriate handler.
// It implements the UpdateHandler signature.
func (r *Router) Handle(ctx context.Context, update types.Update) error {
	if update.Message == nil {
		return nil
	}

	text := update.Message.Text
	if text == "" {
		return nil
	}

	// Check if it's a bot command (starts with "/")
	if text[0] == '/' {
		cmd := extractCommand(text)
		if handler, ok := r.commands[cmd]; ok {
			r.logger.Debug("routing to command handler",
				slog.String("command", cmd),
				slog.Int64("chat_id", update.Message.Chat.ID),
			)
			return handler(ctx, update.Message)
		}
		r.logger.Debug("unknown command, ignoring",
			slog.String("command", cmd),
		)
		return nil
	}

	// Fallback to chat handler for non-command text
	if r.chatHandler != nil {
		return r.chatHandler(ctx, update)
	}

	return nil
}

// extractCommand parses a command from message text.
// "/start" → "start"
// "/help@botname" → "help"
// "/clear arg1 arg2" → "clear"
func extractCommand(text string) string {
	// Remove leading "/"
	cmd := text[1:]

	// Take only the first word
	if idx := strings.IndexByte(cmd, ' '); idx != -1 {
		cmd = cmd[:idx]
	}

	// Remove @botname suffix
	if idx := strings.IndexByte(cmd, '@'); idx != -1 {
		cmd = cmd[:idx]
	}

	return strings.ToLower(cmd)
}
