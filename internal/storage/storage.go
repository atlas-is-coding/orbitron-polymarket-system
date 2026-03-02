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

// Store — объединённый интерфейс хранилища.
type Store interface {
	TradeStore
	OrderStore
	CopyTradeStore
	WalletStatsStore
	Close() error
}
