package binance

import "fmt"

// BinanceError represents an error response from the Binance API.
type BinanceError struct {
	// HTTPStatus is the HTTP status code of the response.
	HTTPStatus int
	// Code is the Binance-specific error code.
	Code int `json:"code"`
	// Msg is the Binance error message.
	Msg string `json:"msg"`
}

// Error implements the error interface for BinanceError.
func (e *BinanceError) Error() string {
	return fmt.Sprintf("binance: [HTTP %d] code=%d: %s", e.HTTPStatus, e.Code, e.Msg)
}

// retryableBinanceCodes contains Binance-specific error codes that are retryable.
var retryableBinanceCodes = map[int]bool{
	-1000: true, // UNKNOWN
	-1001: true, // DISCONNECTED
	-1003: true, // TOO_MANY_REQUESTS
	-1006: true, // UNEXPECTED_RESP
	-1007: true, // TIMEOUT
	-1008: true, // SERVER_BUSY
	-1015: true, // TOO_MANY_ORDERS
	-1034: true, // UNKNOWN_ORDER_COMPOSITION
}

// IsRetryable returns true if the error is transient and can be retried.
// HTTP 429 (rate limit) and 5xx (server errors) are retryable,
// as well as specific Binance error codes.
func (e *BinanceError) IsRetryable() bool {
	if e.HTTPStatus == 429 || e.HTTPStatus >= 500 {
		return true
	}
	return retryableBinanceCodes[e.Code]
}
