package analytics

import (
	"fmt"
	"testing"
)

func TestRecordTrade_BatchTrigger(t *testing.T) {
	h := NewAnalyticsHub(3)

	h.RecordTrade(TradeReport{ID: "t1"})
	h.RecordTrade(TradeReport{ID: "t2"})

	// Trigger should not fire yet (2 < batch size 3)
	select {
	case <-h.Trigger():
		t.Fatal("trigger fired early")
	default:
	}

	h.RecordTrade(TradeReport{ID: "t3"})

	// Now it should fire
	select {
	case <-h.Trigger():
	default:
		t.Fatal("trigger did not fire after batch size reached")
	}
}

func TestFlush_ClearsBuffer(t *testing.T) {
	h := NewAnalyticsHub(10)
	h.RecordTrade(TradeReport{ID: "t1"})
	h.RecordTrade(TradeReport{ID: "t2"})

	got := h.Flush()
	if len(got) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(got))
	}
	if h.Size() != 0 {
		t.Fatalf("expected buffer empty after flush, got size %d", h.Size())
	}
}

func TestTagOrder_StrategyResolution(t *testing.T) {
	h := NewAnalyticsHub(10)

	h.TagOrder("order-1", "arbitrage")
	h.TagOrder("order-2", "market_making")

	got := h.StrategyForOrder("order-1")
	if got != "arbitrage" {
		t.Errorf("expected arbitrage, got %q", got)
	}
	// After lookup, the tag should be consumed (cleanup)
	got2 := h.StrategyForOrder("order-1")
	if got2 != "unknown" {
		t.Errorf("expected unknown after tag consumed, got %q", got2)
	}
}

func TestStrategyForOrder_UnknownFallback(t *testing.T) {
	h := NewAnalyticsHub(10)
	got := h.StrategyForOrder("nonexistent-order")
	if got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}

func TestTagOrder_EmptyInputsIgnored(t *testing.T) {
	h := NewAnalyticsHub(10)
	h.TagOrder("", "strategy") // no order id — should be ignored
	h.TagOrder("order-1", "") // no strategy — should be ignored

	got := h.StrategyForOrder("order-1")
	if got != "unknown" {
		t.Errorf("expected 'unknown' for empty strategy tag, got %q", got)
	}
}

func TestTagOrder_CapacityGuard(t *testing.T) {
	h := NewAnalyticsHub(0)

	for i := range maxOrderTags {
		h.TagOrder(fmt.Sprintf("order-%d", i), "strategy")
	}
	if len(h.orderStrategy) != maxOrderTags {
		t.Fatalf("expected map at capacity (%d), got %d", maxOrderTags, len(h.orderStrategy))
	}

	h.TagOrder("order-over-capacity", "strategy")

	got := h.StrategyForOrder("order-over-capacity")
	if got != "unknown" {
		t.Errorf("expected unknown for dropped tag, got %q", got)
	}
	if len(h.orderStrategy) != maxOrderTags {
		t.Errorf("expected map to stay at capacity, got size %d", len(h.orderStrategy))
	}
}

func TestHub_Concurrency(t *testing.T) {
	h := NewAnalyticsHub(0) // no batch trigger

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(n int) {
			h.RecordTrade(TradeReport{ID: "t"})
			h.TagOrder("order", "strat")
			_ = h.StrategyForOrder("order")
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// Just verifying no race condition — if the test completes without
	// -race detector errors, the locking is correct.
}
