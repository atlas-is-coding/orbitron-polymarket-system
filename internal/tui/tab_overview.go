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

// OverviewModel is the Overview tab sub-model.
type OverviewModel struct {
	subsystems []SubsystemStatus
	balance    float64
	openOrders int
	positions  int
	pnlToday   float64
	traders    int
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

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}
