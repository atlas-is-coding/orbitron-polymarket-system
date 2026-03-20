package monitor

import (
	"context"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/rs/zerolog"
)

// GammaClient определяет методы Gamma API, необходимые монитору.
type GammaClient interface {
	GetMarkets(params gamma.MarketsParams) ([]gamma.Market, error)
}

// Monitor периодически опрашивает рынки и генерирует алерты.
type Monitor struct {
	gamma    GammaClient
	notifier notify.Notifier
	cfg      *config.MonitorConfig
	rules    []Rule
	logger   zerolog.Logger
	store    storage.Store

	// Предыдущие состояния рынков для сравнения
	mu        sync.RWMutex
	prevState map[string]*gamma.Market
}

// New создаёт Monitor.
func New(
	gammaClient GammaClient,
	notifier notify.Notifier,
	cfg *config.MonitorConfig,
	log zerolog.Logger,
) *Monitor {
	return &Monitor{
		gamma:     gammaClient,
		notifier:  notifier,
		cfg:       cfg,
		rules:     DefaultRules,
		logger:    log.With().Str("component", "monitor").Logger(),
		prevState: make(map[string]*gamma.Market),
	}
}

// WithStore добавляет хранилище в монитор
func (m *Monitor) WithStore(s storage.Store) *Monitor {
	m.store = s
	return m
}

// Run запускает мониторинг. Блокирует до отмены ctx.
func (m *Monitor) Run(ctx context.Context) error {
	interval := time.Duration(m.cfg.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.logger.Info().Dur("interval", interval).Msg(i18n.T().LogMonitorStarted)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.poll(ctx)
		}
	}
}

func (m *Monitor) poll(ctx context.Context) {
	active := true
	markets, err := m.gamma.GetMarkets(gamma.MarketsParams{
		Active: &active,
		Limit:  100,
	})
	if err != nil {
		m.logger.Warn().Err(err).Msg("failed to fetch markets")
		return
	}

	for i := range markets {
		mkt := &markets[i]

		// Фильтруем если задан список condition_id
		if len(m.cfg.Markets) > 0 && !contains(m.cfg.Markets, mkt.ConditionID) {
			continue
		}

		alerts := m.evaluate(mkt)
		for _, alert := range alerts {
			// Проверка на дубликаты в БД
			if m.store != nil {
				// Cooldown 24h для рыночных алертов (low liquidity, high volume)
				cooldown := 24 * time.Hour
				sent, err := m.store.WasAlertSent(ctx, string(alert.Type), mkt.ConditionID, cooldown)
				if err == nil && sent {
					continue
				}
			}

			m.logger.Info().
				Str("type", string(alert.Type)).
				Str("market", mkt.ConditionID).
				Str("message", alert.Message).
				Msg("alert triggered")

			go func(a Alert, parentCtx context.Context) {
				maxRetries := 3
				for i := 0; i < maxRetries; i++ {
					if parentCtx.Err() != nil {
						return
					}
					notifCtx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
					err := m.notifier.Send(notifCtx, a.Message)
					cancel()
					if err == nil {
						if m.store != nil {
							storeCtx, storeCancel := context.WithTimeout(parentCtx, 5*time.Second)
							if storeErr := m.store.MarkAlertSent(storeCtx, string(a.Type), a.Market.ConditionID); storeErr != nil {
								m.logger.Warn().Err(storeErr).Msg("failed to mark alert as sent")
							}
							storeCancel()
						}
						return
					}
					m.logger.Warn().Err(err).Int("attempt", i+1).Msg("failed to send alert, retrying")
					select {
					case <-parentCtx.Done():
						return
					case <-time.After(time.Duration(i+1) * 2 * time.Second):
					}
				}
				m.logger.Error().Msg("failed to send alert after max retries")
			}(alert, ctx)
		}

		m.mu.Lock()
		m.prevState[mkt.ConditionID] = mkt
		m.mu.Unlock()
	}
}

// evaluate проверяет правила для рынка и возвращает сработавшие алерты.
func (m *Monitor) evaluate(mkt *gamma.Market) []Alert {
	m.mu.RLock()
	prev, hasPrev := m.prevState[mkt.ConditionID]
	m.mu.RUnlock()

	var alerts []Alert

	for _, rule := range m.rules {
		switch rule.AlertType {
		case AlertPriceChange:
			if hasPrev && len(mkt.OutcomePrices) > 0 && len(prev.OutcomePrices) > 0 {
				currPrice, _ := strconv.ParseFloat(mkt.OutcomePrices[0], 64)
				prevPrice, _ := strconv.ParseFloat(prev.OutcomePrices[0], 64)
				diff := math.Abs(currPrice - prevPrice)
				if diff >= rule.Threshold {
					mktCopy := *mkt
					alerts = append(alerts, Alert{
						Type:    AlertPriceChange,
						Market:  &mktCopy,
						Message: formatAlert(AlertPriceChange, mkt, rule),
					})
				}
			}

		case AlertLowLiquidity:
			if float64(mkt.Liquidity) < rule.Threshold {
				mktCopy := *mkt
				alerts = append(alerts, Alert{
					Type:    AlertLowLiquidity,
					Market:  &mktCopy,
					Message: formatAlert(AlertLowLiquidity, mkt, rule),
				})
			}

		case AlertMarketClosed:
			if !hasPrev && mkt.Closed {
				mktCopy := *mkt
				alerts = append(alerts, Alert{
					Type:    AlertMarketClosed,
					Market:  &mktCopy,
					Message: formatAlert(AlertMarketClosed, mkt, rule),
				})
			}

		case AlertHighVolume:
			if float64(mkt.Volume) > rule.Threshold {
				if !hasPrev || float64(prev.Volume) <= rule.Threshold {
					mktCopy := *mkt
					alerts = append(alerts, Alert{
						Type:    AlertHighVolume,
						Market:  &mktCopy,
						Message: formatAlert(AlertHighVolume, mkt, rule),
					})
				}
			}
		}
	}

	return alerts
}

func formatAlert(t AlertType, mkt *gamma.Market, rule Rule) string {
	switch t {
	case AlertPriceChange:
		price := "0.00"
		if len(mkt.OutcomePrices) > 0 {
			price = mkt.OutcomePrices[0]
		}
		return "⚠️ Изменение цены: " + mkt.Question + " (текущая: " + price + ")"
	case AlertLowLiquidity:
		return "⚠️ Низкая ликвидность: " + mkt.Question + " ($" + fmtFloat(float64(mkt.Liquidity)) + ")"
	case AlertMarketClosed:
		return "🔒 Рынок закрыт: " + mkt.Question
	case AlertHighVolume:
		return "📈 Высокий объём: " + mkt.Question + " ($" + fmtFloat(float64(mkt.Volume)) + ")"
	default:
		return string(t) + ": " + mkt.Question
	}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
