package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StrategyProvider allows managing strategies from the TUI.
type StrategyProvider interface {
	StartStrategy(name string) error
	StopStrategy(name string) error
	SetStrategyWallets(name string, walletIDs []string) error
	AvailableWallets() []string // returns list of wallet IDs
}

// StrategiesModel is the sub-model for the Strategies tab.
type StrategiesModel struct {
	table        table.Model
	rows         []StrategyRow
	width        int
	height       int
	tick         int
	provider     StrategyProvider
	walletPicker bool
}

// Resize updates the model dimensions without losing data.
func (m *StrategiesModel) Resize(w, h int) {
	m.width = w
	m.height = h
	m.table.SetHeight(max(h-10, 3))
	m.table.SetWidth(w - 4)
}

func NewStrategiesModel(width, height int, provider StrategyProvider) StrategiesModel {
	tableH := max(height-6, 1)

	cols := []table.Column{
		{Title: "Strategy", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Wallet", Width: 15},
		{Title: "Details", Width: 60},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(tableH),
	)

	s := table.DefaultStyles()
	s.Header = StyleTableHeader
	s.Selected = StyleTableSelected
	t.SetStyles(s)

	return StrategiesModel{
		table:    t,
		width:    width,
		height:   height,
		provider: provider,
	}
}

func (m *StrategiesModel) SetRows(rows []StrategyRow) {
	m.rows = rows
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		status := StyleMuted.Render("○ " + r.Status)
		if r.Status == "active" {
			status = StyleSuccess.Render("● " + r.Status)
		}
		tableRows[i] = table.Row{r.Name, status, r.WalletLabel, r.Details}
	}
	m.table.SetRows(tableRows)
}

func (m StrategiesModel) Init() tea.Cmd { return nil }

func (m StrategiesModel) Update(msg tea.Msg) (StrategiesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		m.tick++
		return m, nil
	case tea.KeyMsg:
		if m.walletPicker {
			// Handle wallet selection logic here or via messages
			if msg.String() == "esc" {
				m.walletPicker = false
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "space":
			if idx := m.table.Cursor(); idx >= 0 && idx < len(m.rows) {
				strat := m.rows[idx]
				if strat.Status == "active" {
					return m, func() tea.Msg { return StopStrategyMsg{Name: strat.Name} }
				} else {
					return m, func() tea.Msg { return StartStrategyMsg{Name: strat.Name} }
				}
			}
		case "w", "W":
			if idx := m.table.Cursor(); idx >= 0 && idx < len(m.rows) {
				strat := m.rows[idx]
				return m, func() tea.Msg { return CycleStrategyWalletMsg{Name: strat.Name} }
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m StrategiesModel) View() string {
	var content string
	if len(m.rows) == 0 {
		content = renderEmptyState("◈", "No strategies registered", "", m.width)
	} else {
		content = m.table.View()
	}

	tablePanel := renderPanel("", content, m.width, true)

	var detailLine string
	if idx := m.table.Cursor(); idx >= 0 && idx < len(m.rows) {
		r := m.rows[idx]
		detailLine = " " + StyleMuted.Render("Selected:") + " " + StyleValue.Render(r.Name) +
			"  " + StyleMuted.Render("Status:") + " " + StyleValue.Render(r.Status) +
			"  " + StyleMuted.Render("Wallet:") + " " + StyleValue.Render(r.WalletLabel) +
			"\n"
	}

	helpPanel := renderHelpPanel("↑↓=navigate | Enter=enable/disable | Tab=next-tab | q=quit", m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", tablePanel, detailLine, helpPanel)
}

// Messages for strategy management
type StartStrategyMsg struct{ Name string }
type StopStrategyMsg struct{ Name string }
type CycleStrategyWalletMsg struct{ Name string }
