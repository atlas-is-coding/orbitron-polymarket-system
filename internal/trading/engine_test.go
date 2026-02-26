package trading_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/trading"
)

type fakeStrategy struct {
	name    string
	started atomic.Bool
	stopped atomic.Bool
	err     error
}

func (f *fakeStrategy) Name() string { return f.name }
func (f *fakeStrategy) Start(ctx context.Context) error {
	f.started.Store(true)
	<-ctx.Done()
	return f.err
}
func (f *fakeStrategy) Stop() error {
	f.stopped.Store(true)
	return nil
}

type immediateErrStrategy struct{ err error }

func (i *immediateErrStrategy) Name() string                  { return "immediate-err" }
func (i *immediateErrStrategy) Start(_ context.Context) error { return i.err }
func (i *immediateErrStrategy) Stop() error                   { return nil }

func TestEngine_NoStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	err := engine.Start(context.Background())
	assert.NoError(t, err)
}

func TestEngine_StartsAndStops(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s := &fakeStrategy{name: "test-strategy"}
	engine.Register(s)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, s.started.Load())
}

func TestEngine_MultipleStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s1 := &fakeStrategy{name: "strategy-1"}
	s2 := &fakeStrategy{name: "strategy-2"}
	engine.Register(s1)
	engine.Register(s2)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, s1.started.Load())
	assert.True(t, s2.started.Load())
}

func TestEngine_StrategyError_Propagates(t *testing.T) {
	expectedErr := errors.New("strategy boom")
	engine := trading.NewEngine(zerolog.Nop())
	engine.Register(&immediateErrStrategy{err: expectedErr})
	err := engine.Start(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr), "expected error to wrap %v, got: %v", expectedErr, err)
}

func TestEngine_Stop_CallsStopOnStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s := &fakeStrategy{name: "stop-test"}
	engine.Register(s)
	engine.Stop()
	assert.True(t, s.stopped.Load())
}
