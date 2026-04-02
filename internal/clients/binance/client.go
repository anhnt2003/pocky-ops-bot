package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Clock abstracts time.Now for deterministic testing.
type Clock interface {
	Now() time.Time
}

// realClock is the default Clock implementation using the system clock.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// ClientConfig holds configuration options for the Binance client.
type ClientConfig struct {
	// APIKey is the Binance API key (required).
	APIKey string

	// SecretKey is the Binance secret key (required).
	SecretKey string

	// BaseURL is the Binance REST API base URL.
	// Defaults to "https://api.binance.com".
	BaseURL string

	// HTTPClient is the HTTP client to use for requests.
	HTTPClient HTTPClient

	// Logger is the structured logger.
	Logger *slog.Logger

	// Timeout is the HTTP request timeout.
	// Defaults to 10s.
	Timeout time.Duration

	// RecvWindow is the Binance recvWindow parameter in milliseconds.
	// Defaults to 5000.
	RecvWindow int64

	// Clock provides the current time (useful for testing).
	Clock Clock
}

// validate checks the configuration and applies defaults.
func (c *ClientConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("binance: api key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("binance: secret key is required")
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://api.binance.com"
	}

	if c.Timeout <= 0 {
		c.Timeout = 10 * time.Second
	}

	if c.RecvWindow <= 0 {
		c.RecvWindow = 5000
	}

	if c.Clock == nil {
		c.Clock = realClock{}
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

// Client is a Binance REST API client.
type Client struct {
	config ClientConfig
}

// ClientOption is a functional option for configuring the Binance client.
type ClientOption func(*ClientConfig)

// WithBaseURL sets the API base URL (useful for testing).
func WithBaseURL(url string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = url
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client HTTPClient) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = client
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = d
	}
}

// WithRecvWindow sets the Binance recvWindow parameter.
func WithRecvWindow(rw int64) ClientOption {
	return func(c *ClientConfig) {
		c.RecvWindow = rw
	}
}

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *ClientConfig) {
		c.Logger = logger
	}
}

// WithClock sets the clock implementation (useful for testing).
func WithClock(clock Clock) ClientOption {
	return func(c *ClientConfig) {
		c.Clock = clock
	}
}

// NewClient creates a new Binance client with the given API credentials and options.
func NewClient(apiKey, secretKey string, opts ...ClientOption) (*Client, error) {
	config := ClientConfig{
		APIKey:    apiKey,
		SecretKey: secretKey,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &Client{config: config}, nil
}

// doPublicGet performs an unauthenticated GET request to a public Binance endpoint.
func (c *Client) doPublicGet(ctx context.Context, path string, params url.Values) ([]byte, error) {
	reqURL := c.config.BaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	c.config.Logger.Debug("binance request",
		slog.String("method", "GET"),
		slog.String("path", path),
		slog.Bool("signed", false),
	)

	start := c.config.Clock.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("binance: failed to create request: %w", err)
	}

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		c.config.Logger.Error("binance request failed",
			slog.String("path", path),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("binance: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("binance: failed to read response: %w", err)
	}

	latency := c.config.Clock.Now().Sub(start)

	c.config.Logger.Debug("binance response",
		slog.String("path", path),
		slog.Int("status", resp.StatusCode),
		slog.Duration("latency", latency),
	)

	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(body, resp.StatusCode, path)
	}

	return body, nil
}

// doSignedGet performs an authenticated (signed) GET request to a Binance endpoint.
func (c *Client) doSignedGet(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}

	// Add timestamp and recvWindow
	timestamp := c.config.Clock.Now().UnixMilli()
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("recvWindow", strconv.FormatInt(c.config.RecvWindow, 10))

	// Compute signature
	queryString := params.Encode()
	signature := sign(queryString, c.config.SecretKey)
	params.Set("signature", signature)

	reqURL := c.config.BaseURL + path + "?" + params.Encode()

	c.config.Logger.Debug("binance request",
		slog.String("method", "GET"),
		slog.String("path", path),
		slog.Bool("signed", true),
	)

	start := c.config.Clock.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("binance: failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		c.config.Logger.Error("binance request failed",
			slog.String("path", path),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("binance: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("binance: failed to read response: %w", err)
	}

	latency := c.config.Clock.Now().Sub(start)

	c.config.Logger.Debug("binance response",
		slog.String("path", path),
		slog.Int("status", resp.StatusCode),
		slog.Duration("latency", latency),
	)

	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(body, resp.StatusCode, path)
	}

	return body, nil
}

// parseErrorResponse attempts to parse a Binance error response body
// and returns a *BinanceError.
func (c *Client) parseErrorResponse(body []byte, statusCode int, path string) error {
	binErr := &BinanceError{
		HTTPStatus: statusCode,
	}

	// Try to unmarshal Binance error JSON {"code": ..., "msg": ...}
	if err := json.Unmarshal(body, binErr); err != nil {
		// If we can't parse it, use the raw body as the message
		binErr.Msg = string(body)
	}

	c.config.Logger.Error("binance api error",
		slog.String("path", path),
		slog.Int("http_status", statusCode),
		slog.Int("code", binErr.Code),
		slog.String("msg", binErr.Msg),
	)

	return binErr
}
