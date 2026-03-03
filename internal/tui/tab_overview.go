package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// SubsystemStatus holds the name and active state of a bot subsystem.
type SubsystemStatus struct {
	Name   string
	Active bool
}

// walletSummaryRow holds per-wallet stats for the Overview display.
type walletSummaryRow struct {
	id      string
	label   string
	enabled bool
	balance float64
	pnl     float64
}

// OverviewModel is the Overview tab sub-model.
type OverviewModel struct {
	subsystems []SubsystemStatus
	balance    float64
	openOrders int
	positions  int
	pnlToday   float64
	traders    int
	wallets    []walletSummaryRow
	width      int
	height     int
}

// NewOverviewModel creates a new OverviewModel.
func NewOverviewModel(width, height int) OverviewModel {
	return OverviewModel{
		width:  width,
		height: height,
		subsystems: []SubsystemStatus{
			{Name: "WebSocket"},
			{Name: "Monitor"},
			{Name: "Trades Monitor"},
			{Name: "Trading Engine"},
			{Name: "Copytrading"},
		},
	}
}

func (m OverviewModel) Init() tea.Cmd { return nil }

func (m OverviewModel) Update(msg tea.Msg) (OverviewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SubsystemStatusMsg:
		for i, s := range m.subsystems {
			if s.Name == msg.Name {
				m.subsystems[i].Active = msg.Active
			}
		}
	case BalanceMsg:
		m.balance = msg.USDC

	case WalletAddedMsg:
		// Avoid duplicates
		for _, w := range m.wallets {
			if w.id == msg.ID {
				return m, nil
			}
		}
		m.wallets = append(m.wallets, walletSummaryRow{
			id:      msg.ID,
			label:   msg.Label,
			enabled: msg.Enabled,
		})

	case WalletRemovedMsg:
		for i, w := range m.wallets {
			if w.id == msg.ID {
				m.wallets = append(m.wallets[:i], m.wallets[i+1:]...)
				break
			}
		}

	case WalletChangedMsg:
		for i, w := range m.wallets {
			if w.id == msg.ID {
				m.wallets[i].enabled = msg.Enabled
				break
			}
		}

	case WalletStatsMsg:
		found := false
		for i, w := range m.wallets {
			if w.id == msg.ID {
				m.wallets[i].label = msg.Label
				m.wallets[i].enabled = msg.Enabled
				m.wallets[i].balance = msg.BalanceUSD
				m.wallets[i].pnl = msg.PnLUSD
				found = true
				break
			}
		}
		if !found {
			m.wallets = append(m.wallets, walletSummaryRow{
				id:      msg.ID,
				label:   msg.Label,
				enabled: msg.Enabled,
				balance: msg.BalanceUSD,
				pnl:     msg.PnLUSD,
			})
		}
	}
	return m, nil
}

func (m OverviewModel) View() string {
	half := m.width / 2

	// Left: subsystems
	var left strings.Builder
	left.WriteString(StyleSectionTitle.Render(i18n.T().OverviewSubsystems) + "\n")
	for _, s := range m.subsystems {
		dot := StyleSuccess.Render("●")
		name := StyleFgDim.Render(fmt.Sprintf("%-20s", s.Name))
		status := StyleSuccess.Render(i18n.T().OverviewActive)
		if !s.Active {
			dot = StyleMuted.Render("○")
			name = StyleMuted.Render(fmt.Sprintf("%-20s", s.Name))
			status = StyleMuted.Render(i18n.T().OverviewInactive)
		}
		fmt.Fprintf(&left, " %s %s %s\n", dot, name, status)
	}

	// Right: quick stats
	var right strings.Builder
	right.WriteString(StyleSectionTitle.Render(i18n.T().OverviewStats) + "\n")

	label := StyleFgDim.Render
	val := StyleBold.Render

	fmt.Fprintf(&right, " %s  %s\n", label(fmt.Sprintf("%-22s", i18n.T().OverviewBalance)), val(fmt.Sprintf("%.2f", m.balance)))
	fmt.Fprintf(&right, " %s  %s\n", label(fmt.Sprintf("%-22s", i18n.T().OverviewOpenOrders)), val(fmt.Sprintf("%d", m.openOrders)))
	fmt.Fprintf(&right, " %s  %s\n", label(fmt.Sprintf("%-22s", i18n.T().OverviewPositions)), val(fmt.Sprintf("%d", m.positions)))

	pnlStr := fmt.Sprintf("%+.2f", m.pnlToday)
	if m.pnlToday >= 0 {
		pnlStr = StyleSuccess.Render(pnlStr)
	} else {
		pnlStr = StyleError.Render(pnlStr)
	}
	fmt.Fprintf(&right, " %s  %s\n", label(fmt.Sprintf("%-22s", i18n.T().OverviewPnLToday)), pnlStr)
	fmt.Fprintf(&right, " %s  %s\n", label(fmt.Sprintf("%-22s", i18n.T().OverviewCopyTraders)), val(fmt.Sprintf("%d", m.traders)))

	leftBox := StyleBorderActive.Width(half - 2).Render(left.String())
	rightBox := StyleBorder.Width(half - 2).Render(right.String())
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	// Bottom: wallet summary (only when wallets are registered)
	if len(m.wallets) == 0 {
		return topRow
	}

	var totalBal, totalPnL float64
	activeCount := 0
	for _, w := range m.wallets {
		totalBal += w.balance
		totalPnL += w.pnl
		if w.enabled {
			activeCount++
		}
	}

	t := i18n.T()
	var wb strings.Builder
	wb.WriteString(StyleSectionTitle.Render(t.OverviewWallets) + "\n")

	// Aggregate line
	totalPnLStr := fmt.Sprintf("%+.2f", totalPnL)
	if totalPnL >= 0 {
		totalPnLStr = StyleSuccess.Render(totalPnLStr)
	} else {
		totalPnLStr = StyleError.Render(totalPnLStr)
	}
	fmt.Fprintf(&wb, " %s %s  │  %s %s  │  %s %d/%d\n",
		label(t.OverviewTotalBalance+":")+" ", val(fmt.Sprintf("$%.2f", totalBal)),
		label(t.OverviewTotalPnL+":")+" ", totalPnLStr,
		label(t.OverviewActiveWallets+":")+" ", activeCount, len(m.wallets),
	)
	wb.WriteString("\n")

	// Per-wallet rows
	colW := (m.width - 8) / 4
	if colW < 10 {
		colW = 10
	}
	hdr := StyleFgDim.Render(
		fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			colW, "Label",
			12, "Balance",
			12, "P&L",
			"Status",
		),
	)
	wb.WriteString(hdr + "\n")

	for _, w := range m.wallets {
		lbl := w.label
		if len(lbl) > colW {
			lbl = lbl[:colW-1] + "…"
		}
		balStr := fmt.Sprintf("$%.2f", w.balance)
		wPnLStr := fmt.Sprintf("%+.2f", w.pnl)
		if w.pnl >= 0 {
			wPnLStr = StyleSuccess.Render(wPnLStr)
		} else {
			wPnLStr = StyleError.Render(wPnLStr)
		}
		var statusStr string
		if w.enabled {
			statusStr = StyleSuccess.Render("● ON")
		} else {
			statusStr = StyleMuted.Render("○ OFF")
		}
		fmt.Fprintf(&wb, "  %-*s  %-12s  %-12s  %s\n",
			colW, lbl,
			balStr,
			wPnLStr,
			statusStr,
		)
	}

	walletsBox := StyleBorder.Width(m.width - 4).Render(wb.String())
	return lipgloss.JoinVertical(lipgloss.Left, topRow, walletsBox)
}
