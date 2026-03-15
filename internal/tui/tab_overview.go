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
}

// NewOverviewModel creates a new OverviewModel.
func NewOverviewModel(width, height int) OverviewModel {
	return OverviewModel{
		width:      width,
		height:     height,
		subsystems: []SubsystemStatus{},
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
		fmt.Fprintf(&sb, "  %s %-14s %s%s\n", dot, svc.Name, lat, errStr)
	}

	if m.health.Geo != nil {
		geo := m.health.Geo
		if geo.Blocked {
			geoStr := StyleError.Render(fmt.Sprintf("⚠ %s %s (%s)", t.OverviewGeoBlocked, geo.Country, geo.IP))
			fmt.Fprintf(&sb, "  %s %-14s %s\n", StyleError.Render("○"), "Geoblock", geoStr)
		} else {
			geoStr := StyleSuccess.Render(fmt.Sprintf("%s %s", t.OverviewGeoAllowed, geo.Country))
			fmt.Fprintf(&sb, "  %s %-14s %s\n", StyleSuccess.Render("●"), "Geoblock", geoStr)
		}
	}

	age := int(time.Since(m.health.UpdatedAt).Seconds())
	sb.WriteString("\n  " + StyleMuted.Render(fmt.Sprintf(t.OverviewHealthUpdated, age)) + "\n")
	return sb.String()
}

func (m OverviewModel) View() string {
	t := i18n.T()
	half := (m.width - 3) / 2

	// ── Left panel: subsystems ─────────────────────────────────────────────
	var leftContent strings.Builder
	for _, s := range m.subsystems {
		dot := StyleSuccess.Render("●")
		name := StyleFgDim.Render(fmt.Sprintf("%-22s", s.Name))
		status := StyleSuccess.Render(t.OverviewActive)
		if !s.Active {
			dot = StyleMuted.Render("○")
			name = StyleMuted.Render(fmt.Sprintf("%-22s", s.Name))
			status = StyleMuted.Render(t.OverviewInactive)
		}
		fmt.Fprintf(&leftContent, "  %s %s %s\n", dot, name, status)
	}
	leftPanel := renderPanel(t.OverviewSubsystems, leftContent.String(), half, false)

	// ── Right panel: quick stats ───────────────────────────────────────────
	label := StyleFgDim.Render
	val := StyleGlow.Render

	pnlStr := fmt.Sprintf("%+.2f", m.pnlToday)
	if m.pnlToday >= 0 {
		pnlStr = StyleSuccess.Render(pnlStr)
	} else {
		pnlStr = StyleError.Render(pnlStr)
	}

	var rightContent strings.Builder
	fmt.Fprintf(&rightContent, "  %s  %s\n", label(fmt.Sprintf("%-22s", t.OverviewBalance)), val(fmt.Sprintf("$%.2f", m.balance)))
	fmt.Fprintf(&rightContent, "  %s  %s\n", label(fmt.Sprintf("%-22s", t.OverviewOpenOrders)), StyleBold.Render(fmt.Sprintf("%d", m.openOrders)))
	fmt.Fprintf(&rightContent, "  %s  %s\n", label(fmt.Sprintf("%-22s", t.OverviewPositions)), StyleBold.Render(fmt.Sprintf("%d", m.positions)))
	fmt.Fprintf(&rightContent, "  %s  %s\n", label(fmt.Sprintf("%-22s", t.OverviewPnLToday)), pnlStr)
	fmt.Fprintf(&rightContent, "  %s  %s\n", label(fmt.Sprintf("%-22s", t.OverviewCopyTraders)), StyleBold.Render(fmt.Sprintf("%d", m.traders)))
	rightPanel := renderPanel(t.OverviewStats, rightContent.String(), half, true)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	// ── Health panel ───────────────────────────────────────────────────────
	healthPanel := renderPanel(t.OverviewHealth, m.renderHealthBlock(), m.width, false)

	// ── Help panel ─────────────────────────────────────────────────────────
	helpPanel := renderHelpPanel("[1-8] switch tab   [Tab] next   [q] quit", m.width)

	// ── Wallets panel ──────────────────────────────────────────────────────
	if len(m.wallets) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, " ", topRow, " ", healthPanel, " ", helpPanel)
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

	totalPnLStr := fmt.Sprintf("%+.2f", totalPnL)
	if totalPnL >= 0 {
		totalPnLStr = StyleSuccess.Render(totalPnLStr)
	} else {
		totalPnLStr = StyleError.Render(totalPnLStr)
	}

	var walletsContent strings.Builder
	fmt.Fprintf(&walletsContent, "  %s %s  │  %s %s  │  %s %d/%d\n\n",
		label(t.OverviewTotalBalance+":"), val(fmt.Sprintf("$%.2f", totalBal)),
		label(t.OverviewTotalPnL+":"), totalPnLStr,
		label(t.OverviewActiveWallets+":"), activeCount, len(m.wallets),
	)

	colW := (m.width - 12) / 4
	if colW < 12 {
		colW = 12
	}
	hdr := StyleFgDim.Bold(true).Render(fmt.Sprintf("  %-*s  %-14s  %-14s  %s", colW, "LABEL", "BALANCE", "P&L", "STATUS"))
	walletsContent.WriteString(hdr + "\n")
	walletsContent.WriteString(StyleMuted.Render("  "+strings.Repeat("─", m.width-8)) + "\n")

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
		statusStr := StyleMuted.Render("○ OFF")
		if w.enabled {
			statusStr = StyleSuccess.Render("● ON")
		}
		fmt.Fprintf(&walletsContent, "  %-*s  %-14s  %-14s  %s\n", colW, lbl, balStr, wPnLStr, statusStr)
	}

	walletsPanel := renderPanel(t.OverviewWallets, walletsContent.String(), m.width, false)
	return lipgloss.JoinVertical(lipgloss.Left, " ", topRow, " ", walletsPanel, " ", healthPanel, " ", helpPanel)
}
