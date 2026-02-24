// Package sqlite реализует storage.Store поверх SQLite.
// Использует modernc.org/sqlite (pure Go, CGO не требуется).
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/atlasdev/polytrade-bot/internal/storage"
)

// DB — SQLite реализация storage.Store.
type DB struct {
	db *sql.DB
}

// Open открывает (или создаёт) SQLite базу данных и применяет миграции.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	d := &DB{db: db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	return d, nil
}

// Close закрывает соединение с БД.
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate создаёт таблицы если они не существуют.
func (d *DB) migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS trades (
			id           TEXT PRIMARY KEY,
			trade_id     TEXT NOT NULL,
			order_id     TEXT NOT NULL,
			asset_id     TEXT NOT NULL,
			condition_id TEXT NOT NULL,
			side         TEXT NOT NULL,
			price        REAL NOT NULL,
			size         REAL NOT NULL,
			fee          REAL NOT NULL DEFAULT 0,
			timestamp    TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS orders (
			id           TEXT PRIMARY KEY,
			asset_id     TEXT NOT NULL,
			condition_id TEXT NOT NULL,
			side         TEXT NOT NULL,
			order_type   TEXT NOT NULL,
			price        REAL NOT NULL,
			size         REAL NOT NULL,
			status       TEXT NOT NULL,
			created_at   TEXT NOT NULL,
			updated_at   TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS copy_trades (
			id             TEXT PRIMARY KEY,
			trader_address TEXT NOT NULL,
			asset_id       TEXT NOT NULL,
			condition_id   TEXT NOT NULL,
			side           TEXT NOT NULL,
			size           REAL NOT NULL,
			price          REAL NOT NULL,
			our_order_id   TEXT NOT NULL DEFAULT '',
			status         TEXT NOT NULL DEFAULT 'open',
			opened_at      TEXT NOT NULL,
			closed_at      TEXT,
			pnl            REAL
		);

		CREATE INDEX IF NOT EXISTS idx_copy_trades_open
			ON copy_trades(trader_address, asset_id, status);
	`)
	return err
}

// --- TradeStore ---

func (d *DB) SaveTrade(ctx context.Context, t *storage.TradeRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO trades (id, trade_id, order_id, asset_id, condition_id, side, price, size, fee, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.TradeID, t.OrderID, t.AssetID, t.ConditionID, t.Side,
		t.Price, t.Size, t.Fee, t.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

func (d *DB) GetTrades(ctx context.Context, f storage.TradeFilter) ([]*storage.TradeRecord, error) {
	q := `SELECT id, trade_id, order_id, asset_id, condition_id, side, price, size, fee, timestamp
	      FROM trades WHERE 1=1`
	args := []any{}
	if f.AssetID != "" {
		q += " AND asset_id = ?"
		args = append(args, f.AssetID)
	}
	if f.ConditionID != "" {
		q += " AND condition_id = ?"
		args = append(args, f.ConditionID)
	}
	if !f.From.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, f.From.UTC().Format(time.RFC3339))
	}
	if !f.To.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, f.To.UTC().Format(time.RFC3339))
	}
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.TradeRecord
	for rows.Next() {
		var t storage.TradeRecord
		var ts string
		if err := rows.Scan(&t.ID, &t.TradeID, &t.OrderID, &t.AssetID, &t.ConditionID,
			&t.Side, &t.Price, &t.Size, &t.Fee, &ts); err != nil {
			return nil, err
		}
		t.Timestamp, _ = time.Parse(time.RFC3339, ts)
		result = append(result, &t)
	}
	return result, rows.Err()
}

// --- OrderStore ---

func (d *DB) SaveOrder(ctx context.Context, o *storage.OrderRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO orders (id, asset_id, condition_id, side, order_type, price, size, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.AssetID, o.ConditionID, o.Side, o.OrderType, o.Price, o.Size, o.Status,
		o.CreatedAt.UTC().Format(time.RFC3339), o.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (d *DB) UpdateOrderStatus(ctx context.Context, id, status string) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

func (d *DB) GetOrders(ctx context.Context, status string) ([]*storage.OrderRecord, error) {
	q := `SELECT id, asset_id, condition_id, side, order_type, price, size, status, created_at, updated_at
	      FROM orders`
	args := []any{}
	if status != "" {
		q += " WHERE status = ?"
		args = append(args, status)
	}
	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.OrderRecord
	for rows.Next() {
		var o storage.OrderRecord
		var ca, ua string
		if err := rows.Scan(&o.ID, &o.AssetID, &o.ConditionID, &o.Side, &o.OrderType,
			&o.Price, &o.Size, &o.Status, &ca, &ua); err != nil {
			return nil, err
		}
		o.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, ua)
		result = append(result, &o)
	}
	return result, rows.Err()
}

// --- CopyTradeStore ---

func (d *DB) SaveCopyTrade(ctx context.Context, r *storage.CopyTradeRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO copy_trades (id, trader_address, asset_id, condition_id, side, size, price, our_order_id, status, opened_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.TraderAddress, r.AssetID, r.ConditionID, r.Side,
		r.Size, r.Price, r.OurOrderID, r.Status,
		r.OpenedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (d *DB) UpdateCopyTrade(ctx context.Context, id, status string, closedAt *time.Time, pnl *float64) error {
	var closedAtStr *string
	if closedAt != nil {
		s := closedAt.UTC().Format(time.RFC3339)
		closedAtStr = &s
	}
	_, err := d.db.ExecContext(ctx,
		`UPDATE copy_trades SET status = ?, closed_at = ?, pnl = ? WHERE id = ?`,
		status, closedAtStr, pnl, id,
	)
	return err
}

func (d *DB) GetOpenCopyTrades(ctx context.Context, traderAddress string) ([]*storage.CopyTradeRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, trader_address, asset_id, condition_id, side, size, price, our_order_id, status, opened_at, closed_at, pnl
		 FROM copy_trades WHERE trader_address = ? AND status = 'open'`,
		traderAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCopyTrades(rows)
}

func (d *DB) GetAllOpenCopyTrades(ctx context.Context) ([]*storage.CopyTradeRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, trader_address, asset_id, condition_id, side, size, price, our_order_id, status, opened_at, closed_at, pnl
		 FROM copy_trades WHERE status = 'open'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCopyTrades(rows)
}

func scanCopyTrades(rows *sql.Rows) ([]*storage.CopyTradeRecord, error) {
	var result []*storage.CopyTradeRecord
	for rows.Next() {
		var r storage.CopyTradeRecord
		var openedAt string
		var closedAt *string
		var pnl *float64
		if err := rows.Scan(&r.ID, &r.TraderAddress, &r.AssetID, &r.ConditionID,
			&r.Side, &r.Size, &r.Price, &r.OurOrderID, &r.Status,
			&openedAt, &closedAt, &pnl); err != nil {
			return nil, err
		}
		r.OpenedAt, _ = time.Parse(time.RFC3339, openedAt)
		if closedAt != nil {
			t, _ := time.Parse(time.RFC3339, *closedAt)
			r.ClosedAt = &t
		}
		r.PnL = pnl
		result = append(result, &r)
	}
	return result, rows.Err()
}

// Убедимся, что DB реализует storage.Store
var _ storage.Store = (*DB)(nil)
