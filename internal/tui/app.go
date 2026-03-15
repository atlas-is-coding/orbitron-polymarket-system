package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/config"
)

const toastDuration = 3 * time.Second

// toastEntry holds an active toast notification.
type toastEntry struct {
	text    string
	kind    string
	created time.Time
}

// TradingProvider combines wallet and strategy management.
type TradingProvider interface {
	WalletProvider
	StrategyProvider
	CancelOrder(id string) error
	CancelAllOrders() error
}

// AppModel is the root Bubble Tea model for the TUI dashboard.
type AppModel struct {
	activeTab  TabID
	overview   OverviewModel
	trading    TradingModel
	strategies StrategiesModel
	wallets    WalletsModel
	copytrader CopytradingModel
	markets    MarketsModel
	logs       LogsModel
	settings   SettingsModel

	bus     *EventBus
	cfg     *config.Config
	cfgPath string
	onSave  func(string)
	width   int
	height  int

	// Clock
	now time.Time

	// Toast overlay
	toast *toastEntry

	// updateBanner is non-empty when a newer bot version is available.
	updateBanner string
}

// contentWidth returns the usable width for tab content (terminal - sidebar - border).
func (m AppModel) contentWidth() int {
	return max(m.width-sidebarWidth-1, 20)
}

// contentHeight returns the usable height for tab content (terminal - status bar).
func (m AppModel) contentHeight() int {
	return max(m.height-1, 10)
}

// NewAppModel creates the root app model.
func NewAppModel(
	cfg *config.Config,
	cfgPath string,
	bus *EventBus,
	width, height int,
	onSave func(string),
	tp TradingProvider,
) AppModel {
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 40
	}
	cw := max(width-sidebarWidth-1, 20)
	ch := max(height-1, 10)

	return AppModel{
		cfg:        cfg,
		cfgPath:    cfgPath,
		onSave:     onSave,
		bus:        bus,
		width:      width,
		height:     height,
		now:        time.Now(),
		overview:   NewOverviewModel(cw, ch),
		trading:    NewTradingModel(cw, ch),
		strategies: NewStrategiesModel(cw, ch, tp),
		wallets:    NewWalletsModel(tp, cfgPath, cw, ch),
		copytrader: NewCopytradingModel(cfg, cfgPath, cw, ch),
		markets:    NewMarketsModel(nil, ""),
		logs:       NewLogsModel(cw, ch),
		settings:   NewSettingsModel(cfg, cfgPath, cw, ch, onSave),
	}
}

// SetWallet is kept for API compatibility but no longer displays in the UI.
func (m *AppModel) SetWallet(_ string) {}

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
		cw := m.contentWidth()
		ch := m.contentHeight()
		// Update dimensions on all tab models (same package — unexported fields accessible)
		m.overview.width = cw
		m.overview.height = ch
		m.trading.width = cw
		m.trading.height = ch
		m.strategies.width = cw
		m.strategies.height = ch
		m.wallets.width = cw
		m.wallets.height = ch
		m.copytrader.width = cw
		m.copytrader.height = ch
		m.logs = NewLogsModel(cw, ch)
		m.markets, _ = m.markets.Update(msg)
		return m, m.bus.WaitForEvent()

	case MarketsUpdatedMsg:
		var cmd tea.Cmd
		m.markets, cmd = m.markets.Update(msg)
		return m, tea.Batch(cmd, m.bus.WaitForEvent())

	case CancelOrderMsg:
		if tp, ok := m.strategies.provider.(TradingProvider); ok {
			if err := tp.CancelOrder(msg.ID); err != nil {
				return m, func() tea.Msg { return ToastMsg{Text: "Cancel failed: " + err.Error(), Kind: "error"} }
			}
			return m, func() tea.Msg { return ToastMsg{Text: "Order cancelled", Kind: "success"} }
		}
		return m, m.bus.WaitForEvent()

	case CancelAllOrdersMsg:
		if tp, ok := m.strategies.provider.(TradingProvider); ok {
			if err := tp.CancelAllOrders(); err != nil {
				return m, func() tea.Msg { return ToastMsg{Text: "Cancel all failed: " + err.Error(), Kind: "error"} }
			}
			return m, func() tea.Msg { return ToastMsg{Text: "All orders cancelled", Kind: "success"} }
		}
		return m, m.bus.WaitForEvent()

	case PlaceOrderMsg, BatchPlaceOrderMsg:
		return m, m.bus.WaitForEvent()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Suppress q when a text field is being edited
			if m.activeTab == TabSettings && m.settings.IsEditing() {
				break
			}
			if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
				break
			}
			if m.activeTab == TabWallets && m.wallets.IsEditing() {
				break
			}
			return m, tea.Quit

		case "tab":
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

		case "1", "2", "3", "4", "5", "6", "7", "8":
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
				m.activeTab = TabStrategies
			case "4":
				m.activeTab = TabWallets
			case "5":
				m.activeTab = TabCopytrading
			case "6":
				m.activeTab = TabMarkets
			case "7":
				m.activeTab = TabLogs
			case "8":
				m.activeTab = TabSettings
			}
		}

	case StartStrategyMsg:
		if err := m.strategies.provider.StartStrategy(msg.Name); err != nil {
			return m, func() tea.Msg { return ToastMsg{Text: err.Error(), Kind: "error"} }
		}
		return m, m.bus.WaitForEvent()

	case StopStrategyMsg:
		if err := m.strategies.provider.StopStrategy(msg.Name); err != nil {
			return m, func() tea.Msg { return ToastMsg{Text: err.Error(), Kind: "error"} }
		}
		return m, m.bus.WaitForEvent()

	case CycleStrategyWalletMsg:
		wallets := m.strategies.provider.AvailableWallets()
		if len(wallets) == 0 {
			return m, func() tea.Msg { return ToastMsg{Text: "No wallets available", Kind: "error"} }
		}
		var current string
		for _, r := range m.strategies.rows {
			if r.Name == msg.Name {
				current = r.WalletID
				break
			}
		}
		next := wallets[0]
		for i, w := range wallets {
			if w == current {
				next = wallets[(i+1)%len(wallets)]
				break
			}
		}
		if err := m.strategies.provider.SetStrategyWallets(msg.Name, []string{next}); err != nil {
			return m, func() tea.Msg { return ToastMsg{Text: err.Error(), Kind: "error"} }
		}
		return m, m.bus.WaitForEvent()

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

	case HealthSnapshotMsg:
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

	case StrategiesUpdateMsg:
		m.trading.SetStrategyRows(msg.Rows)
		m.strategies.SetRows(msg.Rows)
		return m, m.bus.WaitForEvent()

	case CopytradingTradeMsg:
		m.copytrader.AddTrade(msg.Line)
		return m, m.bus.WaitForEvent()

	case LanguageChangedMsg:
		cw := m.contentWidth()
		ch := m.contentHeight()
		m.trading = NewTradingModel(cw, ch)
		m.copytrader = NewCopytradingModel(m.cfg, m.cfgPath, cw, ch)
		return m, m.bus.WaitForEvent()

	case UpdateAvailableMsg:
		m.updateBanner = fmt.Sprintf(
			" ◈ Update available: v%s — %s (published %s) ",
			msg.Version, msg.ReleaseNotes, msg.PublishedAt,
		)
		return m, m.bus.WaitForEvent()
	}

	// Route key events to active tab
	var cmd tea.Cmd
	switch m.activeTab {
	case TabOverview:
		m.overview, cmd = m.overview.Update(msg)
	case TabTrading:
		m.trading, cmd = m.trading.Update(msg)
	case TabStrategies:
		m.strategies, cmd = m.strategies.Update(msg)
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
	cw := m.contentWidth()

	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.overview.View()
	case TabTrading:
		content = m.trading.View()
	case TabStrategies:
		content = m.strategies.View()
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

	sidebar := RenderSidebar(m.activeTab, m.height, m.overview.subsystems)

	// Content area fills remaining width, height-1 rows
	contentArea := lipgloss.NewStyle().
		Width(cw).
		Height(m.contentHeight()).
		Render(content)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, contentArea)
	statusBar := m.renderStatusBar()

	var parts []string
	if m.updateBanner != "" {
		banner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")).
			Bold(true).
			Width(m.width).
			Render(m.updateBanner)
		parts = append(parts, banner)
	}
	parts = append(parts, body, statusBar)
	full := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if m.toast != nil {
		return m.overlayToast(full)
	}
	return full
}

func (m AppModel) renderStatusBar() string {
	clock := m.now.UTC().Format("15:04:05 UTC")
	live := StyleSuccess.Render("● LIVE")
	content := fmt.Sprintf("  %s   %s  ", clock, live)
	return StyleStatusBar.Width(m.width).Render(content)
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
