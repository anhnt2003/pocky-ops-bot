// Package types provides type definitions for the Telegram Bot API client.
package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIResponse represents the generic response from Telegram Bot API.
type APIResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
	Parameters  *ResponseParams `json:"parameters,omitempty"`
}

// ResponseParams contains additional parameters in error responses.
type ResponseParams struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

// APIError represents an error from the Telegram Bot API.
type APIError struct {
	Code        int
	Description string
	RetryAfter  int
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("telegram api error %d: %s (retry after %ds)", e.Code, e.Description, e.RetryAfter)
	}
	return fmt.Sprintf("telegram api error %d: %s", e.Code, e.Description)
}

// IsRetryable returns true if the error is transient and can be retried.
func (e *APIError) IsRetryable() bool {
	// Rate limiting or server errors are retryable
	return e.RetryAfter > 0 || e.Code >= 500
}

// BackoffStrategy defines the interface for retry backoff calculation.
type BackoffStrategy interface {
	// NextBackoff returns the duration to wait before the next retry.
	// attempt is 0-indexed (first retry is attempt 0).
	NextBackoff(attempt int) time.Duration
	// Reset resets the backoff state.
	Reset()
}

// HTTPClient defines the interface for HTTP operations.
// This allows for easy mocking in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
