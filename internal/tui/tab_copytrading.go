package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
)

// TraderRow is a row in the tracked traders table.
type TraderRow struct {
	Address  string
	Label    string
	Status   string
	AllocPct string
}

type copyMode int

const (
	copyModeTable copyMode = iota
	copyModeAddForm
	copyModeEditForm
	copyModeConfirmDelete
)

// Resize updates the model dimensions without losing data.
func (m *CopytradingModel) Resize(w, h int) {
	m.width = w
	m.height = h
	m.tradersTable.SetHeight(max(h-10, 3))
}

// CopytradingModel is the Copytrading tab sub-model.
type CopytradingModel struct {
	tradersTable table.Model
	recentTrades []string
	width        int
	height       int
	tick         int

	cfg     *config.Config
	cfgPath string

	mode      copyMode
	editIdx   int               // index in cfg.Copytrading.Traders for edit/delete
	formFocus int               // which textinput is active (0-3)
	inputs    []textinput.Model // Address, Label, Alloc%, MaxPositionUSD
	formErr   string            // last save error
}

// NewCopytradingModel creates a new CopytradingModel.
func NewCopytradingModel(cfg *config.Config, cfgPath string, width, height int) CopytradingModel {
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
	s.Header = StyleTableHeader
	s.Selected = StyleTableSelected
	t.SetStyles(s)

	return CopytradingModel{
		tradersTable: t,
		width:        width,
		height:       height,
		cfg:          cfg,
		cfgPath:      cfgPath,
		inputs:       makeFormInputs(),
	}
}

// makeFormInputs creates the four textinputs for the add/edit form.
func makeFormInputs() []textinput.Model {
	placeholders := []string{"0x… wallet address", "label (optional)", "alloc % (e.g. 5)", "max position USD (e.g. 50)"}
	inputs := make([]textinput.Model, 4)
	for i, ph := range placeholders {
		ti := textinput.New()
		ti.Placeholder = ph
		ti.CharLimit = 80
		ti.PromptStyle = StyleAccent
		ti.Cursor.Style = StyleAccent
		inputs[i] = ti
	}
	inputs[0].Focus()
	return inputs
}

// IsEditing reports whether the model is in form-entry mode (blocks global tab switching).
func (m CopytradingModel) IsEditing() bool {
	return m.mode != copyModeTable
}

// SetTraderRows updates the traders table.
func (m *CopytradingModel) SetTraderRows(rows []TraderRow) {
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		icon := copytradingStatusIcon(r.Status, m.tick)
		tableRows[i] = table.Row{r.Address, r.Label, icon + " " + r.Status, r.AllocPct}
	}
	m.tradersTable.SetRows(tableRows)
}

// copytradingStatusIcon returns the spec §7.3 status symbol for the given status string.
func copytradingStatusIcon(status string, tick int) string {
	switch strings.ToUpper(status) {
	case "ACTIVE", "RUNNING":
		return StylePositive.Render("●")
	case "PENDING", "PAUSED":
		return StyleWarning.Render("◆")
	case "FAILED", "STOPPED", "ERROR":
		return StyleNegative.Render("✕")
	case "LOADING", "CONNECTING":
		if tick%2 == 0 {
			return StyleValue.Render("→")
		}
		return StyleMuted.Render("→")
	default:
		return StyleMuted.Render("○")
	}
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
	switch msg.(type) {
	case animTickMsg:
		m.tick++
		return m, nil
	}
	switch m.mode {
	case copyModeTable:
		return m.updateTable(msg)
	case copyModeAddForm, copyModeEditForm:
		return m.updateForm(msg)
	case copyModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	}
	return m, nil
}

func (m CopytradingModel) updateTable(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, CopyKeys.Add):
			m.mode = copyModeAddForm
			m.inputs = makeFormInputs()
			m.formFocus = 0
			m.formErr = ""
			return m, textinput.Blink
		case key.Matches(msg, CopyKeys.Edit):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			m.editIdx = idx
			tr := m.cfg.Copytrading.Traders[idx]
			m.inputs = makeFormInputs()
			m.inputs[0].SetValue(tr.Address)
			m.inputs[0].Blur() // address not editable in edit mode
			m.inputs[1].SetValue(tr.Label)
			m.inputs[2].SetValue(strconv.FormatFloat(tr.AllocationPct, 'f', -1, 64))
			m.inputs[3].SetValue(strconv.FormatFloat(tr.MaxPositionUSD, 'f', -1, 64))
			m.formFocus = 1
			m.inputs[1].Focus()
			m.mode = copyModeEditForm
			m.formErr = ""
			return m, textinput.Blink
		case key.Matches(msg, CopyKeys.Delete):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			m.editIdx = idx
			m.mode = copyModeConfirmDelete
			return m, nil
		case key.Matches(msg, CopyKeys.Toggle):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			addr := m.cfg.Copytrading.Traders[idx].Address
			if err := toggleTrader(m.cfg, addr); err != nil {
				return m, nil
			}
			_ = config.Save(m.cfgPath, m.cfg)
			m.syncTable()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.tradersTable, cmd = m.tradersTable.Update(msg)
	return m, cmd
}

func (m CopytradingModel) updateForm(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = copyModeTable
			m.formErr = ""
			return m, nil
		case "tab", "down":
			m.inputs[m.formFocus].Blur()
			start := 0
			if m.mode == copyModeEditForm {
				start = 1 // address not editable
			}
			m.formFocus++
			if m.formFocus > 3 {
				m.formFocus = start
			}
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.inputs[m.formFocus].Blur()
			start := 0
			if m.mode == copyModeEditForm {
				start = 1
			}
			m.formFocus--
			if m.formFocus < start {
				m.formFocus = 3
			}
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink
		case "enter":
			if m.formFocus < 3 {
				// advance to next field
				m.inputs[m.formFocus].Blur()
				m.formFocus++
				m.inputs[m.formFocus].Focus()
				return m, textinput.Blink
			}
			// last field — save
			return m.saveForm()
		}
	}
	// Route to focused input
	var cmd tea.Cmd
	m.inputs[m.formFocus], cmd = m.inputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m CopytradingModel) updateConfirmDelete(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.cfg != nil && m.editIdx < len(m.cfg.Copytrading.Traders) {
				addr := m.cfg.Copytrading.Traders[m.editIdx].Address
				_ = removeTrader(m.cfg, addr)
				_ = config.Save(m.cfgPath, m.cfg)
				m.syncTable()
			}
			m.mode = copyModeTable
		default:
			m.mode = copyModeTable
		}
	}
	return m, nil
}

// saveForm validates inputs, calls addTrader or editTrader, saves config.
func (m CopytradingModel) saveForm() (CopytradingModel, tea.Cmd) {
	addr := strings.TrimSpace(m.inputs[0].Value())
	label := strings.TrimSpace(m.inputs[1].Value())
	allocStr := strings.TrimSpace(m.inputs[2].Value())
	maxStr := strings.TrimSpace(m.inputs[3].Value())

	allocPct := 5.0
	if allocStr != "" {
		if v, err := strconv.ParseFloat(allocStr, 64); err == nil {
			allocPct = v
		}
	}
	maxPos := 50.0
	if maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxPos = v
		}
	}

	var err error
	if m.mode == copyModeAddForm {
		err = addTrader(m.cfg, addr, label, allocPct, maxPos)
	} else {
		if m.editIdx < len(m.cfg.Copytrading.Traders) {
			addr = m.cfg.Copytrading.Traders[m.editIdx].Address
		}
		err = editTrader(m.cfg, addr, label, allocPct, maxPos)
	}
	if err != nil {
		m.formErr = err.Error()
		return m, nil
	}
	if saveErr := config.Save(m.cfgPath, m.cfg); saveErr != nil {
		m.formErr = saveErr.Error()
		return m, nil
	}
	m.syncTable()
	m.mode = copyModeTable
	m.formErr = ""
	return m, nil
}

// syncTable rebuilds table rows from cfg.
func (m *CopytradingModel) syncTable() {
	if m.cfg == nil {
		return
	}
	rows := make([]TraderRow, len(m.cfg.Copytrading.Traders))
	for i, t := range m.cfg.Copytrading.Traders {
		status := "disabled"
		if t.Enabled {
			status = "active"
		}
		rows[i] = TraderRow{
			Address:  t.Address,
			Label:    t.Label,
			Status:   status,
			AllocPct: strconv.FormatFloat(t.AllocationPct, 'f', 1, 64) + "%",
		}
	}
	m.SetTraderRows(rows)
}

func (m CopytradingModel) View() string {
	switch m.mode {
	case copyModeAddForm:
		return m.viewForm("Add Trader")
	case copyModeEditForm:
		return m.viewForm("Edit Trader")
	case copyModeConfirmDelete:
		return m.viewConfirmDelete()
	}
	return m.viewTable()
}

func (m CopytradingModel) viewTable() string {
	t := i18n.T()

	var tradersContent string
	if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
		tradersContent = renderEmptyState("⇌", "No traders configured", "Press [a] to add one", m.width)
	} else {
		tradersContent = "\n" + m.tradersTable.View()
	}
	tradersPanel := renderPanel(t.CopyTraders, tradersContent, m.width, true)

	var tradesContent strings.Builder
	tradesContent.WriteString("\n")
	if len(m.recentTrades) == 0 {
		tradesContent.WriteString("  " + StyleMuted.Render(t.CopyNoData) + "\n")
	} else {
		for _, tr := range m.recentTrades {
			tradesContent.WriteString("  " + tr + "\n")
		}
	}
	tradesPanel := renderPanel(t.CopyRecentTrades, tradesContent.String(), m.width, false)

	// Detail bar for the selected trader
	detailBar := ""
	if m.cfg != nil {
		rows := m.tradersTable.Rows()
		idx := m.tradersTable.Cursor()
		if idx >= 0 && idx < len(rows) && idx < len(m.cfg.Copytrading.Traders) {
			tr := m.cfg.Copytrading.Traders[idx]
			addr := tr.Address
			if len(addr) > 12 {
				addr = addr[:6] + "..." + addr[len(addr)-4:]
			}
			detailBar = fmt.Sprintf("Trader: %s | Following: $%.0f | PnL: %s | Trades: %s",
				StyleValue.Render(addr),
				tr.MaxPositionUSD,
				StylePositive.Render("+$0.00"),
				StyleMuted.Render("0"),
			)
		}
	}

	helpPanel := renderHelpPanel("↑↓=navigate | f=follow | u=unfollow | Tab=next-tab | q=quit", m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", tradersPanel, " ", tradesPanel, " ", StyleMuted.Render(detailBar), " ", helpPanel)
}

func (m CopytradingModel) viewForm(title string) string {
	labels := []string{"Address:    ", "Label:      ", "Alloc %:    ", "Max Pos USD:"}
	var sb strings.Builder
	sb.WriteString("\n")
	for i, inp := range m.inputs {
		prefix := "   "
		if m.formFocus == i {
			prefix = StyleAccent.Render(" ▶ ")
		}
		sb.WriteString(prefix + StyleMuted.Render(labels[i]) + " " + inp.View() + "\n")
	}
	if m.formErr != "" {
		sb.WriteString("\n   " + StyleError.Render("✖ "+m.formErr) + "\n")
	}
	sb.WriteString("\n")
	formPanel := renderPanel(title, sb.String(), m.width, true)
	helpPanel := renderHelpPanel("[Enter] save   [Tab] next field   [esc] cancel", m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", formPanel, " ", helpPanel)
}

func (m CopytradingModel) viewConfirmDelete() string {
	addr := ""
	if m.cfg != nil && m.editIdx < len(m.cfg.Copytrading.Traders) {
		addr = m.cfg.Copytrading.Traders[m.editIdx].Address
	}
	content := fmt.Sprintf(
		"\n   Delete trader %s?\n\n   %s   %s\n",
		StyleWarning.Render(addr),
		StyleError.Render("[y] Yes, delete"),
		StyleMuted.Render("[n] Cancel"),
	)
	panel := renderPanel("Confirm Delete", content, m.width, true)
	return lipgloss.JoinVertical(lipgloss.Left, " ", panel)
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
