package bot

import (
	"context"
	"fmt"
	"testing"

	"github.com/pocky-ops-bot/internal/bot/types"
)

func TestExtractCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/start", "start"},
		{"/help", "help"},
		{"/clear arg1 arg2", "clear"},
		{"/help@mybot", "help"},
		{"/start@mybot extra args", "start"},
		{"/UPPER", "upper"},
		{"/a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractCommand(tt.input)
			if result != tt.expected {
				t.Errorf("extractCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRouterHandle_Command(t *testing.T) {
	router := NewRouter(nil)

	var handledMsg *types.Message
	router.RegisterCommand("start", func(ctx context.Context, msg *types.Message) error {
		handledMsg = msg
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "/start",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if handledMsg == nil {
		t.Fatal("command handler was not called")
	}
	if handledMsg.ID != 100 {
		t.Errorf("handled message ID = %d, want 100", handledMsg.ID)
	}
}

func TestRouterHandle_CommandWithArgs(t *testing.T) {
	router := NewRouter(nil)

	called := false
	router.RegisterCommand("clear", func(ctx context.Context, msg *types.Message) error {
		called = true
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "/clear all history",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !called {
		t.Fatal("command handler was not called for command with args")
	}
}

func TestRouterHandle_CommandWithBotMention(t *testing.T) {
	router := NewRouter(nil)

	called := false
	router.RegisterCommand("help", func(ctx context.Context, msg *types.Message) error {
		called = true
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "/help@pocky_bot",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !called {
		t.Fatal("command handler was not called for command with bot mention")
	}
}

func TestRouterHandle_UnknownCommand(t *testing.T) {
	router := NewRouter(nil)

	chatHandlerCalled := false
	router.SetChatHandler(func(ctx context.Context, update types.Update) error {
		chatHandlerCalled = true
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "/unknown_cmd",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if chatHandlerCalled {
		t.Error("chat handler should not be called for unknown commands")
	}
}

func TestRouterHandle_TextMessage(t *testing.T) {
	router := NewRouter(nil)

	var handledUpdate types.Update
	router.SetChatHandler(func(ctx context.Context, update types.Update) error {
		handledUpdate = update
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "hello, how are you?",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if handledUpdate.UpdateID != 1 {
		t.Errorf("chat handler received update ID = %d, want 1", handledUpdate.UpdateID)
	}
}

func TestRouterHandle_NilMessage(t *testing.T) {
	router := NewRouter(nil)

	update := types.Update{
		UpdateID: 1,
		Message:  nil,
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v for nil message", err)
	}
}

func TestRouterHandle_EmptyText(t *testing.T) {
	router := NewRouter(nil)

	chatHandlerCalled := false
	router.SetChatHandler(func(ctx context.Context, update types.Update) error {
		chatHandlerCalled = true
		return nil
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if chatHandlerCalled {
		t.Error("chat handler should not be called for empty text")
	}
}

func TestRouterHandle_NoChatHandler(t *testing.T) {
	router := NewRouter(nil)

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "hello",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != nil {
		t.Fatalf("Handle() error = %v when no chat handler set", err)
	}
}

func TestRouterHandle_CommandError(t *testing.T) {
	router := NewRouter(nil)

	expectedErr := fmt.Errorf("command failed")
	router.RegisterCommand("fail", func(ctx context.Context, msg *types.Message) error {
		return expectedErr
	})

	update := types.Update{
		UpdateID: 1,
		Message: &types.Message{
			ID:   100,
			Text: "/fail",
			Chat: types.Chat{ID: 42},
		},
	}

	err := router.Handle(context.Background(), update)
	if err != expectedErr {
		t.Errorf("Handle() error = %v, want %v", err, expectedErr)
	}
}
