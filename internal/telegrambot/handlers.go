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
			tgbotapi.NewInlineKeyboardButtonData("📊 Overview", "cmd:overview"),
			tgbotapi.NewInlineKeyboardButtonData("📈 Trading", "cmd:trading"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Copytrading", "cmd:copytrading"),
			tgbotapi.NewInlineKeyboardButtonData("👛 Wallets", "cmd:wallets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Logs", "cmd:logs"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "cmd:settings"),
		),
	)
}

func walletsKeyboard(wallets []WalletEntry) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, w := range wallets {
		label := w.Label
		if label == "" {
			label = w.ID
		}
		toggleIcon := "▶ Enable"
		if w.Enabled {
			toggleIcon = "⏸ Disable"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("👛 %s  %s", label, toggleIcon),
				"wallet:toggle:"+w.ID,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
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
			tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
		),
	)
}

func copytradingKeyboard(traders []config.TraderConfig) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range traders {
		addr := t.Address
		short := addr
		if len(short) > 12 {
			short = short[:6] + "…" + short[len(short)-4:]
		}
		label := t.Label
		if label == "" {
			label = short
		}
		toggleIcon := "▶ Enable"
		if t.Enabled {
			toggleIcon = "⏸ Disable"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s  %s", label, toggleIcon),
				"trader:toggle:"+addr,
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑 Remove",
				"trader:remove:"+addr,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Add Trader", "addtrader:start"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// settingsSectionsKeyboard returns buttons for each settings section.
func settingsSectionsKeyboard() tgbotapi.InlineKeyboardMarkup {
	sections := []string{"UI", "Monitor", "Trades Monitor", "Trading", "Copytrading", "Telegram", "Database", "Log", "Auth"}
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(sections); i += 2 {
		if i+1 < len(sections) {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚙️ "+sections[i], "settings:section:"+sections[i]),
				tgbotapi.NewInlineKeyboardButtonData("⚙️ "+sections[i+1], "settings:section:"+sections[i+1]),
			))
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚙️ "+sections[i], "settings:section:"+sections[i]),
			))
		}
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// langPickerKeyboard returns an inline keyboard for language selection.
// currentLang is the active language code (e.g. "en", "ru").
func langPickerKeyboard(currentLang string) tgbotapi.InlineKeyboardMarkup {
	type lang struct {
		code  string
		label string
	}
	langs := []lang{
		{"en", "🇬🇧 English"},
		{"ru", "🇷🇺 Русский"},
		{"zh", "🇨🇳 中文"},
		{"ja", "🇯🇵 日本語"},
		{"ko", "🇰🇷 한국어"},
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(langs); i += 2 {
		if i+1 < len(langs) {
			l1, l2 := langs[i], langs[i+1]
			label1, label2 := l1.label, l2.label
			if l1.code == currentLang {
				label1 += " ✓"
			}
			if l2.code == currentLang {
				label2 += " ✓"
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label1, "setlang:"+l1.code),
				tgbotapi.NewInlineKeyboardButtonData(label2, "setlang:"+l2.code),
			))
		} else {
			l := langs[i]
			label := l.label
			if l.code == currentLang {
				label += " ✓"
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label, "setlang:"+l.code),
			))
		}
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Settings", "cmd:settings"),
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// sectionFieldsKeyboard builds per-field buttons for a settings section.
// Bool fields get a toggle button. String/number fields get an edit button.
func sectionFieldsKeyboard(sectionName string, keys []string, cfg *config.Config, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, k := range keys {
		if IsSecretKey(k) && !isAdmin {
			continue
		}
		val, ok := GetSetting(cfg, k)
		if !ok {
			continue
		}
		short := k
		if idx := strings.LastIndex(k, "."); idx >= 0 {
			short = k[idx+1:]
		}

		if val == "true" || val == "false" {
			icon := "🔴"
			if val == "true" {
				icon = "🟢"
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s %s: %s", icon, short, val),
					"toggle:"+k,
				),
			))
		} else {
			display := val
			if display == "" {
				display = "—"
			}
			if len(display) > 20 {
				display = display[:9] + "…" + display[len(display)-8:]
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✏️ %s: %s", short, display),
					"edit:"+k,
				),
			))
		}
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Settings", "cmd:settings"),
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func tradingKeyboard(subTab string, orders []tui.OrderRow) tgbotapi.InlineKeyboardMarkup {
	ordersLabel := "📋 Orders"
	posLabel := "💼 Positions"
	if subTab == "orders" {
		ordersLabel = "📋 Orders ✓"
	} else {
		posLabel = "💼 Positions ✓"
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ordersLabel, "trading:orders"),
		tgbotapi.NewInlineKeyboardButtonData(posLabel, "trading:positions"),
	))

	if subTab == "orders" {
		for i, o := range orders {
			label := fmt.Sprintf("❌ Cancel #%d (%s)", i+1, o.Side)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label, "cancel:"+o.ID),
			))
		}
		if len(orders) > 0 {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Cancel ALL", "cancelall:confirm"),
			))
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// --- Command dispatch ---

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start", "menu":
		b.state.SetMenuMsgID(0) // force new message on /start
		b.sendMenu(msg.Chat.ID)
	case "status", "overview":
		b.sendOverview(msg.Chat.ID)
	case "trading":
		b.sendTrading(msg.Chat.ID, "orders")
	case "orders":
		b.sendTrading(msg.Chat.ID, "orders")
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
		b.sendTrading(msg.Chat.ID, "positions")
	case "wallets":
		b.sendWallets(msg.Chat.ID)
	case "togglewallet":
		id := strings.TrimSpace(msg.CommandArguments())
		if id == "" {
			b.sendText(msg.Chat.ID, RenderError("Usage: /togglewallet &lt;wallet_id&gt;"))
			return
		}
		b.doToggleWallet(ctx, msg.Chat.ID, id)
	case "copy":
		b.sendCopytrading(msg.Chat.ID)
	case "addtrader":
		args := strings.Fields(msg.CommandArguments())
		if len(args) < 1 {
			b.sendText(msg.Chat.ID, RenderError("Usage: /addtrader &lt;address&gt; [label] [alloc_pct]"))
			return
		}
		b.doAddTrader(ctx, msg.Chat.ID, args)
	case "removetrader":
		addr := strings.TrimSpace(msg.CommandArguments())
		if addr == "" {
			b.sendText(msg.Chat.ID, RenderError("Usage: /removetrader &lt;address&gt;"))
			return
		}
		b.doRemoveTrader(ctx, msg.Chat.ID, addr)
	case "toggletrader":
		addr := strings.TrimSpace(msg.CommandArguments())
		if addr == "" {
			b.sendText(msg.Chat.ID, RenderError("Usage: /toggletrader &lt;address&gt;"))
			return
		}
		b.doToggleTrader(ctx, msg.Chat.ID, addr)
	case "logs":
		b.sendLogs(msg.Chat.ID)
	case "settings":
		b.sendSettings(msg.Chat.ID)
	case "set":
		args := strings.Fields(msg.CommandArguments())
		if len(args) < 2 {
			b.sendText(msg.Chat.ID, RenderError("Usage: /set &lt;key&gt; &lt;value&gt;"))
			return
		}
		b.doSetSetting(ctx, msg.Chat.ID, args[0], strings.Join(args[1:], " "))
	default:
		b.sendText(msg.Chat.ID, "❓ Неизвестная команда.\n\nИспользуйте /start для главного меню.")
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
		b.sendMenu(chatID)
	case data == "cmd:overview":
		b.sendOverview(chatID)
	case data == "cmd:orders":
		b.sendTrading(chatID, "orders")
	case data == "cmd:positions":
		b.sendTrading(chatID, "positions")
	case data == "cmd:copytrading":
		b.sendCopytrading(chatID)
	case data == "cmd:logs":
		b.sendLogs(chatID)
	case data == "cmd:settings":
		b.sendSettings(chatID)
	case strings.HasPrefix(data, "settings:section:"):
		section := strings.TrimPrefix(data, "settings:section:")
		b.sendSettingsSection(chatID, section)
	case strings.HasPrefix(data, "toggle:"):
		key := strings.TrimPrefix(data, "toggle:")
		b.doToggleSetting(ctx, chatID, key)
	// Language picker — intercept before generic edit:* handler
	case data == "edit:ui.language":
		b.sendLanguagePicker(chatID)
	case strings.HasPrefix(data, "setlang:"):
		code := strings.TrimPrefix(data, "setlang:")
		b.doSetSetting(ctx, chatID, "ui.language", code)
		b.sendSettingsSection(chatID, "UI")
	case strings.HasPrefix(data, "edit:"):
		key := strings.TrimPrefix(data, "edit:")
		b.state.SetPending("edit:"+key, "")
		b.sendText(chatID, fmt.Sprintf("✏️ Введите новое значение для <code>%s</code>:\n<i>(или /menu для отмены)</i>", key))
	case data == "cmd:wallets":
		b.sendWallets(chatID)
	case strings.HasPrefix(data, "wallet:toggle:"):
		id := strings.TrimPrefix(data, "wallet:toggle:")
		b.doToggleWallet(ctx, chatID, id)
		b.sendWallets(chatID)
	case data == "cancelall:confirm":
		b.sendWithKeyboard(chatID, "⚠️ Are you sure you want to cancel ALL orders?", cancelAllConfirmKeyboard())
	case data == "cancelall:do":
		b.doCancelAll(ctx, chatID)
	case strings.HasPrefix(data, "cancel:"):
		orderID := strings.TrimPrefix(data, "cancel:")
		b.doCancelOrder(ctx, chatID, orderID)
	case strings.HasPrefix(data, "trader:toggle:"):
		addr := strings.TrimPrefix(data, "trader:toggle:")
		b.doToggleTrader(ctx, chatID, addr)
		b.sendCopytrading(chatID)
	case strings.HasPrefix(data, "trader:remove:"):
		addr := strings.TrimPrefix(data, "trader:remove:")
		b.doRemoveTrader(ctx, chatID, addr)
		b.sendCopytrading(chatID)
	case data == "addtrader:start":
		b.state.SetPending("addtrader_addr", "")
		b.sendText(chatID, "📝 Введите адрес кошелька трейдера:\n<i>(или /menu для отмены)</i>")
	case data == "cmd:trading":
		b.sendTrading(chatID, "orders")
	case data == "trading:orders":
		b.sendTrading(chatID, "orders")
	case data == "trading:positions":
		b.sendTrading(chatID, "positions")
	}
}

// --- View helpers ---

func (b *Bot) sendMenu(chatID int64) {
	orders := b.state.Orders()
	positions := b.state.Positions()
	text := RenderWelcome(b.state.Balance(), len(orders), len(positions))
	b.sendOrEdit(chatID, text, mainMenuKeyboard())
}

func (b *Bot) sendOverview(chatID int64) {
	subsystems := b.state.Subsystems()
	orders := b.state.Orders()
	positions := b.state.Positions()
	text := RenderOverview(b.state.Balance(), subsystems, len(orders), len(positions))
	b.sendOrEdit(chatID, text, backKeyboard())
}

func (b *Bot) sendTrading(chatID int64, subTab string) {
	orders := b.state.Orders()
	positions := b.state.Positions()
	text := RenderTrading(subTab, orders, positions)
	b.sendOrEdit(chatID, text, tradingKeyboard(subTab, orders))
}

func (b *Bot) sendCopytrading(chatID int64) {
	b.cfgMu.RLock()
	traders := make([]config.TraderConfig, len(b.cfg.Copytrading.Traders))
	copy(traders, b.cfg.Copytrading.Traders)
	b.cfgMu.RUnlock()

	text := RenderCopytrading(b.state.Traders())
	b.sendOrEdit(chatID, text, copytradingKeyboard(traders))
}

func (b *Bot) sendWallets(chatID int64) {
	wallets := b.state.Wallets()
	text := RenderWallets(wallets)
	b.sendOrEdit(chatID, text, walletsKeyboard(wallets))
}

func (b *Bot) doToggleWallet(_ context.Context, chatID int64, id string) {
	if b.wallets == nil {
		b.sendText(chatID, RenderError("Wallet manager unavailable"))
		return
	}
	enabled := b.wallets.WalletEnabled(id)
	if err := b.wallets.Toggle(id, !enabled); err != nil {
		b.sendText(chatID, RenderError(err.Error()))
		return
	}
	status := "disabled"
	if !enabled {
		status = "enabled"
	}
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Wallet <code>%s</code> %s.", id, status)))
}

func (b *Bot) sendLogs(chatID int64) {
	logsKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "cmd:logs"),
			tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
		),
	)
	b.sendOrEdit(chatID, RenderLogs(b.state.Logs()), logsKeyboard)
}


func (b *Bot) sendLanguagePicker(chatID int64) {
	b.cfgMu.RLock()
	currentLang := b.cfg.UI.Language
	b.cfgMu.RUnlock()

	if currentLang == "" {
		currentLang = "en"
	}
	text := "🌐 <b>Язык интерфейса</b>\n\nВыберите язык:"
	b.sendOrEdit(chatID, text, langPickerKeyboard(currentLang))
}

func (b *Bot) sendSettings(chatID int64) {
	text := "⚙️ <b>Settings</b>\n\nВыберите раздел для просмотра и редактирования:"
	b.sendOrEdit(chatID, text, settingsSectionsKeyboard())
}

// sectionKeys maps display section names to their dot-notation config keys.
var sectionKeys = map[string][]string{
	"UI":             {"ui.language"},
	"Auth":           {"auth.private_key", "auth.api_key", "auth.api_secret", "auth.passphrase", "auth.chain_id"},
	"Monitor":        {"monitor.enabled", "monitor.poll_interval_ms"},
	"Trades Monitor": {"monitor.trades.enabled", "monitor.trades.poll_interval_ms", "monitor.trades.alert_on_fill", "monitor.trades.alert_on_cancel"},
	"Trading":        {"trading.enabled", "trading.max_position_usd", "trading.slippage_pct", "trading.neg_risk"},
	"Copytrading":    {"copytrading.enabled", "copytrading.poll_interval_ms", "copytrading.size_mode"},
	"Telegram":       {"telegram.enabled", "telegram.bot_token", "telegram.admin_chat_id"},
	"Database":       {"database.enabled", "database.path"},
	"Log":            {"log.level", "log.format"},
}

func (b *Bot) sendSettingsSection(chatID int64, sectionName string) {
	keys, ok := sectionKeys[sectionName]
	if !ok {
		b.sendText(chatID, RenderError("Unknown section: "+sectionName))
		return
	}
	isAdmin := b.isAdmin(chatID)

	b.cfgMu.RLock()
	cfg := *b.cfg
	b.cfgMu.RUnlock()

	fields := make([]SettingField, 0, len(keys))
	for _, k := range keys {
		if IsSecretKey(k) && !isAdmin {
			continue
		}
		short := k
		if idx := strings.LastIndex(k, "."); idx >= 0 {
			short = k[idx+1:]
		}
		v, ok2 := GetSetting(&cfg, k)
		if !ok2 {
			continue
		}
		fields = append(fields, SettingField{Key: short, Value: v})
	}
	text := RenderSettingsSection(sectionName, fields, isAdmin)
	b.sendOrEdit(chatID, text, sectionFieldsKeyboard(sectionName, keys, &cfg, isAdmin))
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

func (b *Bot) doToggleSetting(_ context.Context, chatID int64, key string) {
	if IsSecretKey(key) && !b.isAdmin(chatID) {
		b.sendText(chatID, RenderError(fmt.Sprintf("Key %q requires admin access.", key)))
		return
	}
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	cur, ok := GetSetting(&cfgCopy, key)
	if !ok {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Unknown key: %q", key)))
		return
	}
	newVal := "true"
	if cur == "true" {
		newVal = "false"
	}
	if err := SetSetting(&cfgCopy, key, newVal); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(err.Error()))
		return
	}
	if e, ok2 := settingsMap[key]; ok2 && e.onSet != nil {
		e.onSet(newVal)
	}
	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("<code>%s</code> = <code>%s</code>  Config saved.", key, newVal)))
}

func (b *Bot) doAddTrader(_ context.Context, chatID int64, args []string) {
	addr := args[0]
	label := ""
	if len(args) > 1 {
		label = args[1]
	}
	allocPct := 5.0
	if len(args) > 2 {
		if v, err := strconv.ParseFloat(args[2], 64); err == nil {
			allocPct = v
		}
	}

	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	for _, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			b.cfgMu.Unlock()
			b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q already exists.", addr)))
			return
		}
	}

	cfgCopy.Copytrading.Traders = append(cfgCopy.Copytrading.Traders, config.TraderConfig{
		Address:        addr,
		Label:          label,
		Enabled:        true,
		AllocationPct:  allocPct,
		MaxPositionUSD: 50.0,
		SizeMode:       cfgCopy.Copytrading.SizeMode,
	})

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> added (label: %s, alloc: %.1f%%).", addr, label, allocPct)))
}

func (b *Bot) doRemoveTrader(_ context.Context, chatID int64, addr string) {
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	found := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders = append(cfgCopy.Copytrading.Traders[:i], cfgCopy.Copytrading.Traders[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q not found.", addr)))
		return
	}

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> removed.", addr)))
}

func (b *Bot) doToggleTrader(_ context.Context, chatID int64, addr string) {
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	found := false
	newState := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders[i].Enabled = !t.Enabled
			newState = cfgCopy.Copytrading.Traders[i].Enabled
			found = true
			break
		}
	}
	if !found {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q not found.", addr)))
		return
	}

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	state := "disabled"
	if newState {
		state = "enabled"
	}
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> %s.", addr, state)))
}
