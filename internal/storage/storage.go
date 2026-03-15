// Package storage определяет интерфейсы хранилища данных.
// Текущая реализация — SQLite. В будущем можно добавить PostgreSQL.
package storage

import (
	"context"
	"time"
)

// --- Модели ---

// TradeRecord — запись о сделке для хранения.
type TradeRecord struct {
	ID          string
	TradeID     string
	OrderID     string
	AssetID     string
	ConditionID string
	Side        string
	Price       float64
	Size        float64
	Fee         float64
	Timestamp   time.Time
}

// OrderRecord — запись об ордере.
type OrderRecord struct {
	ID          string
	AssetID     string
	ConditionID string
	Side        string
	OrderType   string
	Price       float64
	Size        float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// --- Фильтры ---

// TradeFilter — фильтр для запроса сделок.
type TradeFilter struct {
	AssetID     string
	ConditionID string
	From        time.Time
	To          time.Time
	Limit       int
}

// --- Интерфейсы ---

// TradeStore — хранилище сделок.
type TradeStore interface {
	SaveTrade(ctx context.Context, t *TradeRecord) error
	GetTrades(ctx context.Context, f TradeFilter) ([]*TradeRecord, error)
}

// OrderStore — хранилище ордеров.
type OrderStore interface {
	SaveOrder(ctx context.Context, o *OrderRecord) error
	UpdateOrderStatus(ctx context.Context, id, status string) error
	GetOrders(ctx context.Context, status string) ([]*OrderRecord, error)
}

// CopyTradeRecord — запись о скопированной сделке.
type CopyTradeRecord struct {
	ID            string
	TraderAddress string
	AssetID       string
	ConditionID   string
	Side          string    // "BUY" или "SELL"
	Size          float64
	Price         float64
	OurOrderID    string    // ID нашего ордера в CLOB
	Status        string    // "open", "closed", "failed"
	OpenedAt      time.Time
	ClosedAt      *time.Time
	PnL           *float64
}

// CopyTradeStore — хранилище скопированных сделок.
type CopyTradeStore interface {
	SaveCopyTrade(ctx context.Context, r *CopyTradeRecord) error
	UpdateCopyTrade(ctx context.Context, id, status string, closedAt *time.Time, pnl *float64) error
	GetOpenCopyTrades(ctx context.Context, traderAddress string) ([]*CopyTradeRecord, error)
	GetAllOpenCopyTrades(ctx context.Context) ([]*CopyTradeRecord, error)
}

// WalletStatsRecord — снимок баланса и P&L кошелька.
type WalletStatsRecord struct {
	WalletID   string
	FetchedAt  time.Time
	BalanceUSD float64
	PnLUSD     float64
}

// WalletStatsStore — хранилище снимков статистики кошельков.
type WalletStatsStore interface {
	SaveWalletStats(ctx context.Context, walletID string, balanceUSD, pnlUSD float64) error
	GetWalletStats(ctx context.Context, walletID string, limit int) ([]*WalletStatsRecord, error)
}

// SentAlertStore — хранилище отправленных уведомлений (для дедупликации).
type SentAlertStore interface {
	// WasAlertSent возвращает true если уведомление типа alertType для рынка conditionID
	// уже было отправлено в течение последних cooldown.
	WasAlertSent(ctx context.Context, alertType, conditionID string, cooldown time.Duration) (bool, error)
	// MarkAlertSent записывает факт отправки уведомления.
	MarkAlertSent(ctx context.Context, alertType, conditionID string) error
}

// MarketCacheRecord — снимок рынка для локального кеша.
type MarketCacheRecord struct {
	ConditionID string
	Data        string    // JSON-encoded gamma.Market
	UpdatedAt   time.Time
	FirstSeen   time.Time
}

// MarketCacheStore — кеш маркетов (для быстрого старта).
type MarketCacheStore interface {
	// UpsertMarkets вставляет или обновляет записи. first_seen не перезаписывается.
	UpsertMarkets(ctx context.Context, records []MarketCacheRecord) error
	// GetCachedMarkets возвращает все закешированные маркеты.
	GetCachedMarkets(ctx context.Context) ([]MarketCacheRecord, error)
	// GetNewMarkets возвращает маркеты, first_seen которых >= since.
	GetNewMarkets(ctx context.Context, since time.Time) ([]MarketCacheRecord, error)
}

// Store — объединённый интерфейс хранилища.
type Store interface {
	TradeStore
	OrderStore
	CopyTradeStore
	WalletStatsStore
	SentAlertStore
	MarketCacheStore // NEW
	Close() error
}
