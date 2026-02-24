package trading

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// Engine управляет жизненным циклом торговых стратегий.
type Engine struct {
	strategies []Strategy
	logger     zerolog.Logger
	mu         sync.RWMutex
}

// NewEngine создаёт торговый движок.
func NewEngine(log zerolog.Logger) *Engine {
	return &Engine{
		logger: log.With().Str("component", "trading-engine").Logger(),
	}
}

// Register добавляет стратегию в движок.
func (e *Engine) Register(s Strategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.strategies = append(e.strategies, s)
	e.logger.Info().Str("strategy", s.Name()).Msg("strategy registered")
}

// Start запускает все зарегистрированные стратегии.
// Каждая стратегия работает в отдельной горутине.
// Возвращается после остановки всех стратегий.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.RLock()
	strategies := make([]Strategy, len(e.strategies))
	copy(strategies, e.strategies)
	e.mu.RUnlock()

	if len(strategies) == 0 {
		e.logger.Warn().Msg("no strategies registered")
		return nil
	}

	errCh := make(chan error, len(strategies))
	var wg sync.WaitGroup

	for _, s := range strategies {
		wg.Add(1)
		go func(strategy Strategy) {
			defer wg.Done()
			e.logger.Info().Str("strategy", strategy.Name()).Msg("starting strategy")
			if err := strategy.Start(ctx); err != nil && ctx.Err() == nil {
				e.logger.Error().Err(err).Str("strategy", strategy.Name()).Msg("strategy error")
				errCh <- fmt.Errorf("strategy %q: %w", strategy.Name(), err)
			}
		}(s)
	}

	wg.Wait()
	close(errCh)

	// Возвращаем первую ошибку если есть
	for err := range errCh {
		return err
	}
	return nil
}

// Stop останавливает все стратегии.
func (e *Engine) Stop() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, s := range e.strategies {
		if err := s.Stop(); err != nil {
			e.logger.Error().Err(err).Str("strategy", s.Name()).Msg("error stopping strategy")
		}
	}
}
