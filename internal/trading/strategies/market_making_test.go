package strategies_test

import (
	"testing"

	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestCalcMarketMakingPrices(t *testing.T) {
	bid, ask := strategies.CalcMMPrices(0.50, 2.0) // midpoint 0.50, spread 2%
	// spread_pct/2 = 1% = 0.01 per side
	if bid >= 0.50 || bid <= 0 {
		t.Fatalf("expected bid < 0.50, got %.4f", bid)
	}
	if ask <= 0.50 || ask >= 1.0 {
		t.Fatalf("expected ask > 0.50, got %.4f", ask)
	}
	if ask-bid < 0.009 || ask-bid > 0.011 {
		t.Fatalf("expected spread ~0.01, got %.4f", ask-bid)
	}
}

func TestCalcMMPricesClampedBelow0(t *testing.T) {
	bid, ask := strategies.CalcMMPrices(0.005, 2.0) // very low midpoint
	if bid < 0.001 {
		t.Fatalf("bid too low: %.6f", bid)
	}
	if ask > 0.999 {
		t.Fatalf("ask too high: %.6f", ask)
	}
}

func TestCalcMMPricesAtBoundary(t *testing.T) {
	bid, ask := strategies.CalcMMPrices(0.95, 5.0)
	if bid <= 0 || ask >= 1.0 {
		t.Fatalf("prices out of range: bid=%.4f ask=%.4f", bid, ask)
	}
}
