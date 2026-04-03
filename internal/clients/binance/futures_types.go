package binance

// FuturesAccountResponse represents GET /fapi/v3/account.
type FuturesAccountResponse struct {
	TotalWalletBalance      string         `json:"totalWalletBalance"`
	TotalUnrealizedProfit   string         `json:"totalUnrealizedProfit"`
	TotalMarginBalance      string         `json:"totalMarginBalance"`
	TotalInitialMargin      string         `json:"totalInitialMargin"`
	TotalMaintMargin        string         `json:"totalMaintMargin"`
	AvailableBalance        string         `json:"availableBalance"`
	TotalCrossWalletBalance string         `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl         string         `json:"totalCrossUnPnl"`
	MaxWithdrawAmount       string         `json:"maxWithdrawAmount"`
	Assets                  []FuturesAsset `json:"assets"`
}

// FuturesAsset represents a single asset in the futures account.
type FuturesAsset struct {
	Asset              string `json:"asset"`
	WalletBalance      string `json:"walletBalance"`
	UnrealizedProfit   string `json:"unrealizedProfit"`
	MarginBalance      string `json:"marginBalance"`
	AvailableBalance   string `json:"availableBalance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	InitialMargin      string `json:"initialMargin"`
	MaintMargin        string `json:"maintMargin"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
}

// PositionRisk represents GET /fapi/v3/positionRisk.
type PositionRisk struct {
	Symbol           string `json:"symbol"`
	PositionAmt      string `json:"positionAmt"`
	EntryPrice       string `json:"entryPrice"`
	MarkPrice        string `json:"markPrice"`
	UnRealizedProfit string `json:"unRealizedProfit"`
	LiquidationPrice string `json:"liquidationPrice"`
	Leverage         string `json:"leverage"`
	MarginType       string `json:"marginType"`
	PositionSide     string `json:"positionSide"`
	Notional         string `json:"notional"`
	BreakEvenPrice   string `json:"breakEvenPrice"`
	IsolatedMargin   string `json:"isolatedMargin"`
	UpdateTime       int64  `json:"updateTime"`
}

// FuturesOrder represents GET /fapi/v1/openOrders.
type FuturesOrder struct {
	OrderID      int64  `json:"orderId"`
	Symbol       string `json:"symbol"`
	Side         string `json:"side"`
	PositionSide string `json:"positionSide"`
	Type         string `json:"type"`
	Price        string `json:"price"`
	OrigQty      string `json:"origQty"`
	ExecutedQty  string `json:"executedQty"`
	Status       string `json:"status"`
	StopPrice    string `json:"stopPrice"`
	TimeInForce  string `json:"timeInForce"`
	AvgPrice     string `json:"avgPrice"`
	ReduceOnly   bool   `json:"reduceOnly"`
	UpdateTime   int64  `json:"updateTime"`
}

// FuturesUserTrade represents GET /fapi/v1/userTrades.
type FuturesUserTrade struct {
	ID              int64  `json:"id"`
	Symbol          string `json:"symbol"`
	Side            string `json:"side"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	RealizedPnl     string `json:"realizedPnl"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	PositionSide    string `json:"positionSide"`
	Buyer           bool   `json:"buyer"`
	Maker           bool   `json:"maker"`
}

// IncomeRecord represents GET /fapi/v1/income.
type IncomeRecord struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Time       int64  `json:"time"`
	TranID     int64  `json:"tranId"`
	TradeID    string `json:"tradeId"`
	Info       string `json:"info"`
}

// IncomeHistoryOptions holds optional parameters for GetIncomeHistory.
type IncomeHistoryOptions struct {
	Symbol     string
	IncomeType string // REALIZED_PNL, FUNDING_FEE, COMMISSION, TRANSFER, etc.
	StartTime  int64  // Unix milliseconds
	EndTime    int64  // Unix milliseconds
	Limit      int    // Max 1000
}
