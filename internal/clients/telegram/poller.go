// Package telegram provides a Telegram Bot API client implementation.
package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pocky-ops-bot/internal/bot/types"
)

// PollerConfig holds configuration options for the Poller.
type PollerConfig struct {
	// Token is the Telegram Bot API token (required).
	Token string

	// BaseURL is the Telegram Bot API base URL.
	// Defaults to "https://api.telegram.org" if empty.
	BaseURL string

	// PollInterval is the minimum time between polling requests.
	// Defaults to DefaultPollInterval.
	PollInterval time.Duration

	// Timeout is the long-polling timeout in seconds (0-50).
	// Defaults to DefaultTimeout.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts for transient failures.
	// Defaults to DefaultMaxRetries.
	MaxRetries int

	// Backoff is the strategy for calculating retry delays.
	// Defaults to ExponentialBackoff if nil.
	Backoff BackoffStrategy

	// AllowedUpdates specifies the update types to receive.
	// Empty slice means all update types.
	AllowedUpdates []AllowedUpdateType

	// UpdatesChanSize is the buffer size for the updates channel.
	// Defaults to DefaultUpdatesChanSize.
	UpdatesChanSize int

	// HTTPClient is the HTTP client to use for requests.
	// Defaults to http.DefaultClient with appropriate timeout if nil.
	HTTPClient HTTPClient

	// Logger is the structured logger for debug output.
	// Defaults to slog.Default() if nil.
	Logger *slog.Logger
}

// Poller handles long-polling for Telegram Bot API updates.
type Poller struct {
	config  PollerConfig
	updates chan types.Update
	offset  atomic.Int64
	running atomic.Bool
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

// validate checks the configuration and applies defaults.
func (c *PollerConfig) validate() error {
	if c.Token == "" {
		return fmt.Errorf("telegram: bot token is required")
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://api.telegram.org"
	}

	if c.PollInterval <= 0 {
		c.PollInterval = DefaultPollInterval
	}

	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	// Telegram API allows timeout up to 50 seconds
	if c.Timeout > 50*time.Second {
		c.Timeout = 50 * time.Second
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.Backoff == nil {
		c.Backoff = NewExponentialBackoff()
	}

	if c.UpdatesChanSize <= 0 {
		c.UpdatesChanSize = DefaultUpdatesChanSize
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout + 10*time.Second, // Add buffer for network overhead
		}
	}

	if c.Logger == nil {
		c.Logger = slog.Default()
	}

	return nil
}

// NewPoller creates a new Poller with the given configuration.
func NewPoller(config PollerConfig) (*Poller, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	return &Poller{
		config:  config,
		updates: make(chan types.Update, config.UpdatesChanSize),
		stopCh:  make(chan struct{}),
	}, nil
}

// Updates returns a read-only channel for receiving updates.
func (p *Poller) Updates() <-chan types.Update {
	return p.updates
}

// Start begins the polling loop in a goroutine.
// It returns an error if the poller is already running.
func (p *Poller) Start(ctx context.Context) error {
	if !p.running.CompareAndSwap(false, true) {
		return fmt.Errorf("telegram: poller is already running")
	}

	p.wg.Add(1)
	go p.pollLoop(ctx)

	p.config.Logger.Info("telegram poller started",
		slog.Duration("timeout", p.config.Timeout),
		slog.Duration("poll_interval", p.config.PollInterval),
	)

	return nil
}

// Stop gracefully shuts down the polling loop.
// It blocks until the polling goroutine has finished.
func (p *Poller) Stop() {
	if !p.running.CompareAndSwap(true, false) {
		return // Already stopped
	}

	close(p.stopCh)

	p.wg.Wait()
	close(p.updates)

	p.config.Logger.Info("telegram poller stopped")
}

// IsRunning returns true if the poller is currently running.
func (p *Poller) IsRunning() bool {
	return p.running.Load()
}

// Offset returns the current update offset.
func (p *Poller) Offset() int64 {
	return p.offset.Load()
}

// SetOffset sets the update offset for the next poll.
func (p *Poller) SetOffset(offset int64) {
	p.offset.Store(offset)
}

// pollLoop is the main polling loop that runs in a goroutine.
func (p *Poller) pollLoop(ctx context.Context) {
	defer p.wg.Done()

	retryCount := 0
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.config.Logger.Debug("polling stopped due to context cancellation",
				slog.String("reason", ctx.Err().Error()),
			)
			return
		case <-p.stopCh:
			p.config.Logger.Debug("polling stopped due to stop signal")
			return
		case <-ticker.C:
			updates, err := p.getUpdates(ctx)
			if err != nil {
				if p.handleError(ctx, err, &retryCount) {
					continue
				}
				return
			}

			// Reset retry count on success
			retryCount = 0
			p.config.Backoff.Reset()

			// Process updates
			for _, update := range updates {
				select {
				case <-ctx.Done():
					return
				case <-p.stopCh:
					return
				case p.updates <- update:
					// Update offset to acknowledge this update
					p.offset.Store(int64(update.UpdateID + 1))
				}
			}
		}
	}
}

// handleError processes errors from getUpdates and returns true if polling should continue.
func (p *Poller) handleError(ctx context.Context, err error, retryCount *int) bool {
	// Check if context is cancelled
	if ctx.Err() != nil {
		return false
	}

	apiErr, isAPIErr := err.(*APIError)

	// Log the error
	p.config.Logger.Error("failed to get updates",
		slog.String("error", err.Error()),
		slog.Int("retry_count", *retryCount),
	)

	// Check if we should retry
	if *retryCount >= p.config.MaxRetries {
		p.config.Logger.Error("max retries exceeded, stopping poller",
			slog.Int("max_retries", p.config.MaxRetries),
		)
		return false
	}

	// Calculate backoff duration
	var backoff time.Duration
	if isAPIErr && apiErr.RetryAfter > 0 {
		backoff = time.Duration(apiErr.RetryAfter) * time.Second
	} else if isAPIErr && !apiErr.IsRetryable() {
		// Non-retryable error
		p.config.Logger.Error("non-retryable API error, stopping poller",
			slog.Int("error_code", apiErr.Code),
		)
		return false
	} else {
		backoff = p.config.Backoff.NextBackoff(*retryCount)
	}

	p.config.Logger.Info("retrying after backoff",
		slog.Duration("backoff", backoff),
		slog.Int("attempt", *retryCount+1),
	)

	*retryCount++

	// Wait for backoff duration
	select {
	case <-ctx.Done():
		return false
	case <-p.stopCh:
		return false
	case <-time.After(backoff):
		return true
	}
}

// getUpdates calls the Telegram getUpdates API method.
func (p *Poller) getUpdates(ctx context.Context) ([]types.Update, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("timeout", strconv.Itoa(int(p.config.Timeout.Seconds())))

	offset := p.offset.Load()
	if offset > 0 {
		params.Set("offset", strconv.FormatInt(offset, 10))
	}

	if len(p.config.AllowedUpdates) > 0 {
		updates := make([]string, len(p.config.AllowedUpdates))
		for i, u := range p.config.AllowedUpdates {
			updates[i] = string(u)
		}
		// Telegram expects JSON array format for allowed_updates
		allowedJSON, _ := json.Marshal(updates)
		params.Set("allowed_updates", string(allowedJSON))
	}

	// Build request URL
	apiURL := fmt.Sprintf("%s/bot%s/getUpdates?%s",
		p.config.BaseURL,
		p.config.Token,
		params.Encode(),
	)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	p.config.Logger.Debug("polling for updates",
		slog.Int64("offset", offset),
		slog.Duration("timeout", p.config.Timeout),
	)

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to read response: %w", err)
	}

	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("telegram: failed to parse response: %w", err)
	}

	// Check for API error
	if !apiResp.OK {
		apiErr := &APIError{
			Code:        apiResp.ErrorCode,
			Description: apiResp.Description,
		}
		if apiResp.Parameters != nil {
			apiErr.RetryAfter = apiResp.Parameters.RetryAfter
		}
		return nil, apiErr
	}

	// Parse updates
	var updates []types.Update
	if err := json.Unmarshal(apiResp.Result, &updates); err != nil {
		return nil, fmt.Errorf("telegram: failed to parse updates: %w", err)
	}

	if len(updates) > 0 {
		p.config.Logger.Debug("received updates",
			slog.Int("count", len(updates)),
			slog.Int("first_id", updates[0].UpdateID),
			slog.Int("last_id", updates[len(updates)-1].UpdateID),
		)
	}

	return updates, nil
}

// GetMe calls the getMe API method to test the bot token and get bot info.
func (p *Poller) GetMe(ctx context.Context) (*types.User, error) {
	apiURL := fmt.Sprintf("%s/bot%s/getMe", p.config.BaseURL, p.config.Token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram: failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("telegram: failed to parse response: %w", err)
	}

	if !apiResp.OK {
		return nil, &APIError{
			Code:        apiResp.ErrorCode,
			Description: apiResp.Description,
		}
	}

	var user types.User
	if err := json.Unmarshal(apiResp.Result, &user); err != nil {
		return nil, fmt.Errorf("telegram: failed to parse user: %w", err)
	}

	return &user, nil
}

// PollerOption is a functional option for configuring the Poller.
type PollerOption func(*PollerConfig)

// WithPollInterval sets the polling interval.
func WithPollInterval(d time.Duration) PollerOption {
	return func(c *PollerConfig) {
		c.PollInterval = d
	}
}

// WithTimeout sets the long-polling timeout.
func WithTimeout(d time.Duration) PollerOption {
	return func(c *PollerConfig) {
		c.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(n int) PollerOption {
	return func(c *PollerConfig) {
		c.MaxRetries = n
	}
}

// WithBackoff sets the backoff strategy.
func WithBackoff(b BackoffStrategy) PollerOption {
	return func(c *PollerConfig) {
		c.Backoff = b
	}
}

// WithAllowedUpdates sets the allowed update types.
func WithAllowedUpdates(updates ...AllowedUpdateType) PollerOption {
	return func(c *PollerConfig) {
		c.AllowedUpdates = updates
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client HTTPClient) PollerOption {
	return func(c *PollerConfig) {
		c.HTTPClient = client
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) PollerOption {
	return func(c *PollerConfig) {
		c.Logger = logger
	}
}

// WithBaseURL sets the API base URL (useful for testing).
func WithBaseURL(url string) PollerOption {
	return func(c *PollerConfig) {
		c.BaseURL = url
	}
}

// NewPollerWithOptions creates a new Poller with functional options.
func NewPollerWithOptions(token string, opts ...PollerOption) (*Poller, error) {
	config := PollerConfig{
		Token: token,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return NewPoller(config)
}

// UpdateHandler is a function type for processing updates.
type UpdateHandler func(ctx context.Context, update types.Update) error

// StartWithHandler starts the poller and processes updates with the given handler.
// This is a convenience method that handles the common pattern of consuming updates.
func (p *Poller) StartWithHandler(ctx context.Context, handler UpdateHandler) error {
	if err := p.Start(ctx); err != nil {
		return err
	}

	go func() {
		for update := range p.Updates() {
			if err := handler(ctx, update); err != nil {
				p.config.Logger.Error("handler error",
					slog.String("error", err.Error()),
					slog.Int("update_id", update.UpdateID),
				)
			}
		}
	}()

	return nil
}
