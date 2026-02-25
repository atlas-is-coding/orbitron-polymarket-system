package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// PositionRow represents a single row in the positions table.
type PositionRow struct {
	Market  string
	Side    string
	Size    string
	Entry   string
	Current string
	PnL     string
	PnLPct  string
}

// PositionsModel is the Positions tab sub-model.
type PositionsModel struct {
	table  table.Model
	width  int
	height int
}

// NewPositionsModel creates a new PositionsModel.
func NewPositionsModel(width, height int) PositionsModel {
	cols := []table.Column{
		{Title: i18n.T().PosColMarket, Width: 30},
		{Title: i18n.T().PosColSide, Width: 6},
		{Title: i18n.T().PosColSize, Width: 10},
		{Title: i18n.T().PosColEntry, Width: 10},
		{Title: i18n.T().PosColCurrent, Width: 10},
		{Title: i18n.T().PosColPnL, Width: 12},
		{Title: i18n.T().PosColPnLPct, Width: 8},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(max(height-6, 1)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	t.SetStyles(s)
	return PositionsModel{table: t, width: width, height: height}
}

// SetPositionRows updates positions in the table.
func (m *PositionsModel) SetPositionRows(rows []PositionRow) {
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row{r.Market, r.Side, r.Size, r.Entry, r.Current, r.PnL, r.PnLPct}
	}
	m.table.SetRows(tableRows)
}

func (m PositionsModel) Init() tea.Cmd { return nil }

func (m PositionsModel) Update(msg tea.Msg) (PositionsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m PositionsModel) View() string {
	help := StyleHelpBar.Render(i18n.T().PosHelp)
	return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), help)
}
