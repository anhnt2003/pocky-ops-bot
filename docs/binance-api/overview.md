# Binance Spot REST API - Overview & Authentication

> Source: https://developers.binance.com/docs/binance-spot-api-docs/rest-api
> Fetched: 2026-04-02
> Content type: API Reference

This document covers base URLs, authentication (HMAC SHA256 signatures), rate limits, enums, and general API conventions. Reference this when building or debugging the Binance API client.

---

## Base URLs

| URL | Notes |
|-----|-------|
| `https://api.binance.com` | Primary |
| `https://api-gcp.binance.com` | GCP endpoint |
| `https://api1.binance.com` | Alternative (better perf, less stable) |
| `https://api2.binance.com` | Alternative |
| `https://api3.binance.com` | Alternative |
| `https://api4.binance.com` | Alternative |
| `https://data-api.binance.vision` | Public market data only |

- APIs have a **timeout of 10 seconds** per request (error `-1007` if exceeded).
- All time/timestamp fields in responses are in **milliseconds** by default. Add header `X-MBX-TIME-UNIT:MICROSECOND` for microsecond precision.
- Avoid SQL keywords in requests — they may trigger a WAF security block.

---

## Request Format

- **GET**: parameters must be sent as a `query string`
- **POST / PUT / DELETE**: parameters may be sent as query string or request body with `application/x-www-form-urlencoded`
- Parameters can be sent in any order
- If a parameter appears in both query string and request body, the **query string value takes precedence**

---

## Security Types

| Type | Description | Requires API Key | Requires Signature |
|------|-------------|------------------|-------------------|
| `NONE` | Public market data | No | No |
| `TRADE` | Place/cancel orders | Yes | Yes |
| `USER_DATA` | Private account info, order status, trade history | Yes | Yes |
| `USER_STREAM` | Manage User Data Stream subscriptions | Yes | No |

### API Key Header

All endpoints except `NONE` require:

```
X-MBX-APIKEY: {your-api-key}
```

---

## HMAC SHA256 Signature Generation

Endpoints with `TRADE` or `USER_DATA` security require a signed request.

### Step 1: Build the Payload

Concatenate all query string parameters as `key=value` pairs separated by `&`. If using a request body, concatenate query string + body (no separator between them).

All non-ASCII characters must be **percent-encoded** before signing.

Example payload:
```
symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559
```

### Step 2: Compute HMAC SHA256

Sign the payload using your `secretKey` as the HMAC key. Output as **hexadecimal**.

> The `secretKey` and payload are **case-sensitive**. The resulting signature is **case-insensitive**.

```bash
echo -n "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559" \
  | openssl dgst -sha256 -hmac "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j"
```

Output: `c8db56825ae71d6d79447849e617115f4a920fa2acdcab2b053c4b2838bd6b71`

### Step 3: Attach Signature

Append `&signature={computed_signature}` to the query string.

### Complete Bash Example

```bash
apiKey="vmPUZE6mv9SD5VNHk4HlWFsOr6aKE2zvsw0MuIgwCIPy6utIco14y7Ju91duEh8A"
secretKey="NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j"
payload="symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559"
signature=$(echo -n "$payload" | openssl dgst -sha256 -hmac "$secretKey")
signature=${signature#*= }
curl -H "X-MBX-APIKEY: $apiKey" -X POST "https://api.binance.com/api/v3/order?$payload&signature=$signature"
```

### RSA Signatures

- Algorithm: RSASSA-PKCS1-v1_5 with SHA-256
- Output: **base64-encoded** (must be percent-encoded before appending to URL)
- Signatures are **case-sensitive**

### Ed25519 Signatures

- Output: **base64-encoded** (must be percent-encoded)
- Signatures are **case-sensitive**

---

## Timing Security

All SIGNED requests require:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `timestamp` | LONG | Yes | Current Unix time in milliseconds |
| `recvWindow` | DECIMAL | No | Validity window in ms. Default: `5000`, max: `60000` |

**Server validation logic:**
```
if (timestamp < (serverTime + 1 second) && (serverTime - timestamp) <= recvWindow)
    → request is valid
```

> Recommendation: use a small `recvWindow` of **5000 or less**.

---

## Rate Limits

| Limiter | Limit | Interval |
|---------|-------|----------|
| `REQUEST_WEIGHT` | 6000 | Per minute |
| `ORDERS` | 10 | Per second |
| `RAW_REQUESTS` | 61000 | Per 5 minutes |

---

## Enums

### Symbol Status
`TRADING` | `END_OF_DAY` | `HALT` | `BREAK`

### Order Side
`BUY` | `SELL`

### Order Types
`LIMIT` | `MARKET` | `STOP_LOSS` | `STOP_LOSS_LIMIT` | `TAKE_PROFIT` | `TAKE_PROFIT_LIMIT` | `LIMIT_MAKER`

### Time in Force
| Value | Meaning |
|-------|---------|
| `GTC` | Good Til Canceled |
| `IOC` | Immediate Or Cancel |
| `FOK` | Fill or Kill |

### Order Status
| Value | Meaning |
|-------|---------|
| `NEW` | Accepted by engine |
| `PENDING_NEW` | Pending until working order fully filled |
| `PARTIALLY_FILLED` | Part of order filled |
| `FILLED` | Completed |
| `CANCELED` | Canceled by user |
| `PENDING_CANCEL` | Currently unused |
| `REJECTED` | Not accepted by engine |
| `EXPIRED` | Canceled per order type rules or by exchange |
| `EXPIRED_IN_MATCH` | Expired due to STP |

### Order Response Type
`ACK` | `RESULT` | `FULL`

### Order List Status (listStatusType)
`RESPONSE` | `EXEC_STARTED` | `UPDATED` | `ALL_DONE`

### Order List Order Status (listOrderStatus)
`EXECUTING` | `ALL_DONE` | `REJECT`

### ContingencyType
`OCO` | `OTO`

### Self-Trade Prevention Modes
`NONE` | `EXPIRE_MAKER` | `EXPIRE_TAKER` | `EXPIRE_BOTH` | `DECREMENT` | `TRANSFER`

### Execution Types
`NEW` | `CANCELED` | `REPLACED` | `REJECTED` | `TRADE` | `EXPIRED` | `TRADE_PREVENTION`

### Working Floor
`EXCHANGE` | `SOR`

### AllocationType
`SOR`

### Rate Limit Intervals
`SECOND` | `MINUTE` | `DAY`

### Kline/Candlestick Intervals
| Period | Values |
|--------|--------|
| Seconds | `1s` |
| Minutes | `1m`, `3m`, `5m`, `15m`, `30m` |
| Hours | `1h`, `2h`, `4h`, `6h`, `8h`, `12h` |
| Days | `1d`, `3d` |
| Weeks | `1w` |
| Months | `1M` |

### Account Permissions
`SPOT` | `MARGIN` | `LEVERAGED` | `TRD_GRP_002` through `TRD_GRP_025`
