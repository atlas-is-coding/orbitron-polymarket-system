package monitor

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/rs/zerolog"
)

// Monitor периодически опрашивает рынки и генерирует алерты.
type Monitor struct {
	gamma    *gamma.Client
	notifier notify.Notifier
	cfg      *config.MonitorConfig
	rules    []Rule
	logger   zerolog.Logger
	store    storage.Store

	// Предыдущие состояния рынков для сравнения
	prevState map[string]*gamma.Market
}

// New создаёт Monitor.
func New(
	gammaClient *gamma.Client,
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
			m.logger.Info().
				Str("type", string(alert.Type)).
				Str("market", mkt.ConditionID).
				Str("message", alert.Message).
				Msg("alert triggered")

			if err := m.notifier.Send(ctx, alert.Message); err != nil {
				m.logger.Warn().Err(err).Msg("failed to send alert")
			}
		}

		m.prevState[mkt.ConditionID] = mkt
	}
}

// evaluate проверяет правила для рынка и возвращает сработавшие алерты.
func (m *Monitor) evaluate(mkt *gamma.Market) []Alert {
	prev, hasPrev := m.prevState[mkt.ConditionID]
	var alerts []Alert

	for _, rule := range m.rules {
		switch rule.AlertType {
		case AlertPriceChange:
			if hasPrev && len(mkt.OutcomePrices) > 0 && len(prev.OutcomePrices) > 0 {
				// Сравниваем первый исход (YES)
				// Цены хранятся как строки, упрощаем для примера
				_ = math.Abs // используется ниже
			}

		case AlertLowLiquidity:
			if float64(mkt.Liquidity) < rule.Threshold {
				alerts = append(alerts, Alert{
					Type:    AlertLowLiquidity,
					Market:  mkt,
					Message: formatAlert(AlertLowLiquidity, mkt, rule),
				})
			}

		case AlertMarketClosed:
			if !hasPrev && mkt.Closed {
				alerts = append(alerts, Alert{
					Type:    AlertMarketClosed,
					Market:  mkt,
					Message: formatAlert(AlertMarketClosed, mkt, rule),
				})
			}

		case AlertHighVolume:
			if float64(mkt.Volume) > rule.Threshold {
				if !hasPrev || float64(prev.Volume) <= rule.Threshold {
					alerts = append(alerts, Alert{
						Type:    AlertHighVolume,
						Market:  mkt,
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
