package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/rs/zerolog"
)

// TradesMonitor отслеживает активные ордера, сделки и позиции пользователя.
// Периодически опрашивает CLOB и Data API, кэширует результаты,
// генерирует алерты при изменениях и предоставляет методы для управления ордерами.
type TradesMonitor struct {
	clobClient *clob.Client
	dataClient *data.Client
	notifier   notify.Notifier
	cfg        *config.TradesMonitorConfig
	logger     zerolog.Logger

	mu        sync.RWMutex
	orders    []clob.Order
	trades    []clob.Trade
	positions []clob.Position

	// Множество ID ордеров из предыдущего цикла (для детекта новых/закрытых)
	prevOrderIDs map[string]struct{}
	// Множество ID сделок из предыдущего цикла (для детекта новых исполнений)
	prevTradeIDs map[string]struct{}
}

// NewTradesMonitor создаёт TradesMonitor.
func NewTradesMonitor(
	clobClient *clob.Client,
	dataClient *data.Client,
	notifier notify.Notifier,
	cfg *config.TradesMonitorConfig,
	log zerolog.Logger,
) *TradesMonitor {
	return &TradesMonitor{
		clobClient:   clobClient,
		dataClient:   dataClient,
		notifier:     notifier,
		cfg:          cfg,
		logger:       log.With().Str("component", "trades-monitor").Logger(),
		prevOrderIDs: make(map[string]struct{}),
		prevTradeIDs: make(map[string]struct{}),
	}
}

// Run запускает мониторинг. Блокирует до отмены ctx.
func (tm *TradesMonitor) Run(ctx context.Context) error {
	interval := time.Duration(tm.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	tm.logger.Info().Dur("interval", interval).Msg(i18n.T().LogTradesMonitorStarted)

	// Первый цикл сразу при запуске
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

// poll выполняет один цикл опроса API.
func (tm *TradesMonitor) poll(ctx context.Context) {
	tm.pollOrders(ctx)
	tm.pollTrades(ctx)
	if tm.cfg.TrackPositions {
		tm.pollPositions(ctx)
	}
}

// pollOrders обновляет список открытых ордеров и генерирует алерты.
func (tm *TradesMonitor) pollOrders(ctx context.Context) {
	resp, err := tm.clobClient.GetOrders()
	if err != nil {
		tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedFetchOrders)
		return
	}

	tm.mu.Lock()
	newOrderIDs := make(map[string]struct{}, len(resp.Data))
	for _, o := range resp.Data {
		newOrderIDs[o.ID] = struct{}{}
	}

	// Детект отменённых/исполненных ордеров (были в prevOrderIDs, нет в новых)
	for id := range tm.prevOrderIDs {
		if _, ok := newOrderIDs[id]; !ok {
			tm.logger.Info().Str("order_id", id).Msg(i18n.T().LogOrderClosed)
			go func(orderID string) {
				msg := fmt.Sprintf(i18n.T().TgOrderClosed, orderID)
				if err := tm.notifier.Send(ctx, msg); err != nil {
					tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
				}
			}(id)
		}
	}

	// Детект новых ордеров (есть в новых, не было в prevOrderIDs)
	if len(tm.prevOrderIDs) > 0 {
		for _, o := range resp.Data {
			if _, ok := tm.prevOrderIDs[o.ID]; !ok {
				tm.logger.Info().Str("order_id", o.ID).Str("side", string(o.Side)).Str("price", o.Price).Msg(i18n.T().LogNewOrderDetected)
				go func(order clob.Order) {
					msg := fmt.Sprintf(i18n.T().TgNewOrder, order.Side, order.AssetID[:8]+"...", order.Price, order.OriginalSize)
					if err := tm.notifier.Send(ctx, msg); err != nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(o)
			}
		}
	}

	tm.orders = resp.Data
	tm.prevOrderIDs = newOrderIDs
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(resp.Data)).Msg(i18n.T().LogOrdersUpdated)
}

// pollTrades обновляет список сделок и генерирует алерты о новых исполнениях.
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

	// Новые исполненные сделки
	if len(tm.prevTradeIDs) > 0 {
		for _, t := range resp.Data {
			if _, ok := tm.prevTradeIDs[t.ID]; !ok {
				tm.logger.Info().
					Str("trade_id", t.ID).
					Str("side", string(t.Side)).
					Str("price", t.Price).
					Str("size", t.Size).
					Msg(i18n.T().LogTradeExecuted)
				go func(trade clob.Trade) {
					msg := fmt.Sprintf(i18n.T().TgTradeExecuted,
						trade.Side, trade.AssetID[:8]+"...", trade.Price, trade.Size)
					if err := tm.notifier.Send(ctx, msg); err != nil {
						tm.logger.Warn().Err(err).Msg(i18n.T().LogFailedSendAlert)
					}
				}(t)
			}
		}
	}

	tm.trades = resp.Data
	tm.prevTradeIDs = newTradeIDs
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(resp.Data)).Msg("trades updated")
}

// pollPositions обновляет текущие позиции пользователя.
func (tm *TradesMonitor) pollPositions(ctx context.Context) {
	positions, err := tm.clobClient.GetPositions()
	if err != nil {
		tm.logger.Warn().Err(err).Msg("failed to fetch positions")
		return
	}

	tm.mu.Lock()
	tm.positions = positions
	tm.mu.Unlock()

	tm.logger.Debug().Int("count", len(positions)).Msg("positions updated")
}

// --- Методы для чтения кэша ---

// GetOrders возвращает кэшированный список открытых ордеров.
func (tm *TradesMonitor) GetOrders() []clob.Order {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Order, len(tm.orders))
	copy(result, tm.orders)
	return result
}

// GetTrades возвращает кэшированный список последних сделок.
func (tm *TradesMonitor) GetTrades() []clob.Trade {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Trade, len(tm.trades))
	copy(result, tm.trades)
	return result
}

// GetPositions возвращает кэшированные текущие позиции.
func (tm *TradesMonitor) GetPositions() []clob.Position {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]clob.Position, len(tm.positions))
	copy(result, tm.positions)
	return result
}

// --- Методы взаимодействия с ордерами ---

// CancelOrder отменяет ордер по ID.
func (tm *TradesMonitor) CancelOrder(orderID string) error {
	tm.logger.Info().Str("order_id", orderID).Msg("canceling order")
	resp, err := tm.clobClient.CancelOrder(orderID)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelOrder: %w", err)
	}
	if !resp.Canceled {
		return fmt.Errorf("trades-monitor: CancelOrder: order %s not canceled", orderID)
	}
	tm.logger.Info().Str("order_id", orderID).Msg("order canceled")
	return nil
}

// CancelOrders отменяет несколько ордеров по ID.
func (tm *TradesMonitor) CancelOrders(orderIDs []string) error {
	tm.logger.Info().Strs("order_ids", orderIDs).Msg("canceling orders")
	_, err := tm.clobClient.CancelOrders(orderIDs)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelOrders: %w", err)
	}
	tm.logger.Info().Int("count", len(orderIDs)).Msg("orders canceled")
	return nil
}

// CancelAllOrders отменяет все открытые ордера пользователя.
func (tm *TradesMonitor) CancelAllOrders() error {
	tm.logger.Info().Msg("canceling all orders")
	_, err := tm.clobClient.CancelAllOrders()
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelAllOrders: %w", err)
	}
	tm.logger.Info().Msg("all orders canceled")
	return nil
}

// CancelMarketOrders отменяет все открытые ордера на конкретном рынке.
// marketID = condition_id, assetID = token_id (один из двух обязателен).
func (tm *TradesMonitor) CancelMarketOrders(marketID, assetID string) error {
	tm.logger.Info().Str("market", marketID).Str("asset", assetID).Msg("canceling market orders")
	_, err := tm.clobClient.CancelMarketOrders(marketID, assetID)
	if err != nil {
		return fmt.Errorf("trades-monitor: CancelMarketOrders: %w", err)
	}
	tm.logger.Info().Str("market", marketID).Msg("market orders canceled")
	return nil
}

// --- Методы для получения данных из Data API ---

// GetDataPositions возвращает позиции пользователя из Data API (по адресу кошелька).
// Обогащённые данные: P&L, названия рынков, slug и т.д.
func (tm *TradesMonitor) GetDataPositions(walletAddress string) ([]data.Position, error) {
	return tm.dataClient.GetPositions(data.PositionsParams{
		User:  walletAddress,
		Limit: 100,
	})
}

// GetDataTrades возвращает историю сделок из Data API (по адресу кошелька).
func (tm *TradesMonitor) GetDataTrades(walletAddress string, limit int) ([]data.Trade, error) {
	if limit <= 0 {
		limit = 50
	}
	return tm.dataClient.GetTrades(data.TradesParams{
		User:  walletAddress,
		Limit: limit,
	})
}

// GetMarketTrades возвращает публичную историю сделок на рынке по token_id.
func (tm *TradesMonitor) GetMarketTrades(tokenID string, limit int) ([]clob.Trade, error) {
	resp, err := tm.clobClient.GetMarketTrades(tokenID, limit)
	if err != nil {
		return nil, fmt.Errorf("trades-monitor: GetMarketTrades: %w", err)
	}
	return resp.Data, nil
}
