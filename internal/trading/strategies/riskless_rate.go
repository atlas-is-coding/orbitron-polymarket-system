package strategies

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/atlasdev/polytrade-bot/internal/trading/risk"
	"github.com/atlasdev/polytrade-bot/internal/tui"
	"github.com/rs/zerolog"
)

// RisklessRateStrategy buys NO on long-duration markets with very low YES probability.
type RisklessRateStrategy struct {
	gamma    *gamma.Client
	executor Executor
	notifier notify.Notifier
	bus      *tui.EventBus
	risk     *risk.Manager
	cfg      config.RisklessRateConfig
	log      zerolog.Logger
	done     chan struct{}
}

func NewRisklessRateStrategy(
	gammaClient *gamma.Client,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.RisklessRateConfig,
	log zerolog.Logger,
) *RisklessRateStrategy {
	return &RisklessRateStrategy{
		gamma:    gammaClient,
		executor: executor,
		notifier: notifier,
		bus:      bus,
		risk:     riskMgr,
		cfg:      cfg,
		log:      log.With().Str("strategy", "riskless_rate").Logger(),
		done:     make(chan struct{}),
	}
}

func (s *RisklessRateStrategy) Name() string { return "riskless_rate" }

func (s *RisklessRateStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("riskless rate strategy started")
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

func (s *RisklessRateStrategy) Stop() error {
	close(s.done)
	return nil
}

func (s *RisklessRateStrategy) scan(ctx context.Context) {
	active := true
	markets, err := s.gamma.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 100})
	if err != nil {
		s.log.Warn().Err(err).Msg("riskless_rate: fetch error")
		return
	}
	for _, mkt := range markets {
		if len(mkt.OutcomePrices) < 2 || len(mkt.ClobTokenIDs) < 2 {
			continue
		}
		yes, err := strconv.ParseFloat(string(mkt.OutcomePrices[0]), 64)
		if err != nil {
			continue
		}
		endDate, err := time.Parse(time.RFC3339, mkt.EndDateISO)
		if err != nil {
			continue
		}
		if !CheckRisklessRate(yes, endDate, s.cfg.MinDurationDays, s.cfg.MaxNOPrice) {
			continue
		}
		noPrice := 1.0 - yes
		executed := false
		orderID := ""
		if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() {
			// BUY NO = buy the NO token (index 1)
			res, err := s.executor.Open(string(mkt.ClobTokenIDs[1]), s.cfg.MaxPositionUSD, mkt.NegRisk)
			if err != nil {
				s.log.Warn().Err(err).Str("market", mkt.ConditionID).Msg("riskless_rate: open NO failed")
			} else {
				executed = true
				orderID = res.OrderID
			}
		}
		daysLeft := int(endDate.Sub(time.Now().UTC()).Hours() / 24)
		msg := fmt.Sprintf("[Riskless Rate] %s\nNO=%.3f YES=%.3f duration=%dd\nTime-value distortion opportunity",
			mkt.Question, noPrice, yes, daysLeft)
		_ = s.notifier.Send(ctx, msg)
		if s.bus != nil {
			s.bus.Send(tui.StrategyAlertMsg{
				Strategy: "riskless_rate",
				Market:   mkt.ConditionID,
				Question: mkt.Question,
				Signal:   "BUY_NO",
				Price:    noPrice,
				EdgePct:  (s.cfg.MaxNOPrice - yes) * 100,
				Reason:   fmt.Sprintf("YES=%.3f in %d-day market (time-value distortion)", yes, daysLeft),
				Executed: executed,
				OrderID:  orderID,
			})
		}
	}
}

// CheckRisklessRate returns true when a market qualifies for riskless rate strategy:
// - endDate is at least minDurationDays in the future
// - yesPrice < maxNOPrice (e.g., < 0.05 = absurd event)
func CheckRisklessRate(yesPrice float64, endDate time.Time, minDurationDays int, maxNOPrice float64) bool {
	now := time.Now().UTC()
	if endDate.Before(now) {
		return false
	}
	daysLeft := endDate.Sub(now).Hours() / 24
	if daysLeft < float64(minDurationDays) {
		return false
	}
	// YES price must be very low (absurd/impossible event)
	return yesPrice < maxNOPrice
}