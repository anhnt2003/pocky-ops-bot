package ai

import "fmt"

// AIError represents an error from an AI API provider.
type AIError struct {
	Code        int
	Description string
	RetryAfter  int
	Provider    Provider
}

// Error implements the error interface for AIError.
func (e *AIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s ai error %d: %s (retry after %ds)", e.Provider, e.Code, e.Description, e.RetryAfter)
	}
	return fmt.Sprintf("%s ai error %d: %s", e.Provider, e.Code, e.Description)
}

// IsRetryable returns true if the error is transient and can be retried.
func (e *AIError) IsRetryable() bool {
	// Rate limiting or server errors are retryable
	return e.RetryAfter > 0 || e.Code == 429 || e.Code >= 500
}
