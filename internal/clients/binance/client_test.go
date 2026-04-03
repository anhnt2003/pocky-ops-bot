package binance

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fixedClock is a test Clock that returns a fixed time.
type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

// fixedTime is the fixed time used in tests: 2024-01-15 00:00:00 UTC.
var fixedTime = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		secretKey string
		opts      []ClientOption
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid credentials",
			apiKey:    "test-api-key",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "missing api key",
			apiKey:    "",
			secretKey: "test-secret-key",
			wantErr:   true,
			errMsg:    "api key is required",
		},
		{
			name:      "missing secret key",
			apiKey:    "test-api-key",
			secretKey: "",
			wantErr:   true,
			errMsg:    "secret key is required",
		},
		{
			name:      "with options",
			apiKey:    "test-api-key",
			secretKey: "test-secret-key",
			opts: []ClientOption{
				WithBaseURL("https://testnet.binance.vision"),
				WithTimeout(30 * time.Second),
				WithRecvWindow(10000),
				WithLogger(slog.Default()),
				WithClock(fixedClock{t: fixedTime}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, tt.secretKey, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if client == nil {
				t.Error("NewClient() returned nil without error")
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	client, err := NewClient("api-key", "secret-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.config.BaseURL != "https://api.binance.com" {
		t.Errorf("BaseURL = %q, want %q", client.config.BaseURL, "https://api.binance.com")
	}
	if client.config.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", client.config.Timeout, 10*time.Second)
	}
	if client.config.RecvWindow != 5000 {
		t.Errorf("RecvWindow = %d, want 5000", client.config.RecvWindow)
	}
	if client.config.Clock == nil {
		t.Error("Clock should not be nil")
	}
	if client.config.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
	if client.config.Logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	logger := slog.Default()
	clock := fixedClock{t: fixedTime}

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL("https://custom.url"),
		WithTimeout(30*time.Second),
		WithRecvWindow(10000),
		WithLogger(logger),
		WithClock(clock),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.config.BaseURL != "https://custom.url" {
		t.Errorf("BaseURL = %q, want %q", client.config.BaseURL, "https://custom.url")
	}
	if client.config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", client.config.Timeout, 30*time.Second)
	}
	if client.config.RecvWindow != 10000 {
		t.Errorf("RecvWindow = %d, want 10000", client.config.RecvWindow)
	}
}

func TestDoSignedGet_AddsAuthHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify X-MBX-APIKEY header is set.
		apiKey := r.Header.Get("X-MBX-APIKEY")
		if apiKey != "test-api-key" {
			t.Errorf("X-MBX-APIKEY = %q, want %q", apiKey, "test-api-key")
		}

		// Verify query params include timestamp, recvWindow, signature.
		q := r.URL.Query()
		if q.Get("timestamp") == "" {
			t.Error("missing timestamp parameter")
		}
		if q.Get("recvWindow") == "" {
			t.Error("missing recvWindow parameter")
		}
		if q.Get("signature") == "" {
			t.Error("missing signature parameter")
		}

		// Verify timestamp matches our fixed clock.
		expectedTS := "1705276800000" // fixedTime.UnixMilli()
		if q.Get("timestamp") != expectedTS {
			t.Errorf("timestamp = %q, want %q", q.Get("timestamp"), expectedTS)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", "test-secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	body, err := client.DoSignedGet(context.Background(), "/api/v3/account", nil)
	if err != nil {
		t.Fatalf("DoSignedGet() error = %v", err)
	}

	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
}

func TestDoPublicGet_NoAuthHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no auth headers are present on public endpoints.
		if apiKey := r.Header.Get("X-MBX-APIKEY"); apiKey != "" {
			t.Errorf("X-MBX-APIKEY should be empty for public endpoints, got %q", apiKey)
		}

		// Verify no signature-related params.
		q := r.URL.Query()
		if q.Get("signature") != "" {
			t.Error("signature should not be present for public endpoints")
		}
		if q.Get("timestamp") != "" {
			t.Error("timestamp should not be present for public endpoints")
		}

		// Verify the symbol param is passed through.
		if q.Get("symbol") != "BTCUSDT" {
			t.Errorf("symbol = %q, want %q", q.Get("symbol"), "BTCUSDT")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"symbol":"BTCUSDT","price":"50000.00"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", "test-secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	params := make(map[string][]string)
	params["symbol"] = []string{"BTCUSDT"}

	body, err := client.DoPublicGet(context.Background(), "/api/v3/ticker/price", params)
	if err != nil {
		t.Fatalf("DoPublicGet() error = %v", err)
	}

	if !strings.Contains(string(body), "BTCUSDT") {
		t.Errorf("body = %q, expected to contain BTCUSDT", string(body))
	}
}

func TestDoRequest_ErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   map[string]interface{}
		wantCode       int
		wantMsg        string
		wantRetryable  bool
		wantHTTPStatus int
	}{
		{
			name:       "bad request with binance error",
			statusCode: http.StatusBadRequest,
			responseBody: map[string]interface{}{
				"code": -1100,
				"msg":  "Illegal characters found in parameter.",
			},
			wantCode:       -1100,
			wantMsg:        "Illegal characters found in parameter.",
			wantRetryable:  false,
			wantHTTPStatus: 400,
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			responseBody: map[string]interface{}{
				"code": -1003,
				"msg":  "Too many requests.",
			},
			wantCode:       -1003,
			wantMsg:        "Too many requests.",
			wantRetryable:  true,
			wantHTTPStatus: 429,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			responseBody: map[string]interface{}{
				"code": -1001,
				"msg":  "Internal error.",
			},
			wantCode:       -1001,
			wantMsg:        "Internal error.",
			wantRetryable:  true,
			wantHTTPStatus: 500,
		},
		{
			name:       "non-retryable binance code",
			statusCode: http.StatusBadRequest,
			responseBody: map[string]interface{}{
				"code": -1121,
				"msg":  "Invalid symbol.",
			},
			wantCode:       -1121,
			wantMsg:        "Invalid symbol.",
			wantRetryable:  false,
			wantHTTPStatus: 400,
		},
		{
			name:       "retryable binance code with 4xx status",
			statusCode: http.StatusForbidden,
			responseBody: map[string]interface{}{
				"code": -1015,
				"msg":  "Too many new orders.",
			},
			wantCode:       -1015,
			wantMsg:        "Too many new orders.",
			wantRetryable:  true,
			wantHTTPStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, err := NewClient("test-api-key", "test-secret-key",
				WithBaseURL(server.URL),
				WithClock(fixedClock{t: fixedTime}),
			)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			_, err = client.DoPublicGet(context.Background(), "/api/v3/ticker/price", nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			binErr, ok := err.(*BinanceError)
			if !ok {
				t.Fatalf("expected *BinanceError, got %T: %v", err, err)
			}

			if binErr.HTTPStatus != tt.wantHTTPStatus {
				t.Errorf("HTTPStatus = %d, want %d", binErr.HTTPStatus, tt.wantHTTPStatus)
			}
			if binErr.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", binErr.Code, tt.wantCode)
			}
			if binErr.Msg != tt.wantMsg {
				t.Errorf("Msg = %q, want %q", binErr.Msg, tt.wantMsg)
			}
			if binErr.IsRetryable() != tt.wantRetryable {
				t.Errorf("IsRetryable() = %v, want %v", binErr.IsRetryable(), tt.wantRetryable)
			}
		})
	}
}

func TestDoRequest_NonJSONErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", "test-secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.DoPublicGet(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	binErr, ok := err.(*BinanceError)
	if !ok {
		t.Fatalf("expected *BinanceError, got %T: %v", err, err)
	}

	if binErr.HTTPStatus != 502 {
		t.Errorf("HTTPStatus = %d, want 502", binErr.HTTPStatus)
	}
	// When JSON parsing fails, the raw body should be stored in Msg.
	if !strings.Contains(binErr.Msg, "Bad Gateway") {
		t.Errorf("Msg = %q, want to contain 'Bad Gateway'", binErr.Msg)
	}
}

func TestBinanceError_Error(t *testing.T) {
	err := &BinanceError{
		HTTPStatus: 400,
		Code:       -1100,
		Msg:        "Illegal characters found.",
	}

	expected := "binance: [HTTP 400] code=-1100: Illegal characters found."
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestBinanceError_IsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		err        *BinanceError
		retryable  bool
	}{
		{
			name:      "HTTP 429",
			err:       &BinanceError{HTTPStatus: 429, Code: -1003},
			retryable: true,
		},
		{
			name:      "HTTP 500",
			err:       &BinanceError{HTTPStatus: 500, Code: -1001},
			retryable: true,
		},
		{
			name:      "HTTP 503",
			err:       &BinanceError{HTTPStatus: 503, Code: 0},
			retryable: true,
		},
		{
			name:      "binance code -1000",
			err:       &BinanceError{HTTPStatus: 400, Code: -1000},
			retryable: true,
		},
		{
			name:      "binance code -1006",
			err:       &BinanceError{HTTPStatus: 400, Code: -1006},
			retryable: true,
		},
		{
			name:      "binance code -1007",
			err:       &BinanceError{HTTPStatus: 400, Code: -1007},
			retryable: true,
		},
		{
			name:      "binance code -1008",
			err:       &BinanceError{HTTPStatus: 400, Code: -1008},
			retryable: true,
		},
		{
			name:      "binance code -1034",
			err:       &BinanceError{HTTPStatus: 400, Code: -1034},
			retryable: true,
		},
		{
			name:      "HTTP 400 non-retryable code",
			err:       &BinanceError{HTTPStatus: 400, Code: -1121},
			retryable: false,
		},
		{
			name:      "HTTP 401",
			err:       &BinanceError{HTTPStatus: 401, Code: -2014},
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRetryable(); got != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.retryable)
			}
		})
	}
}
