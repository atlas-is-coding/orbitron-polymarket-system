package analytics

import (
	"sync"
)

type AnalyticsHub struct {
	mu        sync.Mutex
	trades    []TradeReport
	trigger   chan struct{}
	batchSize int
	// orderStrategy maps order ID → strategy name, populated when an order is placed.
	// Used to enrich trade reports with the strategy that generated the underlying order.
	orderStrategy map[string]string
}

func NewAnalyticsHub(batchSize int) *AnalyticsHub {
	return &AnalyticsHub{
		trades:        make([]TradeReport, 0),
		trigger:       make(chan struct{}, 1),
		batchSize:     batchSize,
		orderStrategy: make(map[string]string),
	}
}

// maxOrderTags caps the orderStrategy map to prevent unbounded memory growth
// from cancelled orders that are tagged but never filled.
const maxOrderTags = 10_000

// TagOrder associates an order ID with the strategy that placed it.
// Call this immediately after a successful order placement.
// Silently drops the tag if the map is at capacity.
func (h *AnalyticsHub) TagOrder(orderID, strategy string) {
	if orderID == "" || strategy == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.orderStrategy) >= maxOrderTags {
		return
	}
	h.orderStrategy[orderID] = strategy
}

// StrategyForOrder returns the strategy name for the given order ID, or
// "unknown" if no tag was recorded.
func (h *AnalyticsHub) StrategyForOrder(orderID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.orderStrategy[orderID]; ok {
		// Clean up to avoid unbounded growth
		delete(h.orderStrategy, orderID)
		return s
	}
	return "unknown"
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
