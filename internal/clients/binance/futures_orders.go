package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetOpenOrders retrieves all open orders, optionally filtered by symbol.
// Endpoint: GET /fapi/v1/openOrders (weight: 1 single symbol / 40 all symbols, signed)
func (c *FuturesClient) GetOpenOrders(ctx context.Context, symbol string) ([]FuturesOrder, error) {
	var params url.Values
	if symbol != "" {
		params = url.Values{}
		params.Set("symbol", symbol)
	}

	body, err := c.base.DoSignedGet(ctx, "/fapi/v1/openOrders", params)
	if err != nil {
		return nil, err
	}

	var orders []FuturesOrder
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("futures: failed to parse open orders response: %w", err)
	}

	return orders, nil
}
