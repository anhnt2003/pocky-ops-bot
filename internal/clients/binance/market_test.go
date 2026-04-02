package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTickerPrice_SingleSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path.
		if r.URL.Path != "/api/v3/ticker/price" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v3/ticker/price")
		}

		// Single symbol: should use "symbol" param (singular).
		q := r.URL.Query()
		if q.Get("symbol") != "BTCUSDT" {
			t.Errorf("symbol = %q, want %q", q.Get("symbol"), "BTCUSDT")
		}
		if q.Get("symbols") != "" {
			t.Error("symbols param should not be set for single symbol")
		}

		// No auth headers for public endpoint.
		if r.Header.Get("X-MBX-APIKEY") != "" {
			t.Error("X-MBX-APIKEY should not be set for public endpoint")
		}

		// Single symbol: Binance returns a single JSON object.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TickerPrice{
			Symbol: "BTCUSDT",
			Price:  "50000.00000000",
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tickers, err := client.GetTickerPrice(context.Background(), []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("GetTickerPrice() error = %v", err)
	}

	if len(tickers) != 1 {
		t.Fatalf("len(tickers) = %d, want 1", len(tickers))
	}
	if tickers[0].Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %q, want %q", tickers[0].Symbol, "BTCUSDT")
	}
	if tickers[0].Price != "50000.00000000" {
		t.Errorf("Price = %q, want %q", tickers[0].Price, "50000.00000000")
	}
}

func TestGetTickerPrice_MultipleSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Multiple symbols: should use "symbols" param (plural).
		q := r.URL.Query()
		if q.Get("symbol") != "" {
			t.Error("symbol param should not be set for multiple symbols")
		}

		symbolsParam := q.Get("symbols")
		if symbolsParam == "" {
			t.Fatal("symbols param is empty")
		}

		// Verify the JSON array format.
		var symbols []string
		if err := json.Unmarshal([]byte(symbolsParam), &symbols); err != nil {
			t.Fatalf("failed to parse symbols param: %v", err)
		}
		if len(symbols) != 2 {
			t.Fatalf("len(symbols) = %d, want 2", len(symbols))
		}

		// Multiple symbols: Binance returns a JSON array.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]TickerPrice{
			{Symbol: "BTCUSDT", Price: "50000.00000000"},
			{Symbol: "ETHUSDT", Price: "3000.00000000"},
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tickers, err := client.GetTickerPrice(context.Background(), []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("GetTickerPrice() error = %v", err)
	}

	if len(tickers) != 2 {
		t.Fatalf("len(tickers) = %d, want 2", len(tickers))
	}
	if tickers[0].Symbol != "BTCUSDT" {
		t.Errorf("tickers[0].Symbol = %q, want %q", tickers[0].Symbol, "BTCUSDT")
	}
	if tickers[1].Symbol != "ETHUSDT" {
		t.Errorf("tickers[1].Symbol = %q, want %q", tickers[1].Symbol, "ETHUSDT")
	}
}

func TestGetTickerPrice_EmptySymbols(t *testing.T) {
	client, err := NewClient("api-key", "secret-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetTickerPrice(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error for empty symbols, got nil")
	}
	if !strings.Contains(err.Error(), "at least one symbol is required") {
		t.Errorf("error = %q, want to contain 'at least one symbol is required'", err.Error())
	}
}

func TestGetTickerPrice_NilSymbols(t *testing.T) {
	client, err := NewClient("api-key", "secret-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetTickerPrice(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil symbols, got nil")
	}
	if !strings.Contains(err.Error(), "at least one symbol is required") {
		t.Errorf("error = %q, want to contain 'at least one symbol is required'", err.Error())
	}
}

func TestGetTickerPrice_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": -1121,
			"msg":  "Invalid symbol.",
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetTickerPrice(context.Background(), []string{"INVALID"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	binErr, ok := err.(*BinanceError)
	if !ok {
		t.Fatalf("expected *BinanceError, got %T: %v", err, err)
	}
	if binErr.Code != -1121 {
		t.Errorf("Code = %d, want -1121", binErr.Code)
	}
}

func TestGetTicker24hr_SingleSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path.
		if r.URL.Path != "/api/v3/ticker/24hr" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v3/ticker/24hr")
		}

		q := r.URL.Query()
		if q.Get("symbol") != "BTCUSDT" {
			t.Errorf("symbol = %q, want %q", q.Get("symbol"), "BTCUSDT")
		}

		// Single symbol: returns a single object.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Ticker24hr{
			Symbol:             "BTCUSDT",
			PriceChange:        "500.00000000",
			PriceChangePercent: "1.010",
			WeightedAvgPrice:   "49750.00000000",
			LastPrice:          "50000.00000000",
			HighPrice:          "50500.00000000",
			LowPrice:           "49000.00000000",
			Volume:             "12345.67890000",
			QuoteVolume:        "617283945.00000000",
			OpenTime:           1705190400000,
			CloseTime:          1705276799999,
			Count:              98765,
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tickers, err := client.GetTicker24hr(context.Background(), []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("GetTicker24hr() error = %v", err)
	}

	if len(tickers) != 1 {
		t.Fatalf("len(tickers) = %d, want 1", len(tickers))
	}
	if tickers[0].Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %q, want %q", tickers[0].Symbol, "BTCUSDT")
	}
	if tickers[0].LastPrice != "50000.00000000" {
		t.Errorf("LastPrice = %q, want %q", tickers[0].LastPrice, "50000.00000000")
	}
	if tickers[0].PriceChangePercent != "1.010" {
		t.Errorf("PriceChangePercent = %q, want %q", tickers[0].PriceChangePercent, "1.010")
	}
	if tickers[0].Volume != "12345.67890000" {
		t.Errorf("Volume = %q, want %q", tickers[0].Volume, "12345.67890000")
	}
	if tickers[0].Count != 98765 {
		t.Errorf("Count = %d, want 98765", tickers[0].Count)
	}
}

func TestGetTicker24hr_MultipleSymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("symbol") != "" {
			t.Error("symbol param should not be set for multiple symbols")
		}

		symbolsParam := q.Get("symbols")
		if symbolsParam == "" {
			t.Fatal("symbols param is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Ticker24hr{
			{Symbol: "BTCUSDT", LastPrice: "50000.00"},
			{Symbol: "ETHUSDT", LastPrice: "3000.00"},
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tickers, err := client.GetTicker24hr(context.Background(), []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("GetTicker24hr() error = %v", err)
	}

	if len(tickers) != 2 {
		t.Fatalf("len(tickers) = %d, want 2", len(tickers))
	}
	if tickers[0].Symbol != "BTCUSDT" {
		t.Errorf("tickers[0].Symbol = %q, want %q", tickers[0].Symbol, "BTCUSDT")
	}
	if tickers[1].Symbol != "ETHUSDT" {
		t.Errorf("tickers[1].Symbol = %q, want %q", tickers[1].Symbol, "ETHUSDT")
	}
}

func TestGetTicker24hr_EmptySymbols(t *testing.T) {
	client, err := NewClient("api-key", "secret-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetTicker24hr(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error for empty symbols, got nil")
	}
	if !strings.Contains(err.Error(), "at least one symbol is required") {
		t.Errorf("error = %q, want to contain 'at least one symbol is required'", err.Error())
	}
}

func TestGetTicker24hr_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": -1121,
			"msg":  "Invalid symbol.",
		})
	}))
	defer server.Close()

	client, err := NewClient("api-key", "secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetTicker24hr(context.Background(), []string{"INVALID"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	binErr, ok := err.(*BinanceError)
	if !ok {
		t.Fatalf("expected *BinanceError, got %T: %v", err, err)
	}
	if binErr.Code != -1121 {
		t.Errorf("Code = %d, want -1121", binErr.Code)
	}
}
