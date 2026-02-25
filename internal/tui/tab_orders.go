package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// OrderRow represents a single row in the orders table.
type OrderRow struct {
	Market string
	Side   string
	Price  string
	Size   string
	Filled string
	Status string
	Age    string
	ID     string
}

// OrdersModel is the Orders tab sub-model.
type OrdersModel struct {
	table  table.Model
	rows   []OrderRow
	width  int
	height int
}

// NewOrdersModel creates a new OrdersModel.
func NewOrdersModel(width, height int) OrdersModel {
	cols := []table.Column{
		{Title: i18n.T().OrdersColMarket, Width: 30},
		{Title: i18n.T().OrdersColSide, Width: 6},
		{Title: i18n.T().OrdersColPrice, Width: 10},
		{Title: i18n.T().OrdersColSize, Width: 10},
		{Title: i18n.T().OrdersColFilled, Width: 10},
		{Title: i18n.T().OrdersColStatus, Width: 10},
		{Title: i18n.T().OrdersColAge, Width: 10},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(max(height-6, 1)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	t.SetStyles(s)
	return OrdersModel{table: t, width: width, height: height}
}

// SetOrderRows updates the orders displayed in the table.
func (m *OrdersModel) SetOrderRows(rows []OrderRow) {
	m.rows = rows
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row{r.Market, r.Side, r.Price, r.Size, r.Filled, r.Status, r.Age}
	}
	m.table.SetRows(tableRows)
}

func (m OrdersModel) Init() tea.Cmd { return nil }

func (m OrdersModel) Update(msg tea.Msg) (OrdersModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d", "D":
			if idx := m.table.Cursor(); idx >= 0 && idx < len(m.rows) {
				id := m.rows[idx].ID
				return m, func() tea.Msg { return CancelOrderMsg{ID: id} }
			}
		case "a", "A":
			return m, func() tea.Msg { return CancelAllOrdersMsg{} }
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m OrdersModel) View() string {
	help := StyleHelpBar.Render(i18n.T().OrdersHelp)
	return lipgloss.JoinVertical(lipgloss.Left, m.table.View(), help)
}

// CancelOrderMsg is emitted when user presses D on a selected order.
type CancelOrderMsg struct{ ID string }

// CancelAllOrdersMsg is emitted when user presses A.
type CancelAllOrdersMsg struct{}
