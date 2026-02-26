package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

func baseCfg() *config.Config {
	return &config.Config{
		Copytrading: config.CopytradingConfig{
			Traders: []config.TraderConfig{
				{Address: "0xAAA", Label: "alice", Enabled: true, AllocationPct: 5.0, MaxPositionUSD: 50.0},
				{Address: "0xBBB", Label: "bob", Enabled: false, AllocationPct: 10.0, MaxPositionUSD: 100.0},
			},
		},
	}
}

func TestAddTrader_AppendNew(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "0xCCC", "carol", 3.0, 30.0)
	require.NoError(t, err)
	require.Len(t, cfg.Copytrading.Traders, 3)
	tr := cfg.Copytrading.Traders[2]
	assert.Equal(t, "0xCCC", tr.Address)
	assert.Equal(t, "carol", tr.Label)
	assert.InDelta(t, 3.0, tr.AllocationPct, 0.001)
	assert.InDelta(t, 30.0, tr.MaxPositionUSD, 0.001)
	assert.True(t, tr.Enabled)
}

func TestAddTrader_DuplicateAddress(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "0xAAA", "dup", 5.0, 50.0)
	require.Error(t, err)
	assert.Len(t, cfg.Copytrading.Traders, 2)
}

func TestAddTrader_EmptyAddress(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "", "nobody", 5.0, 50.0)
	require.Error(t, err)
}

func TestRemoveTrader_Existing(t *testing.T) {
	cfg := baseCfg()
	err := removeTrader(cfg, "0xAAA")
	require.NoError(t, err)
	require.Len(t, cfg.Copytrading.Traders, 1)
	assert.Equal(t, "0xBBB", cfg.Copytrading.Traders[0].Address)
}

func TestRemoveTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := removeTrader(cfg, "0xZZZ")
	require.Error(t, err)
}

func TestToggleTrader_EnablesDisabled(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xBBB")
	require.NoError(t, err)
	assert.True(t, cfg.Copytrading.Traders[1].Enabled)
}

func TestToggleTrader_DisablesEnabled(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xAAA")
	require.NoError(t, err)
	assert.False(t, cfg.Copytrading.Traders[0].Enabled)
}

func TestToggleTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xZZZ")
	require.Error(t, err)
}

func TestEditTrader_UpdatesFields(t *testing.T) {
	cfg := baseCfg()
	err := editTrader(cfg, "0xAAA", "ALICE", 7.5, 75.0)
	require.NoError(t, err)
	tr := cfg.Copytrading.Traders[0]
	assert.Equal(t, "ALICE", tr.Label)
	assert.InDelta(t, 7.5, tr.AllocationPct, 0.001)
	assert.InDelta(t, 75.0, tr.MaxPositionUSD, 0.001)
}

func TestEditTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := editTrader(cfg, "0xZZZ", "x", 5.0, 50.0)
	require.Error(t, err)
}
