package bot

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pocky-ops-bot/internal/bot/types"
	"github.com/pocky-ops-bot/internal/clients/ai"
)

// mockSender implements MessageSender for testing.
type mockSender struct {
	mu       sync.Mutex
	texts    []mockText
	actions  []mockAction
}

type mockText struct {
	chatID int64
	text   string
}

type mockAction struct {
	chatID int64
	action string
}

func (m *mockSender) SendText(ctx context.Context, chatID int64, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.texts = append(m.texts, mockText{chatID: chatID, text: text})
	return nil
}

func (m *mockSender) SendChatAction(ctx context.Context, chatID int64, action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.actions = append(m.actions, mockAction{chatID: chatID, action: action})
	return nil
}

func (m *mockSender) getTexts() []mockText {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mockText, len(m.texts))
	copy(result, m.texts)
	return result
}

// mockChat implements ChatCompleter for testing.
type mockChat struct {
	mu       sync.Mutex
	reply    string
	calls    []mockChatCall
}

type mockChatCall struct {
	history  []ai.ChatMessage
	userText string
}

func (m *mockChat) GenerateResponse(ctx context.Context, history []ai.ChatMessage, userText string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	histCopy := make([]ai.ChatMessage, len(history))
	copy(histCopy, history)
	m.calls = append(m.calls, mockChatCall{history: histCopy, userText: userText})
	return m.reply, nil
}

func (m *mockChat) getCalls() []mockChatCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mockChatCall, len(m.calls))
	copy(result, m.calls)
	return result
}

func TestDispatcher_TextMessage(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "AI reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "Hello",
			Chat: types.Chat{ID: 42},
		},
	}

	d.Dispatch(ctx, update)
	time.Sleep(50 * time.Millisecond)

	texts := sender.getTexts()
	if len(texts) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(texts))
	}
	if texts[0].text != "AI reply" {
		t.Errorf("reply = %q, want 'AI reply'", texts[0].text)
	}
	if texts[0].chatID != 42 {
		t.Errorf("chatID = %d, want 42", texts[0].chatID)
	}
}

func TestDispatcher_SequentialSameChat(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send 2 messages to same chat
	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "msg1", Chat: types.Chat{ID: 42}},
	})
	d.Dispatch(ctx, types.Update{
		UpdateID: 2,
		Message:  &types.Message{ID: 2, Text: "msg2", Chat: types.Chat{ID: 42}},
	})

	time.Sleep(100 * time.Millisecond)

	// Second AI call should have history from first
	calls := chat.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 AI calls, got %d", len(calls))
	}

	// First call: no history
	if len(calls[0].history) != 0 {
		t.Errorf("first call history len = %d, want 0", len(calls[0].history))
	}

	// Second call: should have history from first exchange
	if len(calls[1].history) != 2 {
		t.Fatalf("second call history len = %d, want 2", len(calls[1].history))
	}
	if calls[1].history[0].Content != "msg1" {
		t.Errorf("history[0] = %q, want 'msg1'", calls[1].history[0].Content)
	}
	if calls[1].history[1].Content != "reply" {
		t.Errorf("history[1] = %q, want 'reply'", calls[1].history[1].Content)
	}
}

func TestDispatcher_ConcurrentDifferentChats(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send messages to 2 different chats
	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "from chat 1", Chat: types.Chat{ID: 1}},
	})
	d.Dispatch(ctx, types.Update{
		UpdateID: 2,
		Message:  &types.Message{ID: 2, Text: "from chat 2", Chat: types.Chat{ID: 2}},
	})

	time.Sleep(100 * time.Millisecond)

	if d.ActiveWorkers() != 2 {
		t.Errorf("active workers = %d, want 2", d.ActiveWorkers())
	}

	calls := chat.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 AI calls, got %d", len(calls))
	}
}

func TestDispatcher_ClearCommand(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Build some history
	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "hello", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(50 * time.Millisecond)

	// Clear
	d.Dispatch(ctx, types.Update{
		UpdateID: 2,
		Message:  &types.Message{ID: 2, Text: "/clear", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(50 * time.Millisecond)

	// Send another message — should have empty history
	d.Dispatch(ctx, types.Update{
		UpdateID: 3,
		Message:  &types.Message{ID: 3, Text: "after clear", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(50 * time.Millisecond)

	calls := chat.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 AI calls (before and after clear), got %d", len(calls))
	}

	// After clear, history should be empty
	if len(calls[1].history) != 0 {
		t.Errorf("history after clear: len = %d, want 0", len(calls[1].history))
	}

	// Verify clear confirmation was sent
	texts := sender.getTexts()
	found := false
	for _, txt := range texts {
		if txt.text == "Conversation history cleared." {
			found = true
			break
		}
	}
	if !found {
		t.Error("clear confirmation message not sent")
	}
}

func TestDispatcher_IdleCleanup(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "hello", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(30 * time.Millisecond)

	if d.ActiveWorkers() != 1 {
		t.Errorf("active workers = %d, want 1", d.ActiveWorkers())
	}

	// Wait for idle timeout
	time.Sleep(100 * time.Millisecond)

	if d.ActiveWorkers() != 0 {
		t.Errorf("active workers after idle = %d, want 0", d.ActiveWorkers())
	}
}

func TestDispatcher_GracefulShutdown(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Minute))

	ctx, cancel := context.WithCancel(context.Background())

	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "hello", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(50 * time.Millisecond)

	// Cancel context triggers shutdown
	cancel()

	// Shutdown should return promptly
	done := make(chan struct{})
	go func() {
		d.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("Shutdown() did not return within 1 second")
	}

	if d.ActiveWorkers() != 0 {
		t.Errorf("active workers after shutdown = %d, want 0", d.ActiveWorkers())
	}
}

func TestDispatcher_NilMessage(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil)

	ctx := context.Background()

	// Update without message — should be ignored
	err := d.Dispatch(ctx, types.Update{UpdateID: 1})
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if d.ActiveWorkers() != 0 {
		t.Errorf("active workers = %d, want 0 for nil message", d.ActiveWorkers())
	}
}

func TestDispatcher_MaxTurns(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "r"}
	router := NewRouter(nil)
	d := NewDispatcher(router, chat, sender, nil,
		WithIdleTTL(time.Second),
		WithMaxTurns(4), // 2 exchanges max
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send 3 messages — should trim history to maxTurns
	for i := 0; i < 3; i++ {
		d.Dispatch(ctx, types.Update{
			UpdateID: i,
			Message:  &types.Message{ID: i, Text: "msg", Chat: types.Chat{ID: 42}},
		})
		time.Sleep(50 * time.Millisecond)
	}

	calls := chat.getCalls()
	if len(calls) != 3 {
		t.Fatalf("expected 3 AI calls, got %d", len(calls))
	}

	// Third call should have trimmed history (4 messages max, not 4)
	if len(calls[2].history) > 4 {
		t.Errorf("history len = %d, want <= 4", len(calls[2].history))
	}
}

func TestExtractChatID(t *testing.T) {
	tests := []struct {
		name   string
		update types.Update
		want   int64
	}{
		{
			name:   "from message",
			update: types.Update{Message: &types.Message{Chat: types.Chat{ID: 42}}},
			want:   42,
		},
		{
			name: "from callback query",
			update: types.Update{CallbackQuery: &types.CallbackQuery{
				Message: &types.Message{Chat: types.Chat{ID: 99}},
			}},
			want: 99,
		},
		{
			name:   "from edited message",
			update: types.Update{EditedMessage: &types.Message{Chat: types.Chat{ID: 77}}},
			want:   77,
		},
		{
			name:   "empty update",
			update: types.Update{},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractChatID(tt.update)
			if got != tt.want {
				t.Errorf("extractChatID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDispatcher_CommandRouting(t *testing.T) {
	sender := &mockSender{}
	chat := &mockChat{reply: "reply"}
	router := NewRouter(nil)

	startCalled := false
	router.RegisterCommand("start", func(ctx context.Context, msg *types.Message) error {
		startCalled = true
		return nil
	})

	d := NewDispatcher(router, chat, sender, nil, WithIdleTTL(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.Dispatch(ctx, types.Update{
		UpdateID: 1,
		Message:  &types.Message{ID: 1, Text: "/start", Chat: types.Chat{ID: 42}},
	})
	time.Sleep(50 * time.Millisecond)

	if !startCalled {
		t.Error("/start command was not routed to handler")
	}

	// AI should not be called for commands
	calls := chat.getCalls()
	if len(calls) != 0 {
		t.Errorf("AI should not be called for commands, got %d calls", len(calls))
	}
}
