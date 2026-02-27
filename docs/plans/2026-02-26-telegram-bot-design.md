# Telegram Bot Design

**Date:** 2026-02-26
**Status:** Approved

## Overview

Add an interactive Telegram Bot to polytrade-bot that mirrors the Console TUI. Users can view all data (overview, orders, positions, copytrading, logs) and manage settings through Telegram. All state is synchronized bidirectionally via the shared EventBus.

## Architecture

### New Package: `internal/telegrambot/`

```
internal/telegrambot/
├── bot.go          — Bot struct, Run(ctx), two goroutine loops
├── handlers.go     — slash command + callback query handlers
├── renderer.go     — HTML formatting for Telegram messages
└── state.go        — BotState: in-memory cache of latest data
```

### Core Types

**`Bot` struct:**
```go
type Bot struct {
    api         *tgbotapi.BotAPI
    bus         *tui.EventBus        // shared EventBus
    cfg         *config.Config       // pointer updated on ConfigReloadedMsg
    cfgPath     string
    state       *BotState
    allowedIDs  map[int64]bool       // from cfg.Telegram.ChatID (comma-separated or single)
    adminID     int64                // from cfg.Telegram.AdminChatID
    log         zerolog.Logger
}
```

**`BotState` struct (thread-safe cache):**
```go
type BotState struct {
    mu        sync.RWMutex
    balance   float64
    orders    []tui.OrderRow
    positions []tui.PositionRow
    traders   []tui.TraderRow
    logs      []string        // last 50 lines
    subsystems map[string]bool
}
```

### EventBus Extensions

New message types added to `internal/tui/messages.go`:
- `OrdersUpdateMsg{Rows []OrderRow}` — TradesMonitor updated orders
- `PositionsUpdateMsg{Rows []PositionRow}` — positions refreshed

### Run() Goroutine Architecture

`Bot.Run(ctx)` spawns two concurrent loops:

1. **Telegram polling loop** — `bot.GetUpdatesChan()`, dispatches to handlers
2. **EventBus consumer loop** — reads EventBus messages, updates BotState cache, sends Telegram notifications on important events (fill, cancel, alert)

### Integration in main.go

```go
if cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" {
    tgBot, err := telegrambot.New(cfg, cfgPath, bus, log)
    if err == nil {
        startSubsystem("Telegram Bot", func() error { return tgBot.Run(ctx) })
    }
}
```

## Config Changes

`TelegramConfig` in `internal/config/config.go`:
```go
type TelegramConfig struct {
    Enabled     bool   `toml:"enabled"`
    BotToken    string `toml:"bot_token"`
    ChatID      string `toml:"chat_id"`       // allowed user (whitelist)
    AdminChatID string `toml:"admin_chat_id"` // NEW: can edit secret fields
}
```

`config.toml`:
```toml
[telegram]
enabled = true
bot_token = "..."
chat_id = "123456"
admin_chat_id = "789012"  # optional; if empty, nobody can edit secret fields
```

## Commands

| Command | Access | Description |
|---------|--------|-------------|
| `/start` | user | Main menu with inline keyboard navigation |
| `/status` | user | Overview: balance, subsystem status, open orders, positions count |
| `/orders` | user | Active orders table + [Cancel] button per order |
| `/cancel <id>` | user | Cancel specific order by ID |
| `/cancelall` | user | Cancel all orders (requires confirmation via inline button) |
| `/positions` | user | Positions table with P&L |
| `/copy` | user | Copytrading status + tracked traders list |
| `/logs` | user | Last 20 log lines |
| `/settings` | user | View all settings (secrets masked as ••••) |
| `/set <key> <value>` | user/admin | Change a setting (user: safe fields only; admin: all fields) |

### Field Access Control

**Safe fields (all allowed users can /set):**
- `monitor.enabled`, `monitor.poll_interval_ms`
- `monitor.trades.enabled`, `monitor.trades.poll_interval_ms`, `monitor.trades.alert_on_fill`, `monitor.trades.alert_on_cancel`
- `trading.enabled`, `trading.max_position_usd`, `trading.slippage_pct`, `trading.neg_risk`
- `copytrading.enabled`, `copytrading.poll_interval_ms`, `copytrading.size_mode`
- `telegram.enabled`
- `database.enabled`, `database.path`
- `log.level`, `log.format`
- `ui.language`

**Restricted fields (admin only):**
- `auth.private_key`, `auth.api_key`, `auth.api_secret`, `auth.passphrase`
- `auth.chain_id`
- `telegram.bot_token`, `telegram.chat_id`, `telegram.admin_chat_id`

### `/set` dot-notation key mapping

Key → `FieldDef` mapping table in `handlers.go`, reusing the same `FieldDef.Set()` functions from `tab_settings.go`. After applying, calls `saveConfig()` and sends `ConfigReloadedMsg` to EventBus — TUI updates automatically.

## Inline Keyboard Navigation

```
/start → Main menu:
[📊 Orders] [💼 Positions]
[ℹ️ Overview] [🔄 Copytrading]
[📝 Logs] [⚙️ Settings]

Orders view:
1. BTC-USD BUY 0.5 @ 45000  [Cancel]
2. ETH-USD SELL 1.0 @ 3200  [Cancel]
[Cancel All] [🔄 Refresh] [← Back]

Settings view (sections):
[UI] [Auth] [Monitor] [Trades] [Trading] [Copy] [Telegram] [DB] [Log]
```

## Bidirectional Synchronization

| Event | Direction | Mechanism |
|-------|-----------|-----------|
| New orders from TradesMonitor | bot→TUI | `OrdersUpdateMsg` on EventBus |
| New positions from TradesMonitor | bot→TUI | `PositionsUpdateMsg` on EventBus |
| Config changed in TUI (S key) | TUI→bot | `ConfigReloadedMsg` on EventBus |
| Config changed via `/set` | bot→TUI | `ConfigReloadedMsg` on EventBus |
| Balance update | monitor→both | `BalanceMsg` on EventBus |
| Subsystem start/stop | subsystems→both | `SubsystemStatusMsg` on EventBus |
| Order fill/cancel alert | tradesMon→both | `BotEventMsg` + Telegram push |

## Security

- Every incoming Telegram update is checked against `allowedIDs` before processing
- If `chat_id` not in whitelist → silently ignored (no response)
- Admin-only commands checked against `adminID`
- Secret fields shown as `••••` for non-admin `/settings`
- `/cancelall` requires inline button confirmation to prevent accidental execution

## Library

`github.com/go-telegram-bot-api/telegram-bot-api/v5` — polling mode (no webhook server needed).

## Files Changed/Created

**New:**
- `internal/telegrambot/bot.go`
- `internal/telegrambot/handlers.go`
- `internal/telegrambot/renderer.go`
- `internal/telegrambot/state.go`

**Modified:**
- `internal/config/config.go` — add `AdminChatID` to `TelegramConfig`
- `internal/tui/messages.go` — add `OrdersUpdateMsg`, `PositionsUpdateMsg`
- `internal/monitor/trades.go` — send `OrdersUpdateMsg`/`PositionsUpdateMsg` to EventBus
- `cmd/bot/main.go` — instantiate and start Telegram Bot subsystem
- `go.mod` / `go.sum` — add go-telegram-bot-api/v5
