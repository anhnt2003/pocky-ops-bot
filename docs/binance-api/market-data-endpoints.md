# Binance Spot REST API - Market Data Endpoints

> Source: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/market-data-endpoints
> Fetched: 2026-04-02
> Content type: API Reference

All endpoints here are **public** (security: `NONE`) — no API key or signature required. Reference this when retrieving current prices, historical candles, and market stats for P&L calculations.

---

## 1. Symbol Price Ticker

```
GET /api/v3/ticker/price
```

**Weight:** 2 (single symbol) | 4 (all/multiple) | **Data Source:** Memory

Returns the latest price for a symbol or all symbols.

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | No | Cannot combine with `symbols`. Returns all if omitted |
| `symbols` | STRING | No | JSON array format: `["BTCUSDT","BNBUSDT"]` or URL-encoded |
| `symbolStatus` | ENUM | No | Filter: `TRADING`, `HALT`, `BREAK` |

### Response (single)

```json
{
  "symbol": "LTCBTC",
  "price": "4.00000200"
}
```

### Response (multiple)

```json
[
  { "symbol": "LTCBTC", "price": "4.00000200" },
  { "symbol": "ETHBTC", "price": "0.07946600" }
]
```

---

## 2. 24hr Ticker Price Change Statistics

```
GET /api/v3/ticker/24hr
```

**Data Source:** Memory

### Weight

| Scenario | Weight |
|----------|--------|
| Single symbol | 2 |
| 1-20 symbols | 2 |
| 21-100 symbols | 40 |
| 101+ symbols | 80 |
| No symbol (all) | 80 |

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | No | Mutually exclusive with `symbols` |
| `symbols` | STRING | No | JSON array format |
| `type` | ENUM | No | `FULL` (default) or `MINI` |
| `symbolStatus` | ENUM | No | `TRADING`, `HALT`, `BREAK` |

### Response (FULL)

```json
{
  "symbol": "BNBBTC",
  "priceChange": "-94.99999800",
  "priceChangePercent": "-95.960",
  "weightedAvgPrice": "0.29628482",
  "prevClosePrice": "0.10002000",
  "lastPrice": "4.00000200",
  "lastQty": "200.00000000",
  "bidPrice": "4.00000000",
  "bidQty": "100.00000000",
  "askPrice": "4.00000200",
  "askQty": "100.00000000",
  "openPrice": "99.00000000",
  "highPrice": "100.00000000",
  "lowPrice": "0.10000000",
  "volume": "8913.30000000",
  "quoteVolume": "15.30000000",
  "openTime": 1499783499040,
  "closeTime": 1499869899040,
  "firstId": 28385,
  "lastId": 28460,
  "count": 76
}
```

### Response (MINI)

```json
{
  "symbol": "BNBBTC",
  "openPrice": "99.00000000",
  "highPrice": "100.00000000",
  "lowPrice": "0.10000000",
  "lastPrice": "4.00000200",
  "volume": "8913.30000000",
  "quoteVolume": "15.30000000",
  "openTime": 1499783499040,
  "closeTime": 1499869899040,
  "firstId": 28385,
  "lastId": 28460,
  "count": 76
}
```

---

## 3. Current Average Price

```
GET /api/v3/avgPrice
```

**Weight:** 2 | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory |
|------|------|-----------|
| `symbol` | STRING | Yes |

### Response

```json
{
  "mins": 5,
  "price": "9.35751834",
  "closeTime": 1694061154503
}
```

---

## 4. Kline / Candlestick Data

```
GET /api/v3/klines
```

**Weight:** 2 | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `interval` | ENUM | Yes | See intervals below |
| `startTime` | LONG | No | Millisecond timestamp (UTC) |
| `endTime` | LONG | No | Millisecond timestamp (UTC) |
| `timeZone` | STRING | No | Default: `0` (UTC). Format: `-1:00`, `05:45`, or hours. Range: `-12:00` to `+14:00` |
| `limit` | INT | No | Default: 500, max: 1000 |

### Supported Intervals

| Period | Values |
|--------|--------|
| Seconds | `1s` |
| Minutes | `1m`, `3m`, `5m`, `15m`, `30m` |
| Hours | `1h`, `2h`, `4h`, `6h`, `8h`, `12h` |
| Days | `1d`, `3d` |
| Weeks | `1w` |
| Months | `1M` |

> `startTime` and `endTime` always interpret as UTC regardless of `timeZone` parameter.

### Response

```json
[
  [
    1499040000000,      // [0] Kline open time
    "0.01634790",       // [1] Open price
    "0.80000000",       // [2] High price
    "0.01575800",       // [3] Low price
    "0.01577100",       // [4] Close price
    "148976.11427815",  // [5] Volume
    1499644799999,      // [6] Kline close time
    "2434.19055334",    // [7] Quote asset volume
    308,                // [8] Number of trades
    "1756.87402397",    // [9] Taker buy base asset volume
    "28.46694368",      // [10] Taker buy quote asset volume
    "0"                 // [11] Unused field, ignore
  ]
]
```

---

## 5. UI Klines

```
GET /api/v3/uiKlines
```

**Weight:** 2 | **Data Source:** Database

Identical to `/api/v3/klines` with optimized presentation for candlestick charts. Same parameters and response structure.

---

## 6. Trading Day Ticker

```
GET /api/v3/ticker/tradingDay
```

**Weight:** 4 per symbol, capped at 200 for 50+ symbols | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes* | Either `symbol` or `symbols` required |
| `symbols` | STRING | No | Max 100 symbols |
| `timeZone` | STRING | No | Format: `-1:00`, `05:45`, or hours |
| `type` | ENUM | No | `FULL` (default) or `MINI` |
| `symbolStatus` | ENUM | No | `TRADING`, `HALT`, `BREAK` |

### Response (FULL)

```json
{
  "symbol": "BTCUSDT",
  "priceChange": "-83.13000000",
  "priceChangePercent": "-0.317",
  "weightedAvgPrice": "26234.58803036",
  "openPrice": "26304.80000000",
  "highPrice": "26397.46000000",
  "lowPrice": "26088.34000000",
  "lastPrice": "26221.67000000",
  "volume": "18495.35066000",
  "quoteVolume": "485217905.04210480",
  "openTime": 1695686400000,
  "closeTime": 1695772799999,
  "firstId": 3220151555,
  "lastId": 3220849281,
  "count": 697727
}
```

---

## 7. Symbol Order Book Ticker

```
GET /api/v3/ticker/bookTicker
```

**Weight:** 2 (single) | 4 (all/multiple) | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory |
|------|------|-----------|
| `symbol` | STRING | No |
| `symbols` | STRING | No |
| `symbolStatus` | ENUM | No |

### Response

```json
{
  "symbol": "LTCBTC",
  "bidPrice": "4.00000000",
  "bidQty": "431.00000000",
  "askPrice": "4.00000200",
  "askQty": "9.00000000"
}
```

---

## 8. Rolling Window Price Change Statistics

```
GET /api/v3/ticker
```

**Weight:** 4 per symbol, capped at 200 for 50+ symbols | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes* | Either `symbol` or `symbols` required |
| `symbols` | STRING | No | Max 100 symbols |
| `windowSize` | ENUM | No | `1m`-`59m`, `1h`-`23h`, `1d`-`7d`. Default: `1d` |
| `type` | ENUM | No | `FULL` or `MINI` |
| `symbolStatus` | ENUM | No | `TRADING`, `HALT`, `BREAK` |

> Window precision: no more than 59999ms from requested windowSize. `openTime` starts on the minute; `closeTime` is the request time.

### Response (FULL)

```json
{
  "symbol": "BNBBTC",
  "priceChange": "-8.00000000",
  "priceChangePercent": "-88.889",
  "weightedAvgPrice": "2.60427807",
  "openPrice": "9.00000000",
  "highPrice": "9.00000000",
  "lowPrice": "1.00000000",
  "lastPrice": "1.00000000",
  "volume": "187.00000000",
  "quoteVolume": "487.00000000",
  "openTime": 1641859200000,
  "closeTime": 1642031999999,
  "firstId": 0,
  "lastId": 60,
  "count": 61
}
```

---

## 9. Order Book (Depth)

```
GET /api/v3/depth
```

**Weight:** 5 (1-100) | 25 (101-500) | 50 (501-1000) | 250 (1001-5000) | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `limit` | INT | No | Default: 100, max: 5000 |
| `symbolStatus` | ENUM | No | |

---

## 10. Recent Trades

```
GET /api/v3/trades
```

**Weight:** 25 | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `limit` | INT | No | Default: 500, max: 1000 |

---

## 11. Historical Trades

```
GET /api/v3/historicalTrades
```

**Weight:** 25 | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `limit` | INT | No | Default: 500, max: 1000 |
| `fromId` | LONG | No | Trade ID to fetch from |

---

## 12. Aggregate Trades

```
GET /api/v3/aggTrades
```

**Weight:** 4 | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `fromId` | LONG | No | Aggregate trade ID to fetch from |
| `startTime` | LONG | No | ms timestamp |
| `endTime` | LONG | No | ms timestamp |
| `limit` | INT | No | Default: 500, max: 1000 |

---

## 13. Reference Price

```
GET /api/v3/referencePrice
```

**Weight:** 2 | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory |
|------|------|-----------|
| `symbol` | STRING | Yes |

### Response

```json
{
  "symbol": "BAZUSD",
  "referencePrice": "10.00",
  "timestamp": 1770736694138
}
```

> `referencePrice` may be `null` if not set for the symbol.

---

## 14. Reference Price Calculation

```
GET /api/v3/referencePrice/calculation
```

**Weight:** 2 | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory |
|------|------|-----------|
| `symbol` | STRING | Yes |
| `symbolStatus` | ENUM | No |

### Response (arithmetic mean)

```json
{
  "symbol": "BAZUSD",
  "calculationType": "ARITHMETIC_MEAN",
  "bucketCount": 10,
  "bucketWidthMs": 1000
}
```

### Response (external)

```json
{
  "symbol": "BAZUSD",
  "calculationType": "EXTERNAL",
  "externalCalculationId": 42
}
```

> Error `-2043` if symbol has no reference price.

---

## Key Endpoints for Investment Tracking

| Use Case | Endpoint | Why |
|----------|----------|-----|
| Current portfolio value | `GET /api/v3/ticker/price` | Get latest price for each held asset |
| 24h change overview | `GET /api/v3/ticker/24hr` | Price change %, volume, high/low |
| Historical price charts | `GET /api/v3/klines` | OHLCV candles for any timeframe |
| Average price | `GET /api/v3/avgPrice` | 5-minute weighted average |
| Best bid/ask spread | `GET /api/v3/ticker/bookTicker` | Current order book top |
