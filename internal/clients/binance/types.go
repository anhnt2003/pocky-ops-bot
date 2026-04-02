// Package binance provides a client for the Binance REST API.
package binance

// Balance represents a single asset balance in a Binance account.
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// AccountResponse represents the response from the GET /api/v3/account endpoint.
type AccountResponse struct {
	Balances    []Balance `json:"balances"`
	CanTrade    bool      `json:"canTrade"`
	AccountType string    `json:"accountType"`
	UpdateTime  int64     `json:"updateTime"`
}

// TickerPrice represents a symbol price ticker from GET /api/v3/ticker/price.
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// Ticker24hr represents 24hr rolling window price change statistics from GET /api/v3/ticker/24hr.
type Ticker24hr struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	LastPrice          string `json:"lastPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	Count              int64  `json:"count"`
}
