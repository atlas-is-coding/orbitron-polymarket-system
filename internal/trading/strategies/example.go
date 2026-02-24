// Package strategies содержит примеры торговых стратегий.
package strategies

import (
	"context"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/rs/zerolog"
)

// ExampleStrategy — демонстрационная стратегия: периодически запрашивает цены.
// Используйте как шаблон для создания реальных стратегий.
type ExampleStrategy struct {
	clob   *clob.Client
	logger zerolog.Logger
	done   chan struct{}
}

// NewExampleStrategy создаёт пример стратегии.
func NewExampleStrategy(clobClient *clob.Client, log zerolog.Logger) *ExampleStrategy {
	return &ExampleStrategy{
		clob:   clobClient,
		logger: log.With().Str("strategy", "example").Logger(),
		done:   make(chan struct{}),
	}
}

// Name возвращает имя стратегии.
func (s *ExampleStrategy) Name() string { return "example" }

// Start запускает стратегию.
func (s *ExampleStrategy) Start(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	s.logger.Info().Msg("example strategy started")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.done:
			return nil
		case <-ticker.C:
			// Пример: получаем список рынков
			markets, err := s.clob.GetMarkets("")
			if err != nil {
				s.logger.Warn().Err(err).Msg("failed to fetch markets")
				continue
			}
			s.logger.Info().Int("count", len(markets.Data)).Msg("markets fetched")
		}
	}
}

// Stop останавливает стратегию.
func (s *ExampleStrategy) Stop() error {
	close(s.done)
	return nil
}
