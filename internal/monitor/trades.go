package monitor

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/analytics"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/order"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/rs/zerolog"
)

type TradesMonitor struct {
	analyticsHub *analytics.AnalyticsHub
	clobClient   *clob.Client
	dataClient   *data.Client
	notifier     notify.Notifier
	cfg          *config.TradesMonitorConfig
	logger       zerolog.Logger
	address      string

	mu        sync.RWMutex
	orders    []clob.Order
	trades    []clob.Trade
	positions []clob.Position

	prevOrderIDs    map[string]struct{}
	prevTradeIDs    map[string]struct{}
	expiredOrderIDs map[string]struct{}

	bus *tui.EventBus
}

func NewTradesMonitor(
	analyticsHub *analytics.AnalyticsHub,
	clobClient *clob.Client,
	dataClient *data.Client,
	notifier notify.Notifier,
	cfg *config.TradesMonitorConfig,
	log zerolog.Logger,
	address string,
	db ...interface{},
) *TradesMonitor {
	return &TradesMonitor{
		analyticsHub: analyticsHub,
		clobClient:   clobClient,
		dataClient:   dataClient,
		notifier:     notifier,
		cfg:          cfg,
		logger:          log.With().Str("component", "trades-monitor").Logger(),
		address:         address,
		prevOrderIDs:    make(map[string]struct{}),
		prevTradeIDs:    make(map[string]struct{}),
		expiredOrderIDs: make(map[string]struct{}),
	}
}

func (tm *TradesMonitor) SetBus(bus *tui.EventBus) {
	tm.bus = bus
}

func (tm *TradesMonitor) Run(ctx context.Context) error {
	interval := time.Duration(tm.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	tm.logger.Info().Dur("interval", interval).Msg(i18n.T().LogTradesMonitorStarted)
	tm.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tm.poll(ctx)
		}
	}
}

func (tm *TradesMonitor) poll(ctx context.Context) {
	tm.pollOrders(ctx)
	tm.pollTrades(ctx)
	if tm.cfg.TrackPositions {
		tm.pollPositions(ctx)
	}
}

func (tm *TradesMonitor) pollOrders(ctx context.Context) {
	resp, err := tm.clobClient.GetDataOrders(clob.OrdersFilter{
		MakerAddress: tm.address,
		Status:       "LIVE",
	})

	if err != nil {
		tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedFetchOrders)
		return
	}

	tm.mu.Lock()
	newOrderIDs := make(map[string]struct{}, len(resp.Data))
	for _, o := range resp.Data {
		newOrderIDs[o.ID] = struct{}{}
	}

	for id := range tm.prevOrderIDs {
		if _, ok := newOrderIDs[id]; !ok {
			tm.logger.Info().Str("order_id", id).Msg(i18n.T().LogOrderClosed)
			go func(orderID string, parentCtx context.Context) {
				notifCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
				defer cancel()
				msg := fmt.Sprintf(i18n.T().TgOrderClosed, orderID)
				if err := tm.notifier.Send(notifCtx, msg); err != nil && parentCtx.Err() == nil {
					tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
				}
			}(id, ctx)
		}
	}

	if len(tm.prevOrderIDs) > 0 {
		for _, o := range resp.Data {
			if _, ok := tm.prevOrderIDs[o.ID]; !ok {
				tm.logger.Info().Str("order_id", o.ID).Str("side", string(o.Side)).Str("price", o.Price).Msg(i18n.T().LogNewOrderDetected)
				go func(order clob.Order, parentCtx context.Context) {
					notifCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
					defer cancel()
					msg := fmt.Sprintf(i18n.T().TgNewOrder, order.Side, order.AssetID[:8]+"...", order.Price, order.OriginalSize)
					if err := tm.notifier.Send(notifCtx, msg); err != nil && parentCtx.Err() == nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(o, ctx)
			}
		}
	}

	// Mark expired GTD orders
	for idx := range resp.Data {
		if order.IsOrderExpired(&resp.Data[idx]) {
			resp.Data[idx].Status = clob.StatusExpired
			orderID := resp.Data[idx].ID
			if _, alreadyNotified := tm.expiredOrderIDs[orderID]; !alreadyNotified {
				tm.expiredOrderIDs[orderID] = struct{}{}
				tm.logger.Info().Str("order_id", orderID).Msg("GTD order expired")
				go func(expiredOrder clob.Order, parentCtx context.Context) {
					notifCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
					defer cancel()
					msg := fmt.Sprintf("Order %s expired (GTD expiration reached)", expiredOrder.ID[:8]+"...")
					if err := tm.notifier.Send(notifCtx, msg); err != nil && parentCtx.Err() == nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(resp.Data[idx], ctx)
			}
		}
	}

	tm.orders = resp.Data
	tm.prevOrderIDs = newOrderIDs
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(resp.Data)).Msg(i18n.T().LogOrdersUpdated)

	if tm.bus != nil {
		rows := make([]tui.OrderRow, 0, len(resp.Data))
		for _, o := range resp.Data {
			rows = append(rows, tui.OrderRow{
				Market: o.AssetID,
				Side:   string(o.Side),
				Price:  o.Price,
				Size:   o.OriginalSize,
				Filled: o.SizeFilled,
				Status: string(o.Status),
				ID:     o.ID,
			})
		}
		tm.bus.Send(tui.OrdersUpdateMsg{Rows: rows})
	}
}

func (tm *TradesMonitor) pollTrades(ctx context.Context) {
	filter := clob.TradesFilter{Limit: tm.cfg.TradesLimit}
	resp, err := tm.clobClient.GetTrades(filter)
	if err != nil {
		tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedFetchTrades)
		return
	}

	tm.mu.Lock()
	newTradeIDs := make(map[string]struct{}, len(resp.Data))
	for _, t := range resp.Data {
		newTradeIDs[t.ID] = struct{}{}
	}

	if len(tm.prevTradeIDs) > 0 {
		for _, t := range resp.Data {
			if _, ok := tm.prevTradeIDs[t.ID]; !ok {
				tm.logger.Info().
					Str("trade_id", t.ID).
					Str("side", string(t.Side)).
					Str("price", t.Price).
					Str("size", t.Size).
					Msg(i18n.T().LogTradeExecuted)
				
				if tm.analyticsHub != nil {
					p, _ := strconv.ParseFloat(t.Price, 64)
					s, _ := strconv.ParseFloat(t.Size, 64)
					tm.analyticsHub.RecordTrade(analytics.TradeReport{
						ID:        t.ID,
						MarketID:  t.AssetID,
						AssetID:   t.AssetID,
						Side:      string(t.Side),
						Price:     p,
						Size:      s,
						Volume:    p * s,
						Strategy:  "unknown",
						Timestamp: t.Timestamp,
					})
				}

				go func(trade clob.Trade, parentCtx context.Context) {
					notifCtx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
					defer cancel()
					msg := fmt.Sprintf(i18n.T().TgTradeExecuted,
						trade.Side, trade.AssetID[:8]+"...", trade.Price, trade.Size)
					if err := tm.notifier.Send(notifCtx, msg); err != nil && parentCtx.Err() == nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(t, ctx)
			}
		}
	}

	tm.trades = resp.Data
	tm.prevTradeIDs = newTradeIDs
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(resp.Data)).Msg("trades updated")
}

func (tm *TradesMonitor) pollPositions(_ context.Context) {
	if tm.address == "" {
		return
	}
	dataPositions, err := tm.dataClient.GetPositions(data.PositionsParams{
		User:          tm.address,
		SizeThreshold: 0.01,
		Limit:         200,
	})
	if err != nil {
		tm.logger.Warn().Err(err).Msg("failed to fetch positions")
		return
	}

	positions := make([]clob.Position, 0, len(dataPositions))
	for _, p := range dataPositions {
		positions = append(positions, clob.Position{
			AssetID:      p.Asset,
			ConditionID:  p.ConditionID,
			Outcome:      p.Outcome,
			Size:         p.Size,
			AveragePrice: p.AvgPrice,
			InitialValue: p.InitialValue,
			CurrentValue: p.CurrentValue,
			PnL:          p.CashPnl,
			RealizedPnL:  p.RealizedPnl,
		})
	}

	tm.mu.Lock()
	tm.positions = positions
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(positions)).Msg("positions updated")

	if tm.bus != nil {
		rows := make([]tui.PositionRow, 0, len(positions))
		for _, p := range positions {
			rows = append(rows, tui.PositionRow{
				Market:  p.AssetID,
				Side:    p.Outcome,
				Size:    fmt.Sprintf("%.4f", p.Size),
				Entry:   fmt.Sprintf("%.4f", p.AveragePrice),
				Current: fmt.Sprintf("%.4f", p.CurrentValue),
				PnL:     fmt.Sprintf("%+.2f", p.PnL),
				PnLPct:  "",
			})
		}
		tm.bus.Send(tui.PositionsUpdateMsg{Rows: rows})
	}
}

func (tm *TradesMonitor) GetOrders() []clob.Order {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Order, len(tm.orders))
	copy(result, tm.orders)
	return result
}

func (tm *TradesMonitor) GetTrades() []clob.Trade {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Trade, len(tm.trades))
	copy(result, tm.trades)
	return result
}

func (tm *TradesMonitor) GetPositions() []clob.Position {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Position, len(tm.positions))
	copy(result, tm.positions)
	return result
}

func (tm *TradesMonitor) CancelOrder(orderID string) error {
	resp, err := tm.clobClient.CancelOrder(orderID)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelOrder: %w", err)
	}
	if !resp.Canceled {
		return fmt.Errorf("trades-monitor: CancelOrder: order %s not canceled", orderID)
	}
	return nil
}

func (tm *TradesMonitor) CancelOrders(orderIDs []string) error {
	_, err := tm.clobClient.CancelOrders(orderIDs)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelOrders: %w", err)
	}
	return nil
}

func (tm *TradesMonitor) CancelAllOrders() error {
	_, err := tm.clobClient.CancelAllOrders()
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelAllOrders: %w", err)
	}
	return nil
}

func (tm *TradesMonitor) CancelMarketOrders(marketID, assetID string) error {
	_, err := tm.clobClient.CancelMarketOrders(marketID, assetID)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelMarketOrders: %w", err)
	}
	return nil
}

func (tm *TradesMonitor) GetDataPositions(walletAddress string) ([]data.Position, error) {
	return tm.dataClient.GetPositions(data.PositionsParams{
		User:  walletAddress,
		Limit: 100,
	})
}

func (tm *TradesMonitor) GetDataTrades(walletAddress string, limit int) ([]data.Trade, error) {
	if limit <= 0 {
		limit = 50
	}
	return tm.dataClient.GetTrades(data.TradesParams{
		User:  walletAddress,
		Limit: limit,
	})
}

func (tm *TradesMonitor) GetMarketTrades(tokenID string, limit int) ([]clob.Trade, error) {
	resp, err := tm.clobClient.GetMarketTrades(tokenID, limit)
	if err != nil {
		return nil, fmt.Errorf("trades-monitor: GetMarketTrades: %w", err)
	}
	return resp.Data, nil
}