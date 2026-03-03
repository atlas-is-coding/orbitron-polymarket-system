package wallet

import (
	"context"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// dataClient is the subset of data.Client used by the stats poller.
type dataClient interface {
	GetPositions(params data.PositionsParams) ([]data.Position, error)
	GetTrades(params data.TradesParams) ([]data.Trade, error)
}

// RunStatsPoller polls per-wallet stats from the Data API every interval
// and broadcasts WalletStatsMsg via the EventBus.
// It should be launched as a goroutine: go wm.RunStatsPoller(ctx, dataClient, interval).
func (m *Manager) RunStatsPoller(ctx context.Context, dc dataClient, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Poll once immediately on start
	m.pollStats(dc)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.pollStats(dc)
		}
	}
}

func (m *Manager) pollStats(dc dataClient) {
	wallets := m.Wallets()
	for _, inst := range wallets {
		if inst.Address == "" {
			continue
		}
		m.pollWalletStats(inst, dc)
	}
}

func (m *Manager) pollWalletStats(inst *WalletInstance, dc dataClient) {
	addr := inst.Address

	// --- Open orders from TradesMonitor (cheapest: already cached) ---
	openOrders := 0
	totalTrades := 0
	if inst.TradesMon != nil {
		openOrders = len(inst.TradesMon.GetOrders())
		totalTrades = len(inst.TradesMon.GetTrades())
	} else if inst.ClobClient != nil {
		// Fallback: direct CLOB call
		if orders, err := inst.ClobClient.GetOrders(); err == nil {
			openOrders = len(orders.Data)
		}
	}

	// --- Positions P&L and current value from Data API ---
	var totalBalance float64
	var totalPnL float64
	positions, err := dc.GetPositions(data.PositionsParams{User: addr, SizeThreshold: 0.01})
	if err == nil {
		for _, p := range positions {
			totalBalance += p.CurrentValue
			totalPnL += p.CashPnl
		}
	}

	// Update in-memory stats
	inst.Stats.Set(totalBalance, totalPnL, openOrders, totalTrades)

	// Broadcast to TUI
	if m.bus != nil {
		m.bus.Send(tui.WalletStatsMsg{
			ID:          inst.Cfg.ID,
			Label:       inst.Cfg.Label,
			Enabled:     inst.Cfg.Enabled,
			BalanceUSD:  totalBalance,
			PnLUSD:      totalPnL,
			OpenOrders:  openOrders,
			TotalTrades: totalTrades,
		})
	}
}
