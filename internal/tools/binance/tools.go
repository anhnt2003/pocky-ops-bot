// Package binance provides LLM tool implementations for Binance API operations.
package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bnclient "github.com/pocky-ops-bot/internal/clients/binance"
	"github.com/pocky-ops-bot/internal/clients/llm"
)

// BinanceClient is the interface for Binance API operations.
// Defined at the consumer side for testability.
type BinanceClient interface {
	GetAccount(ctx context.Context) (*bnclient.AccountResponse, error)
	GetTickerPrice(ctx context.Context, symbols []string) ([]bnclient.TickerPrice, error)
	GetTicker24hr(ctx context.Context, symbols []string) ([]bnclient.Ticker24hr, error)
}

// --- Tool 1: get_spot_balances ---

// GetBalancesTool retrieves non-zero asset balances from the Binance account.
type GetBalancesTool struct {
	client BinanceClient
	logger *slog.Logger
}

// NewGetBalancesTool creates a new GetBalancesTool.
func NewGetBalancesTool(client BinanceClient, logger *slog.Logger) *GetBalancesTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetBalancesTool{client: client, logger: logger}
}

func (t *GetBalancesTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_spot_balances",
		Description: "Get all non-zero asset balances from the Binance spot account. Returns each asset with its free (available) and locked (in orders) amounts. Use this to see what assets the user holds.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
	}
}

func (t *GetBalancesTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	t.logger.Debug("fetching spot balances")

	resp, err := t.client.GetAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get account: %w", err)
	}

	result, err := json.Marshal(resp.Balances)
	if err != nil {
		return "", fmt.Errorf("failed to marshal balances: %w", err)
	}

	t.logger.Debug("spot balances fetched", slog.Int("count", len(resp.Balances)))
	return string(result), nil
}

// --- Tool 2: get_ticker_prices ---

// GetPricesTool retrieves current prices for trading pairs.
type GetPricesTool struct {
	client BinanceClient
	logger *slog.Logger
}

// NewGetPricesTool creates a new GetPricesTool.
func NewGetPricesTool(client BinanceClient, logger *slog.Logger) *GetPricesTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetPricesTool{client: client, logger: logger}
}

func (t *GetPricesTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_ticker_prices",
		Description: "Get the current price for one or more trading pairs on Binance. Pass symbols like BTCUSDT, ETHUSDT. Use after get_spot_balances to calculate portfolio value in USDT.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbols": {
					"type": "array",
					"items": {"type": "string"},
					"description": "List of trading pair symbols, e.g. [\"BTCUSDT\", \"ETHUSDT\"]. Each must end with the quote asset (usually USDT)."
				}
			},
			"required": ["symbols"]
		}`),
	}
}

type getPricesArgs struct {
	Symbols []string `json:"symbols"`
}

func (t *GetPricesTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args getPricesArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if len(args.Symbols) == 0 {
		return "", fmt.Errorf("symbols is required and must not be empty")
	}

	// Validate each symbol is non-empty
	for i, s := range args.Symbols {
		if s == "" {
			return "", fmt.Errorf("symbols[%d] is empty", i)
		}
	}

	t.logger.Debug("fetching ticker prices", slog.Any("symbols", args.Symbols))

	prices, err := t.client.GetTickerPrice(ctx, args.Symbols)
	if err != nil {
		return "", fmt.Errorf("failed to get ticker prices: %w", err)
	}

	result, err := json.Marshal(prices)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prices: %w", err)
	}

	t.logger.Debug("ticker prices fetched", slog.Int("count", len(prices)))
	return string(result), nil
}

// --- Tool 3: get_24hr_ticker_stats ---

// Get24hrStatsTool retrieves 24-hour price change statistics.
type Get24hrStatsTool struct {
	client BinanceClient
	logger *slog.Logger
}

// NewGet24hrStatsTool creates a new Get24hrStatsTool.
func NewGet24hrStatsTool(client BinanceClient, logger *slog.Logger) *Get24hrStatsTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &Get24hrStatsTool{client: client, logger: logger}
}

func (t *Get24hrStatsTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_24hr_ticker_stats",
		Description: "Get 24-hour price change statistics for one or more trading pairs on Binance. Returns priceChangePercent, highPrice, lowPrice, volume. Use this to calculate today's profit/loss percentage.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbols": {
					"type": "array",
					"items": {"type": "string"},
					"description": "List of trading pair symbols, e.g. [\"BTCUSDT\", \"ETHUSDT\"]."
				}
			},
			"required": ["symbols"]
		}`),
	}
}

type get24hrArgs struct {
	Symbols []string `json:"symbols"`
}

func (t *Get24hrStatsTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args get24hrArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if len(args.Symbols) == 0 {
		return "", fmt.Errorf("symbols is required and must not be empty")
	}

	for i, s := range args.Symbols {
		if s == "" {
			return "", fmt.Errorf("symbols[%d] is empty", i)
		}
	}

	t.logger.Debug("fetching 24hr stats", slog.Any("symbols", args.Symbols))

	stats, err := t.client.GetTicker24hr(ctx, args.Symbols)
	if err != nil {
		return "", fmt.Errorf("failed to get 24hr stats: %w", err)
	}

	result, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal 24hr stats: %w", err)
	}

	t.logger.Debug("24hr stats fetched", slog.Int("count", len(stats)))
	return string(result), nil
}
