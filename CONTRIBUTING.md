# Contributing to Orbitron

Thank you for your interest in contributing. This guide covers everything you need to get started.

---

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 18+ (for Web UI changes only)

### Clone and Build

```bash
git clone https://github.com/atlas-is-coding/orbitron-polymarket-system.git
cd orbitron-polymarket-system

# Copy example config
cp config.toml.example config.toml
# Edit config.toml — add your private_key for integration tests

# Run (TUI wizard appears if config is incomplete)
go run ./cmd/bot/ --config config.toml
```

### Rebuild Web UI (if modifying frontend)

```bash
cd internal/webui/web
npm install
npm run build
```

---

## Running Tests

```bash
# Unit tests
go test ./...

# Integration tests — requires a real Polymarket key
POLY_PRIVATE_KEY=0xYOUR_KEY go test ./... -tags=integration -timeout 90s

# Vet
go vet ./...
```

---

## Project Structure

```
cmd/bot/            — entry point, subsystem wiring, graceful shutdown
internal/
  api/              — HTTP clients: clob/, gamma/, data/, ws/
  auth/             — L1 EIP-712 and L2 HMAC-SHA256 signing
  config/           — Config struct, TOML parsing, fsnotify watcher
  copytrading/      — Copy trader: tracker, sizer, executor
  health/           — Geo-block and service health checks
  i18n/             — Locale loading and translation helpers
  logger/           — zerolog wrapper with TUI writer support
  markets/          — Market service (Gamma polling, price alerts)
  monitor/          — Trade monitor (CLOB polling, fills, positions)
  notify/           — Notifier interface + Telegram implementation
  storage/          — Storage interface + SQLite implementation
  telegrambot/      — Telegram bot: handlers, renderer, state machine
  trading/          — Strategy interface, engine, risk manager
    strategies/     — Six built-in trading strategies
  tui/              — BubbleTea TUI: app, tabs, wizard, styles
  wallet/           — Wallet manager, poller
  webui/            — HTTP server, WebSocket hub, Vue 3 frontend
```

---

## Adding a Trading Strategy

1. Create `internal/trading/strategies/my_strategy.go`
2. Implement the `trading.Strategy` interface:

```go
type MyStrategy struct {
    cfg MyStrategyConfig
    log zerolog.Logger
}

func (s *MyStrategy) Name() string                    { return "my_strategy" }
func (s *MyStrategy) Start(ctx context.Context) error { /* main loop */ return nil }
func (s *MyStrategy) Stop()                           { /* cleanup */ }
```

3. Add config fields to `internal/config/config.go` under `TradingStrategies`
4. Register in `cmd/bot/main.go`: `engine.Register(strategies.NewMyStrategy(cfg, log))`
5. Add i18n strings to all five locale files in `internal/i18n/locales/`

---

## Adding a Config Setting

1. Add field to the relevant struct in `internal/config/config.go`
2. Add to `config.toml.example` with a comment
3. Add i18n key to all five locale files: `internal/i18n/locales/{en,ru,zh,ja,ko}.json`
4. Update TUI: `internal/tui/tab_settings.go`
5. Update Web UI: `internal/webui/web/src/views/SettingsView.vue` and its i18n files

---

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet` must pass)
- Use `zerolog` for all logging — no `fmt.Println` in production paths
- Keep packages focused — `internal/tui/` is a single package to avoid import cycles
- Error messages: lowercase, no punctuation at the end
- Test files alongside source: `foo_test.go` in the same package

---

## Pull Request Process

1. Fork the repository and create a branch from `main`
2. Make your changes with tests
3. Run `go test ./...` and `go vet ./...` — both must pass
4. Update `CHANGELOG.md` under `[Unreleased]`
5. Submit a PR against `main` using the PR template

---

## Reporting Security Issues

Do **not** open a public issue for security vulnerabilities. Contact the maintainers directly via the email listed on [getorbitron.net](https://getorbitron.net).
