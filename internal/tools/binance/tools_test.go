package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	bnclient "github.com/pocky-ops-bot/internal/clients/binance"
)

// mockBinanceClient implements BinanceClient for testing.
type mockBinanceClient struct {
	account  *bnclient.AccountResponse
	prices   []bnclient.TickerPrice
	stats24h []bnclient.Ticker24hr
	err      error
}

func (m *mockBinanceClient) GetAccount(ctx context.Context) (*bnclient.AccountResponse, error) {
	return m.account, m.err
}

func (m *mockBinanceClient) GetTickerPrice(ctx context.Context, symbols []string) ([]bnclient.TickerPrice, error) {
	return m.prices, m.err
}

func (m *mockBinanceClient) GetTicker24hr(ctx context.Context, symbols []string) ([]bnclient.Ticker24hr, error) {
	return m.stats24h, m.err
}

// --- GetBalancesTool Tests ---

func TestGetBalancesTool_Definition(t *testing.T) {
	tool := NewGetBalancesTool(&mockBinanceClient{}, nil)
	def := tool.Definition()

	if def.Name != "get_spot_balances" {
		t.Errorf("Name = %q, want %q", def.Name, "get_spot_balances")
	}
	if def.Description == "" {
		t.Error("Description is empty")
	}
}

func TestGetBalancesTool_Execute_Success(t *testing.T) {
	client := &mockBinanceClient{
		account: &bnclient.AccountResponse{
			Balances: []bnclient.Balance{
				{Asset: "BTC", Free: "0.5", Locked: "0.1"},
				{Asset: "USDT", Free: "1000.00", Locked: "0.00"},
			},
		},
	}
	tool := NewGetBalancesTool(client, nil)

	result, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var balances []bnclient.Balance
	if err := json.Unmarshal([]byte(result), &balances); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(balances) != 2 {
		t.Fatalf("expected 2 balances, got %d", len(balances))
	}
	if balances[0].Asset != "BTC" {
		t.Errorf("balances[0].Asset = %q, want %q", balances[0].Asset, "BTC")
	}
}

func TestGetBalancesTool_Execute_Error(t *testing.T) {
	client := &mockBinanceClient{err: fmt.Errorf("connection refused")}
	tool := NewGetBalancesTool(client, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPricesTool Tests ---

func TestGetPricesTool_Definition(t *testing.T) {
	tool := NewGetPricesTool(&mockBinanceClient{}, nil)
	def := tool.Definition()

	if def.Name != "get_ticker_prices" {
		t.Errorf("Name = %q, want %q", def.Name, "get_ticker_prices")
	}

	// Verify schema has required field
	var schema map[string]interface{}
	if err := json.Unmarshal(def.Parameters, &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	required, ok := schema["required"].([]interface{})
	if !ok || len(required) == 0 {
		t.Error("schema missing required field")
	}
}

func TestGetPricesTool_Execute_Success(t *testing.T) {
	client := &mockBinanceClient{
		prices: []bnclient.TickerPrice{
			{Symbol: "BTCUSDT", Price: "100000.50"},
			{Symbol: "ETHUSDT", Price: "3500.25"},
		},
	}
	tool := NewGetPricesTool(client, nil)

	result, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":["BTCUSDT","ETHUSDT"]}`))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var prices []bnclient.TickerPrice
	if err := json.Unmarshal([]byte(result), &prices); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(prices) != 2 {
		t.Fatalf("expected 2 prices, got %d", len(prices))
	}
}

func TestGetPricesTool_Execute_EmptySymbols(t *testing.T) {
	tool := NewGetPricesTool(&mockBinanceClient{}, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":[]}`))
	if err == nil {
		t.Fatal("expected error for empty symbols")
	}
}

func TestGetPricesTool_Execute_InvalidArgs(t *testing.T) {
	tool := NewGetPricesTool(&mockBinanceClient{}, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetPricesTool_Execute_EmptySymbolString(t *testing.T) {
	tool := NewGetPricesTool(&mockBinanceClient{}, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":["BTCUSDT",""]}`))
	if err == nil {
		t.Fatal("expected error for empty symbol string")
	}
}

func TestGetPricesTool_Execute_APIError(t *testing.T) {
	client := &mockBinanceClient{err: fmt.Errorf("rate limited")}
	tool := NewGetPricesTool(client, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":["BTCUSDT"]}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Get24hrStatsTool Tests ---

func TestGet24hrStatsTool_Definition(t *testing.T) {
	tool := NewGet24hrStatsTool(&mockBinanceClient{}, nil)
	def := tool.Definition()

	if def.Name != "get_24hr_ticker_stats" {
		t.Errorf("Name = %q, want %q", def.Name, "get_24hr_ticker_stats")
	}
}

func TestGet24hrStatsTool_Execute_Success(t *testing.T) {
	client := &mockBinanceClient{
		stats24h: []bnclient.Ticker24hr{
			{
				Symbol:             "BTCUSDT",
				PriceChangePercent: "2.50",
				LastPrice:          "100000.50",
				HighPrice:          "101000.00",
				LowPrice:           "98000.00",
				Volume:             "15000.00",
			},
		},
	}
	tool := NewGet24hrStatsTool(client, nil)

	result, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":["BTCUSDT"]}`))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var stats []bnclient.Ticker24hr
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].PriceChangePercent != "2.50" {
		t.Errorf("PriceChangePercent = %q, want %q", stats[0].PriceChangePercent, "2.50")
	}
}

func TestGet24hrStatsTool_Execute_EmptySymbols(t *testing.T) {
	tool := NewGet24hrStatsTool(&mockBinanceClient{}, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":[]}`))
	if err == nil {
		t.Fatal("expected error for empty symbols")
	}
}

func TestGet24hrStatsTool_Execute_APIError(t *testing.T) {
	client := &mockBinanceClient{err: fmt.Errorf("timeout")}
	tool := NewGet24hrStatsTool(client, nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"symbols":["BTCUSDT"]}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
