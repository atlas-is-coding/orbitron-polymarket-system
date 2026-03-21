package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WalletProvider is implemented by *wallet.Manager (see internal/wallet).
// Using an interface here avoids an import cycle (wallet imports tui).
type WalletProvider interface {
	// List
	WalletIDs() []string
	WalletLabel(id string) string
	WalletAddress(id string) string
	WalletEnabled(id string) bool
	WalletStats(id string) (balanceUSD, pnlUSD float64, openOrders, totalTrades int)

	// Mutations (return error if id not found)
	UpdateLabel(id, label string) error
	Toggle(id string, enabled bool) error
	Remove(id string) error
}

type walletMode int

const (
	walletModeTable         walletMode = iota
	walletModeDetail
	walletModeAddForm
	walletModeEditForm
	walletModeConfirmDelete
)

// walletFormField indices
const (
	wfLabel      = 0
	wfPrivKey    = 1
	wfAPIKey     = 2
	wfAPISecret  = 3
	wfPassphrase = 4
	wfChainID    = 5
	wfCount      = 6
)

// WalletsModel is the Wallets tab sub-model.
type WalletsModel struct {
	wm      WalletProvider
	cfgPath string
	width   int
	height  int

	walletsTable table.Model
	mode         walletMode
	selectedID   string // selected wallet ID for delete confirmation
	detailID     string // wallet being viewed in detail mode
	editID       string // wallet being edited (for editForm)
	formFocus    int    // which input is focused (0..wfCount-1)
	inputs       []textinput.Model
	formErr      string
}

// Resize updates the model dimensions without losing data.
func (m *WalletsModel) Resize(w, h int) {
	m.width = w
	m.height = h
}

// NewWalletsModel creates a new WalletsModel.
func NewWalletsModel(wm WalletProvider, cfgPath string, width, height int) WalletsModel {
	cols := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Label", Width: 16},
		{Title: "Address", Width: 14},
		{Title: "Balance", Width: 10},
		{Title: "P&L", Width: 10},
		{Title: "Status", Width: 8},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(max(height/2-3, 3)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(ColorAccent).
		Background(ColorSurface)
	s.Selected = s.Selected.
		Foreground(ColorBg).
		Background(ColorAccent).
		Bold(true)
	t.SetStyles(s)

	m := WalletsModel{
		wm:           wm,
		cfgPath:      cfgPath,
		width:        width,
		height:       height,
		walletsTable: t,
		inputs:       makeWalletFormInputs(),
	}
	m.refreshTable()
	return m
}

func makeWalletFormInputs() []textinput.Model {
	labels := []string{"Label", "Private Key", "API Key", "API Secret", "Passphrase", "Chain ID"}
	passwords := []bool{false, true, false, true, true, false}
	inputs := make([]textinput.Model, wfCount)
	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = labels[i]
		ti.CharLimit = 256
		ti.PromptStyle = StyleAccent
		ti.Cursor.Style = StyleAccent
		if passwords[i] {
			ti.EchoMode = textinput.EchoPassword
		}
		inputs[i] = ti
	}
	return inputs
}

// refreshTable rebuilds table rows from the WalletProvider.
// Must only be called from Update (pointer context) or constructor.
func (m *WalletsModel) refreshTable() {
	if m.wm == nil {
		return
	}
	ids := m.wm.WalletIDs()
	rows := make([]table.Row, 0, len(ids))
	for i, id := range ids {
		label := m.wm.WalletLabel(id)
		addr := m.wm.WalletAddress(id)
		addrShort := addr
		if len(addrShort) > 10 {
			addrShort = addrShort[:6] + "..." + addrShort[len(addrShort)-4:]
		}
		if addrShort == "" {
			addrShort = "—"
		}
		enabled := m.wm.WalletEnabled(id)
		bal, pnl, _, _ := m.wm.WalletStats(id)

		var status string
		if enabled {
			status = StyleSuccess.Render("● ON")
		} else {
			status = StyleMuted.Render("○ OFF")
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			label,
			addrShort,
			fmt.Sprintf("$%.2f", bal),
			fmt.Sprintf("%+.2f", pnl),
			status,
		})
	}
	m.walletsTable.SetRows(rows)
	// Keep cursor within bounds
	if cur := m.walletsTable.Cursor(); cur >= len(rows) && len(rows) > 0 {
		m.walletsTable.SetCursor(len(rows) - 1)
	}
}

// selectedWalletID returns the wallet ID at the current table cursor position.
func (m WalletsModel) selectedWalletID() string {
	if m.wm == nil {
		return ""
	}
	ids := m.wm.WalletIDs()
	cursor := m.walletsTable.Cursor()
	if cursor < 0 || cursor >= len(ids) {
		return ""
	}
	return ids[cursor]
}

// Init implements tea.Model.
func (m WalletsModel) Init() tea.Cmd { return nil }

// IsEditing reports whether a form field is currently being edited.
// When true, global tab-switching keys are suppressed.
func (m WalletsModel) IsEditing() bool {
	return m.mode == walletModeAddForm || m.mode == walletModeEditForm
}

// Update implements tea.Model.
func (m WalletsModel) Update(msg tea.Msg) (WalletsModel, tea.Cmd) {
	switch m.mode {
	case walletModeAddForm, walletModeEditForm:
		return m.updateForm(msg)
	case walletModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case walletModeDetail:
		return m.updateDetail(msg)
	default:
		return m.updateTable(msg)
	}
}

func (m WalletsModel) updateTable(msg tea.Msg) (WalletsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case WalletAddedMsg, WalletRemovedMsg, WalletChangedMsg, WalletStatsMsg:
		m.refreshTable()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			// Open add form
			m.mode = walletModeAddForm
			m.formFocus = 0
			m.formErr = ""
			m.inputs = makeWalletFormInputs()
			m.inputs[wfChainID].SetValue("137")
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink

		case "e":
			// Open edit form for selected wallet
			id := m.selectedWalletID()
			if id == "" {
				return m, nil
			}
			m.editID = id
			m.mode = walletModeEditForm
			m.formFocus = 0
			m.formErr = ""
			m.inputs = makeWalletFormInputs()
			m.inputs[wfLabel].SetValue(m.wm.WalletLabel(id))
			m.inputs[wfChainID].SetValue("137")
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink

		case "d":
			// Ask for delete confirmation
			id := m.selectedWalletID()
			if id == "" {
				return m, nil
			}
			m.selectedID = id
			m.mode = walletModeConfirmDelete
			return m, nil

		case " ":
			// Toggle enabled state
			id := m.selectedWalletID()
			if id == "" {
				return m, nil
			}
			enabled := m.wm.WalletEnabled(id)
			_ = m.wm.Toggle(id, !enabled)
			m.refreshTable()
			return m, nil

		case "enter":
			// Open detail view
			id := m.selectedWalletID()
			if id == "" {
				return m, nil
			}
			m.detailID = id
			m.mode = walletModeDetail
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.walletsTable, cmd = m.walletsTable.Update(msg)
	return m, cmd
}

func (m WalletsModel) updateDetail(msg tea.Msg) (WalletsModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "q", "enter":
			m.mode = walletModeTable
			return m, nil
		}
	}
	return m, nil
}

func (m WalletsModel) updateConfirmDelete(msg tea.Msg) (WalletsModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "y", "Y":
			if m.selectedID != "" && m.wm != nil {
				_ = m.wm.Remove(m.selectedID)
				m.selectedID = ""
			}
			m.mode = walletModeTable
			m.refreshTable()
			return m, nil
		case "n", "N", "esc":
			m.selectedID = ""
			m.mode = walletModeTable
			return m, nil
		}
	}
	return m, nil
}

func (m WalletsModel) updateForm(msg tea.Msg) (WalletsModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.mode = walletModeTable
			m.inputs[m.formFocus].Blur()
			m.formErr = ""
			return m, nil

		case "tab", "down":
			m.inputs[m.formFocus].Blur()
			m.formFocus = (m.formFocus + 1) % wfCount
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink

		case "shift+tab", "up":
			m.inputs[m.formFocus].Blur()
			m.formFocus = (m.formFocus - 1 + wfCount) % wfCount
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink

		case "enter":
			if m.formFocus < wfCount-1 {
				// Advance to next field
				m.inputs[m.formFocus].Blur()
				m.formFocus++
				m.inputs[m.formFocus].Focus()
				return m, textinput.Blink
			}
			// Last field — save
			return m.saveForm()

		case "ctrl+s":
			return m.saveForm()
		}
	}
	var cmd tea.Cmd
	m.inputs[m.formFocus], cmd = m.inputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m WalletsModel) saveForm() (WalletsModel, tea.Cmd) {
	label := strings.TrimSpace(m.inputs[wfLabel].Value())
	if label == "" {
		m.formErr = "Label is required"
		return m, nil
	}

	if m.mode == walletModeEditForm {
		// Update label in memory; caller's config.Save handles persistence
		if err := m.wm.UpdateLabel(m.editID, label); err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		m.inputs[m.formFocus].Blur()
		m.mode = walletModeTable
		m.formErr = ""
		m.refreshTable()
		return m, nil
	}

	// Add mode: wallet creation requires full subsystem wiring in main.go (AddActive).
	// The form validates inputs; actual instantiation is a future task (AddInactive path).
	// For now, clear the form and return to the table — the user will need to restart
	// the bot with the new wallet entry added to config.toml directly.
	m.inputs[m.formFocus].Blur()
	m.mode = walletModeTable
	m.formErr = ""
	m.refreshTable()
	return m, nil
}

// View implements tea.Model.
func (m WalletsModel) View() string {
	switch m.mode {
	case walletModeAddForm:
		return m.viewForm("Add Wallet", "[ctrl+s] save  [enter] next field  [esc] cancel  [tab] cycle fields")
	case walletModeEditForm:
		return m.viewForm("Edit Wallet", "[ctrl+s] save  [enter] next field  [esc] cancel  [tab] cycle fields")
	case walletModeConfirmDelete:
		return m.viewConfirmDelete()
	case walletModeDetail:
		return m.viewDetail()
	default:
		return m.viewTable()
	}
}

func (m WalletsModel) viewTable() string {
	var content string
	if m.wm == nil || len(m.wm.WalletIDs()) == 0 {
		content = renderEmptyState("◎", "No wallets configured", "Press [a] to add one", m.width)
	} else {
		content = "\n" + m.walletsTable.View()
	}
	tablePanel := renderPanel("Wallets", content, m.width, true)
	helpPanel := renderHelpPanel("[↑↓] navigate   [a] add   [e] edit   [d] delete   [Space] toggle   [Enter] details", m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", tablePanel, " ", helpPanel)
}

func (m WalletsModel) viewDetail() string {
	id := m.detailID
	if m.wm == nil {
		return lipgloss.JoinVertical(lipgloss.Left, " ",
			renderPanel("Wallet Detail", StyleError.Render("  Wallet provider unavailable"), m.width, true))
	}

	label := m.wm.WalletLabel(id)
	addr := m.wm.WalletAddress(id)
	enabled := m.wm.WalletEnabled(id)
	bal, pnl, orders, total := m.wm.WalletStats(id)

	var statusStr string
	if enabled {
		statusStr = StyleSuccess.Render("● ACTIVE")
	} else {
		statusStr = StyleMuted.Render("○ INACTIVE")
	}

	if addr == "" {
		addr = "—"
	}

	var pnlStr string
	if pnl >= 0 {
		pnlStr = StyleSuccess.Render(fmt.Sprintf("%+.2f", pnl))
	} else {
		pnlStr = StyleError.Render(fmt.Sprintf("%+.2f", pnl))
	}

	var sb strings.Builder
	sb.WriteString("\n   " + StyleBold.Render(fmt.Sprintf("Wallet: %s", label)))
	sb.WriteString("  ")
	sb.WriteString(statusStr)
	sb.WriteString("\n   ")
	sb.WriteString(StyleFgDim.Render("Address: "))
	sb.WriteString(addr)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("   %-14s  $%.2f\n", StyleFgDim.Render("Balance:"), bal))
	sb.WriteString(fmt.Sprintf("   %-14s  %s\n", StyleFgDim.Render("P&L:"), pnlStr))
	sb.WriteString(fmt.Sprintf("   %-14s  %d\n", StyleFgDim.Render("Open Orders:"), orders))
	sb.WriteString(fmt.Sprintf("   %-14s  %d\n\n", StyleFgDim.Render("Total Trades:"), total))

	detailPanel := renderPanel("Wallet Detail", sb.String(), m.width, true)
	helpPanel := renderHelpPanel("esc back", m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", detailPanel, " ", helpPanel)
}

func (m WalletsModel) viewForm(title, helpLine string) string {
	fieldNames := []string{"Label", "Private Key", "API Key", "API Secret", "Passphrase", "Chain ID"}

	var sb strings.Builder
	sb.WriteString("\n")

	for i, ti := range m.inputs {
		var prefix string
		var labelStr string
		if i == m.formFocus {
			prefix = StyleAccent.Render(" ▶ ")
			labelStr = StyleBold.Render(fmt.Sprintf("%-14s", fieldNames[i]))
		} else {
			prefix = "   "
			labelStr = StyleFgDim.Render(fmt.Sprintf("%-14s", fieldNames[i]))
		}
		sb.WriteString(prefix)
		sb.WriteString(labelStr)
		sb.WriteString("  ")
		sb.WriteString(ti.View())
		sb.WriteString("\n")
	}

	if m.formErr != "" {
		sb.WriteString("\n")
		sb.WriteString(StyleError.Render("   ✖ " + m.formErr))
	}

	sb.WriteString("\n")
	formPanel := renderPanel(title, sb.String(), m.width, true)
	helpPanel := renderHelpPanel(helpLine, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, " ", formPanel, " ", helpPanel)
}

func (m WalletsModel) viewConfirmDelete() string {
	id := m.selectedID
	label := ""
	if m.wm != nil {
		label = m.wm.WalletLabel(id)
	}
	if label == "" {
		label = id
	}

	content := fmt.Sprintf(
		"\n   Delete wallet %s?\n\n   %s   %s\n",
		StyleWarning.Render(fmt.Sprintf("%q", label)),
		StyleError.Render("[y] Yes, delete"),
		StyleMuted.Render("[n] Cancel"),
	)
	panel := renderPanel("Confirm Delete", content, m.width, true)
	return lipgloss.JoinVertical(lipgloss.Left, " ", panel)
}
