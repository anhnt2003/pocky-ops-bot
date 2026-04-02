# Binance Spot REST API Reference

> Source: https://developers.binance.com/docs/binance-spot-api-docs/rest-api
> Fetched: 2026-04-02

API reference for building Binance integration into pocky-ops-bot. Covers account data retrieval, current price lookups, trade history, and everything needed to calculate investment amounts and P&L.

## Documents

| File | Description |
|------|-------------|
| [overview.md](overview.md) | Base URLs, authentication (HMAC SHA256), rate limits, enums |
| [account-endpoints.md](account-endpoints.md) | Account info, balances, orders, trade history, commissions |
| [market-data-endpoints.md](market-data-endpoints.md) | Price tickers, klines, order book, 24hr stats |
| [error-codes.md](error-codes.md) | All error codes with retryable/non-retryable classification |

## Quick Reference: Investment Tracking Workflow

```
1. GET /api/v3/account (omitZeroBalances=true)
   → Get all held assets with free + locked balances

2. GET /api/v3/ticker/price (symbols=["BTCUSDT","ETHUSDT",...])
   → Get current price for each asset

3. GET /api/v3/myTrades (symbol=BTCUSDT, paginate with fromId)
   → Full trade history for cost basis calculation

4. Calculate:
   - Current Value = sum(holding[asset] * price[asset])
   - Cost Basis    = sum(buy_trades.quoteQty) - sum(sell_trades.quoteQty)
   - Unrealized PnL = Current Value - Cost Basis
```

## Key Endpoints Summary

| Endpoint | Weight | Auth | Purpose |
|----------|--------|------|---------|
| `GET /api/v3/account` | 20 | USER_DATA | Balances & account info |
| `GET /api/v3/myTrades` | 20/5 | USER_DATA | Trade history (P&L calc) |
| `GET /api/v3/allOrders` | 20 | USER_DATA | Order history |
| `GET /api/v3/openOrders` | 6/80 | USER_DATA | Open orders |
| `GET /api/v3/ticker/price` | 2/4 | NONE | Current prices |
| `GET /api/v3/ticker/24hr` | 2-80 | NONE | 24h price change stats |
| `GET /api/v3/klines` | 2 | NONE | Historical OHLCV candles |
| `GET /api/v3/account/commission` | 20 | USER_DATA | Commission rates |

## Rate Limits

| Type | Limit | Window |
|------|-------|--------|
| Request Weight | 6,000 | 1 minute |
| Orders | 10 | 1 second |
| Raw Requests | 61,000 | 5 minutes |

## Authentication

All USER_DATA endpoints require:
1. API key in header: `X-MBX-APIKEY: {key}`
2. HMAC SHA256 signature of query params using secret key
3. `timestamp` parameter (Unix ms)
4. Optional `recvWindow` (default 5000ms, max 60000ms)

See [overview.md](overview.md) for full signature generation details and examples.
