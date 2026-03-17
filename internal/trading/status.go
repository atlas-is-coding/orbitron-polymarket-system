package trading

import (
	"strings"

	"github.com/atlasdev/orbitron/internal/tui"
)

// WalletProvider is an interface to get wallet labels and addresses.
type WalletProvider interface {
	WalletLabel(id string) string
	WalletAddress(id string) string
}

// walletAwareStrategy is optionally implemented by strategies that track wallets.
type walletAwareStrategy interface {
	WalletIDs() []string
}

// runnableStrategy is optionally implemented by strategies that report running state.
type runnableStrategy interface {
	IsRunning() bool
}

// StrategyDetailsProvider allows extracting details from strategies.
type StrategyDetailsProvider interface {
	Details() string
}

// GetStrategyRows builds a list of strategy status rows for the TUI.
func GetStrategyRows(engine *Engine, wp WalletProvider) []tui.StrategyRow {
	var rows []tui.StrategyRow

	strategies := engine.Strategies()
	for _, s := range strategies {
		var wids []string
		if wa, ok := s.(walletAwareStrategy); ok {
			wids = wa.WalletIDs()
		}

		var labels []string
		var addresses []string
		for _, wid := range wids {
			if wid == "" {
				continue
			}
			wlabel := wp.WalletLabel(wid)
			waddr := wp.WalletAddress(wid)
			if wlabel == "" {
				if waddr != "" {
					wlabel = waddr[:min(len(waddr), 8)] + "…" + waddr[max(0, len(waddr)-4):]
				} else {
					wlabel = wid[:min(len(wid), 8)]
				}
			}
			labels = append(labels, wlabel)

			if waddr != "" {
				addresses = append(addresses, waddr)
			}
		}

		wlabel := "—"
		if len(labels) > 0 {
			wlabel = strings.Join(labels, ", ")
		}
		waddr := "—"
		if len(addresses) > 0 {
			waddr = strings.Join(addresses, ", ")
		}
		wid := ""
		if len(wids) > 0 {
			wid = wids[0]
		}

		status := "off"
		if rs, ok := s.(runnableStrategy); ok && rs.IsRunning() {
			status = "active"
		}

		rows = append(rows, tui.StrategyRow{
			Name:          s.Name(),
			Status:        status,
			WalletID:      wid,
			WalletLabel:   wlabel,
			WalletAddress: waddr,
			Details:       GetDetails(s),
		})
	}

	return rows
}

func GetDetails(s Strategy) string {
	if p, ok := s.(StrategyDetailsProvider); ok {
		return p.Details()
	}
	return ""
}
