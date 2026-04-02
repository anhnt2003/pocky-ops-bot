# Binance Spot REST API - Account Endpoints

> Source: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/account-endpoints
> Fetched: 2026-04-02
> Content type: API Reference

All endpoints in this document require `USER_DATA` security (API key + HMAC SHA256 signature). Reference this when implementing account balance retrieval, trade history, and P&L calculations.

---

## 1. Account Information

```
GET /api/v3/account
```

**Weight:** 20 | **Security:** USER_DATA | **Data Source:** Memory => Database

Returns balances, commission rates, account permissions, and trading status.

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `omitZeroBalances` | BOOLEAN | No | When `true`, returns only non-zero balances. Default: `false` |
| `recvWindow` | DECIMAL | No | Max `60000`. Supports 3 decimal places (e.g. `6000.346`) |
| `timestamp` | LONG | Yes | Unix time in milliseconds |

### Response

```json
{
  "makerCommission": 15,
  "takerCommission": 15,
  "buyerCommission": 0,
  "sellerCommission": 0,
  "commissionRates": {
    "maker": "0.00150000",
    "taker": "0.00150000",
    "buyer": "0.00000000",
    "seller": "0.00000000"
  },
  "canTrade": true,
  "canWithdraw": true,
  "canDeposit": true,
  "brokered": false,
  "requireSelfTradePrevention": false,
  "preventSor": false,
  "updateTime": 123456789,
  "accountType": "SPOT",
  "balances": [
    {
      "asset": "BTC",
      "free": "4723846.89208129",
      "locked": "0.00000000"
    }
  ],
  "permissions": ["SPOT"],
  "uid": 354937868
}
```

> **Key fields for investment tracking:**
> - `balances[].asset` — asset symbol (e.g. "BTC", "USDT")
> - `balances[].free` — available balance
> - `balances[].locked` — balance locked in open orders
> - Total holding = `free` + `locked`

---

## 2. Query Order

```
GET /api/v3/order
```

**Weight:** 4 | **Security:** USER_DATA | **Data Source:** Memory => Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair (e.g. `BTCUSDT`) |
| `orderId` | LONG | No | Order ID |
| `origClientOrderId` | STRING | No | Original client order ID |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

> Either `orderId` or `origClientOrderId` must be provided. If both given, `orderId` is searched first, then `origClientOrderId` is verified.

### Response

```json
{
  "symbol": "LTCBTC",
  "orderId": 1,
  "orderListId": -1,
  "clientOrderId": "myOrder1",
  "price": "0.1",
  "origQty": "1.0",
  "executedQty": "0.0",
  "cummulativeQuoteQty": "0.0",
  "status": "NEW",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY",
  "stopPrice": "0.0",
  "icebergQty": "0.0",
  "time": 1499827319559,
  "updateTime": 1499827319559,
  "isWorking": true,
  "workingTime": 1499827319559,
  "origQuoteOrderQty": "0.000000",
  "selfTradePreventionMode": "NONE"
}
```

> Some historical orders may have `cummulativeQuoteQty` < 0 (meaning data unavailable).

---

## 3. Current Open Orders

```
GET /api/v3/openOrders
```

**Weight:** 6 (single symbol) | 80 (all symbols) | **Security:** USER_DATA | **Data Source:** Memory => Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | No | If omitted, returns open orders for all symbols |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Response

Array of order objects (same structure as Query Order).

---

## 4. All Orders

```
GET /api/v3/allOrders
```

**Weight:** 20 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `orderId` | LONG | No | Get orders >= this ID |
| `startTime` | LONG | No | Start time filter (ms) |
| `endTime` | LONG | No | End time filter (ms) |
| `limit` | INT | No | Default: 500, max: 1000 |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

> - If `orderId` set, returns orders >= that ID; otherwise returns most recent
> - If `startTime`/`endTime` provided, `orderId` is not required
> - **Time range cannot exceed 24 hours**

### Response

Array of order objects (same structure as Query Order).

---

## 5. Account Trade List

```
GET /api/v3/myTrades
```

**Weight:** 20 (without `orderId`) | 5 (with `orderId`) | **Security:** USER_DATA | **Data Source:** Memory => Database

This is the **primary endpoint for calculating investment P&L** — it returns all executed trades with prices, quantities, and commissions.

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `orderId` | LONG | No | Must be combined with `symbol` |
| `startTime` | LONG | No | Start time filter (ms) |
| `endTime` | LONG | No | End time filter (ms) |
| `fromId` | LONG | No | Trade ID to fetch from (default: most recent) |
| `limit` | INT | No | Default: 500, max: 1000 |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Supported Parameter Combinations

- `symbol`
- `symbol` + `orderId`
- `symbol` + `startTime`
- `symbol` + `endTime`
- `symbol` + `fromId`
- `symbol` + `startTime` + `endTime`
- `symbol` + `orderId` + `fromId`

> - If `fromId` set, returns trades >= that ID; otherwise most recent
> - **Time range cannot exceed 24 hours**

### Response

```json
[
  {
    "symbol": "BNBBTC",
    "id": 28457,
    "orderId": 100234,
    "orderListId": -1,
    "price": "4.00000100",
    "qty": "12.00000000",
    "quoteQty": "48.000012",
    "commission": "10.10000000",
    "commissionAsset": "BNB",
    "time": 1499865549590,
    "isBuyer": true,
    "isMaker": false,
    "isBestMatch": true
  }
]
```

> **Key fields for P&L calculation:**
> - `price` — execution price
> - `qty` — base asset quantity traded
> - `quoteQty` — quote asset amount (`price * qty`)
> - `commission` — fee charged
> - `commissionAsset` — asset used for fee (often BNB)
> - `isBuyer` — `true` = buy, `false` = sell
> - `time` — trade execution timestamp

---

## 6. Query Order List (OCO)

```
GET /api/v3/orderList
```

**Weight:** 4 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `orderListId` | LONG | No* | Query by order list ID |
| `origClientOrderId` | STRING | No* | Query by client order ID |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

> Either `orderListId` or `origClientOrderId` is required.

### Response

```json
{
  "orderListId": 27,
  "contingencyType": "OCO",
  "listStatusType": "EXEC_STARTED",
  "listOrderStatus": "EXECUTING",
  "listClientOrderId": "h2USkA5YQpaXHPIrkd96xE",
  "transactionTime": 1565245656253,
  "symbol": "LTCBTC",
  "orders": [
    {
      "symbol": "LTCBTC",
      "orderId": 4,
      "clientOrderId": "qD1gy3kc3Gx0rihm9Y3xwS"
    },
    {
      "symbol": "LTCBTC",
      "orderId": 5,
      "clientOrderId": "ARzZ9I00CPM8i3NhmU9Ega"
    }
  ]
}
```

---

## 7. Query All Order Lists

```
GET /api/v3/allOrderList
```

**Weight:** 20 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `fromId` | LONG | No | If supplied, `startTime`/`endTime` cannot be provided |
| `startTime` | LONG | No | |
| `endTime` | LONG | No | |
| `limit` | INT | No | Default: 500, max: 1000 |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

> Time range cannot exceed 24 hours.

---

## 8. Query Open Order Lists

```
GET /api/v3/openOrderList
```

**Weight:** 6 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

---

## 9. Query Unfilled Order Count (Rate Limits)

```
GET /api/v3/rateLimit/order
```

**Weight:** 40 | **Security:** USER_DATA | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Response

```json
[
  {
    "rateLimitType": "ORDERS",
    "interval": "SECOND",
    "intervalNum": 10,
    "limit": 50,
    "count": 0
  },
  {
    "rateLimitType": "ORDERS",
    "interval": "DAY",
    "intervalNum": 1,
    "limit": 160000,
    "count": 0
  }
]
```

---

## 10. Query Prevented Matches

```
GET /api/v3/myPreventedMatches
```

**Weight:** 2 (by `preventedMatchId`) | 20 (by `orderId`) | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `preventedMatchId` | LONG | No | |
| `orderId` | LONG | No | |
| `fromPreventedMatchId` | LONG | No | |
| `limit` | INT | No | Default: 500, max: 1000 |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Supported Combinations

- `symbol` + `preventedMatchId`
- `symbol` + `orderId`
- `symbol` + `orderId` + `fromPreventedMatchId`
- `symbol` + `orderId` + `fromPreventedMatchId` + `limit`

### Response

```json
[
  {
    "symbol": "BTCUSDT",
    "preventedMatchId": 1,
    "takerOrderId": 5,
    "makerSymbol": "BTCUSDT",
    "makerOrderId": 3,
    "tradeGroupId": 1,
    "selfTradePreventionMode": "EXPIRE_MAKER",
    "price": "1.100000",
    "makerPreventedQuantity": "1.300000",
    "transactTime": 1669101687094
  }
]
```

---

## 11. Query Allocations

```
GET /api/v3/myAllocations
```

**Weight:** 20 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `startTime` | LONG | No | |
| `endTime` | LONG | No | |
| `fromAllocationId` | INT | No | |
| `limit` | INT | No | Default: 500, max: 1000 |
| `orderId` | LONG | No | |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | No | |

### Supported Parameter Combinations

- `symbol` — oldest to newest
- `symbol` + `startTime` — oldest since startTime
- `symbol` + `endTime` — newest until endTime
- `symbol` + `startTime` + `endTime` — within range
- `symbol` + `fromAllocationId` — by allocation ID
- `symbol` + `orderId` — by order, oldest first
- `symbol` + `orderId` + `fromAllocationId`

> Time range cannot exceed 24 hours.

### Response

```json
[
  {
    "symbol": "BTCUSDT",
    "allocationId": 0,
    "allocationType": "SOR",
    "orderId": 1,
    "orderListId": -1,
    "price": "1.00000000",
    "qty": "5.00000000",
    "quoteQty": "5.00000000",
    "commission": "0.00000000",
    "commissionAsset": "BTC",
    "time": 1687506878118,
    "isBuyer": true,
    "isMaker": false,
    "isAllocator": false
  }
]
```

---

## 12. Query Commission Rates

```
GET /api/v3/account/commission
```

**Weight:** 20 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |

### Response

```json
{
  "symbol": "BTCUSDT",
  "standardCommission": {
    "maker": "0.00000010",
    "taker": "0.00000020",
    "buyer": "0.00000030",
    "seller": "0.00000040"
  },
  "specialCommission": {
    "maker": "0.01000000",
    "taker": "0.02000000",
    "buyer": "0.03000000",
    "seller": "0.04000000"
  },
  "taxCommission": {
    "maker": "0.00000112",
    "taker": "0.00000114",
    "buyer": "0.00000118",
    "seller": "0.00000116"
  },
  "discount": {
    "enabledForAccount": true,
    "enabledForSymbol": true,
    "discountAsset": "BNB",
    "discount": "0.75000000"
  }
}
```

---

## 13. Query Order Amendments

```
GET /api/v3/order/amendments
```

**Weight:** 4 | **Security:** USER_DATA | **Data Source:** Database

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `orderId` | LONG | Yes | |
| `fromExecutionId` | LONG | No | |
| `limit` | LONG | No | Default: 500, max: 1000 |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Response

```json
[
  {
    "symbol": "BTCUSDT",
    "orderId": 9,
    "executionId": 22,
    "origClientOrderId": "W0fJ9fiLKHOJutovPK3oJp",
    "newClientOrderId": "UQ1Np3bmQ71jJzsSDW9Vpi",
    "origQty": "5.00000000",
    "newQty": "4.00000000",
    "time": 1741669661670
  }
]
```

---

## 14. Query Relevant Filters

```
GET /api/v3/myFilters
```

**Weight:** 40 | **Security:** USER_DATA | **Data Source:** Memory

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | Yes | Trading pair |
| `recvWindow` | DECIMAL | No | Max `60000` |
| `timestamp` | LONG | Yes | |

### Response

```json
{
  "exchangeFilters": [
    {
      "filterType": "EXCHANGE_MAX_NUM_ORDERS",
      "maxNumOrders": 1000
    }
  ],
  "symbolFilters": [
    {
      "filterType": "MAX_NUM_ORDER_LISTS",
      "maxNumOrderLists": 20
    }
  ],
  "assetFilters": [
    {
      "filterType": "MAX_ASSET",
      "asset": "JPY",
      "limit": "1000000.00000000"
    }
  ]
}
```

---

## Investment Tracking Strategy

To calculate current investment amount, P&L, and portfolio info using these endpoints:

### 1. Get Current Holdings
- Call `GET /api/v3/account` with `omitZeroBalances=true`
- For each asset: total = `free` + `locked`

### 2. Get Current Prices
- Call `GET /api/v3/ticker/price` (see [market-data-endpoints.md](market-data-endpoints.md))
- Current value per asset = `total_holding * current_price`

### 3. Calculate Cost Basis (from Trade History)
- Call `GET /api/v3/myTrades` for each symbol
- For buys (`isBuyer=true`): accumulate `quoteQty` as cost
- For sells (`isBuyer=false`): reduce cost basis proportionally
- Account for `commission` in cost calculation

### 4. Compute P&L
- **Unrealized P&L** = current_value - cost_basis (for held positions)
- **Realized P&L** = sell_proceeds - cost_basis_of_sold (from historical trades)
- **Total P&L** = unrealized + realized

### 5. Pagination for Full History
- `myTrades` returns max 1000 per call; use `fromId` to paginate
- `allOrders` time range limited to 24h; iterate with `startTime`/`endTime` windows
