package webui

import (
	"maps"
	"sync"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/health"
	"github.com/atlasdev/orbitron/internal/tui"
)

const maxLogs = 200

// LogEntry is a single log line with level.
type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// WalletEntry holds the latest stats snapshot for one wallet.
type WalletEntry struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Enabled     bool    `json:"enabled"`
	Primary     bool    `json:"primary"`
	BalanceUSD  float64 `json:"balance_usd"`
	PnLUSD      float64 `json:"pnl_usd"`
	OpenOrders  int     `json:"open_orders"`
	TotalTrades int     `json:"total_trades"`
}

// WebState is a thread-safe snapshot of bot data for the web panel.
type WebState struct {
        mu         sync.RWMutex
        balance    float64
        orders     []tui.OrderRow
        positions  []tui.PositionRow
        traders    []tui.TraderRow
        strategies []tui.StrategyRow
        logs       []LogEntry
        subsystems map[string]bool
        wallets    map[string]*WalletEntry
        cfg        *config.Config
        health     health.HealthSnapshot
}

func newWebState() *WebState {
        return &WebState{
                subsystems: make(map[string]bool),
                wallets:    make(map[string]*WalletEntry),
                strategies: make([]tui.StrategyRow, 0),
        }
}

func (s *WebState) SetBalance(v float64) {
        s.mu.Lock()
        s.balance = v
        s.mu.Unlock()
}

func (s *WebState) Balance() float64 {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.balance
}

func (s *WebState) SetOrders(rows []tui.OrderRow) {
        s.mu.Lock()
        s.orders = rows
        s.mu.Unlock()
}

func (s *WebState) Orders() []tui.OrderRow {
        s.mu.RLock()
        defer s.mu.RUnlock()
        cp := make([]tui.OrderRow, len(s.orders))
        copy(cp, s.orders)
        return cp
}

func (s *WebState) SetPositions(rows []tui.PositionRow) {
        s.mu.Lock()
        s.positions = rows
        s.mu.Unlock()
}

func (s *WebState) Positions() []tui.PositionRow {
        s.mu.RLock()
        defer s.mu.RUnlock()
        cp := make([]tui.PositionRow, len(s.positions))
        copy(cp, s.positions)
        return cp
}

func (s *WebState) SetTraders(rows []tui.TraderRow) {
        s.mu.Lock()
        s.traders = rows
        s.mu.Unlock()
}

func (s *WebState) Traders() []tui.TraderRow {
        s.mu.RLock()
        defer s.mu.RUnlock()
        cp := make([]tui.TraderRow, len(s.traders))
        copy(cp, s.traders)
        return cp
}

func (s *WebState) SetStrategies(rows []tui.StrategyRow) {
        s.mu.Lock()
        s.strategies = rows
        s.mu.Unlock()
}

func (s *WebState) Strategies() []tui.StrategyRow {
        s.mu.RLock()
        defer s.mu.RUnlock()
        cp := make([]tui.StrategyRow, len(s.strategies))
        copy(cp, s.strategies)
        return cp
}
func (s *WebState) AddLog(level, msg string) {
	s.mu.Lock()
	s.logs = append(s.logs, LogEntry{Level: level, Message: msg})
	if len(s.logs) > maxLogs {
		s.logs = s.logs[len(s.logs)-maxLogs:]
	}
	s.mu.Unlock()
}

func (s *WebState) Logs() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]LogEntry, len(s.logs))
	copy(cp, s.logs)
	return cp
}

func (s *WebState) SetSubsystem(name string, active bool) {
	s.mu.Lock()
	s.subsystems[name] = active
	s.mu.Unlock()
}

func (s *WebState) Subsystems() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]bool, len(s.subsystems))
	maps.Copy(cp, s.subsystems)
	return cp
}

func (s *WebState) UpsertWallet(e WalletEntry) {
	s.mu.Lock()
	s.wallets[e.ID] = &e
	s.mu.Unlock()
}

func (s *WebState) RemoveWallet(id string) {
	s.mu.Lock()
	delete(s.wallets, id)
	s.mu.Unlock()
}

func (s *WebState) Wallets() []WalletEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]WalletEntry, 0, len(s.wallets))
	for _, w := range s.wallets {
		out = append(out, *w)
	}
	return out
}

func (s *WebState) SetConfig(cfg *config.Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

func (s *WebState) Config() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *WebState) SetHealth(snap health.HealthSnapshot) {
	s.mu.Lock()
	s.health = snap
	s.mu.Unlock()
}

func (s *WebState) GetHealth() health.HealthSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.health
}
