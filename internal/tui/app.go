package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// AppModel is the root Bubble Tea model for the TUI dashboard.
type AppModel struct {
	activeTab  TabID
	overview   OverviewModel
	orders     OrdersModel
	positions  PositionsModel
	copytrader CopytradingModel
	logs       LogsModel
	settings   SettingsModel

	bus     *EventBus
	cfg     *config.Config
	cfgPath string
	onSave  func(string)
	wallet  string
	width   int
	height  int
}

// NewAppModel creates the root app model.
func NewAppModel(
	cfg *config.Config,
	cfgPath string,
	bus *EventBus,
	width, height int,
	onSave func(string),
) AppModel {
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 40
	}
	cw := max(height-6, 10)

	return AppModel{
		cfg:        cfg,
		cfgPath:    cfgPath,
		onSave:     onSave,
		bus:        bus,
		width:      width,
		height:     height,
		overview:   NewOverviewModel(width, cw),
		orders:     NewOrdersModel(width, cw),
		positions:  NewPositionsModel(width, cw),
		copytrader: NewCopytradingModel(cfg, cfgPath, width, cw),
		logs:       NewLogsModel(width, cw),
		settings:   NewSettingsModel(cfg, cfgPath, width, cw, onSave),
	}
}

// SetWallet sets the wallet address shown in the header.
func (m *AppModel) SetWallet(addr string) {
	m.wallet = addr
}

func (m AppModel) Init() tea.Cmd {
	return m.bus.WaitForEvent()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		cw := max(m.height-6, 10)
		m.logs = NewLogsModel(m.width, cw)
		return m, m.bus.WaitForEvent()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			// Don't switch tabs when a text field is being edited in settings
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			m.activeTab = (m.activeTab + 1) % tabCount
			return m, m.bus.WaitForEvent()
		case "shift+tab":
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			if m.activeTab == 0 {
				m.activeTab = tabCount - 1
			} else {
				m.activeTab--
			}
			return m, m.bus.WaitForEvent()
		case "1", "2", "3", "4", "5", "6":
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			switch msg.String() {
			case "1":
				m.activeTab = TabOverview
			case "2":
				m.activeTab = TabOrders
			case "3":
				m.activeTab = TabPositions
			case "4":
				m.activeTab = TabCopytrading
			case "5":
				m.activeTab = TabLogs
			case "6":
				m.activeTab = TabSettings
			}
		}

	case ConfigReloadedMsg:
		m.cfg = msg.Config
		m.copytrader.cfg = msg.Config
		var cmd tea.Cmd
		m.settings, cmd = m.settings.Update(msg)
		return m, tea.Batch(cmd, m.bus.WaitForEvent())

	case BalanceMsg:
		m.overview, _ = m.overview.Update(msg)
		return m, m.bus.WaitForEvent()

	case SubsystemStatusMsg:
		m.overview, _ = m.overview.Update(msg)
		return m, m.bus.WaitForEvent()

	case BotEventMsg:
		m.logs, _ = m.logs.Update(msg)
		return m, m.bus.WaitForEvent()

	case LanguageChangedMsg:
		cw := max(m.height-6, 10)
		m.orders = NewOrdersModel(m.width, cw)
		m.positions = NewPositionsModel(m.width, cw)
		m.copytrader = NewCopytradingModel(m.cfg, m.cfgPath, m.width, cw)
		// settings NOT rebuilt — FieldDef labels/tooltips use func() string closures,
		// sectionNames() is computed dynamically, and optionIdx is already updated.
		// Rebuilding from m.cfg would reset the Language field to the unsaved (old) value.
		return m, m.bus.WaitForEvent()
	}

	// Route key events to active tab
	var cmd tea.Cmd
	switch m.activeTab {
	case TabOverview:
		m.overview, cmd = m.overview.Update(msg)
	case TabOrders:
		m.orders, cmd = m.orders.Update(msg)
	case TabPositions:
		m.positions, cmd = m.positions.Update(msg)
	case TabCopytrading:
		m.copytrader, cmd = m.copytrader.Update(msg)
	case TabLogs:
		m.logs, cmd = m.logs.Update(msg)
	case TabSettings:
		m.settings, cmd = m.settings.Update(msg)
	}

	return m, tea.Batch(cmd, m.bus.WaitForEvent())
}

func (m AppModel) View() string {
	walletShort := m.wallet
	if len(walletShort) > 12 {
		walletShort = walletShort[:6] + "..." + walletShort[len(walletShort)-4:]
	}
	dot := StyleHeaderDot.Render("●")
	walletPart := ""
	if walletShort != "" {
		walletPart = "  │  " + i18n.T().AppWallet + ": " + walletShort
	}
	header := StyleHeader.Width(m.width).Render(
		fmt.Sprintf(" polytrade-bot  %s %s%s ", dot, i18n.T().AppRunning, walletPart),
	)
	tabBar := RenderTabBar(m.activeTab, m.width)

	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.overview.View()
	case TabOrders:
		content = m.orders.View()
	case TabPositions:
		content = m.positions.View()
	case TabCopytrading:
		content = m.copytrader.View()
	case TabLogs:
		content = m.logs.View()
	case TabSettings:
		content = m.settings.View()
	}

	helpBar := StyleHelpBar.Width(m.width).Render(
		"  " + i18n.T().HelpGlobal + "  ",
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, content, helpBar)
}
