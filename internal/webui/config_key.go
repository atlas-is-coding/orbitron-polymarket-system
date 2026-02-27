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
	default:
		return fmt.Errorf("unknown or read-only key: %q", key)
	}
	return nil
}

func parseBoolKey(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}
