# Polymarket API — Complete Reference Documentation

> Compiled: 2026-02-24. Sources: docs.polymarket.com, GitHub (py-clob-client, real-time-data-client), NautilusTrader docs, community guides.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Base URLs](#2-base-urls)
3. [Authentication](#3-authentication)
   - [L1 — EIP-712 Wallet Signature](#31-l1--eip-712-wallet-signature)
   - [L2 — HMAC-SHA256 API Key](#32-l2--hmac-sha256-api-key)
   - [API Key Creation Endpoints](#33-api-key-creation-endpoints)
   - [Request Headers Summary](#34-request-headers-summary)
4. [Rate Limits](#4-rate-limits)
5. [CLOB API — Market Data Endpoints (Public)](#5-clob-api--market-data-endpoints-public)
6. [CLOB API — Trading Endpoints (Auth Required)](#6-clob-api--trading-endpoints-auth-required)
7. [CLOB API — Order Model](#7-clob-api--order-model)
8. [CLOB API — Trade Model](#8-clob-api--trade-model)
9. [CLOB API — Order Types & Statuses](#9-clob-api--order-types--statuses)
10. [CLOB API — Heartbeat](#10-clob-api--heartbeat)
11. [CLOB API — L1 Client Methods](#11-clob-api--l1-client-methods)
12. [CLOB API — L2 Client Methods](#12-clob-api--l2-client-methods)
13. [Gamma API — Endpoints](#13-gamma-api--endpoints)
14. [Gamma API — Query Parameters](#14-gamma-api--query-parameters)
15. [Gamma API — Data Models](#15-gamma-api--data-models)
16. [Data API — Endpoints](#16-data-api--endpoints)
17. [Bridge API](#17-bridge-api)
18. [WebSocket — Market Channel](#18-websocket--market-channel)
19. [WebSocket — User Channel](#19-websocket--user-channel)
20. [WebSocket — Sports Channel](#20-websocket--sports-channel)
21. [WebSocket — RTDS (Real-Time Data Streaming)](#21-websocket--rtds-real-time-data-streaming)
22. [Smart Contract Addresses](#22-smart-contract-addresses)
23. [Error Codes](#23-error-codes)
24. [Negative Risk Markets](#24-negative-risk-markets)
25. [SDK Quick Reference](#25-sdk-quick-reference)

---

## 1. Architecture Overview

Polymarket operates a **hybrid-decentralized trading system**:

- **Off-chain order matching**: An operator hosts the order book and matches orders.
- **On-chain settlement**: Final settlement executes via an audited Exchange contract on Polygon (chain ID 137).
- **Non-custodial**: Orders are EIP-712 signed messages. Users retain independent on-chain cancellation rights. The operator cannot manipulate prices or execute unauthorized trades.
- **Settlement currency**: USDC.e (bridged USDC on Polygon)

### API Components

| Component | Purpose |
|-----------|---------|
| **Gamma API** | Market metadata, discovery, events, tags, sports |
| **CLOB API** | Order book, pricing, order placement/cancellation |
| **Data API** | User positions, trades, activity, leaderboards |
| **Bridge API** | Deposits and withdrawals (proxy for fun.xyz) |
| **WebSocket** | Real-time orderbook, prices, user events |
| **RTDS** | Real-time data streaming (trades, comments, RFQ, crypto prices) |

### Wallet / Signature Types

| Type | Value | Description |
|------|-------|-------------|
| EOA | 0 | Standard wallet (MetaMask, hardware), pays own gas |
| POLY_PROXY | 1 | Magic Link / email wallet, requires exported keys |
| GNOSIS_SAFE | 2 | Browser or embedded wallet (most common for Polymarket users) |

---

## 2. Base URLs

| Service | Base URL |
|---------|---------|
| CLOB API | `https://clob.polymarket.com` |
| Gamma API | `https://gamma-api.polymarket.com` |
| Data API | `https://data-api.polymarket.com` |
| Bridge API | `https://bridge.polymarket.com` |
| WebSocket (Market/User) | `wss://ws-subscriptions-clob.polymarket.com/ws/` |
| WebSocket (Sports) | `wss://sports-api.polymarket.com/ws` |
| WebSocket (RTDS) | `wss://ws-live-data.polymarket.com` |

---

## 3. Authentication

Polymarket uses a **two-tier authentication** system. Public endpoints (orderbook, pricing, Gamma market data) require no authentication.

### 3.1 L1 — EIP-712 Wallet Signature

Used for:
- Creating/deriving API keys
- Proving wallet ownership
- Signing orders locally

**EIP-712 Domain:**
```json
{
  "name": "ClobAuthDomain",
  "version": "1",
  "chainId": 137
}
```

**EIP-712 Types:**
```json
{
  "ClobAuth": [
    { "name": "address",   "type": "address" },
    { "name": "timestamp", "type": "string"  },
    { "name": "nonce",     "type": "uint256" },
    { "name": "message",   "type": "string"  }
  ]
}
```

**Value to sign:**
```json
{
  "address":   "<signingAddress>",
  "timestamp": "<unix_timestamp_string>",
  "nonce":     <nonce_uint256>,
  "message":   "This message attests that I control the given wallet"
}
```

**TypeScript signing example:**
```typescript
const sig = await signer._signTypedData(domain, types, value);
```

**L1 Request Headers:**

| Header | Value |
|--------|-------|
| `POLY_ADDRESS` | Signer's Ethereum address |
| `POLY_SIGNATURE` | EIP-712 signature |
| `POLY_TIMESTAMP` | Unix timestamp (seconds) |
| `POLY_NONCE` | Nonce integer used in signing |

### 3.2 L2 — HMAC-SHA256 API Key

Used for all trading endpoints once API credentials are obtained.

**Credentials object:**
```json
{
  "apiKey":     "550e8400-e29b-41d4-a716-446655440000",
  "secret":     "<base64EncodedSecret>",
  "passphrase": "<randomPassphrase>"
}
```

Requests are signed using HMAC-SHA256 with the `secret`. Requests expire after **30 seconds** (timestamp must be within 30s of server time).

**Each wallet can have only one active API key.** Creating a new key invalidates the previous one.

**L2 Request Headers:**

| Header | Value |
|--------|-------|
| `POLY_ADDRESS` | Signer's Ethereum address |
| `POLY_SIGNATURE` | HMAC-SHA256 signature of payload |
| `POLY_TIMESTAMP` | Unix timestamp (seconds) |
| `POLY_API_KEY` | API key (UUID) |
| `POLY_PASSPHRASE` | Passphrase value |

### 3.3 API Key Creation Endpoints

| Method | Endpoint | Auth | Description |
|--------|---------|------|-------------|
| `POST` | `/auth/api-key` | L1 headers | Create a new API key |
| `GET` | `/auth/derive-api-key` | L1 headers | Derive existing key (by nonce) |
| `GET` | `/auth/api-key` | L2 headers | Get all API keys for account |
| `DELETE` | `/auth/api-key` | L2 headers | Delete/revoke current API key |

**SDK equivalents:**
```python
# Python
creds = client.create_or_derive_api_creds()   # recommended
creds = client.create_api_creds()
creds = client.derive_api_creds()
```
```typescript
// TypeScript
const creds = await client.createOrDeriveApiKey();
const creds = await client.createApiKey(nonce?);
const creds = await client.deriveApiKey(nonce?);
```

### 3.4 Request Headers Summary

| Scenario | Required Headers |
|----------|----------------|
| Public (no auth) | None |
| L1 (key creation) | `POLY_ADDRESS`, `POLY_SIGNATURE`, `POLY_TIMESTAMP`, `POLY_NONCE` |
| L2 (trading) | `POLY_ADDRESS`, `POLY_SIGNATURE`, `POLY_TIMESTAMP`, `POLY_API_KEY`, `POLY_PASSPHRASE` |

---

## 4. Rate Limits

Rate limits are enforced via **Cloudflare throttling** (sliding windows). Exceeding limits results in HTTP **429**. Implement exponential backoff. Prefer WebSocket for real-time data over polling.

### CLOB API (`https://clob.polymarket.com`)

#### General
| Endpoint | Limit |
|---------|-------|
| General (catch-all) | 9,000 req / 10s |
| `GET /ok` (health check) | 100 req / 10s |
| GET balance allowance | 200 req / 10s |
| UPDATE balance allowance | 50 req / 10s |

#### Market Data
| Endpoint | Limit |
|---------|-------|
| `GET /book` | 1,500 req / 10s |
| `GET /price` | 1,500 req / 10s |
| `GET /midpoint` | 1,500 req / 10s |
| `POST /books` (batch) | 500 req / 10s |
| `POST /prices` (batch) | 500 req / 10s |
| `POST /midpoints` (batch) | 500 req / 10s |
| `GET /prices-history` | 1,000 req / 10s |
| Market tick size | 200 req / 10s |

#### Ledger / Orders / Trades
| Endpoint | Limit |
|---------|-------|
| `GET /trades` | 900 req / 10s |
| `GET /orders` | 900 req / 10s |
| `GET /notifications` | 125 req / 10s |
| `GET /order` (single) | 900 req / 10s |
| `GET /data/orders` | 500 req / 10s |
| `GET /data/trades` | 500 req / 10s |

#### Authentication
| Endpoint | Limit |
|---------|-------|
| API key endpoints | 100 req / 10s |

#### Trading (Burst + Sustained)
| Endpoint | Burst (per 10s) | Sustained (per 10 min) |
|---------|----------------|----------------------|
| `POST /order` | 3,500 req | 36,000 req |
| `DELETE /order` | 3,000 req | 30,000 req |
| `POST /orders` (batch) | 1,000 req | 15,000 req |
| `DELETE /orders` (batch) | 1,000 req | 15,000 req |
| `DELETE /cancel-all` | 250 req | 6,000 req |
| `DELETE /cancel-market-orders` | 1,000 req | 1,500 req |

### Gamma API (`https://gamma-api.polymarket.com`)

| Endpoint | Limit |
|---------|-------|
| General | 4,000 req / 10s |
| `GET /events` | 500 req / 10s |
| `GET /markets` | 300 req / 10s |
| `/markets` + `/events` listing combined | 900 req / 10s |
| `GET /comments` | 200 req / 10s |
| `GET /tags` | 200 req / 10s |
| `GET /public-search` | 350 req / 10s |

### Data API (`https://data-api.polymarket.com`)

| Endpoint | Limit |
|---------|-------|
| General | 1,000 req / 10s |
| `GET /trades` | 200 req / 10s |
| `GET /positions` | 150 req / 10s |
| `GET /closed-positions` | 150 req / 10s |
| `GET /ok` (health check) | 100 req / 10s |

### Other
| Endpoint | Limit |
|---------|-------|
| Relayer `/submit` | 25 req / min |
| User PNL API | 200 req / 10s |

---

## 5. CLOB API — Market Data Endpoints (Public)

Base URL: `https://clob.polymarket.com`
No authentication required.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/ok` | Health check |
| `GET` | `/time` | Server timestamp |
| `GET` | `/book` | Order book for a single token |
| `POST` | `/books` | Order books for multiple tokens (batch) |
| `GET` | `/price` | Best price for a token (BUY or SELL side) |
| `POST` | `/prices` | Prices for multiple tokens (batch) |
| `GET` | `/midpoint` | Midpoint price for a token |
| `POST` | `/midpoints` | Midpoints for multiple tokens (batch) |
| `GET` | `/spread` | Bid-ask spread for a token |
| `GET` | `/prices-history` | Historical price data for a token |
| `GET` | `/tick-size` | Tick size for a market |
| `GET` | `/neg-risk` | Whether a market is negative-risk |
| `GET` | `/trades` | Trade history (public) |
| `GET` | `/markets` | Market info by condition ID |
| `GET` | `/simplified-markets` | Simplified market listing |
| `GET` | `/sampling-markets` | Sampling market data |
| `GET` | `/sampling-simplified-markets` | Simplified sampling markets |

### GET /book — Order Book

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token_id` | string | Yes | Token/asset ID |

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `market` | string | Condition ID of the market |
| `asset_id` | string | Token ID |
| `timestamp` | string | Snapshot timestamp |
| `hash` | string | Order book hash |
| `bids` | array | Bid levels `[{price, size}]` |
| `asks` | array | Ask levels `[{price, size}]` |
| `min_order_size` | string | Minimum order size |
| `tick_size` | string | Price tick size |
| `neg_risk` | boolean | Whether market is negative-risk |

### GET /price

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token_id` | string | Yes | Token/asset ID |
| `side` | string | Yes | `"BUY"` or `"SELL"` |

### GET /midpoint

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token_id` | string | Yes | Token/asset ID |

### GET /prices-history

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `market` | string | Condition ID |
| `startTs` | int | Start timestamp (Unix) |
| `endTs` | int | End timestamp (Unix) |
| `fidelity` | int | Interval in minutes |

---

## 6. CLOB API — Trading Endpoints (Auth Required)

Base URL: `https://clob.polymarket.com`
All require L2 authentication headers unless noted.

| Method | Path | Auth Level | Description |
|--------|------|------------|-------------|
| `POST` | `/auth/api-key` | L1 | Create API key |
| `GET` | `/auth/derive-api-key` | L1 | Derive existing API key |
| `GET` | `/auth/api-key` | L2 | List API keys |
| `DELETE` | `/auth/api-key` | L2 | Delete API key |
| `POST` | `/order` | L2 | Place single order |
| `POST` | `/orders` | L2 | Place batch orders (up to 15) |
| `DELETE` | `/order` | L2 | Cancel single order |
| `DELETE` | `/orders` | L2 | Cancel multiple orders |
| `DELETE` | `/cancel-all` | L2 | Cancel all open orders |
| `DELETE` | `/cancel-market-orders` | L2 | Cancel orders for a specific market |
| `GET` | `/order` | L2 | Get single order by ID |
| `GET` | `/orders` | L2 | Get open orders (filtered) |
| `GET` | `/data/orders` | L2 | Order history |
| `GET` | `/trades` | L2 | User trade history |
| `GET` | `/data/trades` | L2 | Trade history (paginated) |
| `GET` | `/positions` | L2 | Current open positions |
| `GET` | `/balance-allowance` | L2 | Balance and allowance check |
| `POST` | `/balance-allowance` | L2 | Update cached balance/allowance |
| `GET` | `/notifications` | L2 | User event notifications |
| `DELETE` | `/notifications` | L2 | Dismiss notifications |
| `POST` | `/heartbeat` | L2 | Session keep-alive |

### POST /order — Place Single Order

**Request Body (JSON):**
```json
{
  "order": {
    "salt":          "<random_salt>",
    "maker":         "<maker_address>",
    "signer":        "<signer_address>",
    "taker":         "0x0000000000000000000000000000000000000000",
    "tokenId":       "<token_id>",
    "makerAmount":   "<amount_in_usdc_base_units>",
    "takerAmount":   "<amount_in_shares>",
    "expiration":    "<unix_timestamp_or_0>",
    "nonce":         "<nonce>",
    "feeRateBps":    "<fee_rate>",
    "side":          0,
    "signatureType": 0,
    "signature":     "<eip712_signature>"
  },
  "owner":     "<api_key>",
  "orderType": "GTC"
}
```

**Order Parameters (before signing):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tokenID` | string | Yes | Market token ID |
| `price` | string | Yes | Limit price (0.0001 to 0.9999), must match tick size |
| `size` | string | Yes | Number of shares (limit) or USD amount (market BUY) |
| `side` | string | Yes | `"BUY"` or `"SELL"` |
| `orderType` | string | No | `"GTC"` (default), `"GTD"`, `"FOK"`, `"FAK"` |
| `expiration` | int | No | Unix timestamp; required for GTD |
| `feeRateBps` | string | No | Custom fee in basis points |
| `nonce` | string | No | Custom nonce for uniqueness |
| `taker` | string | No | Restrict order to specific taker address |
| `tickSize` | string | Yes (option) | `"0.1"`, `"0.01"`, `"0.001"`, `"0.0001"` |
| `negRisk` | boolean | Yes (option) | `true` for 3+ outcome markets |
| `postOnly` | boolean | No | Reject if order would immediately match |

**Response:**
```json
{
  "success":          true,
  "errorMsg":         "",
  "orderID":          "0xabc123...",
  "takingAmount":     "",
  "makingAmount":     "",
  "status":           "live",
  "transactionsHashes": [],
  "tradeIDs":         []
}
```

**Insert Status Values:**

| Status | Description |
|--------|-------------|
| `live` | Order resting on the book |
| `matched` | Immediately matched with existing order |
| `delayed` | Marketable, subject to matching delay |
| `unmatched` | Marketable but delay failed; placement succeeded |

### POST /orders — Batch Orders

- Up to **15 orders** per request (increased from 5 in 2025).
- Request body: array of order objects with same structure as single order.

### DELETE /order — Cancel Single Order

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `orderID` | string | Yes | Order ID to cancel |

### GET /orders — Open Orders

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Filter by order ID |
| `market` | string | Filter by condition ID |
| `asset_id` | string | Filter by token ID |

---

## 7. CLOB API — Order Model

### OpenOrder Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Order identifier |
| `status` | string | Current state (`live`, `matched`, etc.) |
| `market` | string | Condition ID |
| `asset_id` | string | Token ID |
| `side` | string | `"BUY"` or `"SELL"` |
| `original_size` | string | Initial placement size (shares) |
| `size_matched` | string | Amount already filled |
| `price` | string | Limit price |
| `outcome` | string | Human-readable outcome label |
| `order_type` | string | `"GTC"`, `"GTD"`, `"FOK"`, `"FAK"` |
| `maker_address` | string | Funder wallet address |
| `owner` | string | API key that created the order |
| `expiration` | string | Unix timestamp (`"0"` if none) |
| `associate_trades` | string[] | IDs of related trades |
| `created_at` | string | Creation timestamp |

---

## 8. CLOB API — Trade Model

### Trade Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Trade identifier |
| `taker_order_id` | string | Taker's order ID |
| `market` | string | Condition ID |
| `asset_id` | string | Token ID |
| `side` | string | Trade side |
| `size` | string | Trade size (shares) |
| `fee_rate_bps` | string | Fee rate in basis points |
| `price` | string | Execution price |
| `status` | string | Trade lifecycle status |
| `match_time` | string | Match timestamp |
| `last_update` | string | Last status update timestamp |
| `outcome` | string | Human-readable outcome |
| `owner` | string | API key |
| `maker_address` | string | Maker's wallet address |
| `trader_side` | string | Whether user is maker or taker |
| `transaction_hash` | string | On-chain transaction hash |
| `maker_orders` | array | Array of MakerOrder objects |

### MakerOrder Object (nested in Trade)

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Maker's order ID |
| `owner` | string | Maker's API key |
| `maker_address` | string | Maker's wallet address |
| `matched_amount` | string | Amount matched against this maker order |
| `price` | string | Maker's limit price |
| `fee_rate_bps` | string | Maker's fee rate |
| `asset_id` | string | Token ID |
| `outcome` | string | Outcome label |
| `side` | string | Maker's side |

---

## 9. CLOB API — Order Types & Statuses

### Order Types

| Type | Behavior |
|------|----------|
| `GTC` | Good-Till-Cancelled — rests on book until filled or cancelled (default) |
| `GTD` | Good-Till-Date — expires at specified Unix timestamp unless filled/cancelled earlier |
| `FOK` | Fill-Or-Kill — entire order must fill immediately or whole order is cancelled |
| `FAK` | Fill-And-Kill — fills as many shares as available immediately, cancels remainder |

**GTD Security Threshold:** Add 60 seconds minimum buffer. For 90-second life, set `expiration = now + 60 + 30`.

**Post-Only Orders:**
- Rejected if they would cross the spread and immediately match.
- Cannot combine with `FOK` or `FAK`.
- Only valid with `GTC` or `GTD`.

### Trade/Order Lifecycle Statuses

| Status | Terminal | Description |
|--------|----------|-------------|
| `MATCHED` | No | Matched and sent to executor for on-chain submission |
| `MINED` | No | Observed as mined, no finality threshold yet |
| `CONFIRMED` | Yes | Strong probabilistic finality — trade successful |
| `RETRYING` | No | Transaction failed (revert/reorg), being retried |
| `FAILED` | Yes | Permanently failed, not being retried |

### Tick Size Rules

Price must conform to the market's tick size or the order is rejected.

| Tick Size | Valid Price Examples |
|-----------|---------------------|
| `0.1` | 0.1, 0.2, ..., 0.9 |
| `0.01` | 0.01, 0.50, 0.99 |
| `0.001` | 0.001, 0.500, 0.999 |
| `0.0001` | 0.0001, 0.5000, 0.9999 |

### Balance Constraints

```
maxOrderSize = underlyingAssetBalance - Σ(orderSize - orderFillAmount)
```

- **BUY orders**: Require USDC.e allowance ≥ spend amount.
- **SELL orders**: Require conditional token allowance ≥ sell amount.

---

## 10. CLOB API — Heartbeat

To prevent automatic cancellation of all open orders, send a heartbeat within **10 seconds** (5-second buffer recommended).

**Endpoint:** `POST /heartbeat`

**Request:**
```json
{ "heartbeat_id": "<previous_heartbeat_id_or_empty>" }
```

**Response:**
```json
{ "heartbeat_id": "<new_heartbeat_id>" }
```

**Python example:**
```python
heartbeat_id = ""
while True:
    resp = client.post_heartbeat(heartbeat_id)
    heartbeat_id = resp["heartbeat_id"]
    time.sleep(5)
```

**TypeScript example:**
```typescript
let heartbeatId = "";
setInterval(async () => {
  const resp = await client.postHeartbeat(heartbeatId);
  heartbeatId = resp.heartbeat_id;
}, 5000);
```

---

## 11. CLOB API — L1 Client Methods

Require wallet private key; no API credentials needed.

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| `createApiKey(nonce?)` | nonce: number (optional) | `ApiKeyCreds` | Create new L2 credentials (invalidates previous key) |
| `deriveApiKey(nonce?)` | nonce: number (optional) | `ApiKeyCreds` | Retrieve existing credentials by nonce |
| `createOrDeriveApiKey(nonce?)` | nonce: number (optional) | `ApiKeyCreds` | Derive first; create if not found. **Recommended** |
| `createOrder(userOrder, options?)` | tokenID, price, size, side, [feeRateBps, nonce, expiration, taker] | `SignedOrder` | Sign limit order locally without posting |
| `createMarketOrder(userMarketOrder, options?)` | tokenID, amount, side, [price, orderType] | `SignedOrder` | Sign market order locally |

---

## 12. CLOB API — L2 Client Methods

Require API credentials (apiKey, secret, passphrase).

| Method | Corresponding REST | Description |
|--------|------------------|-------------|
| `createAndPostOrder(args, options, orderType?)` | `POST /order` | Create, sign, and post limit order in one call |
| `createAndPostMarketOrder(args, options, orderType?)` | `POST /order` | Create, sign, and post market order |
| `postOrder(signedOrder, orderType, postOnly?)` | `POST /order` | Post pre-signed order |
| `postOrders(orders[])` | `POST /orders` | Post up to 15 pre-signed orders |
| `cancelOrder(orderId)` | `DELETE /order` | Cancel single order |
| `cancelOrders(orderIds[])` | `DELETE /orders` | Cancel multiple orders |
| `cancelAll()` | `DELETE /cancel-all` | Cancel all open orders |
| `cancelMarketOrders(market?, assetId?)` | `DELETE /cancel-market-orders` | Cancel orders for market/asset |
| `getOrder(orderId)` | `GET /order` | Get single order details |
| `getOpenOrders(filters?)` | `GET /orders` | List open orders (filter by id/market/asset) |
| `getTrades(filters?)` | `GET /trades` | User trade history |
| `getTradesPaginated(filters?)` | `GET /data/trades` | Trade history with pagination |
| `getBalanceAllowance(params)` | `GET /balance-allowance` | Balance and allowance for token |
| `updateBalanceAllowance(params)` | `POST /balance-allowance` | Update cached balance/allowance |
| `getApiKeys()` | `GET /auth/api-key` | List all API keys for account |
| `deleteApiKey()` | `DELETE /auth/api-key` | Revoke current API key |
| `getNotifications()` | `GET /notifications` | Get event notifications (expire in 48h) |
| `dropNotifications(ids[])` | `DELETE /notifications` | Dismiss notifications |
| `postHeartbeat(heartbeatId)` | `POST /heartbeat` | Keep session alive |

### Notification Types
- Order cancellation
- Order fills
- Market resolution

---

## 13. Gamma API — Endpoints

Base URL: `https://gamma-api.polymarket.com`
**No authentication required.**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/events` | List events (with filtering and pagination) |
| `GET` | `/events/{id}` | Get event by ID |
| `GET` | `/events?slug={slug}` | Get event by slug (query param) |
| `GET` | `/events/slug/{slug}` | Get event by slug (path param) |
| `GET` | `/markets` | List markets (with filtering and pagination) |
| `GET` | `/markets/{id}` | Get market by ID |
| `GET` | `/markets?slug={slug}` | Get market by slug (query param) |
| `GET` | `/markets/slug/{slug}` | Get market by slug (path param) |
| `GET` | `/tags` | List all tags/categories |
| `GET` | `/series` | List series (grouped events) |
| `GET` | `/sports` | Sports metadata (images, resolution sources, tag IDs) |
| `GET` | `/teams` | Team information |
| `GET` | `/public-search` | Search across events, markets, profiles |

---

## 14. Gamma API — Query Parameters

### Common Parameters (Events & Markets)

| Parameter | Type | Description |
|-----------|------|-------------|
| `slug` | string | Filter by unique slug identifier |
| `tag_id` | integer | Filter by tag/category ID |
| `related_tags` | boolean | Include related tag markets |
| `exclude_tag_id` | integer | Exclude specific tag from results |
| `active` | boolean | `true` = live/tradable only |
| `closed` | boolean | `true` = include closed, `false` = exclude closed |
| `series_id` | integer | Filter by sports series (e.g., NBA) |
| `order` | string | Sort field: `volume_24hr`, `volume`, `liquidity`, `start_date`, `end_date`, `competitive`, `closed_time` |
| `ascending` | boolean | Sort direction (`false` = descending, default) |
| `limit` | integer | Results per page |
| `offset` | integer | Number of results to skip (pagination) |

### Recommended Patterns

```
# All active markets (efficient — events contain associated markets)
GET /events?active=true&closed=false&limit=100

# Highest volume events
GET /events?active=true&closed=false&order=volume_24hr&ascending=false&limit=100

# Events by tag
GET /events?tag_id=<id>&active=true&closed=false

# Events by tag with related tags
GET /events?tag_id=<id>&active=true&closed=false&related_tags=true

# Paginated markets
GET /markets?active=true&closed=false&limit=50&offset=0
GET /markets?active=true&closed=false&limit=50&offset=50

# Market by slug (from URL: polymarket.com/event/<slug>)
GET /events?slug=fed-decision-in-october
GET /events/slug/fed-decision-in-october

# Sports events by series
GET /events?series_id=<nba_series_id>&active=true&closed=false

# All tags
GET /tags?limit=100
```

---

## 15. Gamma API — Data Models

### Event Object (key fields)

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Event ID |
| `slug` | string | URL slug |
| `title` | string | Event title |
| `description` | string | Event description |
| `active` | boolean | Is currently active |
| `closed` | boolean | Is closed |
| `markets` | array | Associated market objects |
| `enableOrderBook` | boolean | Whether order book is enabled |
| `volume` | string | Total volume |
| `volume24hr` | string | 24-hour volume |
| `liquidity` | string | Total liquidity |
| `startDate` | string | Event start date |
| `endDate` | string | Event end date |
| `tags` | array | Tag objects |

### Market Object (key fields)

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Market ID |
| `slug` | string | URL slug |
| `question` | string | Market question |
| `conditionId` | string | Condition ID (used for CLOB/Data API) |
| `questionId` | string | Question ID |
| `tokenIds` | array | Array of 2 token IDs: `[yesTokenId, noTokenId]` |
| `clobTokenIds` | array | CLOB token IDs (same as tokenIds) |
| `outcomes` | array | Outcome labels: `["Yes", "No"]` |
| `outcomePrices` | array | Implied probabilities (1:1 with outcomes) |
| `active` | boolean | Is currently tradeable |
| `closed` | boolean | Is closed |
| `volume` | string | Total volume |
| `volume24hr` | string | 24-hour volume |
| `liquidity` | string | Liquidity |
| `bestBid` | string | Current best bid |
| `bestAsk` | string | Current best ask |
| `lastTradePrice` | string | Most recent trade price |
| `startDate` | string | Start date |
| `endDate` | string | End date |
| `resolutionSource` | string | Resolution source URL |
| `negRisk` | boolean | Is negative-risk market |
| `negRiskMarketID` | string | Neg-risk market ID (if applicable) |
| `enableOrderBook` | boolean | Whether CLOB is enabled |

### Key Identifiers

| Identifier | Used In | Description |
|-----------|---------|-------------|
| `conditionId` | CLOB API, Data API | Market-level identifier |
| `questionId` | On-chain | Unique question hash |
| `tokenId` (Yes) | CLOB API | ERC1155 token for Yes outcome |
| `tokenId` (No) | CLOB API | ERC1155 token for No outcome |
| `slug` | Gamma API | URL-safe unique identifier |

### Sports Metadata Object

| Field | Type | Description |
|-------|------|-------------|
| `sport` | string | Sport identifier/abbreviation |
| `image` | URI | Sport logo URL |
| `resolution` | URI | Official resolution source URL |
| `ordering` | string | Display preference (`"home"` or `"away"`) |
| `tags` | string | Comma-separated tag IDs |
| `series` | string | Series identifier (tournament/season) |

---

## 16. Data API — Endpoints

Base URL: `https://data-api.polymarket.com`
**No authentication required.**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/positions` | User positions (by address) |
| `GET` | `/closed-positions` | Closed positions |
| `GET` | `/trades` | Trade history |
| `GET` | `/activity` | User activity |
| `GET` | `/holders` | Market holder information |
| `GET` | `/open-interest` | Open interest data |
| `GET` | `/leaderboard` | Trader leaderboard rankings |
| `GET` | `/builder-analytics` | Builder program analytics |
| `GET` | `/ok` | Health check |

---

## 17. Bridge API

Base URL: `https://bridge.polymarket.com`
Note: Operates as a proxy for fun.xyz service. Not directly managed by Polymarket.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/deposit-address` | Get deposit address for a chain |
| `GET` | `/withdraw-address` | Get withdrawal address |
| `GET` | `/quote` | Get quote for bridge operation |
| `GET` | `/supported-assets` | Chains and tokens supported for deposits |
| `GET` | `/tx-status` | Transaction status |
| `POST` | `/withdraw` | Bridge USDC.e to any supported chain/token (added 2025) |

---

## 18. WebSocket — Market Channel

**URL:** `wss://ws-subscriptions-clob.polymarket.com/ws/market`
**Authentication:** Not required.

### Connection & Subscription

After connecting, immediately send a subscription message:

```json
{
  "auth": null,
  "markets": ["<condition_id_1>", "<condition_id_2>"],
  "assets_ids": ["<token_id_1>", "<token_id_2>"],
  "type": "market",
  "custom_feature_enabled": true
}
```

- `custom_feature_enabled: true` activates advanced event types.
- No limit on number of token IDs (100-token limit removed in 2025).

### Heartbeat

Send `PING` every **10 seconds**; server responds with `PONG`.

```
Client → "PING"
Server → "PONG"
```

### Message Types (Market Channel)

| Type | Description |
|------|-------------|
| `book` | Full order book snapshot |
| `price_change` | Order book price level update (delta) |
| `tick_size_change` | Tick size change notification |
| `last_trade_price` | Last executed trade price |
| `best_bid` | Best bid update |
| `best_ask` | Best ask update |
| `market_open` | New market notification |
| `market_resolved` | Market resolution notification |

### Dynamic Subscribe/Unsubscribe (without reconnecting)

```json
{
  "assets_ids": ["<token_id>"],
  "type": "market",
  "op": "subscribe"
}
```
```json
{
  "assets_ids": ["<token_id>"],
  "type": "market",
  "op": "unsubscribe"
}
```

---

## 19. WebSocket — User Channel

**URL:** `wss://ws-subscriptions-clob.polymarket.com/ws/user`
**Authentication:** Required (API credentials).

### Subscription Message

Note: uses **condition IDs** (not token/asset IDs).

```json
{
  "auth": {
    "apiKey":     "<api_key>",
    "secret":     "<secret>",
    "passphrase": "<passphrase>"
  },
  "markets": ["<condition_id>"],
  "type": "user"
}
```

### Heartbeat

Same as market channel — send `PING` every 10 seconds.

### Message Types (User Channel)

| Type | Description |
|------|-------------|
| `trade` | Trade lifecycle updates (MATCHED → CONFIRMED) |
| `order` | Order events (placement, update, cancellation) |

---

## 20. WebSocket — Sports Channel

**URL:** `wss://sports-api.polymarket.com/ws`
**Authentication:** Not required.

### Message Types

| Type | Description |
|------|-------------|
| Live scores | Game scores, periods, status |

### Heartbeat (reversed from Market/User channel)

Server sends `ping` every **5 seconds**. Client must respond with `pong` within **10 seconds** or connection is dropped.

```
Server → "ping"
Client → "pong"
```

---

## 21. WebSocket — RTDS (Real-Time Data Streaming)

**URL:** `wss://ws-live-data.polymarket.com`
**Authentication:** Optional (for `clob_user` topic).

### Subscription Format

```json
{
  "subscriptions": [
    {
      "topic":   "<topic>",
      "type":    "<message_type_or_wildcard>",
      "filters": "<json_filter_string>"
    }
  ]
}
```

### Topics and Message Types

| Topic | Type | Description | Filters |
|-------|------|-------------|---------|
| `activity` | `trades` | Trade executions | `{"event_slug":".."}` or `{"market_slug":".."}` |
| `activity` | `orders_matched` | Matched orders | Same as trades |
| `comments` | `comment_created` | New comment | `{"parentEntityID": 100, "parentEntityType": "Event"}` |
| `comments` | `comment_removed` | Deleted comment | Same |
| `comments` | `reaction_created` | New reaction | Same |
| `comments` | `reaction_removed` | Removed reaction | Same |
| `rfq` | `request_created` | RFQ created | — |
| `rfq` | `request_edited` | RFQ edited | — |
| `rfq` | `request_canceled` | RFQ canceled | — |
| `rfq` | `request_expired` | RFQ expired | — |
| `rfq` | `quote_created` | Quote created | — |
| `rfq` | `quote_edited` | Quote edited | — |
| `rfq` | `quote_canceled` | Quote canceled | — |
| `rfq` | `quote_expired` | Quote expired | — |
| `crypto_prices` | `update` | Crypto price tick | `{"symbol": "BTCUSDT"}` |
| `equity_prices` | `update` | Stock price tick | `{"symbol": "AAPL"}` |
| `clob_user` | `*` | All user CLOB events | Requires `clob_auth` |

### Wildcard: use `"*"` as type to receive all message types for a topic.

### Authenticated RTDS Subscription (clob_user)

```json
{
  "subscriptions": [{
    "topic": "clob_user",
    "type": "*",
    "clob_auth": {
      "key":        "<api_key_uuid>",
      "secret":     "<base64_secret>",
      "passphrase": "<base64_passphrase>"
    }
  }]
}
```

### Supported Crypto Symbols
`BTCUSDT`, `ETHUSDT`, `XRPUSDT`, `SOLUSDT`, `DOGEUSDT`

### Supported Equity Symbols
`AAPL`, `TSLA`, `MSFT`, `GOOGL`, `AMZN`, `META`, `NVDA`, `NFLX`, `PLTR`, `OPEN`, `RKLB`, `ABNB`

### Key Message Fields

**Trade:** `asset`, `outcome`, `outcomeIndex`, `price`, `side` (BUY/SELL), `size`, `timestamp`

**Comment:** `body`, `parentEntityType` (Event/Series), `userAddress`, `createdAt`, `updatedAt`

**RFQ Request/Quote:** `requestId`/`quoteId`, `side`, `sizeIn`, `sizeOut`, `price`, `expiry` (Unix), `state`

**Price Update:** `symbol`, `value`, `timestamp` (milliseconds)

### Important Constraints
- `crypto_prices` topic: only **one symbol per connection**. New symbol subscription replaces previous. Use separate connections for multiple symbols.
- Server sends initial historical snapshot on connection for price topics.
- Max **500 instruments per WebSocket connection**.

---

## 22. Smart Contract Addresses

All on Polygon (chain ID 137).

| Contract | Address |
|---------|---------|
| USDC.e (Collateral) | `0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174` |
| Conditional Tokens (CTF) | `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045` |
| Exchange (Main) | `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E` |
| Neg Risk Exchange | `0xC5d563A36AE78145C45a50134d48A1215220f80a` |
| Neg Risk Adapter | `0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296` |

### Required Token Allowances (EOA/MetaMask wallets only)

For **USDC.e** — approve all three exchange contracts:
- `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E`
- `0xC5d563A36AE78145C45a50134d48A1215220f80a`
- `0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296`

For **Conditional Tokens** — approve same three contracts.

Email/Magic wallets: allowances set automatically.

---

## 23. Error Codes

### HTTP Status Codes

| Code | Meaning | Action |
|------|---------|--------|
| 400 | Bad Request — invalid parameters | Fix request fields |
| 401 | Unauthorized — invalid credentials or expired signature | Re-authenticate |
| 403 | Forbidden — insufficient permissions or geo-restricted | Check permissions |
| 404 | Not Found — market or order doesn't exist | Verify IDs |
| 429 | Rate Limit Exceeded | Exponential backoff |
| 500 | Server Error | Retry with backoff |

### Order-Level Error Messages

| Error Code | Meaning |
|-----------|---------|
| `INVALID_ORDER_MIN_TICK_SIZE` | Price precision violates market tick size |
| `INVALID_ORDER_MIN_SIZE` | Order size below minimum |
| `INVALID_ORDER_DUPLICATED` | Identical order already exists |
| `INVALID_ORDER_NOT_ENOUGH_BALANCE` | Insufficient balance or allowance |
| `INVALID_ORDER_EXPIRATION` | Expiration timestamp is in the past |
| `INVALID_POST_ONLY_ORDER_TYPE` | Post-only combined with FOK/FAK |
| `INVALID_POST_ONLY_ORDER` | Post-only order would cross spread |
| `FOK_ORDER_NOT_FILLED_ERROR` | FOK could not be fully filled |
| `MARKET_NOT_READY` | Market is not accepting orders |

---

## 24. Negative Risk Markets

Markets with 3 or more outcomes use the **Negative Risk** model for capital efficiency.

**Detection:** `negRisk: true` field on the market object.

**Requirement:** Set `negRisk: true` in order options when placing orders on these markets.

```typescript
const order = await client.createAndPostOrder(
  { tokenID, price: 0.30, size: 10, side: Side.BUY },
  { tickSize: "0.01", negRisk: true }  // <-- required for neg-risk markets
);
```

```python
order = client.create_and_post_order(
    OrderArgs(token_id=token_id, price=0.30, size=10, side=BUY),
    options={"tick_size": "0.01", "neg_risk": True}
)
```

The Neg Risk Adapter contract (`0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296`) handles these markets on-chain.

---

## 25. SDK Quick Reference

### TypeScript / Node.js

**Install:**
```bash
npm install @polymarket/clob-client ethers@5
```

**Basic setup:**
```typescript
import { ClobClient, Side, OrderType } from "@polymarket/clob-client";
import { Wallet } from "ethers";

const signer = new Wallet(process.env.PRIVATE_KEY!);
const client = new ClobClient("https://clob.polymarket.com", 137, signer);

// Create or derive API credentials
const creds = await client.createOrDeriveApiKey();
client.setCreds(creds);

// Place limit order
const order = await client.createAndPostOrder(
  { tokenID: "<token_id>", price: 0.50, size: 10, side: Side.BUY },
  { tickSize: "0.01", negRisk: false },
  OrderType.GTC
);
```

### Python

**Install:**
```bash
pip install py-clob-client
```

**Basic setup:**
```python
from py_clob_client.client import ClobClient
from py_clob_client.clob_types import OrderArgs, OrderType
from py_clob_client.order_builder.constants import BUY

# Initialize with private key
client = ClobClient(
    host="https://clob.polymarket.com",
    key="<private_key>",
    chain_id=137,
    signature_type=0,    # 0=EOA, 1=email/magic, 2=gnosis safe
    funder="<wallet_address>"  # required for types 1 and 2
)

# Set API credentials
client.set_api_creds(client.create_or_derive_api_creds())

# Place limit order
resp = client.create_and_post_order(
    OrderArgs(token_id="<token_id>", price=0.50, size=10, side=BUY),
    options={"tick_size": "0.01", "neg_risk": False},
    order_type=OrderType.GTC
)
```

**Read-only (no auth):**
```python
client = ClobClient("https://clob.polymarket.com")
book = client.get_order_book("<token_id>")
price = client.get_price("<token_id>", side="BUY")
mid = client.get_midpoint("<token_id>")
```

### Additional SDKs

| Language | Package | Source |
|----------|---------|--------|
| TypeScript | `@polymarket/clob-client` | npm |
| Python | `py-clob-client` | PyPI |
| Rust | `polymarket-client-sdk` | crates.io |
| Builder Tools (TS) | `@polymarket/builder-signing-sdk` | npm |
| Builder Tools (Py) | `py_builder_signing_sdk` | PyPI |
| Gasless Relay (TS) | `@polymarket/builder-relayer-client` | npm |
| Gasless Relay (Py) | `py-builder-relayer-client` | PyPI |

---

## Appendix: Key SDK Data Types (Python)

```python
from py_clob_client.clob_types import (
    BookParams,        # { token_id: str }
    OrderArgs,         # { token_id, price, size, side, [expiration, feeRateBps, nonce, taker] }
    MarketOrderArgs,   # { token_id, amount, side, [price, order_type] }
    OpenOrderParams,   # { id?, market?, asset_id? }
    OrderType,         # GTC, GTD, FOK, FAK
)
from py_clob_client.order_builder.constants import BUY, SELL
```

## Appendix: Finding Token IDs

```python
# From Gamma API
import requests

resp = requests.get(
    "https://gamma-api.polymarket.com/markets",
    params={"active": "true", "closed": "false", "limit": 1}
)
market = resp.json()[0]
token_ids = market["clobTokenIds"]     # [yes_token_id, no_token_id]
condition_id = market["conditionId"]   # for user channel WebSocket
tick_size = market.get("minTickSize", "0.01")
neg_risk = market.get("negRisk", False)
```

---

## Appendix: Fees

- Most markets: **zero fees**.
- 15-minute crypto markets: **0.2%–1.6%** on buys, **0.8%–3.7%** on sells.
- Sports taker fees enabled on select leagues (NCAAB, Serie A as of Feb 2026).

---

*Sources:*
- [docs.polymarket.com](https://docs.polymarket.com)
- [CLOB Introduction](https://docs.polymarket.com/developers/CLOB/introduction)
- [Authentication](https://docs.polymarket.com/developers/CLOB/authentication)
- [Orders Overview](https://docs.polymarket.com/developers/CLOB/orders/orders)
- [Place Single Order](https://docs.polymarket.com/developers/CLOB/orders/create-order)
- [WSS Overview](https://docs.polymarket.com/developers/CLOB/websocket/wss-overview)
- [API Rate Limits](https://docs.polymarket.com/quickstart/introduction/rate-limits)
- [Gamma API Overview](https://docs.polymarket.com/developers/gamma-markets-api/overview)
- [Endpoints Reference](https://docs.polymarket.com/quickstart/reference/endpoints)
- [L1 Methods](https://docs.polymarket.com/developers/CLOB/clients/methods-l1)
- [L2 Methods](https://docs.polymarket.com/developers/CLOB/clients/methods-l2)
- [py-clob-client GitHub](https://github.com/Polymarket/py-clob-client)
- [real-time-data-client GitHub](https://github.com/Polymarket/real-time-data-client)
- [NautilusTrader Polymarket Integration](https://nautilustrader.io/docs/latest/integrations/polymarket/)
