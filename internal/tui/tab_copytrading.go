package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// TraderRow is a row in the tracked traders table.
type TraderRow struct {
	Address  string
	Label    string
	Status   string
	AllocPct string
}

// CopytradingModel is the Copytrading tab sub-model.
type CopytradingModel struct {
	tradersTable table.Model
	recentTrades []string
	width        int
	height       int
}

// NewCopytradingModel creates a new CopytradingModel.
func NewCopytradingModel(width, height int) CopytradingModel {
	cols := []table.Column{
		{Title: i18n.T().CopyColAddress, Width: 20},
		{Title: i18n.T().CopyColLabel, Width: 18},
		{Title: i18n.T().CopyColStatus, Width: 10},
		{Title: i18n.T().CopyColAlloc, Width: 8},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(max(height/2-3, 1)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	t.SetStyles(s)
	return CopytradingModel{tradersTable: t, width: width, height: height}
}

// SetTraderRows updates the traders table.
func (m *CopytradingModel) SetTraderRows(rows []TraderRow) {
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row{r.Address, r.Label, r.Status, r.AllocPct}
	}
	m.tradersTable.SetRows(tableRows)
}

// AddTrade appends a line to the recent trades feed (keeps last 20).
func (m *CopytradingModel) AddTrade(line string) {
	m.recentTrades = append(m.recentTrades, line)
	if len(m.recentTrades) > 20 {
		m.recentTrades = m.recentTrades[len(m.recentTrades)-20:]
	}
}

func (m CopytradingModel) Init() tea.Cmd { return nil }

func (m CopytradingModel) Update(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	var cmd tea.Cmd
	m.tradersTable, cmd = m.tradersTable.Update(msg)
	return m, cmd
}

func (m CopytradingModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleBold.Render(i18n.T().CopyTraders) + "\n")
	sb.WriteString(m.tradersTable.View() + "\n\n")
	sb.WriteString(StyleBold.Render(i18n.T().CopyRecentTrades) + "\n")
	if len(m.recentTrades) == 0 {
		sb.WriteString(StyleMuted.Render("  " + i18n.T().CopyNoData + "\n"))
	}
	for _, t := range m.recentTrades {
		sb.WriteString("  " + t + "\n")
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(sb.String())
}

// addTrader appends a new trader to cfg. Returns error if address is empty or already exists.
func addTrader(cfg *config.Config, address, label string, allocPct, maxPositionUSD float64) error {
	if address == "" {
		return fmt.Errorf("address is required")
	}
	for _, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			return fmt.Errorf("trader %q already exists", address)
		}
	}
	cfg.Copytrading.Traders = append(cfg.Copytrading.Traders, config.TraderConfig{
		Address:        address,
		Label:          label,
		Enabled:        true,
		AllocationPct:  allocPct,
		MaxPositionUSD: maxPositionUSD,
		SizeMode:       cfg.Copytrading.SizeMode,
	})
	return nil
}

// removeTrader removes the trader with the given address. Returns error if not found.
func removeTrader(cfg *config.Config, address string) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders = append(cfg.Copytrading.Traders[:i], cfg.Copytrading.Traders[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}

// toggleTrader flips the Enabled flag on the trader with the given address.
func toggleTrader(cfg *config.Config, address string) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders[i].Enabled = !t.Enabled
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}

// editTrader updates label, allocPct, and maxPositionUSD for the trader with the given address.
func editTrader(cfg *config.Config, address, label string, allocPct, maxPositionUSD float64) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders[i].Label = label
			cfg.Copytrading.Traders[i].AllocationPct = allocPct
			cfg.Copytrading.Traders[i].MaxPositionUSD = maxPositionUSD
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}
