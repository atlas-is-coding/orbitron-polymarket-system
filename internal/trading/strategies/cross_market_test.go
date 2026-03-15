package strategies_test

import (
	"testing"

	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestCrossMarketNegRiskSum(t *testing.T) {
	// 3 outcomes in a negRisk event should sum to ~1.0
	// If sum = 0.90, divergence = 10%
	prices := []float64{0.30, 0.30, 0.30} // sum = 0.90
	div := strategies.CalcNegRiskDivergence(prices)
	if div < 9.9 || div > 10.1 {
		t.Fatalf("expected divergence ~10%%, got %.2f", div)
	}
}

func TestCrossMarketNegRiskNoOpportunity(t *testing.T) {
	// sum = 1.00 → no divergence
	prices := []float64{0.50, 0.30, 0.20}
	div := strategies.CalcNegRiskDivergence(prices)
	if div > 0.1 {
		t.Fatalf("expected ~0%% divergence, got %.2f", div)
	}
}

func TestCrossMarketNegRiskOverpriced(t *testing.T) {
	// sum = 1.10 → overpriced
	prices := []float64{0.40, 0.40, 0.30}
	div := strategies.CalcNegRiskDivergence(prices)
	// divergence is |sum - 1| * 100
	if div < 9.9 || div > 10.1 {
		t.Fatalf("expected divergence ~10%%, got %.2f", div)
	}
}

func TestCrossMarketEmptyPrices(t *testing.T) {
	div := strategies.CalcNegRiskDivergence(nil)
	if div != 0 {
		t.Fatal("expected 0 divergence for empty slice")
	}
}
