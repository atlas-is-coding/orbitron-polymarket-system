package analytics

import (
	"sync"
)

type AnalyticsHub struct {
	mu     sync.Mutex
	trades []TradeReport
}

func NewAnalyticsHub() *AnalyticsHub {
	return &AnalyticsHub{
		trades: make([]TradeReport, 0),
	}
}

func (h *AnalyticsHub) RecordTrade(t TradeReport) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.trades = append(h.trades, t)
}

func (h *AnalyticsHub) Flush() []TradeReport {
	h.mu.Lock()
	defer h.mu.Unlock()
	trades := h.trades
	h.trades = make([]TradeReport, 0)
	return trades
}
