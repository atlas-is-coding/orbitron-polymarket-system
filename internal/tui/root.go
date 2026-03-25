package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/nexus"
)

// RootModel is the top-level BubbleTea model.
// It shows SplashModel on startup, then hands off to AppModel.
type RootModel struct {
	splash     SplashModel
	app        AppModel
	nx         *Nexus
	nex        *nexus.Nexus // may be nil
	showSplash bool
}

// NewRootModel creates the root model. All AppModel params forwarded directly.
func NewRootModel(
	cfg *config.Config,
	cfgPath string,
	bus *EventBus,
	nx *Nexus,
	nex *nexus.Nexus, // may be nil
	width, height int,
	onSave func(string),
	tp TradingProvider,
) RootModel {
	return RootModel{
		splash:     NewSplashModel(width, height),
		app:        NewAppModel(cfg, cfgPath, bus, nx, nex, width, height, onSave, tp),
		nx:         nx,
		nex:        nex,
		showSplash: true,
	}
}

// SetWallet forwards the wallet address to AppModel.
func (m *RootModel) SetWallet(addr string) {
	m.app.SetWallet(addr)
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(m.splash.Init(), m.app.Init())
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showSplash {
		switch msg.(type) {
		case SplashDoneMsg:
			m.showSplash = false
			return m, m.app.bus.WaitForEvent()
		}
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		// Also forward to app so EventBus drains correctly
		updated, appCmd := m.app.Update(msg)
		if a, ok := updated.(AppModel); ok {
			m.app = a
		}
		return m, tea.Batch(cmd, appCmd)
	}

	updated, cmd := m.app.Update(msg)
	if a, ok := updated.(AppModel); ok {
		m.app = a
	}
	return m, cmd
}

func (m RootModel) View() string {
	if m.showSplash {
		return m.splash.View()
	}
	return m.app.View()
}
