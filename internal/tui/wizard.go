package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// wizardStep describes one wizard input step.
type wizardStep struct {
	Label  string
	Hint   string
	IsPass bool
}

// buildWizardSteps returns wizard steps using the current locale.
func buildWizardSteps() []wizardStep {
	t := i18n.T()
	return []wizardStep{
		{Label: t.WizardStep1Label, Hint: t.WizardStep1Hint, IsPass: true},
	}
}

// WizardDoneMsg is emitted when the wizard completes and config.toml is written.
type WizardDoneMsg struct {
	ConfigPath string
}

// WizardModel is the Bubble Tea model for the first-run wizard.
type WizardModel struct {
	step    int
	steps   []wizardStep
	inputs  []textinput.Model
	errMsg  string
	width   int
	height  int
	outPath string
}

// NewWizardModel creates a new WizardModel.
func NewWizardModel(width, height int, outPath string) WizardModel {
	steps := buildWizardSteps()
	inputs := make([]textinput.Model, len(steps))
	for i, s := range steps {
		ti := textinput.New()
		ti.Placeholder = s.Label
		ti.CharLimit = 256
		if s.IsPass {
			ti.EchoMode = textinput.EchoPassword
		}
		inputs[i] = ti
	}
	inputs[0].Focus()
	return WizardModel{steps: steps, inputs: inputs, width: width, height: height, outPath: outPath}
}

func (m WizardModel) Init() tea.Cmd { return textinput.Blink }

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			val := strings.TrimSpace(m.inputs[m.step].Value())
			if val == "" {
				m.errMsg = i18n.T().WizardEmptyField
				return m, nil
			}
			m.errMsg = ""
			m.inputs[m.step].Blur()

			if m.step < len(m.steps)-1 {
				m.step++
				m.inputs[m.step].Focus()
				return m, textinput.Blink
			}

			if err := m.writeConfig(); err != nil {
				m.errMsg = fmt.Sprintf(i18n.T().WizardWriteError, err)
				return m, nil
			}
			return m, func() tea.Msg { return WizardDoneMsg{ConfigPath: m.outPath} }
		}
	}

	var cmd tea.Cmd
	m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
	return m, cmd
}

func (m WizardModel) writeConfig() error {
	type authSection struct {
		PrivateKey string `toml:"private_key"`
		ChainID    int    `toml:"chain_id"`
	}
	type apiSection struct {
		ClobURL    string `toml:"clob_url"`
		GammaURL   string `toml:"gamma_url"`
		DataURL    string `toml:"data_url"`
		WSURL      string `toml:"ws_url"`
		TimeoutSec int    `toml:"timeout_sec"`
		MaxRetries int    `toml:"max_retries"`
	}
	type logSection struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
	}
	type uiSection struct {
		Language string `toml:"language"`
	}
	type minCfg struct {
		API  apiSection  `toml:"api"`
		Auth authSection `toml:"auth"`
		Log  logSection  `toml:"log"`
		UI   uiSection   `toml:"ui"`
	}

	cfg := minCfg{
		API: apiSection{
			ClobURL:    "https://clob.polymarket.com",
			GammaURL:   "https://gamma-api.polymarket.com",
			DataURL:    "https://data-api.polymarket.com",
			WSURL:      "wss://ws-subscriptions-clob.polymarket.com/ws/",
			TimeoutSec: 10,
			MaxRetries: 3,
		},
		Auth: authSection{
			PrivateKey: m.inputs[0].Value(),
			ChainID:    137,
		},
		Log: logSection{Level: "info", Format: "pretty"},
		UI:  uiSection{Language: "en"},
	}

	f, err := os.Create(m.outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func (m WizardModel) View() string {
	s := m.steps[m.step]
	progress := fmt.Sprintf(i18n.T().WizardProgress, m.step+1, len(m.steps), s.Label)

	errLine := ""
	if m.errMsg != "" {
		errLine = "\n" + StyleError.Render(m.errMsg)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		StyleBold.Render(progress),
		"",
		m.inputs[m.step].View(),
		"",
		StyleTooltip.Render(s.Hint),
		errLine,
		"",
		StyleMuted.Render(i18n.T().WizardContinue),
	)

	w := m.width - 6
	if w < 40 {
		w = 40
	}
	box := StyleBorder.Width(w).Padding(1, 2).Render(body)
	title := StyleHeader.Render(i18n.T().WizardTitle)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", box)
}
