package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

const toastDuration = 3 * time.Second

// toastEntry holds an active toast notification.
type toastEntry struct {
	text    string
	kind    string
	created time.Time
}

// AppModel is the root Bubble Tea model for the TUI dashboard.
type AppModel struct {
	activeTab  TabID
	overview   OverviewModel
	trading    TradingModel
	wallets    WalletsModel
	copytrader CopytradingModel
	markets    MarketsModel
	logs       LogsModel
	settings   SettingsModel

	bus     *EventBus
	cfg     *config.Config
	cfgPath string
	onSave  func(string)
	wallet  string
	width   int
	height  int

	// Clock
	now time.Time

	// Toast overlay
	toast *toastEntry
}

// NewAppModel creates the root app model.
func NewAppModel(
	cfg *config.Config,
	cfgPath string,
	bus *EventBus,
	width, height int,
	onSave func(string),
	wm WalletProvider,
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
		now:        time.Now(),
		overview:   NewOverviewModel(width, cw),
		trading:    NewTradingModel(width, cw),
		wallets:    NewWalletsModel(wm, cfgPath, width, cw),
		copytrader: NewCopytradingModel(cfg, cfgPath, width, cw),
		markets:    NewMarketsModel(nil, ""),
		logs:       NewLogsModel(width, cw),
		settings:   NewSettingsModel(cfg, cfgPath, width, cw, onSave),
	}
}

// SetWallet sets the wallet address shown in the header.
func (m *AppModel) SetWallet(addr string) {
	m.wallet = addr
}

func clockTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return clockTickMsg{} })
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.bus.WaitForEvent(), clockTick())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clockTickMsg:
		m.now = time.Now()
		if m.toast != nil && time.Since(m.toast.created) >= toastDuration {
			m.toast = nil
		}
		return m, tea.Batch(clockTick(), m.bus.WaitForEvent())

	case ToastMsg:
		m.toast = &toastEntry{text: msg.Text, kind: msg.Kind, created: time.Now()}
		return m, m.bus.WaitForEvent()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		cw := max(m.height-6, 10)
		m.logs = NewLogsModel(m.width, cw)
		m.markets, _ = m.markets.Update(msg)
		return m, m.bus.WaitForEvent()

	case MarketsUpdatedMsg:
		var cmd tea.Cmd
		m.markets, cmd = m.markets.Update(msg)
		return m, tea.Batch(cmd, m.bus.WaitForEvent())

	case PlaceOrderMsg:
		// Actual order execution is wired in a later task.
		return m, m.bus.WaitForEvent()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			// Don't switch tabs when a text field is being edited
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			if m.activeTab == TabWallets && m.wallets.IsEditing() {
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
			if m.activeTab == TabWallets && m.wallets.IsEditing() {
				break
			}
			if m.activeTab == 0 {
				m.activeTab = tabCount - 1
			} else {
				m.activeTab--
			}
			return m, m.bus.WaitForEvent()
		case "1", "2", "3", "4", "5", "6", "7":
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			if m.activeTab == TabWallets && m.wallets.IsEditing() {
				break
			}
			switch msg.String() {
			case "1":
				m.activeTab = TabOverview
			case "2":
				m.activeTab = TabTrading
			case "3":
				m.activeTab = TabWallets
			case "4":
				m.activeTab = TabCopytrading
			case "5":
				m.activeTab = TabMarkets
			case "6":
				m.activeTab = TabLogs
			case "7":
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

	case WalletAddedMsg, WalletRemovedMsg, WalletChangedMsg, WalletStatsMsg:
		m.wallets, _ = m.wallets.Update(msg)
		m.overview, _ = m.overview.Update(msg)
		return m, m.bus.WaitForEvent()

	case OrdersUpdateMsg:
		m.trading.SetOrderRows(msg.Rows)
		return m, m.bus.WaitForEvent()

	case PositionsUpdateMsg:
		m.trading.SetPositionRows(msg.Rows)
		return m, m.bus.WaitForEvent()

	case LanguageChangedMsg:
		cw := max(m.height-6, 10)
		m.trading = NewTradingModel(m.width, cw)
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
	case TabTrading:
		m.trading, cmd = m.trading.Update(msg)
	case TabWallets:
		m.wallets, cmd = m.wallets.Update(msg)
	case TabCopytrading:
		m.copytrader, cmd = m.copytrader.Update(msg)
	case TabMarkets:
		m.markets, cmd = m.markets.Update(msg)
	case TabLogs:
		m.logs, cmd = m.logs.Update(msg)
	case TabSettings:
		m.settings, cmd = m.settings.Update(msg)
	}

	return m, tea.Batch(cmd, m.bus.WaitForEvent())
}

func (m AppModel) View() string {
	header := m.renderHeader()
	tabBar := RenderTabBar(m.activeTab, m.width)

	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.overview.View()
	case TabTrading:
		content = m.trading.View()
	case TabWallets:
		content = m.wallets.View()
	case TabCopytrading:
		content = m.copytrader.View()
	case TabMarkets:
		content = m.markets.View()
	case TabLogs:
		content = m.logs.View()
	case TabSettings:
		content = m.settings.View()
	}

	helpBar := StyleHelpBar.Width(m.width).Render(
		"  " + i18n.T().HelpGlobal + "  ",
	)

	base := lipgloss.JoinVertical(lipgloss.Left, header, tabBar, content, helpBar)

	if m.toast != nil {
		return m.overlayToast(base)
	}
	return base
}

func (m AppModel) renderHeader() string {
	walletShort := m.wallet
	if len(walletShort) > 12 {
		walletShort = walletShort[:6] + "..." + walletShort[len(walletShort)-4:]
	}

	dot := StyleHeaderDot.Render("●")
	live := StyleHeaderDot.Render("LIVE")
	logo := StyleHeaderGlow.Render("◈ POLYTRADE")
	clock := StyleHeaderMuted.Render(m.now.UTC().Format("15:04:05 UTC"))

	walletPart := ""
	if walletShort != "" {
		walletPart = StyleHeaderMuted.Render("  │  " + i18n.T().AppWallet + ": " + walletShort)
	}

	content := fmt.Sprintf(" %s  %s %s%s  %s  Chain:Polygon ", logo, dot, live, walletPart, clock)
	return StyleHeader.Width(m.width).Render(content)
}

func (m AppModel) overlayToast(base string) string {
	var s lipgloss.Style
	var prefix string
	switch m.toast.kind {
	case "success":
		s = StyleToastSuccess
		prefix = "✓ "
	case "error":
		s = StyleToastError
		prefix = "✗ "
	case "warning":
		s = StyleToastWarning
		prefix = "⚠ "
	default:
		s = StyleToastInfo
		prefix = "◈ "
	}
	toast := s.Render(prefix + m.toast.text)
	return base + "\n" + lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Right).
		Render(toast)
}
