package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/config"
	"github.com/atlasdev/orbitron/internal/i18n"
)

// FieldKind is the type of a settings field.
type FieldKind int

const (
	KindString   FieldKind = iota
	KindPassword           // masked textinput
	KindInt                // numeric textinput
	KindBool               // toggle: Space/Enter flips true/false
	KindEnum               // cycle: Space/Enter cycles through Options
)

// FieldDef describes one editable setting.
type FieldDef struct {
	Section  func() string
	Label    func() string
	Tooltip  func() string
	Kind     FieldKind
	Options  []string // for KindEnum: valid values in cycle order
	OnChange func(v string) tea.Cmd
	Get      func(*config.Config) string
	Set      func(*config.Config, string) error
}

// allFields is the complete list of editable settings in display order.
var allFields = []FieldDef{
	// ── UI ───────────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionUI },
		Label:   func() string { return i18n.T().FieldLanguage },
		Tooltip: func() string { return i18n.T().TooltipLanguage },
		Kind:    KindEnum,
		Options: []string{"en", "ru", "zh", "ja", "ko"},
		Get:     func(c *config.Config) string { return c.UI.Language },
		Set:     func(c *config.Config, v string) error { c.UI.Language = v; return nil },
		OnChange: func(v string) tea.Cmd {
			i18n.SetLanguage(v)
			return func() tea.Msg { return LanguageChangedMsg{} }
		},
	},
	// ── Auth ─────────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionAuth },
		Label:   func() string { return i18n.T().FieldPrivKey },
		Tooltip: func() string { return i18n.T().TooltipPrivKey },
		Kind:    KindPassword,
		Get:     func(c *config.Config) string { return c.Auth.PrivateKey },
		Set:     func(c *config.Config, v string) error { c.Auth.PrivateKey = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionAuth },
		Label:   func() string { return i18n.T().FieldChainID },
		Tooltip: func() string { return i18n.T().TooltipChainID },
		Kind:    KindEnum,
		Options: []string{"137", "80002"},
		Get:     func(c *config.Config) string { return strconv.FormatInt(c.Auth.ChainID, 10) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			c.Auth.ChainID = n
			return nil
		},
	},
	// ── API ──────────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionAPI },
		Label:   func() string { return i18n.T().FieldTimeout },
		Tooltip: func() string { return i18n.T().TooltipTimeout },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.API.TimeoutSec) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.API.TimeoutSec = n
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionAPI },
		Label:   func() string { return i18n.T().FieldMaxRetries },
		Tooltip: func() string { return i18n.T().TooltipMaxRetries },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.API.MaxRetries) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.API.MaxRetries = n
			return nil
		},
	},
	// ── Monitor ──────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionMonitor },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipMonitorEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Monitor.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Monitor.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionMonitor },
		Label:   func() string { return i18n.T().FieldPollInterval },
		Tooltip: func() string { return i18n.T().TooltipMonitorPoll },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.Monitor.PollIntervalMs) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Monitor.PollIntervalMs = n
			return nil
		},
	},
	// ── Trades Monitor ───────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipTradesEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Monitor.Trades.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Monitor.Trades.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldPollInterval },
		Tooltip: func() string { return i18n.T().TooltipTradesPoll },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.PollIntervalMs) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Monitor.Trades.PollIntervalMs = n
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldAlertOnFill },
		Tooltip: func() string { return i18n.T().TooltipAlertOnFill },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnFill) },
		Set:     func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnFill = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldAlertOnCancel },
		Tooltip: func() string { return i18n.T().TooltipAlertOnCancel },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnCancel) },
		Set:     func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnCancel = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldTradesLimit },
		Tooltip: func() string { return i18n.T().TooltipTradesLimit },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.TradesLimit) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Monitor.Trades.TradesLimit = n
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionTradesMonitor },
		Label:   func() string { return i18n.T().FieldTrackPositions },
		Tooltip: func() string { return i18n.T().TooltipTradesTrack },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Monitor.Trades.TrackPositions) },
		Set:     func(c *config.Config, v string) error { c.Monitor.Trades.TrackPositions = parseBool(v); return nil },
	},
	// ── Trading ──────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionTrading },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipTradingEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Trading.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Trading.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTrading },
		Label:   func() string { return i18n.T().FieldMaxPositionUSD },
		Tooltip: func() string { return i18n.T().TooltipMaxPosition },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.MaxPositionUSD) },
		Set: func(c *config.Config, v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			c.Trading.MaxPositionUSD = f
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionTrading },
		Label:   func() string { return i18n.T().FieldSlippagePct },
		Tooltip: func() string { return i18n.T().TooltipSlippage },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.SlippagePct) },
		Set: func(c *config.Config, v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			c.Trading.SlippagePct = f
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionTrading },
		Label:   func() string { return i18n.T().FieldNegRisk },
		Tooltip: func() string { return i18n.T().TooltipNegRisk },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Trading.NegRisk) },
		Set:     func(c *config.Config, v string) error { c.Trading.NegRisk = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTrading },
		Label:   func() string { return i18n.T().FieldDefaultOrderType },
		Tooltip: func() string { return i18n.T().TooltipDefaultOrderType },
		Kind:    KindEnum,
		Options: []string{"GTC", "GTD", "FOK", "FAK"},
		Get:     func(c *config.Config) string { return c.Trading.DefaultOrderType },
		Set:     func(c *config.Config, v string) error { c.Trading.DefaultOrderType = v; return nil },
	},
	// ── Copytrading ──────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionCopytrading },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipCopyEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Copytrading.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Copytrading.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionCopytrading },
		Label:   func() string { return i18n.T().FieldPollInterval },
		Tooltip: func() string { return i18n.T().TooltipCopyPoll },
		Kind:    KindInt,
		Get:     func(c *config.Config) string { return strconv.Itoa(c.Copytrading.PollIntervalMs) },
		Set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Copytrading.PollIntervalMs = n
			return nil
		},
	},
	{
		Section: func() string { return i18n.T().SectionCopytrading },
		Label:   func() string { return i18n.T().FieldSizeMode },
		Tooltip: func() string { return i18n.T().TooltipSizeMode },
		Kind:    KindEnum,
		Options: []string{"proportional", "fixed_pct"},
		Get:     func(c *config.Config) string { return c.Copytrading.SizeMode },
		Set:     func(c *config.Config, v string) error { c.Copytrading.SizeMode = v; return nil },
	},
	// ── Telegram ─────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionTelegram },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipTelegramEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Telegram.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Telegram.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTelegram },
		Label:   func() string { return i18n.T().FieldBotToken },
		Tooltip: func() string { return i18n.T().TooltipBotToken },
		Kind:    KindPassword,
		Get:     func(c *config.Config) string { return c.Telegram.BotToken },
		Set:     func(c *config.Config, v string) error { c.Telegram.BotToken = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionTelegram },
		Label:   func() string { return i18n.T().FieldAdminChatID },
		Tooltip: func() string { return i18n.T().TooltipAdminChatID },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.Telegram.AdminChatID },
		Set:     func(c *config.Config, v string) error { c.Telegram.AdminChatID = v; return nil },
	},
	// ── Database ─────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionDatabase },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipDBEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Database.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Database.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionDatabase },
		Label:   func() string { return i18n.T().FieldDBPath },
		Tooltip: func() string { return i18n.T().TooltipDBPath },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.Database.Path },
		Set:     func(c *config.Config, v string) error { c.Database.Path = v; return nil },
	},
	// ── Log ──────────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionLog },
		Label:   func() string { return i18n.T().FieldLogLevel },
		Tooltip: func() string { return i18n.T().TooltipLogLevel },
		Kind:    KindEnum,
		Options: []string{"trace", "debug", "info", "warn", "error"},
		Get:     func(c *config.Config) string { return c.Log.Level },
		Set:     func(c *config.Config, v string) error { c.Log.Level = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionLog },
		Label:   func() string { return i18n.T().FieldLogFormat },
		Tooltip: func() string { return i18n.T().TooltipLogFormat },
		Kind:    KindEnum,
		Options: []string{"pretty", "json"},
		Get:     func(c *config.Config) string { return c.Log.Format },
		Set:     func(c *config.Config, v string) error { c.Log.Format = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionLog },
		Label:   func() string { return i18n.T().FieldLogFile },
		Tooltip: func() string { return i18n.T().TooltipLogFile },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.Log.File },
		Set:     func(c *config.Config, v string) error { c.Log.File = v; return nil },
	},
	// ── Web UI ───────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionWebUI },
		Label:   func() string { return i18n.T().FieldEnabled },
		Tooltip: func() string { return i18n.T().TooltipWebUIEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.WebUI.Enabled) },
		Set:     func(c *config.Config, v string) error { c.WebUI.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionWebUI },
		Label:   func() string { return i18n.T().FieldWebUIListen },
		Tooltip: func() string { return i18n.T().TooltipWebUIListen },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.WebUI.Listen },
		Set:     func(c *config.Config, v string) error { c.WebUI.Listen = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionWebUI },
		Label:   func() string { return i18n.T().FieldWebUIJWTSecret },
		Tooltip: func() string { return i18n.T().TooltipWebUIJWTSecret },
		Kind:    KindPassword,
		Get:     func(c *config.Config) string { return c.WebUI.JWTSecret },
		Set:     func(c *config.Config, v string) error { c.WebUI.JWTSecret = v; return nil },
	},
	// ── Proxy ─────────────────────────────────────────────────────────────
	{
		Section: func() string { return i18n.T().SectionProxy },
		Label:   func() string { return i18n.T().SettingsProxyEnabled },
		Tooltip: func() string { return i18n.T().SettingsProxyEnabled },
		Kind:    KindBool,
		Get:     func(c *config.Config) string { return boolStr(c.Proxy.Enabled) },
		Set:     func(c *config.Config, v string) error { c.Proxy.Enabled = parseBool(v); return nil },
	},
	{
		Section: func() string { return i18n.T().SectionProxy },
		Label:   func() string { return i18n.T().SettingsProxyType },
		Tooltip: func() string { return i18n.T().SettingsProxyType },
		Kind:    KindEnum,
		Options: []string{"socks5", "http"},
		Get:     func(c *config.Config) string { return c.Proxy.Type },
		Set:     func(c *config.Config, v string) error { c.Proxy.Type = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionProxy },
		Label:   func() string { return i18n.T().SettingsProxyAddr },
		Tooltip: func() string { return i18n.T().SettingsProxyAddr },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.Proxy.Addr },
		Set:     func(c *config.Config, v string) error { c.Proxy.Addr = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionProxy },
		Label:   func() string { return i18n.T().SettingsProxyUsername },
		Tooltip: func() string { return i18n.T().SettingsProxyUsername },
		Kind:    KindString,
		Get:     func(c *config.Config) string { return c.Proxy.Username },
		Set:     func(c *config.Config, v string) error { c.Proxy.Username = v; return nil },
	},
	{
		Section: func() string { return i18n.T().SectionProxy },
		Label:   func() string { return i18n.T().SettingsProxyPassword },
		Tooltip: func() string { return i18n.T().SettingsProxyPassword },
		Kind:    KindPassword,
		Get:     func(c *config.Config) string { return c.Proxy.Password },
		Set:     func(c *config.Config, v string) error { c.Proxy.Password = v; return nil },
	},
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

// sectionNames computes the ordered, deduplicated list of section names dynamically.
func sectionNames() []string {
	seen := map[string]bool{}
	var result []string
	for _, f := range allFields {
		name := f.Section()
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}

// SettingsModel is the Settings tab sub-model.
type SettingsModel struct {
	cfg      config.Config
	original config.Config
	cfgPath  string

	fields    []FieldDef
	inputs    []textinput.Model // for KindString / KindPassword / KindInt
	optionIdx []int             // for KindBool and KindEnum: current option index
	cursor    int
	editing   bool // true when a text field is being edited
	modified  []bool
	errMsg    string

	activeSection int

	width  int
	height int

	OnSave func(path string)
}

// NewSettingsModel creates a new SettingsModel.
func NewSettingsModel(cfg *config.Config, cfgPath string, width, height int, onSave func(string)) SettingsModel {
	m := SettingsModel{
		cfg:      *cfg,
		original: *cfg,
		cfgPath:  cfgPath,
		fields:   allFields,
		width:    width,
		height:   height,
		OnSave:   onSave,
	}

	m.inputs = make([]textinput.Model, len(allFields))
	m.optionIdx = make([]int, len(allFields))
	m.modified = make([]bool, len(allFields))

	for i, f := range allFields {
		cur := f.Get(cfg)
		switch f.Kind {
		case KindBool:
			if parseBool(cur) {
				m.optionIdx[i] = 1
			}
		case KindEnum:
			for j, opt := range f.Options {
				if opt == cur {
					m.optionIdx[i] = j
					break
				}
			}
		default:
			ti := textinput.New()
			ti.SetValue(cur)
			ti.CharLimit = 256
			ti.PromptStyle = StyleAccent
			ti.Cursor.Style = StyleAccent
			if f.Kind == KindPassword {
				ti.EchoMode = textinput.EchoPassword
			}
			m.inputs[i] = ti
		}
	}

	if idxs := m.sectionIndexes(0); len(idxs) > 0 {
		m.cursor = idxs[0]
	}
	return m
}

// IsEditing reports whether a text field is currently being edited.
func (m SettingsModel) IsEditing() bool { return m.editing }

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) sectionIndexes(slot int) []int {
	names := sectionNames()
	if slot < 0 || slot >= len(names) {
		return nil
	}
	name := names[slot]
	var idxs []int
	for i, f := range m.fields {
		if f.Section() == name {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// currentValue returns the effective string value for field i.
func (m SettingsModel) currentValue(i int) string {
	f := m.fields[i]
	switch f.Kind {
	case KindBool:
		if m.optionIdx[i] == 1 {
			return "true"
		}
		return "false"
	case KindEnum:
		if len(f.Options) == 0 {
			return ""
		}
		return f.Options[m.optionIdx[i]]
	default:
		return m.inputs[i].Value()
	}
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	if m.editing {
		return m.updateEditing(msg)
	}
	switch msg := msg.(type) {
	case ConfigReloadedMsg:
		m.cfg = *msg.Config
		m.original = *msg.Config
		for i, f := range m.fields {
			cur := f.Get(msg.Config)
			switch f.Kind {
			case KindBool:
				if parseBool(cur) {
					m.optionIdx[i] = 1
				} else {
					m.optionIdx[i] = 0
				}
			case KindEnum:
				for j, opt := range f.Options {
					if opt == cur {
						m.optionIdx[i] = j
						break
					}
				}
			default:
				m.inputs[i].SetValue(cur)
			}
			m.modified[i] = false
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			idxs := m.sectionIndexes(m.activeSection)
			if pos := sliceIndex(idxs, m.cursor); pos > 0 {
				m.cursor = idxs[pos-1]
			}
		case "down", "j":
			idxs := m.sectionIndexes(m.activeSection)
			if pos := sliceIndex(idxs, m.cursor); pos < len(idxs)-1 {
				m.cursor = idxs[pos+1]
			}
		case "left", "h":
			if m.activeSection > 0 {
				m.activeSection--
				if idxs := m.sectionIndexes(m.activeSection); len(idxs) > 0 {
					m.cursor = idxs[0]
				}
			}
		case "right", "l":
			if m.activeSection < len(sectionNames())-1 {
				m.activeSection++
				if idxs := m.sectionIndexes(m.activeSection); len(idxs) > 0 {
					m.cursor = idxs[0]
				}
			}

		case "enter", " ":
			f := m.fields[m.cursor]
			switch f.Kind {
			case KindBool:
				m.optionIdx[m.cursor] = 1 - m.optionIdx[m.cursor]
				orig := f.Get(&m.original)
				m.modified[m.cursor] = m.currentValue(m.cursor) != orig
			case KindEnum:
				if len(f.Options) > 0 {
					m.optionIdx[m.cursor] = (m.optionIdx[m.cursor] + 1) % len(f.Options)
					orig := f.Get(&m.original)
					m.modified[m.cursor] = m.currentValue(m.cursor) != orig
					if f.OnChange != nil {
						return m, f.OnChange(f.Options[m.optionIdx[m.cursor]])
					}
				}
			default:
				// Enter editing mode for text fields
				m.editing = true
				m.inputs[m.cursor].Focus()
				if f.Kind == KindPassword {
					m.inputs[m.cursor].EchoMode = textinput.EchoNormal
				}
				return m, textinput.Blink
			}

		case "s", "S":
			m.errMsg = ""
			cfgCopy := m.cfg
			for i, f := range m.fields {
				if err := f.Set(&cfgCopy, m.currentValue(i)); err != nil {
					m.errMsg = fmt.Sprintf(i18n.T().SettingsErrField, f.Label(), err)
					return m, nil
				}
			}
			if err := config.Save(m.cfgPath, &cfgCopy); err != nil {
				m.errMsg = fmt.Sprintf(i18n.T().SettingsErrSave, err)
				return m, nil
			}
			m.cfg = cfgCopy
			m.original = cfgCopy
			for i := range m.modified {
				m.modified[i] = false
			}
			if m.OnSave != nil {
				m.OnSave(m.cfgPath)
			}

		case "r", "R":
			m.cfg = m.original
			for i, f := range m.fields {
				cur := f.Get(&m.original)
				switch f.Kind {
				case KindBool:
					if parseBool(cur) {
						m.optionIdx[i] = 1
					} else {
						m.optionIdx[i] = 0
					}
				case KindEnum:
					for j, opt := range f.Options {
						if opt == cur {
							m.optionIdx[i] = j
							break
						}
					}
				default:
					m.inputs[i].SetValue(cur)
				}
				m.modified[i] = false
			}
			m.errMsg = ""
		}
	}
	return m, nil
}

func (m SettingsModel) updateEditing(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			m.editing = false
			m.inputs[m.cursor].Blur()
			orig := m.fields[m.cursor].Get(&m.original)
			m.modified[m.cursor] = m.inputs[m.cursor].Value() != orig
			if m.fields[m.cursor].Kind == KindPassword {
				m.inputs[m.cursor].EchoMode = textinput.EchoPassword
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
	return m, cmd
}

func sliceIndex(slice []int, val int) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

// renderWidget renders the interactive widget for field i.
func (m SettingsModel) renderWidget(i int) string {
	f := m.fields[i]
	focused := i == m.cursor

	switch f.Kind {
	case KindBool:
		if m.optionIdx[i] == 1 {
			s := StyleToggleOn.Render("● ON ")
			if focused {
				return lipgloss.NewStyle().
					Border(BorderRounded).
					BorderForeground(ColorSuccess).
					Render(s)
			}
			return lipgloss.NewStyle().
				Border(BorderRounded).
				BorderForeground(ColorBorder).
				Render(s)
		}
		s := StyleToggleOff.Render("○ OFF")
		if focused {
			return lipgloss.NewStyle().
				Border(BorderRounded).
				BorderForeground(ColorAccent).
				Render(s)
		}
		return lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorBorder).
			Render(s)

	case KindEnum:
		cur := ""
		if len(f.Options) > 0 {
			cur = f.Options[m.optionIdx[i]]
		}
		left := StyleEnumArrow.Render("‹")
		right := StyleEnumArrow.Render("›")
		val := StyleEnumValue.Render(fmt.Sprintf(" %-16s", cur))
		if focused {
			return lipgloss.NewStyle().
				Border(BorderRounded).
				BorderForeground(ColorAccent).
				Render(left + val + right)
		}
		return lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorBorder).
			Render(StyleMuted.Render("‹") + StyleFgDim.Render(fmt.Sprintf(" %-16s", cur)) + StyleMuted.Render("›"))

	default:
		return m.inputs[i].View()
	}
}

func (m SettingsModel) View() string {
	halfW := max((m.width-6)/2, 28)
	t := i18n.T()

	// ── Section sub-tab selector ────────────────────────────────────────────
	var sectionBar strings.Builder
	for i, s := range sectionNames() {
		if i == m.activeSection {
			sectionBar.WriteString(StyleSubTabActive.Render(" " + s + " "))
		} else {
			sectionBar.WriteString(StyleSubTabInactive.Render(" " + s + " "))
		}
		sectionBar.WriteString("  ")
	}

	// ── Left: fields list ───────────────────────────────────────────────────
	idxs := m.sectionIndexes(m.activeSection)
	var leftLines []string
	leftLines = append(leftLines, "")
	for _, idx := range idxs {
		f := m.fields[idx]
		mod := ""
		if m.modified[idx] {
			mod = " " + StyleWarning.Render("●")
		}
		cur := "   "
		if idx == m.cursor {
			cur = StyleAccent.Render(" ▶ ")
		}
		widget := m.renderWidget(idx)
		label := fmt.Sprintf("%-24s", f.Label())
		if idx == m.cursor {
			label = StyleBold.Render(label)
		} else {
			label = StyleFgDim.Render(label)
		}
		line := fmt.Sprintf("%s%s  %s%s", cur, label, widget, mod)
		leftLines = append(leftLines, line)
		leftLines = append(leftLines, "")
	}
	leftBox := renderPanel("", strings.Join(leftLines, "\n"), halfW, true)

	// ── Right: tooltip ──────────────────────────────────────────────────────
	var tipLines []string
	tipLines = append(tipLines, "")
	if m.cursor >= 0 && m.cursor < len(m.fields) {
		f := m.fields[m.cursor]
		tipLines = append(tipLines, "   "+StyleGlow.Render(f.Label()))
		tipLines = append(tipLines, "")
		for _, line := range strings.Split(f.Tooltip(), "\n") {
			tipLines = append(tipLines, "   "+StyleTooltip.Render(line))
		}

		if f.Kind == KindEnum && len(f.Options) > 0 {
			tipLines = append(tipLines, "")
			tipLines = append(tipLines, "   "+StyleFgDim.Render(t.SettingsOptions))
			for j, opt := range f.Options {
				mark := "     "
				if j == m.optionIdx[m.cursor] {
					mark = StyleSuccess.Render("   ▶ ")
				}
				tipLines = append(tipLines, mark+StyleEnumValue.Render(opt))
			}
		}

		tipLines = append(tipLines, "")
		curVal := m.currentValue(m.cursor)
		if f.Kind == KindPassword && !m.editing {
			curVal = strings.Repeat("•", min(len(curVal), 20))
		}
		tipLines = append(tipLines, "   "+StyleMuted.Render(t.SettingsValue)+StyleFgDim.Render(curVal))
		if m.modified[m.cursor] {
			tipLines = append(tipLines, "   "+StyleWarning.Render(t.SettingsUnsaved))
		}
	}
	rightBox := renderPanel("", strings.Join(tipLines, "\n"), halfW, false)

	// ── Error line ──────────────────────────────────────────────────────────
	errLine := ""
	if m.errMsg != "" {
		errLine = " " + StyleError.Render("✖ "+m.errMsg)
	}

	// ── Help bar (context-aware) ─────────────────────────────────────────────
	var helpParts []string
	helpParts = append(helpParts, t.HelpField)
	helpParts = append(helpParts, t.HelpSection)

	if m.cursor >= 0 && m.cursor < len(m.fields) {
		f := m.fields[m.cursor]
		switch f.Kind {
		case KindBool:
			helpParts = append(helpParts, t.HelpToggle)
		case KindEnum:
			helpParts = append(helpParts, t.HelpNextOption)
		default:
			helpParts = append(helpParts, t.HelpEdit)
		}
	}
	helpParts = append(helpParts, t.HelpSave)
	helpParts = append(helpParts, t.HelpReset)

	helpPanel := renderHelpPanel(strings.Join(helpParts, "  │  "), m.width)

	rows := []string{
		" " + sectionBar.String(),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, " ", leftBox, " ", rightBox),
	}
	if errLine != "" {
		rows = append(rows, errLine)
	}
	rows = append(rows, " ", helpPanel)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
