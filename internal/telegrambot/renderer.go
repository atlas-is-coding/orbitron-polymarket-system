package telegrambot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/health"
	"github.com/atlasdev/polytrade-bot/internal/i18n"
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
	l := i18n.T()
	sorted := make([]SubsystemStatus, len(subsystems))
	copy(sorted, subsystems)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var sb strings.Builder
	sb.WriteString(l.TgOverviewTitle + "\n\n")
	sb.WriteString(fmt.Sprintf(l.TgOverviewBalance+"\n", balance))
	sb.WriteString(fmt.Sprintf(l.TgOverviewStats+"\n\n", openOrders, positions))
	if len(sorted) > 0 {
		sb.WriteString(l.TgOverviewSubsystems + "\n")
		for _, s := range sorted {
			dot := "🔴"
			status := l.TgStatusInactive
			if s.Active {
				dot = "🟢"
				status = l.TgStatusActive
			}
			sb.WriteString(fmt.Sprintf("%s %s — %s\n", dot, s.Name, status))
		}
	}
	return sb.String()
}

// RenderHealth formats the health snapshot as Telegram HTML.
func RenderHealth(snap health.HealthSnapshot, loaded bool) string {
	l := i18n.T()
	var sb strings.Builder
	sb.WriteString("\n" + l.TgHealthTitle + "\n")
	if !loaded {
		sb.WriteString(l.TgHealthNever + "\n")
		return sb.String()
	}

	icon := func(s health.ServiceStatus) string {
		switch s {
		case health.StatusOK:
			return "✅"
		case health.StatusDegraded:
			return "⚠️"
		default:
			return "🔴"
		}
	}
	latStr := func(ms int64) string {
		if ms < 1000 {
			return fmt.Sprintf("%dms", ms)
		}
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}

	for _, svc := range snap.Services {
		if svc.Name == "Geoblock" {
			continue
		}
		errPart := ""
		if svc.Error != "" {
			errPart = " — " + svc.Error
		}
		sb.WriteString(fmt.Sprintf("%s <b>%-10s</b> %s%s\n",
			icon(svc.Status), svc.Name, latStr(svc.LatencyMs), errPart))
	}

	if snap.Geo != nil {
		if snap.Geo.Blocked {
			sb.WriteString(fmt.Sprintf("🚫 <b>%-10s</b> %s · %s\n",
				"Geoblock", snap.Geo.Country, snap.Geo.IP))
		} else {
			sb.WriteString(fmt.Sprintf("✅ <b>%-10s</b> %s\n",
				"Geoblock", snap.Geo.Country))
		}
	}

	age := int(time.Since(snap.UpdatedAt).Seconds())
	sb.WriteString(fmt.Sprintf("\n<i>"+l.TgHealthUpdated+"</i>", age))
	return sb.String()
}

// RenderOrders formats the orders list. Order IDs are included for cancel buttons.
func RenderOrders(rows []tui.OrderRow) string {
	l := i18n.T()
	if len(rows) == 0 {
		return l.TgOrdersEmpty
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(l.TgOrdersTitle+"\n\n", len(rows)))
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
	l := i18n.T()
	if len(rows) == 0 {
		return l.TgPositionsEmpty
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(l.TgPositionsTitle+"\n\n", len(rows)))
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
// trades is the recent copy-trade feed (last N lines); may be nil/empty.
func RenderCopytrading(traders []tui.TraderRow, trades []string) string {
	l := i18n.T()
	var sb strings.Builder
	if len(traders) == 0 {
		sb.WriteString(l.TgCopyEmpty)
	} else {
		sb.WriteString(fmt.Sprintf(l.TgCopyTitle+"\n\n", len(traders)))
		for _, t := range traders {
			addr := t.Address
			if len(addr) > 12 {
				addr = addr[:6] + "…" + addr[len(addr)-4:]
			}
			sb.WriteString(fmt.Sprintf("• %s (%s)  %s  alloc: %s\n", t.Label, addr, t.Status, t.AllocPct))
		}
	}
	if len(trades) > 0 {
		sb.WriteString("\n" + l.TgCopyRecentTrades + "\n")
		for _, line := range trades {
			sb.WriteString("  " + line + "\n")
		}
	}
	return sb.String()
}

// RenderLogs formats the last 20 log lines.
func RenderLogs(lines []string) string {
	l := i18n.T()
	if len(lines) == 0 {
		return l.TgLogsEmpty
	}
	last := lines
	if len(last) > 20 {
		last = last[len(last)-20:]
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(l.TgLogsTitle+"\n\n<pre>", len(last)))
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
	l := i18n.T()
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
				display = "<i>" + l.TgNotSet + "</i>"
			}
		} else if display == "" {
			display = "<i>" + l.TgNotSet + "</i>"
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
	l := i18n.T()
	return fmt.Sprintf(
		"🤖 <b>polytrade-bot</b>\n\n"+
			l.TgWelcome+"\n\n"+
			l.TgOverviewBalance+"\n"+
			l.TgOverviewStats+"\n\n"+
			l.TgChooseSection,
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

// RenderMarkets formats the markets list as an HTML Telegram message.
// tagSlug filters by tag (empty = all). totalTags is used to indicate filter state.
func RenderMarkets(mkts []gamma.Market, tagSlug string, totalTags int) string {
	l := i18n.T()
	var sb strings.Builder
	sb.WriteString(l.TgMarketsTitle)
	if tagSlug != "" {
		sb.WriteString(fmt.Sprintf(" — <i>%s</i>", tagSlug))
	}
	sb.WriteString("\n\n")

	if len(mkts) == 0 {
		sb.WriteString(l.TgMarketsEmpty)
		if totalTags > 0 {
			sb.WriteString(l.TgMarketsFilterHint)
		}
		return sb.String()
	}

	shown := mkts
	if len(shown) > 10 {
		shown = shown[:10]
	}
	sb.WriteString(fmt.Sprintf(l.TgMarketsShowing, len(shown), len(mkts)))
	if len(mkts) > 10 {
		sb.WriteString(l.TgMarketsTapHint)
	}
	return sb.String()
}

// RenderMarketDetail formats a single market as an HTML Telegram message.
// Handles both binary (YES/NO) and multi-outcome (categorical) markets.
func RenderMarketDetail(m gamma.Market) string {
	l := i18n.T()
	var sb strings.Builder
	sb.WriteString(l.TgMarketDetail + "\n\n")
	sb.WriteString(fmt.Sprintf("<b>%s</b>\n\n", m.Question))

	// Prices — iterate all outcomes (handles binary and categorical markets)
	if len(m.OutcomePrices) > 0 && len(m.Outcomes) > 0 {
		n := len(m.OutcomePrices)
		if len(m.Outcomes) < n {
			n = len(m.Outcomes)
		}
		for i := 0; i < n; i++ {
			p, _ := strconv.ParseFloat(string(m.OutcomePrices[i]), 64)
			outcome := string(m.Outcomes[i])
			sb.WriteString(fmt.Sprintf("• %s: <b>%.1f%%</b>\n", outcome, p*100))
		}
		sb.WriteString("\n")
	} else if len(m.OutcomePrices) > 0 {
		// Fallback: prices but no outcome labels
		p, _ := strconv.ParseFloat(string(m.OutcomePrices[0]), 64)
		sb.WriteString(fmt.Sprintf("YES: <b>%.1f%%</b>  NO: <b>%.1f%%</b>\n\n", p*100, (1-p)*100))
	}

	// Stats
	liq := float64(m.Liquidity)
	vol := float64(m.Volume)
	if liq > 0 {
		sb.WriteString(fmt.Sprintf(l.TgMarketLiquidity+"\n", liq))
	}
	if vol > 0 {
		sb.WriteString(fmt.Sprintf(l.TgMarketVolume+"\n", vol))
	}

	if m.EndDateISO != "" {
		end := m.EndDateISO
		if len(end) > 10 {
			end = end[:10]
		}
		sb.WriteString(fmt.Sprintf(l.TgMarketEnds+"\n", end))
	}

	if m.Category != "" {
		sb.WriteString(fmt.Sprintf(l.TgMarketCategory+"\n", m.Category))
	}

	sb.WriteString(fmt.Sprintf("\n<code>%s</code>", m.ConditionID))
	return sb.String()
}

// RenderWallets formats the wallet list as an HTML Telegram message.
func RenderWallets(wallets []WalletEntry) string {
	l := i18n.T()
	if len(wallets) == 0 {
		return l.TgWalletsEmpty
	}
	var sb strings.Builder
	sb.WriteString(l.TgWalletsTitle + "\n\n")

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
	sb.WriteString(fmt.Sprintf(l.TgWalletTotal+"\n\n",
		totalBal, pnlSign, totalPnL, activeCount, len(wallets)))

	for _, w := range wallets {
		status := l.TgStatusOff
		if w.Enabled {
			status = l.TgStatusOn
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
