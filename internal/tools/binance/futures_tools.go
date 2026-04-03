package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	bnclient "github.com/pocky-ops-bot/internal/clients/binance"
	"github.com/pocky-ops-bot/internal/clients/llm"
)

// FuturesClient is the interface for Binance Futures API operations.
// Defined at the consumer side for testability.
type FuturesClient interface {
	GetAccount(ctx context.Context) (*bnclient.FuturesAccountResponse, error)
	GetPositionRisk(ctx context.Context, symbol string) ([]bnclient.PositionRisk, error)
	GetOpenOrders(ctx context.Context, symbol string) ([]bnclient.FuturesOrder, error)
	GetUserTrades(ctx context.Context, symbol string, limit int) ([]bnclient.FuturesUserTrade, error)
	GetIncomeHistory(ctx context.Context, opts bnclient.IncomeHistoryOptions) ([]bnclient.IncomeRecord, error)
}

// --- Tool 4: get_futures_account ---

// GetFuturesAccountTool retrieves futures account summary.
type GetFuturesAccountTool struct {
	client FuturesClient
	logger *slog.Logger
}

// NewGetFuturesAccountTool creates a new GetFuturesAccountTool.
func NewGetFuturesAccountTool(client FuturesClient, logger *slog.Logger) *GetFuturesAccountTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetFuturesAccountTool{client: client, logger: logger}
}

func (t *GetFuturesAccountTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_futures_account",
		Description: "Get Binance USD-M Futures account summary including total wallet balance, unrealized profit, margin balance, and available balance. Use this to see the user's futures account overview.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
	}
}

func (t *GetFuturesAccountTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	t.logger.Debug("fetching futures account")

	resp, err := t.client.GetAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get futures account: %w", err)
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal futures account: %w", err)
	}

	t.logger.Debug("futures account fetched", slog.Int("assets", len(resp.Assets)))
	return string(result), nil
}

// --- Tool 5: get_futures_positions ---

// GetFuturesPositionsTool retrieves open futures positions.
type GetFuturesPositionsTool struct {
	client FuturesClient
	logger *slog.Logger
}

// NewGetFuturesPositionsTool creates a new GetFuturesPositionsTool.
func NewGetFuturesPositionsTool(client FuturesClient, logger *slog.Logger) *GetFuturesPositionsTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetFuturesPositionsTool{client: client, logger: logger}
}

func (t *GetFuturesPositionsTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_futures_positions",
		Description: "Get open USD-M Futures positions with entry price, mark price, unrealized P&L, leverage, and liquidation price. Omit symbol to get all open positions. Only returns positions with non-zero amounts.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbol": {
					"type": "string",
					"description": "Trading pair symbol, e.g. \"BTCUSDT\". Optional — omit to get all open positions."
				}
			}
		}`),
	}
}

type getPositionsArgs struct {
	Symbol string `json:"symbol"`
}

func (t *GetFuturesPositionsTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args getPositionsArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	t.logger.Debug("fetching futures positions", slog.String("symbol", args.Symbol))

	positions, err := t.client.GetPositionRisk(ctx, args.Symbol)
	if err != nil {
		return "", fmt.Errorf("failed to get futures positions: %w", err)
	}

	// Filter out zero-amount positions (Binance returns all configured symbols).
	open := make([]bnclient.PositionRisk, 0, len(positions))
	for _, p := range positions {
		if p.PositionAmt != "0" && p.PositionAmt != "0.000" && p.PositionAmt != "0.00000000" {
			open = append(open, p)
		}
	}

	result, err := json.Marshal(open)
	if err != nil {
		return "", fmt.Errorf("failed to marshal positions: %w", err)
	}

	t.logger.Debug("futures positions fetched",
		slog.Int("total", len(positions)),
		slog.Int("open", len(open)),
	)
	return string(result), nil
}

// --- Tool 6: get_futures_open_orders ---

// GetFuturesOpenOrdersTool retrieves open futures orders.
type GetFuturesOpenOrdersTool struct {
	client FuturesClient
	logger *slog.Logger
}

// NewGetFuturesOpenOrdersTool creates a new GetFuturesOpenOrdersTool.
func NewGetFuturesOpenOrdersTool(client FuturesClient, logger *slog.Logger) *GetFuturesOpenOrdersTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetFuturesOpenOrdersTool{client: client, logger: logger}
}

func (t *GetFuturesOpenOrdersTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_futures_open_orders",
		Description: "Get open USD-M Futures orders including order type, side, price, quantity, and stop price. Optionally filter by symbol.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbol": {
					"type": "string",
					"description": "Trading pair symbol, e.g. \"BTCUSDT\". Optional — omit to get all open orders (higher API weight)."
				}
			}
		}`),
	}
}

type getOpenOrdersArgs struct {
	Symbol string `json:"symbol"`
}

func (t *GetFuturesOpenOrdersTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args getOpenOrdersArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	t.logger.Debug("fetching futures open orders", slog.String("symbol", args.Symbol))

	orders, err := t.client.GetOpenOrders(ctx, args.Symbol)
	if err != nil {
		return "", fmt.Errorf("failed to get futures open orders: %w", err)
	}

	result, err := json.Marshal(orders)
	if err != nil {
		return "", fmt.Errorf("failed to marshal open orders: %w", err)
	}

	t.logger.Debug("futures open orders fetched", slog.Int("count", len(orders)))
	return string(result), nil
}

// --- Tool 7: get_futures_trades ---

// GetFuturesTradesTool retrieves recent futures trade history.
type GetFuturesTradesTool struct {
	client FuturesClient
	logger *slog.Logger
}

// NewGetFuturesTradesTool creates a new GetFuturesTradesTool.
func NewGetFuturesTradesTool(client FuturesClient, logger *slog.Logger) *GetFuturesTradesTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetFuturesTradesTool{client: client, logger: logger}
}

func (t *GetFuturesTradesTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_futures_trades",
		Description: "Get recent USD-M Futures trade history for a specific symbol. Returns realized P&L and commissions per trade. Symbol is required.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbol": {
					"type": "string",
					"description": "Trading pair symbol, e.g. \"BTCUSDT\". Required."
				},
				"limit": {
					"type": "integer",
					"description": "Number of trades to return. Default 20, max 1000."
				}
			},
			"required": ["symbol"]
		}`),
	}
}

type getTradesArgs struct {
	Symbol string `json:"symbol"`
	Limit  int    `json:"limit"`
}

func (t *GetFuturesTradesTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args getTradesArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 20
	}

	t.logger.Debug("fetching futures trades",
		slog.String("symbol", args.Symbol),
		slog.Int("limit", limit),
	)

	trades, err := t.client.GetUserTrades(ctx, args.Symbol, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get futures trades: %w", err)
	}

	result, err := json.Marshal(trades)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trades: %w", err)
	}

	t.logger.Debug("futures trades fetched", slog.Int("count", len(trades)))
	return string(result), nil
}

// --- Tool 8: get_futures_income ---

// GetFuturesIncomeTool retrieves futures income/PnL history.
type GetFuturesIncomeTool struct {
	client FuturesClient
	logger *slog.Logger
}

// NewGetFuturesIncomeTool creates a new GetFuturesIncomeTool.
func NewGetFuturesIncomeTool(client FuturesClient, logger *slog.Logger) *GetFuturesIncomeTool {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetFuturesIncomeTool{client: client, logger: logger}
}

func (t *GetFuturesIncomeTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        "get_futures_income",
		Description: "Get USD-M Futures income history including realized PnL, funding fees, and commissions. Filter by symbol and/or income type. Without filters, returns last 7 days of all income types.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"symbol": {
					"type": "string",
					"description": "Trading pair symbol, e.g. \"BTCUSDT\". Optional."
				},
				"income_type": {
					"type": "string",
					"enum": ["REALIZED_PNL", "FUNDING_FEE", "COMMISSION", "TRANSFER"],
					"description": "Filter by income type. Optional — omit to get all types."
				},
				"limit": {
					"type": "integer",
					"description": "Number of records to return. Default 50, max 1000."
				}
			}
		}`),
	}
}

type getIncomeArgs struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"income_type"`
	Limit      int    `json:"limit"`
}

func (t *GetFuturesIncomeTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args getIncomeArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}

	t.logger.Debug("fetching futures income",
		slog.String("symbol", args.Symbol),
		slog.String("income_type", args.IncomeType),
		slog.Int("limit", limit),
	)

	records, err := t.client.GetIncomeHistory(ctx, bnclient.IncomeHistoryOptions{
		Symbol:     args.Symbol,
		IncomeType: args.IncomeType,
		Limit:      limit,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get futures income: %w", err)
	}

	result, err := json.Marshal(records)
	if err != nil {
		return "", fmt.Errorf("failed to marshal income records: %w", err)
	}

	t.logger.Debug("futures income fetched", slog.Int("count", len(records)))
	return string(result), nil
}
