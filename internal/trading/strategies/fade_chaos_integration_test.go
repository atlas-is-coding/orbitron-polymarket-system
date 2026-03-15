package strategies_test

import (
	"testing"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/trading/strategies"
)

func TestFadeTheChaosRealMarkets(t *testing.T) {
	httpClient := api.NewClient("https://gamma-api.polymarket.com", 10, 1)
	gammaClient := gamma.NewClient(httpClient)
	active := true
	markets, err := gammaClient.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 20})
	if err != nil {
		t.Skipf("gamma API unavailable: %v", err)
	}
	// Simulate two polls: first poll sets baseline, second detects spike
	// With real data we won't have spikes but we verify no panics + DetectSpike logic
	t.Log(markets)
	for _, mkt := range markets {
		if len(mkt.OutcomePrices) < 1 {
			continue
		}
		// DetectSpike with artificially high current price
		found, risePct := strategies.DetectSpike(0.40, 0.60, 10.0)
		if !found || risePct < 49 {
			t.Errorf("synthetic spike not detected: found=%v rise=%.2f", found, risePct)
		}
		break
	}
}
