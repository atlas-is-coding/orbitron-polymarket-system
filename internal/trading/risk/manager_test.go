package risk_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/trading/risk"
)

func newTestManager(maxDailyLoss float64) *risk.Manager {
	cfg := config.RiskConfig{
		StopLossPct:     20.0,
		TakeProfitPct:   50.0,
		MaxDailyLossUSD: maxDailyLoss,
	}
	return risk.NewManager(cfg)
}

func TestCanTradeInitially(t *testing.T) {
	m := newTestManager(100.0)
	if !m.CanTrade() {
		t.Fatal("expected CanTrade() = true initially")
	}
}

func TestCircuitBreakerTrips(t *testing.T) {
	m := newTestManager(50.0)
	m.RecordLoss(60.0) // exceeds max
	if m.CanTrade() {
		t.Fatal("expected circuit breaker to trip after loss exceeds limit")
	}
}

func TestCircuitBreakerNotTripIfUnder(t *testing.T) {
	m := newTestManager(100.0)
	m.RecordLoss(30.0)
	m.RecordLoss(40.0) // total 70 < 100
	if !m.CanTrade() {
		t.Fatal("expected CanTrade() = true when loss under limit")
	}
}

func TestCircuitBreakerTripsCumulative(t *testing.T) {
	m := newTestManager(100.0)
	m.RecordLoss(60.0)
	m.RecordLoss(60.0) // total 120 > 100
	if m.CanTrade() {
		t.Fatal("expected circuit breaker to trip cumulatively")
	}
}

func TestResetRestoresTrade(t *testing.T) {
	m := newTestManager(50.0)
	m.RecordLoss(100.0)
	if m.CanTrade() {
		t.Fatal("expected circuit breaker tripped")
	}
	m.Reset()
	if !m.CanTrade() {
		t.Fatal("expected CanTrade() = true after reset")
	}
}

func TestShouldStopLoss(t *testing.T) {
	m := newTestManager(100.0) // stop_loss_pct = 20
	// entry 1.0, current 0.75 → loss = 25% > 20% → stop
	if !m.ShouldStopLoss(1.0, 0.75) {
		t.Fatal("expected ShouldStopLoss = true")
	}
	// entry 1.0, current 0.85 → loss = 15% < 20% → no stop
	if m.ShouldStopLoss(1.0, 0.85) {
		t.Fatal("expected ShouldStopLoss = false")
	}
}

func TestShouldTakeProfit(t *testing.T) {
	m := newTestManager(100.0) // take_profit_pct = 50
	// entry 0.50, current 0.80 → gain = 60% > 50% → take profit
	if !m.ShouldTakeProfit(0.50, 0.80) {
		t.Fatal("expected ShouldTakeProfit = true")
	}
	// entry 0.50, current 0.65 → gain = 30% < 50% → no
	if m.ShouldTakeProfit(0.50, 0.65) {
		t.Fatal("expected ShouldTakeProfit = false")
	}
}

func TestDailyLoss(t *testing.T) {
	m := newTestManager(100.0)
	m.RecordLoss(30.0)
	if m.DailyLossUSD() != 30.0 {
		t.Fatalf("expected DailyLossUSD = 30, got %.2f", m.DailyLossUSD())
	}
}

func TestLastResetTime(t *testing.T) {
	before := time.Now()
	m := newTestManager(100.0)
	after := time.Now()
	lr := m.LastReset()
	if lr.Before(before) || lr.After(after) {
		t.Fatal("LastReset should be near construction time")
	}
}
