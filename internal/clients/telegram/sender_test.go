package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSender(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		opts    []SenderOption
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "test-token",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:  "with options",
			token: "test-token",
			opts: []SenderOption{
				WithSenderBaseURL("https://custom.api.org"),
				WithSenderTimeout(10 * time.Second),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewSender(tt.token, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sender == nil {
				t.Error("NewSender() returned nil sender without error")
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/bottest-token/sendMessage" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify content type
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		// Verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var reqBody map[string]interface{}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// chat_id is sent as JSON number, which unmarshals to float64
		if chatID, ok := reqBody["chat_id"].(float64); !ok || int64(chatID) != 12345 {
			t.Errorf("chat_id = %v, want 12345", reqBody["chat_id"])
		}

		if text, ok := reqBody["text"].(string); !ok || text != "Hello, World!" {
			t.Errorf("text = %v, want %q", reqBody["text"], "Hello, World!")
		}

		if parseMode, ok := reqBody["parse_mode"].(string); !ok || parseMode != "Markdown" {
			t.Errorf("parse_mode = %v, want %q", reqBody["parse_mode"], "Markdown")
		}

		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{OK: true}
		respJSON, _ := json.Marshal(resp)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx := context.Background()
	if err := sender.SendMessage(ctx, 12345, "Hello, World!"); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
}

func TestSendMessageWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var reqBody map[string]interface{}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		if parseMode, ok := reqBody["parse_mode"].(string); !ok || parseMode != "HTML" {
			t.Errorf("parse_mode = %v, want %q", reqBody["parse_mode"], "HTML")
		}

		if dn, ok := reqBody["disable_notification"].(bool); !ok || !dn {
			t.Errorf("disable_notification = %v, want true", reqBody["disable_notification"])
		}

		if replyID, ok := reqBody["reply_to_message_id"].(float64); !ok || int(replyID) != 42 {
			t.Errorf("reply_to_message_id = %v, want 42", reqBody["reply_to_message_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{OK: true}
		respJSON, _ := json.Marshal(resp)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx := context.Background()
	err = sender.SendMessage(ctx, 12345, "Hello",
		WithParseMode("HTML"),
		WithDisableNotification(),
		WithReplyToMessageID(42),
	)
	if err != nil {
		t.Fatalf("SendMessage() with options error = %v", err)
	}
}

func TestSendMessageAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{
			OK:          false,
			ErrorCode:   400,
			Description: "Bad Request: chat not found",
		}
		respJSON, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx := context.Background()
	err = sender.SendMessage(ctx, 99999, "Hello")
	if err == nil {
		t.Fatal("SendMessage() expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T: %v", err, err)
	}

	if apiErr.Code != 400 {
		t.Errorf("APIError.Code = %d, want 400", apiErr.Code)
	}

	if apiErr.Description != "Bad Request: chat not found" {
		t.Errorf("APIError.Description = %q, want %q", apiErr.Description, "Bad Request: chat not found")
	}

	if apiErr.IsRetryable() {
		t.Error("Expected non-retryable error")
	}
}

func TestSendMessageAPIErrorWithRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{
			OK:          false,
			ErrorCode:   429,
			Description: "Too Many Requests",
			Parameters:  &ResponseParams{RetryAfter: 30},
		}
		respJSON, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx := context.Background()
	err = sender.SendMessage(ctx, 12345, "Hello")
	if err == nil {
		t.Fatal("SendMessage() expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T: %v", err, err)
	}

	if apiErr.Code != 429 {
		t.Errorf("APIError.Code = %d, want 429", apiErr.Code)
	}

	if apiErr.RetryAfter != 30 {
		t.Errorf("APIError.RetryAfter = %d, want 30", apiErr.RetryAfter)
	}

	if !apiErr.IsRetryable() {
		t.Error("Expected retryable error")
	}
}

func TestSendChatAction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/bottest-token/sendChatAction" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Verify content type
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		// Verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var reqBody map[string]interface{}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		if chatID, ok := reqBody["chat_id"].(float64); !ok || int64(chatID) != 12345 {
			t.Errorf("chat_id = %v, want 12345", reqBody["chat_id"])
		}

		if action, ok := reqBody["action"].(string); !ok || action != "typing" {
			t.Errorf("action = %v, want %q", reqBody["action"], "typing")
		}

		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{OK: true}
		respJSON, _ := json.Marshal(resp)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx := context.Background()
	if err := sender.SendChatAction(ctx, 12345, "typing"); err != nil {
		t.Fatalf("SendChatAction() error = %v", err)
	}
}

func TestSenderOptions(t *testing.T) {
	mockClient := &mockHTTPClient{}

	sender, err := NewSender("test-token",
		WithSenderBaseURL("https://custom.api.org"),
		WithSenderHTTPClient(mockClient),
		WithSenderTimeout(15*time.Second),
	)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	if sender.config.BaseURL != "https://custom.api.org" {
		t.Errorf("BaseURL = %q, want %q", sender.config.BaseURL, "https://custom.api.org")
	}

	if sender.config.Timeout != 15*time.Second {
		t.Errorf("Timeout = %v, want %v", sender.config.Timeout, 15*time.Second)
	}

	// HTTPClient should be the mock we provided, not a default
	if sender.config.HTTPClient != mockClient {
		t.Error("HTTPClient was not set to provided mock")
	}
}

func TestSenderDefaults(t *testing.T) {
	sender, err := NewSender("test-token")
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	if sender.config.BaseURL != "https://api.telegram.org" {
		t.Errorf("BaseURL = %q, want %q", sender.config.BaseURL, "https://api.telegram.org")
	}

	if sender.config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", sender.config.Timeout, 30*time.Second)
	}

	if sender.config.HTTPClient == nil {
		t.Error("HTTPClient should not be nil after validation")
	}

	if sender.config.Logger == nil {
		t.Error("Logger should not be nil after validation")
	}
}

func TestSendMessageContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		resp := APIResponse{OK: true}
		respJSON, _ := json.Marshal(resp)
		w.Write(respJSON)
	}))
	defer server.Close()

	sender, err := NewSender("test-token", WithSenderBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = sender.SendMessage(ctx, 12345, "Hello")
	if err == nil {
		t.Error("SendMessage() expected error for cancelled context, got nil")
	}
}
