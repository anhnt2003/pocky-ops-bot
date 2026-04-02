package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAccount_Success(t *testing.T) {
	expectedResp := AccountResponse{
		Balances: []Balance{
			{Asset: "BTC", Free: "0.50000000", Locked: "0.10000000"},
			{Asset: "USDT", Free: "1000.00000000", Locked: "0.00000000"},
		},
		CanTrade:    true,
		AccountType: "SPOT",
		UpdateTime:  1705276800000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a signed request.
		if r.Header.Get("X-MBX-APIKEY") == "" {
			t.Error("missing X-MBX-APIKEY header")
		}
		q := r.URL.Query()
		if q.Get("signature") == "" {
			t.Error("missing signature parameter")
		}
		if q.Get("timestamp") == "" {
			t.Error("missing timestamp parameter")
		}

		// Verify omitZeroBalances is set.
		if q.Get("omitZeroBalances") != "true" {
			t.Errorf("omitZeroBalances = %q, want %q", q.Get("omitZeroBalances"), "true")
		}

		// Verify path.
		if r.URL.Path != "/api/v3/account" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v3/account")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", "test-secret-key",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.GetAccount(context.Background())
	if err != nil {
		t.Fatalf("GetAccount() error = %v", err)
	}

	if !resp.CanTrade {
		t.Error("CanTrade = false, want true")
	}
	if resp.AccountType != "SPOT" {
		t.Errorf("AccountType = %q, want %q", resp.AccountType, "SPOT")
	}
	if len(resp.Balances) != 2 {
		t.Fatalf("len(Balances) = %d, want 2", len(resp.Balances))
	}
	if resp.Balances[0].Asset != "BTC" {
		t.Errorf("Balances[0].Asset = %q, want %q", resp.Balances[0].Asset, "BTC")
	}
	if resp.Balances[0].Free != "0.50000000" {
		t.Errorf("Balances[0].Free = %q, want %q", resp.Balances[0].Free, "0.50000000")
	}
	if resp.Balances[1].Asset != "USDT" {
		t.Errorf("Balances[1].Asset = %q, want %q", resp.Balances[1].Asset, "USDT")
	}
}

func TestGetAccount_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": -2014,
			"msg":  "API-key format invalid.",
		})
	}))
	defer server.Close()

	client, err := NewClient("bad-key", "bad-secret",
		WithBaseURL(server.URL),
		WithClock(fixedClock{t: fixedTime}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetAccount(context.Background())
	if err == nil {
		t.Fatal("GetAccount() expected error, got nil")
	}

	binErr, ok := err.(*BinanceError)
	if !ok {
		t.Fatalf("expected *BinanceError, got %T: %v", err, err)
	}

	if binErr.HTTPStatus != 401 {
		t.Errorf("HTTPStatus = %d, want 401", binErr.HTTPStatus)
	}
	if binErr.Code != -2014 {
		t.Errorf("Code = %d, want -2014", binErr.Code)
	}
	if binErr.IsRetryable() {
		t.Error("expected non-retryable error for 401")
	}
}
