package trading

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// strategyEntry wraps a strategy with per-goroutine lifecycle control.
type strategyEntry struct {
	strategy Strategy
	cancel   context.CancelFunc // nil if not running
}

// Engine управляет жизненным циклом торговых стратегий.
type Engine struct {
	entries []strategyEntry
	logger  zerolog.Logger
	mu      sync.RWMutex
}

// NewEngine создаёт торговый движок. Additional arguments are accepted but ignored
// to allow callers to pass optional parameters (e.g. wallet manager).
func NewEngine(log zerolog.Logger, _ ...interface{}) *Engine {
	return &Engine{
		logger: log.With().Str("component", "trading-engine").Logger(),
	}
}

// Register добавляет стратегию в движок.
func (e *Engine) Register(s Strategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.entries = append(e.entries, strategyEntry{strategy: s})
	e.logger.Info().Str("strategy", s.Name()).Msg("strategy registered")
}

// Get returns the strategy with the given name and true, or nil and false if not found.
func (e *Engine) Get(name string) (Strategy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, en := range e.entries {
		if en.strategy.Name() == name {
			return en.strategy, true
		}
	}
	return nil, false
}

// Strategies returns a snapshot of all registered strategies.
func (e *Engine) Strategies() []Strategy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Strategy, len(e.entries))
	for i, en := range e.entries {
		out[i] = en.strategy
	}
	return out
}

// Start запускает все зарегистрированные стратегии.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.RLock()
	entries := make([]strategyEntry, len(e.entries))
	copy(entries, e.entries)
	e.mu.RUnlock()

	if len(entries) == 0 {
		e.logger.Warn().Msg("no strategies registered")
		return nil
	}

	errCh := make(chan error, len(entries))
	var wg sync.WaitGroup

	for _, en := range entries {
		wg.Add(1)
		go func(s Strategy) {
			defer wg.Done()
			e.logger.Info().Str("strategy", s.Name()).Msg("starting strategy")
			if err := s.Start(ctx); err != nil && ctx.Err() == nil {
				e.logger.Error().Err(err).Str("strategy", s.Name()).Msg("strategy error")
				errCh <- fmt.Errorf("strategy %q: %w", s.Name(), err)
			}
		}(en.strategy)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

// StartStrategy starts a single strategy by name in a background goroutine.
func (e *Engine) StartStrategy(ctx context.Context, name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, en := range e.entries {
		if en.strategy.Name() == name {
			if e.entries[i].cancel != nil {
				return fmt.Errorf("strategy %q is already running", name)
			}
			sCtx, cancel := context.WithCancel(ctx)
			e.entries[i].cancel = cancel
			go func(s Strategy, c context.CancelFunc) {
				defer c()
				if err := s.Start(sCtx); err != nil && sCtx.Err() == nil {
					e.logger.Error().Err(err).Str("strategy", s.Name()).Msg("strategy error")
				}
			}(en.strategy, cancel)
			e.logger.Info().Str("strategy", name).Msg("strategy started")
			return nil
		}
	}
	return fmt.Errorf("strategy %q not found", name)
}

// StopStrategy stops a single strategy by name.
func (e *Engine) StopStrategy(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, en := range e.entries {
		if en.strategy.Name() == name {
			if e.entries[i].cancel != nil {
				e.entries[i].cancel()
				e.entries[i].cancel = nil
			}
			if err := en.strategy.Stop(); err != nil {
				return fmt.Errorf("stop strategy %q: %w", name, err)
			}
			e.logger.Info().Str("strategy", name).Msg("strategy stopped")
			return nil
		}
	}
	return fmt.Errorf("strategy %q not found", name)
}

// executorSetter is optionally implemented by strategies that support runtime executor updates.
type executorSetter interface {
	SetExecutors(executors map[string]interface{})
}

// walletIDSetter is optionally implemented by strategies that track wallet assignment.
type walletIDSetter interface {
	SetWalletIDs(ids []string)
}

// SetStrategyWallets updates the executor map for a named strategy.
// executors maps wallet ID → strategies.Executor (passed as interface{}).
func (e *Engine) SetStrategyWallets(name string, walletIDs []string, executors map[string]interface{}) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, en := range e.entries {
		if en.strategy.Name() == name {
			if es, ok := en.strategy.(executorSetter); ok {
				es.SetExecutors(executors)
			}
			if ws, ok := en.strategy.(walletIDSetter); ok {
				ws.SetWalletIDs(walletIDs)
			}
			return nil
		}
	}
	return fmt.Errorf("strategy %q not found", name)
}

// Stop останавливает все стратегии.
func (e *Engine) Stop() {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, en := range e.entries {
		if err := en.strategy.Stop(); err != nil {
			e.logger.Error().Err(err).Str("strategy", en.strategy.Name()).Msg("error stopping strategy")
		}
	}
}
