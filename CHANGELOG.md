# Changelog

All notable changes to Orbitron are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

---

## [1.0.0] — 2025-03-14

### Added

**Core**
- Seven context-cancellable subsystems: WebSocket client, Monitor, Trading Engine, Copy Trader, Notifier, Telegram Bot, Web UI
- Graceful shutdown with signal handling (SIGINT / SIGTERM)
- Config hot-reload via fsnotify with 300 ms debounce (language, copytrading traders list)

**Trading Engine**
- Pluggable `trading.Strategy` interface with goroutine-per-strategy execution model
- Six built-in strategies: Arbitrage, Cross-Market, Fade Chaos, Market Making, Positive EV, Riskless Rate
- Global risk manager: stop-loss, take-profit, max daily loss limits

**Authentication**
- L1 EIP-712 signing (go-ethereum) for API key derivation
- L2 HMAC-SHA256 signing for all authenticated CLOB requests
- 30-second signature expiry; keys derived in-memory, never stored on disk

**Multi-Wallet**
- `[[wallets]]` array in config — multiple wallets with per-wallet `enabled`/`primary` flags
- Wallet manager with aggregated P&L view across all active wallets
- Environment variable overlay for secrets (`POLY_PRIVATE_KEY`, `POLY_API_KEY`, etc.)

**Copy Trading**
- Data API polling with configurable `poll_interval_ms`
- Two size modes: `proportional` and `fixed_pct`
- SQLite persistence for tracked traders and copied positions

**Monitor & Alerts**
- Trades monitor: open orders, fills, position tracking
- Market monitor: rule-based price alerts evaluated against Gamma API diffs
- Telegram delivery for all alert types

**Terminal UI (TUI)**
- BubbleTea v1.3.10 — tabs: Markets, Trading, Copy Trading, Wallets, Strategies, Settings, Logs
- First-run wizard — generates `config.toml` interactively
- Log capture with ring buffer; real-time log tab inside TUI
- `--no-tui` flag for headless/server mode

**Web UI**
- Vue 3 + Vite + Pinia + vue-router + vue-i18n SPA
- Embedded into the Go binary via `embed.FS`
- JWT authentication; WebSocket hub for real-time updates
- Views: Overview, Markets, Orders, Positions, Copytrading, Wallets, Settings, Logs

**Telegram Bot**
- Inline keyboard navigation mirroring TUI functionality
- Single-admin model via `admin_chat_id`
- Multi-step conversation flows for order placement and configuration

**Internationalization**
- Five languages: English, Russian, Chinese, Japanese, Korean
- Hot-reload on `ui.language` config change — no restart required

**Infrastructure**
- `setup.sh` — universal setup script (Linux, macOS, Windows via Git Bash/WSL)
- SQLite storage layer with migration support
- Proxy support (HTTP / SOCKS5)
- Structured logging via zerolog with `pretty` and `json` formats
- Builder Program license integration via ldflags embedding

[Unreleased]: https://github.com/atlas-is-coding/orbitron-polymarket-system/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/atlas-is-coding/orbitron-polymarket-system/releases/tag/v1.0.0
