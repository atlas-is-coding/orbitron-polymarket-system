# Polytrade Bot 📈🤖

[![Go Report Card](https://goreportcard.com/badge/github.com/polytrade/bot)](https://goreportcard.com/report/github.com/polytrade/bot)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

*Read this in other languages: [Русский](README_ru.md), [中文](README_zh.md), [한국어](README_ko.md), [日本語](README_ja.md).*

Polytrade Bot is an advanced algorithmic trading and management bot for the **Polymarket CTF Exchange**. It features a robust multi-interface architecture including an interactive Terminal User Interface (TUI), a Vue 3 Web UI, and a comprehensive Telegram bot for remote management.

## 🌟 Comprehensive Features

*   **Multi-Interface Experience:**
    *   **TUI:** A beautiful terminal interface with separate tabs for Markets, Trading, Copytrading, Wallets, Strategies, Settings, and Logs.
    *   **Web UI:** A Vue 3 SPA with real-time WebSocket updates, JWT authentication, and a responsive design.
    *   **Telegram Bot:** An interactive bot using inline keyboards and multi-step conversations that mirrors the TUI functionality.
*   **Algorithmic Trading Engine:** Built-in strategies including Arbitrage, Cross-Market, Fade Chaos, Market Making, Positive EV, and Riskless Rate. Easily register new custom strategies.
*   **Advanced Copy Trading:** Monitor target wallets via the Data API and automatically copy positions via the CLOB API. Supports dynamic size-mode allocations (`proportional` or `fixed_pct`).
*   **Real-Time Monitoring & Alerts:**
    *   **Trades Monitor:** Tracks open orders, trade fills, and positions.
    *   **Market Alerts:** Evaluates real-time alert conditions against market state diffs.
*   **Secure Authentication:** L1/L2 credentials architecture. Automatic EIP-712 signature derivation—L2 credentials are kept entirely in memory and are never stored in your config file. Signatures expire automatically in 30 seconds for security.
*   **Multi-Wallet Support:** Manage multiple active wallets, toggle them on/off, and view aggregated stats directly from any of the UIs.
*   **Internationalization (i18n):** Native, hot-reloading support for English, Russian, Chinese, Japanese, and Korean across all interfaces.

## 🏗 Architecture Overview

The bot operates across seven core, context-cancellable subsystems:

1.  **WebSocket Client:** Persistent connections with auto-reconnect to Polymarket CLOB (`market`, `user`, `asset` channels).
2.  **Monitor:** Polls Gamma & Data API for market state diffs and evaluates real-time alert conditions.
3.  **Trading Engine:** Scalable Goroutine-based execution layer for pluggable trading strategies (`trading.Strategy`).
4.  **Notifier:** Configurable alerting system (defaulting to Telegram).
5.  **Copy Trader:** Tracks configured wallets and replicates their positions. Hot-reloads on configuration changes without restarting the bot.
6.  **Telegram Bot:** Interactive mirror of the TUI using a single-admin model (`AdminChatID`).
7.  **Web UI:** Embedded HTTP server + WebSocket hub serving a Vue 3 SPA.

## ⚙️ Configuration (`config.toml`)

The bot is controlled entirely by `config.toml`. Trading and database features are disabled by default for safety.

Key sections include:
*   `[auth]`: Requires `private_key` (hex, no `0x` prefix). L2 credentials are auto-derived at startup.
*   `[webui]`: `enabled` (true/false), `listen` (e.g., `127.0.0.1:8080`), `jwt_secret` (used for signing and as the login password).
*   `[ui]`: `language` (`en`, `ru`, `zh`, `ja`, `ko`). Hot-reloads instantly.
*   `[monitor.trades]`: `enabled`, `poll_interval_ms`. Requires L2 auth.
*   `[copytrading]`: `enabled`, `size_mode` (`proportional`/`fixed_pct`), and the `[[copytrading.traders]]` list. Requires database and L2 auth.
*   `[telegram]`: `enabled`, `bot_token`, `admin_chat_id` (single admin target).
*   `[database]`: `enabled`, `path` (SQLite DB path).
*   `chain_id`: `137` for Polygon Mainnet, `80002` for Amoy Testnet.

## 🚀 Installation & Setup

### Prerequisites

*   [Go 1.24+](https://golang.org/doc/install)
*   [Node.js 18+](https://nodejs.org/) (Only needed if modifying the Web UI)
*   Polymarket Wallet Private Key (for L1 Signature derivation)

### Setup Steps

#### Option 1: Universal Setup Script (Recommended)
We provide a universal `setup.sh` script that works on Linux, macOS, and Windows (via Git Bash/WSL). It automatically installs Go and Node.js (if missing), sets up your `config.toml`, builds the Vue 3 frontend, and compiles the Go backend.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **Run the setup script:**
    ```bash
    ./setup.sh
    ```

#### Option 2: Manual Setup
1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **Configure the Bot:**
    Create a `config.toml` file in the root directory. You can start the bot without it to use the TUI wizard, which will securely help you configure your `private_key`.

3.  **Build and Run:**
    ```bash
    # Build the binary
    go build ./...

    # Run the bot
    go run ./cmd/bot/ --config config.toml
    ```

### Headless Mode
To run the bot in a server environment without the TUI, use the headless flag:
```bash
go run ./cmd/bot/ --config config.toml --no-tui
```

## 🛠 Troubleshooting & Common Issues

*   **API Key / 401 Unauthorized:** Ensure your `private_key` is correct. The bot automatically derives the L2 API key, secret, and passphrase at startup. L2 signatures expire in 30 seconds; ensure your system clock is synchronized.
*   **Web UI "Network Error":** If a Go HTTP handler panics, the browser reports a generic "Network Error" because Go closes the TCP connection without a JSON body. Check the terminal logs for the actual Go panic stack trace.
*   **Missing Market Data in Web UI / TUI:** The internal EventBus drops messages silently if the buffer is full. If you have logging set to `trace` and the bot is producing too many logs, important messages like `MarketsUpdatedMsg` might be dropped. Reduce log level to `info` or `debug`.
*   **WebSocket "Bad Handshake":** The bot connects to specific channels (`.../ws/market`), not the root WS URL. This is handled internally, but verify your firewall/network allows WebSocket connections to Polymarket.
*   **Polymarket Token IDs Parsing:** Token IDs from Polymarket Gamma API are decimal strings. The bot parses them correctly using `big.Int.SetString(id, 10)`. Do not attempt to parse them as hex directly in your own scripts.

## 💻 Development Guide

### Building the Web UI
The Vue 3 Web UI is embedded into the Go binary. If you modify the files in `internal/webui/web/src`, you must rebuild the frontend for the Go binary to pick up the changes:
```bash
cd internal/webui/web
npm install
npm run build
```

### Extending the Bot
*   **New Trading Strategy:** Implement the `trading.Strategy` interface (`Name`, `Start`, `Stop`), instantiate it in `main.go`, and call `engine.Register(s)`.
*   **New Configuration Setting:** Add the field to `tab_settings.go`, `Locale` struct, update the 5 `locales/*.json` files, and add logic to `applyConfigKey()` in `config_key.go`. Update the Vue UI in `SettingsView.vue` and its locale files.
*   **New Telegram Command:** Add the handler in `internal/telegrambot/handlers.go` within the `handleCommand` switch statement.

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests (requires real Polymarket API & L1 Key)
POLY_PRIVATE_KEY=0xYOUR_KEY go test ./... -tags=integration -timeout 90s
```

## 📜 License
This project is licensed under the MIT License - see the LICENSE file for details.
