package trading

import (
	"strings"

	"github.com/atlasdev/orbitron/internal/tui"
)

// WalletProvider is an interface to get wallet labels.
type WalletProvider interface {
	WalletLabel(id string) string
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
		for _, wid := range wids {
			if wid == "" {
				continue
			}
			wlabel := wp.WalletLabel(wid)
			if wlabel == "" {
				wlabel = wid[:min(len(wid), 8)]
			}
			labels = append(labels, wlabel)
		}

		wlabel := "—"
		if len(labels) > 0 {
			wlabel = strings.Join(labels, ", ")
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
			Name:        s.Name(),
			Status:      status,
			WalletID:    wid,
			WalletLabel: wlabel,
			Details:     GetDetails(s),
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
