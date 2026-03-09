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
	Enabled          bool    `toml:"enabled"            json:"enabled"`
	MaxPositionUSD   float64 `toml:"max_position_usd"   json:"max_position_usd"`
	SlippagePct      float64 `toml:"slippage_pct"       json:"slippage_pct"`
	DefaultOrderType string  `toml:"default_order_type" json:"default_order_type"`
	NegRisk          bool    `toml:"neg_risk"           json:"neg_risk"`
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
	Type     string `toml:"type"     json:"type"`     // "socks5" | "http"
	Addr     string `toml:"addr"     json:"addr"`     // "host:port"
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
