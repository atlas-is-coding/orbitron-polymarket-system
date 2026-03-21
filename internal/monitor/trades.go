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
	"github.com/atlasdev/orbitron/internal/notification"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/order"
	"github.com/atlasdev/orbitron/internal/storage"
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

	db         storage.Store
	optCache   *order.OptimisticCache
	notifQueue *notification.Queue
}

func NewTradesMonitor(
	analyticsHub *analytics.AnalyticsHub,
	clobClient *clob.Client,
	dataClient *data.Client,
	notifier notify.Notifier,
	cfg *config.TradesMonitorConfig,
	log zerolog.Logger,
	address string,
	dbStore storage.Store,
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
		db:              dbStore,
		optCache:        order.NewOptimisticCache(),
		notifQueue:      notification.NewQueue(notifier, dbStore, log),
	}
}

func (tm *TradesMonitor) SetBus(bus *tui.EventBus) {
	tm.bus = bus
}

func (tm *TradesMonitor) Run(ctx context.Context) error {
	orderTicker := time.NewTicker(time.Duration(tm.cfg.PollIntervalMs) * time.Millisecond)
	defer orderTicker.Stop()

	var positionsTicker *time.Ticker
	if tm.cfg.TrackPositions {
		positionsTicker = time.NewTicker(time.Duration(tm.cfg.PositionsPollMs) * time.Millisecond)
		defer positionsTicker.Stop()
	}

	tm.logger.Info().
		Int("orders_poll_ms", tm.cfg.PollIntervalMs).
		Int("positions_poll_ms", tm.cfg.PositionsPollMs).
		Int("max_backoff_ms", tm.cfg.MaxBackoffMs).
		Msg(i18n.T().LogTradesMonitorStarted)

	tm.poll(ctx)

	backoffMs := tm.cfg.PollIntervalMs
	errorCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-orderTicker.C:
			err := tm.pollWithError(ctx)
			if err != nil {
				errorCount++
				backoffMs = tm.cfg.PollIntervalMs * (1 << uint(errorCount))
				if backoffMs > tm.cfg.MaxBackoffMs {
					backoffMs = tm.cfg.MaxBackoffMs
				}
				tm.logger.Warn().Err(err).Int("backoff_ms", backoffMs).Int("error_count", errorCount).
					Msg("orders poll failed, applying exponential backoff")
				orderTicker.Reset(time.Duration(backoffMs) * time.Millisecond)
			} else {
				if errorCount > 0 {
					tm.logger.Info().Msg("orders poll succeeded, resetting backoff")
				}
				errorCount = 0
				backoffMs = tm.cfg.PollIntervalMs
				orderTicker.Reset(time.Duration(tm.cfg.PollIntervalMs) * time.Millisecond)
			}
		case <-positionsTicker.C:
			if tm.cfg.TrackPositions {
				tm.pollPositions(ctx)
			}
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

// pollWithError performs polling and returns error if any critical operation fails
func (tm *TradesMonitor) pollWithError(ctx context.Context) error {
	tm.pollOrders(ctx)
	tm.pollTrades(ctx)
	// Note: positions polling is handled separately with its own ticker
	return nil
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
				// Enqueue notification instead of sending directly
				notif := &storage.Notification{
					WalletAddress: tm.address,
					EventType:     "ORDER_CLOSED",
					Payload:       msg,
				}
				if err := tm.notifQueue.Enqueue(notifCtx, notif); err != nil && parentCtx.Err() == nil {
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
					// Enqueue notification instead of sending directly
					notif := &storage.Notification{
						WalletAddress: tm.address,
						EventType:     "ORDER_PLACED",
						Payload:       msg,
					}
					if err := tm.notifQueue.Enqueue(notifCtx, notif); err != nil && parentCtx.Err() == nil {
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
					// Enqueue notification instead of sending directly
					notif := &storage.Notification{
						WalletAddress: tm.address,
						EventType:     "ORDER_EXPIRED",
						Payload:       msg,
					}
					if err := tm.notifQueue.Enqueue(notifCtx, notif); err != nil && parentCtx.Err() == nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(resp.Data[idx], ctx)
			}
		}
	}

	// Save orders to database and reconcile cache
	if tm.db != nil {
		for idx := range resp.Data {
			dbOrder := &storage.Order{
				ID:            resp.Data[idx].ID,
				WalletAddress: tm.address,
				ConditionID:   "", // ConditionID is not provided by CLOB API, using AssetID (token_id) as primary identifier
				AssetID:       resp.Data[idx].AssetID,
				Side:          string(resp.Data[idx].Side),
				OrderType:     string(resp.Data[idx].OrderType),
				Price:         parseFloat(resp.Data[idx].Price),
				Size:          parseFloat(resp.Data[idx].OriginalSize),
				Status:        string(resp.Data[idx].Status),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if resp.Data[idx].ExpiresAt > 0 {
				expiresAt := time.UnixMilli(resp.Data[idx].ExpiresAt)
				dbOrder.ExpiresAt = &expiresAt
			}
			if err := tm.db.InsertOrder(ctx, dbOrder); err != nil {
				tm.logger.Warn().Err(err).Str("order_id", resp.Data[idx].ID).Msg("failed to save order to database")
			}
			// Reconcile optimistic cache with API data
			tm.optCache.Reconcile(ctx, resp.Data[idx].ID, &resp.Data[idx])
		}
	}

	tm.orders = resp.Data
	tm.prevOrderIDs = newOrderIDs
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(resp.Data)).Msg(i18n.T().LogOrdersUpdated)

	// Process pending notifications
	if err := tm.notifQueue.ProcessPending(ctx); err != nil {
		tm.logger.Warn().Err(err).Msg("failed to process pending notifications")
	}

	if tm.bus != nil {
		rows := make([]tui.OrderRow, 0, len(resp.Data))
		for _, o := range resp.Data {
			status := string(o.Status)
			// Check if order is optimistically canceled and show "CANCELING..." instead of "LIVE"
			if tm.optCache.IsOptimisticallyCanceled(ctx, o.ID) && status == string(clob.StatusLive) {
				status = "CANCELING..."
			}
			rows = append(rows, tui.OrderRow{
				Market: o.AssetID,
				Side:   string(o.Side),
				Price:  o.Price,
				Size:   o.OriginalSize,
				Filled: o.SizeFilled,
				Status: status,
				ID:     o.ID,
			})
		}
		tm.bus.Send(tui.OrdersUpdateMsg{Rows: rows})
	}
}

// Helper function to parse float strings
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
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
	// Apply optimistic update immediately for instant UI feedback
	tm.optCache.MarkCanceled(context.Background(), orderID)

	// Execute API call asynchronously (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := tm.clobClient.CancelOrder(orderID)
		if err != nil {
			tm.logger.Warn().Err(err).Str("order_id", orderID).Msg("failed to cancel order via API")
			// Rollback optimistic update on error
			tm.optCache.Reconcile(ctx, orderID, nil)
			return
		}
		if !resp.Canceled {
			tm.logger.Warn().Str("order_id", orderID).Msg("order cancellation not confirmed by API")
			// Rollback optimistic update on error
			tm.optCache.Reconcile(ctx, orderID, nil)
			return
		}

		tm.logger.Info().Str("order_id", orderID).Msg("order canceled successfully")
		// Reconcile cache with API response (mark as Canceled)
		canceledOrder := &clob.Order{
			ID:     orderID,
			Status: clob.StatusCanceled,
		}
		tm.optCache.Reconcile(ctx, orderID, canceledOrder)
	}()

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