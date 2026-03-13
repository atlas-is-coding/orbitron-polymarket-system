package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const splashDuration = 2 * time.Second

// splashTickMsg fires when splash timer expires.
type splashTickMsg struct{}

// SplashModel is the startup welcome screen shown for splashDuration.
type SplashModel struct {
	spinner spinner.Model
	width   int
	height  int
	version string
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
		tea.Tick(splashDuration, func(time.Time) tea.Msg { return splashTickMsg{} }),
	)
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case splashTickMsg:
		return m, func() tea.Msg { return SplashDoneMsg{} }
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
	loading := fmt.Sprintf("  %s INITIALIZING SUBSYSTEMS...", m.spinner.View())

	body := lipgloss.JoinVertical(lipgloss.Left,
		"",
		logo,
		"",
		subtitle,
		divider,
		"",
		StyleFgDim.Render(loading),
		"",
	)

	boxW := 64
	if m.width > 0 && m.width-8 > boxW {
		boxW = min(72, m.width-8)
	}
	box := StyleSplashBox.Width(boxW).Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
