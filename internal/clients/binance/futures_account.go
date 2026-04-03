package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetAccount retrieves futures account information including balances and margin.
// Endpoint: GET /fapi/v3/account (weight: 5, signed)
func (c *FuturesClient) GetAccount(ctx context.Context) (*FuturesAccountResponse, error) {
	body, err := c.base.DoSignedGet(ctx, "/fapi/v3/account", nil)
	if err != nil {
		return nil, err
	}

	var resp FuturesAccountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("futures: failed to parse account response: %w", err)
	}

	return &resp, nil
}

// GetPositionRisk retrieves position information.
// If symbol is empty, returns all positions (including zero-amount).
// Endpoint: GET /fapi/v3/positionRisk (weight: 5, signed)
func (c *FuturesClient) GetPositionRisk(ctx context.Context, symbol string) ([]PositionRisk, error) {
	var params url.Values
	if symbol != "" {
		params = url.Values{}
		params.Set("symbol", symbol)
	}

	body, err := c.base.DoSignedGet(ctx, "/fapi/v3/positionRisk", params)
	if err != nil {
		return nil, err
	}

	var positions []PositionRisk
	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, fmt.Errorf("futures: failed to parse position risk response: %w", err)
	}

	return positions, nil
}
