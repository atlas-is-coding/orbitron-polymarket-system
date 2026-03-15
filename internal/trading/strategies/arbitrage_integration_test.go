//go:build integration

package strategies_test

import (
	"testing"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestArbitrageRealMarkets(t *testing.T) {
	httpClient := api.NewClient("https://gamma-api.polymarket.com", 10, 1)
	gammaClient := gamma.NewClient(httpClient)
	active := true
	markets, err := gammaClient.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 20})
	if err != nil {
		t.Skipf("gamma API unavailable: %v", err)
	}
	for _, mkt := range markets {
		found, profitPct := strategies.CheckArbitrageOpportunity(mkt)
		if found {
			t.Logf("arbitrage found: %s profit=%.4f", mkt.Question, profitPct)
		}
	}
	// No assertion — just verifying no panics and parsing works
}
