package binance

const defaultFuturesBaseURL = "https://fapi.binance.com"

// FuturesClient is a Binance USD-M Futures API client.
// It wraps the base Client with the futures base URL (fapi.binance.com).
type FuturesClient struct {
	base *Client
}

// NewFuturesClient creates a new Futures client reusing the same API key/secret
// as spot but targeting the futures base URL.
func NewFuturesClient(apiKey, secretKey string, opts ...ClientOption) (*FuturesClient, error) {
	// Prepend default futures base URL; caller can override with WithBaseURL.
	allOpts := append([]ClientOption{WithBaseURL(defaultFuturesBaseURL)}, opts...)

	base, err := NewClient(apiKey, secretKey, allOpts...)
	if err != nil {
		return nil, err
	}

	return &FuturesClient{base: base}, nil
}
