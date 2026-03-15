package strategies_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestRisklessRateDetects(t *testing.T) {
	// Long duration market, YES very low (absurd event), NO below 0.97 → opportunity
	endFar := time.Now().UTC().Add(90 * 24 * time.Hour)
	found := strategies.CheckRisklessRate(0.03, endFar, 30, 0.05)
	if !found {
		t.Fatal("expected riskless rate opportunity: YES=0.03, duration=90d")
	}
}

func TestRisklessRateNotTooShort(t *testing.T) {
	// Market ends in 10 days — not long enough
	endSoon := time.Now().UTC().Add(10 * 24 * time.Hour)
	found := strategies.CheckRisklessRate(0.03, endSoon, 30, 0.05)
	if found {
		t.Fatal("expected no opportunity: duration < min_duration_days")
	}
}

func TestRisklessRateYESPriceToHigh(t *testing.T) {
	// YES = 0.10 → NO = 0.90 → above max_no_price=0.05, but YES too high to be "riskless"
	// Only trigger when YES < max_no_price (e.g., 0.03)
	endFar := time.Now().UTC().Add(90 * 24 * time.Hour)
	found := strategies.CheckRisklessRate(0.10, endFar, 30, 0.05)
	if found {
		t.Fatal("expected no opportunity: YES price too high for riskless classification")
	}
}

func TestRisklessRateAlreadyExpired(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	found := strategies.CheckRisklessRate(0.02, past, 30, 0.05)
	if found {
		t.Fatal("expected no opportunity: market already expired")
	}
}
