package tui

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

// ConfigReloadedMsg is sent when config.toml changes on disk.
type ConfigReloadedMsg struct {
	Config *config.Config
}

// BotEventMsg carries a log line from a subsystem.
type BotEventMsg struct {
	Level   string // "trace","debug","info","warn","error"
	Message string
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

// LanguageChangedMsg is sent when the user switches the UI language.
type LanguageChangedMsg struct{}

// OrdersUpdateMsg carries a fresh snapshot of open orders from TradesMonitor.
type OrdersUpdateMsg struct {
	Rows []OrderRow
}

// PositionsUpdateMsg carries a fresh snapshot of positions from TradesMonitor.
type PositionsUpdateMsg struct {
	Rows []PositionRow
}

// EventBus bridges bot goroutines to the Bubble Tea loop.
// Supports multiple subscribers via Tap(); the primary channel is
// used by the TUI via WaitForEvent().
type EventBus struct {
	ch   chan tea.Msg
	mu   sync.Mutex
	taps []chan tea.Msg
}

// NewEventBus creates an EventBus with a buffered channel.
func NewEventBus() *EventBus {
	return &EventBus{ch: make(chan tea.Msg, 512)}
}

// Send enqueues a message to the TUI channel and all tap subscribers (non-blocking; drops if full).
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

// Tap creates a new subscriber channel that receives a copy of every future Send() call.
// The caller is responsible for draining the channel to prevent blocking.
func (b *EventBus) Tap() <-chan tea.Msg {
	ch := make(chan tea.Msg, 512)
	b.mu.Lock()
	b.taps = append(b.taps, ch)
	b.mu.Unlock()
	return ch
}

// WaitForEvent returns a tea.Cmd that blocks until the next EventBus message.
func (b *EventBus) WaitForEvent() tea.Cmd {
	return func() tea.Msg {
		return <-b.ch
	}
}

// WalletAddedMsg is sent when a new wallet is added to the manager.
type WalletAddedMsg struct{ ID string }

// WalletRemovedMsg is sent when a wallet is removed.
type WalletRemovedMsg struct{ ID string }

// WalletChangedMsg is sent when a wallet's enabled state changes.
type WalletChangedMsg struct {
	ID      string
	Enabled bool
}

// WalletStatsMsg carries a statistics snapshot for one wallet.
type WalletStatsMsg struct {
	ID          string
	BalanceUSD  float64
	PnLUSD      float64
	OpenOrders  int
	TotalTrades int
}
