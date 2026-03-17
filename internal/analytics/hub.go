package analytics

import (
	"sync"
)


type AnalyticsHub struct {
	mu        sync.Mutex
	trades    []TradeReport
	trigger   chan struct{}
	batchSize int
}

func NewAnalyticsHub(batchSize int) *AnalyticsHub {
	return &AnalyticsHub{
		trades:    make([]TradeReport, 0),
		trigger:   make(chan struct{}, 1),
		batchSize: batchSize,
	}
}

func (h *AnalyticsHub) RecordTrade(t TradeReport) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.trades = append(h.trades, t)
	if h.batchSize > 0 && len(h.trades) >= h.batchSize {
		select {
		case h.trigger <- struct{}{}:
		default:
		}
	}
}

func (h *AnalyticsHub) Trigger() <-chan struct{} {
	return h.trigger
}

func (h *AnalyticsHub) Size() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.trades)
}

func (h *AnalyticsHub) Flush() []TradeReport {
	h.mu.Lock()
	defer h.mu.Unlock()
	trades := h.trades
	h.trades = make([]TradeReport, 0)
	return trades
}
