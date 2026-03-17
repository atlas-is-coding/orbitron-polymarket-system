package strategies

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

// CooldownTracker prevents repeat alerts on the same market within cooldownSec.
type CooldownTracker struct {
	cooldown time.Duration
	last     map[string]time.Time
	mu       sync.Mutex
}

// NewCooldownTracker creates a tracker with the given cooldown in seconds.
func NewCooldownTracker(cooldownSec int) *CooldownTracker {
	return &CooldownTracker{
		cooldown: time.Duration(cooldownSec) * time.Second,
		last:     make(map[string]time.Time),
	}
}

// InCooldown returns true if the market was alerted within the cooldown window.
func (c *CooldownTracker) InCooldown(conditionID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.last[conditionID]
	return ok && time.Since(t) < c.cooldown
}

// Record marks the market as alerted now.
func (c *CooldownTracker) Record(conditionID string) {
	c.mu.Lock()
	c.last[conditionID] = time.Now()
	c.mu.Unlock()
}

// FadeTheChaosStrategy detects emotional YES price spikes and signals contrarian NO positions.
type FadeTheChaosStrategy struct {
	gamma    GammaClient
	executor Executor
	notifier notify.Notifier
	bus      *tui.EventBus
	risk     *risk.Manager
	cfg      config.FadeChaosConfig
	log      zerolog.Logger
	done     chan struct{}
	// previous poll prices: conditionID → yesPrice
	prevPrices map[string]float64
	priceMu    sync.Mutex
	cooldown   *CooldownTracker
}

// NewFadeTheChaosStrategy creates the strategy.
func NewFadeTheChaosStrategy(
	gammaClient GammaClient,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.FadeChaosConfig,
	log zerolog.Logger,
) *FadeTheChaosStrategy {
	return &FadeTheChaosStrategy{
		gamma:      gammaClient,
		executor:   executor,
		notifier:   notifier,
		bus:        bus,
		risk:       riskMgr,
		cfg:        cfg,
		log:        log.With().Str("strategy", "fade_chaos").Logger(),
		done:       make(chan struct{}),
		prevPrices: make(map[string]float64),
		cooldown:   NewCooldownTracker(cfg.CooldownSec),
	}
}

func (s *FadeTheChaosStrategy) Name() string { return "fade_chaos" }

func (s *FadeTheChaosStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("fade the chaos strategy started")
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

func (s *FadeTheChaosStrategy) Stop() error {
	close(s.done)
	return nil
}

func (s *FadeTheChaosStrategy) scan(ctx context.Context) {
	active := true
	markets, err := s.gamma.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 100})
	if err != nil {
		s.log.Warn().Err(err).Msg("fade_chaos: fetch error")
		return
	}
	s.priceMu.Lock()
	defer s.priceMu.Unlock()

	for _, mkt := range markets {
		if len(mkt.OutcomePrices) < 2 || len(mkt.ClobTokenIDs) < 2 {
			continue
		}
		yes, err := strconv.ParseFloat(string(mkt.OutcomePrices[0]), 64)
		if err != nil {
			continue
		}
		prev, hasPrev := s.prevPrices[mkt.ConditionID]
		s.prevPrices[mkt.ConditionID] = yes

		if !hasPrev {
			continue // first poll, no baseline
		}
		if s.cooldown.InCooldown(mkt.ConditionID) {
			continue
		}
		found, risePct := DetectSpike(prev, yes, s.cfg.SpikeThresholdPct)
		if !found {
			continue
		}
		s.cooldown.Record(mkt.ConditionID)
		s.onSpike(ctx, mkt, yes, prev, risePct)
	}
}

func (s *FadeTheChaosStrategy) onSpike(ctx context.Context, mkt gamma.Market, yesNow, yesPrev, risePct float64) {
	executed := false
	orderID := ""
	noTokenID := string(mkt.ClobTokenIDs[1])
	noPrice := 1.0 - yesNow

	if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() {
		res, err := s.executor.Open(noTokenID, s.cfg.MaxPositionUSD, mkt.NegRisk)
		if err != nil {
			s.log.Warn().Err(err).Str("market", mkt.ConditionID).Msg("fade_chaos: open NO failed")
		} else {
			executed = true
			orderID = res.OrderID
		}
	}

	msg := fmt.Sprintf("[Fade the Chaos] %s\nYES spiked %.1f%% (%.3f→%.3f) → buying NO @ %.3f",
		mkt.Question, risePct, yesPrev, yesNow, noPrice)
	_ = s.notifier.Send(ctx, msg)
	if s.bus != nil {
		s.bus.Send(tui.StrategyAlertMsg{
			Strategy: "fade_chaos",
			Market:   mkt.ConditionID,
			Question: mkt.Question,
			Signal:   "BUY_NO",
			Price:    noPrice,
			EdgePct:  risePct,
			Reason:   fmt.Sprintf("YES spiked %.1f%% (%.3f→%.3f), fading emotional move", risePct, yesPrev, yesNow),
			Executed: executed,
			OrderID:  orderID,
		})
	}
	s.log.Info().Str("market", mkt.ConditionID).
		Float64("prev_yes", yesPrev).Float64("curr_yes", yesNow).
		Float64("rise_pct", risePct).Msg("fade_chaos: spike detected")
}

// DetectSpike returns (true, risePct) when yesNow is more than thresholdPct% higher than yesPrev.
// risePct is the percentage increase relative to yesPrev.
func DetectSpike(yesPrev, yesNow, thresholdPct float64) (bool, float64) {
	if yesPrev <= 0 {
		return false, 0
	}
	if yesNow <= yesPrev {
		return false, 0
	}
	risePct := (yesNow - yesPrev) / yesPrev * 100
	return risePct >= thresholdPct, risePct
}
