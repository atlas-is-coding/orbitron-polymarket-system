package telegrambot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/auth"
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
			tgbotapi.NewInlineKeyboardButtonData("🏪 Markets", "cmd:markets"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Logs", "cmd:logs"),
		),
		tgbotapi.NewInlineKeyboardRow(
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
			tgbotapi.NewInlineKeyboardButtonData("🗑 Remove", "wallet:remove:"+w.ID),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Add Wallet", "wallet:add:start"),
	))
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
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit", "trader:edit:"+addr),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Remove", "trader:remove:"+addr),
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
func sectionFieldsKeyboard(_ string, keys []string, cfg *config.Config, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
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

// categoryIcon returns an emoji for known category slugs.
func categoryIcon(slug string) string {
	icons := map[string]string{
		"politics":    "🏛",
		"us-politics": "🏛",
		"sports":      "⚽",
		"crypto":      "🔮",
		"science":     "🔬",
		"business":    "💼",
		"culture":     "🎭",
		"tech":        "💻",
		"weather":     "🌦",
		"entertainment": "🎬",
		"economics":   "📈",
		"world":       "🌍",
		"nba":         "🏀",
		"nfl":         "🏈",
		"soccer":      "⚽",
	}
	if icon, ok := icons[slug]; ok {
		return icon + " "
	}
	return ""
}

// marketsListKeyboard builds the Markets list view keyboard.
// Shows up to 5 tag filter buttons and up to 10 market items.
func marketsListKeyboard(mkts []gamma.Market, tags []gamma.Tag, currentTag string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Tag filter row(s) — up to 5 tags, 2 per row
	if len(tags) > 0 {
		shown := tags
		if len(shown) > 5 {
			shown = shown[:5]
		}
		for i := 0; i < len(shown); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			t1 := shown[i]
			label1 := categoryIcon(t1.Slug) + t1.Label
			if t1.Slug == currentTag {
				label1 = "✓ " + label1
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label1, "markets:tag:"+t1.Slug))
			if i+1 < len(shown) {
				t2 := shown[i+1]
				label2 := categoryIcon(t2.Slug) + t2.Label
				if t2.Slug == currentTag {
					label2 = "✓ " + label2
				}
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(label2, "markets:tag:"+t2.Slug))
			}
			rows = append(rows, row)
		}
		// "All" filter
		allLabel := "🌐 All markets"
		if currentTag == "" {
			allLabel = "✓ 🌐 All markets"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(allLabel, "markets:tag:"),
		))
	}

	// Market items — up to 10
	shown := mkts
	if len(shown) > 10 {
		shown = shown[:10]
	}
	for i, m := range shown {
		q := m.Question
		if len(q) > 38 {
			q = q[:35] + "…"
		}
		suffix := ""
		nOutcomes := len(m.OutcomePrices)
		switch {
		case nOutcomes == 2:
			// Binary YES/NO — show YES price
			suffix = " [" + string(m.OutcomePrices[0]) + "]"
		case nOutcomes > 2:
			// Categorical — show number of options
			suffix = fmt.Sprintf(" [%d opts]", nOutcomes)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(q+suffix, fmt.Sprintf("market:detail:%d", i)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// marketDetailKeyboard builds the market detail view keyboard.
// conditionID is used to route YES/NO quick buy callbacks.
// yesPrice and noPrice are pre-filled price strings (e.g. "0.72"), "" if unavailable.
func marketDetailKeyboard(conditionID, yesPrice, noPrice string) tgbotapi.InlineKeyboardMarkup {
	yesLabel := "💚 YES"
	if yesPrice != "" {
		yesLabel = "💚 YES " + yesPrice
	}
	noLabel := "❤️ NO"
	if noPrice != "" {
		noLabel = "❤️ NO " + noPrice
	}
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(yesLabel, "quickbuy:YES:"+conditionID),
			tgbotapi.NewInlineKeyboardButtonData(noLabel, "quickbuy:NO:"+conditionID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Full Order", "order:start"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 Set Alert", "market:alert"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Markets", "cmd:markets"),
			tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
		),
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func orderSideKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 YES (Buy)", "order:side:YES"),
			tgbotapi.NewInlineKeyboardButtonData("📉 NO (Sell)", "order:side:NO"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Market", "market:back"),
		),
	)
}

func orderTypeKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("GTC (Good Till Cancel)", "order:type:GTC"),
			tgbotapi.NewInlineKeyboardButtonData("FOK (Fill or Kill)", "order:type:FOK"),
		),
	)
}

func orderWalletKeyboard(wallets []WalletEntry) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, w := range wallets {
		if !w.Enabled {
			continue
		}
		label := w.Label
		if label == "" {
			label = w.ID
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👛 "+label, "order:wallet:"+w.ID),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func orderConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "order:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "market:back"),
		),
	)
}

func quickbuyConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "quickbuy:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "market:back"),
		),
	)
}

// alertDirectionKeyboard builds the alert direction picker keyboard.
func alertDirectionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Above (price rises above)", "alert:above"),
			tgbotapi.NewInlineKeyboardButtonData("📉 Below (price falls below)", "alert:below"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("← Market", "market:back"),
			tgbotapi.NewInlineKeyboardButtonData("← Главное меню", "cmd:menu"),
		),
	)
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
	case "markets":
		b.sendMarkets(msg.Chat.ID, "")
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
	case data == "wallet:add:start":
		b.state.SetPending("wallet_add_key", "")
		b.sendText(chatID, "🔑 Введите <b>private key</b> кошелька (hex, без 0x префикса):\n<i>(или /menu для отмены)</i>")
	case strings.HasPrefix(data, "wallet:remove:"):
		id := strings.TrimPrefix(data, "wallet:remove:")
		b.state.SetPending("wallet_remove_confirm", id)
		b.sendText(chatID, fmt.Sprintf("⚠️ Удалить кошелёк <code>%s</code>? Введите <b>yes</b> для подтверждения:", id))
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
	case strings.HasPrefix(data, "trader:edit:"):
		addr := strings.TrimPrefix(data, "trader:edit:")
		b.state.SetPending("edittrader_label", addr)
		b.sendText(chatID, fmt.Sprintf("✏️ <b>Edit Trader</b> <code>%s</code>\n\nВведите новый label (или <code>-</code> чтобы оставить пустым):", addr))
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

	// Markets
	case data == "cmd:markets":
		b.sendMarkets(chatID, "")
	case strings.HasPrefix(data, "markets:tag:"):
		slug := strings.TrimPrefix(data, "markets:tag:")
		b.sendMarkets(chatID, slug)
	case strings.HasPrefix(data, "market:detail:"):
		idxStr := strings.TrimPrefix(data, "market:detail:")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			b.sendText(chatID, RenderError("Invalid market index"))
			return
		}
		m, ok := b.state.ViewMarket(idx)
		if !ok {
			b.sendText(chatID, RenderError("Market not found. Please refresh the market list."))
			return
		}
		b.sendMarketDetail(chatID, m)
	case data == "market:alert":
		// Ask for direction; conditionID is in pendingData
		_, condID := b.state.Pending()
		if condID == "" {
			b.sendText(chatID, RenderError("Market context lost. Please reopen the market."))
			return
		}
		b.sendOrEdit(chatID, "🔔 <b>Set Alert</b>\n\nВыберите направление:", alertDirectionKeyboard())
	case data == "alert:above":
		_, condID := b.state.Pending()
		if condID == "" {
			b.sendText(chatID, RenderError("Market context lost. Please reopen the market."))
			return
		}
		b.state.SetPending("alert_threshold", "above|"+condID)
		b.sendText(chatID, "📈 Введите цену порога (0.01–0.99), <b>выше</b> которого придёт алерт:\n<i>(или /menu для отмены)</i>")
	case data == "alert:below":
		_, condID := b.state.Pending()
		if condID == "" {
			b.sendText(chatID, RenderError("Market context lost. Please reopen the market."))
			return
		}
		b.state.SetPending("alert_threshold", "below|"+condID)
		b.sendText(chatID, "📉 Введите цену порога (0.01–0.99), <b>ниже</b> которого придёт алерт:\n<i>(или /menu для отмены)</i>")
	case data == "order:start":
		_, condID := b.state.Pending()
		if condID == "" {
			b.sendText(chatID, RenderError("Market context lost. Reopen the market."))
			return
		}
		if b.mkts == nil {
			b.sendText(chatID, RenderError("Markets service unavailable"))
			return
		}
		mkt, ok := b.mkts.GetMarket(condID)
		tokenID := ""
		if ok && len(mkt.ClobTokenIDs) > 0 {
			tokenID = string(mkt.ClobTokenIDs[0])
		}
		b.state.SetPending("order_side", condID+"|"+tokenID)
		b.sendOrEdit(chatID, "📊 <b>Place Order</b>\n\nВыберите сторону:", orderSideKeyboard())

	case strings.HasPrefix(data, "order:side:"):
		side := strings.TrimPrefix(data, "order:side:")
		_, orderData := b.state.Pending()
		parts := strings.SplitN(orderData, "|", 2)
		condID := parts[0]
		tokenID := ""
		if b.mkts != nil {
			mkt, ok := b.mkts.GetMarket(condID)
			if ok {
				if side == "YES" && len(mkt.ClobTokenIDs) > 0 {
					tokenID = string(mkt.ClobTokenIDs[0])
				} else if side == "NO" && len(mkt.ClobTokenIDs) > 1 {
					tokenID = string(mkt.ClobTokenIDs[1])
				}
			}
		}
		b.state.SetPending("order_price", condID+"|"+tokenID+"|"+side)
		b.sendText(chatID, fmt.Sprintf(
			"📊 Side: <b>%s</b>\n\nВведите цену (0.01–0.99):\n<i>(или /menu для отмены)</i>", side,
		))

	case strings.HasPrefix(data, "order:type:"):
		orderType := strings.TrimPrefix(data, "order:type:")
		_, orderData := b.state.Pending()
		wallets := b.state.Wallets()
		var enabled []WalletEntry
		for _, w := range wallets {
			if w.Enabled {
				enabled = append(enabled, w)
			}
		}
		if len(enabled) == 0 {
			b.sendText(chatID, RenderError("Нет активных кошельков."))
			return
		}
		if len(enabled) == 1 {
			b.state.SetPending("order_confirm", orderData+"|"+orderType+"|"+enabled[0].ID)
			b.sendOrderConfirm(chatID)
			return
		}
		b.state.SetPending("order_wallet", orderData+"|"+orderType)
		b.sendOrEdit(chatID, "👛 <b>Выберите кошелёк:</b>", orderWalletKeyboard(enabled))

	case strings.HasPrefix(data, "order:wallet:"):
		walletID := strings.TrimPrefix(data, "order:wallet:")
		_, orderData := b.state.Pending()
		b.state.SetPending("order_confirm", orderData+"|"+walletID)
		b.sendOrderConfirm(chatID)

	case data == "order:confirm":
		_, orderData := b.state.Pending()
		b.state.ClearPending()
		b.doPlaceOrder(ctx, chatID, orderData)

	// Quick buy — Step 1: user taps YES/NO on market detail
	case strings.HasPrefix(data, "quickbuy:YES:") || strings.HasPrefix(data, "quickbuy:NO:"):
		var side, condID string
		if s, ok := strings.CutPrefix(data, "quickbuy:YES:"); ok {
			side, condID = "YES", s
		} else {
			condID, _ = strings.CutPrefix(data, "quickbuy:NO:")
			side = "NO"
		}
		if b.mkts == nil {
			b.sendText(chatID, RenderError("Markets service unavailable"))
			return
		}
		mkt, ok := b.mkts.GetMarket(condID)
		if !ok {
			b.sendText(chatID, RenderError("Market not found. Please refresh."))
			return
		}
		tokenID := ""
		price := 0.0
		if side == "YES" && len(mkt.ClobTokenIDs) > 0 {
			tokenID = string(mkt.ClobTokenIDs[0])
		} else if side == "NO" && len(mkt.ClobTokenIDs) > 1 {
			tokenID = string(mkt.ClobTokenIDs[1])
		}
		if side == "YES" && len(mkt.OutcomePrices) > 0 {
			price, _ = strconv.ParseFloat(mkt.OutcomePrices[0], 64)
		} else if side == "NO" && len(mkt.OutcomePrices) > 1 {
			price, _ = strconv.ParseFloat(mkt.OutcomePrices[1], 64)
		}
		// pendingData: condID|tokenID|side|price
		pendingData := fmt.Sprintf("%s|%s|%s|%.4f", condID, tokenID, side, price)
		b.state.SetPending("market_quickbuy_size", pendingData)
		b.sendText(chatID, fmt.Sprintf(
			"💚 <b>Quick Buy %s</b>\n<i>%s</i>\n\nPrice: <b>%.4f</b>\n\nВведите размер ставки в USD:",
			side, mkt.Question, price,
		))

	// Quick buy — Step 3: confirm
	case data == "quickbuy:confirm":
		_, orderData := b.state.Pending()
		b.state.ClearPending()
		b.doPlaceOrder(ctx, chatID, orderData)

	case data == "market:back":
		b.sendMarkets(chatID, "")
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

	text := RenderCopytrading(b.state.Traders(), b.state.CopyTrades())
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


func (b *Bot) sendMarkets(chatID int64, tagSlug string) {
	if b.mkts == nil {
		b.sendOrEdit(chatID, "🏪 <b>Markets</b>\n\n<i>Markets service not running.</i>", backKeyboard())
		return
	}
	mkts := b.mkts.GetByTag(tagSlug)
	tags := b.mkts.Tags()
	b.state.SetViewMarkets(mkts)
	text := RenderMarkets(mkts, tagSlug, len(tags))
	b.sendOrEdit(chatID, text, marketsListKeyboard(mkts, tags, tagSlug))
}

func (b *Bot) sendMarketDetail(chatID int64, m gamma.Market) {
	// Store conditionID in pending so alert/quickbuy callbacks can retrieve it.
	b.state.SetPending("market_view", m.ConditionID)
	text := RenderMarketDetail(m)
	yesPrice, noPrice := "", ""
	if len(m.OutcomePrices) >= 2 {
		if p, err := strconv.ParseFloat(m.OutcomePrices[0], 64); err == nil {
			yesPrice = fmt.Sprintf("%.2f", p)
		}
		if p, err := strconv.ParseFloat(m.OutcomePrices[1], 64); err == nil {
			noPrice = fmt.Sprintf("%.2f", p)
		}
	} else if len(m.OutcomePrices) == 1 {
		if p, err := strconv.ParseFloat(m.OutcomePrices[0], 64); err == nil {
			yesPrice = fmt.Sprintf("%.2f", p)
		}
	}
	b.sendOrEdit(chatID, text, marketDetailKeyboard(m.ConditionID, yesPrice, noPrice))
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

func (b *Bot) doEditTrader(_ context.Context, chatID int64, addr, label string, allocPct, maxPos float64) {
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	found := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders[i].Label = label
			cfgCopy.Copytrading.Traders[i].AllocationPct = allocPct
			cfgCopy.Copytrading.Traders[i].MaxPositionUSD = maxPos
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
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()
	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf(
		"Trader <code>%s</code> updated.\nlabel: %s | alloc: %.1f%% | max: $%.0f",
		addr, label, allocPct, maxPos,
	)))
}

func (b *Bot) doAddWallet(_ context.Context, chatID int64, privateKey string) {
	l1, err := auth.NewL1Signer(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		b.sendText(chatID, RenderError("Invalid private key: "+err.Error()))
		return
	}
	addr := l1.Address()
	b.cfgMu.Lock()
	for _, wc := range b.cfg.Wallets {
		existing, err2 := auth.NewL1Signer(wc.PrivateKey)
		if err2 == nil && existing.Address() == addr {
			b.cfgMu.Unlock()
			b.sendText(chatID, RenderError("Кошелёк уже существует: "+addr))
			return
		}
	}
	id := fmt.Sprintf("w%d", time.Now().UnixMilli())
	chainID := int64(137)
	if len(b.cfg.Wallets) > 0 && b.cfg.Wallets[0].ChainID != 0 {
		chainID = b.cfg.Wallets[0].ChainID
	}
	wCfg := config.WalletConfig{
		ID:         id,
		Label:      addr[:8] + "…" + addr[len(addr)-4:],
		PrivateKey: strings.TrimPrefix(privateKey, "0x"),
		ChainID:    chainID,
		Enabled:    true,
	}
	cfgCopy := *b.cfg
	newWallets := make([]config.WalletConfig, len(b.cfg.Wallets)+1)
	copy(newWallets, b.cfg.Wallets)
	newWallets[len(b.cfg.Wallets)] = wCfg
	cfgCopy.Wallets = newWallets
	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError("Failed to save: "+err.Error()))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()
	if b.adder != nil {
		b.adder.AddInactive(wCfg)
	}
	b.bus.Send(tui.WalletAddedMsg{ID: id, Label: wCfg.Label, Enabled: true})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf(
		"Кошелёк добавлен.\nAddress: <code>%s</code>\nID: <code>%s</code>", addr, id,
	)))
}

func (b *Bot) doRemoveWallet(_ context.Context, chatID int64, id string) {
	type remover interface{ Remove(id string) error }
	if r, ok := b.wallets.(remover); ok {
		if err := r.Remove(id); err != nil {
			b.sendText(chatID, RenderError(err.Error()))
			return
		}
		b.bus.Send(tui.WalletRemovedMsg{ID: id})
		b.cfgMu.Lock()
		cfgCopy := *b.cfg
		wallets := make([]config.WalletConfig, 0, len(cfgCopy.Wallets))
		for _, w := range cfgCopy.Wallets {
			if w.ID != id {
				wallets = append(wallets, w)
			}
		}
		cfgCopy.Wallets = wallets
		_ = config.Save(b.cfgPath, &cfgCopy)
		*b.cfg = cfgCopy
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderSuccess(fmt.Sprintf("Кошелёк <code>%s</code> удалён.", id)))
		return
	}
	b.sendText(chatID, RenderError("Remove not supported by wallet manager"))
}

func (b *Bot) sendOrderConfirm(chatID int64) {
	_, orderData := b.state.Pending()
	// format: condID|tokenID|side|price|size|orderType|walletID
	parts := strings.Split(orderData, "|")
	if len(parts) < 7 {
		b.sendText(chatID, RenderError("Потеряны данные ордера. Начните заново."))
		return
	}
	side, price, size, orderType, walletID := parts[2], parts[3], parts[4], parts[5], parts[6]
	text := fmt.Sprintf(
		"📊 <b>Подтвердите ордер</b>\n\nСторона: <b>%s</b>\nЦена: <b>%s</b>\nРазмер: <b>$%s</b>\nТип: <b>%s</b>\nКошелёк: <b>%s</b>",
		side, price, size, orderType, walletID,
	)
	b.sendOrEdit(chatID, text, orderConfirmKeyboard())
}

func (b *Bot) doPlaceOrder(_ context.Context, chatID int64, orderData string) {
	if b.placer == nil {
		b.sendText(chatID, RenderError("Order placement unavailable"))
		return
	}
	parts := strings.Split(orderData, "|")
	if len(parts) < 7 {
		b.sendText(chatID, RenderError("Данные ордера повреждены."))
		return
	}
	tokenID, side, priceStr, sizeStr, orderType, walletID := parts[1], parts[2], parts[3], parts[4], parts[5], parts[6]
	price, _ := strconv.ParseFloat(priceStr, 64)
	sizeUSD, _ := strconv.ParseFloat(sizeStr, 64)

	orderID, err := b.placer.PlaceOrder(walletID, tokenID, side, orderType, price, sizeUSD)
	if err != nil {
		b.sendText(chatID, RenderError("Ошибка размещения ордера: "+err.Error()))
		return
	}
	b.sendText(chatID, RenderSuccess(fmt.Sprintf(
		"Ордер размещён!\n\nID: <code>%s</code>\nСторона: <b>%s</b> | Цена: <b>%.4f</b> | Размер: <b>$%.2f</b>",
		orderID, side, price, sizeUSD,
	)))
}
