package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// WalletConfig — настройки одного кошелька.
type WalletConfig struct {
	ID         string `toml:"id"           json:"id"`
	Label      string `toml:"label"        json:"label"`
	PrivateKey string `toml:"private_key"  json:"private_key"`
	APIKey     string `toml:"api_key"      json:"api_key"`
	APISecret  string `toml:"api_secret"   json:"api_secret"`
	Passphrase string `toml:"passphrase"   json:"passphrase"`
	ChainID    int64  `toml:"chain_id"     json:"chain_id"`
	Enabled    bool   `toml:"enabled"      json:"enabled"`
	Primary    bool   `toml:"primary"      json:"primary"`
	NegRisk    bool   `toml:"neg_risk"     json:"neg_risk"`
}

// Config — корневая структура конфигурации.
type Config struct {
	Wallets     []WalletConfig    `toml:"wallets"     json:"wallets"`
	Auth        AuthConfig        `toml:"auth"        json:"auth"` // Deprecated: use [[wallets]]; kept for migration
	API         APIConfig         `toml:"api"         json:"api"`
	Trading     TradingConfig     `toml:"trading"     json:"trading"`
	Monitor     MonitorConfig     `toml:"monitor"     json:"monitor"`
	Telegram    TelegramConfig    `toml:"telegram"    json:"telegram"`
	Database    DatabaseConfig    `toml:"database"    json:"database"`
	Log         LogConfig         `toml:"log"         json:"log"`
	Copytrading CopytradingConfig `toml:"copytrading" json:"copytrading"`
	UI          UIConfig          `toml:"ui"          json:"ui"`
	WebUI       WebUIConfig       `toml:"webui"       json:"webui"`
	Proxy       ProxyConfig       `toml:"proxy"       json:"proxy"`
}

type APIConfig struct {
	ClobURL    string `toml:"clob_url"    json:"clob_url"`
	GammaURL   string `toml:"gamma_url"   json:"gamma_url"`
	DataURL    string `toml:"data_url"    json:"data_url"`
	WSURL      string `toml:"ws_url"      json:"ws_url"`
	TimeoutSec int    `toml:"timeout_sec" json:"timeout_sec"`
	MaxRetries int    `toml:"max_retries" json:"max_retries"`
}

type AuthConfig struct {
	PrivateKey string `toml:"private_key" json:"private_key"`
	APIKey     string `toml:"api_key"     json:"api_key"`
	APISecret  string `toml:"api_secret"  json:"api_secret"`
	Passphrase string `toml:"passphrase"  json:"passphrase"`
	ChainID    int64  `toml:"chain_id"    json:"chain_id"`
}

type TradingConfig struct {
	Enabled          bool             `toml:"enabled"            json:"enabled"`
	MaxPositionUSD   float64          `toml:"max_position_usd"   json:"max_position_usd"`
	SlippagePct      float64          `toml:"slippage_pct"       json:"slippage_pct"`
	DefaultOrderType string           `toml:"default_order_type" json:"default_order_type"`
	NegRisk          bool             `toml:"neg_risk"           json:"neg_risk"`
	Strategies       StrategiesConfig `toml:"strategies"         json:"strategies"`
	Risk             RiskConfig       `toml:"risk"               json:"risk"`
}

type MonitorConfig struct {
	Enabled        bool                `toml:"enabled"         json:"enabled"`
	PollIntervalMs int                 `toml:"poll_interval_ms" json:"poll_interval_ms"`
	Markets        []string            `toml:"markets"         json:"markets"`
	Trades         TradesMonitorConfig `toml:"trades"          json:"trades"`
}

// TradesMonitorConfig — конфигурация монитора сделок и позиций.
type TradesMonitorConfig struct {
	// Enabled — включить мониторинг ордеров/сделок/позиций
	Enabled bool `toml:"enabled" json:"enabled"`
	// PollIntervalMs — интервал опроса API в миллисекундах
	PollIntervalMs int `toml:"poll_interval_ms" json:"poll_interval_ms"`
	// TrackPositions — отслеживать позиции через CLOB /positions
	TrackPositions bool `toml:"track_positions" json:"track_positions"`
	// AlertOnFill — отправлять алерт при исполнении ордера
	AlertOnFill bool `toml:"alert_on_fill" json:"alert_on_fill"`
	// AlertOnCancel — отправлять алерт при отмене ордера
	AlertOnCancel bool `toml:"alert_on_cancel" json:"alert_on_cancel"`
	// TradesLimit — максимальное количество сделок в одном запросе
	TradesLimit int `toml:"trades_limit" json:"trades_limit"`
}

type TelegramConfig struct {
	Enabled     bool   `toml:"enabled"       json:"enabled"`
	BotToken    string `toml:"bot_token"     json:"bot_token"`
	AdminChatID string `toml:"admin_chat_id" json:"admin_chat_id"`
}

type DatabaseConfig struct {
	Enabled bool   `toml:"enabled" json:"enabled"`
	Path    string `toml:"path"    json:"path"`
}

type LogConfig struct {
	Level  string `toml:"level"  json:"level"`
	Format string `toml:"format" json:"format"`
	File   string `toml:"file"   json:"file"`
}

type UIConfig struct {
	Language string `toml:"language" json:"language"`
}

type WebUIConfig struct {
	Enabled   bool   `toml:"enabled"    json:"enabled"`
	Listen    string `toml:"listen"     json:"listen"`
	JWTSecret string `toml:"jwt_secret" json:"jwt_secret"`
}

// ProxyConfig — optional outbound proxy for all Polymarket API calls.
type ProxyConfig struct {
	Enabled  bool   `toml:"enabled"  json:"enabled"`
	Type     string `toml:"type"     json:"type"` // "socks5" | "http"
	Addr     string `toml:"addr"     json:"addr"` // "host:port"
	Username string `toml:"username" json:"username"`
	Password string `toml:"password" json:"password"`
}

// CopytradingConfig — конфигурация подсистемы копитрейдинга.
type CopytradingConfig struct {
	// Enabled — включить копитрейдинг
	Enabled bool `toml:"enabled" json:"enabled"`
	// PollIntervalMs — интервал опроса позиций трейдеров (миллисекунды)
	PollIntervalMs int `toml:"poll_interval_ms" json:"poll_interval_ms"`
	// SizeMode — глобальный метод расчёта размера: "proportional" или "fixed_pct"
	SizeMode string `toml:"size_mode" json:"size_mode"`
	// Traders — список отслеживаемых трейдеров
	Traders []TraderConfig `toml:"traders" json:"traders"`
}

// TraderConfig — настройки одного копируемого трейдера.
type TraderConfig struct {
	// Address — proxy-wallet адрес трейдера (из Data API)
	Address string `toml:"address" json:"address"`
	// Label — метка для логов и алертов
	Label string `toml:"label" json:"label"`
	// Enabled — можно временно отключить без удаления из конфига
	Enabled bool `toml:"enabled" json:"enabled"`
	// AllocationPct — % нашего баланса, выделяемый этому трейдеру
	AllocationPct float64 `toml:"allocation_pct" json:"allocation_pct"`
	// MaxPositionUSD — максимальный размер одной позиции в USD
	MaxPositionUSD float64 `toml:"max_position_usd" json:"max_position_usd"`
	// SizeMode — переопределяет глобальный (если не пустая строка)
	SizeMode string `toml:"size_mode" json:"size_mode"`
}

// StrategiesConfig — конфигурация всех торговых стратегий.
type StrategiesConfig struct {
	Arbitrage    ArbitrageConfig    `toml:"arbitrage"     json:"arbitrage"`
	MarketMaking MarketMakingConfig `toml:"market_making" json:"market_making"`
	PositiveEV   PositiveEVConfig   `toml:"positive_ev"   json:"positive_ev"`
	RisklessRate RisklessRateConfig `toml:"riskless_rate" json:"riskless_rate"`
	FadeChaos    FadeChaosConfig    `toml:"fade_chaos"    json:"fade_chaos"`
	CrossMarket  CrossMarketConfig  `toml:"cross_market"  json:"cross_market"`
}

// RiskConfig — глобальные параметры риск-менеджмента.
type RiskConfig struct {
	StopLossPct     float64 `toml:"stop_loss_pct"      json:"stop_loss_pct"`
	TakeProfitPct   float64 `toml:"take_profit_pct"    json:"take_profit_pct"`
	MaxDailyLossUSD float64 `toml:"max_daily_loss_usd" json:"max_daily_loss_usd"`
}

type ArbitrageConfig struct {
	Enabled        bool    `toml:"enabled"          json:"enabled"`
	MinProfitUSD   float64 `toml:"min_profit_usd"   json:"min_profit_usd"`
	MaxPositionUSD float64 `toml:"max_position_usd" json:"max_position_usd"`
	PollIntervalMs int     `toml:"poll_interval_ms" json:"poll_interval_ms"`
	ExecuteOrders  bool    `toml:"execute_orders"   json:"execute_orders"`
}

type MarketMakingConfig struct {
	Enabled              bool    `toml:"enabled"                json:"enabled"`
	SpreadPct            float64 `toml:"spread_pct"             json:"spread_pct"`
	MaxPositionUSD       float64 `toml:"max_position_usd"       json:"max_position_usd"`
	RebalanceIntervalSec int     `toml:"rebalance_interval_sec" json:"rebalance_interval_sec"`
	MinLiquidityUSD      float64 `toml:"min_liquidity_usd"      json:"min_liquidity_usd"`
	ExecuteOrders        bool    `toml:"execute_orders"         json:"execute_orders"`
}

type PositiveEVConfig struct {
	Enabled         bool    `toml:"enabled"          json:"enabled"`
	MinEdgePct      float64 `toml:"min_edge_pct"     json:"min_edge_pct"`
	MinLiquidityUSD float64 `toml:"min_liquidity_usd" json:"min_liquidity_usd"`
	MaxPositionUSD  float64 `toml:"max_position_usd" json:"max_position_usd"`
	PollIntervalMs  int     `toml:"poll_interval_ms" json:"poll_interval_ms"`
	ExecuteOrders   bool    `toml:"execute_orders"   json:"execute_orders"`
}

type RisklessRateConfig struct {
	Enabled         bool    `toml:"enabled"          json:"enabled"`
	MinDurationDays int     `toml:"min_duration_days" json:"min_duration_days"`
	MaxNOPrice      float64 `toml:"max_no_price"     json:"max_no_price"`
	MaxPositionUSD  float64 `toml:"max_position_usd" json:"max_position_usd"`
	PollIntervalMs  int     `toml:"poll_interval_ms" json:"poll_interval_ms"`
	ExecuteOrders   bool    `toml:"execute_orders"   json:"execute_orders"`
}

type FadeChaosConfig struct {
	Enabled           bool    `toml:"enabled"              json:"enabled"`
	SpikeThresholdPct float64 `toml:"spike_threshold_pct"  json:"spike_threshold_pct"`
	CooldownSec       int     `toml:"cooldown_sec"         json:"cooldown_sec"`
	MaxPositionUSD    float64 `toml:"max_position_usd"     json:"max_position_usd"`
	PollIntervalMs    int     `toml:"poll_interval_ms"     json:"poll_interval_ms"`
	ExecuteOrders     bool    `toml:"execute_orders"       json:"execute_orders"`
}

type CrossMarketConfig struct {
	Enabled          bool    `toml:"enabled"            json:"enabled"`
	MinDivergencePct float64 `toml:"min_divergence_pct" json:"min_divergence_pct"`
	MaxPositionUSD   float64 `toml:"max_position_usd"   json:"max_position_usd"`
	PollIntervalMs   int     `toml:"poll_interval_ms"   json:"poll_interval_ms"`
	ExecuteOrders    bool    `toml:"execute_orders"     json:"execute_orders"`
}

// Save сериализует cfg в TOML и записывает в файл path (создаёт или перезаписывает).
func Save(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("config: save %q: %w", path, err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// Load читает и парсит TOML-конфиг из указанного файла.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file %q: %w", path, err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("config: parse toml: %w", err)
	}

	cfg.migrateAuth()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.API.ClobURL == "" {
		return fmt.Errorf("api.clob_url is required")
	}
	if c.API.TimeoutSec <= 0 {
		c.API.TimeoutSec = 10
	}
	if c.API.MaxRetries <= 0 {
		c.API.MaxRetries = 3
	}
	if c.Monitor.PollIntervalMs <= 0 {
		c.Monitor.PollIntervalMs = 1000
	}
	if c.Monitor.Trades.PollIntervalMs <= 0 {
		c.Monitor.Trades.PollIntervalMs = 5000
	}
	if c.Monitor.Trades.TradesLimit <= 0 {
		c.Monitor.Trades.TradesLimit = 50
	}
	if c.Trading.DefaultOrderType == "" {
		c.Trading.DefaultOrderType = "GTC"
	}
	if c.Copytrading.PollIntervalMs <= 0 {
		c.Copytrading.PollIntervalMs = 10000
	}
	if c.Copytrading.SizeMode == "" {
		c.Copytrading.SizeMode = "proportional"
	}
	if c.Trading.Risk.StopLossPct <= 0 {
		c.Trading.Risk.StopLossPct = 20.0
	}
	if c.Trading.Risk.TakeProfitPct <= 0 {
		c.Trading.Risk.TakeProfitPct = 50.0
	}
	if c.Trading.Risk.MaxDailyLossUSD <= 0 {
		c.Trading.Risk.MaxDailyLossUSD = 100.0
	}
	if c.Trading.Strategies.Arbitrage.PollIntervalMs <= 0 {
		c.Trading.Strategies.Arbitrage.PollIntervalMs = 5000
	}
	if c.Trading.Strategies.Arbitrage.MinProfitUSD <= 0 {
		c.Trading.Strategies.Arbitrage.MinProfitUSD = 0.50
	}
	if c.Trading.Strategies.Arbitrage.MaxPositionUSD <= 0 {
		c.Trading.Strategies.Arbitrage.MaxPositionUSD = 100.0
	}
	if c.Trading.Strategies.MarketMaking.SpreadPct <= 0 {
		c.Trading.Strategies.MarketMaking.SpreadPct = 2.0
	}
	if c.Trading.Strategies.MarketMaking.RebalanceIntervalSec <= 0 {
		c.Trading.Strategies.MarketMaking.RebalanceIntervalSec = 30
	}
	if c.Trading.Strategies.MarketMaking.MaxPositionUSD <= 0 {
		c.Trading.Strategies.MarketMaking.MaxPositionUSD = 200.0
	}
	if c.Trading.Strategies.MarketMaking.MinLiquidityUSD <= 0 {
		c.Trading.Strategies.MarketMaking.MinLiquidityUSD = 10000.0
	}
	if c.Trading.Strategies.PositiveEV.MinEdgePct <= 0 {
		c.Trading.Strategies.PositiveEV.MinEdgePct = 5.0
	}
	if c.Trading.Strategies.PositiveEV.MinLiquidityUSD <= 0 {
		c.Trading.Strategies.PositiveEV.MinLiquidityUSD = 5000.0
	}
	if c.Trading.Strategies.PositiveEV.MaxPositionUSD <= 0 {
		c.Trading.Strategies.PositiveEV.MaxPositionUSD = 50.0
	}
	if c.Trading.Strategies.PositiveEV.PollIntervalMs <= 0 {
		c.Trading.Strategies.PositiveEV.PollIntervalMs = 30000
	}
	if c.Trading.Strategies.RisklessRate.MinDurationDays <= 0 {
		c.Trading.Strategies.RisklessRate.MinDurationDays = 30
	}
	if c.Trading.Strategies.RisklessRate.MaxNOPrice <= 0 {
		c.Trading.Strategies.RisklessRate.MaxNOPrice = 0.05
	}
	if c.Trading.Strategies.RisklessRate.MaxPositionUSD <= 0 {
		c.Trading.Strategies.RisklessRate.MaxPositionUSD = 50.0
	}
	if c.Trading.Strategies.RisklessRate.PollIntervalMs <= 0 {
		c.Trading.Strategies.RisklessRate.PollIntervalMs = 60000
	}
	if c.Trading.Strategies.FadeChaos.SpikeThresholdPct <= 0 {
		c.Trading.Strategies.FadeChaos.SpikeThresholdPct = 10.0
	}
	if c.Trading.Strategies.FadeChaos.CooldownSec <= 0 {
		c.Trading.Strategies.FadeChaos.CooldownSec = 300
	}
	if c.Trading.Strategies.FadeChaos.MaxPositionUSD <= 0 {
		c.Trading.Strategies.FadeChaos.MaxPositionUSD = 50.0
	}
	if c.Trading.Strategies.FadeChaos.PollIntervalMs <= 0 {
		c.Trading.Strategies.FadeChaos.PollIntervalMs = 10000
	}
	if c.Trading.Strategies.CrossMarket.MinDivergencePct <= 0 {
		c.Trading.Strategies.CrossMarket.MinDivergencePct = 5.0
	}
	if c.Trading.Strategies.CrossMarket.MaxPositionUSD <= 0 {
		c.Trading.Strategies.CrossMarket.MaxPositionUSD = 75.0
	}
	if c.Trading.Strategies.CrossMarket.PollIntervalMs <= 0 {
		c.Trading.Strategies.CrossMarket.PollIntervalMs = 30000
	}
	for i := range c.Wallets {
		if c.Wallets[i].ChainID == 0 {
			c.Wallets[i].ChainID = 137 // default: Polygon mainnet
		}
	}
	for i := range c.Copytrading.Traders {
		if c.Copytrading.Traders[i].SizeMode == "" {
			c.Copytrading.Traders[i].SizeMode = c.Copytrading.SizeMode
		}
		if c.Copytrading.Traders[i].MaxPositionUSD <= 0 {
			c.Copytrading.Traders[i].MaxPositionUSD = 50.0
		}
		if c.Copytrading.Traders[i].AllocationPct <= 0 {
			c.Copytrading.Traders[i].AllocationPct = 5.0
		}
	}
	if c.UI.Language == "" {
		c.UI.Language = "en"
	}
	if c.WebUI.Listen == "" {
		c.WebUI.Listen = "127.0.0.1:8080"
	}
	if c.WebUI.JWTSecret == "" {
		c.WebUI.JWTSecret = "change-me-in-production"
	}
	return nil
}
