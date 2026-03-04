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

The bot is structured around seven independent, context-cancellable subsystems launched in `cmd/bot/main.go`:

1. **WebSocket client** (`internal/api/ws/`) — persistent connection to `wss://ws-subscriptions-clob.polymarket.com/ws/` with auto-reconnect. Supports three channels: `market` (order book), `user` (private events), `asset` (prices).

2. **Monitor** (`internal/monitor/`) — polls Gamma API on a configurable interval, evaluates `Rule` conditions against market state diff, and dispatches `Alert`s via `Notifier`.

   **TradesMonitor** (`internal/monitor/trades.go`) — separate monitor that polls CLOB + Data API. Tracks open orders, trade fills, and positions; generates alerts on fill/cancel. Exposes `CancelOrder`, `CancelAllOrders`, `GetDataPositions`, etc. Requires L2 auth (`monitor.trades.enabled = true` in config).

3. **Trading Engine** (`internal/trading/`) — manages `Strategy` implementations. Each `Strategy` runs in its own goroutine. Register strategies before calling `engine.Start(ctx)`.

4. **Notifier** (`internal/notify/`) — interface-based; `NoopNotifier` is the default, `telegram.Notifier` is activated via config.

5. **Copy Trader** (`internal/copytrading/`) — monitors configured traders via Data API, copies positions via CLOB API. Requires L2 auth + private_key + database. Hot-reloads trader list on config.toml changes via fsnotify. Enable with `copytrading.enabled = true`.
   **Trader management**: Tab 4 (Copytrading) supports inline add/edit/delete/toggle. Keys: `a` add, `e` edit selected, `d` delete (confirm with `y`), `space` toggle enabled. All ops call `config.Save()` — fsnotify picks up changes automatically.

6. **Telegram Bot** (`internal/telegrambot/`) — interactive bot mirroring the TUI. Single-admin model: `cfg.Telegram.AdminChatID` is both notification target and bot controller. `telegrambot.New()` returns `(nil, nil)` when token is empty. `OrderCanceler` interface wraps `TradesMonitor` when trades are enabled. Wire with `startSubsystem("Telegram Bot", ...)` in `main.go`.
   **Navigation (edit-in-place)**: `sendOrEdit(chatID, text, keyboard)` in `bot.go` edits the existing menu message (`BotState.menuMsgID`) or sends a new one; `/start` resets `menuMsgID` to 0 to force a fresh message.
   **Conversation state**: `BotState.SetPending(input, data)` / `ClearPending()` / `Pending()` track multi-step text input. `handlePendingInput()` is in `bot.go` and handles steps like `"addtrader_addr"` → `"addtrader_label"` → `"addtrader_alloc"` and generic `"edit:<key>"`.
   **Trading screen**: `sendTrading(chatID, "orders"|"positions")` — NOT `sendOrders`/`sendPositions` (removed).
   **Settings UX**: `sectionKeys` map (not `settingsSections`) maps section display names to dot-notation keys. `sendSettingsSection(chatID, name)` renders a section with per-field toggle/edit buttons. `doToggleSetting` flips bool fields inline. Specific callbacks like `data == "edit:ui.language"` must appear **before** generic `strings.HasPrefix(data, "edit:")` in the switch — Go evaluates top-to-bottom.
   **Language picker**: `edit:ui.language` callback is intercepted to show `sendLanguagePicker`; `setlang:<code>` saves via `doSetSetting` + refreshes Settings UI section.
   **Trader management**: `/addtrader <addr> [label] [alloc_pct]` command still works; also available via `➕ Add Trader` inline button in Copytrading screen (guided conversation). Toggle/remove via inline buttons in `copytradingKeyboard()`.

7. **Web UI** (`internal/webui/`) — HTTP server + WebSocket hub embedded in the binary. Serves a Vue 3 SPA from `//go:embed web/dist`. JWT authentication (password = `cfg.WebUI.JWTSecret`). EventBus fan-out via `bus.Tap()` pushes live state to WebSocket clients. Enable with `webui.enabled = true`; EventBus is created even in `--no-tui` mode when webui is enabled.
   **Adding new REST routes**: add handler in `internal/webui/handlers.go`, register in `server.go`'s `Run()`. Use `s.jwtMiddleware(handler)` for protected routes.
   **Adding new config keys**: extend `applyConfigKey()` in `internal/webui/config_key.go`.

### API Layer

Three Polymarket APIs, each with its own `*api.Client` (fasthttp wrapper with retry):
- **CLOB** (`internal/api/clob/`) — trading and order book. Requires L2 credentials for authenticated endpoints. `GetOrders()` returns `OrdersResponse` with `.Data []Order` field (NOT `.Orders`).
- **Gamma** (`internal/api/gamma/`) — market metadata and events. Public, no auth needed.
- **Data** (`internal/api/data/`) — enriched history: positions with P&L, trades by wallet. Public, no auth needed. Base URL: `https://data-api.polymarket.com`.

### Authentication

- **L1** (`internal/auth/l1.go`) — `L1Signer` wraps an Ethereum private key; used to derive the wallet address and create API keys via `POST /auth/api-key`.
- **L2** (`internal/auth/l2.go`) — `L2Credentials` signs each request with HMAC-SHA256. Headers: `POLY_ADDRESS`, `POLY_API_KEY`, `POLY_PASSPHRASE`, `POLY_TIMESTAMP`, `POLY_SIGNATURE`. Signatures expire in **30 seconds**.
- **L2 auto-derivation**: `pubClobClient.DeriveAPIKey(l1)` (`internal/api/clob/auth.go`) is called at startup in `main.go` if `api_key` is empty — uses `GET /auth/derive-api-key`. L2 credentials are never stored in config.
- **EIP-712** (`internal/auth/order_signer.go`) — `OrderSigner` signs CLOB orders. Domain: `"Polymarket CTF Exchange"`, v1, chainId from config, contract `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E` (standard) or `0xC5d563A36AE78145C45a50134d48A1215220f80a` (negRisk). `RandomSalt()` generates per-order entropy.

### Storage

`internal/storage/storage.go` defines `TradeStore`, `OrderStore`, `CopyTradeStore`, and `Store` interfaces. The SQLite implementation in `internal/storage/sqlite/` uses `modernc.org/sqlite` (pure-Go, no CGO). Enable with `database.enabled = true` and `database.path` in config.

### Extending the bot

- **New trading strategy**: implement `trading.Strategy` (3 methods: `Name`, `Start`, `Stop`), instantiate in `main.go`, call `engine.Register(s)`.
- **New notification channel**: implement `notify.Notifier` (`Send(ctx, msg) error`), wire in `main.go`.
- **New storage backend**: implement `storage.Store`, swap in `main.go`.
- **New Telegram Bot command**: add handler in `internal/telegrambot/handlers.go` (`handleCommand` switch). For main-menu sections add to `mainMenuKeyboard()`; for trader mutations use the `doAddTrader`/`doRemoveTrader`/`doToggleTrader` pattern (copy slice → mutate → `config.Save` → emit `ConfigReloadedMsg`). Add key to `settingsMap` only for scalar config fields.

## Key Config Fields

`config.toml` controls all behaviour. Trading and database are disabled by default (`enabled = false`). Set `auth.private_key` (hex, no `0x` prefix) — L2 credentials are **auto-derived** at startup and kept in memory only (never stored in config).

Key subsections:
- `[webui]` — `enabled`, `listen` (default `127.0.0.1:8080`), `jwt_secret` (JWT signing key and login password)
- `[ui]` — `language` (en/ru/zh/ja/ko); default "en"; hot-reloads instantly via `i18n.SetLanguage`
- `[monitor.trades]` — enables TradesMonitor; requires L2 auth; options: `poll_interval_ms`, `track_positions`, `trades_limit`
- `[copytrading]` — enables copy trading; requires L2 auth + private_key + database; options: `poll_interval_ms`, `size_mode` ("proportional"/"fixed_pct"), `[[copytrading.traders]]` list
- `[telegram]` — `enabled`, `bot_token`, `admin_chat_id` (single admin; replaces the old `chat_id` field)
- `[log]` — `level` (trace/debug/info/warn/error) and `format` (pretty / json)
- `chain_id` — 137 = Polygon Mainnet, 80002 = Amoy Testnet

## TUI Package

- Go version: 1.24.4 (`go.mod`). Built-ins `min`/`max` available; do not redefine them.
- All TUI code lives in `internal/tui/` (single package, no sub-packages) — sub-packages create import cycles.
- Key files: `app.go` (root model), `styles.go` (Lipgloss theme), `keys.go` (keybindings), `messages.go` (EventBus + message types), `tab_*.go` (per-tab sub-models).
- `--no-tui` flag in `main.go` enables headless/CI mode (plain zerolog output).
- `logger.NewWithWriter(level, format, io.Writer)` — use to redirect logs into TUI log capture.
- i18n: `internal/i18n/` — `i18n.T()` returns `*Locale` (atomic, thread-safe). `FieldDef.Section/Label/Tooltip` are `func() string` closures; do NOT cache them. On `LanguageChangedMsg`, rebuild only table models (orders/positions/copytrading) — not `SettingsModel`, which already holds updated state.
- i18n locales: `internal/i18n/locales/{en,ru,zh,ja,ko}.json` — when adding a new locale, populate all fields from `Locale` struct. `i18n.Available()` returns list of loaded language codes.
- `internal/config/watcher.go` — `config.NewWatcher(path, onReload func(*Config))` with 300ms debounce; used by copytrading for hot-reload of trader list.
- `docs/` — gitignored; design docs stored locally only, not committed.
- Wizard (`internal/tui/wizard.go`) is **single-step** — private_key only; L2 credentials derived automatically at startup.
- **Adding/removing a setting**: (1) TUI: `allFields []FieldDef` in `tab_settings.go` + `Locale` struct in `locale.go` + 5 `internal/i18n/locales/*.json`; (2) WebUI: `applyConfigKey()` in `config_key.go` + `SettingsView.vue` + 5 `src/i18n/*.json`; (3) `npm run build`.

## Shared Utilities & Patterns

- `config.Save(path, cfg)` — canonical TOML persist; used by `tab_settings.go` and `internal/telegrambot/handlers.go`. Do NOT add a private `saveConfig` helper.
- `EventBus` (`internal/tui/messages.go`) — `Send(msg)` broadcasts to TUI + all `Tap()` subscribers. `Tap()` returns a `<-chan tea.Msg` for fan-out; used by `internal/telegrambot/`.
- `TradesMonitor.SetBus(bus)` — call after `NewTradesMonitor`, before `Run(ctx)`; emits `OrdersUpdateMsg`/`PositionsUpdateMsg` after each poll.
- `go-telegram-bot-api/v5` is in `go.mod` as an indirect dep — do NOT `go get` it again; just import and run `go mod tidy`.
- **Multi-wallet interfaces**: `internal/wallet` defines `WalletProvider`; consumers (TUI, webui, telegrambot) each define their own minimal `WalletMutator`/`WalletProvider` interface to avoid import cycles. EventBus messages (e.g. `WalletStatsMsg`) carry `Label`+`Enabled` so consumers display full context without importing `internal/wallet`.

## Testing

- Unit tests (no tag): `go test ./...`
- Integration tests (real Polymarket API): `POLY_PRIVATE_KEY=0x... go test ./... -tags=integration -timeout 90s`
- `internal/testutil/` — shared helpers: `NewCLOBClient`, `NewGammaClient`, `NewDataClient`, `LoadL1Signer`, `LoadL2Creds` (derives L2 via `GET /auth/derive-api-key`)
- `POLY_PRIVATE_KEY` env var: accepts `0x`-prefixed hex; testutil strips prefix and auto-derives L2 credentials
- Network tests are flaky in this dev environment (MITM proxy blocks TLS intermittently)

## API Quirks

- **Gamma**: `liquidity` and `volume` fields may be JSON strings instead of numbers. Use `flexFloat64` type (in `internal/api/gamma/models.go`) for any numeric field that exhibits this behaviour.
- **Gamma**: `outcomes`, `outcomePrices`, `clobTokenIds` may also be JSON-encoded strings instead of arrays. Use `flexStringSlice` type (same file) for any array field that exhibits this.
- **Gamma query params**: Use `limit`/`offset`/`order`/`ascending` — NOT `_limit`/`_offset`/`_sort`/`_order`.
- **Gamma `GET /markets/{id}`**: `{id}` is `Market.ID` (numeric string), NOT `Market.ConditionID` (hex).

## Web UI (Frontend)

- Source: `internal/webui/web/src/` — Vue 3 + Vite + Pinia + vue-router + vue-i18n v11
- **Rebuild**: `cd internal/webui/web && npm run build` — outputs to `web/dist/` (embedded by Go)
- **Dev server**: `cd internal/webui/web && npm run dev` — proxies `/api` and `/ws` to `localhost:8080`
- `vue-i18n` must be **v11** (`legacy: false`); v9/v10 are EOL. Do NOT downgrade.
- i18n locales: `src/i18n/{en,ru,zh,ja,ko}.json` — parallel to Go's `internal/i18n/locales/`
- **Adding a new Vue page**: (1) add route to `router/index.js`, (2) add nav item to `AppHeader.vue` `navItems`, (3) add `nav.pagename` key + full `pagename.*` section to all 5 `src/i18n/*.json` locale files, (4) run `npm run build`.
- node_modules: `internal/webui/web/node_modules/` (gitignored)

## Polymarket API Reference

See `POLYMARKET_DOCS.md` for the full API reference (endpoints, rate limits, WebSocket channels, auth flows, contract addresses).

## Frontend Design

- Theme: CSS custom properties dark/light; JetBrains Mono (mono) + DM Sans (UI) — avoid Inter/Roboto/Arial
- Colors: deep navy backgrounds (`#0d1117`, `#161b22`), electric blue accent (`#58a6ff` dark / `#0969da` light)
- Avoid "AI slop": no purple gradients, no generic layouts, no Space Grotesk by default
