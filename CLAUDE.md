# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build binary
go build ./...

# Build and run
go run ./cmd/bot/ --config config.toml

# Vet
go vet ./...

# Test
go test ./...

# Tidy dependencies
go mod tidy
```

## Architecture

The bot is structured around four independent, context-cancellable subsystems launched in `cmd/bot/main.go`:

1. **WebSocket client** (`internal/api/ws/`) — persistent connection to `wss://ws-subscriptions-clob.polymarket.com/ws/` with auto-reconnect. Supports three channels: `market` (order book), `user` (private events), `asset` (prices).

2. **Monitor** (`internal/monitor/`) — polls Gamma API on a configurable interval, evaluates `Rule` conditions against market state diff, and dispatches `Alert`s via `Notifier`.

   **TradesMonitor** (`internal/monitor/trades.go`) — separate monitor that polls CLOB + Data API. Tracks open orders, trade fills, and positions; generates alerts on fill/cancel. Exposes `CancelOrder`, `CancelAllOrders`, `GetDataPositions`, etc. Requires L2 auth (`monitor.trades.enabled = true` in config).

3. **Trading Engine** (`internal/trading/`) — manages `Strategy` implementations. Each `Strategy` runs in its own goroutine. Register strategies before calling `engine.Start(ctx)`.

4. **Notifier** (`internal/notify/`) — interface-based; `NoopNotifier` is the default, `telegram.Notifier` is activated via config.

### API Layer

Three Polymarket APIs, each with its own `*api.Client` (fasthttp wrapper with retry):
- **CLOB** (`internal/api/clob/`) — trading and order book. Requires L2 credentials for authenticated endpoints.
- **Gamma** (`internal/api/gamma/`) — market metadata and events. Public, no auth needed.
- **Data** (`internal/api/data/`) — enriched history: positions with P&L, trades by wallet. Public, no auth needed. Base URL: `https://data-api.polymarket.com`.

### Authentication

- **L1** (`internal/auth/l1.go`) — `L1Signer` wraps an Ethereum private key; used to derive the wallet address and create API keys via `POST /auth/api-key`.
- **L2** (`internal/auth/l2.go`) — `L2Credentials` signs each request with HMAC-SHA256. Headers: `POLY_ADDRESS`, `POLY_API_KEY`, `POLY_PASSPHRASE`, `POLY_TIMESTAMP`, `POLY_SIGNATURE`. Signatures expire in **30 seconds**.

### Storage

`internal/storage/storage.go` defines `TradeStore`, `OrderStore`, and `Store` interfaces. The SQLite implementation in `internal/storage/sqlite/` is a stub — activate by adding `github.com/mattn/go-sqlite3` (requires CGO/gcc) and implementing the methods.

### Extending the bot

- **New trading strategy**: implement `trading.Strategy` (3 methods: `Name`, `Start`, `Stop`), instantiate in `main.go`, call `engine.Register(s)`.
- **New notification channel**: implement `notify.Notifier` (`Send(ctx, msg) error`), wire in `main.go`.
- **New storage backend**: implement `storage.Store`, swap in `main.go`.

## Key Config Fields

`config.toml` controls all behaviour. Trading and database are disabled by default (`enabled = false`). Set `auth.private_key` (hex, no `0x` prefix) and `auth.api_key/secret/passphrase` to enable authenticated trading.

Key subsections:
- `[monitor.trades]` — enables TradesMonitor; requires L2 auth; options: `poll_interval_ms`, `track_positions`, `trades_limit`
- `[log]` — `level` (trace/debug/info/warn/error) and `format` (pretty / json)
- `chain_id` — 137 = Polygon Mainnet, 80002 = Amoy Testnet

## Polymarket API Reference

See `POLYMARKET_DOCS.md` for the full API reference (endpoints, rate limits, WebSocket channels, auth flows, contract addresses).
