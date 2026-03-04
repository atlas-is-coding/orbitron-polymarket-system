package telegrambot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/atlasdev/polytrade-bot/internal/tui"
)

// secretFields lists field labels that are masked for non-admins.
var secretFields = map[string]bool{
	"private_key":   true,
	"api_key":       true,
	"api_secret":    true,
	"passphrase":    true,
	"bot_token":     true,
	"admin_chat_id": true,
	"chain_id":      true,
}

// RenderOverview formats the overview page as an HTML Telegram message.
func RenderOverview(balance float64, subsystems []SubsystemStatus, openOrders, positions int) string {
	sorted := make([]SubsystemStatus, len(subsystems))
	copy(sorted, subsystems)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var sb strings.Builder
	sb.WriteString("📊 <b>Overview</b>\n\n")
	sb.WriteString(fmt.Sprintf("💰 Баланс: <b>%.2f USDC</b>\n", balance))
	sb.WriteString(fmt.Sprintf("📋 Ордеров: <b>%d</b>  |  💼 Позиций: <b>%d</b>\n\n", openOrders, positions))
	if len(sorted) > 0 {
		sb.WriteString("<b>Subsystems:</b>\n")
		for _, s := range sorted {
			dot := "🔴"
			status := "inactive"
			if s.Active {
				dot = "🟢"
				status = "active"
			}
			sb.WriteString(fmt.Sprintf("%s %s — %s\n", dot, s.Name, status))
		}
	}
	return sb.String()
}

// RenderOrders formats the orders list. Order IDs are included for cancel buttons.
func RenderOrders(rows []tui.OrderRow) string {
	if len(rows) == 0 {
		return "📋 <b>Orders</b>\n\nNo open orders."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 <b>Orders</b> (%d)\n\n", len(rows)))
	for i, r := range rows {
		market := r.Market
		if len(market) > 12 {
			market = market[:6] + "…" + market[len(market)-4:]
		}
		sb.WriteString(fmt.Sprintf(
			"%d. <b>%s</b> %s @ %s  size: %s  [%s]\n   <code>%s</code>\n",
			i+1, market, r.Side, r.Price, r.Size, r.Status, r.ID,
		))
	}
	return sb.String()
}

// RenderPositions formats the positions list.
func RenderPositions(rows []tui.PositionRow) string {
	if len(rows) == 0 {
		return "💼 <b>Positions</b>\n\nNo open positions."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💼 <b>Positions</b> (%d)\n\n", len(rows)))
	for i, r := range rows {
		market := r.Market
		if len(market) > 12 {
			market = market[:6] + "…" + market[len(market)-4:]
		}
		sb.WriteString(fmt.Sprintf(
			"%d. <b>%s</b> %s  size: %s  entry: %s  P&amp;L: %s\n",
			i+1, market, r.Side, r.Size, r.Entry, r.PnL,
		))
	}
	return sb.String()
}

// RenderCopytrading formats the copytrading status.
func RenderCopytrading(traders []tui.TraderRow) string {
	if len(traders) == 0 {
		return "🔄 <b>Copytrading</b>\n\nNo traders configured."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔄 <b>Copytrading</b> (%d traders)\n\n", len(traders)))
	for _, t := range traders {
		addr := t.Address
		if len(addr) > 12 {
			addr = addr[:6] + "…" + addr[len(addr)-4:]
		}
		sb.WriteString(fmt.Sprintf("• %s (%s)  %s  alloc: %s\n", t.Label, addr, t.Status, t.AllocPct))
	}
	return sb.String()
}

// RenderLogs formats the last 20 log lines.
func RenderLogs(lines []string) string {
	if len(lines) == 0 {
		return "📝 <b>Logs</b>\n\nNo log entries yet."
	}
	last := lines
	if len(last) > 20 {
		last = last[len(last)-20:]
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📝 <b>Logs</b> (last %d)\n\n<pre>", len(last)))
	for _, line := range last {
		line = strings.ReplaceAll(line, "&", "&amp;")
		line = strings.ReplaceAll(line, "<", "&lt;")
		line = strings.ReplaceAll(line, ">", "&gt;")
		sb.WriteString(line + "\n")
	}
	sb.WriteString("</pre>")
	return sb.String()
}

// RenderSettingsSection formats one settings section as HTML.
// isAdmin controls whether secret field values are shown in plain text.
// fields is an ordered slice of (key, value) pairs.
func RenderSettingsSection(section string, fields []SettingField, isAdmin bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚙️ <b>%s</b>\n", section))
	for _, f := range fields {
		display := f.Value
		if secretFields[f.Key] && !isAdmin {
			if len(display) > 0 {
				n := len(display)
				if n > 8 {
					n = 8
				}
				display = strings.Repeat("•", n)
			} else {
				display = "<i>not set</i>"
			}
		} else if display == "" {
			display = "<i>not set</i>"
		}
		sb.WriteString(fmt.Sprintf("  <code>%-28s</code> %s\n", f.Key+":", display))
	}
	return sb.String()
}

// SettingField holds a key-value pair for rendering.
type SettingField struct {
	Key   string
	Value string
}

// RenderWelcome formats the welcome / main menu message shown on /start.
func RenderWelcome(balance float64, openOrders, positions int) string {
	return fmt.Sprintf(
		"🤖 <b>polytrade-bot</b>\n\n"+
			"Добро пожаловать! Я помогу вам управлять торговлей на Polymarket.\n\n"+
			"💰 Баланс: <b>%.2f USDC</b>\n"+
			"📋 Ордеров: <b>%d</b>  |  💼 Позиций: <b>%d</b>\n\n"+
			"Выберите раздел:",
		balance, openOrders, positions,
	)
}

// RenderError formats an error message.
func RenderError(msg string) string {
	return fmt.Sprintf("❌ <b>Error:</b> %s", msg)
}

// RenderSuccess formats a success message.
func RenderSuccess(msg string) string {
	return fmt.Sprintf("✅ %s", msg)
}

// RenderTrading renders Orders or Positions depending on subTab.
// subTab: "orders" | "positions"
func RenderTrading(subTab string, orders []tui.OrderRow, positions []tui.PositionRow) string {
	if subTab == "positions" {
		return RenderPositions(positions)
	}
	return RenderOrders(orders)
}

// RenderWallets formats the wallet list as an HTML Telegram message.
func RenderWallets(wallets []WalletEntry) string {
	if len(wallets) == 0 {
		return "<b>👛 Wallets</b>\n\nNo wallets configured."
	}
	var sb strings.Builder
	sb.WriteString("<b>👛 Wallets</b>\n\n")

	var totalBal, totalPnL float64
	activeCount := 0
	for _, w := range wallets {
		totalBal += w.Balance
		totalPnL += w.PnL
		if w.Enabled {
			activeCount++
		}
	}
	pnlSign := "+"
	if totalPnL < 0 {
		pnlSign = ""
	}
	sb.WriteString(fmt.Sprintf("Total: <b>$%.2f</b>  P&L: <b>%s%.2f</b>  Active: <b>%d/%d</b>\n\n",
		totalBal, pnlSign, totalPnL, activeCount, len(wallets)))

	for _, w := range wallets {
		status := "🔴 OFF"
		if w.Enabled {
			status = "🟢 ON"
		}
		label := w.Label
		if label == "" {
			label = w.ID
		}
		pSign := "+"
		if w.PnL < 0 {
			pSign = ""
		}
		sb.WriteString(fmt.Sprintf("• <b>%s</b>  %s\n  Balance: $%.2f  P&L: %s%.2f\n  ID: <code>%s</code>\n\n",
			label, status, w.Balance, pSign, w.PnL, w.ID))
	}
	return strings.TrimRight(sb.String(), "\n")
}
