package binance

import (
	"encoding/hex"
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		secretKey   string
		expected    string
	}{
		{
			// Binance official test vector from API documentation.
			name:        "binance official test vector",
			queryString: "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559",
			secretKey:   "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j",
			expected:    "c8db56825ae71d6d79447849e617115f4a920fa2acdcab2b053c4b2838bd6b71",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sign(tt.queryString, tt.secretKey)
			if got != tt.expected {
				t.Errorf("sign() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSign_OutputFormat(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		secretKey   string
	}{
		{
			name:        "empty query string",
			queryString: "",
			secretKey:   "secret",
		},
		{
			name:        "empty secret key",
			queryString: "symbol=BTCUSDT",
			secretKey:   "",
		},
		{
			name:        "both empty",
			queryString: "",
			secretKey:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sign(tt.queryString, tt.secretKey)

			// SHA256 output is always 32 bytes = 64 hex chars.
			if len(got) != 64 {
				t.Errorf("sign() length = %d, want 64", len(got))
			}

			// Must be valid hex.
			if _, err := hex.DecodeString(got); err != nil {
				t.Errorf("sign() produced invalid hex: %v", err)
			}
		})
	}
}

func TestSign_Deterministic(t *testing.T) {
	qs := "timestamp=1234567890"
	key := "mykey"

	first := sign(qs, key)
	second := sign(qs, key)

	if first != second {
		t.Errorf("sign() not deterministic: %q != %q", first, second)
	}
}
