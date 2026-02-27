# Copytrading Wallet Management — Design

**Date:** 2026-02-26
**Status:** Approved

## Goal

Allow the user to add, remove, toggle, and edit copytrading wallets on-the-fly without manually editing `config.toml`. Available in both the TUI (Tab 4: Copytrading) and Telegram Bot.

## Background

`CopyTrader` already hot-reloads via fsnotify when `config.toml` changes. Any mutation of `cfg.Copytrading.Traders` followed by `config.Save()` automatically starts/stops the relevant `TraderTracker` goroutines — no extra plumbing needed.

## TUI Design

### Approach: Modal inline form in the Copytrading tab

`CopytradingModel` gains a `mode` state enum:

```go
type copyMode int
const (
    modeTable copyMode = iota
    modeAddForm
    modeEditForm
    modeConfirmDelete
)
```

The table is always shown. When a form mode is active, a form panel renders below the table.

### Updated constructor

```go
func NewCopytradingModel(cfg *config.Config, cfgPath string, width, height int) CopytradingModel
```

`cfg` and `cfgPath` are stored on the model for direct mutation and save.

### Key bindings (table mode)

| Key    | Action                                      |
|--------|---------------------------------------------|
| `↑`/`↓` | Navigate rows                              |
| `a`    | Open add form (blank fields)               |
| `e`    | Open edit form (pre-filled with selected)  |
| `d`    | Enter confirm-delete mode                  |
| `space`| Toggle `enabled` on selected row, save     |

### Key bindings (form mode)

| Key              | Action                        |
|------------------|-------------------------------|
| `Tab` / `↓`      | Next field                    |
| `Shift+Tab` / `↑`| Previous field                |
| `Enter` (last)   | Save, return to table         |
| `Esc`            | Cancel, return to table       |

### Form fields

- **Address** (string, required) — proxy-wallet address
- **Label** (string, optional)
- **Alloc%** (float, default 5.0) — allocation percentage
- **MaxPositionUSD** (float, default 50.0) — max position size

### Delete confirmation

Shows inline prompt: `Delete 0xabc...? [y/N]`
`y` → remove from `cfg.Copytrading.Traders`, save.
Any other key → cancel.

### Data flow

1. User presses `a` → `mode = modeAddForm`, blank textinputs shown.
2. User fills fields, presses `Enter` on last field.
3. `saveForm()` appends/updates `cfg.Copytrading.Traders`.
4. `config.Save(cfgPath, cfg)` writes to disk.
5. `CopyTrader`'s fsnotify detects change → `applyConfig()` starts/stops trackers.
6. Existing `ConfigReloadedMsg` flow updates `AppModel.cfg`.

### Files changed (TUI)

- `internal/tui/tab_copytrading.go` — mode state machine, textinput form, key handling
- `internal/tui/app.go` — pass `cfg`/`cfgPath` to `NewCopytradingModel`; update `cfg` on `ConfigReloadedMsg`
- `internal/tui/keys.go` — add `CopyAdd`, `CopyEdit`, `CopyDelete`, `CopyToggle` key bindings

## Telegram Bot Design

### Commands

Consistent with existing `/set <key> <value>` pattern — no new conversation state machine required.

| Command                                    | Action                                   |
|--------------------------------------------|------------------------------------------|
| `/addtrader <address> <label> <alloc_pct>` | Add a new trader to config               |
| `/removetrader <address>`                   | Remove trader by address                 |
| `/toggletrader <address>`                   | Flip `enabled` flag on trader            |

### Inline keyboard on Copytrading screen

```
📋 Tracked Traders

0xabc… alice | ● active | 5%
[⏸ Disable] [🗑 Remove]

0xdef… bob | ● active | 10%
[⏸ Disable] [🗑 Remove]

[➕ Add (use /addtrader)]  [← Back]
```

Callback data patterns:
- `trader:remove:<address>`
- `trader:toggle:<address>`

### Mutation pattern

Same as `doSetSetting`: lock `cfgMu` → mutate `cfg.Copytrading.Traders` → `config.Save()` → emit `ConfigReloadedMsg` via bus.

### Files changed (Telegram)

- `internal/telegrambot/handlers.go` — add commands, update `copytradingKeyboard()`, handle new callbacks, add `doAddTrader` / `doRemoveTrader` / `doToggleTrader` helpers

## Non-goals

- Editing `poll_interval_ms` or `size_mode` per-trader via the form (can use `/set` for those)
- Multi-step wizard in Telegram (commands with args are sufficient)
- Any changes to `internal/copytrading/` — the existing hot-reload handles everything
