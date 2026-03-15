package gamma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/api/gamma"
	"github.com/atlasdev/orbitron/internal/testutil"
)

func TestGetMarkets_ReturnsResults(t *testing.T) {
	client := testutil.NewGammaClient()
	active := true
	markets, err := client.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 5})
	require.NoError(t, err)
	require.NotEmpty(t, markets)
	m := markets[0]
	assert.NotEmpty(t, m.ConditionID)
	assert.NotEmpty(t, m.Question)
}

func TestGetMarkets_WithLimit(t *testing.T) {
	client := testutil.NewGammaClient()
	markets, err := client.GetMarkets(gamma.MarketsParams{Limit: 3})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(markets), 3)
}

func TestGetMarket_ByConditionID(t *testing.T) {
	client := testutil.NewGammaClient()
	active := true
	markets, err := client.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 1})
	require.NoError(t, err)
	require.NotEmpty(t, markets)
	condID := markets[0].ConditionID
	m, err := client.GetMarket(condID)
	require.NoError(t, err)
	assert.Equal(t, condID, m.ConditionID)
}

func TestGetEvents_ReturnsResults(t *testing.T) {
	client := testutil.NewGammaClient()
	active := true
	events, err := client.GetEvents(gamma.EventsParams{Active: &active, Limit: 5})
	require.NoError(t, err)
	require.NotEmpty(t, events)
	assert.NotEmpty(t, events[0].ID)
}

func TestGetEvent_ByID(t *testing.T) {
	client := testutil.NewGammaClient()
	events, err := client.GetEvents(gamma.EventsParams{Limit: 1})
	require.NoError(t, err)
	require.NotEmpty(t, events)
	ev, err := client.GetEvent(events[0].ID)
	require.NoError(t, err)
	assert.Equal(t, events[0].ID, ev.ID)
}

func TestFlexFloat64_NonNegative(t *testing.T) {
	client := testutil.NewGammaClient()
	markets, err := client.GetMarkets(gamma.MarketsParams{Limit: 10})
	require.NoError(t, err)
	for _, m := range markets {
		assert.GreaterOrEqual(t, float64(m.Volume), 0.0)
		assert.GreaterOrEqual(t, float64(m.Liquidity), 0.0)
	}
}
