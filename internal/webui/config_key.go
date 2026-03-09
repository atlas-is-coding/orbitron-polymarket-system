package webui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

// applyConfigKey sets a dot-notation config key on cfg.
// Only non-secret, writable keys are supported.
func applyConfigKey(cfg *config.Config, key, value string) error {
	switch key {
	case "ui.language":
		cfg.UI.Language = value
	case "monitor.enabled":
		cfg.Monitor.Enabled = parseBoolKey(value)
	case "monitor.poll_interval_ms":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.Monitor.PollIntervalMs = n
	case "monitor.trades.enabled":
		cfg.Monitor.Trades.Enabled = parseBoolKey(value)
	case "monitor.trades.poll_interval_ms":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.Monitor.Trades.PollIntervalMs = n
	case "monitor.trades.track_positions":
		cfg.Monitor.Trades.TrackPositions = parseBoolKey(value)
	case "monitor.trades.trades_limit":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.Monitor.Trades.TradesLimit = n
	case "trading.enabled":
		cfg.Trading.Enabled = parseBoolKey(value)
	case "trading.max_position_usd":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		cfg.Trading.MaxPositionUSD = f
	case "trading.slippage_pct":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		cfg.Trading.SlippagePct = f
	case "copytrading.enabled":
		cfg.Copytrading.Enabled = parseBoolKey(value)
	case "copytrading.poll_interval_ms":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.Copytrading.PollIntervalMs = n
	case "copytrading.size_mode":
		cfg.Copytrading.SizeMode = value
	case "log.level":
		cfg.Log.Level = value
	case "log.format":
		cfg.Log.Format = value
	case "log.file":
		cfg.Log.File = value
	case "monitor.trades.alert_on_fill":
		cfg.Monitor.Trades.AlertOnFill = parseBoolKey(value)
	case "monitor.trades.alert_on_cancel":
		cfg.Monitor.Trades.AlertOnCancel = parseBoolKey(value)
	case "trading.default_order_type":
		cfg.Trading.DefaultOrderType = value
	case "trading.neg_risk":
		cfg.Trading.NegRisk = parseBoolKey(value)
	case "auth.chain_id":
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		cfg.Auth.ChainID = n
	case "api.timeout_sec":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.API.TimeoutSec = n
	case "api.max_retries":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.API.MaxRetries = n
	case "telegram.enabled":
		cfg.Telegram.Enabled = parseBoolKey(value)
	case "telegram.bot_token":
		cfg.Telegram.BotToken = value
	case "telegram.admin_chat_id":
		cfg.Telegram.AdminChatID = value
	case "database.enabled":
		cfg.Database.Enabled = parseBoolKey(value)
	case "database.path":
		cfg.Database.Path = value
	case "webui.enabled":
		cfg.WebUI.Enabled = parseBoolKey(value)
	case "webui.listen":
		cfg.WebUI.Listen = value
	case "webui.jwt_secret":
		cfg.WebUI.JWTSecret = value
	case "proxy.enabled":
		cfg.Proxy.Enabled = parseBoolKey(value)
	case "proxy.type":
		cfg.Proxy.Type = value
	case "proxy.addr":
		cfg.Proxy.Addr = value
	case "proxy.username":
		cfg.Proxy.Username = value
	case "proxy.password":
		cfg.Proxy.Password = value
	default:
		return fmt.Errorf("unknown or read-only key: %q", key)
	}
	return nil
}

func parseBoolKey(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}
