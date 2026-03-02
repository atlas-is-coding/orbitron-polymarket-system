// Package wallet manages runtime state for multiple wallet instances.
package wallet

import "sync"

// WalletStats holds the in-memory cached statistics for one wallet.
// All fields are unexported; access is only through Set/Get under the mutex.
type WalletStats struct {
	mu          sync.RWMutex
	balanceUSD  float64
	pnlUSD      float64
	openOrders  int
	totalTrades int
}

// Set updates all stats atomically.
func (s *WalletStats) Set(balance, pnl float64, openOrders, totalTrades int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balanceUSD = balance
	s.pnlUSD = pnl
	s.openOrders = openOrders
	s.totalTrades = totalTrades
}

// Get returns a snapshot of all stats.
func (s *WalletStats) Get() (balance, pnl float64, openOrders, totalTrades int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balanceUSD, s.pnlUSD, s.openOrders, s.totalTrades
}
