package strategies_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func endDate(daysFromNow int) time.Time {
	return time.Now().UTC().Add(time.Duration(daysFromNow) * 24 * time.Hour)
}

func TestPositiveEVDetects(t *testing.T) {
	found, edge := strategies.CheckPositiveEV(0.80, 10000, endDate(5), 14, 5000, 5.0)
	if !found {
		t.Fatal("expected opportunity")
	}
	if edge < 5.0 {
		t.Fatalf("edge too low: %.2f", edge)
	}
}

func TestPositiveEVTooFarOut(t *testing.T) {
	found, _ := strategies.CheckPositiveEV(0.80, 10000, endDate(30), 14, 5000, 5.0)
	if found {
		t.Fatal("expected no opportunity: too far out")
	}
}

func TestPositiveEVLowLiquidity(t *testing.T) {
	found, _ := strategies.CheckPositiveEV(0.80, 1000, endDate(5), 14, 5000, 5.0)
	if found {
		t.Fatal("expected no opportunity: low liquidity")
	}
}

func TestPositiveEVPriceAlreadyAtTarget(t *testing.T) {
	found, _ := strategies.CheckPositiveEV(0.93, 10000, endDate(5), 14, 5000, 5.0)
	if found {
		t.Fatal("expected no opportunity: already at target")
	}
}

func TestPositiveEVPriceTooLow(t *testing.T) {
	// Below 0.75 — not "near certain" enough
	found, _ := strategies.CheckPositiveEV(0.60, 10000, endDate(5), 14, 5000, 5.0)
	if found {
		t.Fatal("expected no opportunity: price too low for near-certain heuristic")
	}
}
