package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/i18n"
	"github.com/atlasdev/orbitron/internal/ui"
)

// TradingSubTab identifies which sub-tab is active inside Trading.
type TradingSubTab int

const (
	SubTabOrders    TradingSubTab = iota
	SubTabPositions               // switch with o/p keys
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

// CancelOrderMsg is emitted when user presses D on a selected order.
type CancelOrderMsg struct{ ID string }

// CancelAllOrdersMsg is emitted when user presses A.
type CancelAllOrdersMsg struct{}

// TradingModel is the Trading tab sub-model (Orders + Positions with sub-tabs).
type TradingModel struct {
	subTab         TradingSubTab
	orders         table.Model
	orderRows      []OrderRow
	positions      table.Model
	width          int
	height         int
	cancelDebounce *ui.Debouncer
}

// Resize updates the model dimensions without losing data.
func (m *TradingModel) Resize(w, h int) {
	m.width = w
	m.height = h
}

// NewTradingModel creates a new TradingModel.
func NewTradingModel(width, height int) TradingModel {
	tableH := max(height-8, 1)

	// Orders table
	orderCols := []table.Column{
		{Title: i18n.T().OrdersColMarket, Width: 28},
		{Title: i18n.T().OrdersColSide, Width: 6},
		{Title: i18n.T().OrdersColPrice, Width: 10},
		{Title: i18n.T().OrdersColSize, Width: 10},
		{Title: i18n.T().OrdersColFilled, Width: 10},
		{Title: i18n.T().OrdersColStatus, Width: 10},
		{Title: i18n.T().OrdersColAge, Width: 10},
	}
	ot := table.New(
		table.WithColumns(orderCols),
		table.WithFocused(true),
		table.WithHeight(tableH),
	)
	os := table.DefaultStyles()
	os.Header = os.Header.
		Bold(true).
		Foreground(ColorAccent).
		Background(ColorSurface)
	os.Selected = os.Selected.
		Foreground(ColorBg).
		Background(ColorAccent).
		Bold(true)
	ot.SetStyles(os)

	// Positions table
	posCols := []table.Column{
		{Title: i18n.T().PosColMarket, Width: 28},
		{Title: i18n.T().PosColSide, Width: 6},
		{Title: i18n.T().PosColSize, Width: 10},
		{Title: i18n.T().PosColEntry, Width: 10},
		{Title: i18n.T().PosColCurrent, Width: 10},
		{Title: i18n.T().PosColPnL, Width: 12},
		{Title: i18n.T().PosColPnLPct, Width: 8},
	}
	pt := table.New(
		table.WithColumns(posCols),
		table.WithFocused(false),
		table.WithHeight(tableH),
	)
	ps := table.DefaultStyles()
	ps.Header = ps.Header.
		Bold(true).
		Foreground(ColorAccent).
		Background(ColorSurface)
	ps.Selected = ps.Selected.
		Foreground(ColorBg).
		Background(ColorAccent).
		Bold(true)
	pt.SetStyles(ps)

	return TradingModel{
		subTab:         SubTabOrders,
		orders:         ot,
		positions:      pt,
		width:          width,
		height:         height,
		cancelDebounce: ui.NewDebouncer(200 * time.Millisecond),
	}
}

// SetOrderRows updates the orders table data.
func (m *TradingModel) SetOrderRows(rows []OrderRow) {
	m.orderRows = rows
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row{r.Market, r.Side, r.Price, r.Size, r.Filled, r.Status, r.Age}
	}
	m.orders.SetRows(tableRows)
}

// SetPositionRows updates the positions table data.
func (m *TradingModel) SetPositionRows(rows []PositionRow) {
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row{r.Market, r.Side, r.Size, r.Entry, r.Current, r.PnL, r.PnLPct}
	}
	m.positions.SetRows(tableRows)
}

// SetStrategyRows is a no-op here since strategies were moved to a dedicated tab.
func (m *TradingModel) SetStrategyRows(rows []StrategyRow) {}

func (m TradingModel) Init() tea.Cmd { return nil }

func (m TradingModel) Update(msg tea.Msg) (TradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "o", "O":
			m.subTab = SubTabOrders
			m.orders.Focus()
			m.positions.Blur()
			return m, nil
		case "p", "P":
			m.subTab = SubTabPositions
			m.positions.Focus()
			m.orders.Blur()
			return m, nil
		case "x", "X":
			if !m.cancelDebounce.Allow() {
				return m, nil // ignore rapid presses
			}
			if m.subTab == SubTabOrders {
				if idx := m.orders.Cursor(); idx >= 0 && idx < len(m.orderRows) {
					id := m.orderRows[idx].ID
					return m, func() tea.Msg { return CancelOrderMsg{ID: id} }
				}
			}
			return m, nil
		case "ctrl+x":
			if !m.cancelDebounce.Allow() {
				return m, nil // ignore rapid presses
			}
			if m.subTab == SubTabOrders {
				return m, func() tea.Msg { return CancelAllOrdersMsg{} }
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.subTab == SubTabOrders {
		m.orders, cmd = m.orders.Update(msg)
	} else if m.subTab == SubTabPositions {
		m.positions, cmd = m.positions.Update(msg)
	}
	return m, cmd
}

func (m TradingModel) View() string {
	t := i18n.T()

	// ── Inline sub-tab selector ─────────────────────────────────────────────
	ordersLabel := fmt.Sprintf(" o: %s ", t.TabOrders)
	posLabel := fmt.Sprintf(" p: %s ", t.TabPositions)
	var subTabLine string
	if m.subTab == SubTabOrders {
		subTabLine = StyleSubTabActive.Render(ordersLabel) + " " + StyleSubTabInactive.Render(posLabel)
	} else {
		subTabLine = StyleSubTabInactive.Render(ordersLabel) + " " + StyleSubTabActive.Render(posLabel)
	}

	// ── Content + help ──────────────────────────────────────────────────────
	var content string
	var helpKeys string
	if m.subTab == SubTabOrders {
		if len(m.orderRows) == 0 {
			content = renderEmptyState("◈", t.OrdersEmpty, "", m.width)
		} else {
			content = m.orders.View()
		}
		helpKeys = "[↑↓] navigate   [x] cancel order   [ctrl+x] cancel all   [o/p] switch"
	} else {
		if len(m.positions.Rows()) == 0 {
			content = renderEmptyState("◈", t.PosEmpty, "", m.width)
		} else {
			content = m.positions.View()
		}
		helpKeys = "[↑↓] navigate   [o/p] switch"
	}

	tablePanel := renderPanel("", content, m.width, true)
	helpPanel := renderHelpPanel(helpKeys, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " " + subTabLine, " ", tablePanel, " ", helpPanel)
}
