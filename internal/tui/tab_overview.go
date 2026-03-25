package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/health"
	"github.com/atlasdev/orbitron/internal/i18n"
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
	subsystems   []SubsystemStatus
	balance      float64
	openOrders   int
	positions    int
	pnlToday     float64
	traders      int
	wallets      []walletSummaryRow
	width        int
	height       int
	health       health.HealthSnapshot
	healthLoaded bool
	tick         int
}

// NewOverviewModel creates a new OverviewModel.
func NewOverviewModel(width, height int) OverviewModel {
	return OverviewModel{
		width:      width,
		height:     height,
		subsystems: []SubsystemStatus{},
	}
}

// Resize updates the model dimensions without losing data.
func (m *OverviewModel) Resize(w, h int) {
	m.width = w
	m.height = h
	m.rebuildLayout()
}

// rebuildLayout pre-computes layout measurements for View().
func (m *OverviewModel) rebuildLayout() {
	// Currently a hook for future pre-computation; measurements done inline in View().
}

func (m *OverviewModel) LoadSnapshot(snap map[string]any) {
	if bal, ok := snap["balance"].(float64); ok {
		m.balance = bal
	}
	if pnl, ok := snap["pnl"].(float64); ok {
		m.pnlToday = pnl
	}
	if subs, ok := snap["subsystems"].(map[string]bool); ok {
		m.subsystems = make([]SubsystemStatus, 0, len(subs))
		for name, active := range subs {
			m.subsystems = append(m.subsystems, SubsystemStatus{Name: name, Active: active})
		}
	}
	if wallets, ok := snap["wallets"].([]WalletStatsMsg); ok {
		m.wallets = make([]walletSummaryRow, 0, len(wallets))
		for _, w := range wallets {
			m.wallets = append(m.wallets, walletSummaryRow{
				id:      w.ID,
				label:   w.Label,
				enabled: w.Enabled,
				balance: w.BalanceUSD,
				pnl:     w.PnLUSD,
			})
		}
	}
}

func (m OverviewModel) Init() tea.Cmd { return nil }

func (m OverviewModel) Update(msg tea.Msg) (OverviewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SubsystemStatusMsg:
		found := false
		for i, s := range m.subsystems {
			if s.Name == msg.Name {
				m.subsystems[i].Active = msg.Active
				found = true
				break
			}
		}
		if !found {
			m.subsystems = append(m.subsystems, SubsystemStatus{Name: msg.Name, Active: msg.Active})
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

	case HealthSnapshotMsg:
		m.health = msg.Snapshot
		m.healthLoaded = true

	case animTickMsg:
		m.tick++
		return m, nil
	}
	return m, nil
}

func (m OverviewModel) renderHealthBlock() string {
	t := i18n.T()
	var sb strings.Builder

	if !m.healthLoaded {
		sb.WriteString("   " + StyleMuted.Render(t.OverviewHealthNever) + "\n")
		return sb.String()
	}

	statusDot := func(s health.ServiceStatus) string {
		switch s {
		case health.StatusOK:
			return StyleSuccess.Render("●")
		case health.StatusDegraded:
			return StyleWarning.Render("◐")
		default:
			return StyleError.Render("○")
		}
	}
	latStr := func(ms int64) string {
		if ms < 1000 {
			return fmt.Sprintf("%dms", ms)
		}
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}

	for _, svc := range m.health.Services {
		if svc.Name == "Geoblock" {
			continue
		}
		dot := statusDot(svc.Status)
		lat := StyleFgDim.Render(latStr(svc.LatencyMs))
		errStr := ""
		if svc.Error != "" {
			errStr = " " + StyleError.Render(svc.Error)
		}
		fmt.Fprintf(&sb, "   %s %-16s %s%s\n", dot, svc.Name, lat, errStr)
	}

	if m.health.Geo != nil {
		geo := m.health.Geo
		if geo.Blocked {
			geoStr := StyleError.Render(fmt.Sprintf("⚠ %s %s (%s)", t.OverviewGeoBlocked, geo.Country, geo.IP))
			fmt.Fprintf(&sb, "   %s %-16s %s\n", StyleError.Render("○"), "Geoblock", geoStr)
		} else {
			geoStr := StyleSuccess.Render(fmt.Sprintf("%s %s", t.OverviewGeoAllowed, geo.Country))
			fmt.Fprintf(&sb, "   %s %-16s %s\n", StyleSuccess.Render("●"), "Geoblock", geoStr)
		}
	}

	age := int(time.Since(m.health.UpdatedAt).Seconds())
	sb.WriteString("\n   " + StyleMuted.Render(fmt.Sprintf(t.OverviewHealthUpdated, age)) + "\n")
	return sb.String()
}

// formatPnL returns a coloured PnL string.
func formatPnL(v float64) string {
	if v >= 0 {
		return StylePositive.Render(fmt.Sprintf("+$%.2f", v))
	}
	return StyleNegative.Render(fmt.Sprintf("-$%.2f", -v))
}

func (m OverviewModel) View() string {
	t := i18n.T()
	helpPanel := renderHelpPanel("r=refresh | Tab=next-tab | q=quit", m.width)

	bp := breakpoint(m.width)

	// ── Pulsing status dot for tiny/mobile ────────────────────────────────
	statusDot := StyleSuccess.Render("●")
	if m.tick%2 == 1 {
		statusDot = StyleMuted.Render("●")
	}

	switch bp {
	case "tiny":
		// Single column, no panels — plain critical data
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Balance: $%.2f  PnL: %s\n", m.balance, formatPnL(m.pnlToday)))
		sb.WriteString(fmt.Sprintf("Orders: %d  Positions: %d  %s\n", m.openOrders, m.positions, statusDot))
		return lipgloss.JoinVertical(lipgloss.Left, sb.String(), helpPanel)

	case "mobile":
		// Single column panels stacked
		label := StyleFgDim.Render
		val := StyleGlow.Render

		var statsContent strings.Builder
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewBalance)), val(fmt.Sprintf("$%.2f", m.balance)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewOpenOrders)), StyleBold.Render(fmt.Sprintf("%d", m.openOrders)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPositions)), StyleBold.Render(fmt.Sprintf("%d", m.positions)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPnLToday)), formatPnL(m.pnlToday))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewCopyTraders)), StyleBold.Render(fmt.Sprintf("%d", m.traders)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label("Status"), statusDot)
		statsPanel := renderCard(t.OverviewStats, statsContent.String(), m.width, true)
		return lipgloss.JoinVertical(lipgloss.Left, " ", statsPanel, " ", helpPanel)

	case "standard":
		// Two-column layout: left=wallet/balance, right=orders/positions
		half := (m.width - 4) / 2
		label := StyleFgDim.Render
		val := StyleGlow.Render

		var leftContent strings.Builder
		leftContent.WriteString(" " + StyleSidebarLogo.Render("◈ ORBITRON") + "\n")
		leftContent.WriteString(" " + StyleSidebarSubtitle.Render(" NEXUS TERM v1.0") + "\n\n")
		fmt.Fprintf(&leftContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewBalance)), val(fmt.Sprintf("$%.2f", m.balance)))
		fmt.Fprintf(&leftContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPnLToday)), formatPnL(m.pnlToday))
		fmt.Fprintf(&leftContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewCopyTraders)), StyleBold.Render(fmt.Sprintf("%d", m.traders)))
		leftPanel := renderCard(t.OverviewStats, leftContent.String(), half, true)

		var rightContent strings.Builder
		fmt.Fprintf(&rightContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewOpenOrders)), StyleBold.Render(fmt.Sprintf("%d", m.openOrders)))
		fmt.Fprintf(&rightContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPositions)), StyleBold.Render(fmt.Sprintf("%d", m.positions)))
		rightContent.WriteString("\n")
		rightContent.WriteString(" " + StyleSidebarLabel.Render("SUBSYSTEMS") + "\n")
		for _, s := range m.subsystems {
			dot := StyleSuccess.Render("●")
			status := StyleSuccess.Render(t.OverviewActive)
			if !s.Active {
				dot = StyleMuted.Render("○")
				status = StyleMuted.Render(t.OverviewInactive)
			}
			fmt.Fprintf(&rightContent, " %s %-14s %s\n", dot, StyleFgDim.Render(s.Name), status)
		}
		rightPanel := renderCard(t.OverviewHealth, rightContent.String(), half, false)

		topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)
		return lipgloss.JoinVertical(lipgloss.Left, " ", topRow, " ", helpPanel)

	default:
		// Large+ (>140): three-column layout with full detail
		third := (m.width - 4) / 3
		label := StyleFgDim.Render
		val := StyleGlow.Render

		// ── Left: Logo & Subsystems ───────────────────────────────────────────
		var logoSubsystems strings.Builder
		logoSubsystems.WriteString(" " + StyleSidebarLogo.Render("◈ ORBITRON") + "\n")
		logoSubsystems.WriteString(" " + StyleSidebarSubtitle.Render(" NEXUS TERM v1.0") + "\n\n")
		logoSubsystems.WriteString(" " + StyleSidebarLabel.Render("SUBSYSTEMS") + "\n")
		for _, s := range m.subsystems {
			dot := StyleSuccess.Render("●")
			status := StyleSuccess.Render(t.OverviewActive)
			if !s.Active {
				dot = StyleMuted.Render("○")
				status = StyleMuted.Render(t.OverviewInactive)
			}
			fmt.Fprintf(&logoSubsystems, " %s %-16s %s\n", dot, StyleFgDim.Render(s.Name), status)
		}
		leftPanel := renderCard("", logoSubsystems.String(), third, false)

		// ── Middle: Quick Stats ───────────────────────────────────────────────
		var statsContent strings.Builder
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewBalance)), val(fmt.Sprintf("$%.2f", m.balance)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewOpenOrders)), StyleBold.Render(fmt.Sprintf("%d", m.openOrders)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPositions)), StyleBold.Render(fmt.Sprintf("%d", m.positions)))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewPnLToday)), formatPnL(m.pnlToday))
		fmt.Fprintf(&statsContent, " %s  %s\n", label(fmt.Sprintf("%-20s", t.OverviewCopyTraders)), StyleBold.Render(fmt.Sprintf("%d", m.traders)))
		middlePanel := renderCard(t.OverviewStats, statsContent.String(), third, true)

		// ── Right: Health ─────────────────────────────────────────────────────
		rightPanel := renderCard(t.OverviewHealth, m.renderHealthBlock(), third, false)

		topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", middlePanel, " ", rightPanel)

		// ── Bottom: Wallets Table ──────────────────────────────────────────────
		var walletsPanel string
		if len(m.wallets) > 0 {
			var totalBal, totalPnL float64
			activeCount := 0
			for _, w := range m.wallets {
				totalBal += w.balance
				totalPnL += w.pnl
				if w.enabled {
					activeCount++
				}
			}

			var walletsContent strings.Builder
			fmt.Fprintf(&walletsContent, " %s %s  │  %s %s  │  %s %d/%d\n\n",
				label(t.OverviewTotalBalance+":"), val(fmt.Sprintf("$%.2f", totalBal)),
				label(t.OverviewTotalPnL+":"), formatPnL(totalPnL),
				label(t.OverviewActiveWallets+":"), activeCount, len(m.wallets),
			)

			colW := (m.width - 12) / 4
			if colW < 12 {
				colW = 12
			}
			hdr := StyleFgDimBold.Render(fmt.Sprintf(" %-*s  %-14s  %-14s  %s", colW, "LABEL", "BALANCE", "P&L", "STATUS"))
			walletsContent.WriteString(hdr + "\n")
			walletsContent.WriteString(StyleMuted.Render(" "+strings.Repeat("─", m.width-6)) + "\n")

			for _, w := range m.wallets {
				lbl := w.label
				if len(lbl) > colW {
					lbl = lbl[:colW-1] + "…"
				}
				balStr := fmt.Sprintf("$%.2f", w.balance)
				statusStr := StyleMuted.Render("○ OFF")
				if w.enabled {
					statusStr = StyleSuccess.Render("● ON")
				}
				fmt.Fprintf(&walletsContent, " %-*s  %-14s  %-14s  %s\n", colW, lbl, balStr, formatPnL(w.pnl), statusStr)
			}
			walletsPanel = renderCard(t.OverviewWallets, walletsContent.String(), m.width, false)
		}

		if walletsPanel != "" {
			return lipgloss.JoinVertical(lipgloss.Left, " ", topRow, " ", walletsPanel, " ", helpPanel)
		}
		return lipgloss.JoinVertical(lipgloss.Left, " ", topRow, " ", helpPanel)
	}
}

