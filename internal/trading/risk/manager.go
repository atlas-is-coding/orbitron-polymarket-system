// Package risk implements shared risk management for trading strategies.
package risk

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/atlasdev/orbitron/internal/config"
)

// Manager tracks daily P&L and enforces circuit breaker + stop-loss/take-profit.
type Manager struct {
	cfg       config.RiskConfig
	dailyLoss float64
	lastReset time.Time
	broken    atomic.Bool
	mu        sync.Mutex
}

// NewManager creates a RiskManager with the given config.
func NewManager(cfg config.RiskConfig) *Manager {
	return &Manager{
		cfg:       cfg,
		lastReset: time.Now(),
	}
}

// CanTrade returns false when the circuit breaker has tripped.
func (m *Manager) CanTrade() bool {
	return !m.broken.Load()
}

// RecordLoss adds usd to the daily loss accumulator.
// If the total exceeds MaxDailyLossUSD, the circuit breaker trips.
func (m *Manager) RecordLoss(usd float64) {
	m.mu.Lock()
	m.dailyLoss += usd
	tripped := m.dailyLoss >= m.cfg.MaxDailyLossUSD
	m.mu.Unlock()
	if tripped {
		m.broken.Store(true)
	}
}

// Reset clears the daily loss and resets the circuit breaker.
// Call at midnight or after manual review.
func (m *Manager) Reset() {
	m.mu.Lock()
	m.dailyLoss = 0
	m.lastReset = time.Now()
	m.mu.Unlock()
	m.broken.Store(false)
}

// DailyLossUSD returns the accumulated daily loss in USD.
func (m *Manager) DailyLossUSD() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dailyLoss
}

// LastReset returns the time of the last Reset() call.
func (m *Manager) LastReset() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastReset
}

// ShouldStopLoss returns true if the position has lost more than StopLossPct.
// entryPrice and currentPrice are in (0, 1).
func (m *Manager) ShouldStopLoss(entryPrice, currentPrice float64) bool {
	if entryPrice <= 0 {
		return false
	}
	lossPct := (entryPrice - currentPrice) / entryPrice * 100
	return lossPct >= m.cfg.StopLossPct
}

// ShouldTakeProfit returns true if the position has gained more than TakeProfitPct.
func (m *Manager) ShouldTakeProfit(entryPrice, currentPrice float64) bool {
	if entryPrice <= 0 {
		return false
	}
	gainPct := (currentPrice - entryPrice) / entryPrice * 100
	return gainPct >= m.cfg.TakeProfitPct
}
