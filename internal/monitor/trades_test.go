//go:build integration

package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/api"
	"github.com/atlasdev/orbitron/internal/api/clob"
	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/monitor"
	"github.com/atlasdev/orbitron/internal/notify"
	"github.com/atlasdev/orbitron/internal/testutil"
	"github.com/atlasdev/orbitron/internal/tui"
)

func TestTradesMonitor_SinglePoll(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)

	clobHTTP := api.NewClient(testutil.ClobURL, 10, 1)
	dataHTTP := api.NewClient(testutil.DataURL, 10, 1)
	clobClient := clob.NewClient(clobHTTP, creds)
	dataClient := data.NewClient(dataHTTP)

	cfg := &config.TradesMonitorConfig{
		PollIntervalMs: 5000,
		TrackPositions: true,
		TradesLimit:    10,
	}

	tm := monitor.NewTradesMonitor(nil, clobClient, dataClient, &notify.NoopNotifier{}, cfg, zerolog.Nop())

	bus := tui.NewEventBus()
	tap := bus.Tap()
	tm.SetBus(bus)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go tm.Run(ctx)

	gotUpdate := false
	deadline := time.After(12 * time.Second)
	for !gotUpdate {
		select {
		case msg := <-tap:
			switch msg.(type) {
			case tui.OrdersUpdateMsg:
				gotUpdate = true
				t.Log("received OrdersUpdateMsg")
			case tui.PositionsUpdateMsg:
				gotUpdate = true
				t.Log("received PositionsUpdateMsg")
			}
		case <-deadline:
			t.Fatal("did not receive any update from TradesMonitor within timeout")
		}
	}
	require.True(t, gotUpdate)
	assert.True(t, gotUpdate)
}
