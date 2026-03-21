package tui

import (
	"fmt"
	"strings"
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

// animTickMsg is fired every 100ms for animation updates.
type animTickMsg struct{}

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
	nx      *Nexus
	cfg     *config.Config
	cfgPath string
	onSave  func(string)
	width   int
	height  int

	// Clock
	now time.Time

	// Animation tick counter (incremented every 100ms)
	tickCount int

	// Toast overlay
	toast *toastEntry

	// updateBanner is non-empty when a newer bot version is available.
	updateBanner string
}

// contentWidth returns the usable width for tab content.
func (m AppModel) contentWidth() int {
	return max(m.width, 20)
}

// contentHeight returns the usable height for tab content (terminal - top bar - status bar).
func (m AppModel) contentHeight() int {
	return max(m.height-topBarHeight-1, 10)
}

// NewAppModel creates the root app model.
func NewAppModel(
	cfg *config.Config,
	cfgPath string,
	bus *EventBus,
	nx *Nexus,
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
	cw := max(width, 20)
	ch := max(height-topBarHeight-1, 10)

	return AppModel{
		cfg:        cfg,
		cfgPath:    cfgPath,
		onSave:     onSave,
		bus:        bus,
		nx:         nx,
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

func animTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return animTickMsg{} })
}

func (m AppModel) Init() tea.Cmd {
	// Initialize models from Nexus snapshot
	if m.nx != nil {
		snap := m.nx.Snapshot()
		m.overview.LoadSnapshot(snap)
		if strats, ok := snap["strategies"].([]StrategyRow); ok {
			m.strategies.SetRows(strats)
			m.trading.SetStrategyRows(strats)
		}
		if orders, ok := snap["orders"].([]OrderRow); ok {
			m.trading.SetOrderRows(orders)
		}
		if positions, ok := snap["positions"].([]PositionRow); ok {
			m.trading.SetPositionRows(positions)
		}
	}
	return tea.Batch(m.bus.WaitForEvent(), clockTick(), animTick())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		m.tickCount++
		m.overview, _ = m.overview.Update(msg)
		m.trading, _ = m.trading.Update(msg)
		m.strategies, _ = m.strategies.Update(msg)
		m.wallets, _ = m.wallets.Update(msg)
		m.copytrader, _ = m.copytrader.Update(msg)
		m.markets, _ = m.markets.Update(msg)
		m.logs, _ = m.logs.Update(msg)
		m.settings, _ = m.settings.Update(msg)
		return m, animTick()

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
		m.overview.Resize(cw, ch)
		m.trading.Resize(cw, ch)
		m.strategies.Resize(cw, ch)
		m.wallets.Resize(cw, ch)
		m.copytrader.Resize(cw, ch)
		m.markets.Resize(cw, ch)
		m.logs.Resize(cw, ch)
		m.settings.Resize(cw, ch)
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

	case WalletAddedMsg:
		m.wallets, _ = m.wallets.Update(msg)
		m.overview, _ = m.overview.Update(msg)
		if len(msg.Allowances) > 0 {
			missing := 0
			for _, a := range msg.Allowances {
				if !a.Approved {
					missing++
				}
			}
			if missing > 0 {
				return m, tea.Batch(
					func() tea.Msg {
						return ToastMsg{
							Text: fmt.Sprintf("Wallet added, but %d/6 allowances missing!", missing),
							Kind: "warning",
						}
					},
					m.bus.WaitForEvent(),
				)
			}
		}
		return m, m.bus.WaitForEvent()

	case WalletRemovedMsg, WalletChangedMsg, WalletStatsMsg:
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

	topBar := RenderTopBar(m.activeTab, m.width)

	// Content area fills remaining width, height - topBarHeight - status bar
	contentArea := lipgloss.NewStyle().
		Width(cw).
		Height(m.contentHeight()).
		Render(content)

	statusBar := m.renderStatusBar()

	var parts []string
	if m.updateBanner != "" {
		banner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")).
			Background(lipgloss.Color("#2E2005")).
			Bold(true).
			Width(m.width).
			Padding(0, 1).
			Render(m.updateBanner)
		parts = append(parts, banner)
	}
	parts = append(parts, topBar, contentArea, statusBar)
	full := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if m.toast != nil {
		return m.overlayToast(full)
	}
	return full
}


func (m AppModel) renderStatusBar() string {
	clock := m.now.UTC().Format("15:04:05 UTC")
	
	mode := StyleStatusMode.Render(" NORMAL ")
	live := StyleStatusLive.Render(" ● LIVE ")
	timeStr := StyleStatusTime.Render(" " + clock + " ")

	left := mode
	right := timeStr + live

	padWidth := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if padWidth < 0 {
		padWidth = 0
	}
	middle := strings.Repeat(" ", padWidth)

	return StyleStatusBar.Width(m.width).Render(left + middle + right)
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
		MarginRight(2).
		MarginBottom(1).
		Render(toast)
}
