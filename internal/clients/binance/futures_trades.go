package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// GetUserTrades retrieves trade history for a specific symbol.
// Symbol is required. Limit defaults to 500, max 1000.
// Endpoint: GET /fapi/v1/userTrades (weight: 5, signed)
func (c *FuturesClient) GetUserTrades(ctx context.Context, symbol string, limit int) ([]FuturesUserTrade, error) {
	if symbol == "" {
		return nil, fmt.Errorf("futures: symbol is required for user trades")
	}

	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := c.base.DoSignedGet(ctx, "/fapi/v1/userTrades", params)
	if err != nil {
		return nil, err
	}

	var trades []FuturesUserTrade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, fmt.Errorf("futures: failed to parse user trades response: %w", err)
	}

	return trades, nil
}

// GetIncomeHistory retrieves income/PnL history including realized PnL,
// funding fees, and commissions.
// Endpoint: GET /fapi/v1/income (weight: 30, signed)
func (c *FuturesClient) GetIncomeHistory(ctx context.Context, opts IncomeHistoryOptions) ([]IncomeRecord, error) {
	params := url.Values{}

	if opts.Symbol != "" {
		params.Set("symbol", opts.Symbol)
	}
	if opts.IncomeType != "" {
		params.Set("incomeType", opts.IncomeType)
	}
	if opts.StartTime > 0 {
		params.Set("startTime", strconv.FormatInt(opts.StartTime, 10))
	}
	if opts.EndTime > 0 {
		params.Set("endTime", strconv.FormatInt(opts.EndTime, 10))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	body, err := c.base.DoSignedGet(ctx, "/fapi/v1/income", params)
	if err != nil {
		return nil, err
	}

	var records []IncomeRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, fmt.Errorf("futures: failed to parse income history response: %w", err)
	}

	return records, nil
}
