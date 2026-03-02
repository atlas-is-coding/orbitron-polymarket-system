// Package wallet manages runtime state for multiple wallet instances.
package wallet

import "sync"

// WalletStats holds the in-memory cached statistics for one wallet.
// Thread-safe via RWMutex.
type WalletStats struct {
	mu          sync.RWMutex
	BalanceUSD  float64
	PnLUSD      float64
	OpenOrders  int
	TotalTrades int
}

// Set updates all stats atomically.
func (s *WalletStats) Set(balance, pnl float64, openOrders, totalTrades int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BalanceUSD = balance
	s.PnLUSD = pnl
	s.OpenOrders = openOrders
	s.TotalTrades = totalTrades
}

// Get returns a snapshot of all stats.
func (s *WalletStats) Get() (balance, pnl float64, openOrders, totalTrades int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BalanceUSD, s.PnLUSD, s.OpenOrders, s.TotalTrades
}
