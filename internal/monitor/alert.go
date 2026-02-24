// Package monitor отслеживает рынки и генерирует алерты.
package monitor

import "github.com/atlasdev/polytrade-bot/internal/api/gamma"

// AlertType — тип алерта.
type AlertType string

const (
	// AlertPriceChange — цена изменилась более чем на threshold.
	AlertPriceChange AlertType = "price_change"
	// AlertLowLiquidity — ликвидность упала ниже threshold.
	AlertLowLiquidity AlertType = "low_liquidity"
	// AlertMarketClosed — рынок закрылся.
	AlertMarketClosed AlertType = "market_closed"
	// AlertHighVolume — необычно высокий объём торгов.
	AlertHighVolume AlertType = "high_volume"
)

// Alert — уведомление о событии на рынке.
type Alert struct {
	Type    AlertType
	Market  *gamma.Market
	Message string
	// Дополнительные данные алерта
	Data map[string]interface{}
}

// Rule — правило для генерации алерта.
type Rule struct {
	// Тип алерта который генерирует это правило
	AlertType AlertType
	// Порог срабатывания (смысл зависит от типа)
	Threshold float64
	// Метка для логов
	Label string
}

// DefaultRules — набор стандартных правил мониторинга.
var DefaultRules = []Rule{
	{AlertType: AlertPriceChange, Threshold: 0.05, Label: "price_change_5pct"},
	{AlertType: AlertLowLiquidity, Threshold: 1000, Label: "low_liquidity_1k"},
	{AlertType: AlertHighVolume, Threshold: 100000, Label: "high_volume_100k"},
}
