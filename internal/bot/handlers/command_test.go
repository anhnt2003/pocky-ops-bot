package handlers

import (
	"context"
	"testing"

	"github.com/pocky-ops-bot/internal/bot/types"
)

// mockSender implements MessageSender for testing.
type mockSender struct {
	messages []sentMessage
	err      error
}

type sentMessage struct {
	chatID int64
	text   string
}

func (m *mockSender) SendText(ctx context.Context, chatID int64, text string) error {
	m.messages = append(m.messages, sentMessage{chatID: chatID, text: text})
	return m.err
}

func TestCommandHandler_Start(t *testing.T) {
	sender := &mockSender{}
	handler := NewCommandHandler(sender, nil)

	msg := &types.Message{
		ID:   1,
		Chat: types.Chat{ID: 42},
		From: &types.User{FirstName: "Alice"},
	}

	err := handler.Start(context.Background(), msg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.messages))
	}

	if sender.messages[0].chatID != 42 {
		t.Errorf("chatID = %d, want 42", sender.messages[0].chatID)
	}

	if sender.messages[0].text == "" {
		t.Error("message text should not be empty")
	}
}

func TestCommandHandler_Start_NoFrom(t *testing.T) {
	sender := &mockSender{}
	handler := NewCommandHandler(sender, nil)

	msg := &types.Message{
		ID:   1,
		Chat: types.Chat{ID: 42},
		From: nil,
	}

	err := handler.Start(context.Background(), msg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.messages))
	}
}

func TestCommandHandler_Help(t *testing.T) {
	sender := &mockSender{}
	handler := NewCommandHandler(sender, nil)

	msg := &types.Message{
		ID:   1,
		Chat: types.Chat{ID: 42},
	}

	err := handler.Help(context.Background(), msg)
	if err != nil {
		t.Fatalf("Help() error = %v", err)
	}

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.messages))
	}

	if sender.messages[0].chatID != 42 {
		t.Errorf("chatID = %d, want 42", sender.messages[0].chatID)
	}
}

