package strategies_test

import (
	"fmt"
	"testing"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

// mockGammaForArb returns markets with a given set of YES prices.
func marketWithPrices(condID, question string, yesPrice, noPrice float64) gamma.Market {
	yesStr := fmt.Sprintf("%.4f", yesPrice)
	noStr := fmt.Sprintf("%.4f", noPrice)
	return gamma.Market{
		ConditionID:   condID,
		Question:      question,
		Active:        true,
		ClobTokenIDs:  []string{"111", "222"},
		OutcomePrices: []string{yesStr, noStr},
		Liquidity:     1000,
	}
}

func TestArbitrageDetectOpportunity(t *testing.T) {
	detected := make(chan strategies.ArbitrageSignal, 1)
	cfg := config.ArbitrageConfig{
		MinProfitUSD:   0.01,
		MaxPositionUSD: 100.0,
		ExecuteOrders:  false,
	}
	riskMgr := risk.NewManager(config.RiskConfig{MaxDailyLossUSD: 100})

	// YES=0.45, NO=0.45 → sum=0.90 < 1.00, profit=0.10 per $1
	mkt := marketWithPrices("cid1", "Will X happen?", 0.45, 0.45)

	found, profit := strategies.CheckArbitrageOpportunity(mkt)
	if !found {
		t.Fatal("expected arbitrage opportunity detected")
	}
	if profit < 0.09 || profit > 0.11 {
		t.Fatalf("expected profit ~0.10, got %.4f", profit)
	}
	_ = detected
	_ = cfg
	_ = riskMgr
}

func TestArbitrageNoOpportunityWhenSumAbove1(t *testing.T) {
	mkt := marketWithPrices("cid2", "Will Y happen?", 0.52, 0.50)
	found, _ := strategies.CheckArbitrageOpportunity(mkt)
	if found {
		t.Fatal("expected no arbitrage when sum > 1.00")
	}
}

func TestArbitrageNoOpportunityWhenSumEquals1(t *testing.T) {
	mkt := marketWithPrices("cid3", "Will Z happen?", 0.50, 0.50)
	found, _ := strategies.CheckArbitrageOpportunity(mkt)
	if found {
		t.Fatal("expected no arbitrage when sum = 1.00")
	}
}

func TestArbitrageCircuitBreakerPreventsExecution(t *testing.T) {
	riskMgr := risk.NewManager(config.RiskConfig{MaxDailyLossUSD: 10})
	riskMgr.RecordLoss(20) // trip breaker
	if riskMgr.CanTrade() {
		t.Fatal("expected circuit breaker tripped")
	}
}
