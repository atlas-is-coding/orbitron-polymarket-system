package strategies

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/trading/risk"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

// MarketMakingStrategy posts bid+ask limit orders around the midpoint.
type MarketMakingStrategy struct {
	gamma     GammaClient
	clob      *clob.Client
	executor  Executor
	notifier  notify.Notifier
	bus       *tui.EventBus
	risk      *risk.Manager
	cfg       config.MarketMakingConfig
	log       zerolog.Logger
	done      chan struct{}
	// active orders: tokenID → orderID
	bidOrders map[string]string
	askOrders map[string]string
}

// NewMarketMakingStrategy creates the strategy.
func NewMarketMakingStrategy(
	gammaClient GammaClient,
	clobClient *clob.Client,
	executor Executor,
	notifier notify.Notifier,
	bus *tui.EventBus,
	riskMgr *risk.Manager,
	cfg config.MarketMakingConfig,
	log zerolog.Logger,
) *MarketMakingStrategy {
	return &MarketMakingStrategy{
		gamma:     gammaClient,
		clob:      clobClient,
		executor:  executor,
		notifier:  notifier,
		bus:       bus,
		risk:      riskMgr,
		cfg:       cfg,
		log:       log.With().Str("strategy", "market_making").Logger(),
		done:      make(chan struct{}),
		bidOrders: make(map[string]string),
		askOrders: make(map[string]string),
	}
}

func (s *MarketMakingStrategy) Name() string { return "market_making" }

func (s *MarketMakingStrategy) Start(ctx context.Context) error {
	interval := time.Duration(s.cfg.RebalanceIntervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", interval).Msg("market making strategy started")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.done:
			return nil
		case <-ticker.C:
			s.rebalance(ctx)
		}
	}
}

func (s *MarketMakingStrategy) Stop() error {
	close(s.done)
	return nil
}

func (s *MarketMakingStrategy) rebalance(ctx context.Context) {
	active := true
	markets, err := s.gamma.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 50})
	if err != nil {
		s.log.Warn().Err(err).Msg("mm: failed to fetch markets")
		return
	}
	for _, mkt := range markets {
		if float64(mkt.Liquidity) < s.cfg.MinLiquidityUSD {
			continue
		}
		if len(mkt.OutcomePrices) < 1 || len(mkt.ClobTokenIDs) < 2 {
			continue
		}
		midStr := string(mkt.OutcomePrices[0])
		mid, err := strconv.ParseFloat(midStr, 64)
		if err != nil || mid <= 0 || mid >= 1 {
			continue
		}
		bid, ask := CalcMMPrices(mid, s.cfg.SpreadPct)
		tokenYes := string(mkt.ClobTokenIDs[0])

		if s.cfg.ExecuteOrders && s.executor != nil && s.risk.CanTrade() {
			half := s.cfg.MaxPositionUSD / 2
			bidID, errB := s.executor.PlaceLimit(tokenYes, "YES", "GTC", bid, half)
			askID, errA := s.executor.PlaceLimit(tokenYes, "NO", "GTC", ask, half)
			if errB == nil {
				s.bidOrders[tokenYes] = bidID
			}
			if errA == nil {
				s.askOrders[tokenYes] = askID
			}
		}

		msg := fmt.Sprintf("[Market Making] %s\nbid=%.4f ask=%.4f spread=%.2f%%",
			mkt.Question, bid, ask, s.cfg.SpreadPct)
		_ = s.notifier.Send(ctx, msg)
		if s.bus != nil {
			s.bus.Send(tui.StrategyAlertMsg{
				Strategy: "market_making",
				Market:   mkt.ConditionID,
				Question: mkt.Question,
				Signal:   "MARKET_MAKE",
				Price:    mid,
				EdgePct:  s.cfg.SpreadPct / 2,
				Reason:   fmt.Sprintf("bid=%.4f ask=%.4f", bid, ask),
				Executed: s.cfg.ExecuteOrders && s.executor != nil,
			})
		}
	}
}

// CalcMMPrices returns bid and ask prices around midpoint with the given spread %.
// bid = mid * (1 - spreadPct/100/2), ask = mid * (1 + spreadPct/100/2), clamped to (0.001, 0.999).
func CalcMMPrices(mid, spreadPct float64) (bid, ask float64) {
	half := spreadPct / 100 / 2
	bid = mid * (1 - half)
	ask = mid * (1 + half)
	if bid < 0.001 {
		bid = 0.001
	}
	if ask > 0.999 {
		ask = 0.999
	}
	return bid, ask
}
