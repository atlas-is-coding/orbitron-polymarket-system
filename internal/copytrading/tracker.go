package copytrading

import (
	"context"
	"fmt"
	"time"

	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/tui"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TraderTracker следит за позициями одного трейдера и копирует изменения.
// Каждый трекер работает в отдельной горутине.
type TraderTracker struct {
	trader       config.TraderConfig
	dataClient   *data.Client
	executor     *OrderExecutor
	sizer        *SizeCalculator
	store        storage.CopyTradeStore
	notifier     notify.Notifier
	bus          *tui.EventBus
	logger       zerolog.Logger
	getMyBalance func() (float64, error)

	prev TraderState
}

// NewTraderTracker создаёт трекер для одного трейдера.
func NewTraderTracker(
	trader config.TraderConfig,
	dataClient *data.Client,
	executor *OrderExecutor,
	store storage.CopyTradeStore,
	notifier notify.Notifier,
	getMyBalance func() (float64, error),
	bus *tui.EventBus,
	log zerolog.Logger,
) *TraderTracker {
	sizer := NewSizeCalculator(trader.SizeMode, trader.AllocationPct, trader.MaxPositionUSD)
	return &TraderTracker{
		trader:       trader,
		dataClient:   dataClient,
		executor:     executor,
		sizer:        sizer,
		store:        store,
		notifier:     notifier,
		bus:          bus,
		logger:       log.With().Str("trader", trader.Label).Str("address", trader.Address).Logger(),
		getMyBalance: getMyBalance,
		prev:         make(TraderState),
	}
}

// Run запускает цикл мониторинга позиций трейдера. Блокирует до отмены ctx.
func (t *TraderTracker) Run(ctx context.Context, pollInterval time.Duration) error {
	t.logger.Info().Dur("interval", pollInterval).Msg("trader tracker started")

	// При старте: сверяем открытые DB-записи с реальными позициями
	if err := t.reconcile(ctx); err != nil {
		t.logger.Warn().Err(err).Msg("reconcile on startup failed")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.logger.Info().Msg("trader tracker stopped")
			return nil
		case <-ticker.C:
			t.poll(ctx)
		}
	}
}

// reconcile сверяет открытые DB-записи с актуальными позициями трейдера.
// Закрывает записи для позиций, которые трейдер закрыл пока бот был оффлайн.
func (t *TraderTracker) reconcile(ctx context.Context) error {
	openTrades, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		return fmt.Errorf("load open trades: %w", err)
	}
	if len(openTrades) == 0 {
		return nil
	}

	positions, err := t.dataClient.GetPositions(data.PositionsParams{
		User:  t.trader.Address,
		Limit: 200,
	})
	if err != nil {
		return fmt.Errorf("fetch positions for reconcile: %w", err)
	}

	curr := toTraderState(positions)
	t.prev = curr

	for _, trade := range openTrades {
		if _, exists := curr[trade.AssetID]; !exists {
			t.logger.Warn().
				Str("asset_id", trade.AssetID).
				Str("trade_id", trade.ID).
				Msg("orphaned position found: trader closed while bot was offline")
			t.closePosition(ctx, trade)
		}
	}
	return nil
}

// poll выполняет один цикл опроса: получает текущие позиции и применяет diff.
func (t *TraderTracker) poll(ctx context.Context) {
	positions, err := t.dataClient.GetPositions(data.PositionsParams{
		User:  t.trader.Address,
		Limit: 200,
	})
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to fetch trader positions")
		return
	}

	curr := toTraderState(positions)
	diff := diffStates(t.prev, curr)
	t.prev = curr

	for _, pos := range diff.Opened {
		t.openPosition(ctx, pos)
	}
	for _, pos := range diff.Closed {
		t.handleTraderClosed(ctx, pos)
	}
}

// openPosition копирует открытие новой позиции трейдера.
func (t *TraderTracker) openPosition(ctx context.Context, pos data.Position) {
	// Проверяем, не отслеживаем ли уже эту позицию
	existing, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to check existing copy trades")
		return
	}
	for _, e := range existing {
		if e.AssetID == pos.Asset {
			t.logger.Debug().Str("asset_id", pos.Asset).Msg("position already tracked, skipping")
			return
		}
	}

	myBalance, err := t.getMyBalance()
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to get our balance")
		return
	}
	if myBalance <= 0 {
		t.logger.Warn().Float64("balance", myBalance).Msg("our balance is zero, skipping copy")
		return
	}

	traderTotalBalance := t.estimateTraderBalance()
	sizeUSD := t.sizer.Calculate(pos.CurrentValue, traderTotalBalance, myBalance)
	if sizeUSD <= 0 {
		t.logger.Warn().
			Float64("size_usd", sizeUSD).
			Str("asset_id", pos.Asset).
			Msg("calculated size is 0, skipping")
		return
	}

	t.logger.Info().
		Str("asset_id", pos.Asset).
		Str("market", pos.Title).
		Str("outcome", pos.Outcome).
		Float64("size_usd", sizeUSD).
		Float64("trader_avg_price", pos.AvgPrice).
		Msg("opening copy position")

	result, err := t.executor.Open(pos.Asset, sizeUSD, false)
	if err != nil {
		t.logger.Error().Err(err).Str("asset_id", pos.Asset).Msg("failed to open copy position")
		t.sendAlert(ctx, fmt.Sprintf(
			"❌ Failed to copy open: [%s] %s (%s)\nError: %v",
			t.trader.Label, pos.Title, pos.Outcome, err,
		))
		return
	}

	rec := &storage.CopyTradeRecord{
		ID:            uuid.New().String(),
		TraderAddress: t.trader.Address,
		AssetID:       pos.Asset,
		ConditionID:   pos.ConditionID,
		Side:          "BUY",
		Size:          result.Size,
		Price:         result.Price,
		OurOrderID:    result.OrderID,
		Status:        "open",
		OpenedAt:      time.Now().UTC(),
	}
	if err := t.store.SaveCopyTrade(ctx, rec); err != nil {
		t.logger.Error().Err(err).Msg("failed to save copy trade to DB")
	}

	t.sendAlert(ctx, fmt.Sprintf(
		"📈 Opened copy trade: [%s] %s (%s)\nSize: $%.2f @ %.4f",
		t.trader.Label, pos.Title, pos.Outcome, sizeUSD, result.Price,
	))
	if t.bus != nil {
		t.bus.Send(tui.CopytradingTradeMsg{Line: fmt.Sprintf(
			"📈 Opened [%s] %s (%s) $%.2f @ %.4f",
			t.trader.Label, pos.Title, pos.Outcome, sizeUSD, result.Price,
		)})
	}
}

// handleTraderClosed обрабатывает исчезновение позиции у трейдера.
func (t *TraderTracker) handleTraderClosed(ctx context.Context, pos data.Position) {
	openTrades, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to load open trades")
		return
	}
	for _, trade := range openTrades {
		if trade.AssetID == pos.Asset {
			t.closePosition(ctx, trade)
			return
		}
	}
	t.logger.Debug().Str("asset_id", pos.Asset).Msg("trader closed position we weren't tracking")
}

// closePosition закрывает нашу скопированную позицию рыночным ордером.
func (t *TraderTracker) closePosition(ctx context.Context, trade *storage.CopyTradeRecord) {
	t.logger.Info().
		Str("asset_id", trade.AssetID).
		Str("trade_id", trade.ID).
		Float64("size", trade.Size).
		Msg("closing copy position")

	result, err := t.executor.Close(trade.AssetID, trade.Size, trade.Price, false)

	now := time.Now().UTC()
	status := "closed"
	var pnl *float64

	if err != nil {
		t.logger.Error().Err(err).Str("trade_id", trade.ID).Msg("failed to close copy position")
		status = "failed"
		t.sendAlert(ctx, fmt.Sprintf(
			"❌ Failed to close copy: [%s] asset=%s\nError: %v",
			t.trader.Label, trade.AssetID, err,
		))
	} else {
		pnl = &result.PnL
	}

	if err := t.store.UpdateCopyTrade(ctx, trade.ID, status, &now, pnl); err != nil {
		t.logger.Error().Err(err).Msg("failed to update copy trade in DB")
	}

	if result != nil {
		t.sendAlert(ctx, fmt.Sprintf(
			"📉 Closed copy trade: [%s] asset=%s\nPnL: $%.2f",
			t.trader.Label, trade.AssetID, result.PnL,
		))
		if t.bus != nil {
			t.bus.Send(tui.CopytradingTradeMsg{Line: fmt.Sprintf(
				"📉 Closed [%s] asset=%s P&L: $%.2f",
				t.trader.Label, trade.AssetID, result.PnL,
			)})
		}
	}
}

// estimateTraderBalance суммирует CurrentValue всех известных позиций трейдера.
// Используется как приближение общего баланса для пропорционального расчёта.
func (t *TraderTracker) estimateTraderBalance() float64 {
	total := 0.0
	for _, pos := range t.prev {
		total += pos.CurrentValue
	}
	return total
}

func (t *TraderTracker) sendAlert(ctx context.Context, msg string) {
	if err := t.notifier.Send(ctx, msg); err != nil {
		t.logger.Warn().Err(err).Msg("failed to send alert")
	}
}
