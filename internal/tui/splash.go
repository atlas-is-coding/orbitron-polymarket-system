package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// splashTimerMsg fires when the 3-second minimum display time elapses.
type splashTimerMsg struct{}

// splashTimeoutMsg fires when the 15-second hard timeout elapses.
type splashTimeoutMsg struct{}

// SplashModel is the startup welcome screen.
// Waits for both timerDone and marketsReady before transitioning,
// with a 15-second hard timeout as a fallback.
type SplashModel struct {
	spinner      spinner.Model
	width        int
	height       int
	version      string
	done         bool // guard: SplashDoneMsg already emitted
	timerDone    bool // 3s minimum display time elapsed
	marketsReady bool // MarketsReadyMsg received
	loadedCount  int
	totalCount   int // 500 during initial load, 0 before first MarketsLoadingMsg
}

// NewSplashModel creates the splash screen.
func NewSplashModel(width, height int) SplashModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(ColorGlow)
	return SplashModel{
		spinner: s,
		width:   width,
		height:  height,
		version: "v0.1.0",
	}
}

func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(3*time.Second, func(time.Time) tea.Msg { return splashTimerMsg{} }),
		tea.Tick(15*time.Second, func(time.Time) tea.Msg { return splashTimeoutMsg{} }),
	)
}

// tryDone emits SplashDoneMsg if both conditions are met and not already done.
func (m *SplashModel) tryDone() tea.Cmd {
	if !m.done && m.timerDone && m.marketsReady {
		m.done = true
		return func() tea.Msg { return SplashDoneMsg{} }
	}
	return nil
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case splashTimerMsg:
		m.timerDone = true
		cmd := (&m).tryDone()
		return m, cmd
	case splashTimeoutMsg:
		if !m.done {
			m.timerDone = true
			m.marketsReady = true
			m.done = true
			return m, func() tea.Msg { return SplashDoneMsg{} }
		}
		return m, nil
	case MarketsLoadingMsg:
		m.loadedCount = msg.Loaded
		m.totalCount = msg.Total
		return m, nil
	case MarketsReadyMsg:
		m.marketsReady = true
		cmd := (&m).tryDone()
		return m, cmd
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m SplashModel) View() string {
	// Block-letter logo using box-drawing chars — rendered in ColorAccent
	logo := StyleSplashLogo.Render(strings.Join([]string{
		` ██████╗ ██████╗ ██████╗ ██╗████████╗██████╗  ██████╗ ███╗  `,
		`██╔═══██╗██╔══██╗██╔══██╗██║╚══██╔══╝██╔══██╗██╔═══██╗████╗ `,
		`██║   ██║██████╔╝██████╔╝██║   ██║   ██████╔╝██║   ██║╚═══╝ `,
		`██║   ██║██╔══██╗██╔══██╗██║   ██║   ██╔══██╗██║   ██║      `,
		`╚██████╔╝██║  ██║██████╔╝██║   ██║   ██║  ██║╚██████╔╝      `,
		` ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝       `,
	}, "\n"))

	subtitle := StyleSplashSubtitle.Render("            NEXUS TERMINAL  " + m.version)
	divider := StyleMuted.Render("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var statusLine string
	if m.marketsReady {
		if m.loadedCount > 0 {
			statusLine = fmt.Sprintf("  %s Markets ready (%d)", m.spinner.View(), m.loadedCount)
		} else {
			statusLine = fmt.Sprintf("  %s Markets ready", m.spinner.View())
		}
	} else if m.totalCount > 0 {
		statusLine = fmt.Sprintf("  %s Loading markets... (%d / %d)",
			m.spinner.View(), m.loadedCount, m.totalCount)
	} else {
		statusLine = fmt.Sprintf("  %s INITIALIZING SUBSYSTEMS...", m.spinner.View())
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		"",
		logo,
		"",
		subtitle,
		divider,
		"",
		StyleFgDim.Render(statusLine),
		"",
	)

	boxW := 64
	if m.width > 0 && m.width-8 > boxW {
		boxW = min(72, m.width-8)
	}
	box := StyleSplashBox.Width(boxW).Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
