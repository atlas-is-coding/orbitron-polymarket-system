# Copytrading Design

**Date**: 2026-02-24
**Status**: Approved

## Overview

Add copytrading capability to polytrade-bot: monitor selected Polymarket traders via the public Data API, automatically replicate their positions, and close them when the trader exits.

## Requirements

- Select traders from Polymarket leaderboard / Predictfolio (by proxy-wallet address)
- Detect new and closed positions by polling `data-api.polymarket.com/positions?user=<address>`
- Scale trade size proportionally (relative to trader's balance) OR as a fixed % of our balance
- Auto-close our position when trader closes theirs
- Persist full copy-trade history in SQLite (deduplication, crash recovery)
- Hot-reload trader list from config.toml without restart
- Telegram alerts on open/close/failure

## Architecture

### New package: `internal/copytrading/`

```
internal/copytrading/
├── copier.go    — CopyTrader orchestrator (Run loop, hot-reload)
├── tracker.go   — TraderTracker (one goroutine per trader, diff positions)
├── sizer.go     — SizeCalculator (proportional and fixed_pct modes)
├── executor.go  — OrderExecutor (create/close orders via CLOB)
└── models.go    — CopyTrade, TraderState, CopyTradeRecord
```

**CopyTrader** — top-level component started in `main.go`:
- Creates one `TraderTracker` goroutine per enabled trader from config
- Watches config.toml via `fsnotify`: adds/stops trackers on change
- Passes shared `OrderExecutor` and `SizeCalculator` to each tracker

**TraderTracker** — per-trader polling loop:
- Polls `data.Client.GetPositions(trader_address)` every `poll_interval_ms`
- Diffs current vs previous snapshot (`map[assetID]Position`)
- New asset → call executor to open; missing asset → call executor to close
- On startup: reconciles open DB records with live positions (handles downtime)

**SizeCalculator**:
- `proportional`: `mySize = (traderPositionValue / traderTotalBalance) * myBalance * allocation_pct / 100`
- `fixed_pct`: `mySize = myBalance * allocation_pct / 100`
- Per-trader `size_mode` overrides global
- Always clamps to `max_position_usd`

**OrderExecutor**:
- Open: `clob.CreateOrder` (market BUY at best ask, type GTC)
- Close: sell all held tokens (market SELL at best bid)
- Retry up to 3 times on failure, then mark `CopyTradeRecord.status = "failed"` and notify

## Configuration

New section in `config.toml`:

```toml
[copytrading]
  enabled          = false
  poll_interval_ms = 10000
  # Global size mode: "proportional" or "fixed_pct"
  size_mode        = "proportional"

  [[copytrading.traders]]
    address          = "0xABC..."
    label            = "whale1"
    enabled          = true
    allocation_pct   = 10.0
    max_position_usd = 50.0
    # size_mode = "fixed_pct"  # overrides global if set
```

Hot-reload: `fsnotify` watches config file. On change, reloads trader list:
- New trader → start TraderTracker goroutine
- Removed trader → stop goroutine (existing open positions not auto-closed, alert sent)
- Changed params → stop + restart tracker

## Data Flow

### Opening a position

1. `TraderTracker` detects new `asset_id` in trader's positions snapshot
2. Check SQLite: skip if open `CopyTradeRecord` already exists for this `asset_id`
3. `SizeCalculator` computes `mySize` (clamped to `max_position_usd`)
4. `OrderExecutor` places market-buy via `clob.CreateOrder`
5. Save `CopyTradeRecord{status: "open", our_order_id: ...}` to SQLite
6. Send Telegram alert: "📈 Opened copy trade: [label] [market] size=$X"

### Closing a position

1. `TraderTracker` detects `asset_id` disappeared from trader's snapshot
2. Load open `CopyTradeRecord` from SQLite for this `asset_id`
3. `OrderExecutor` places market-sell for full position size
4. Update SQLite: `status = "closed"`, `closed_at`, `pnl`
5. Send Telegram alert: "📉 Closed copy trade: [label] [market] pnl=$X"

### Edge cases

| Scenario | Handling |
|---|---|
| Bot was offline, trader closed position | On startup, reconcile open DB records with live positions; close orphaned |
| Order not filled (retry exhausted) | Mark `status = "failed"`, Telegram alert |
| Trader removed from config | Stop tracker, alert user; positions remain open |
| Duplicate detection | SQLite unique constraint on `(trader_address, asset_id, status=open)` |
| My balance insufficient | Log warning, skip trade, alert |

## Storage Schema

New tables in `internal/storage/sqlite/`:

```sql
CREATE TABLE IF NOT EXISTS copy_traders (
    address    TEXT PRIMARY KEY,
    label      TEXT NOT NULL,
    enabled    INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS copy_trades (
    id             TEXT PRIMARY KEY,
    trader_address TEXT NOT NULL,
    asset_id       TEXT NOT NULL,
    condition_id   TEXT NOT NULL,
    side           TEXT NOT NULL,
    size           REAL NOT NULL,
    price          REAL NOT NULL,
    our_order_id   TEXT,
    status         TEXT NOT NULL DEFAULT 'open',
    opened_at      TEXT NOT NULL,
    closed_at      TEXT,
    pnl            REAL
);

CREATE INDEX IF NOT EXISTS idx_copy_trades_open
    ON copy_trades(trader_address, asset_id, status);
```

New interface methods on `storage.Store`:

```go
SaveCopyTrade(ctx context.Context, r *CopyTradeRecord) error
UpdateCopyTrade(ctx context.Context, id, status string, closedAt time.Time, pnl float64) error
GetOpenCopyTrades(ctx context.Context, traderAddress string) ([]*CopyTradeRecord, error)
```

## Integration in main.go

```go
// After existing components are initialized:
if cfg.Copytrading.Enabled {
    copier := copytrading.New(clobClient, dataClient, store, notifier, &cfg.Copytrading, log)
    go func() {
        if err := copier.Run(ctx); err != nil && ctx.Err() == nil {
            errCh <- fmt.Errorf("copytrading: %w", err)
        }
    }()
}
```

## Dependencies

- `github.com/fsnotify/fsnotify` — hot-reload config file watching
- `github.com/mattn/go-sqlite3` — SQLite driver (CGO required, gcc needed)
- `github.com/google/uuid` — generate copy trade IDs

## Out of Scope

- Polymarket leaderboard scraping (user provides wallet addresses manually)
- Predictfolio scraping (same — user adds addresses to config)
- Partial position sync (we open when trader opens, close when trader closes entirely)
- Position sizing based on live CLOB balance query (use configured `allocation_pct`)
