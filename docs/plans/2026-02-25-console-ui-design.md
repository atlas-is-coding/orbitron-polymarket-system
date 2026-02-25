# Console UI Design

**Date:** 2026-02-25
**Status:** Approved

## Overview

Add a Bubble Tea + Lipgloss TUI dashboard to polytrade-bot. The UI wraps all existing subsystems without modifying their core logic. A first-run wizard handles initial config generation. A Settings tab provides full in-app configuration with tooltips and hot reload via fsnotify.

## Architecture

**Approach:** Sub-models. A root `AppModel` delegates to per-tab sub-models. Each sub-model handles its own `Update`/`View`. The root model routes `tea.Msg` to the active tab and broadcasts shared events (config reload, bot events) to all tabs.

### File Structure

```
internal/tui/
  app.go              — AppModel (root model, tab routing)
  tabs.go             — TabBar component
  styles.go           — all Lipgloss styles (colors, borders, theme)
  keys.go             — global keybindings
  messages.go         — custom tea.Msg types (ConfigReloaded, BotEvent, ...)
  wizard/
    wizard.go         — first-run setup wizard model
  tabs/
    overview.go       — subsystem status + quick stats
    orders.go         — open orders table
    positions.go      — positions table with P&L
    copytrading.go    — tracked traders + copied trades
    logs.go           — scrollable log buffer
    settings.go       — full config editor with tooltips

internal/config/
  watcher.go          — fsnotify-based ConfigWatcher (new)
```

## Layout

### Main Dashboard
```
┌─────────────────────────────────────────────────────────┐
│  polytrade-bot  ●Running   Wallet: 0x1234...  USDC: 420 │  ← Header
├─────────────────────────────────────────────────────────┤
│ [Overview] [Orders] [Positions] [Copytrading] [Logs] [Settings] │  ← TabBar
├─────────────────────────────────────────────────────────┤
│                                                         │
│                    <Tab Content>                        │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  Tab/Shift+Tab: switch  ↑↓: navigate  q: quit           │  ← Help bar
└─────────────────────────────────────────────────────────┘
```

### First-Run Wizard (no config.toml)
```
┌── First Run Setup ─────────────────────────────────────┐
│  Step 1/4: Private Key                                  │
│  ┌──────────────────────────────────────────────────┐  │
│  │ ••••••••••••••••••••••••••••••••••••••••••••     │  │
│  └──────────────────────────────────────────────────┘  │
│  Hex Ethereum wallet key (no 0x prefix).               │
│  Used for EIP-712 order signing.                        │
│                                              [Enter →]  │
└─────────────────────────────────────────────────────────┘
```
Steps: `private_key` → `api_key` → `api_secret` → `passphrase` → generate `config.toml` → launch.

## Data Flow

Subsystems run in goroutines as before. A shared `EventBus` (`chan tea.Msg`) bridges them to the TUI. Each subsystem writes events; TUI polls via `tea.Cmd` (non-blocking).

```
monitor.Run() ──→ eventBus ──→ WaitForEvent (tea.Cmd) ──→ AppModel.Update()
trades.Run()  ──→ eventBus
ws.Run()      ──→ eventBus
```

No mutexes in TUI — all state changes via immutable Bubble Tea messages.

## Hot Reload

```
fsnotify.Watcher
      │ (file changed)
      ▼
ConfigWatcher.Run()  →  debounce 300ms  →  chan ConfigReloadedMsg
      │
      ▼
WaitForConfigReload() tea.Cmd
      │
      ▼
AppModel.Update(ConfigReloadedMsg)
  ├── updates in-memory cfg
  ├── notifies each subsystem via its OnReload callback
  └── Settings tab re-renders with new values
```

`internal/config/watcher.go` — new file. Uses fsnotify (already in go.mod). Exports `ConfigWatcher` with `Run(ctx)` and `OnReload func(*Config)` callback. Subsystems receive updated config via the same callback pattern already used in copytrading (`func() *CopytradingConfig`).

## Tab Designs

### Overview
- Left panel: subsystem status (WebSocket, Monitor, Trades Monitor, Trading Engine, Copytrading) with ●/○ indicators
- Right panel: quick stats (USDC balance, open orders count, positions count, today's P&L, tracked traders count)

### Orders
Table columns: Market | Side | Price | Size | Filled | Status | Age
Keybindings: `D` cancel selected order, `A` cancel all orders.

### Positions
Table columns: Market | Side | Size | Entry Price | Current Price | P&L | P&L%
Default sort: P&L descending.

### Copytrading
Two sections: tracked traders list (address, label, status, allocation%) + recent copied trades feed.

### Logs
- Scrollable buffer, last 500 lines
- Level filter: `T/D/I/W/E` keys
- `F` to freeze/unfreeze autoscroll
- Colors: WARN=yellow, ERROR=red, DEBUG=dim

### Settings
```
┌── Settings ────────────────────────────────────────────────┐
│  [Auth] [API] [Monitor] [Trading] [Copytrading] [Telegram] │
├────────────────────────────────┬───────────────────────────┤
│  Auth                          │  Tooltip                  │
│  ─────────────────────────     │  ─────────────────────    │
│  Private Key     [••••••••]    │  Hex Ethereum wallet key  │
│  API Key         [abc123  ] ◄  │  without 0x prefix.       │
│  API Secret      [••••••••]    │  Used for EIP-712 order   │
│  Passphrase      [••••••••]    │  signing and address      │
│  Chain ID        [137     ]    │  derivation.              │
│                                │                           │
│                                │  Current: abc123...       │
│                                │  Modified: yes ●          │
├────────────────────────────────┴───────────────────────────┤
│  [S] Save    [R] Reset    [Esc] Cancel                     │
└────────────────────────────────────────────────────────────┘
```

- `↑↓` navigate fields, right panel shows tooltip for focused field
- `Enter` enter edit mode for field
- `Esc` exit edit without saving
- `S` save to `config.toml` (triggers hot reload)
- `R` reset all unsaved changes
- Modified fields marked with `●`
- Password/key fields masked with `••••`, revealed during editing

**Settings sections:** Auth / API / Monitor / Monitor.Trades / Trading / Copytrading / Telegram / Database / Log

## Entry Point

`cmd/bot/main.go` checks for `config.toml` on startup:
- Missing → launch wizard, generate config, then start TUI
- Present → load config, start all subsystems, launch TUI

TUI mode is the default. A `--no-tui` flag preserves the current plain log mode for headless/CI use.
