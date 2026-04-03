package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetTickerPrice returns the latest price for the given symbol(s).
// At least one symbol is required; pass a single symbol or multiple.
// For a single symbol, Binance returns a single object; for multiple, an array.
func (c *Client) GetTickerPrice(ctx context.Context, symbols []string) ([]TickerPrice, error) {
	params, err := buildSymbolParams(symbols)
	if err != nil {
		return nil, err
	}

	body, err := c.DoPublicGet(ctx, "/api/v3/ticker/price", params)
	if err != nil {
		return nil, err
	}

	// Single symbol: Binance returns a single JSON object, not an array.
	if len(symbols) == 1 {
		var ticker TickerPrice
		if err := json.Unmarshal(body, &ticker); err != nil {
			return nil, fmt.Errorf("binance: failed to parse ticker price response: %w", err)
		}
		return []TickerPrice{ticker}, nil
	}

	// Multiple symbols: Binance returns a JSON array.
	var tickers []TickerPrice
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("binance: failed to parse ticker price response: %w", err)
	}
	return tickers, nil
}

// GetTicker24hr returns 24hr rolling window statistics for the given symbol(s).
// At least one symbol is required; pass a single symbol or multiple.
// For a single symbol, Binance returns a single object; for multiple, an array.
func (c *Client) GetTicker24hr(ctx context.Context, symbols []string) ([]Ticker24hr, error) {
	params, err := buildSymbolParams(symbols)
	if err != nil {
		return nil, err
	}

	body, err := c.DoPublicGet(ctx, "/api/v3/ticker/24hr", params)
	if err != nil {
		return nil, err
	}

	// Single symbol: Binance returns a single JSON object, not an array.
	if len(symbols) == 1 {
		var ticker Ticker24hr
		if err := json.Unmarshal(body, &ticker); err != nil {
			return nil, fmt.Errorf("binance: failed to parse ticker 24hr response: %w", err)
		}
		return []Ticker24hr{ticker}, nil
	}

	// Multiple symbols: Binance returns a JSON array.
	var tickers []Ticker24hr
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("binance: failed to parse ticker 24hr response: %w", err)
	}
	return tickers, nil
}

// buildSymbolParams constructs the query parameters for symbol-based endpoints.
// Returns an error if no symbols are provided.
func buildSymbolParams(symbols []string) (url.Values, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("binance: at least one symbol is required")
	}

	params := url.Values{}

	if len(symbols) == 1 {
		params.Set("symbol", symbols[0])
	} else {
		// Marshal symbols as JSON array for the "symbols" param.
		symbolsJSON, err := json.Marshal(symbols)
		if err != nil {
			return nil, fmt.Errorf("binance: failed to marshal symbols: %w", err)
		}
		params.Set("symbols", string(symbolsJSON))
	}

	return params, nil
}
