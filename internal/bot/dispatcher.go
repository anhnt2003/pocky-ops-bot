package bot

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pocky-ops-bot/internal/bot/types"
	"github.com/pocky-ops-bot/internal/clients/llm"
)

// MessageSender sends text messages and chat actions to Telegram.
type MessageSender interface {
	SendText(ctx context.Context, chatID int64, text string) error
	SendChatAction(ctx context.Context, chatID int64, action string) error
}

// ChatCompleter generates an AI response given conversation history and user text.
type ChatCompleter interface {
	GenerateResponse(ctx context.Context, history []llm.ChatMessage, userText string) (string, error)
}

// chatWorker represents an active per-chat goroutine.
type chatWorker struct {
	ch chan types.Update
}

// Dispatcher routes Telegram updates to per-chat worker goroutines.
// Each chat gets its own goroutine and channel, ensuring sequential
// processing per chat while allowing concurrent processing across chats.
// Conversation history is owned locally by each worker goroutine — no shared state.
type Dispatcher struct {
	workers  sync.Map // map[int64]*chatWorker
	router   *Router
	chat     ChatCompleter
	sender   MessageSender
	bufSize  int
	idleTTL  time.Duration
	maxTurns int
	logger   *slog.Logger
	wg       sync.WaitGroup
	active   atomic.Int64
}

// DispatcherOption is a functional option for configuring the Dispatcher.
type DispatcherOption func(*Dispatcher)

// WithBufferSize sets the per-chat channel buffer size.
func WithBufferSize(n int) DispatcherOption {
	return func(d *Dispatcher) {
		d.bufSize = n
	}
}

// WithIdleTTL sets the idle timeout for worker goroutines.
func WithIdleTTL(ttl time.Duration) DispatcherOption {
	return func(d *Dispatcher) {
		d.idleTTL = ttl
	}
}

// WithMaxTurns sets the maximum conversation history length.
func WithMaxTurns(n int) DispatcherOption {
	return func(d *Dispatcher) {
		d.maxTurns = n
	}
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(router *Router, chat ChatCompleter, sender MessageSender, logger *slog.Logger, opts ...DispatcherOption) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	d := &Dispatcher{
		router:   router,
		chat:     chat,
		sender:   sender,
		bufSize:  5,
		idleTTL:  30 * time.Minute,
		maxTurns: 40,
		logger:   logger,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Dispatch routes an update to the appropriate chat worker.
// It implements the UpdateHandler signature for use with Poller.StartWithHandler.
func (d *Dispatcher) Dispatch(ctx context.Context, update types.Update) error {
	chatID := extractChatID(update)
	if chatID == 0 {
		return nil
	}

	// Get or create worker for this chat
	val, loaded := d.workers.LoadOrStore(chatID, &chatWorker{
		ch: make(chan types.Update, d.bufSize),
	})
	worker := val.(*chatWorker)

	if !loaded {
		// New worker — spawn goroutine
		d.wg.Add(1)
		d.active.Add(1)
		d.logger.Debug("spawning chat worker",
			slog.Int64("chat_id", chatID),
		)
		go d.runWorker(ctx, chatID, worker.ch)
	}

	// Non-blocking send
	select {
	case worker.ch <- update:
	default:
		d.logger.Warn("chat queue full, dropping update",
			slog.Int64("chat_id", chatID),
			slog.Int("update_id", update.UpdateID),
		)
	}

	return nil
}

// Shutdown waits for all worker goroutines to finish.
func (d *Dispatcher) Shutdown() {
	d.wg.Wait()
}

// ActiveWorkers returns the number of active worker goroutines.
func (d *Dispatcher) ActiveWorkers() int {
	return int(d.active.Load())
}

// runWorker is the per-chat goroutine that processes updates sequentially.
func (d *Dispatcher) runWorker(ctx context.Context, chatID int64, ch <-chan types.Update) {
	defer d.wg.Done()
	defer d.active.Add(-1)
	defer d.workers.Delete(chatID)

	history := make([]llm.ChatMessage, 0)
	idle := time.NewTimer(d.idleTTL)
	defer idle.Stop()

	for {
		select {
		case update, ok := <-ch:
			if !ok {
				return
			}
			if !idle.Stop() {
				select {
				case <-idle.C:
				default:
				}
			}
			idle.Reset(d.idleTTL)

			history = d.handleUpdate(ctx, chatID, update, history)

		case <-idle.C:
			d.logger.Debug("chat worker idle, shutting down",
				slog.Int64("chat_id", chatID),
			)
			return

		case <-ctx.Done():
			return
		}
	}
}

// handleUpdate processes a single update within the worker goroutine.
// It returns the (possibly updated) history.
func (d *Dispatcher) handleUpdate(ctx context.Context, chatID int64, update types.Update, history []llm.ChatMessage) []llm.ChatMessage {
	msg := update.Message
	if msg == nil {
		// Handle callback queries, etc. through router
		d.router.Handle(ctx, update)
		return history
	}

	text := msg.Text
	if text == "" {
		return history
	}

	// Command handling
	if text[0] == '/' {
		cmd := extractCommand(text)
		switch cmd {
		case "xoa":
			history = history[:0]
			d.logger.Info("conversation cleared",
				slog.Int64("chat_id", chatID),
			)
			_ = d.sender.SendText(ctx, chatID, "🗑️ Đã xoá lịch sử trò chuyện.")
			return history

		case "dautu", "dautư":
			text = "Hiển thị tổng quan danh mục đầu tư Binance của tôi, bao gồm cả Spot và Futures:\n" +
				"1. Spot: liệt kê từng tài sản với giá trị USDT, tổng giá trị portfolio, và % lãi/lỗ 24h.\n" +
				"2. Futures: tổng số dư ví, lãi/lỗ chưa thực hiện, margin khả dụng, tất cả vị thế đang mở (giá vào, giá mark, P&L, đòn bẩy, giá thanh lý), và các lệnh đang chờ.\n" +
				"Dùng các tool có sẵn để lấy dữ liệu realtime."

		default:
			// Other commands (/start, /help) — delegate to router
			d.router.Handle(ctx, update)
			return history
		}
	}

	// Text message (or /balance prompt) → AI
	_ = d.sender.SendChatAction(ctx, chatID, "typing")

	reply, err := d.chat.GenerateResponse(ctx, history, text)
	if err != nil {
		d.logger.Error("ai response failed",
			slog.Int64("chat_id", chatID),
			slog.String("error", err.Error()),
		)
		_ = d.sender.SendText(ctx, chatID, "Sorry, I couldn't process that. Please try again.")
		return history
	}

	// Update local history
	history = append(history,
		llm.ChatMessage{Role: llm.RoleUser, Content: text},
		llm.ChatMessage{Role: llm.RoleAssistant, Content: reply},
	)

	// Trim to max turns
	if len(history) > d.maxTurns {
		history = history[len(history)-d.maxTurns:]
	}

	_ = d.sender.SendText(ctx, chatID, reply)
	return history
}

// extractChatID extracts the chat ID from an update.
func extractChatID(update types.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	if update.EditedMessage != nil {
		return update.EditedMessage.Chat.ID
	}
	return 0
}

// extractChatID and handleUpdate reuse extractCommand() from router.go (same package).
