// Package telegram provides a Telegram Bot API client implementation.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// SenderConfig holds configuration options for the Sender.
type SenderConfig struct {
	// Token is the Telegram Bot API token (required).
	Token string

	// BaseURL is the Telegram Bot API base URL.
	// Defaults to "https://api.telegram.org" if empty.
	BaseURL string

	// HTTPClient is the HTTP client to use for requests.
	// Defaults to http.Client with Timeout if nil.
	HTTPClient HTTPClient

	// Logger is the structured logger for debug output.
	// Defaults to slog.Default() if nil.
	Logger *slog.Logger

	// Timeout is the HTTP client timeout.
	// Defaults to 30s. Used only when HTTPClient is nil.
	Timeout time.Duration
}

// validate checks the configuration and applies defaults.
func (c *SenderConfig) validate() error {
	if c.Token == "" {
		return fmt.Errorf("telegram: bot token is required")
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://api.telegram.org"
	}

	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
		}
	}

	if c.Logger == nil {
		c.Logger = slog.Default()
	}

	return nil
}

// Sender sends messages and actions via the Telegram Bot API.
type Sender struct {
	config SenderConfig
}

// SenderOption is a functional option for configuring the Sender.
type SenderOption func(*SenderConfig)

// WithSenderBaseURL sets the API base URL (useful for testing).
func WithSenderBaseURL(url string) SenderOption {
	return func(c *SenderConfig) {
		c.BaseURL = url
	}
}

// WithSenderHTTPClient sets the HTTP client.
func WithSenderHTTPClient(client HTTPClient) SenderOption {
	return func(c *SenderConfig) {
		c.HTTPClient = client
	}
}

// WithSenderLogger sets the logger.
func WithSenderLogger(logger *slog.Logger) SenderOption {
	return func(c *SenderConfig) {
		c.Logger = logger
	}
}

// WithSenderTimeout sets the HTTP client timeout.
func WithSenderTimeout(d time.Duration) SenderOption {
	return func(c *SenderConfig) {
		c.Timeout = d
	}
}

// NewSender creates a new Sender with the given token and functional options.
func NewSender(token string, opts ...SenderOption) (*Sender, error) {
	config := SenderConfig{
		Token: token,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &Sender{config: config}, nil
}

// SendOption is a functional option for configuring individual send requests.
type SendOption func(map[string]interface{})

// WithParseMode sets the parse mode for the message (e.g. "Markdown", "HTML").
func WithParseMode(mode string) SendOption {
	return func(body map[string]interface{}) {
		body["parse_mode"] = mode
	}
}

// WithDisableNotification sends the message silently.
func WithDisableNotification() SendOption {
	return func(body map[string]interface{}) {
		body["disable_notification"] = true
	}
}

// WithReplyToMessageID sets the message to reply to.
func WithReplyToMessageID(id int) SendOption {
	return func(body map[string]interface{}) {
		body["reply_to_message_id"] = id
	}
}

// SendText sends a plain text message to the specified chat.
// This is the simplified version of SendMessage without options.
func (s *Sender) SendText(ctx context.Context, chatID int64, text string) error {
	return s.SendMessage(ctx, chatID, text)
}

// SendMessage sends a text message to the specified chat.
// By default it uses Markdown parse mode. Use WithParseMode to override.
func (s *Sender) SendMessage(ctx context.Context, chatID int64, text string, opts ...SendOption) error {
	body := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	for _, opt := range opts {
		opt(body)
	}

	s.config.Logger.Debug("sending message",
		slog.Int64("chat_id", chatID),
		slog.Int("text_len", len(text)),
	)

	return s.doPost(ctx, "sendMessage", body)
}

// SendChatAction sends a chat action (e.g. "typing") to the specified chat.
func (s *Sender) SendChatAction(ctx context.Context, chatID int64, action string) error {
	body := map[string]interface{}{
		"chat_id": chatID,
		"action":  action,
	}

	s.config.Logger.Debug("sending chat action",
		slog.Int64("chat_id", chatID),
		slog.String("action", action),
	)

	return s.doPost(ctx, "sendChatAction", body)
}

// BotCommand represents a bot command for the Telegram command menu.
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// SetMyCommands registers the bot's command list with Telegram.
// These commands appear in the autocomplete menu when a user types "/".
func (s *Sender) SetMyCommands(ctx context.Context, commands []BotCommand) error {
	body := map[string]interface{}{
		"commands": commands,
	}

	s.config.Logger.Debug("setting bot commands",
		slog.Int("count", len(commands)),
	)

	return s.doPost(ctx, "setMyCommands", body)
}

// doPost performs a POST request to the given Telegram Bot API method.
func (s *Sender) doPost(ctx context.Context, method string, body map[string]interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("telegram: failed to marshal request body: %w", err)
	}

	apiURL := fmt.Sprintf("%s/bot%s/%s", s.config.BaseURL, s.config.Token, method)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("telegram: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("telegram: failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("telegram: failed to parse response: %w", err)
	}

	if !apiResp.OK {
		apiErr := &APIError{
			Code:        apiResp.ErrorCode,
			Description: apiResp.Description,
		}
		if apiResp.Parameters != nil {
			apiErr.RetryAfter = apiResp.Parameters.RetryAfter
		}
		return apiErr
	}

	return nil
}
