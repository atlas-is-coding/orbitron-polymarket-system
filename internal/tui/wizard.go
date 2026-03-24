package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/i18n"
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
		{Label: t.WizardStep2Label, Hint: t.WizardStep2Hint, IsPass: false},
		{Label: t.WizardStep3Label, Hint: t.WizardStep3Hint, IsPass: false},
		{Label: t.WizardStep4Label, Hint: t.WizardStep4Hint, IsPass: true},
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
		ti.PromptStyle = StyleAccent
		ti.Cursor.Style = StyleAccent
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
			// Steps 0 (private_key) and 3 (jwt_secret) are required; steps 1 and 2 are optional
			required := m.step == 0 || m.step == len(m.steps)-1
			if val == "" && required {
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
	lang := strings.TrimSpace(m.inputs[1].Value())
	if lang == "" {
		lang = "en"
	}
	portStr := strings.TrimSpace(m.inputs[2].Value())
	if portStr == "" {
		portStr = "8080"
	}
	listenAddr := "127.0.0.1:" + portStr
	jwtSecret := strings.TrimSpace(m.inputs[3].Value())

	type walletSection struct {
		ID         string `toml:"id"`
		Label      string `toml:"label"`
		PrivateKey string `toml:"private_key"`
		APIKey     string `toml:"api_key"`
		APISecret  string `toml:"api_secret"`
		Passphrase string `toml:"passphrase"`
		ChainID    int    `toml:"chain_id"`
		Enabled    bool   `toml:"enabled"`
		Primary    bool   `toml:"primary"`
		NegRisk    bool   `toml:"neg_risk"`
	}
	type authSection struct {
		PrivateKey string `toml:"private_key"`
		APIKey     string `toml:"api_key"`
		APISecret  string `toml:"api_secret"`
		Passphrase string `toml:"passphrase"`
		ChainID    int    `toml:"chain_id"`
	}
	type apiSection struct {
		ClobURL    string `toml:"clob_url"`
		GammaURL   string `toml:"gamma_url"`
		DataURL    string `toml:"data_url"`
		WSURL      string `toml:"ws_url"`
		PolygonRPC string `toml:"polygon_rpc"`
		TimeoutSec int    `toml:"timeout_sec"`
		MaxRetries int    `toml:"max_retries"`
	}
	type analyticsSection struct {
		Enabled        bool   `toml:"enabled"`
		Endpoint       string `toml:"endpoint"`
		ReportInterval int    `toml:"report_interval"`
		BatchSize      int    `toml:"batch_size"`
	}
	type strategyArbitrage struct {
		Enabled        bool    `toml:"enabled"`
		MinProfitUSD   float64 `toml:"min_profit_usd"`
		MaxPositionUSD float64 `toml:"max_position_usd"`
		PollIntervalMs int     `toml:"poll_interval_ms"`
		ExecuteOrders  bool    `toml:"execute_orders"`
	}
	type strategyMarketMaking struct {
		Enabled              bool    `toml:"enabled"`
		SpreadPct            float64 `toml:"spread_pct"`
		MaxPositionUSD       float64 `toml:"max_position_usd"`
		RebalanceIntervalSec int     `toml:"rebalance_interval_sec"`
		MinLiquidityUSD      float64 `toml:"min_liquidity_usd"`
		ExecuteOrders        bool    `toml:"execute_orders"`
	}
	type strategyPositiveEV struct {
		Enabled         bool    `toml:"enabled"`
		MinEdgePct      float64 `toml:"min_edge_pct"`
		MinLiquidityUSD float64 `toml:"min_liquidity_usd"`
		MaxPositionUSD  float64 `toml:"max_position_usd"`
		MaxDurationDays int     `toml:"max_duration_days"`
		PollIntervalMs  int     `toml:"poll_interval_ms"`
		ExecuteOrders   bool    `toml:"execute_orders"`
	}
	type strategyRisklessRate struct {
		Enabled         bool    `toml:"enabled"`
		MinDurationDays int     `toml:"min_duration_days"`
		MaxNoPrice      float64 `toml:"max_no_price"`
		MaxPositionUSD  float64 `toml:"max_position_usd"`
		PollIntervalMs  int     `toml:"poll_interval_ms"`
		ExecuteOrders   bool    `toml:"execute_orders"`
	}
	type strategyFadeChaos struct {
		Enabled           bool     `toml:"enabled"`
		WalletIDs         []string `toml:"wallet_ids"`
		SpikeThresholdPct float64  `toml:"spike_threshold_pct"`
		CooldownSec       int      `toml:"cooldown_sec"`
		MaxPositionUSD    float64  `toml:"max_position_usd"`
		PollIntervalMs    int      `toml:"poll_interval_ms"`
		ExecuteOrders     bool     `toml:"execute_orders"`
	}
	type strategyCrossMarket struct {
		Enabled          bool    `toml:"enabled"`
		MinDivergencePct float64 `toml:"min_divergence_pct"`
		MaxPositionUSD   float64 `toml:"max_position_usd"`
		CooldownSec      int     `toml:"cooldown_sec"`
		PollIntervalMs   int     `toml:"poll_interval_ms"`
		ExecuteOrders    bool    `toml:"execute_orders"`
	}
	type strategiesSection struct {
		Arbitrage    strategyArbitrage    `toml:"arbitrage"`
		MarketMaking strategyMarketMaking `toml:"market_making"`
		PositiveEV   strategyPositiveEV   `toml:"positive_ev"`
		RisklessRate strategyRisklessRate `toml:"riskless_rate"`
		FadeChaos    strategyFadeChaos    `toml:"fade_chaos"`
		CrossMarket  strategyCrossMarket  `toml:"cross_market"`
	}
	type riskSection struct {
		StopLossPct     float64 `toml:"stop_loss_pct"`
		TakeProfitPct   float64 `toml:"take_profit_pct"`
		MaxDailyLossUSD float64 `toml:"max_daily_loss_usd"`
	}
	type tradingSection struct {
		Enabled          bool              `toml:"enabled"`
		MaxPositionUSD   float64           `toml:"max_position_usd"`
		SlippagePct      float64           `toml:"slippage_pct"`
		DefaultOrderType string            `toml:"default_order_type"`
		NegRisk          bool              `toml:"neg_risk"`
		Strategies       strategiesSection `toml:"strategies"`
		Risk             riskSection       `toml:"risk"`
	}
	type tradesSection struct {
		Enabled         bool `toml:"enabled"`
		PollIntervalMs  int  `toml:"poll_interval_ms"`
		TrackPositions  bool `toml:"track_positions"`
		AlertOnFill     bool `toml:"alert_on_fill"`
		AlertOnCancel   bool `toml:"alert_on_cancel"`
		TradesLimit     int  `toml:"trades_limit"`
		PositionsPollMs int  `toml:"positions_poll_ms"`
		MaxBackoffMs    int  `toml:"max_backoff_ms"`
	}
	type monitorSection struct {
		Enabled        bool          `toml:"enabled"`
		PollIntervalMs int           `toml:"poll_interval_ms"`
		Trades         tradesSection `toml:"trades"`
	}
	type telegramSection struct {
		Enabled     bool   `toml:"enabled"`
		BotToken    string `toml:"bot_token"`
		AdminChatID string `toml:"admin_chat_id"`
	}
	type databaseSection struct {
		Enabled bool   `toml:"enabled"`
		Path    string `toml:"path"`
	}
	type logSection struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		File   string `toml:"file"`
	}
	type copytradingSection struct {
		Enabled        bool   `toml:"enabled"`
		PollIntervalMs int    `toml:"poll_interval_ms"`
		SizeMode       string `toml:"size_mode"`
	}
	type uiSection struct {
		Language string `toml:"language"`
	}
	type webuiSection struct {
		Enabled   bool   `toml:"enabled"`
		Listen    string `toml:"listen"`
		JWTSecret string `toml:"jwt_secret"`
	}
	type proxySection struct {
		Enabled  bool   `toml:"enabled"`
		Type     string `toml:"type"`
		Addr     string `toml:"addr"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	}
	type fullCfg struct {
		Wallets     []walletSection    `toml:"wallets"`
		Auth        authSection        `toml:"auth"`
		API         apiSection         `toml:"api"`
		Analytics   analyticsSection   `toml:"analytics"`
		Trading     tradingSection     `toml:"trading"`
		Monitor     monitorSection     `toml:"monitor"`
		Telegram    telegramSection    `toml:"telegram"`
		Database    databaseSection    `toml:"database"`
		Log         logSection         `toml:"log"`
		Copytrading copytradingSection `toml:"copytrading"`
		UI          uiSection          `toml:"ui"`
		WebUI       webuiSection       `toml:"webui"`
		Proxy       proxySection       `toml:"proxy"`
	}

	privKey := m.inputs[0].Value()
	cfg := fullCfg{
		Wallets: []walletSection{
			{
				ID:         "default",
				Label:      "Default",
				PrivateKey: privKey,
				APIKey:     "",
				APISecret:  "",
				Passphrase: "",
				ChainID:    137,
				Enabled:    true,
				Primary:    true,
				NegRisk:    false,
			},
		},
		Auth: authSection{
			PrivateKey: privKey,
			APIKey:     "",
			APISecret:  "",
			Passphrase: "",
			ChainID:    137,
		},
		API: apiSection{
			ClobURL:    "https://clob.polymarket.com",
			GammaURL:   "https://gamma-api.polymarket.com",
			DataURL:    "https://data-api.polymarket.com",
			WSURL:      "wss://ws-subscriptions-clob.polymarket.com/ws/",
			PolygonRPC: "https://polygon.drpc.org",
			TimeoutSec: 10,
			MaxRetries: 3,
		},
		Analytics: analyticsSection{
			Enabled:        true,
			Endpoint:       "http://getorbitron.net/api/v1/analytics/report",
			ReportInterval: 30,
			BatchSize:      10,
		},
		Trading: tradingSection{
			Enabled:          false,
			MaxPositionUSD:   50.0,
			SlippagePct:      0.0,
			DefaultOrderType: "GTC",
			NegRisk:          false,
			Strategies: strategiesSection{
				Arbitrage:    strategyArbitrage{Enabled: false, MinProfitUSD: 0.5, MaxPositionUSD: 100.0, PollIntervalMs: 5000, ExecuteOrders: false},
				MarketMaking: strategyMarketMaking{Enabled: false, SpreadPct: 2.0, MaxPositionUSD: 200.0, RebalanceIntervalSec: 30, MinLiquidityUSD: 10000.0, ExecuteOrders: false},
				PositiveEV:   strategyPositiveEV{Enabled: false, MinEdgePct: 5.0, MinLiquidityUSD: 5000.0, MaxPositionUSD: 50.0, MaxDurationDays: 14, PollIntervalMs: 30000, ExecuteOrders: false},
				RisklessRate: strategyRisklessRate{Enabled: false, MinDurationDays: 30, MaxNoPrice: 0.05, MaxPositionUSD: 50.0, PollIntervalMs: 60000, ExecuteOrders: false},
				FadeChaos:    strategyFadeChaos{Enabled: false, WalletIDs: []string{}, SpikeThresholdPct: 10.0, CooldownSec: 300, MaxPositionUSD: 50.0, PollIntervalMs: 10000, ExecuteOrders: false},
				CrossMarket:  strategyCrossMarket{Enabled: false, MinDivergencePct: 5.0, MaxPositionUSD: 75.0, CooldownSec: 300, PollIntervalMs: 30000, ExecuteOrders: false},
			},
			Risk: riskSection{StopLossPct: 20.0, TakeProfitPct: 50.0, MaxDailyLossUSD: 100.0},
		},
		Monitor: monitorSection{
			Enabled:        true,
			PollIntervalMs: 1000,
			Trades: tradesSection{
				Enabled:         false,
				PollIntervalMs:  5000,
				TrackPositions:  false,
				AlertOnFill:     false,
				AlertOnCancel:   false,
				TradesLimit:     50,
				PositionsPollMs: 5000,
				MaxBackoffMs:    30000,
			},
		},
		Telegram:    telegramSection{Enabled: false, BotToken: "", AdminChatID: ""},
		Database:    databaseSection{Enabled: true, Path: ""},
		Log:         logSection{Level: "info", Format: "pretty", File: ""},
		Copytrading: copytradingSection{Enabled: false, PollIntervalMs: 10000, SizeMode: "proportional"},
		UI:          uiSection{Language: lang},
		WebUI:       webuiSection{Enabled: true, Listen: listenAddr, JWTSecret: jwtSecret},
		Proxy:       proxySection{Enabled: false, Type: "socks5", Addr: "", Username: "", Password: ""},
	}

	f, err := os.Create(m.outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write a header comment. BurntSushi/toml encoder does not support inline comments.
	if _, err := fmt.Fprintf(f, "# Password for logging in to the web interface is set in [webui] jwt_secret\n\n"); err != nil {
		return err
	}
	return toml.NewEncoder(f).Encode(cfg)
}

func (m WizardModel) View() string {
	s := m.steps[m.step]
	t := i18n.T()

	// Progress dots
	var progressDots strings.Builder
	for i := range m.steps {
		if i < m.step {
			progressDots.WriteString(StyleSuccess.Render("● "))
		} else if i == m.step {
			progressDots.WriteString(StyleGlow.Render("● "))
		} else {
			progressDots.WriteString(StyleMuted.Render("○ "))
		}
	}

	stepLabel := StyleAccent.Render(fmt.Sprintf(t.WizardProgress, m.step+1, len(m.steps), s.Label))

	errLine := ""
	if m.errMsg != "" {
		errLine = "\n" + StyleError.Render("  ✗ "+m.errMsg)
	}

	sep := StyleMuted.Render("  " + strings.Repeat("─", 36))

	body := lipgloss.JoinVertical(lipgloss.Left,
		"",
		" "+StyleGlow.Bold(true).Render("◈ ORBITRON")+"  "+StyleMuted.Render("—  First Run Setup"),
		" "+StyleFgDim.Render("Polymarket Trading Terminal"),
		"",
		sep,
		"",
		"  "+progressDots.String(),
		"  "+stepLabel,
		"",
		"  "+m.inputs[m.step].View(),
		"",
		StyleTooltip.Render("  "+s.Hint),
		errLine,
		"",
		StyleMuted.Render("  "+t.WizardContinue),
		"",
	)

	w := m.width - 8
	if w < 52 {
		w = 52
	}
	box := StyleSplashBox.Width(w).Render(body)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
