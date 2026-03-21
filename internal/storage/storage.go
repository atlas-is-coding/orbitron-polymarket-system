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

// MarketAlertRecord — запись о пользовательском алерте на цену.
type MarketAlertRecord struct {
	ID          string
	ConditionID string
	TokenID     string
	Direction   string
	Threshold   float64
	CreatedAt   time.Time
	Triggered   bool
}

// MarketAlertStore — хранилище пользовательских алертов.
type MarketAlertStore interface {
	SaveAlert(ctx context.Context, a *MarketAlertRecord) error
	DeleteAlert(ctx context.Context, id string) error
	GetAlerts(ctx context.Context) ([]*MarketAlertRecord, error)
	MarkAlertTriggered(ctx context.Context, id string) error
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
	// Order and trade history (new interfaces with better signatures)
	OrderHistoryStore
	NotificationQueueStore
	WalletStatisticsStore

	// Market and alert stores
	CopyTradeStore
	SentAlertStore
	MarketAlertStore
	MarketCacheStore

	// Old-style stores (for backwards compatibility where still used)
	// Note: TradeStore, OrderStore, and WalletStatsStore methods are now
	// better represented in OrderHistoryStore and WalletStatisticsStore
	SaveTrade(ctx context.Context, t *TradeRecord) error
	SaveOrder(ctx context.Context, o *OrderRecord) error
	UpdateOrderStatus(ctx context.Context, id, status string) error

	Close() error
}

// --- New structures for orders history ---

// Order — расширенная запись об ордере с информацией о кошельке и GTD.
type Order struct {
	ID            string
	WalletAddress string
	ConditionID   string
	AssetID       string
	Side          string    // "BUY" or "SELL"
	OrderType     string
	Price         float64
	Size          float64
	Status        string    // "PENDING", "OPEN", "FILLED", "CANCELED", "EXPIRED"
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SyncedAt      *time.Time
}

// Trade — расширенная запись о сделке.
type Trade struct {
	ID            string
	WalletAddress string
	OrderID       string
	TradeID       string
	ConditionID   string
	AssetID       string
	Side          string
	Price         float64
	Size          float64
	Fee           float64
	Timestamp     time.Time
}

// WalletStats — статистика кошелька.
type WalletStats struct {
	WalletAddress string
	FetchedAt     time.Time
	BalanceUSD    float64
	PnLUSD        float64
	WinRate       float64   // percentage
	TotalTrades   int
	TotalVolume   float64
}

// Notification — уведомление в очереди.
type Notification struct {
	ID            string
	WalletAddress string
	EventType     string // "ORDER_PLACED", "ORDER_FILLED", "ORDER_CANCELED", etc.
	Payload       string // JSON
	Status        string // "PENDING", "SENT", "FAILED", "DELIVERED"
	RetryCount    int
	MaxRetries    int
	NextRetryAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OrderFilters — фильтры для поиска ордеров.
type OrderFilters struct {
	WalletAddress string
	ConditionID   string
	Status        string
	Side          string
	From          time.Time
	To            time.Time
	Limit         int
}

// --- Interfaces for orders history ---

// OrderHistoryStore — интерфейс для хранения истории ордеров.
type OrderHistoryStore interface {
	InsertOrder(ctx context.Context, order *Order) error
	UpdateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, id string) (*Order, error)
	GetOrders(ctx context.Context, filters OrderFilters) ([]*Order, error)
	GetExpiredOrders(ctx context.Context, before time.Time) ([]*Order, error)

	InsertTrade(ctx context.Context, trade *Trade) error
	GetTrades(ctx context.Context, walletAddress string, from, to time.Time) ([]*Trade, error)
	GetWalletTradesByCondition(ctx context.Context, walletAddress, conditionID string) ([]*Trade, error)
	GetWalletOrdersByCondition(ctx context.Context, walletAddress, conditionID string) ([]*Order, error)

	// Statistics computation
	UpdateWalletStats(ctx context.Context, walletAddress string) error
	GetWalletStatsComputed(ctx context.Context, walletAddress string) (*WalletStats, error)
}

// NotificationQueueStore — интерфейс для очереди уведомлений.
type NotificationQueueStore interface {
	EnqueueNotification(ctx context.Context, notif *Notification) error
	GetPendingNotifications(ctx context.Context, walletAddress string) ([]*Notification, error)
	UpdateNotificationStatus(ctx context.Context, id, status string, retryCount int, nextRetryAt *time.Time) error
	DeleteNotification(ctx context.Context, id string) error
}

// WalletStatisticsStore — интерфейс для статистики кошельков.
type WalletStatisticsStore interface {
	SaveWalletStats(ctx context.Context, stats *WalletStats) error
	GetWalletStats(ctx context.Context, walletAddress string, limit int) ([]*WalletStats, error)
}
