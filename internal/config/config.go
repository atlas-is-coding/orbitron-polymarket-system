package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// WalletConfig — настройки одного кошелька.
type WalletConfig struct {
	ID         string `toml:"id"`
	Label      string `toml:"label"`
	PrivateKey string `toml:"private_key"`
	APIKey     string `toml:"api_key"`
	APISecret  string `toml:"api_secret"`
	Passphrase string `toml:"passphrase"`
	ChainID    int64  `toml:"chain_id"`
	Enabled    bool   `toml:"enabled"`
	NegRisk    bool   `toml:"neg_risk"`
}

// Config — корневая структура конфигурации.
type Config struct {
	Wallets     []WalletConfig    `toml:"wallets"`
	Auth        AuthConfig        `toml:"auth"` // Deprecated: use [[wallets]]; kept for migration
	API         APIConfig         `toml:"api"`
	Trading     TradingConfig     `toml:"trading"`
	Monitor     MonitorConfig     `toml:"monitor"`
	Telegram    TelegramConfig    `toml:"telegram"`
	Database    DatabaseConfig    `toml:"database"`
	Log         LogConfig         `toml:"log"`
	Copytrading CopytradingConfig `toml:"copytrading"`
	UI          UIConfig          `toml:"ui"`
	WebUI       WebUIConfig       `toml:"webui"`
}

type APIConfig struct {
	ClobURL    string `toml:"clob_url"`
	GammaURL   string `toml:"gamma_url"`
	DataURL    string `toml:"data_url"`
	WSURL      string `toml:"ws_url"`
	TimeoutSec int    `toml:"timeout_sec"`
	MaxRetries int    `toml:"max_retries"`
}

type AuthConfig struct {
	PrivateKey string `toml:"private_key"`
	APIKey     string `toml:"api_key"`
	APISecret  string `toml:"api_secret"`
	Passphrase string `toml:"passphrase"`
	ChainID    int64  `toml:"chain_id"`
}

type TradingConfig struct {
	Enabled          bool    `toml:"enabled"`
	MaxPositionUSD   float64 `toml:"max_position_usd"`
	SlippagePct      float64 `toml:"slippage_pct"`
	DefaultOrderType string  `toml:"default_order_type"`
	NegRisk          bool    `toml:"neg_risk"`
}

type MonitorConfig struct {
	Enabled        bool               `toml:"enabled"`
	PollIntervalMs int                `toml:"poll_interval_ms"`
	Markets        []string           `toml:"markets"`
	Trades         TradesMonitorConfig `toml:"trades"`
}

// TradesMonitorConfig — конфигурация монитора сделок и позиций.
type TradesMonitorConfig struct {
	// Enabled — включить мониторинг ордеров/сделок/позиций
	Enabled bool `toml:"enabled"`
	// PollIntervalMs — интервал опроса API в миллисекундах
	PollIntervalMs int `toml:"poll_interval_ms"`
	// TrackPositions — отслеживать позиции через CLOB /positions
	TrackPositions bool `toml:"track_positions"`
	// AlertOnFill — отправлять алерт при исполнении ордера
	AlertOnFill bool `toml:"alert_on_fill"`
	// AlertOnCancel — отправлять алерт при отмене ордера
	AlertOnCancel bool `toml:"alert_on_cancel"`
	// TradesLimit — максимальное количество сделок в одном запросе
	TradesLimit int `toml:"trades_limit"`
}

type TelegramConfig struct {
	Enabled     bool   `toml:"enabled"`
	BotToken    string `toml:"bot_token"`
	AdminChatID string `toml:"admin_chat_id"` // admin: receives notifications and can control the bot
}

type DatabaseConfig struct {
	Enabled bool   `toml:"enabled"`
	Path    string `toml:"path"`
}

type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

type UIConfig struct {
	Language string `toml:"language"`
}

type WebUIConfig struct {
	Enabled   bool   `toml:"enabled"`
	Listen    string `toml:"listen"`
	JWTSecret string `toml:"jwt_secret"`
}

// CopytradingConfig — конфигурация подсистемы копитрейдинга.
type CopytradingConfig struct {
	// Enabled — включить копитрейдинг
	Enabled bool `toml:"enabled"`
	// PollIntervalMs — интервал опроса позиций трейдеров (миллисекунды)
	PollIntervalMs int `toml:"poll_interval_ms"`
	// SizeMode — глобальный метод расчёта размера: "proportional" или "fixed_pct"
	SizeMode string `toml:"size_mode"`
	// Traders — список отслеживаемых трейдеров
	Traders []TraderConfig `toml:"traders"`
}

// TraderConfig — настройки одного копируемого трейдера.
type TraderConfig struct {
	// Address — proxy-wallet адрес трейдера (из Data API)
	Address string `toml:"address"`
	// Label — метка для логов и алертов
	Label string `toml:"label"`
	// Enabled — можно временно отключить без удаления из конфига
	Enabled bool `toml:"enabled"`
	// AllocationPct — % нашего баланса, выделяемый этому трейдеру
	AllocationPct float64 `toml:"allocation_pct"`
	// MaxPositionUSD — максимальный размер одной позиции в USD
	MaxPositionUSD float64 `toml:"max_position_usd"`
	// SizeMode — переопределяет глобальный (если не пустая строка)
	SizeMode string `toml:"size_mode"`
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
