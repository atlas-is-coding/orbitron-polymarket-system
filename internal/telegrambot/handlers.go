package telegrambot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// settingEntry describes one editable config field accessible via /set.
type settingEntry struct {
	get    func(*config.Config) string
	set    func(*config.Config, string) error
	secret bool           // if true, admin-only
	onSet  func(v string) // side effects (e.g. i18n language change)
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

// settingsMap maps dot-notation keys to config accessors.
var settingsMap = map[string]settingEntry{
	// UI
	"ui.language": {
		get: func(c *config.Config) string { return c.UI.Language },
		set: func(c *config.Config, v string) error { c.UI.Language = v; return nil },
		onSet: func(v string) { i18n.SetLanguage(v) },
	},
	// Monitor
	"monitor.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Monitor.Enabled) },
		set: func(c *config.Config, v string) error { c.Monitor.Enabled = parseBool(v); return nil },
	},
	"monitor.poll_interval_ms": {
		get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.PollIntervalMs) },
		set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Monitor.PollIntervalMs = n
			return nil
		},
	},
	// Trades Monitor
	"monitor.trades.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.Enabled) },
		set: func(c *config.Config, v string) error { c.Monitor.Trades.Enabled = parseBool(v); return nil },
	},
	"monitor.trades.poll_interval_ms": {
		get: func(c *config.Config) string { return strconv.Itoa(c.Monitor.Trades.PollIntervalMs) },
		set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Monitor.Trades.PollIntervalMs = n
			return nil
		},
	},
	"monitor.trades.alert_on_fill": {
		get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnFill) },
		set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnFill = parseBool(v); return nil },
	},
	"monitor.trades.alert_on_cancel": {
		get: func(c *config.Config) string { return boolStr(c.Monitor.Trades.AlertOnCancel) },
		set: func(c *config.Config, v string) error { c.Monitor.Trades.AlertOnCancel = parseBool(v); return nil },
	},
	// Trading
	"trading.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Trading.Enabled) },
		set: func(c *config.Config, v string) error { c.Trading.Enabled = parseBool(v); return nil },
	},
	"trading.max_position_usd": {
		get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.MaxPositionUSD) },
		set: func(c *config.Config, v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			c.Trading.MaxPositionUSD = f
			return nil
		},
	},
	"trading.slippage_pct": {
		get: func(c *config.Config) string { return fmt.Sprintf("%.2f", c.Trading.SlippagePct) },
		set: func(c *config.Config, v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			c.Trading.SlippagePct = f
			return nil
		},
	},
	"trading.neg_risk": {
		get: func(c *config.Config) string { return boolStr(c.Trading.NegRisk) },
		set: func(c *config.Config, v string) error { c.Trading.NegRisk = parseBool(v); return nil },
	},
	// Copytrading
	"copytrading.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Copytrading.Enabled) },
		set: func(c *config.Config, v string) error { c.Copytrading.Enabled = parseBool(v); return nil },
	},
	"copytrading.poll_interval_ms": {
		get: func(c *config.Config) string { return strconv.Itoa(c.Copytrading.PollIntervalMs) },
		set: func(c *config.Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.Copytrading.PollIntervalMs = n
			return nil
		},
	},
	"copytrading.size_mode": {
		get: func(c *config.Config) string { return c.Copytrading.SizeMode },
		set: func(c *config.Config, v string) error { c.Copytrading.SizeMode = v; return nil },
	},
	// Telegram (non-secret)
	"telegram.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Telegram.Enabled) },
		set: func(c *config.Config, v string) error { c.Telegram.Enabled = parseBool(v); return nil },
	},
	// Database
	"database.enabled": {
		get: func(c *config.Config) string { return boolStr(c.Database.Enabled) },
		set: func(c *config.Config, v string) error { c.Database.Enabled = parseBool(v); return nil },
	},
	"database.path": {
		get: func(c *config.Config) string { return c.Database.Path },
		set: func(c *config.Config, v string) error { c.Database.Path = v; return nil },
	},
	// Log
	"log.level": {
		get: func(c *config.Config) string { return c.Log.Level },
		set: func(c *config.Config, v string) error { c.Log.Level = v; return nil },
	},
	"log.format": {
		get: func(c *config.Config) string { return c.Log.Format },
		set: func(c *config.Config, v string) error { c.Log.Format = v; return nil },
	},
	// Auth (admin-only)
	"auth.private_key": {
		secret: true,
		get:    func(c *config.Config) string { return c.Auth.PrivateKey },
		set:    func(c *config.Config, v string) error { c.Auth.PrivateKey = v; return nil },
	},
	"auth.api_key": {
		secret: true,
		get:    func(c *config.Config) string { return c.Auth.APIKey },
		set:    func(c *config.Config, v string) error { c.Auth.APIKey = v; return nil },
	},
	"auth.api_secret": {
		secret: true,
		get:    func(c *config.Config) string { return c.Auth.APISecret },
		set:    func(c *config.Config, v string) error { c.Auth.APISecret = v; return nil },
	},
	"auth.passphrase": {
		secret: true,
		get:    func(c *config.Config) string { return c.Auth.Passphrase },
		set:    func(c *config.Config, v string) error { c.Auth.Passphrase = v; return nil },
	},
	"auth.chain_id": {
		secret: true,
		get: func(c *config.Config) string { return strconv.FormatInt(c.Auth.ChainID, 10) },
		set: func(c *config.Config, v string) error {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			c.Auth.ChainID = n
			return nil
		},
	},
	// Telegram secrets (admin-only)
	"telegram.bot_token": {
		secret: true,
		get:    func(c *config.Config) string { return c.Telegram.BotToken },
		set:    func(c *config.Config, v string) error { c.Telegram.BotToken = v; return nil },
	},
	"telegram.admin_chat_id": {
		secret: true,
		get:    func(c *config.Config) string { return c.Telegram.AdminChatID },
		set:    func(c *config.Config, v string) error { c.Telegram.AdminChatID = v; return nil },
	},
}

// GetSetting returns the current value for a dot-notation key.
func GetSetting(cfg *config.Config, key string) (string, bool) {
	e, ok := settingsMap[key]
	if !ok {
		return "", false
	}
	return e.get(cfg), true
}

// SetSetting applies a value for a dot-notation key and returns an error if key is unknown or value is invalid.
func SetSetting(cfg *config.Config, key, value string) error {
	e, ok := settingsMap[key]
	if !ok {
		return fmt.Errorf("unknown setting key: %q", key)
	}
	return e.set(cfg, value)
}

// IsSecretKey reports whether the key is admin-only.
func IsSecretKey(key string) bool {
	e, ok := settingsMap[key]
	return ok && e.secret
}

// --- Inline keyboards ---

func mainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Orders", "cmd:orders"),
			tgbotapi.NewInlineKeyboardButtonData("💼 Positions", "cmd:positions"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Overview", "cmd:overview"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Copytrading", "cmd:copytrading"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Logs", "cmd:logs"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "cmd:settings"),
		),
	)
}

func ordersKeyboard(orders []tui.OrderRow) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, o := range orders {
		label := fmt.Sprintf("❌ Cancel #%d (%s)", i+1, o.Side)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "cancel:"+o.ID),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel ALL", "cancelall:confirm"),
		tgbotapi.NewInlineKeyboardButtonData("← Back", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func cancelAllConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Yes, cancel all", "cancelall:do"),
			tgbotapi.NewInlineKeyboardButtonData("🚫 No, go back", "cmd:orders"),
		),
	)
}

func backKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Back to Menu", "cmd:menu"),
		),
	)
}

// --- Command dispatch ---

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start", "menu":
		b.sendWithKeyboard(msg.Chat.ID, "polytrade-bot\n\nChoose a section:", mainMenuKeyboard())
	case "status", "overview":
		b.sendOverview(msg.Chat.ID)
	case "orders":
		b.sendOrders(msg.Chat.ID)
	case "cancel":
		id := strings.TrimSpace(msg.CommandArguments())
		if id == "" {
			b.sendText(msg.Chat.ID, RenderError("Usage: /cancel &lt;order_id&gt;"))
			return
		}
		b.doCancelOrder(ctx, msg.Chat.ID, id)
	case "cancelall":
		b.sendWithKeyboard(msg.Chat.ID, "⚠️ Cancel ALL open orders?", cancelAllConfirmKeyboard())
	case "positions":
		b.sendPositions(msg.Chat.ID)
	case "copy":
		b.sendCopytrading(msg.Chat.ID)
	case "logs":
		b.sendLogs(msg.Chat.ID)
	case "settings":
		b.sendSettings(msg.Chat.ID, b.isAdmin(msg.Chat.ID))
	case "set":
		args := strings.Fields(msg.CommandArguments())
		if len(args) < 2 {
			b.sendText(msg.Chat.ID, RenderError("Usage: /set &lt;key&gt; &lt;value&gt;"))
			return
		}
		b.doSetSetting(ctx, msg.Chat.ID, args[0], strings.Join(args[1:], " "))
	default:
		b.sendText(msg.Chat.ID, "Unknown command. Use /start for the menu.")
	}
}

// --- Callback dispatch ---

func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	answer := tgbotapi.NewCallback(cb.ID, "")
	b.api.Send(answer) //nolint:errcheck

	chatID := cb.Message.Chat.ID
	data := cb.Data

	switch {
	case data == "cmd:menu":
		b.sendWithKeyboard(chatID, "polytrade-bot\n\nChoose a section:", mainMenuKeyboard())
	case data == "cmd:overview":
		b.sendOverview(chatID)
	case data == "cmd:orders":
		b.sendOrders(chatID)
	case data == "cmd:positions":
		b.sendPositions(chatID)
	case data == "cmd:copytrading":
		b.sendCopytrading(chatID)
	case data == "cmd:logs":
		b.sendLogs(chatID)
	case data == "cmd:settings":
		b.sendSettings(chatID, b.isAdmin(chatID))
	case data == "cancelall:confirm":
		b.sendWithKeyboard(chatID, "⚠️ Are you sure you want to cancel ALL orders?", cancelAllConfirmKeyboard())
	case data == "cancelall:do":
		b.doCancelAll(ctx, chatID)
	case strings.HasPrefix(data, "cancel:"):
		orderID := strings.TrimPrefix(data, "cancel:")
		b.doCancelOrder(ctx, chatID, orderID)
	}
}

// --- View helpers ---

func (b *Bot) sendOverview(chatID int64) {
	subsystems := b.state.Subsystems()
	orders := b.state.Orders()
	positions := b.state.Positions()
	text := RenderOverview(b.state.Balance(), subsystems, len(orders), len(positions))
	b.sendWithKeyboard(chatID, text, backKeyboard())
}

func (b *Bot) sendOrders(chatID int64) {
	orders := b.state.Orders()
	text := RenderOrders(orders)
	b.sendWithKeyboard(chatID, text, ordersKeyboard(orders))
}

func (b *Bot) sendPositions(chatID int64) {
	b.sendWithKeyboard(chatID, RenderPositions(b.state.Positions()), backKeyboard())
}

func (b *Bot) sendCopytrading(chatID int64) {
	b.sendWithKeyboard(chatID, RenderCopytrading(b.state.Traders()), backKeyboard())
}

func (b *Bot) sendLogs(chatID int64) {
	b.sendWithKeyboard(chatID, RenderLogs(b.state.Logs()), backKeyboard())
}

// settingsSections defines the display order of settings sections with their keys.
var settingsSections = []struct {
	name string
	keys []string
}{
	{"UI", []string{"ui.language"}},
	{"Auth", []string{"auth.private_key", "auth.api_key", "auth.api_secret", "auth.passphrase", "auth.chain_id"}},
	{"Monitor", []string{"monitor.enabled", "monitor.poll_interval_ms"}},
	{"Trades Monitor", []string{"monitor.trades.enabled", "monitor.trades.poll_interval_ms", "monitor.trades.alert_on_fill", "monitor.trades.alert_on_cancel"}},
	{"Trading", []string{"trading.enabled", "trading.max_position_usd", "trading.slippage_pct", "trading.neg_risk"}},
	{"Copytrading", []string{"copytrading.enabled", "copytrading.poll_interval_ms", "copytrading.size_mode"}},
	{"Telegram", []string{"telegram.enabled", "telegram.bot_token", "telegram.admin_chat_id"}},
	{"Database", []string{"database.enabled", "database.path"}},
	{"Log", []string{"log.level", "log.format"}},
}

func (b *Bot) sendSettings(chatID int64, isAdmin bool) {
	b.cfgMu.RLock()
	cfg := *b.cfg
	b.cfgMu.RUnlock()

	var parts []string
	for _, sec := range settingsSections {
		fields := make([]SettingField, 0, len(sec.keys))
		for _, k := range sec.keys {
			// short key = last segment after the last dot
			short := k
			if idx := strings.LastIndex(k, "."); idx >= 0 {
				short = k[idx+1:]
			}
			v, ok := GetSetting(&cfg, k)
			if !ok {
				continue
			}
			fields = append(fields, SettingField{Key: short, Value: v})
		}
		parts = append(parts, RenderSettingsSection(sec.name, fields, isAdmin))
	}

	footer := "\n<i>Use /set &lt;key&gt; &lt;value&gt; to change a setting.</i>"
	if isAdmin {
		footer = "\n<i>Admin mode — all fields editable.\nUse /set &lt;key&gt; &lt;value&gt;</i>"
	}
	text := strings.Join(parts, "\n") + footer
	b.sendWithKeyboard(chatID, text, backKeyboard())
}

// --- Action helpers ---

func (b *Bot) doCancelOrder(_ context.Context, chatID int64, orderID string) {
	if b.canceler == nil {
		b.sendText(chatID, RenderError("Order cancellation unavailable (TradesMonitor not enabled)"))
		return
	}
	if err := b.canceler.CancelOrder(orderID); err != nil {
		b.sendText(chatID, RenderError(err.Error()))
		return
	}
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Order <code>%s</code> cancelled.", orderID)))
}

func (b *Bot) doCancelAll(_ context.Context, chatID int64) {
	if b.canceler == nil {
		b.sendText(chatID, RenderError("Order cancellation unavailable (TradesMonitor not enabled)"))
		return
	}
	if err := b.canceler.CancelAllOrders(); err != nil {
		b.sendText(chatID, RenderError(err.Error()))
		return
	}
	b.sendText(chatID, RenderSuccess("All orders cancelled."))
}

func (b *Bot) doSetSetting(_ context.Context, chatID int64, key, value string) {
	if IsSecretKey(key) && !b.isAdmin(chatID) {
		b.sendText(chatID, RenderError(fmt.Sprintf("Key %q requires admin access.", key)))
		return
	}

	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	if err := SetSetting(&cfgCopy, key, value); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Invalid value for %q: %v", key, err)))
		return
	}

	// Apply side effects (e.g. i18n language change)
	if e, ok := settingsMap[key]; ok && e.onSet != nil {
		e.onSet(value)
	}

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	// Notify TUI of config change via EventBus
	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})

	b.sendText(chatID, RenderSuccess(fmt.Sprintf("<code>%s</code> = <code>%s</code>\nConfig saved. TUI updated.", key, value)))
}
