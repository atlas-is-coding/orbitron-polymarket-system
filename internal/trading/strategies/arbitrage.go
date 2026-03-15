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

// ArbitrageSignal carries details of a detected arbitrage opportunity.
type ArbitrageSignal struct {
	ConditionID string
	Question    string
	YesTokenID  string
	NoTokenID   string
	YesPrice    float64
	NoPrice     float64
	ProfitPct   float64
}

// ArbitrageStrategy buys YES+NO when their sum < $1.00.
type ArbitrageStrategy struct {
	gamma    *gamma.Client
	executor Executor // nil = signal-only
	notifier notify.Notifier
	bus      *tui.EventBus
	risk     *risk.Manager
	cfg      config.ArbitrageConfig
	log      zerolog.Logger
	done     chan struct{}
}

// NewArbitrageStrategy creates the strategy. executor may be nil for signal-only mode.
func NewArbitrageStrategy(
	gammaClient *gamma.Client,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.ArbitrageConfig,
	log zerolog.Logger,
) *ArbitrageStrategy {
	return &ArbitrageStrategy{
		gamma:    gammaClient,
		executor: executor,
		notifier: notifier,
		bus:      bus,
		risk:     riskMgr,
		cfg:      cfg,
		log:      log.With().Str("strategy", "arbitrage").Logger(),
		done:     make(chan struct{}),
	}
}

func (s *ArbitrageStrategy) Name() string { return "arbitrage" }

func (s *ArbitrageStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("arbitrage strategy started")
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

func (s *ArbitrageStrategy) Stop() error {
	close(s.done)
	return nil
}

func (s *ArbitrageStrategy) scan(ctx context.Context) {
	active := true
	markets, err := s.gamma.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 100})
	if err != nil {
		s.log.Warn().Err(err).Msg("failed to fetch markets")
		return
	}
	for _, mkt := range markets {
		found, profitPct := CheckArbitrageOpportunity(mkt)
		if !found {
			continue
		}
		profitUSD := profitPct * s.cfg.MaxPositionUSD
		if profitUSD < s.cfg.MinProfitUSD {
			continue
		}
		sig := ArbitrageSignal{
			ConditionID: mkt.ConditionID,
			Question:    mkt.Question,
			YesTokenID:  string(mkt.ClobTokenIDs[0]),
			NoTokenID:   string(mkt.ClobTokenIDs[1]),
			ProfitPct:   profitPct,
		}
		s.onSignal(ctx, sig)
	}
}

func (s *ArbitrageStrategy) onSignal(ctx context.Context, sig ArbitrageSignal) {
	executed := false
	orderID := ""

	if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() {
		half := s.cfg.MaxPositionUSD / 2
		resYes, errYes := s.executor.Open(sig.YesTokenID, half, false)
		resNo, errNo := s.executor.Open(sig.NoTokenID, half, false)
		if errYes == nil && errNo == nil {
			executed = true
			orderID = resYes.OrderID + "+" + resNo.OrderID
			s.log.Info().Str("market", sig.ConditionID).
				Float64("profit_pct", sig.ProfitPct).Msg("arbitrage executed")
		} else {
			if errYes != nil {
				s.log.Warn().Err(errYes).Msg("arb: open YES failed")
			}
			if errNo != nil {
				s.log.Warn().Err(errNo).Msg("arb: open NO failed")
			}
		}
	}

	msg := fmt.Sprintf("[Arbitrage] %s\nYES+NO = %.2f%% below $1 → profit %.2f%%\nMarket: %s",
		sig.Question, sig.ProfitPct*100, sig.ProfitPct*100, sig.ConditionID)
	if executed {
		msg += fmt.Sprintf("\n✅ Executed: %s", orderID)
	}

	_ = s.notifier.Send(ctx, msg)
	if s.bus != nil {
		s.bus.Send(tui.StrategyAlertMsg{
			Strategy: "arbitrage",
			Market:   sig.ConditionID,
			Question: sig.Question,
			Signal:   "BUY_YES+NO",
			Price:    sig.YesPrice,
			EdgePct:  sig.ProfitPct * 100,
			Reason:   fmt.Sprintf("YES+NO sum is %.4f (below 1.00)", 1-sig.ProfitPct),
			Executed: executed,
			OrderID:  orderID,
		})
	}
}

// CheckArbitrageOpportunity checks if a market's combined YES+NO price < 1.00.
// Returns (true, profitFraction) when opportunity exists. profitFraction = 1 - (yesPrice + noPrice).
func CheckArbitrageOpportunity(mkt gamma.Market) (bool, float64) {
	if !mkt.Active || mkt.Closed {
		return false, 0
	}
	if len(mkt.OutcomePrices) < 2 || len(mkt.ClobTokenIDs) < 2 {
		return false, 0
	}
	yes, errY := strconv.ParseFloat(string(mkt.OutcomePrices[0]), 64)
	no, errN := strconv.ParseFloat(string(mkt.OutcomePrices[1]), 64)
	if errY != nil || errN != nil || yes <= 0 || no <= 0 {
		return false, 0
	}
	sum := yes + no
	if sum >= 1.0 {
		return false, 0
	}
	return true, 1.0 - sum
}
