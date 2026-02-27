# Telegram Bot Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an interactive Telegram Bot to polytrade-bot that mirrors the Console TUI, with bidirectional state synchronization via a shared fan-out EventBus.

**Architecture:** New package `internal/telegrambot/` with 4 files (bot.go, handlers.go, renderer.go, state.go). EventBus gains fan-out support (`Tap()` method). Bot and TUI share the same EventBus, each reading from their own subscription channel. Config changes in either direction propagate through `ConfigReloadedMsg`. Order cancellations are delegated to `TradesMonitor` via an interface.

**Tech Stack:** `github.com/go-telegram-bot-api/telegram-bot-api/v5` (polling mode), existing `tui.EventBus`, `config.Save()` (new), zerolog.

---

## Task 1: Add go-telegram-bot-api/v5 dependency

**Files:**
- Modify: `go.mod`, `go.sum`

**Step 1: Add the dependency**

```bash
cd /path/to/polytrade-bot
go get github.com/go-telegram-bot-api/telegram-bot-api/v5@v5.5.1
go mod tidy
```

**Step 2: Verify go.mod has the new entry**

```bash
grep "telegram-bot-api" go.mod
```
Expected: `github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: add go-telegram-bot-api/v5 dependency"
```

---

## Task 2: Add `config.Save()` and extend `TelegramConfig`

**Files:**
- Modify: `internal/config/config.go`

**Step 1: Write the failing test**

Create `internal/config/save_test.go`:
```go
package config_test

import (
    "os"
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/config"
)

func TestSave_RoundTrip(t *testing.T) {
    f, err := os.CreateTemp(t.TempDir(), "cfg-*.toml")
    if err != nil {
        t.Fatal(err)
    }
    f.Close()

    cfg := &config.Config{}
    cfg.API.ClobURL = "https://clob.polymarket.com"
    cfg.UI.Language = "ru"

    if err := config.Save(f.Name(), cfg); err != nil {
        t.Fatalf("Save: %v", err)
    }

    loaded, err := config.Load(f.Name())
    if err != nil {
        t.Fatalf("Load after Save: %v", err)
    }
    if loaded.UI.Language != "ru" {
        t.Errorf("Language = %q, want %q", loaded.UI.Language, "ru")
    }
}

func TestSave_AdminChatIDRoundTrip(t *testing.T) {
    f, err := os.CreateTemp(t.TempDir(), "cfg-*.toml")
    if err != nil {
        t.Fatal(err)
    }
    f.Close()

    cfg := &config.Config{}
    cfg.API.ClobURL = "https://clob.polymarket.com"
    cfg.Telegram.AdminChatID = "999888"

    if err := config.Save(f.Name(), cfg); err != nil {
        t.Fatalf("Save: %v", err)
    }

    loaded, err := config.Load(f.Name())
    if err != nil {
        t.Fatalf("Load: %v", err)
    }
    if loaded.Telegram.AdminChatID != "999888" {
        t.Errorf("AdminChatID = %q, want %q", loaded.Telegram.AdminChatID, "999888")
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/config/... -run "TestSave|TestAdmin" -v
```
Expected: FAIL — `config.Save undefined`

**Step 3: Add `AdminChatID` to `TelegramConfig` and add `config.Save()`**

In `internal/config/config.go`, change `TelegramConfig`:
```go
type TelegramConfig struct {
    Enabled     bool   `toml:"enabled"`
    BotToken    string `toml:"bot_token"`
    ChatID      string `toml:"chat_id"`
    AdminChatID string `toml:"admin_chat_id"` // optional; if set, this user can edit secret fields
}
```

Add at the bottom of `internal/config/config.go`:
```go
// Save serialises cfg to TOML at path, creating or overwriting the file.
func Save(path string, cfg *Config) error {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("config: save %q: %w", path, err)
    }
    defer f.Close()
    return toml.NewEncoder(f).Encode(cfg)
}
```

**Step 4: Run tests**

```bash
go test ./internal/config/... -run "TestSave|TestAdmin" -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/save_test.go
git commit -m "feat(config): add Save(), add TelegramConfig.AdminChatID"
```

---

## Task 3: Fan-out EventBus with `Tap()` method

The current `EventBus` sends to a single channel. We need both TUI and Telegram Bot to receive all messages. Solution: add `Tap()` which returns a new subscriber channel that receives copies of all future `Send()` calls.

**Files:**
- Modify: `internal/tui/messages.go`

**Step 1: Write the failing test**

Create `internal/tui/messages_test.go`:
```go
package tui_test

import (
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestEventBus_Tap_ReceivesCopy(t *testing.T) {
    bus := tui.NewEventBus()
    tap := bus.Tap()

    bus.Send(tui.BotEventMsg{Level: "info", Message: "hello"})

    select {
    case msg := <-tap:
        evt, ok := msg.(tui.BotEventMsg)
        if !ok {
            t.Fatalf("got %T, want BotEventMsg", msg)
        }
        if evt.Message != "hello" {
            t.Errorf("Message = %q, want %q", evt.Message, "hello")
        }
    default:
        t.Fatal("tap channel empty after Send")
    }
}

func TestEventBus_Tap_IndependentOfMain(t *testing.T) {
    bus := tui.NewEventBus()
    tap := bus.Tap()

    // Send 3 messages
    for i := 0; i < 3; i++ {
        bus.Send(tui.BotEventMsg{Level: "info", Message: "msg"})
    }

    // tap should have 3, main channel should also have 3
    tapCount := 0
    for {
        select {
        case <-tap:
            tapCount++
        default:
            goto done
        }
    }
done:
    if tapCount != 3 {
        t.Errorf("tap received %d messages, want 3", tapCount)
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/tui/... -run "TestEventBus" -v
```
Expected: FAIL — `bus.Tap undefined`

**Step 3: Extend EventBus in `internal/tui/messages.go`**

Replace the `EventBus` struct and related code (keep `WaitForEvent` unchanged):
```go
// EventBus bridges bot goroutines to the Bubble Tea loop.
// Supports multiple subscribers via Tap(); the primary channel is
// used by the TUI via WaitForEvent().
type EventBus struct {
    ch   chan tea.Msg
    mu   sync.Mutex
    taps []chan tea.Msg
}
```

Add `"sync"` to imports.

Change `NewEventBus()`:
```go
func NewEventBus() *EventBus {
    return &EventBus{ch: make(chan tea.Msg, 512)}
}
```

Change `Send()` to also broadcast to taps:
```go
func (b *EventBus) Send(msg tea.Msg) {
    select {
    case b.ch <- msg:
    default:
    }
    b.mu.Lock()
    for _, tap := range b.taps {
        select {
        case tap <- msg:
        default:
        }
    }
    b.mu.Unlock()
}
```

Add `Tap()`:
```go
// Tap creates a new subscriber channel that receives a copy of every
// future Send() call. The caller is responsible for draining it.
func (b *EventBus) Tap() <-chan tea.Msg {
    ch := make(chan tea.Msg, 512)
    b.mu.Lock()
    b.taps = append(b.taps, ch)
    b.mu.Unlock()
    return ch
}
```

Keep `WaitForEvent()` unchanged (reads from `b.ch`).

**Step 4: Run tests**

```bash
go test ./internal/tui/... -run "TestEventBus" -v
```
Expected: PASS

**Step 5: Make sure existing code still compiles**

```bash
go build ./...
```

**Step 6: Commit**

```bash
git add internal/tui/messages.go internal/tui/messages_test.go
git commit -m "feat(tui): add EventBus.Tap() for multi-subscriber fan-out"
```

---

## Task 4: Add `OrdersUpdateMsg` and `PositionsUpdateMsg` to EventBus types

**Files:**
- Modify: `internal/tui/messages.go`

**Step 1: Add new message types at the end of the messages block in `messages.go`**

After `LanguageChangedMsg`:
```go
// OrdersUpdateMsg carries a fresh snapshot of open orders from TradesMonitor.
type OrdersUpdateMsg struct {
    Rows []OrderRow
}

// PositionsUpdateMsg carries a fresh snapshot of positions from TradesMonitor.
type PositionsUpdateMsg struct {
    Rows []PositionRow
}
```

`OrderRow` and `PositionRow` are already defined in `tab_orders.go` and `tab_positions.go` (same `tui` package), so no import needed.

**Step 2: Build**

```bash
go build ./...
```
Expected: success (new types, no usage yet)

**Step 3: Commit**

```bash
git add internal/tui/messages.go
git commit -m "feat(tui): add OrdersUpdateMsg, PositionsUpdateMsg event types"
```

---

## Task 5: Wire `TradesMonitor` to emit `OrdersUpdateMsg` / `PositionsUpdateMsg`

`TradesMonitor` needs an optional EventBus to push snapshots after each poll. We pass `bus *tui.EventBus` — if nil, no push (backward compat).

**Files:**
- Modify: `internal/monitor/trades.go`

> Note: This creates an import of `tui` from `monitor`. Both are internal packages; no cycle.

**Step 1: Write the failing test**

Create `internal/monitor/trades_bus_test.go`:
```go
package monitor_test

import (
    "context"
    "testing"
    "time"

    "github.com/atlasdev/polytrade-bot/internal/monitor"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestTradesMonitor_SetBus(t *testing.T) {
    // Just verify SetBus doesn't panic with nil TradesMonitor fields
    bus := tui.NewEventBus()
    tap := bus.Tap()

    tm := &monitor.TradesMonitor{}
    tm.SetBus(bus)

    // Send a synthetic event (SetBus itself shouldn't send)
    bus.Send(tui.BotEventMsg{Level: "info", Message: "test"})

    select {
    case msg := <-tap:
        if _, ok := msg.(tui.BotEventMsg); !ok {
            t.Fatalf("unexpected message type %T", msg)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("tap did not receive forwarded message")
    }
}
```

> Note: This test is minimal — just verifies the exported `SetBus` method exists and the bus works. Full integration test requires a mock CLOB client.

**Step 2: Run to verify it fails**

```bash
go test ./internal/monitor/... -run "TestTradesMonitor_SetBus" -v
```
Expected: FAIL — `tm.SetBus undefined` (and `TradesMonitor` fields unexported)

**Step 3: Modify `TradesMonitor` in `internal/monitor/trades.go`**

Add import:
```go
"github.com/atlasdev/polytrade-bot/internal/tui"
```

Add `bus` field to `TradesMonitor` struct:
```go
type TradesMonitor struct {
    // ... existing fields ...
    bus *tui.EventBus // optional; if set, emits OrdersUpdateMsg/PositionsUpdateMsg
}
```

Add `SetBus` method (after `NewTradesMonitor`):
```go
// SetBus wires an EventBus so TradesMonitor pushes data snapshots after each poll.
// Call before Run(). Pass nil to disable.
func (tm *TradesMonitor) SetBus(bus *tui.EventBus) {
    tm.bus = bus
}
```

At the end of `pollOrders()`, after `tm.mu.Unlock()`, add:
```go
if tm.bus != nil {
    rows := make([]tui.OrderRow, 0, len(tm.orders))
    for _, o := range tm.orders {
        rows = append(rows, tui.OrderRow{
            Market: o.AssetID,
            Side:   string(o.Side),
            Price:  o.Price,
            Size:   o.OriginalSize,
            Filled: o.SizeMatched,
            Status: string(o.Status),
            Age:    "",
            ID:     o.ID,
        })
    }
    tm.bus.Send(tui.OrdersUpdateMsg{Rows: rows})
}
```

At the end of `pollPositions()`, after `tm.mu.Unlock()`, add:
```go
if tm.bus != nil {
    rows := make([]tui.PositionRow, 0, len(tm.positions))
    for _, p := range tm.positions {
        rows = append(rows, tui.PositionRow{
            Market:  p.AssetID,
            Side:    string(p.Side),
            Size:    p.Size,
            Entry:   p.AvgPrice,
            Current: "",
            PnL:     "",
            PnLPct:  "",
        })
    }
    tm.bus.Send(tui.PositionsUpdateMsg{Rows: rows})
}
```

Note: `clob.Order` and `clob.Position` field names — check `internal/api/clob/models.go` before filling in. Use the same fields as `tab_orders.go` SetOrderRows caller.

**Step 4: Run test**

```bash
go test ./internal/monitor/... -run "TestTradesMonitor_SetBus" -v
go build ./...
```
Expected: PASS, clean build

**Step 5: Commit**

```bash
git add internal/monitor/trades.go internal/monitor/trades_bus_test.go
git commit -m "feat(monitor): wire TradesMonitor to emit OrdersUpdateMsg/PositionsUpdateMsg"
```

---

## Task 6: Create `internal/telegrambot/state.go`

Thread-safe cache of latest bot data, updated by EventBus consumer.

**Files:**
- Create: `internal/telegrambot/state.go`

**Step 1: Write the failing test**

Create `internal/telegrambot/state_test.go`:
```go
package telegrambot_test

import (
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/telegrambot"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestBotState_Orders(t *testing.T) {
    s := telegrambot.NewBotState()
    rows := []tui.OrderRow{
        {Market: "BTC-USD", Side: "BUY", Price: "45000", Size: "1.0", ID: "abc"},
    }
    s.SetOrders(rows)
    got := s.Orders()
    if len(got) != 1 || got[0].ID != "abc" {
        t.Errorf("Orders() = %v, want 1 order with ID abc", got)
    }
}

func TestBotState_Logs(t *testing.T) {
    s := telegrambot.NewBotState()
    for i := 0; i < 55; i++ {
        s.AddLog("line")
    }
    logs := s.Logs()
    if len(logs) != 50 {
        t.Errorf("Logs() len = %d, want 50 (capped)", len(logs))
    }
}

func TestBotState_SubsystemStatus(t *testing.T) {
    s := telegrambot.NewBotState()
    s.SetSubsystem("WebSocket", true)
    s.SetSubsystem("Monitor", false)
    statuses := s.Subsystems()
    active := 0
    for _, st := range statuses {
        if st.Active {
            active++
        }
    }
    if active != 1 {
        t.Errorf("active subsystems = %d, want 1", active)
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/telegrambot/... -run "TestBotState" -v
```
Expected: FAIL — package doesn't exist

**Step 3: Create `internal/telegrambot/state.go`**

```go
package telegrambot

import (
    "sync"

    "github.com/atlasdev/polytrade-bot/internal/tui"
)

// SubsystemStatus holds name + active state.
type SubsystemStatus struct {
    Name   string
    Active bool
}

// BotState is a thread-safe cache of the latest bot data,
// updated by the EventBus consumer goroutine.
type BotState struct {
    mu         sync.RWMutex
    balance    float64
    orders     []tui.OrderRow
    positions  []tui.PositionRow
    traders    []tui.TraderRow
    logs       []string
    subsystems map[string]bool
}

// NewBotState creates an empty BotState.
func NewBotState() *BotState {
    return &BotState{subsystems: make(map[string]bool)}
}

func (s *BotState) SetBalance(v float64) {
    s.mu.Lock()
    s.balance = v
    s.mu.Unlock()
}

func (s *BotState) Balance() float64 {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.balance
}

func (s *BotState) SetOrders(rows []tui.OrderRow) {
    s.mu.Lock()
    s.orders = rows
    s.mu.Unlock()
}

func (s *BotState) Orders() []tui.OrderRow {
    s.mu.RLock()
    defer s.mu.RUnlock()
    cp := make([]tui.OrderRow, len(s.orders))
    copy(cp, s.orders)
    return cp
}

func (s *BotState) SetPositions(rows []tui.PositionRow) {
    s.mu.Lock()
    s.positions = rows
    s.mu.Unlock()
}

func (s *BotState) Positions() []tui.PositionRow {
    s.mu.RLock()
    defer s.mu.RUnlock()
    cp := make([]tui.PositionRow, len(s.positions))
    copy(cp, s.positions)
    return cp
}

func (s *BotState) SetTraders(rows []tui.TraderRow) {
    s.mu.Lock()
    s.traders = rows
    s.mu.Unlock()
}

func (s *BotState) Traders() []tui.TraderRow {
    s.mu.RLock()
    defer s.mu.RUnlock()
    cp := make([]tui.TraderRow, len(s.traders))
    copy(cp, s.traders)
    return cp
}

// AddLog appends a log line and caps the buffer at 50 lines.
func (s *BotState) AddLog(line string) {
    s.mu.Lock()
    s.logs = append(s.logs, line)
    if len(s.logs) > 50 {
        s.logs = s.logs[len(s.logs)-50:]
    }
    s.mu.Unlock()
}

func (s *BotState) Logs() []string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    cp := make([]string, len(s.logs))
    copy(cp, s.logs)
    return cp
}

func (s *BotState) SetSubsystem(name string, active bool) {
    s.mu.Lock()
    s.subsystems[name] = active
    s.mu.Unlock()
}

func (s *BotState) Subsystems() []SubsystemStatus {
    s.mu.RLock()
    defer s.mu.RUnlock()
    result := make([]SubsystemStatus, 0, len(s.subsystems))
    for name, active := range s.subsystems {
        result = append(result, SubsystemStatus{Name: name, Active: active})
    }
    return result
}
```

**Step 4: Run tests**

```bash
go test ./internal/telegrambot/... -run "TestBotState" -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/telegrambot/state.go internal/telegrambot/state_test.go
git commit -m "feat(telegrambot): add thread-safe BotState cache"
```

---

## Task 7: Create `internal/telegrambot/renderer.go`

Formats data as HTML strings for Telegram messages.

**Files:**
- Create: `internal/telegrambot/renderer.go`
- Create: `internal/telegrambot/renderer_test.go`

**Step 1: Write the failing test**

Create `internal/telegrambot/renderer_test.go`:
```go
package telegrambot_test

import (
    "strings"
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/telegrambot"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestRenderOverview(t *testing.T) {
    subsystems := []telegrambot.SubsystemStatus{
        {Name: "WebSocket", Active: true},
        {Name: "Monitor", Active: false},
    }
    result := telegrambot.RenderOverview(100.50, subsystems, 3, 2)
    if !strings.Contains(result, "100.50") {
        t.Error("balance missing from overview")
    }
    if !strings.Contains(result, "WebSocket") {
        t.Error("subsystem name missing")
    }
}

func TestRenderOrders_Empty(t *testing.T) {
    result := telegrambot.RenderOrders(nil)
    if !strings.Contains(result, "No") && !strings.Contains(result, "empty") && !strings.Contains(result, "нет") {
        // just check it returns something reasonable
        if result == "" {
            t.Error("RenderOrders(nil) returned empty string")
        }
    }
}

func TestRenderOrders_WithRows(t *testing.T) {
    rows := []tui.OrderRow{
        {Market: "BTC-USD", Side: "BUY", Price: "45000", Size: "1.0", Status: "LIVE", ID: "abc123"},
    }
    result := telegrambot.RenderOrders(rows)
    if !strings.Contains(result, "BTC-USD") {
        t.Error("market missing from orders render")
    }
    if !strings.Contains(result, "abc123") {
        t.Error("order ID missing — needed for cancel button")
    }
}

func TestRenderSettings_MasksSecrets(t *testing.T) {
    result := telegrambot.RenderSettings(
        "Auth",
        map[string]string{"api_key": "mykey", "private_key": "mysecret"},
        false, // not admin
    )
    if strings.Contains(result, "mysecret") {
        t.Error("secret leaked for non-admin")
    }
    if strings.Contains(result, "mykey") {
        t.Error("api_key leaked for non-admin")
    }
}

func TestRenderSettings_AdminSeesSecrets(t *testing.T) {
    result := telegrambot.RenderSettings(
        "Auth",
        map[string]string{"api_key": "mykey", "private_key": "mysecret"},
        true, // admin
    )
    if !strings.Contains(result, "mykey") {
        t.Error("admin should see api_key")
    }
}
```

**Step 2: Run to verify it fails**

```bash
go test ./internal/telegrambot/... -run "TestRender" -v
```
Expected: FAIL

**Step 3: Create `internal/telegrambot/renderer.go`**

```go
package telegrambot

import (
    "fmt"
    "strings"

    "github.com/atlasdev/polytrade-bot/internal/tui"
)

// secretFields lists field labels that are masked for non-admins.
var secretFields = map[string]bool{
    "private_key": true,
    "api_key":     true,
    "api_secret":  true,
    "passphrase":  true,
    "bot_token":   true,
    "chat_id":     true,
    "admin_chat_id": true,
}

// RenderOverview formats the overview page as an HTML Telegram message.
func RenderOverview(balance float64, subsystems []SubsystemStatus, openOrders, positions int) string {
    var sb strings.Builder
    sb.WriteString("<b>📊 Overview</b>\n\n")
    sb.WriteString(fmt.Sprintf("💰 Balance: <b>%.2f USDC</b>\n", balance))
    sb.WriteString(fmt.Sprintf("📋 Open orders: <b>%d</b>\n", openOrders))
    sb.WriteString(fmt.Sprintf("💼 Positions: <b>%d</b>\n\n", positions))
    sb.WriteString("<b>Subsystems:</b>\n")
    for _, s := range subsystems {
        dot := "🔴"
        status := "inactive"
        if s.Active {
            dot = "🟢"
            status = "active"
        }
        sb.WriteString(fmt.Sprintf("%s %s — %s\n", dot, s.Name, status))
    }
    return sb.String()
}

// RenderOrders formats the orders list with cancel button data.
// Returns both the text and a list of (label, callbackData) for inline buttons.
func RenderOrders(rows []tui.OrderRow) string {
    if len(rows) == 0 {
        return "📋 <b>Orders</b>\n\nNo open orders."
    }
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("📋 <b>Orders</b> (%d)\n\n", len(rows)))
    for i, r := range rows {
        sb.WriteString(fmt.Sprintf(
            "%d. <b>%s</b> %s @ %s  size: %s  [%s]\n   ID: <code>%s</code>\n",
            i+1, r.Market, r.Side, r.Price, r.Size, r.Status, r.ID,
        ))
    }
    return sb.String()
}

// RenderPositions formats the positions list.
func RenderPositions(rows []tui.PositionRow) string {
    if len(rows) == 0 {
        return "💼 <b>Positions</b>\n\nNo open positions."
    }
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("💼 <b>Positions</b> (%d)\n\n", len(rows)))
    for i, r := range rows {
        sb.WriteString(fmt.Sprintf(
            "%d. <b>%s</b> %s  size: %s  entry: %s  P&L: %s (%s)\n",
            i+1, r.Market, r.Side, r.Size, r.Entry, r.PnL, r.PnLPct,
        ))
    }
    return sb.String()
}

// RenderCopytrading formats the copytrading status.
func RenderCopytrading(traders []tui.TraderRow) string {
    if len(traders) == 0 {
        return "🔄 <b>Copytrading</b>\n\nNo traders configured."
    }
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("🔄 <b>Copytrading</b> (%d traders)\n\n", len(traders)))
    for _, t := range traders {
        sb.WriteString(fmt.Sprintf("• %s (%s)  %s  alloc: %s\n", t.Label, t.Address, t.Status, t.AllocPct))
    }
    return sb.String()
}

// RenderLogs formats the last N log lines.
func RenderLogs(lines []string) string {
    if len(lines) == 0 {
        return "📝 <b>Logs</b>\n\nNo log entries yet."
    }
    last := lines
    if len(last) > 20 {
        last = last[len(last)-20:]
    }
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("📝 <b>Logs</b> (last %d)\n\n<pre>", len(last)))
    for _, line := range last {
        // Escape HTML special chars in log lines
        line = strings.ReplaceAll(line, "&", "&amp;")
        line = strings.ReplaceAll(line, "<", "&lt;")
        line = strings.ReplaceAll(line, ">", "&gt;")
        sb.WriteString(line + "\n")
    }
    sb.WriteString("</pre>")
    return sb.String()
}

// RenderSettings formats a settings section.
// isAdmin controls whether secret field values are shown.
// fields is a map[key]value for the section.
func RenderSettings(section string, fields map[string]string, isAdmin bool) string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("⚙️ <b>Settings — %s</b>\n\n", section))
    for k, v := range fields {
        display := v
        if secretFields[k] && !isAdmin {
            if len(v) > 0 {
                display = strings.Repeat("•", min(len(v), 8))
            } else {
                display = "<i>not set</i>"
            }
        }
        if display == "" {
            display = "<i>not set</i>"
        }
        sb.WriteString(fmt.Sprintf("<code>%-25s</code> %s\n", k, display))
    }
    return sb.String()
}

// RenderError formats an error message.
func RenderError(msg string) string {
    return fmt.Sprintf("❌ <b>Error:</b> %s", msg)
}

// RenderSuccess formats a success message.
func RenderSuccess(msg string) string {
    return fmt.Sprintf("✅ %s", msg)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**Step 4: Run tests**

```bash
go test ./internal/telegrambot/... -run "TestRender" -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/telegrambot/renderer.go internal/telegrambot/renderer_test.go
git commit -m "feat(telegrambot): add HTML renderer for all TUI sections"
```

---

## Task 8: Create `internal/telegrambot/handlers.go`

All command and callback handlers. Contains the settings key map for `/set`.

**Files:**
- Create: `internal/telegrambot/handlers.go`
- Create: `internal/telegrambot/handlers_test.go`

**Step 1: Write the failing tests**

Create `internal/telegrambot/handlers_test.go`:
```go
package telegrambot_test

import (
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/telegrambot"
)

func TestSettingsKey_Get(t *testing.T) {
    cfg := &config.Config{}
    cfg.UI.Language = "ru"
    cfg.Monitor.Enabled = true

    v, ok := telegrambot.GetSetting(cfg, "ui.language")
    if !ok {
        t.Fatal("ui.language key not found")
    }
    if v != "ru" {
        t.Errorf("ui.language = %q, want %q", v, "ru")
    }

    v2, ok2 := telegrambot.GetSetting(cfg, "monitor.enabled")
    if !ok2 {
        t.Fatal("monitor.enabled key not found")
    }
    if v2 != "true" {
        t.Errorf("monitor.enabled = %q, want %q", v2, "true")
    }
}

func TestSettingsKey_Set(t *testing.T) {
    cfg := &config.Config{}
    cfg.API.ClobURL = "https://clob.polymarket.com"

    err := telegrambot.SetSetting(cfg, "ui.language", "zh")
    if err != nil {
        t.Fatalf("SetSetting: %v", err)
    }
    if cfg.UI.Language != "zh" {
        t.Errorf("Language = %q, want zh", cfg.UI.Language)
    }
}

func TestSettingsKey_Secret(t *testing.T) {
    if !telegrambot.IsSecretKey("auth.private_key") {
        t.Error("auth.private_key should be secret")
    }
    if telegrambot.IsSecretKey("monitor.enabled") {
        t.Error("monitor.enabled should not be secret")
    }
}

func TestSettingsKey_UnknownKey(t *testing.T) {
    cfg := &config.Config{}
    _, ok := telegrambot.GetSetting(cfg, "nonexistent.field")
    if ok {
        t.Error("unknown key should return ok=false")
    }
    err := telegrambot.SetSetting(cfg, "nonexistent.field", "value")
    if err == nil {
        t.Error("SetSetting on unknown key should return error")
    }
}
```

**Step 2: Run to verify fails**

```bash
go test ./internal/telegrambot/... -run "TestSettingsKey" -v
```
Expected: FAIL

**Step 3: Create `internal/telegrambot/handlers.go`**

```go
package telegrambot

import (
    "context"
    "fmt"
    "strconv"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/i18n"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

// settingEntry describes one editable config field accessible via /set.
type settingEntry struct {
    get    func(*config.Config) string
    set    func(*config.Config, string) error
    secret bool
    onSet  func(v string) // side effects (e.g. i18n change)
}

func boolStr(b bool) string {
    if b {
        return "true"
    }
    return "false"
}

func parseBool(s string) bool {
    s = strings.ToLower(strings.TrimSpace(s))
    return s == "true" || s == "1" || s == "yes"
}

// settingsMap maps dot-notation keys to config accessors.
var settingsMap = map[string]settingEntry{
    // UI
    "ui.language": {
        get: func(c *config.Config) string { return c.UI.Language },
        set: func(c *config.Config, v string) error { c.UI.Language = v; return nil },
        onSet: func(v string) { i18n.SetLanguage(v) },
    },
    // Monitor
    "monitor.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Monitor.Enabled) },
        set: func(c *config.Config, v string) error { c.Monitor.Enabled = parseBool(v); return nil },
    },
    "monitor.poll_interval_ms": {
        get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.PollIntervalMs) },
        set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil {
                return err
            }
            c.Monitor.PollIntervalMs = n
            return nil
        },
    },
    // Trades Monitor
    "monitor.trades.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.Enabled) },
        set: func(c *config.Config, v string) error { c.Monitor.Trades.Enabled = parseBool(v); return nil },
    },
    "monitor.trades.poll_interval_ms": {
        get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.PollIntervalMs) },
        set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil {
                return err
            }
            c.Monitor.Trades.PollIntervalMs = n
            return nil
        },
    },
    "monitor.trades.alert_on_fill": {
        get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnFill) },
        set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnFill = parseBool(v); return nil },
    },
    "monitor.trades.alert_on_cancel": {
        get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnCancel) },
        set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnCancel = parseBool(v); return nil },
    },
    // Trading
    "trading.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Trading.Enabled) },
        set: func(c *config.Config, v string) error { c.Trading.Enabled = parseBool(v); return nil },
    },
    "trading.max_position_usd": {
        get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.MaxPositionUSD) },
        set: func(c *config.Config, v string) error {
            f, err := strconv.ParseFloat(v, 64)
            if err != nil {
                return err
            }
            c.Trading.MaxPositionUSD = f
            return nil
        },
    },
    "trading.slippage_pct": {
        get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.SlippagePct) },
        set: func(c *config.Config, v string) error {
            f, err := strconv.ParseFloat(v, 64)
            if err != nil {
                return err
            }
            c.Trading.SlippagePct = f
            return nil
        },
    },
    "trading.neg_risk": {
        get: func(c *config.Config) string { return boolStr(c.Trading.NegRisk) },
        set: func(c *config.Config, v string) error { c.Trading.NegRisk = parseBool(v); return nil },
    },
    // Copytrading
    "copytrading.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Copytrading.Enabled) },
        set: func(c *config.Config, v string) error { c.Copytrading.Enabled = parseBool(v); return nil },
    },
    "copytrading.poll_interval_ms": {
        get: func(c *config.Config) string { return strconv.Itoa(c.Copytrading.PollIntervalMs) },
        set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil {
                return err
            }
            c.Copytrading.PollIntervalMs = n
            return nil
        },
    },
    "copytrading.size_mode": {
        get: func(c *config.Config) string { return c.Copytrading.SizeMode },
        set: func(c *config.Config, v string) error { c.Copytrading.SizeMode = v; return nil },
    },
    // Telegram (non-secret)
    "telegram.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Telegram.Enabled) },
        set: func(c *config.Config, v string) error { c.Telegram.Enabled = parseBool(v); return nil },
    },
    // Database
    "database.enabled": {
        get: func(c *config.Config) string { return boolStr(c.Database.Enabled) },
        set: func(c *config.Config, v string) error { c.Database.Enabled = parseBool(v); return nil },
    },
    "database.path": {
        get: func(c *config.Config) string { return c.Database.Path },
        set: func(c *config.Config, v string) error { c.Database.Path = v; return nil },
    },
    // Log
    "log.level": {
        get: func(c *config.Config) string { return c.Log.Level },
        set: func(c *config.Config, v string) error { c.Log.Level = v; return nil },
    },
    "log.format": {
        get: func(c *config.Config) string { return c.Log.Format },
        set: func(c *config.Config, v string) error { c.Log.Format = v; return nil },
    },
    // Auth (admin-only)
    "auth.private_key": {
        secret: true,
        get:    func(c *config.Config) string { return c.Auth.PrivateKey },
        set:    func(c *config.Config, v string) error { c.Auth.PrivateKey = v; return nil },
    },
    "auth.api_key": {
        secret: true,
        get:    func(c *config.Config) string { return c.Auth.APIKey },
        set:    func(c *config.Config, v string) error { c.Auth.APIKey = v; return nil },
    },
    "auth.api_secret": {
        secret: true,
        get:    func(c *config.Config) string { return c.Auth.APISecret },
        set:    func(c *config.Config, v string) error { c.Auth.APISecret = v; return nil },
    },
    "auth.passphrase": {
        secret: true,
        get:    func(c *config.Config) string { return c.Auth.Passphrase },
        set:    func(c *config.Config, v string) error { c.Auth.Passphrase = v; return nil },
    },
    "auth.chain_id": {
        secret: true,
        get: func(c *config.Config) string { return strconv.FormatInt(c.Auth.ChainID, 10) },
        set: func(c *config.Config, v string) error {
            n, err := strconv.ParseInt(v, 10, 64)
            if err != nil {
                return err
            }
            c.Auth.ChainID = n
            return nil
        },
    },
    "telegram.bot_token": {
        secret: true,
        get:    func(c *config.Config) string { return c.Telegram.BotToken },
        set:    func(c *config.Config, v string) error { c.Telegram.BotToken = v; return nil },
    },
    "telegram.chat_id": {
        secret: true,
        get:    func(c *config.Config) string { return c.Telegram.ChatID },
        set:    func(c *config.Config, v string) error { c.Telegram.ChatID = v; return nil },
    },
    "telegram.admin_chat_id": {
        secret: true,
        get:    func(c *config.Config) string { return c.Telegram.AdminChatID },
        set:    func(c *config.Config, v string) error { c.Telegram.AdminChatID = v; return nil },
    },
}

// GetSetting returns the current value for a dot-notation key.
func GetSetting(cfg *config.Config, key string) (string, bool) {
    e, ok := settingsMap[key]
    if !ok {
        return "", false
    }
    return e.get(cfg), true
}

// SetSetting applies a value for a dot-notation key.
func SetSetting(cfg *config.Config, key, value string) error {
    e, ok := settingsMap[key]
    if !ok {
        return fmt.Errorf("unknown setting key: %q", key)
    }
    return e.set(cfg, value)
}

// IsSecretKey reports whether the key is admin-only.
func IsSecretKey(key string) bool {
    e, ok := settingsMap[key]
    return ok && e.secret
}

// handlersImpl holds the handler methods (attached to Bot in bot.go).
// Command dispatch is in Bot.handleUpdate().

// mainMenuKeyboard builds the main navigation keyboard.
func mainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📊 Orders", "cmd:orders"),
            tgbotapi.NewInlineKeyboardButtonData("💼 Positions", "cmd:positions"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("ℹ️ Overview", "cmd:overview"),
            tgbotapi.NewInlineKeyboardButtonData("🔄 Copytrading", "cmd:copytrading"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📝 Logs", "cmd:logs"),
            tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "cmd:settings"),
        ),
    )
}

// ordersKeyboard builds per-order cancel buttons + Cancel All.
func ordersKeyboard(orders []tui.OrderRow) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    for i, o := range orders {
        label := fmt.Sprintf("❌ Cancel #%d (%s)", i+1, o.Side)
        rows = append(rows, tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData(label, "cancel:"+o.ID),
        ))
    }
    rows = append(rows, tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("❌ Cancel ALL", "cancelall:confirm"),
        tgbotapi.NewInlineKeyboardButtonData("← Back", "cmd:menu"),
    ))
    return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// cancelAllConfirmKeyboard asks for confirmation before canceling all.
func cancelAllConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("✅ Yes, cancel all", "cancelall:do"),
            tgbotapi.NewInlineKeyboardButtonData("🚫 No, go back", "cmd:orders"),
        ),
    )
}

// backKeyboard is a simple [← Back to Menu] row.
func backKeyboard() tgbotapi.InlineKeyboardMarkup {
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("← Back to Menu", "cmd:menu"),
        ),
    )
}

// handleCommand dispatches slash commands. Called from Bot.handleUpdate.
func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
    switch msg.Command() {
    case "start", "menu":
        b.sendWithKeyboard(msg.Chat.ID, "polytrade-bot\n\nChoose a section:", mainMenuKeyboard())
    case "status", "overview":
        b.sendOverview(ctx, msg.Chat.ID)
    case "orders":
        b.sendOrders(ctx, msg.Chat.ID)
    case "cancel":
        id := strings.TrimSpace(msg.CommandArguments())
        if id == "" {
            b.sendText(msg.Chat.ID, RenderError("Usage: /cancel <order_id>"))
            return
        }
        b.doCancelOrder(ctx, msg.Chat.ID, id)
    case "cancelall":
        b.sendWithKeyboard(msg.Chat.ID, "⚠️ Cancel ALL open orders?", cancelAllConfirmKeyboard())
    case "positions":
        b.sendPositions(msg.Chat.ID)
    case "copy":
        b.sendCopytrading(msg.Chat.ID)
    case "logs":
        b.sendLogs(msg.Chat.ID)
    case "settings":
        b.sendSettings(msg.Chat.ID, b.isAdmin(msg.Chat.ID))
    case "set":
        args := strings.Fields(msg.CommandArguments())
        if len(args) < 2 {
            b.sendText(msg.Chat.ID, RenderError("Usage: /set <key> <value>"))
            return
        }
        b.doSetSetting(ctx, msg.Chat.ID, args[0], strings.Join(args[1:], " "))
    default:
        b.sendText(msg.Chat.ID, "Unknown command. Use /start for the menu.")
    }
}

// handleCallback dispatches inline keyboard callbacks.
func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
    // Always answer the callback to clear the "loading" spinner
    answer := tgbotapi.NewCallback(cb.ID, "")
    b.api.Send(answer) //nolint:errcheck

    chatID := cb.Message.Chat.ID
    data := cb.Data

    switch {
    case data == "cmd:menu":
        b.sendWithKeyboard(chatID, "polytrade-bot\n\nChoose a section:", mainMenuKeyboard())
    case data == "cmd:overview":
        b.sendOverview(ctx, chatID)
    case data == "cmd:orders":
        b.sendOrders(ctx, chatID)
    case data == "cmd:positions":
        b.sendPositions(chatID)
    case data == "cmd:copytrading":
        b.sendCopytrading(chatID)
    case data == "cmd:logs":
        b.sendLogs(chatID)
    case data == "cmd:settings":
        b.sendSettings(chatID, b.isAdmin(chatID))
    case data == "cancelall:confirm":
        b.sendWithKeyboard(chatID, "⚠️ Are you sure you want to cancel ALL orders?", cancelAllConfirmKeyboard())
    case data == "cancelall:do":
        b.doCancelAll(ctx, chatID)
    case strings.HasPrefix(data, "cancel:"):
        orderID := strings.TrimPrefix(data, "cancel:")
        b.doCancelOrder(ctx, chatID, orderID)
    }
}

// --- Action helpers ---

func (b *Bot) sendOverview(_ context.Context, chatID int64) {
    subsystems := b.state.Subsystems()
    orders := b.state.Orders()
    positions := b.state.Positions()
    text := RenderOverview(b.state.Balance(), subsystems, len(orders), len(positions))
    b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) sendOrders(_ context.Context, chatID int64) {
    orders := b.state.Orders()
    text := RenderOrders(orders)
    b.sendWithKeyboard(chatID, text, ordersKeyboard(orders))
}

func (b *Bot) sendPositions(chatID int64) {
    text := RenderPositions(b.state.Positions())
    b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) sendCopytrading(chatID int64) {
    text := RenderCopytrading(b.state.Traders())
    b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) sendLogs(chatID int64) {
    text := RenderLogs(b.state.Logs())
    b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) sendSettings(chatID int64, isAdmin bool) {
    b.cfgMu.RLock()
    cfg := b.cfg
    b.cfgMu.RUnlock()

    // Build one unified settings view grouped by section
    sections := []struct {
        name string
        keys []string
    }{
        {"UI", []string{"ui.language"}},
        {"Auth", []string{"auth.private_key", "auth.api_key", "auth.api_secret", "auth.passphrase", "auth.chain_id"}},
        {"Monitor", []string{"monitor.enabled", "monitor.poll_interval_ms"}},
        {"Trades Monitor", []string{"monitor.trades.enabled", "monitor.trades.poll_interval_ms", "monitor.trades.alert_on_fill", "monitor.trades.alert_on_cancel"}},
        {"Trading", []string{"trading.enabled", "trading.max_position_usd", "trading.slippage_pct", "trading.neg_risk"}},
        {"Copytrading", []string{"copytrading.enabled", "copytrading.poll_interval_ms", "copytrading.size_mode"}},
        {"Telegram", []string{"telegram.enabled", "telegram.bot_token", "telegram.chat_id", "telegram.admin_chat_id"}},
        {"Database", []string{"database.enabled", "database.path"}},
        {"Log", []string{"log.level", "log.format"}},
    }

    var parts []string
    for _, sec := range sections {
        fields := make(map[string]string, len(sec.keys))
        for _, k := range sec.keys {
            shortKey := k[strings.Index(k, ".")+1:]
            if strings.Contains(k, ".") {
                parts2 := strings.SplitN(k, ".", 2)
                shortKey = parts2[1]
            }
            if v, ok := GetSetting(cfg, k); ok {
                fields[shortKey] = v
            }
        }
        parts = append(parts, RenderSettings(sec.name, fields, isAdmin))
    }

    text := strings.Join(parts, "\n")
    if !isAdmin {
        text += "\n<i>Use /set &lt;key&gt; &lt;value&gt; to change settings.\nAdmin: secret fields require admin_chat_id.</i>"
    } else {
        text += "\n<i>Admin mode: all fields editable.\nUse /set &lt;key&gt; &lt;value&gt;</i>"
    }
    b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) doCancelOrder(ctx context.Context, chatID int64, orderID string) {
    if b.canceler == nil {
        b.sendText(chatID, RenderError("Order cancellation unavailable (TradesMonitor not enabled)"))
        return
    }
    if err := b.canceler.CancelOrder(orderID); err != nil {
        b.sendText(chatID, RenderError(err.Error()))
        return
    }
    b.sendText(chatID, RenderSuccess(fmt.Sprintf("Order <code>%s</code> canceled.", orderID)))
}

func (b *Bot) doCancelAll(ctx context.Context, chatID int64) {
    if b.canceler == nil {
        b.sendText(chatID, RenderError("Order cancellation unavailable (TradesMonitor not enabled)"))
        return
    }
    if err := b.canceler.CancelAllOrders(); err != nil {
        b.sendText(chatID, RenderError(err.Error()))
        return
    }
    b.sendText(chatID, RenderSuccess("All orders canceled."))
}

func (b *Bot) doSetSetting(ctx context.Context, chatID int64, key, value string) {
    isAdmin := b.isAdmin(chatID)
    if IsSecretKey(key) && !isAdmin {
        b.sendText(chatID, RenderError(fmt.Sprintf("Key %q requires admin access.", key)))
        return
    }

    b.cfgMu.Lock()
    cfgCopy := *b.cfg
    if err := SetSetting(&cfgCopy, key, value); err != nil {
        b.cfgMu.Unlock()
        b.sendText(chatID, RenderError(fmt.Sprintf("Invalid value for %q: %v", key, err)))
        return
    }

    // Apply side effects (e.g. i18n language change)
    if e, ok := settingsMap[key]; ok && e.onSet != nil {
        e.onSet(value)
    }

    if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
        b.cfgMu.Unlock()
        b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
        return
    }
    *b.cfg = cfgCopy
    b.cfgMu.Unlock()

    // Notify TUI of config change
    b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})

    b.sendText(chatID, RenderSuccess(fmt.Sprintf("<code>%s</code> = <code>%s</code>\nConfig saved. TUI updated.", key, value)))
}
```

**Step 4: Run tests**

```bash
go test ./internal/telegrambot/... -run "TestSettingsKey" -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/telegrambot/handlers.go internal/telegrambot/handlers_test.go
git commit -m "feat(telegrambot): add command handlers, settings key map, keyboards"
```

---

## Task 9: Create `internal/telegrambot/bot.go`

The main `Bot` struct, `New()`, `Run()`, and helper send methods.

**Files:**
- Create: `internal/telegrambot/bot.go`

**Step 1: Write the failing test**

Create `internal/telegrambot/bot_test.go`:
```go
package telegrambot_test

import (
    "testing"

    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/telegrambot"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestBot_AccessControl(t *testing.T) {
    cfg := &config.Config{}
    cfg.API.ClobURL = "https://clob.polymarket.com"
    cfg.Telegram.ChatID = "111"
    cfg.Telegram.AdminChatID = "999"
    bus := tui.NewEventBus()

    b, err := telegrambot.New(cfg, "config.toml", bus, nil, nil)
    if err != nil {
        t.Fatalf("New: %v", err)
    }

    if !b.IsAllowed(111) {
        t.Error("chat 111 should be allowed (chat_id)")
    }
    if b.IsAllowed(222) {
        t.Error("chat 222 should NOT be allowed")
    }
    if !b.IsAllowed(999) {
        t.Error("admin 999 should be allowed")
    }
    if !b.IsAdmin(999) {
        t.Error("chat 999 should be admin")
    }
    if b.IsAdmin(111) {
        t.Error("chat 111 should NOT be admin")
    }
}
```

**Step 2: Run to verify fails**

```bash
go test ./internal/telegrambot/... -run "TestBot_AccessControl" -v
```
Expected: FAIL

**Step 3: Create `internal/telegrambot/bot.go`**

```go
// Package telegrambot implements an interactive Telegram Bot that mirrors
// the Console TUI, synchronized via the shared EventBus.
package telegrambot

import (
    "context"
    "strconv"
    "sync"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/rs/zerolog"

    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

// OrderCanceler is the subset of TradesMonitor used by the bot.
type OrderCanceler interface {
    CancelOrder(id string) error
    CancelAllOrders() error
}

// Bot is the interactive Telegram Bot.
type Bot struct {
    api      *tgbotapi.BotAPI
    bus      *tui.EventBus
    state    *BotState
    canceler OrderCanceler // optional; nil if TradesMonitor not running
    log      zerolog.Logger

    cfgMu   sync.RWMutex
    cfg     *config.Config
    cfgPath string

    allowedIDs map[int64]bool
    adminID    int64
}

// New creates a new Bot. canceler may be nil if order management is not needed.
// Returns an error only if the Telegram API token is invalid.
// When cfg.Telegram.BotToken is empty, returns (nil, nil) — caller must check.
func New(cfg *config.Config, cfgPath string, bus *tui.EventBus, canceler OrderCanceler, log *zerolog.Logger) (*Bot, error) {
    allowed := make(map[int64]bool)
    if cfg.Telegram.ChatID != "" {
        if id, err := strconv.ParseInt(cfg.Telegram.ChatID, 10, 64); err == nil {
            allowed[id] = true
        }
    }

    var adminID int64
    if cfg.Telegram.AdminChatID != "" {
        if id, err := strconv.ParseInt(cfg.Telegram.AdminChatID, 10, 64); err == nil {
            allowed[id] = true // admin is also allowed
            adminID = id
        }
    }

    l := zerolog.Nop()
    if log != nil {
        l = log.With().Str("component", "telegram-bot").Logger()
    }

    b := &Bot{
        bus:        bus,
        state:      NewBotState(),
        canceler:   canceler,
        log:        l,
        cfg:        cfg,
        cfgPath:    cfgPath,
        allowedIDs: allowed,
        adminID:    adminID,
    }

    // Only create the API client if token is provided
    if cfg.Telegram.BotToken != "" {
        api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
        if err != nil {
            return nil, err
        }
        b.api = api
    }

    return b, nil
}

// IsAllowed reports whether chatID is authorized to use the bot.
func (b *Bot) IsAllowed(chatID int64) bool {
    return b.allowedIDs[chatID]
}

// IsAdmin reports whether chatID has admin privileges.
func (b *Bot) isAdmin(chatID int64) bool {
    return b.adminID != 0 && chatID == b.adminID
}

// IsAdmin is exported for testing.
func (b *Bot) IsAdmin(chatID int64) bool {
    return b.isAdmin(chatID)
}

// Run starts the bot. Blocks until ctx is cancelled.
// If api is nil (no token configured), returns immediately.
func (b *Bot) Run(ctx context.Context) error {
    if b.api == nil {
        b.log.Warn().Msg("Telegram bot token not set, bot disabled")
        <-ctx.Done()
        return nil
    }

    b.log.Info().Str("username", b.api.Self.UserName).Msg("Telegram bot started")

    // EventBus tap: receive copies of all bus events
    tap := b.bus.Tap()

    // Start EventBus consumer goroutine
    go b.consumeEvents(ctx, tap)

    // Start Telegram polling
    return b.pollTelegram(ctx)
}

// consumeEvents reads from the EventBus tap and updates BotState.
func (b *Bot) consumeEvents(ctx context.Context, tap <-chan tgbotapi.Msg) {
    // Note: tap is <-chan tea.Msg (tgbotapi.Msg is the wrong type here)
    // The actual type is <-chan tea.Msg from tui package.
    // See below for the correct signature.
}

// pollTelegram runs the long-polling getUpdates loop.
func (b *Bot) pollTelegram(ctx context.Context) error {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 30
    updates := b.api.GetUpdatesChan(u)

    for {
        select {
        case <-ctx.Done():
            b.api.StopReceivingUpdates()
            b.log.Info().Msg("Telegram bot stopped")
            return nil
        case update, ok := <-updates:
            if !ok {
                return nil
            }
            b.handleUpdate(ctx, update)
        }
    }
}

// handleUpdate routes an incoming Telegram update.
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
    switch {
    case update.Message != nil:
        if !b.IsAllowed(update.Message.Chat.ID) {
            b.log.Debug().Int64("chat_id", update.Message.Chat.ID).Msg("ignoring message from unauthorized chat")
            return
        }
        if update.Message.IsCommand() {
            b.handleCommand(ctx, update.Message)
        }
    case update.CallbackQuery != nil:
        if !b.IsAllowed(update.CallbackQuery.Message.Chat.ID) {
            return
        }
        b.handleCallback(ctx, update.CallbackQuery)
    }
}

// sendText sends a plain HTML text message.
func (b *Bot) sendText(chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = tgbotapi.ModeHTML
    if _, err := b.api.Send(msg); err != nil {
        b.log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to send message")
    }
}

// sendWithKeyboard sends an HTML text message with an inline keyboard.
func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = tgbotapi.ModeHTML
    msg.ReplyMarkup = keyboard
    if _, err := b.api.Send(msg); err != nil {
        b.log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to send keyboard message")
    }
}
```

Then add the real `consumeEvents` implementation — note `tap` is `<-chan tea.Msg`:

```go
// In bot.go, replace the placeholder consumeEvents with:

import tea "github.com/charmbracelet/bubbletea"

func (b *Bot) consumeEvents(ctx context.Context, tap <-chan tea.Msg) {
    for {
        select {
        case <-ctx.Done():
            return
        case msg, ok := <-tap:
            if !ok {
                return
            }
            b.processBusMsg(msg)
        }
    }
}

func (b *Bot) processBusMsg(msg tea.Msg) {
    switch m := msg.(type) {
    case tui.BalanceMsg:
        b.state.SetBalance(m.USDC)

    case tui.SubsystemStatusMsg:
        b.state.SetSubsystem(m.Name, m.Active)

    case tui.BotEventMsg:
        b.state.AddLog(m.Level + " " + m.Message)

    case tui.OrdersUpdateMsg:
        b.state.SetOrders(m.Rows)

    case tui.PositionsUpdateMsg:
        b.state.SetPositions(m.Rows)

    case tui.ConfigReloadedMsg:
        b.cfgMu.Lock()
        b.cfg = m.Config
        b.cfgMu.Unlock()
    }
}
```

**Step 4: Run tests and build**

```bash
go test ./internal/telegrambot/... -run "TestBot_AccessControl" -v
go build ./...
```
Expected: PASS, clean build

**Step 5: Commit**

```bash
git add internal/telegrambot/bot.go internal/telegrambot/bot_test.go
git commit -m "feat(telegrambot): add Bot struct, Run(), polling loop, EventBus consumer"
```

---

## Task 10: Wire Telegram Bot in `cmd/bot/main.go`

**Files:**
- Modify: `cmd/bot/main.go`

**Step 1: Import the new package and add initialization**

In `main.go`, add import:
```go
"github.com/atlasdev/polytrade-bot/internal/telegrambot"
```

After the Notifier block (line ~116), add:
```go
// --- Telegram Bot (interactive) ---
var tgBot *telegrambot.Bot
if cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" {
    var cancelerForBot telegrambot.OrderCanceler
    if cfg.Monitor.Trades.Enabled && l2Creds != nil {
        cancelerForBot = tradesMon
    }
    var err error
    tgBot, err = telegrambot.New(cfg, *cfgPath, bus, cancelerForBot, &log)
    if err != nil {
        log.Warn().Err(err).Msg("telegram bot init failed, continuing without it")
        tgBot = nil
    }
}
```

Then in the subsystem start block, after TradesMonitor start:
```go
if tgBot != nil {
    startSubsystem("Telegram Bot", func() error { return tgBot.Run(ctx) })
}
```

Also wire `TradesMonitor` to the EventBus:
```go
if bus != nil {
    tradesMon.SetBus(bus)
}
```
(Add this line after `tradesMon` is created, before subsystems are started.)

**Step 2: Build and verify**

```bash
go build ./...
go vet ./...
```
Expected: clean

**Step 3: Commit**

```bash
git add cmd/bot/main.go
git commit -m "feat(main): wire Telegram Bot subsystem and TradesMonitor EventBus"
```

---

## Task 11: Update `tab_settings.go` — remove duplicate `saveConfig`, use `config.Save`

The private `saveConfig()` in `tab_settings.go` duplicates the new `config.Save()`. Replace it.

**Files:**
- Modify: `internal/tui/tab_settings.go`

**Step 1: Replace `saveConfig` at the bottom of `tab_settings.go`**

Remove the old `saveConfig` function:
```go
// DELETE this:
func saveConfig(path string, cfg *config.Config) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    return toml.NewEncoder(f).Encode(cfg)
}
```

Replace the two import usages: find all calls to `saveConfig(m.cfgPath, &cfgCopy)` in the file and replace with `config.Save(m.cfgPath, &cfgCopy)`.

Remove the `"os"` and `"github.com/BurntSushi/toml"` imports if they're now unused (check first — `toml` might still be used elsewhere).

**Step 2: Build**

```bash
go build ./...
```
Expected: clean (if `os` or `toml` are still needed for other things, keep them)

**Step 3: Run all tests**

```bash
go test ./...
```
Expected: all PASS

**Step 4: Commit**

```bash
git add internal/tui/tab_settings.go
git commit -m "refactor(tui): replace saveConfig() with config.Save()"
```

---

## Task 12: Manual Integration Verification

**Prerequisite:** Valid `config.toml` with `[telegram]` section filled.

**Step 1: Build**

```bash
go build -o polytrade-bot ./cmd/bot/
```

**Step 2: Set up Telegram test bot**

1. Message @BotFather → `/newbot` → get token
2. Get your chat ID: message @userinfobot
3. Update `config.toml`:
   ```toml
   [telegram]
   enabled = true
   bot_token = "YOUR_TOKEN"
   chat_id = "YOUR_CHAT_ID"
   admin_chat_id = "YOUR_CHAT_ID"  # for testing, set same as chat_id
   ```

**Step 3: Run and verify commands**

```bash
./polytrade-bot --config config.toml
```

In Telegram chat:
- `/start` → main menu with 6 inline buttons ✓
- Click "ℹ️ Overview" → balance, subsystem status ✓
- Click "📊 Orders" → order list (or "No open orders") ✓
- `/settings` → settings view with masked secrets ✓
- `/set ui.language ru` → language changes, config saved ✓
- Check TUI: language should update without restart ✓
- Change setting in TUI (save with S) → check bot `/settings` reflects new value ✓

**Step 4: Verify access control**

Try messaging from a different account → no response (silently ignored) ✓

**Step 5: Final commit tag**

```bash
git tag -a v0.x.0-telegram -m "Telegram Bot interactive interface"
```

---

## Summary of Files

| Action | File |
|--------|------|
| Create | `internal/telegrambot/state.go` |
| Create | `internal/telegrambot/state_test.go` |
| Create | `internal/telegrambot/renderer.go` |
| Create | `internal/telegrambot/renderer_test.go` |
| Create | `internal/telegrambot/handlers.go` |
| Create | `internal/telegrambot/handlers_test.go` |
| Create | `internal/telegrambot/bot.go` |
| Create | `internal/telegrambot/bot_test.go` |
| Create | `internal/config/save_test.go` |
| Create | `internal/tui/messages_test.go` |
| Create | `internal/monitor/trades_bus_test.go` |
| Modify | `internal/config/config.go` — add `AdminChatID`, `Save()` |
| Modify | `internal/tui/messages.go` — fan-out EventBus, new msg types |
| Modify | `internal/tui/tab_settings.go` — use `config.Save()` |
| Modify | `internal/monitor/trades.go` — `SetBus()`, emit update msgs |
| Modify | `cmd/bot/main.go` — wire tgBot subsystem |
| Modify | `go.mod`, `go.sum` — add go-telegram-bot-api/v5 |
