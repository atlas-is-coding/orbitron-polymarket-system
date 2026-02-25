package tui

import (
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

// EventBus bridges bot goroutines to the Bubble Tea loop.
type EventBus struct {
	ch chan tea.Msg
}

// NewEventBus creates an EventBus with a buffered channel.
func NewEventBus() *EventBus {
	return &EventBus{ch: make(chan tea.Msg, 512)}
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
