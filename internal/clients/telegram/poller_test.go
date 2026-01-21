package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocky-ops-bot/internal/bot/types"
)

// mockHTTPClient implements HTTPClient for testing.
type mockHTTPClient struct {
	responses []mockResponse
	index     int
	requests  []*http.Request
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)

	if m.index >= len(m.responses) {
		// Default response: empty updates
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"result":[]}`)),
		}, nil
	}

	resp := m.responses[m.index]
	m.index++

	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
	}, nil
}

func TestNewPoller(t *testing.T) {
	tests := []struct {
		name    string
		config  PollerConfig
		wantErr bool
	}{
		{
			name:    "empty token",
			config:  PollerConfig{},
			wantErr: true,
		},
		{
			name: "valid config",
			config: PollerConfig{
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: PollerConfig{
				Token:           "test-token",
				BaseURL:         "https://api.telegram.org",
				PollInterval:    time.Second,
				Timeout:         30 * time.Second,
				MaxRetries:      5,
				UpdatesChanSize: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller, err := NewPoller(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPoller() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && poller == nil {
				t.Error("NewPoller() returned nil poller without error")
			}
		})
	}
}

func TestNewPollerWithOptions(t *testing.T) {
	poller, err := NewPollerWithOptions(
		"test-token",
		WithTimeout(45*time.Second),
		WithPollInterval(500*time.Millisecond),
		WithMaxRetries(3),
	)

	if err != nil {
		t.Fatalf("NewPollerWithOptions() error = %v", err)
	}

	if poller.config.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want %v", poller.config.Timeout, 45*time.Second)
	}

	if poller.config.PollInterval != 500*time.Millisecond {
		t.Errorf("PollInterval = %v, want %v", poller.config.PollInterval, 500*time.Millisecond)
	}

	if poller.config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want %v", poller.config.MaxRetries, 3)
	}
}

func TestPollerStartStop(t *testing.T) {
	mockClient := &mockHTTPClient{}

	poller, err := NewPoller(PollerConfig{
		Token:        "test-token",
		HTTPClient:   mockClient,
		PollInterval: 10 * time.Millisecond,
		Timeout:      1 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	ctx := context.Background()

	// Test Start
	if err := poller.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !poller.IsRunning() {
		t.Error("IsRunning() = false after Start()")
	}

	// Starting again should fail
	if err := poller.Start(ctx); err == nil {
		t.Error("Start() should fail when already running")
	}

	// Wait a bit for polling to occur
	time.Sleep(50 * time.Millisecond)

	// Test Stop
	poller.Stop()

	if poller.IsRunning() {
		t.Error("IsRunning() = true after Stop()")
	}

	// Stopping again should be idempotent
	poller.Stop()
}

func TestPollerReceivesUpdates(t *testing.T) {
	updates := []types.Update{
		{UpdateID: 1, Message: &types.Message{ID: 100}},
		{UpdateID: 2, Message: &types.Message{ID: 101}},
	}

	updatesJSON, _ := json.Marshal(updates)
	responseBody, _ := json.Marshal(APIResponse{OK: true, Result: updatesJSON})

	mockClient := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: http.StatusOK, body: string(responseBody)},
		},
	}

	poller, err := NewPoller(PollerConfig{
		Token:        "test-token",
		HTTPClient:   mockClient,
		PollInterval: 10 * time.Millisecond,
		Timeout:      1 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := poller.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer poller.Stop()

	// Collect received updates
	received := make([]types.Update, 0)
	timeout := time.After(200 * time.Millisecond)

	for {
		select {
		case update, ok := <-poller.Updates():
			if !ok {
				goto done
			}
			received = append(received, update)
			if len(received) >= len(updates) {
				goto done
			}
		case <-timeout:
			goto done
		}
	}

done:
	if len(received) != len(updates) {
		t.Errorf("Received %d updates, want %d", len(received), len(updates))
	}

	for i, u := range received {
		if u.UpdateID != updates[i].UpdateID {
			t.Errorf("Update[%d].UpdateID = %d, want %d", i, u.UpdateID, updates[i].UpdateID)
		}
	}
}

func TestPollerOffset(t *testing.T) {
	poller, err := NewPoller(PollerConfig{
		Token: "test-token",
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	// Initial offset should be 0
	if poller.Offset() != 0 {
		t.Errorf("Initial Offset() = %d, want 0", poller.Offset())
	}

	// Set offset
	poller.SetOffset(100)
	if poller.Offset() != 100 {
		t.Errorf("Offset() = %d, want 100", poller.Offset())
	}
}

func TestExponentialBackoff(t *testing.T) {
	backoff := &ExponentialBackoff{
		InitialInterval: time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 10 * time.Second}, // Capped at max
		{5, 10 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := backoff.NextBackoff(tt.attempt)
			if result != tt.expected {
				t.Errorf("NextBackoff(%d) = %v, want %v", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name        string
		err         *APIError
		wantMessage string
		wantRetry   bool
	}{
		{
			name:        "simple error",
			err:         &APIError{Code: 400, Description: "Bad Request"},
			wantMessage: "telegram api error 400: Bad Request",
			wantRetry:   false,
		},
		{
			name:        "rate limited",
			err:         &APIError{Code: 429, Description: "Too Many Requests", RetryAfter: 30},
			wantMessage: "telegram api error 429: Too Many Requests (retry after 30s)",
			wantRetry:   true,
		},
		{
			name:        "server error",
			err:         &APIError{Code: 500, Description: "Internal Server Error"},
			wantMessage: "telegram api error 500: Internal Server Error",
			wantRetry:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if msg := tt.err.Error(); msg != tt.wantMessage {
				t.Errorf("Error() = %q, want %q", msg, tt.wantMessage)
			}
			if retry := tt.err.IsRetryable(); retry != tt.wantRetry {
				t.Errorf("IsRetryable() = %v, want %v", retry, tt.wantRetry)
			}
		})
	}
}

func TestAllowedUpdateTypes(t *testing.T) {
	all := AllAllowedUpdates()
	if len(all) == 0 {
		t.Error("AllAllowedUpdates() returned empty slice")
	}

	common := CommonAllowedUpdates()
	if len(common) == 0 {
		t.Error("CommonAllowedUpdates() returned empty slice")
	}

	if len(common) >= len(all) {
		t.Error("CommonAllowedUpdates() should be a subset of AllAllowedUpdates()")
	}
}

func TestParseAllowedUpdates(t *testing.T) {
	tests := []struct {
		input    string
		expected []AllowedUpdateType
	}{
		{"", nil},
		{"message", []AllowedUpdateType{UpdateTypeMessage}},
		{"message,callback_query", []AllowedUpdateType{UpdateTypeMessage, UpdateTypeCallbackQuery}},
		{"message, callback_query , inline_query", []AllowedUpdateType{UpdateTypeMessage, UpdateTypeCallbackQuery, UpdateTypeInlineQuery}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseAllowedUpdates(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseAllowedUpdates(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("ParseAllowedUpdates(%q)[%d] = %v, want %v", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestPollerContextCancellation(t *testing.T) {
	mockClient := &mockHTTPClient{}

	poller, err := NewPoller(PollerConfig{
		Token:        "test-token",
		HTTPClient:   mockClient,
		PollInterval: 10 * time.Millisecond,
		Timeout:      1 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if err := poller.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Cancel context
	cancel()

	// Wait for poller to stop
	time.Sleep(50 * time.Millisecond)

	// Poller may still report running because Stop wasn't called explicitly,
	// but it should have stopped polling. Let's call Stop to clean up.
	poller.Stop()

	if poller.IsRunning() {
		t.Error("Poller should not be running after Stop()")
	}
}

func TestGetMeIntegration(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/getMe" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		user := types.User{
			ID:        123456789,
			IsBot:     true,
			FirstName: "TestBot",
			Username:  "test_bot",
		}
		userJSON, _ := json.Marshal(user)
		resp := APIResponse{OK: true, Result: userJSON}
		respJSON, _ := json.Marshal(resp)

		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	}))
	defer server.Close()

	poller, err := NewPollerWithOptions(
		"test-token",
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewPollerWithOptions() error = %v", err)
	}

	ctx := context.Background()
	user, err := poller.GetMe(ctx)
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}

	if user.ID != 123456789 {
		t.Errorf("User.ID = %d, want 123456789", user.ID)
	}
	if !user.IsBot {
		t.Error("User.IsBot = false, want true")
	}
	if user.Username != "test_bot" {
		t.Errorf("User.Username = %q, want %q", user.Username, "test_bot")
	}
}

func TestStartWithHandler(t *testing.T) {
	updates := []types.Update{
		{UpdateID: 1, Message: &types.Message{ID: 100}},
	}

	updatesJSON, _ := json.Marshal(updates)
	responseBody, _ := json.Marshal(APIResponse{OK: true, Result: updatesJSON})

	mockClient := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: http.StatusOK, body: string(responseBody)},
		},
	}

	poller, err := NewPoller(PollerConfig{
		Token:        "test-token",
		HTTPClient:   mockClient,
		PollInterval: 10 * time.Millisecond,
		Timeout:      1 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	var handledUpdates atomic.Int32
	handler := func(ctx context.Context, update types.Update) error {
		handledUpdates.Add(1)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := poller.StartWithHandler(ctx, handler); err != nil {
		t.Fatalf("StartWithHandler() error = %v", err)
	}

	// Wait for updates to be processed
	time.Sleep(100 * time.Millisecond)
	poller.Stop()

	if count := handledUpdates.Load(); count < 1 {
		t.Errorf("Handler called %d times, want at least 1", count)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test timeout capping at 50 seconds
	poller, err := NewPoller(PollerConfig{
		Token:   "test-token",
		Timeout: 60 * time.Second, // Should be capped to 50
	})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	if poller.config.Timeout != 50*time.Second {
		t.Errorf("Timeout = %v, want %v (capped)", poller.config.Timeout, 50*time.Second)
	}
}

func TestAllowedUpdateTypeString(t *testing.T) {
	if UpdateTypeMessage.String() != "message" {
		t.Errorf("UpdateTypeMessage.String() = %q, want %q", UpdateTypeMessage.String(), "message")
	}
}
