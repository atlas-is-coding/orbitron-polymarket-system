package telegrambot

import (
	"sync"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/health"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// SubsystemStatus holds name + active state.
type SubsystemStatus struct {
	Name   string
	Active bool
}

// WalletEntry holds the latest stats snapshot for one wallet in the bot state.
type WalletEntry struct {
	ID      string
	Label   string
	Enabled bool
	Primary bool
	Balance float64
	PnL     float64
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
	wallets    map[string]WalletEntry

	// Recent copy trades feed (capped at 10).
	copyTrades []string

	// Health: latest snapshot from health.Service.
	healthSnap   health.HealthSnapshot
	healthLoaded bool

	// Markets: snapshot of the currently displayed market list (for index-based navigation).
	viewMarkets []gamma.Market

	// Navigation: ID of the active menu message (for edit-in-place).
	menuMsgID int

	// Conversation state: non-empty while bot awaits text input from user.
	// Examples: "addtrader_addr", "addtrader_label", "addtrader_alloc",
	//           "edit:monitor.poll_interval_ms"
	pendingInput string
	// pendingData accumulates values across conversation steps.
	// For addtrader: "addr|label" (pipe-separated as steps complete).
	pendingData string
}

// NewBotState creates an empty BotState.
func NewBotState() *BotState {
	return &BotState{
		subsystems: make(map[string]bool),
		wallets:    make(map[string]WalletEntry),
	}
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

func (s *BotState) UpsertWallet(e WalletEntry) {
	s.mu.Lock()
	s.wallets[e.ID] = e
	s.mu.Unlock()
}

func (s *BotState) RemoveWallet(id string) {
	s.mu.Lock()
	delete(s.wallets, id)
	s.mu.Unlock()
}

func (s *BotState) Wallets() []WalletEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]WalletEntry, 0, len(s.wallets))
	for _, w := range s.wallets {
		out = append(out, w)
	}
	return out
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

// --- Navigation state ---

func (s *BotState) SetMenuMsgID(id int) {
	s.mu.Lock()
	s.menuMsgID = id
	s.mu.Unlock()
}

func (s *BotState) MenuMsgID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.menuMsgID
}

// --- Conversation state ---

func (s *BotState) SetPending(input, data string) {
	s.mu.Lock()
	s.pendingInput = input
	s.pendingData = data
	s.mu.Unlock()
}

func (s *BotState) ClearPending() {
	s.mu.Lock()
	s.pendingInput = ""
	s.pendingData = ""
	s.mu.Unlock()
}

func (s *BotState) Pending() (input, data string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingInput, s.pendingData
}

func (s *BotState) AddCopyTrade(line string) {
	s.mu.Lock()
	s.copyTrades = append([]string{line}, s.copyTrades...)
	if len(s.copyTrades) > 10 {
		s.copyTrades = s.copyTrades[:10]
	}
	s.mu.Unlock()
}

func (s *BotState) CopyTrades() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]string, len(s.copyTrades))
	copy(cp, s.copyTrades)
	return cp
}

// --- Markets navigation state ---

// --- Health state ---

func (s *BotState) SetHealth(snap health.HealthSnapshot) {
	s.mu.Lock()
	s.healthSnap = snap
	s.healthLoaded = true
	s.mu.Unlock()
}

func (s *BotState) Health() (health.HealthSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthSnap, s.healthLoaded
}

// SetViewMarkets stores the current market list snapshot (for index-based callbacks).
func (s *BotState) SetViewMarkets(markets []gamma.Market) {
	s.mu.Lock()
	s.viewMarkets = markets
	s.mu.Unlock()
}

// ViewMarket returns the market at index idx from the last SetViewMarkets snapshot.
func (s *BotState) ViewMarket(idx int) (gamma.Market, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if idx < 0 || idx >= len(s.viewMarkets) {
		return gamma.Market{}, false
	}
	return s.viewMarkets[idx], true
}
