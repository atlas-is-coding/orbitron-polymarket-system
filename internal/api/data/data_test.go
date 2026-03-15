package data_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/orbitron/internal/api/data"
	"github.com/atlasdev/orbitron/internal/testutil"
)

const knownTraderAddr = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

func TestGetPositions_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	positions, err := client.GetPositions(data.PositionsParams{User: knownTraderAddr, Limit: 5})
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetPositions_ZeroAddress(t *testing.T) {
	client := testutil.NewDataClient()
	positions, err := client.GetPositions(data.PositionsParams{
		User: "0x0000000000000000000000000000000000000000",
	})
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetClosedPositions_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	positions, err := client.GetClosedPositions(knownTraderAddr, 5, 0)
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetTrades_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	trades, err := client.GetTrades(data.TradesParams{User: knownTraderAddr, Limit: 5})
	require.NoError(t, err)
	assert.NotNil(t, trades)
}
