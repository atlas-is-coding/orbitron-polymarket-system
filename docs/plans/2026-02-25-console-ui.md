# Console UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Bubble Tea TUI dashboard to polytrade-bot with tabbed layout, first-run wizard, full settings editor with tooltips, and fsnotify hot reload.

**Architecture:** Sub-model pattern — root `AppModel` delegates to per-tab sub-models, each self-contained. A shared `EventBus` (`chan tea.Msg`) bridges bot subsystems to TUI without mutexes. `ConfigWatcher` wraps fsnotify with 300ms debounce and broadcasts `ConfigReloadedMsg` through the TUI loop.

**Tech Stack:** `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles` (textinput, table, viewport), `github.com/fsnotify/fsnotify` (already in go.mod)

---

## Task 1: Add TUI dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

**Step 1: Add dependencies**

```bash
cd /path/to/polytrade-bot
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go mod tidy
```

**Step 2: Verify build still works**

```bash
go build ./...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add bubbletea, lipgloss, bubbles for TUI"
```

---

## Task 2: ConfigWatcher

**Files:**
- Create: `internal/config/watcher.go`
- Create: `internal/config/watcher_test.go`

**Step 1: Write failing test**

```go
// internal/config/watcher_test.go
package config_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/atlasdev/polytrade-bot/internal/config"
)

func TestConfigWatcher_NotifiesOnChange(t *testing.T) {
    // create a temp config file
    f, err := os.CreateTemp(t.TempDir(), "config-*.toml")
    if err != nil {
        t.Fatal(err)
    }
    f.WriteString(`[api]
clob_url = "https://clob.polymarket.com"
`)
    f.Close()

    reloaded := make(chan struct{}, 1)
    w, err := config.NewWatcher(f.Name(), func(cfg *config.Config) {
        reloaded <- struct{}{}
    })
    if err != nil {
        t.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    go w.Run(ctx)

    // modify file to trigger reload
    time.Sleep(100 * time.Millisecond)
    os.WriteFile(f.Name(), []byte(`[api]
clob_url = "https://clob.polymarket.com"
timeout_sec = 15
`), 0644)

    select {
    case <-reloaded:
        // pass
    case <-ctx.Done():
        t.Fatal("timeout: watcher did not fire")
    }
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/config/... -run TestConfigWatcher -v
```
Expected: FAIL — `config.NewWatcher` undefined.

**Step 3: Implement ConfigWatcher**

```go
// internal/config/watcher.go
package config

import (
    "context"
    "time"

    "github.com/fsnotify/fsnotify"
)

// Watcher watches a config file and calls OnReload when it changes.
type Watcher struct {
    path     string
    onReload func(*Config)
    debounce time.Duration
}

// NewWatcher creates a Watcher for the given config file path.
func NewWatcher(path string, onReload func(*Config)) (*Watcher, error) {
    return &Watcher{
        path:     path,
        onReload: onReload,
        debounce: 300 * time.Millisecond,
    }, nil
}

// Run starts watching the config file. Blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
    fw, err := fsnotify.NewWatcher()
    if err != nil {
        return
    }
    defer fw.Close()

    if err := fw.Add(w.path); err != nil {
        return
    }

    var timer *time.Timer
    for {
        select {
        case <-ctx.Done():
            return
        case event, ok := <-fw.Events:
            if !ok {
                return
            }
            if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
                // debounce: reset timer on each event
                if timer != nil {
                    timer.Stop()
                }
                timer = time.AfterFunc(w.debounce, func() {
                    cfg, err := Load(w.path)
                    if err != nil {
                        return
                    }
                    w.onReload(cfg)
                })
            }
        case _, ok := <-fw.Errors:
            if !ok {
                return
            }
        }
    }
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/config/... -run TestConfigWatcher -v
```
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/config/watcher.go internal/config/watcher_test.go
git commit -m "feat: add ConfigWatcher with fsnotify debounce hot reload"
```

---

## Task 3: TUI messages and EventBus

**Files:**
- Create: `internal/tui/messages.go`

**Step 1: Write messages.go**

```go
// internal/tui/messages.go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/atlasdev/polytrade-bot/internal/config"
)

// ConfigReloadedMsg is sent when config.toml changes on disk.
type ConfigReloadedMsg struct {
    Config *config.Config
}

// BotEventMsg carries a log line or status update from a subsystem.
type BotEventMsg struct {
    Level   string // "trace","debug","info","warn","error"
    Message string
    Time    string
}

// SubsystemStatusMsg updates the running/stopped state of a subsystem.
type SubsystemStatusMsg struct {
    Name   string
    Active bool
}

// BalanceMsg carries the current USDC balance.
type BalanceMsg struct {
    USDC float64
}

// EventBus bridges bot goroutines to the Bubble Tea loop.
type EventBus struct {
    ch chan tea.Msg
}

// NewEventBus creates an EventBus with a buffered channel.
func NewEventBus() *EventBus {
    return &EventBus{ch: make(chan tea.Msg, 256)}
}

// Send enqueues a message (non-blocking; drops if full).
func (b *EventBus) Send(msg tea.Msg) {
    select {
    case b.ch <- msg:
    default:
    }
}

// WaitForEvent returns a tea.Cmd that blocks until the next EventBus message.
func (b *EventBus) WaitForEvent() tea.Cmd {
    return func() tea.Msg {
        return <-b.ch
    }
}
```

**Step 2: Build check**

```bash
go build ./internal/tui/...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add internal/tui/messages.go
git commit -m "feat: add TUI messages and EventBus"
```

---

## Task 4: Styles and keybindings

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/keys.go`

**Step 1: Create styles.go**

```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    ColorPrimary  = lipgloss.Color("#7C3AED") // purple
    ColorSuccess  = lipgloss.Color("#10B981") // green
    ColorWarning  = lipgloss.Color("#F59E0B") // yellow
    ColorError    = lipgloss.Color("#EF4444") // red
    ColorMuted    = lipgloss.Color("#6B7280") // gray
    ColorBg       = lipgloss.Color("#111827")
    ColorFg       = lipgloss.Color("#F9FAFB")
    ColorBorder   = lipgloss.Color("#374151")
    ColorSelected = lipgloss.Color("#1D4ED8")

    StyleHeader = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorFg).
            Background(ColorPrimary).
            Padding(0, 2)

    StyleTabActive = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorFg).
            Background(ColorSelected).
            Padding(0, 2)

    StyleTabInactive = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Padding(0, 2)

    StyleHelpBar = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Padding(0, 1)

    StyleBorder = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder)

    StyleTooltip = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Padding(0, 1)

    StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
    StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
    StyleError   = lipgloss.NewStyle().Foreground(ColorError)
    StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
    StyleBold    = lipgloss.NewStyle().Bold(true)
)
```

**Step 2: Create keys.go**

```go
// internal/tui/keys.go
package tui

import "github.com/charmbracelet/bubbles/key"

type GlobalKeyMap struct {
    NextTab key.Binding
    PrevTab key.Binding
    Quit    key.Binding
}

var GlobalKeys = GlobalKeyMap{
    NextTab: key.NewBinding(
        key.WithKeys("tab"),
        key.WithHelp("tab", "next tab"),
    ),
    PrevTab: key.NewBinding(
        key.WithKeys("shift+tab"),
        key.WithHelp("shift+tab", "prev tab"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("ctrl+c", "q"),
        key.WithHelp("q", "quit"),
    ),
}
```

**Step 3: Build check**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/styles.go internal/tui/keys.go
git commit -m "feat: add TUI styles and keybindings"
```

---

## Task 5: Logs tab (zerolog integration)

**Files:**
- Create: `internal/tui/tabs/logs.go`

The logs tab acts as a zerolog writer, capturing all bot log output into a ring buffer and rendering it as a scrollable viewport.

**Step 1: Create logs.go**

```go
// internal/tui/tabs/logs.go
package tabs

import (
    "fmt"
    "strings"
    "sync"

    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

const maxLogLines = 500

// LogWriter implements io.Writer for zerolog; feeds lines to LogsModel.
type LogWriter struct {
    mu   sync.Mutex
    bus  *tui.EventBus
}

func NewLogWriter(bus *tui.EventBus) *LogWriter {
    return &LogWriter{bus: bus}
}

func (w *LogWriter) Write(p []byte) (int, error) {
    line := strings.TrimRight(string(p), "\n")
    level := detectLevel(line)
    w.bus.Send(tui.BotEventMsg{
        Level:   level,
        Message: line,
    })
    return len(p), nil
}

func detectLevel(line string) string {
    for _, lvl := range []string{"ERR", "WRN", "INF", "DBG", "TRC"} {
        if strings.Contains(line, `"`+lvl+`"`) || strings.Contains(line, lvl) {
            switch lvl {
            case "ERR":
                return "error"
            case "WRN":
                return "warn"
            case "INF":
                return "info"
            case "DBG":
                return "debug"
            case "TRC":
                return "trace"
            }
        }
    }
    return "info"
}

// LogsModel is the Logs tab sub-model.
type LogsModel struct {
    viewport  viewport.Model
    lines     []tui.BotEventMsg
    filter    string // "", "trace","debug","info","warn","error"
    freeze    bool
    width     int
    height    int
}

func NewLogsModel(width, height int) LogsModel {
    vp := viewport.New(width, height-4)
    return LogsModel{viewport: vp, width: width, height: height}
}

func (m LogsModel) Init() tea.Cmd { return nil }

func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tui.BotEventMsg:
        m.lines = append(m.lines, msg)
        if len(m.lines) > maxLogLines {
            m.lines = m.lines[len(m.lines)-maxLogLines:]
        }
        if !m.freeze {
            m.viewport.SetContent(m.renderLines())
            m.viewport.GotoBottom()
        }
    case tea.KeyMsg:
        switch msg.String() {
        case "f", "F":
            m.freeze = !m.freeze
        case "t", "T":
            m.toggleFilter("trace")
        case "d", "D":
            m.toggleFilter("debug")
        case "i", "I":
            m.toggleFilter("info")
        case "w", "W":
            m.toggleFilter("warn")
        case "e", "E":
            m.toggleFilter("error")
        default:
            var cmd tea.Cmd
            m.viewport, cmd = m.viewport.Update(msg)
            return m, cmd
        }
        m.viewport.SetContent(m.renderLines())
    }
    return m, nil
}

func (m *LogsModel) toggleFilter(level string) {
    if m.filter == level {
        m.filter = ""
    } else {
        m.filter = level
    }
}

func (m LogsModel) renderLines() string {
    var sb strings.Builder
    for _, l := range m.lines {
        if m.filter != "" && l.Level != m.filter {
            continue
        }
        sb.WriteString(colorLine(l))
        sb.WriteString("\n")
    }
    return sb.String()
}

func colorLine(l tui.BotEventMsg) string {
    switch l.Level {
    case "error":
        return tui.StyleError.Render(l.Message)
    case "warn":
        return tui.StyleWarning.Render(l.Message)
    case "debug", "trace":
        return tui.StyleMuted.Render(l.Message)
    default:
        return l.Message
    }
}

func (m LogsModel) View() string {
    freeze := ""
    if m.freeze {
        freeze = tui.StyleWarning.Render(" [FROZEN]")
    }
    filter := ""
    if m.filter != "" {
        filter = fmt.Sprintf(" filter:%s", m.filter)
    }
    help := tui.StyleHelpBar.Render(
        fmt.Sprintf("F:freeze%s%s  T/D/I/W/E:filter  ↑↓:scroll", freeze, filter),
    )
    return lipgloss.JoinVertical(lipgloss.Left,
        m.viewport.View(),
        help,
    )
}
```

**Step 2: Build check**

```bash
go build ./internal/tui/...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add internal/tui/tabs/logs.go
git commit -m "feat: add Logs tab with zerolog integration and filter"
```

---

## Task 6: Overview tab

**Files:**
- Create: `internal/tui/tabs/overview.go`

**Step 1: Create overview.go**

```go
// internal/tui/tabs/overview.go
package tabs

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

type SubsystemStatus struct {
    Name   string
    Active bool
}

type OverviewModel struct {
    subsystems []SubsystemStatus
    balance    float64
    openOrders int
    positions  int
    pnlToday   float64
    traders    int
    width      int
    height     int
}

func NewOverviewModel(width, height int) OverviewModel {
    return OverviewModel{
        width:  width,
        height: height,
        subsystems: []SubsystemStatus{
            {Name: "WebSocket", Active: false},
            {Name: "Monitor", Active: false},
            {Name: "Trades Monitor", Active: false},
            {Name: "Trading Engine", Active: false},
            {Name: "Copytrading", Active: false},
        },
    }
}

func (m OverviewModel) Init() tea.Cmd { return nil }

func (m OverviewModel) Update(msg tea.Msg) (OverviewModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tui.SubsystemStatusMsg:
        for i, s := range m.subsystems {
            if s.Name == msg.Name {
                m.subsystems[i].Active = msg.Active
            }
        }
    case tui.BalanceMsg:
        m.balance = msg.USDC
    }
    return m, nil
}

func (m OverviewModel) View() string {
    half := m.width / 2

    // Left: subsystems
    var left strings.Builder
    left.WriteString(tui.StyleBold.Render("Подсистемы") + "\n")
    left.WriteString(strings.Repeat("─", half-4) + "\n")
    for _, s := range m.subsystems {
        dot := tui.StyleSuccess.Render("●")
        if !s.Active {
            dot = tui.StyleMuted.Render("○")
        }
        left.WriteString(fmt.Sprintf("  %s  %-20s\n", dot, s.Name))
    }

    // Right: quick stats
    var right strings.Builder
    right.WriteString(tui.StyleBold.Render("Быстрая статистика") + "\n")
    right.WriteString(strings.Repeat("─", half-4) + "\n")
    right.WriteString(fmt.Sprintf("  Баланс USDC      %.2f\n", m.balance))
    right.WriteString(fmt.Sprintf("  Открытых ордеров %d\n", m.openOrders))
    right.WriteString(fmt.Sprintf("  Позиций          %d\n", m.positions))
    pnlStr := fmt.Sprintf("%+.2f", m.pnlToday)
    if m.pnlToday >= 0 {
        pnlStr = tui.StyleSuccess.Render(pnlStr)
    } else {
        pnlStr = tui.StyleError.Render(pnlStr)
    }
    right.WriteString(fmt.Sprintf("  P&L сегодня      %s\n", pnlStr))
    right.WriteString(fmt.Sprintf("  Копируемых трейд. %d\n", m.traders))

    leftBox := tui.StyleBorder.Width(half - 2).Render(left.String())
    rightBox := tui.StyleBorder.Width(half - 2).Render(right.String())

    return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}
```

**Step 2: Build check**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/tabs/overview.go
git commit -m "feat: add Overview tab with subsystem status and quick stats"
```

---

## Task 7: Orders, Positions, Copytrading tabs (scaffold)

**Files:**
- Create: `internal/tui/tabs/orders.go`
- Create: `internal/tui/tabs/positions.go`
- Create: `internal/tui/tabs/copytrading.go`

These tabs use `github.com/charmbracelet/bubbles/table`. Scaffold with placeholder data first; wire real API data in Task 14.

**Step 1: Create orders.go**

```go
// internal/tui/tabs/orders.go
package tabs

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type OrdersModel struct {
    table  table.Model
    width  int
    height int
}

func NewOrdersModel(width, height int) OrdersModel {
    cols := []table.Column{
        {Title: "Market", Width: 30},
        {Title: "Side", Width: 6},
        {Title: "Price", Width: 10},
        {Title: "Size", Width: 10},
        {Title: "Filled", Width: 10},
        {Title: "Status", Width: 10},
        {Title: "Age", Width: 10},
    }
    t := table.New(
        table.WithColumns(cols),
        table.WithFocused(true),
        table.WithHeight(height-6),
    )
    s := table.DefaultStyles()
    s.Header = s.Header.Bold(true)
    t.SetStyles(s)
    return OrdersModel{table: t, width: width, height: height}
}

func (m OrdersModel) Init() tea.Cmd { return nil }

func (m OrdersModel) Update(msg tea.Msg) (OrdersModel, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "d", "D":
            // TODO: cancel selected order (Task 14)
        case "a", "A":
            // TODO: cancel all orders (Task 14)
        }
    }
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m OrdersModel) View() string {
    help := StyleHelpBar.Render("↑↓: navigate  D: cancel order  A: cancel all")
    return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), help)
}
```

**Step 2: Create positions.go**

```go
// internal/tui/tabs/positions.go
package tabs

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type PositionsModel struct {
    table  table.Model
    width  int
    height int
}

func NewPositionsModel(width, height int) PositionsModel {
    cols := []table.Column{
        {Title: "Market", Width: 30},
        {Title: "Side", Width: 6},
        {Title: "Size", Width: 10},
        {Title: "Entry", Width: 10},
        {Title: "Current", Width: 10},
        {Title: "P&L", Width: 10},
        {Title: "P&L%", Width: 8},
    }
    t := table.New(
        table.WithColumns(cols),
        table.WithFocused(true),
        table.WithHeight(height-6),
    )
    s := table.DefaultStyles()
    s.Header = s.Header.Bold(true)
    t.SetStyles(s)
    return PositionsModel{table: t, width: width, height: height}
}

func (m PositionsModel) Init() tea.Cmd { return nil }

func (m PositionsModel) Update(msg tea.Msg) (PositionsModel, tea.Cmd) {
    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m PositionsModel) View() string {
    help := StyleHelpBar.Render("↑↓: navigate  sorted by P&L")
    return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), help)
}
```

**Step 3: Create copytrading.go**

```go
// internal/tui/tabs/copytrading.go
package tabs

import (
    "strings"

    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

type CopytradingModel struct {
    tradersTable table.Model
    recentTrades []string
    width        int
    height       int
}

func NewCopytradingModel(width, height int) CopytradingModel {
    cols := []table.Column{
        {Title: "Адрес", Width: 20},
        {Title: "Label", Width: 16},
        {Title: "Статус", Width: 8},
        {Title: "Алл. %", Width: 8},
    }
    t := table.New(
        table.WithColumns(cols),
        table.WithFocused(true),
        table.WithHeight(height/2-3),
    )
    s := table.DefaultStyles()
    s.Header = s.Header.Bold(true)
    t.SetStyles(s)
    return CopytradingModel{tradersTable: t, width: width, height: height}
}

func (m CopytradingModel) Init() tea.Cmd { return nil }

func (m CopytradingModel) Update(msg tea.Msg) (CopytradingModel, tea.Cmd) {
    var cmd tea.Cmd
    switch msg.(type) {
    case tui.BotEventMsg:
        // TODO: filter copytrading events (Task 14)
    }
    m.tradersTable, cmd = m.tradersTable.Update(msg)
    return m, cmd
}

func (m CopytradingModel) View() string {
    var sb strings.Builder
    sb.WriteString(tui.StyleBold.Render("Отслеживаемые трейдеры") + "\n")
    sb.WriteString(m.tradersTable.View() + "\n\n")
    sb.WriteString(tui.StyleBold.Render("Последние скопированные сделки") + "\n")
    for _, t := range m.recentTrades {
        sb.WriteString("  " + t + "\n")
    }
    return lipgloss.NewStyle().Padding(0, 1).Render(sb.String())
}
```

**Step 4: Build check**

```bash
go build ./internal/tui/...
```

**Step 5: Commit**

```bash
git add internal/tui/tabs/orders.go internal/tui/tabs/positions.go internal/tui/tabs/copytrading.go
git commit -m "feat: scaffold Orders, Positions, Copytrading tabs"
```

---

## Task 8: Settings tab

**Files:**
- Create: `internal/tui/tabs/settings.go`

This is the largest tab. It uses `github.com/charmbracelet/bubbles/textinput`.

**Step 1: Define field descriptors**

```go
// internal/tui/tabs/settings.go
package tabs

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

type FieldKind int

const (
    KindString FieldKind = iota
    KindPassword
    KindInt
    KindBool
)

type FieldDef struct {
    Section string
    Label   string
    Tooltip string
    Kind    FieldKind
    Get     func(*config.Config) string
    Set     func(*config.Config, string) error
}

var allFields = []FieldDef{
    // Auth
    {
        Section: "Auth", Label: "Private Key", Kind: KindPassword,
        Tooltip: "Hex-ключ вашего Ethereum кошелька (без 0x).\nИспользуется для подписи ордеров EIP-712\nи деривации адреса кошелька.",
        Get: func(c *config.Config) string { return c.Auth.PrivateKey },
        Set: func(c *config.Config, v string) error { c.Auth.PrivateKey = v; return nil },
    },
    {
        Section: "Auth", Label: "API Key", Kind: KindString,
        Tooltip: "API-ключ Polymarket CLOB.\nПолучите через POST /auth/api-key\nили в личном кабинете.",
        Get: func(c *config.Config) string { return c.Auth.APIKey },
        Set: func(c *config.Config, v string) error { c.Auth.APIKey = v; return nil },
    },
    {
        Section: "Auth", Label: "API Secret", Kind: KindPassword,
        Tooltip: "Секрет для HMAC-SHA256 подписи\nкаждого запроса к CLOB API (L2 auth).",
        Get: func(c *config.Config) string { return c.Auth.APISecret },
        Set: func(c *config.Config, v string) error { c.Auth.APISecret = v; return nil },
    },
    {
        Section: "Auth", Label: "Passphrase", Kind: KindPassword,
        Tooltip: "Passphrase для L2 аутентификации.\nОтправляется в заголовке POLY_PASSPHRASE.",
        Get: func(c *config.Config) string { return c.Auth.Passphrase },
        Set: func(c *config.Config, v string) error { c.Auth.Passphrase = v; return nil },
    },
    {
        Section: "Auth", Label: "Chain ID", Kind: KindInt,
        Tooltip: "ID блокчейна.\n137 = Polygon Mainnet\n80002 = Amoy Testnet",
        Get: func(c *config.Config) string { return strconv.FormatInt(c.Auth.ChainID, 10) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.ParseInt(v, 10, 64)
            if err != nil { return err }
            c.Auth.ChainID = n; return nil
        },
    },
    // API
    {
        Section: "API", Label: "Timeout (sec)", Kind: KindInt,
        Tooltip: "Таймаут HTTP запросов в секундах.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.API.TimeoutSec) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.API.TimeoutSec = n; return nil
        },
    },
    {
        Section: "API", Label: "Max Retries", Kind: KindInt,
        Tooltip: "Максимальное число повторных попыток\nпри ошибках HTTP запросов.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.API.MaxRetries) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.API.MaxRetries = n; return nil
        },
    },
    // Monitor
    {
        Section: "Monitor", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить мониторинг рынков через Gamma API.\nГенерирует алерты при совпадении правил.",
        Get: func(c *config.Config) string { return boolStr(c.Monitor.Enabled) },
        Set: func(c *config.Config, v string) error { c.Monitor.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Monitor", Label: "Poll Interval (ms)", Kind: KindInt,
        Tooltip: "Интервал опроса Gamma API в миллисекундах.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.PollIntervalMs) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.Monitor.PollIntervalMs = n; return nil
        },
    },
    // Monitor.Trades
    {
        Section: "Trades Monitor", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить TradesMonitor — отслеживание\nордеров, сделок и позиций. Требует L2 auth.",
        Get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.Enabled) },
        Set: func(c *config.Config, v string) error { c.Monitor.Trades.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Trades Monitor", Label: "Poll Interval (ms)", Kind: KindInt,
        Tooltip: "Интервал опроса CLOB/Data API в миллисекундах.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.PollIntervalMs) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.Monitor.Trades.PollIntervalMs = n; return nil
        },
    },
    {
        Section: "Trades Monitor", Label: "Alert On Fill", Kind: KindBool,
        Tooltip: "Отправлять уведомление при исполнении ордера.",
        Get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnFill) },
        Set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnFill = parseBool(v); return nil },
    },
    {
        Section: "Trades Monitor", Label: "Alert On Cancel", Kind: KindBool,
        Tooltip: "Отправлять уведомление при отмене ордера.",
        Get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnCancel) },
        Set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnCancel = parseBool(v); return nil },
    },
    {
        Section: "Trades Monitor", Label: "Trades Limit", Kind: KindInt,
        Tooltip: "Максимальное количество сделок\nв одном запросе к API.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.TradesLimit) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.Monitor.Trades.TradesLimit = n; return nil
        },
    },
    // Trading
    {
        Section: "Trading", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить торговый движок и стратегии.",
        Get: func(c *config.Config) string { return boolStr(c.Trading.Enabled) },
        Set: func(c *config.Config, v string) error { c.Trading.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Trading", Label: "Max Position USD", Kind: KindString,
        Tooltip: "Максимальный размер одной позиции в USD.",
        Get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.MaxPositionUSD) },
        Set: func(c *config.Config, v string) error {
            f, err := strconv.ParseFloat(v, 64)
            if err != nil { return err }
            c.Trading.MaxPositionUSD = f; return nil
        },
    },
    {
        Section: "Trading", Label: "Slippage %", Kind: KindString,
        Tooltip: "Допустимое проскальзывание в процентах.",
        Get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.SlippagePct) },
        Set: func(c *config.Config, v string) error {
            f, err := strconv.ParseFloat(v, 64)
            if err != nil { return err }
            c.Trading.SlippagePct = f; return nil
        },
    },
    {
        Section: "Trading", Label: "NegRisk", Kind: KindBool,
        Tooltip: "Использовать NegRisk контракт\n(0xC5d563A36AE78145C45a50134d48A1215220f80a)\nвместо стандартного.",
        Get: func(c *config.Config) string { return boolStr(c.Trading.NegRisk) },
        Set: func(c *config.Config, v string) error { c.Trading.NegRisk = parseBool(v); return nil },
    },
    // Copytrading
    {
        Section: "Copytrading", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить копитрейдинг.\nТребует L2 auth + private_key + database.",
        Get: func(c *config.Config) string { return boolStr(c.Copytrading.Enabled) },
        Set: func(c *config.Config, v string) error { c.Copytrading.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Copytrading", Label: "Poll Interval (ms)", Kind: KindInt,
        Tooltip: "Интервал проверки позиций трейдеров.",
        Get: func(c *config.Config) string { return strconv.Itoa(c.Copytrading.PollIntervalMs) },
        Set: func(c *config.Config, v string) error {
            n, err := strconv.Atoi(v)
            if err != nil { return err }
            c.Copytrading.PollIntervalMs = n; return nil
        },
    },
    {
        Section: "Copytrading", Label: "Size Mode", Kind: KindString,
        Tooltip: "Метод расчёта размера позиции:\n'proportional' — пропорционально балансу\n'fixed_pct' — фиксированный % от баланса",
        Get: func(c *config.Config) string { return c.Copytrading.SizeMode },
        Set: func(c *config.Config, v string) error { c.Copytrading.SizeMode = v; return nil },
    },
    // Telegram
    {
        Section: "Telegram", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить Telegram уведомления.",
        Get: func(c *config.Config) string { return boolStr(c.Telegram.Enabled) },
        Set: func(c *config.Config, v string) error { c.Telegram.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Telegram", Label: "Bot Token", Kind: KindPassword,
        Tooltip: "Токен Telegram бота от @BotFather.",
        Get: func(c *config.Config) string { return c.Telegram.BotToken },
        Set: func(c *config.Config, v string) error { c.Telegram.BotToken = v; return nil },
    },
    {
        Section: "Telegram", Label: "Chat ID", Kind: KindString,
        Tooltip: "ID чата или канала для уведомлений.",
        Get: func(c *config.Config) string { return c.Telegram.ChatID },
        Set: func(c *config.Config, v string) error { c.Telegram.ChatID = v; return nil },
    },
    // Database
    {
        Section: "Database", Label: "Enabled", Kind: KindBool,
        Tooltip: "Включить SQLite базу данных.\nТребуется для копитрейдинга.",
        Get: func(c *config.Config) string { return boolStr(c.Database.Enabled) },
        Set: func(c *config.Config, v string) error { c.Database.Enabled = parseBool(v); return nil },
    },
    {
        Section: "Database", Label: "Path", Kind: KindString,
        Tooltip: "Путь к файлу SQLite базы данных.",
        Get: func(c *config.Config) string { return c.Database.Path },
        Set: func(c *config.Config, v string) error { c.Database.Path = v; return nil },
    },
    // Log
    {
        Section: "Log", Label: "Level", Kind: KindString,
        Tooltip: "Уровень логирования:\ntrace / debug / info / warn / error",
        Get: func(c *config.Config) string { return c.Log.Level },
        Set: func(c *config.Config, v string) error { c.Log.Level = v; return nil },
    },
    {
        Section: "Log", Label: "Format", Kind: KindString,
        Tooltip: "Формат логов:\n'pretty' — цветной человекочитаемый\n'json' — структурированный JSON",
        Get: func(c *config.Config) string { return c.Log.Format },
        Set: func(c *config.Config, v string) error { c.Log.Format = v; return nil },
    },
}

func boolStr(b bool) string {
    if b { return "true" }
    return "false"
}
func parseBool(s string) bool {
    return strings.ToLower(s) == "true" || s == "1" || s == "yes"
}
```

**Step 2: Implement SettingsModel**

Append to the same file:

```go
type SettingsModel struct {
    cfg      config.Config
    original config.Config
    cfgPath  string

    fields   []FieldDef
    inputs   []textinput.Model
    cursor   int
    editing  bool
    modified []bool
    err      string

    sections     []string
    activeSection int

    width  int
    height int

    // OnSave is called after writing config.toml (triggers hot reload)
    OnSave func(path string)
}

func NewSettingsModel(cfg *config.Config, cfgPath string, width, height int, onSave func(string)) SettingsModel {
    m := SettingsModel{
        cfg:     *cfg,
        original: *cfg,
        cfgPath: cfgPath,
        fields:  allFields,
        width:   width,
        height:  height,
        OnSave:  onSave,
    }

    // build section list
    seen := map[string]bool{}
    for _, f := range allFields {
        if !seen[f.Section] {
            m.sections = append(m.sections, f.Section)
            seen[f.Section] = true
        }
    }

    // create textinput for each field
    m.inputs = make([]textinput.Model, len(allFields))
    m.modified = make([]bool, len(allFields))
    for i, f := range allFields {
        ti := textinput.New()
        ti.SetValue(f.Get(cfg))
        if f.Kind == KindPassword {
            ti.EchoMode = textinput.EchoPassword
        }
        m.inputs[i] = ti
    }
    return m
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) currentSectionFields() ([]int, []FieldDef) {
    section := m.sections[m.activeSection]
    var idxs []int
    var fields []FieldDef
    for i, f := range m.fields {
        if f.Section == section {
            idxs = append(idxs, i)
            fields = append(fields, f)
        }
    }
    return idxs, fields
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
    if m.editing {
        return m.updateEditing(msg)
    }
    switch msg := msg.(type) {
    case tui.ConfigReloadedMsg:
        m.cfg = *msg.Config
        m.original = *msg.Config
        for i, f := range m.fields {
            m.inputs[i].SetValue(f.Get(msg.Config))
            m.modified[i] = false
        }
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            idxs, _ := m.currentSectionFields()
            if len(idxs) > 0 {
                pos := indexOf(idxs, m.cursor)
                if pos > 0 { m.cursor = idxs[pos-1] }
            }
        case "down", "j":
            idxs, _ := m.currentSectionFields()
            if len(idxs) > 0 {
                pos := indexOf(idxs, m.cursor)
                if pos < len(idxs)-1 { m.cursor = idxs[pos+1] }
            }
        case "left", "h":
            if m.activeSection > 0 { m.activeSection-- }
            idxs, _ := m.currentSectionFields()
            if len(idxs) > 0 { m.cursor = idxs[0] }
        case "right", "l":
            if m.activeSection < len(m.sections)-1 { m.activeSection++ }
            idxs, _ := m.currentSectionFields()
            if len(idxs) > 0 { m.cursor = idxs[0] }
        case "enter":
            m.editing = true
            m.inputs[m.cursor].Focus()
            if m.fields[m.cursor].Kind == KindPassword {
                m.inputs[m.cursor].EchoMode = textinput.EchoNormal
            }
        case "s", "S":
            m.err = ""
            // apply all modified inputs to cfg
            for i, f := range m.fields {
                if err := f.Set(&m.cfg, m.inputs[i].Value()); err != nil {
                    m.err = fmt.Sprintf("Ошибка в поле %s: %v", f.Label, err)
                    return m, nil
                }
            }
            if err := saveConfig(m.cfgPath, &m.cfg); err != nil {
                m.err = fmt.Sprintf("Ошибка сохранения: %v", err)
                return m, nil
            }
            m.original = m.cfg
            for i := range m.modified { m.modified[i] = false }
            if m.OnSave != nil { m.OnSave(m.cfgPath) }
        case "r", "R":
            m.cfg = m.original
            for i, f := range m.fields {
                m.inputs[i].SetValue(f.Get(&m.original))
                m.modified[i] = false
            }
        }
    }
    return m, nil
}

func (m SettingsModel) updateEditing(msg tea.Msg) (SettingsModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter", "esc":
            m.editing = false
            m.inputs[m.cursor].Blur()
            orig := m.fields[m.cursor].Get(&m.original)
            m.modified[m.cursor] = m.inputs[m.cursor].Value() != orig
            if m.fields[m.cursor].Kind == KindPassword {
                m.inputs[m.cursor].EchoMode = textinput.EchoPassword
            }
            return m, nil
        }
    }
    var cmd tea.Cmd
    m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
    return m, cmd
}

func indexOf(slice []int, val int) int {
    for i, v := range slice {
        if v == val { return i }
    }
    return 0
}

func (m SettingsModel) View() string {
    halfW := m.width / 2

    // Section tabs
    var sectionBar strings.Builder
    for i, s := range m.sections {
        if i == m.activeSection {
            sectionBar.WriteString(tui.StyleTabActive.Render(" " + s + " "))
        } else {
            sectionBar.WriteString(tui.StyleTabInactive.Render(" " + s + " "))
        }
    }

    // Left: fields
    idxs, _ := m.currentSectionFields()
    var leftContent strings.Builder
    for _, idx := range idxs {
        f := m.fields[idx]
        val := m.inputs[idx].View()
        mod := ""
        if m.modified[idx] { mod = tui.StyleWarning.Render(" ●") }
        cursor := "  "
        if idx == m.cursor { cursor = tui.StylePrimary.Render("▶ ") }
        leftContent.WriteString(fmt.Sprintf("%s%-20s %s%s\n", cursor, f.Label, val, mod))
    }
    leftBox := tui.StyleBorder.Width(halfW - 2).Height(m.height - 8).Render(leftContent.String())

    // Right: tooltip
    tooltip := ""
    if m.cursor < len(m.fields) {
        tooltip = m.fields[m.cursor].Tooltip
    }
    rightBox := tui.StyleBorder.Width(halfW - 2).Height(m.height - 8).Render(
        tui.StyleBold.Render("Подсказка") + "\n\n" + tui.StyleTooltip.Render(tooltip),
    )

    errLine := ""
    if m.err != "" {
        errLine = "\n" + tui.StyleError.Render(m.err)
    }

    help := tui.StyleHelpBar.Render("↑↓:поле  ←→:секция  Enter:редактировать  S:сохранить  R:сбросить")

    return lipgloss.JoinVertical(lipgloss.Left,
        sectionBar.String(),
        lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox),
        errLine,
        help,
    )
}

// saveConfig serialises cfg back to TOML at path.
// Uses BurntSushi/toml encoder.
func saveConfig(path string, cfg *config.Config) error {
    // Simple approach: marshal to TOML using encoder
    // Note: import "github.com/BurntSushi/toml" at top of file
    f, err := os.Create(path)
    if err != nil { return err }
    defer f.Close()
    return toml.NewEncoder(f).Encode(cfg)
}
```

Add missing imports at the top of the file:
```go
import (
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/BurntSushi/toml"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)
```

Also add `StylePrimary` to `styles.go`:
```go
StylePrimary = lipgloss.NewStyle().Foreground(ColorPrimary)
```

**Step 3: Build check**

```bash
go build ./internal/tui/...
```
Expected: no errors.

**Step 4: Commit**

```bash
git add internal/tui/tabs/settings.go internal/tui/styles.go
git commit -m "feat: add Settings tab with all fields, tooltips, save/reset"
```

---

## Task 9: First-run Wizard

**Files:**
- Create: `internal/tui/wizard/wizard.go`

**Step 1: Create wizard.go**

```go
// internal/tui/wizard/wizard.go
package wizard

import (
    "fmt"
    "os"
    "strings"

    "github.com/BurntSushi/toml"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/tui"
)

type Step struct {
    Label   string
    Hint    string
    IsPass  bool
}

var steps = []Step{
    {
        Label:  "Private Key",
        Hint:   "Hex-ключ вашего Ethereum кошелька (без 0x).\nИспользуется для подписи ордеров и деривации адреса.",
        IsPass: true,
    },
    {
        Label:  "API Key",
        Hint:   "API-ключ Polymarket CLOB.\nПолучите через POST /auth/api-key.",
        IsPass: false,
    },
    {
        Label:  "API Secret",
        Hint:   "Секрет для HMAC-SHA256 подписи запросов.",
        IsPass: true,
    },
    {
        Label:  "Passphrase",
        Hint:   "Passphrase для L2 аутентификации.",
        IsPass: true,
    },
}

// DoneMsg is emitted when wizard completes. Contains the generated config path.
type DoneMsg struct {
    ConfigPath string
}

type Model struct {
    step   int
    inputs []textinput.Model
    err    string
    width  int
    height int
    outPath string
}

func New(width, height int, outPath string) Model {
    inputs := make([]textinput.Model, len(steps))
    for i, s := range steps {
        ti := textinput.New()
        ti.Placeholder = s.Label
        if s.IsPass {
            ti.EchoMode = textinput.EchoPassword
        }
        inputs[i] = ti
    }
    inputs[0].Focus()
    return Model{inputs: inputs, width: width, height: height, outPath: outPath}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            val := strings.TrimSpace(m.inputs[m.step].Value())
            if val == "" {
                m.err = "Поле не может быть пустым"
                return m, nil
            }
            m.err = ""
            m.inputs[m.step].Blur()
            if m.step < len(steps)-1 {
                m.step++
                m.inputs[m.step].Focus()
                return m, textinput.Blink
            }
            // last step — generate config
            if err := m.writeConfig(); err != nil {
                m.err = fmt.Sprintf("Ошибка: %v", err)
                return m, nil
            }
            return m, func() tea.Msg { return DoneMsg{ConfigPath: m.outPath} }
        case "ctrl+c":
            return m, tea.Quit
        }
    }
    var cmd tea.Cmd
    m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
    return m, cmd
}

func (m Model) writeConfig() error {
    type minConfig struct {
        API  map[string]interface{} `toml:"api"`
        Auth map[string]interface{} `toml:"auth"`
        Log  map[string]interface{} `toml:"log"`
    }
    cfg := minConfig{
        API: map[string]interface{}{
            "clob_url":    "https://clob.polymarket.com",
            "gamma_url":   "https://gamma-api.polymarket.com",
            "data_url":    "https://data-api.polymarket.com",
            "ws_url":      "wss://ws-subscriptions-clob.polymarket.com/ws/",
            "timeout_sec": 10,
            "max_retries": 3,
        },
        Auth: map[string]interface{}{
            "private_key": m.inputs[0].Value(),
            "api_key":     m.inputs[1].Value(),
            "api_secret":  m.inputs[2].Value(),
            "passphrase":  m.inputs[3].Value(),
            "chain_id":    137,
        },
        Log: map[string]interface{}{
            "level":  "info",
            "format": "pretty",
        },
    }
    f, err := os.Create(m.outPath)
    if err != nil { return err }
    defer f.Close()
    return toml.NewEncoder(f).Encode(cfg)
}

func (m Model) View() string {
    s := steps[m.step]
    progress := fmt.Sprintf("Шаг %d/%d: %s", m.step+1, len(steps), s.Label)

    errLine := ""
    if m.err != "" {
        errLine = "\n" + tui.StyleError.Render(m.err)
    }

    body := lipgloss.JoinVertical(lipgloss.Left,
        tui.StyleBold.Render(progress),
        "\n",
        m.inputs[m.step].View(),
        "\n",
        tui.StyleTooltip.Render(s.Hint),
        errLine,
        "\n",
        tui.StyleMuted.Render("[Enter] продолжить  [Ctrl+C] выход"),
    )

    box := tui.StyleBorder.
        Width(m.width - 4).
        Padding(1, 2).
        Render(body)

    title := tui.StyleHeader.Render("  polytrade-bot — Первичная настройка  ")
    return lipgloss.JoinVertical(lipgloss.Left, title, "\n", box)
}
```

**Step 2: Build check**

```bash
go build ./internal/tui/...
```

**Step 3: Commit**

```bash
git add internal/tui/wizard/wizard.go
git commit -m "feat: add first-run wizard for initial config generation"
```

---

## Task 10: TabBar and AppModel

**Files:**
- Create: `internal/tui/tabs.go`
- Create: `internal/tui/app.go`

**Step 1: Create tabs.go**

```go
// internal/tui/tabs.go
package tui

import (
    "strings"
)

type TabID int

const (
    TabOverview TabID = iota
    TabOrders
    TabPositions
    TabCopytrading
    TabLogs
    TabSettings
)

var tabNames = []string{
    "Overview", "Orders", "Positions", "Copytrading", "Logs", "Settings",
}

func RenderTabBar(active TabID, width int) string {
    var sb strings.Builder
    for i, name := range tabNames {
        if TabID(i) == active {
            sb.WriteString(StyleTabActive.Render(" " + name + " "))
        } else {
            sb.WriteString(StyleTabInactive.Render(" " + name + " "))
        }
    }
    // pad to width
    bar := sb.String()
    return StyleBorder.Width(width).Render(bar)
}
```

**Step 2: Create app.go**

```go
// internal/tui/app.go
package tui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/atlasdev/polytrade-bot/internal/config"
    "github.com/atlasdev/polytrade-bot/internal/tui/tabs"
)

type AppModel struct {
    activeTab  TabID
    overview   tabs.OverviewModel
    orders     tabs.OrdersModel
    positions  tabs.PositionsModel
    copytrader tabs.CopytradingModel
    logs       tabs.LogsModel
    settings   tabs.SettingsModel
    bus        *EventBus
    width      int
    height     int
    cfg        *config.Config
    wallet     string
    balance    float64
}

func NewAppModel(
    cfg *config.Config,
    cfgPath string,
    bus *EventBus,
    width, height int,
    onSave func(string),
) AppModel {
    cw := height - 6 // content height (minus header + tabbar + helpbar)
    return AppModel{
        cfg:        cfg,
        bus:        bus,
        width:      width,
        height:     height,
        overview:   tabs.NewOverviewModel(width, cw),
        orders:     tabs.NewOrdersModel(width, cw),
        positions:  tabs.NewPositionsModel(width, cw),
        copytrader: tabs.NewCopytradingModel(width, cw),
        logs:       tabs.NewLogsModel(width, cw),
        settings:   tabs.NewSettingsModel(cfg, cfgPath, width, cw, onSave),
    }
}

func (m AppModel) Init() tea.Cmd {
    return m.bus.WaitForEvent()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

    case tea.KeyMsg:
        // Global keys (skip when settings is editing)
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "tab":
            m.activeTab = (m.activeTab + 1) % TabSettings + 1 % (TabSettings + 1)
        case "shift+tab":
            if m.activeTab == 0 {
                m.activeTab = TabSettings
            } else {
                m.activeTab--
            }
        case "1":
            m.activeTab = TabOverview
        case "2":
            m.activeTab = TabOrders
        case "3":
            m.activeTab = TabPositions
        case "4":
            m.activeTab = TabCopytrading
        case "5":
            m.activeTab = TabLogs
        case "6":
            m.activeTab = TabSettings
        }

    case ConfigReloadedMsg:
        m.cfg = msg.Config
        var cmd tea.Cmd
        m.settings, cmd = m.settings.Update(msg)
        return m, tea.Batch(cmd, m.bus.WaitForEvent())

    case BalanceMsg:
        m.balance = msg.USDC

    case SubsystemStatusMsg:
        var cmd tea.Cmd
        m.overview, cmd = m.overview.Update(msg)
        return m, tea.Batch(cmd, m.bus.WaitForEvent())

    case BotEventMsg:
        var cmd tea.Cmd
        m.logs, cmd = m.logs.Update(msg)
        return m, tea.Batch(cmd, m.bus.WaitForEvent())
    }

    // Route to active tab
    var cmd tea.Cmd
    switch m.activeTab {
    case TabOverview:
        m.overview, cmd = m.overview.Update(msg)
    case TabOrders:
        m.orders, cmd = m.orders.Update(msg)
    case TabPositions:
        m.positions, cmd = m.positions.Update(msg)
    case TabCopytrading:
        m.copytrader, cmd = m.copytrader.Update(msg)
    case TabLogs:
        m.logs, cmd = m.logs.Update(msg)
    case TabSettings:
        m.settings, cmd = m.settings.Update(msg)
    }

    return m, tea.Batch(cmd, m.bus.WaitForEvent())
}

func (m AppModel) View() string {
    walletShort := ""
    if len(m.wallet) > 10 {
        walletShort = m.wallet[:6] + "..." + m.wallet[len(m.wallet)-4:]
    }
    header := StyleHeader.Width(m.width).Render(
        fmt.Sprintf("  polytrade-bot  ●Running   Wallet: %s   USDC: %.2f  ", walletShort, m.balance),
    )

    tabBar := RenderTabBar(m.activeTab, m.width)

    var content string
    switch m.activeTab {
    case TabOverview:
        content = m.overview.View()
    case TabOrders:
        content = m.orders.View()
    case TabPositions:
        content = m.positions.View()
    case TabCopytrading:
        content = m.copytrader.View()
    case TabLogs:
        content = m.logs.View()
    case TabSettings:
        content = m.settings.View()
    }

    helpBar := StyleHelpBar.Width(m.width).Render(
        "  Tab/Shift+Tab: вкладка  1-6: прямой переход  q: выход  ",
    )

    return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, content, helpBar)
}
```

**Step 3: Build check**

```bash
go build ./internal/tui/...
```

**Step 4: Commit**

```bash
git add internal/tui/tabs.go internal/tui/app.go
git commit -m "feat: add TabBar and AppModel (root Bubble Tea model)"
```

---

## Task 11: Wire TUI into main.go

**Files:**
- Modify: `cmd/bot/main.go`
- Modify: `internal/logger/logger.go`

**Step 1: Add --no-tui flag and wizard check to main.go**

In `cmd/bot/main.go`, add after `cfgPath := flag.String(...)`:

```go
noTUI := flag.Bool("no-tui", false, "disable TUI, use plain log output")
```

After `flag.Parse()`, add wizard check:

```go
// First-run wizard if config doesn't exist
if _, err := os.Stat(*cfgPath); os.IsNotExist(err) && !*noTUI {
    p := tea.NewProgram(wizard.New(80, 24, *cfgPath), tea.WithAltScreen())
    finalModel, err := p.Run()
    if err != nil {
        return fmt.Errorf("wizard: %w", err)
    }
    _ = finalModel
    // reload config after wizard wrote it
}
```

After all subsystems are started but before the blocking select, add TUI launch:

```go
if !*noTUI {
    bus := tui.NewEventBus()

    // Wire log output to TUI EventBus
    logWriter := tabs.NewLogWriter(bus)
    log = logger.NewWithWriter(cfg.Log.Level, cfg.Log.Format, logWriter)

    // ConfigWatcher
    watcher, _ := config.NewWatcher(*cfgPath, func(newCfg *config.Config) {
        bus.Send(tui.ConfigReloadedMsg{Config: newCfg})
    })
    go watcher.Run(ctx)

    // Send initial subsystem statuses
    bus.Send(tui.SubsystemStatusMsg{Name: "WebSocket", Active: true})
    if cfg.Monitor.Enabled {
        bus.Send(tui.SubsystemStatusMsg{Name: "Monitor", Active: true})
    }
    if cfg.Monitor.Trades.Enabled && l2Creds != nil {
        bus.Send(tui.SubsystemStatusMsg{Name: "Trades Monitor", Active: true})
    }
    if cfg.Trading.Enabled {
        bus.Send(tui.SubsystemStatusMsg{Name: "Trading Engine", Active: true})
    }
    if cfg.Copytrading.Enabled {
        bus.Send(tui.SubsystemStatusMsg{Name: "Copytrading", Active: true})
    }

    appModel := tui.NewAppModel(cfg, *cfgPath, bus, 0, 0, func(path string) {
        // onSave: ConfigWatcher will pick up the file change automatically
    })

    p := tea.NewProgram(appModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
    if _, err := p.Run(); err != nil {
        return fmt.Errorf("tui: %w", err)
    }
    cancel() // shut down bot when TUI exits
    return nil
}
```

**Step 2: Add NewWithWriter to logger.go**

In `internal/logger/logger.go`, add:

```go
import "io"

// NewWithWriter creates a logger that writes to the given writer (for TUI integration).
func NewWithWriter(level, format string, w io.Writer) zerolog.Logger {
    // same as New() but use w instead of os.Stdout
    // copy existing New() logic, replace os.Stdout with w
}
```

**Step 3: Add required imports to main.go**

```go
import (
    // existing imports...
    "github.com/atlasdev/polytrade-bot/internal/tui"
    "github.com/atlasdev/polytrade-bot/internal/tui/tabs"
    "github.com/atlasdev/polytrade-bot/internal/tui/wizard"
    tea "github.com/charmbracelet/bubbletea"
)
```

**Step 4: Build and test**

```bash
go build ./...
go vet ./...
```

Run without config to test wizard:
```bash
./polytrade-bot --config /tmp/test-config.toml
```
Expected: wizard appears, fills in 4 fields, generates config.

Run with existing config:
```bash
./polytrade-bot --config config.toml
```
Expected: TUI dashboard appears with all tabs navigable.

Run headless:
```bash
./polytrade-bot --config config.toml --no-tui
```
Expected: plain zerolog output, no TUI.

**Step 5: Commit**

```bash
git add cmd/bot/main.go internal/logger/logger.go
git commit -m "feat: wire TUI into main.go, add --no-tui flag and first-run wizard"
```

---

## Task 12: Final polish and verification

**Files:**
- Modify: various tabs as needed

**Step 1: Fix tab cycling in app.go**

The tab cycling logic `(m.activeTab + 1) % TabSettings + 1 % (TabSettings + 1)` is wrong. Replace with:

```go
case "tab":
    if m.activeTab < TabSettings {
        m.activeTab++
    } else {
        m.activeTab = TabOverview
    }
```

**Step 2: Run all tests**

```bash
go test ./...
```
Expected: all pass.

**Step 3: Build release binary**

```bash
go build -ldflags="-s -w" -o polytrade-bot ./cmd/bot/
```

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete TUI dashboard with tabs, wizard, settings, hot reload"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Add deps | go.mod |
| 2 | ConfigWatcher | internal/config/watcher.go |
| 3 | Messages + EventBus | internal/tui/messages.go |
| 4 | Styles + Keys | internal/tui/styles.go, keys.go |
| 5 | Logs tab | internal/tui/tabs/logs.go |
| 6 | Overview tab | internal/tui/tabs/overview.go |
| 7 | Orders/Positions/Copytrading | internal/tui/tabs/*.go |
| 8 | Settings tab | internal/tui/tabs/settings.go |
| 9 | Wizard | internal/tui/wizard/wizard.go |
| 10 | TabBar + AppModel | internal/tui/app.go, tabs.go |
| 11 | Wire main.go | cmd/bot/main.go |
| 12 | Polish + verify | various |
