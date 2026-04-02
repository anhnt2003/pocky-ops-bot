package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetAccount retrieves the current account information, including balances.
// This is a signed endpoint requiring API key and secret.
func (c *Client) GetAccount(ctx context.Context) (*AccountResponse, error) {
	params := url.Values{}
	params.Set("omitZeroBalances", "true")

	body, err := c.doSignedGet(ctx, "/api/v3/account", params)
	if err != nil {
		return nil, err
	}

	var resp AccountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("binance: failed to parse account response: %w", err)
	}

	return &resp, nil
}
