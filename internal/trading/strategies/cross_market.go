// AI-generated strategy: CrossMarketStrategy.
// Detects price inconsistencies between logically related markets within the same event.
package strategies

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

// CrossMarketStrategy finds price inconsistencies within multi-outcome events.
// AI-generated: targets negRisk events where outcome prices don't sum to 1.0.
type CrossMarketStrategy struct {
	gamma    *gamma.Client
	executor Executor
	notifier notify.Notifier
	bus      *tui.EventBus
	risk     *risk.Manager
	cfg      config.CrossMarketConfig
	log      zerolog.Logger
	done     chan struct{}
	cooldown *CooldownTracker
}

// NewCrossMarketStrategy creates the AI-generated cross-market strategy.
func NewCrossMarketStrategy(
	gammaClient *gamma.Client,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.CrossMarketConfig,
	log zerolog.Logger,
) *CrossMarketStrategy {
	return &CrossMarketStrategy{
		gamma:    gammaClient,
		executor: executor,
		notifier: notifier,
		bus:      bus,
		risk:     riskMgr,
		cfg:      cfg,
		log:      log.With().Str("strategy", "cross_market").Logger(),
		done:     make(chan struct{}),
		cooldown: NewCooldownTracker(300),
	}
}

func (s *CrossMarketStrategy) Name() string { return "cross_market" }

func (s *CrossMarketStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("cross-market strategy started (AI-generated)")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.done:
			return nil
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

func (s *CrossMarketStrategy) Stop() error {
	close(s.done)
	return nil
}

func (s *CrossMarketStrategy) scan(ctx context.Context) {
	active := true
	events, err := s.gamma.GetEvents(gamma.EventsParams{Active: &active, Limit: 50})
	if err != nil {
		s.log.Warn().Err(err).Msg("cross_market: fetch events error")
		return
	}
	for _, ev := range events {
		// Collect YES prices for all negRisk markets in this event
		var negRiskPrices []float64
		var negRiskMarkets []gamma.Market
		for _, mkt := range ev.Markets {
			if !mkt.Active || mkt.Closed || !mkt.NegRisk {
				continue
			}
			if len(mkt.OutcomePrices) < 1 {
				continue
			}
			yes, err := strconv.ParseFloat(string(mkt.OutcomePrices[0]), 64)
			if err != nil || yes <= 0 {
				continue
			}
			negRiskPrices = append(negRiskPrices, yes)
			negRiskMarkets = append(negRiskMarkets, mkt)
		}
		if len(negRiskPrices) < 2 {
			continue
		}
		divPct := CalcNegRiskDivergence(negRiskPrices)
		if divPct < s.cfg.MinDivergencePct {
			continue
		}
		eventKey := ev.ID
		if s.cooldown.InCooldown(eventKey) {
			continue
		}
		s.cooldown.Record(eventKey)
		s.onDivergence(ctx, ev, negRiskMarkets, negRiskPrices, divPct)
	}
}

func (s *CrossMarketStrategy) onDivergence(ctx context.Context, ev gamma.Event, markets []gamma.Market, prices []float64, divPct float64) {
	sum := 0.0
	for _, p := range prices {
		sum += p
	}

	// If sum < 1.0: buy all outcomes (like arbitrage)
	// If sum > 1.0: sell most expensive outcome (fade overpriced side)
	signal := "BUY_YES"
	if sum > 1.0 {
		signal = "SELL"
	}

	executed := false
	if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() && sum < 1.0 {
		perMarket := s.cfg.MaxPositionUSD / float64(len(markets))
		for _, mkt := range markets {
			if len(mkt.ClobTokenIDs) < 1 {
				continue
			}
			_, err := s.executor.Open(string(mkt.ClobTokenIDs[0]), perMarket, true)
			if err != nil {
				s.log.Warn().Err(err).Str("market", mkt.ConditionID).Msg("cross_market: open failed")
			} else {
				executed = true
			}
		}
	}

	msg := fmt.Sprintf("[Cross-Market AI] Event: %s\n%d negRisk markets sum=%.3f divergence=%.1f%%",
		ev.Title, len(markets), sum, divPct)
	_ = s.notifier.Send(ctx, msg)
	if s.bus != nil && len(markets) > 0 {
		s.bus.Send(tui.StrategyAlertMsg{
			Strategy: "cross_market",
			Market:   markets[0].ConditionID,
			Question: fmt.Sprintf("Event: %s (%d markets)", ev.Title, len(markets)),
			Signal:   signal,
			Price:    sum / float64(len(prices)),
			EdgePct:  divPct,
			Reason:   fmt.Sprintf("negRisk sum=%.3f (expected 1.0), divergence=%.1f%%", sum, divPct),
			Executed: executed,
		})
	}
	s.log.Info().Str("event", ev.ID).Float64("sum", sum).Float64("div_pct", divPct).
		Msg("cross_market: divergence detected")
}

// CalcNegRiskDivergence returns |sum(prices) - 1.0| * 100 as a percentage.
// For a consistent negRisk event, prices should sum to exactly 1.0.
func CalcNegRiskDivergence(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	return math.Abs(sum-1.0) * 100
}
