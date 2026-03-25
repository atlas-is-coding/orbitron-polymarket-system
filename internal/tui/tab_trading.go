package tui

import (
	"fmt"
	"strings"
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
	Market string `json:"market"`
	Side   string `json:"side"`
	Price  string `json:"price"`
	Size   string `json:"size"`
	Filled string `json:"filled"`
	Status string `json:"status"`
	Age    string `json:"age"`
	ID     string `json:"id"`
}

// PositionRow represents a single row in the positions table.
type PositionRow struct {
	Market  string `json:"market"`
	Side    string `json:"side"`
	Size    string `json:"size"`
	Entry   string `json:"entry"`
	Current string `json:"current"`
	PnL     string `json:"pnl"`
	PnLPct  string `json:"pnl_pct"`
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
	positionRows   []PositionRow
	width          int
	height         int
	tick           int
	cancelDebounce *ui.Debouncer
}

// Resize updates the model dimensions and rebuilds table columns/heights.
func (m *TradingModel) Resize(w, h int) {
	m.width = w
	m.height = h
	m.rebuildTables()
}

// rebuildTables resizes inner tables based on current width/height and breakpoint.
func (m *TradingModel) rebuildTables() {
	tableH := max(m.height-10, 1)
	bp := breakpoint(m.width)

	var orderCols []table.Column
	var posCols []table.Column

	switch bp {
	case "tiny": // ≤80: show minimal columns
		orderCols = []table.Column{
			{Title: i18n.T().OrdersColPrice, Width: 10},
			{Title: i18n.T().OrdersColSize, Width: 10},
			{Title: i18n.T().OrdersColStatus, Width: 10},
		}
		posCols = []table.Column{
			{Title: i18n.T().PosColSide, Width: 6},
			{Title: i18n.T().PosColSize, Width: 10},
			{Title: i18n.T().PosColPnL, Width: 12},
		}
	case "mobile": // ≤100: show 3-4 columns
		orderCols = []table.Column{
			{Title: i18n.T().OrdersColSide, Width: 6},
			{Title: i18n.T().OrdersColPrice, Width: 10},
			{Title: i18n.T().OrdersColSize, Width: 10},
			{Title: i18n.T().OrdersColStatus, Width: 10},
		}
		posCols = []table.Column{
			{Title: i18n.T().PosColSide, Width: 6},
			{Title: i18n.T().PosColSize, Width: 10},
			{Title: i18n.T().PosColEntry, Width: 10},
			{Title: i18n.T().PosColPnL, Width: 12},
		}
	case "standard": // ≤140: all main columns
		orderCols = []table.Column{
			{Title: i18n.T().OrdersColMarket, Width: 20},
			{Title: i18n.T().OrdersColSide, Width: 6},
			{Title: i18n.T().OrdersColPrice, Width: 10},
			{Title: i18n.T().OrdersColSize, Width: 10},
			{Title: i18n.T().OrdersColFilled, Width: 10},
			{Title: i18n.T().OrdersColStatus, Width: 10},
		}
		posCols = []table.Column{
			{Title: i18n.T().PosColMarket, Width: 20},
			{Title: i18n.T().PosColSide, Width: 6},
			{Title: i18n.T().PosColSize, Width: 10},
			{Title: i18n.T().PosColEntry, Width: 10},
			{Title: i18n.T().PosColCurrent, Width: 10},
			{Title: i18n.T().PosColPnL, Width: 12},
		}
	default: // large/xl: all columns including optional
		orderCols = []table.Column{
			{Title: i18n.T().OrdersColMarket, Width: 28},
			{Title: i18n.T().OrdersColSide, Width: 6},
			{Title: i18n.T().OrdersColPrice, Width: 10},
			{Title: i18n.T().OrdersColSize, Width: 10},
			{Title: i18n.T().OrdersColFilled, Width: 10},
			{Title: i18n.T().OrdersColStatus, Width: 10},
			{Title: i18n.T().OrdersColAge, Width: 10},
		}
		posCols = []table.Column{
			{Title: i18n.T().PosColMarket, Width: 28},
			{Title: i18n.T().PosColSide, Width: 6},
			{Title: i18n.T().PosColSize, Width: 10},
			{Title: i18n.T().PosColEntry, Width: 10},
			{Title: i18n.T().PosColCurrent, Width: 10},
			{Title: i18n.T().PosColPnL, Width: 12},
			{Title: i18n.T().PosColPnLPct, Width: 8},
		}
	}

	m.orders.SetColumns(orderCols)
	m.orders.SetHeight(tableH)
	m.positions.SetColumns(posCols)
	m.positions.SetHeight(tableH)

	// Re-apply rows with new column count
	m.SetOrderRows(m.orderRows)
	m.SetPositionRows(m.positionRows)
}

// colorStatus returns a status string rendered with the appropriate style.
func colorStatus(status string) string {
	switch strings.ToUpper(status) {
	case "OPEN", "FILLED", "MATCHED":
		return StylePositive.Render(status)
	case "CANCELLED", "CANCELED", "FAILED":
		return StyleNegative.Render(status)
	case "PENDING":
		return StyleWarning.Render(status)
	default:
		return StyleBody.Render(status)
	}
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
	os.Header = StyleTableHeader
	os.Selected = StyleTableSelected
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
	ps.Header = StyleTableHeader
	ps.Selected = StyleTableSelected
	pt.SetStyles(ps)

	return TradingModel{
		subTab:         SubTabOrders,
		orders:         ot,
		orderRows:      nil,
		positions:      pt,
		positionRows:   nil,
		width:          width,
		height:         height,
		cancelDebounce: ui.NewDebouncer(200 * time.Millisecond),
	}
}

// SetOrderRows updates the orders table data, adapting to current breakpoint column count.
func (m *TradingModel) SetOrderRows(rows []OrderRow) {
	m.orderRows = rows
	bp := breakpoint(m.width)
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		switch bp {
		case "tiny":
			tableRows[i] = table.Row{r.Price, r.Size, r.Status}
		case "mobile":
			tableRows[i] = table.Row{r.Side, r.Price, r.Size, r.Status}
		case "standard":
			tableRows[i] = table.Row{r.Market, r.Side, r.Price, r.Size, r.Filled, r.Status}
		default:
			tableRows[i] = table.Row{r.Market, r.Side, r.Price, r.Size, r.Filled, r.Status, r.Age}
		}
	}
	m.orders.SetRows(tableRows)
}

// SetPositionRows updates the positions table data, adapting to current breakpoint column count.
func (m *TradingModel) SetPositionRows(rows []PositionRow) {
	m.positionRows = rows
	bp := breakpoint(m.width)
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		switch bp {
		case "tiny":
			tableRows[i] = table.Row{r.Side, r.Size, r.PnL}
		case "mobile":
			tableRows[i] = table.Row{r.Side, r.Size, r.Entry, r.PnL}
		case "standard":
			tableRows[i] = table.Row{r.Market, r.Side, r.Size, r.Entry, r.Current, r.PnL}
		default:
			tableRows[i] = table.Row{r.Market, r.Side, r.Size, r.Entry, r.Current, r.PnL, r.PnLPct}
		}
	}
	m.positions.SetRows(tableRows)
}

// SetStrategyRows is a no-op here since strategies were moved to a dedicated tab.
func (m *TradingModel) SetStrategyRows(rows []StrategyRow) {}

func (m TradingModel) Init() tea.Cmd { return nil }

func (m TradingModel) Update(msg tea.Msg) (TradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		m.tick++
		return m, nil
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

	// ── Content + detail bar + help ─────────────────────────────────────────
	var content string
	var detailBar string
	if m.subTab == SubTabOrders {
		if len(m.orderRows) == 0 {
			content = renderEmptyState("◈", t.OrdersEmpty, "", m.width)
		} else {
			content = m.orders.View()
			// Selected-row detail bar
			if idx := m.orders.Cursor(); idx >= 0 && idx < len(m.orderRows) {
				r := m.orderRows[idx]
				detailBar = StyleMuted.Render("Selected: ") +
					StyleValue.Render(fmt.Sprintf("Order #%s", r.ID)) +
					StyleMuted.Render(" | Side: ") + StyleValue.Render(r.Side) +
					StyleMuted.Render(" | Price: ") + StyleValue.Render("$"+r.Price) +
					StyleMuted.Render(" | Size: ") + StyleValue.Render(r.Size+" USDC") +
					StyleMuted.Render(" | Status: ") + colorStatus(r.Status)
			}
		}
	} else {
		if len(m.positions.Rows()) == 0 {
			content = renderEmptyState("◈", t.PosEmpty, "", m.width)
		} else {
			content = m.positions.View()
			// Selected-row detail bar
			if idx := m.positions.Cursor(); idx >= 0 && idx < len(m.positionRows) {
				r := m.positionRows[idx]
				detailBar = StyleMuted.Render("Selected: ") +
					StyleValue.Render(r.Market) +
					StyleMuted.Render(" | Side: ") + StyleValue.Render(r.Side) +
					StyleMuted.Render(" | Size: ") + StyleValue.Render(r.Size) +
					StyleMuted.Render(" | Entry: ") + StyleValue.Render("$"+r.Entry) +
					StyleMuted.Render(" | PnL: ") + StyleValue.Render(r.PnL)
			}
		}
	}

	tablePanel := renderPanel("", content, m.width, true)
	helpPanel := renderHelpPanel("↑↓=navigate | Tab=switch-tab | x=cancel | q=quit", m.width)

	parts := []string{" " + subTabLine, " ", tablePanel}
	if detailBar != "" {
		parts = append(parts, " "+detailBar)
	}
	parts = append(parts, " ", helpPanel)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
