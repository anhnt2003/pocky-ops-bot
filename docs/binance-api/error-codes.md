# Binance Spot REST API - Error Codes

> Source: https://developers.binance.com/docs/binance-spot-api-docs/errors
> Fetched: 2026-04-02
> Content type: API Reference

Complete list of error codes returned by the Binance Spot API. Use this to implement error classification (retryable vs non-retryable) in the API client.

---

## Error Response Format

```json
{
  "code": -1121,
  "msg": "Invalid symbol."
}
```

---

## 10xx — General Server / Network Issues

| Code | Name | Description | Retryable |
|------|------|-------------|-----------|
| -1000 | UNKNOWN | An unknown error occurred while processing the request | Yes |
| -1001 | DISCONNECTED | Internal error; unable to process your request. Please try again | Yes |
| -1002 | UNAUTHORIZED | You are not authorized to execute this request | No |
| -1003 | TOO_MANY_REQUESTS | Rate limit exceeded (queued, weight limit, or IP ban) | Yes (backoff) |
| -1006 | UNEXPECTED_RESP | Unexpected response from message bus | Yes |
| -1007 | TIMEOUT | Timeout waiting for response from backend server | Yes |
| -1008 | SERVER_BUSY | Server overloaded with other requests | Yes (backoff) |
| -1013 | INVALID_MESSAGE | Request rejected by the API | No |
| -1014 | UNKNOWN_ORDER_COMPOSITION | Unsupported order combination | No |
| -1015 | TOO_MANY_ORDERS | Order placement limit exceeded | Yes (backoff) |
| -1016 | SERVICE_SHUTTING_DOWN | This service is no longer available | No |
| -1020 | UNSUPPORTED_OPERATION | This operation is not supported | No |
| -1021 | INVALID_TIMESTAMP | Timestamp outside recvWindow or ahead of server time | No (fix clock) |
| -1022 | INVALID_SIGNATURE | Signature for this request is not valid | No (fix signing) |
| -1033 | COMP_ID_IN_USE | SenderCompId(49) is currently in use | No |
| -1034 | TOO_MANY_CONNECTIONS | Connection limit exceeded | Yes (backoff) |
| -1035 | LOGGED_OUT | Please send Logout message to close the session | No |

---

## 11xx — Request Issues

| Code | Name | Description |
|------|------|-------------|
| -1100 | ILLEGAL_CHARS | Illegal characters found in a parameter |
| -1101 | TOO_MANY_PARAMETERS | Excessive or duplicate parameters sent |
| -1102 | MANDATORY_PARAM_EMPTY_OR_MALFORMED | Required parameter missing, empty, null, or malformed |
| -1103 | UNKNOWN_PARAM | An unknown parameter was sent |
| -1104 | UNREAD_PARAMETERS | Not all sent parameters were read |
| -1105 | PARAM_EMPTY | A parameter was empty |
| -1106 | PARAM_NOT_REQUIRED | A parameter was sent when not required |
| -1108 | PARAM_OVERFLOW | Parameter '%s' overflowed |
| -1111 | BAD_PRECISION | Parameter '%s' has too much precision |
| -1112 | NO_DEPTH | No orders on book for symbol |
| -1114 | TIF_NOT_REQUIRED | TimeInForce parameter sent when not required |
| -1115 | INVALID_TIF | Invalid timeInForce |
| -1116 | INVALID_ORDER_TYPE | Invalid orderType |
| -1117 | INVALID_SIDE | Invalid side |
| -1118 | EMPTY_NEW_CL_ORD_ID | New client order ID was empty |
| -1119 | EMPTY_ORG_CL_ORD_ID | Original client order ID was empty |
| -1120 | BAD_INTERVAL | Invalid interval |
| -1121 | BAD_SYMBOL | Invalid symbol |
| -1122 | INVALID_SYMBOLSTATUS | Invalid symbolStatus |
| -1125 | INVALID_LISTEN_KEY | This listenKey does not exist |
| -1127 | MORE_THAN_XX_HOURS | Lookup interval is too big |
| -1128 | OPTIONAL_PARAMS_BAD_COMBO | Combination of optional parameters invalid |
| -1130 | INVALID_PARAMETER | Invalid data sent for a parameter |
| -1134 | BAD_STRATEGY_TYPE | strategyType was less than 1000000 |
| -1135 | INVALID_JSON | Invalid JSON Request |
| -1139 | INVALID_TICKER_TYPE | Invalid ticker type |
| -1145 | INVALID_CANCEL_RESTRICTIONS | cancelRestrictions must be `ONLY_NEW` or `ONLY_PARTIALLY_FILLED` |
| -1151 | DUPLICATE_SYMBOLS | Symbol is present multiple times in the list |
| -1152 | INVALID_SBE_HEADER | Invalid X-MBX-SBE header |
| -1153 | UNSUPPORTED_SCHEMA_ID | Unsupported SBE schema ID or version |
| -1155 | SBE_DISABLED | SBE is not enabled |
| -1158 | OCO_ORDER_TYPE_REJECTED | Order type not supported in OCO |
| -1160 | OCO_ICEBERGQTY_TIMEINFORCE | Parameter not supported with specified timeInForce |
| -1161 | DEPRECATED_SCHEMA | Unable to encode the response in SBE schema |
| -1165 | BUY_OCO_LIMIT_MUST_BE_BELOW | A limit order in a buy OCO must be below |
| -1166 | SELL_OCO_LIMIT_MUST_BE_ABOVE | A limit order in a sell OCO must be above |
| -1168 | BOTH_OCO_ORDERS_CANNOT_BE_LIMIT | At least one OCO order must be contingent |
| -1169 | INVALID_TAG_NUMBER | Invalid tag number |
| -1170 | TAG_NOT_DEFINED_IN_MESSAGE | Tag '%s' not defined for this message type |
| -1171 | TAG_APPEARS_MORE_THAN_ONCE | Tag '%s' appears more than once |
| -1172 | TAG_OUT_OF_ORDER | Tag '%s' specified out of required order |
| -1173 | GROUP_FIELDS_OUT_OF_ORDER | Repeating group '%s' fields out of order |
| -1174 | INVALID_COMPONENT | Component '%s' is incorrectly populated |
| -1175 | RESET_SEQ_NUM_SUPPORT | Continuation of sequence numbers to new session is unsupported |
| -1176 | ALREADY_LOGGED_IN | Logon should only be sent once |
| -1177 | GARBLED_MESSAGE | Various message format errors |
| -1178 | BAD_SENDER_COMPID | SenderCompId(49) contains an incorrect value |
| -1179 | BAD_SEQ_NUM | MsgSeqNum(34) contains an unexpected value |
| -1180 | EXPECTED_LOGON | Logon must be the first message in the session |
| -1181 | TOO_MANY_MESSAGES | Too many messages; current limit exceeded |
| -1182 | PARAMS_BAD_COMBO | Conflicting fields: [%s] |
| -1183 | NOT_ALLOWED_IN_DROP_COPY_SESSIONS | Requested operation not allowed in DropCopy sessions |
| -1184 | DROP_COPY_SESSION_NOT_ALLOWED | DropCopy sessions not supported on this server |
| -1185 | DROP_COPY_SESSION_REQUIRED | Only DropCopy sessions supported on this server |
| -1186 | NOT_ALLOWED_IN_ORDER_ENTRY_SESSIONS | Requested operation not allowed in order entry sessions |
| -1187 | NOT_ALLOWED_IN_MARKET_DATA_SESSIONS | Requested operation not allowed in market data sessions |
| -1188 | INCORRECT_NUM_IN_GROUP_COUNT | Incorrect NumInGroup count for repeating group |
| -1189 | DUPLICATE_ENTRIES_IN_A_GROUP | Group '%s' contains duplicate entries |
| -1190 | INVALID_REQUEST_ID | MDReqID contains invalid subscription request |
| -1191 | TOO_MANY_SUBSCRIPTIONS | Too many subscriptions |
| -1194 | INVALID_TIME_UNIT | Invalid value for time unit |
| -1196 | BUY_OCO_STOP_LOSS_MUST_BE_ABOVE | A stop loss order in a buy OCO must be above |
| -1197 | SELL_OCO_STOP_LOSS_MUST_BE_BELOW | A stop loss order in a sell OCO must be below |
| -1198 | BUY_OCO_TAKE_PROFIT_MUST_BE_BELOW | A take profit order in a buy OCO must be below |
| -1199 | SELL_OCO_TAKE_PROFIT_MUST_BE_ABOVE | A take profit order in a sell OCO must be above |
| -1210 | INVALID_PEG_PRICE_TYPE | Invalid pegPriceType |
| -1211 | INVALID_PEG_OFFSET_TYPE | Invalid pegOffsetType |
| -1220 | SYMBOL_DOES_NOT_MATCH_STATUS | The symbol's status does not match requested status |
| -1221 | INVALID_SBE_MESSAGE_FIELD | Invalid/missing field(s) in SBE message |
| -1222 | OPO_WORKING_MUST_BE_BUY | Working order in an OPO list must be a bid |
| -1223 | OPO_PENDING_MUST_BE_SELL | Pending orders in an OPO list must be asks |
| -1224 | WORKING_PARAM_REQUIRED | Working order must include the '{param}' tag |
| -1225 | PENDING_PARAM_NOT_REQUIRED | Pending orders should not include the '%s' tag |

> All 11xx errors are **non-retryable** — fix the request parameters.

---

## 20xx — Order-Related Errors

| Code | Name | Description |
|------|------|-------------|
| -2010 | NEW_ORDER_REJECTED | Order rejected (see rejection messages below) |
| -2011 | CANCEL_REJECTED | Cancel rejected |
| -2013 | NO_SUCH_ORDER | Order does not exist |
| -2014 | BAD_API_KEY_FMT | API-key format invalid |
| -2015 | REJECTED_MBX_KEY | Invalid API-key, IP, or permissions for action |
| -2016 | NO_TRADING_WINDOW | No trading window could be found for the symbol |
| -2021 | ORDER_CANCEL_REPLACE_PARTIALLY_FAILED | Either cancellation or placement failed but not both |
| -2022 | ORDER_CANCEL_REPLACE_FAILED | Both cancellation and placement failed |
| -2026 | ORDER_ARCHIVED | Order was canceled or expired over 90 days ago |
| -2035 | SUBSCRIPTION_ACTIVE | User Data Stream subscription already active |
| -2036 | SUBSCRIPTION_INACTIVE | User Data Stream subscription not active |
| -2039 | CLIENT_ORDER_ID_INVALID | Client order ID is not correct for this order |
| -2042 | MAXIMUM_SUBSCRIPTION_IDS | Maximum subscription ID reached for this connection |
| -2043 | NO_REFERENCE_PRICE | This symbol doesn't have a reference price |

---

## Order Rejection Messages (with -2010, -2011)

These are the `msg` values returned when orders are rejected:

- "Unknown order sent."
- "Duplicate order sent."
- "Market is closed."
- "Account has insufficient balance for requested action."
- "Market orders are not supported for this symbol."
- "Iceberg orders are not supported for this symbol."
- "Stop loss orders are not supported for this symbol."
- "Stop loss limit orders are not supported for this symbol."
- "Take profit orders are not supported for this symbol."
- "Take profit limit orders are not supported for this symbol."
- "Order amend is not supported for this symbol."
- "Price * QTY is zero or less."
- "IcebergQty exceeds QTY."
- "This action is disabled on this account."
- "This account may not place or cancel orders."
- "Unsupported order combination"
- "Order would trigger immediately."
- "Cancel order is invalid. Check origClOrdId and orderId."
- "Order would immediately match and take."
- "The relationship of the prices for the orders is not correct."
- "OCO orders are not supported for this symbol"
- "Quote order qty market orders are not support for this symbol."
- "Trailing stop orders are not supported for this symbol."
- "Order cancel-replace is not supported for this symbol."
- "This symbol is not permitted for this account."
- "This symbol is restricted for this account."
- "Order was not canceled due to cancel restrictions."
- "Rest API trading is not enabled." / "WebSocket API trading is not enabled." / "FIX API trading is not enabled."
- "Order book liquidity is less than LOT_SIZE filter minimum."
- "Order book liquidity is less than MARKET_LOT_SIZE filter minimum."
- "Order book liquidity is less than symbol minimum quantity."
- "Order amend (quantity increase) is not supported."
- "The requested action would change no state; rejecting."
- "Pegged orders are not supported for this symbol."
- "This price peg cannot be used with this order type."
- "Order book liquidity is too low for this pegged order."
- "OPO orders are not supported for this symbol."
- "Order amend (pending OPO order) is not supported."

---

## Filter Failure Messages

Returned when order parameters violate symbol/exchange filters:

| Filter | Message |
|--------|---------|
| PRICE_FILTER | price is too high, too low, and/or not following tick size |
| PERCENT_PRICE | price is X% too high or low from average weighted price |
| LOT_SIZE | quantity is too high, too low, and/or not following step size |
| MIN_NOTIONAL | price * quantity is too low for valid order |
| NOTIONAL | price * quantity not within minNotional/maxNotional range |
| ICEBERG_PARTS | ICEBERG order would break into too many parts |
| MARKET_LOT_SIZE | MARKET order's quantity too high, low, or violates step size |
| MAX_POSITION | Account's position reached maximum defined limit |
| MAX_NUM_ORDERS | Account has too many open orders on the symbol |
| MAX_NUM_ALGO_ORDERS | Account has too many stop loss/take profit orders |
| MAX_NUM_ICEBERG_ORDERS | Account has too many open iceberg orders |
| MAX_NUM_ORDER_AMENDS | Account made too many amendments to single order |
| MAX_NUM_ORDER_LISTS | Account has too many open order lists |
| TRAILING_DELTA | trailingDelta not within defined range for order type |
| EXCHANGE_MAX_NUM_ORDERS | Account has too many open orders on exchange |
| EXCHANGE_MAX_NUM_ALGO_ORDERS | Account has too many algo orders on exchange |
| EXCHANGE_MAX_NUM_ICEBERG_ORDERS | Account has too many iceberg orders on exchange |
| EXCHANGE_MAX_NUM_ORDER_LISTS | Account has too many order lists on exchange |

---

## Error Classification for Client Implementation

### Retryable (with exponential backoff)
- `-1000` (UNKNOWN)
- `-1001` (DISCONNECTED)
- `-1003` (TOO_MANY_REQUESTS) — respect `Retry-After` header
- `-1006` (UNEXPECTED_RESP)
- `-1007` (TIMEOUT)
- `-1008` (SERVER_BUSY)
- `-1015` (TOO_MANY_ORDERS) — wait for rate limit reset
- `-1034` (TOO_MANY_CONNECTIONS)

### Non-retryable (fix request)
- All `-11xx` codes — parameter/validation errors
- `-1002` (UNAUTHORIZED)
- `-1021` (INVALID_TIMESTAMP) — sync clock with server
- `-1022` (INVALID_SIGNATURE) — fix signing logic
- `-2010` through `-2043` — order/business logic errors
