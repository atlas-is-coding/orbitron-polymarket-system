package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexus_HandleWalletStats(t *testing.T) {
	n := NewNexus()
	msg := WalletStatsMsg{
		ID:         "w1",
		BalanceUSD: 100.0,
		PnLUSD:     10.0,
	}

	n.Handle(msg)

	bal, pnl := n.GetTotals()
	assert.Equal(t, 100.0, bal)
	assert.Equal(t, 10.0, pnl)

	// Update existing wallet
	msg2 := WalletStatsMsg{
		ID:         "w1",
		BalanceUSD: 150.0,
		PnLUSD:     15.0,
	}
	n.Handle(msg2)
	bal, pnl = n.GetTotals()
	assert.Equal(t, 150.0, bal)
	assert.Equal(t, 15.0, pnl)

	// Add second wallet
	msg3 := WalletStatsMsg{
		ID:         "w2",
		BalanceUSD: 50.0,
		PnLUSD:     -5.0,
	}
	n.Handle(msg3)
	bal, pnl = n.GetTotals()
	assert.Equal(t, 200.0, bal)
	assert.Equal(t, 10.0, pnl)
}

func TestNexus_HandleSubsystemStatus(t *testing.T) {
	n := NewNexus()
	msg := SubsystemStatusMsg{Name: "trading", Active: true}
	n.Handle(msg)

	snap := n.Snapshot()
	subs := snap["subsystems"].(map[string]bool)
	assert.True(t, subs["trading"])

	msg2 := SubsystemStatusMsg{Name: "trading", Active: false}
	n.Handle(msg2)
	snap = n.Snapshot()
	subs = snap["subsystems"].(map[string]bool)
	assert.False(t, subs["trading"])
}

func TestNexus_HandleStrategiesUpdate(t *testing.T) {
	n := NewNexus()
	rows := []StrategyRow{
		{Name: "strat1", Status: "active"},
		{Name: "strat2", Status: "stopped"},
	}
	msg := StrategiesUpdateMsg{Rows: rows}
	n.Handle(msg)

	snap := n.Snapshot()
	strats := snap["strategies"].([]StrategyRow)
	assert.Len(t, strats, 2)
	assert.Equal(t, "active", strats[0].Status)
}

func TestNexus_SnapshotDeepCopy(t *testing.T) {
	n := NewNexus()
	rows := []StrategyRow{{Name: "strat1", Status: "active"}}
	n.Handle(StrategiesUpdateMsg{Rows: rows})

	snap := n.Snapshot()
	strats := snap["strategies"].([]StrategyRow)
	
	// Modify the original rows in Nexus (simulated via another Handle)
	n.Handle(StrategiesUpdateMsg{Rows: []StrategyRow{{Name: "strat1", Status: "stopped"}}})
	
	// Snapshot should still have the old value if it was a deep copy (or at least if it was a copy of the slice)
	assert.Equal(t, "active", strats[0].Status)
}
