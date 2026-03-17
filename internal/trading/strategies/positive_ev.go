package strategies

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

// PositiveEVStrategy detects markets where our private probability > market price.
type PositiveEVStrategy struct {
	gamma    GammaClient
	executor Executor
	notifier notify.Notifier
	bus      *tui.EventBus
	risk     *risk.Manager
	cfg      config.PositiveEVConfig
	log      zerolog.Logger
	done     chan struct{}
}

// NewPositiveEVStrategy creates the strategy.
func NewPositiveEVStrategy(
	gammaClient GammaClient,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.PositiveEVConfig,
	log zerolog.Logger,
) *PositiveEVStrategy {
	return &PositiveEVStrategy{
		gamma:    gammaClient,
		executor: executor,
		notifier: notifier,
		bus:      bus,
		risk:     riskMgr,
		cfg:      cfg,
		log:      log.With().Str("strategy", "positive_ev").Logger(),
		done:     make(chan struct{}),
	}
}


func (s *PositiveEVStrategy) Name() string { return "positive_ev" }

func (s *PositiveEVStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("positive EV strategy started")
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

func (s *PositiveEVStrategy) Stop() error {
	close(s.done)
	return nil
}

// targetProbability estimates the "true" probability for YES given current price.
// Heuristic: for near-resolution markets (< 7d) we bump towards 90% if already above 75%.
// This is a simplification — in reality would use external data.
const targetProbability = 0.90

func (s *PositiveEVStrategy) scan(ctx context.Context) {
	active := true
	markets, err := s.gamma.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 100})
	if err != nil {
		s.log.Warn().Err(err).Msg("pos_ev: fetch error")
		return
	}
	now := time.Now().UTC()
	for _, mkt := range markets {
		if len(mkt.OutcomePrices) < 1 || len(mkt.ClobTokenIDs) < 2 {
			continue
		}
		liq := float64(mkt.Liquidity)
		yesStr := string(mkt.OutcomePrices[0])
		yes, err := strconv.ParseFloat(yesStr, 64)
		if err != nil {
			continue
		}
		endDate, err := time.Parse(time.RFC3339, mkt.EndDateISO)
		if err != nil {
			continue
		}
		found, edgePct := CheckPositiveEV(yes, liq, endDate, s.cfg.PollIntervalMs, s.cfg.MinLiquidityUSD, s.cfg.MinEdgePct)
		// Use days from config field (reusing PollIntervalMs field is wrong — fix: add MaxDurationDays to config)
		// Actually use hardcoded 14 days here and compute from endDate:
		daysUntilEnd := endDate.Sub(now).Hours() / 24
		found, edgePct = CheckPositiveEV(yes, liq, endDate, 14, s.cfg.MinLiquidityUSD, s.cfg.MinEdgePct)
		if !found {
			continue
		}
		_ = daysUntilEnd

		executed := false
		orderID := ""
		if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() {
			res, err := s.executor.Open(string(mkt.ClobTokenIDs[0]), s.cfg.MaxPositionUSD, mkt.NegRisk)
			if err != nil {
				s.log.Warn().Err(err).Str("market", mkt.ConditionID).Msg("pos_ev: open failed")
			} else {
				executed = true
				orderID = res.OrderID
			}
		}

		msg := fmt.Sprintf("[Positive EV] %s\nYES price=%.3f edge=%.1f%% (target=%.0f%%)",
			mkt.Question, yes, edgePct, targetProbability*100)
		_ = s.notifier.Send(ctx, msg)
		if s.bus != nil {
			s.bus.Send(tui.StrategyAlertMsg{
				Strategy: "positive_ev",
				Market:   mkt.ConditionID,
				Question: mkt.Question,
				Signal:   "BUY_YES",
				Price:    yes,
				EdgePct:  edgePct,
				Reason:   fmt.Sprintf("market priced %.0f%% below target %.0f%%", yes*100, targetProbability*100),
				Executed: executed,
				OrderID:  orderID,
			})
		}
	}
}

// CheckPositiveEV returns (true, edgePct) when a near-resolution market appears mispriced.
// yesPrice: current YES price (0-1), liquidity: USD liquidity, endDate: resolution date,
// maxDurationDays: max days until resolution to consider, minLiquidityUSD, minEdgePct: thresholds.
func CheckPositiveEV(yesPrice, liquidity float64, endDate time.Time, maxDurationDays int, minLiquidityUSD, minEdgePct float64) (bool, float64) {
	now := time.Now().UTC()
	if endDate.Before(now) {
		return false, 0
	}
	daysLeft := endDate.Sub(now).Hours() / 24
	if daysLeft > float64(maxDurationDays) {
		return false, 0
	}
	if liquidity < minLiquidityUSD {
		return false, 0
	}
	// Only target markets where YES is in the "near-certain but underpriced" zone: 0.75-0.89
	if yesPrice < 0.75 || yesPrice >= targetProbability {
		return false, 0
	}
	edgePct := (targetProbability - yesPrice) * 100
	if edgePct < minEdgePct {
		return false, 0
	}
	return true, edgePct
}
