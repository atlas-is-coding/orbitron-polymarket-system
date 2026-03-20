// Package sqlite реализует storage.Store поверх SQLite.
// Использует modernc.org/sqlite (pure Go, CGO не требуется).
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/atlasdev/orbitron/internal/storage"
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
	if err != nil {
		return err
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS wallet_stats (
			wallet_id   TEXT    NOT NULL,
			fetched_at  INTEGER NOT NULL,
			balance_usd REAL    NOT NULL DEFAULT 0,
			pnl_usd     REAL    NOT NULL DEFAULT 0,
			PRIMARY KEY (wallet_id, fetched_at)
		);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create wallet_stats: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS sent_alerts (
			alert_type   TEXT    NOT NULL,
			condition_id TEXT    NOT NULL,
			sent_at      INTEGER NOT NULL,
			PRIMARY KEY (alert_type, condition_id)
		);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create sent_alerts: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS market_alerts (
			id           TEXT    PRIMARY KEY,
			condition_id TEXT    NOT NULL,
			token_id     TEXT    NOT NULL,
			direction    TEXT    NOT NULL,
			threshold    REAL    NOT NULL,
			created_at   TEXT    NOT NULL,
			triggered    INTEGER NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create market_alerts: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS markets_cache (
			condition_id TEXT    PRIMARY KEY,
			data         TEXT    NOT NULL,
			updated_at   INTEGER NOT NULL,
			first_seen   INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_markets_cache_first_seen
			ON markets_cache(first_seen);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create markets_cache: %w", err)
	}
	// Orders history schema (new)
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS wallets (
			address TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create wallets: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS orders_history (
			id TEXT PRIMARY KEY,
			wallet_address TEXT NOT NULL,
			condition_id TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			side TEXT NOT NULL,
			order_type TEXT NOT NULL,
			price REAL NOT NULL,
			size REAL NOT NULL,
			status TEXT NOT NULL,
			expires_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			synced_at TEXT,
			FOREIGN KEY (wallet_address) REFERENCES wallets(address)
		);
		CREATE INDEX IF NOT EXISTS idx_orders_wallet_status
			ON orders_history(wallet_address, status, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_orders_condition
			ON orders_history(condition_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_orders_expires
			ON orders_history(expires_at)
			WHERE expires_at IS NOT NULL;
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create orders_history: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS trades_history (
			id TEXT PRIMARY KEY,
			wallet_address TEXT NOT NULL,
			order_id TEXT NOT NULL,
			trade_id TEXT NOT NULL,
			condition_id TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			side TEXT NOT NULL,
			price REAL NOT NULL,
			size REAL NOT NULL,
			fee REAL NOT NULL DEFAULT 0,
			timestamp TEXT NOT NULL,
			FOREIGN KEY (wallet_address) REFERENCES wallets(address),
			FOREIGN KEY (order_id) REFERENCES orders_history(id)
		);
		CREATE INDEX IF NOT EXISTS idx_trades_wallet_time
			ON trades_history(wallet_address, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_trades_condition
			ON trades_history(condition_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_trades_order
			ON trades_history(order_id);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create trades_history: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS wallet_statistics (
			wallet_address TEXT NOT NULL,
			fetched_at INTEGER NOT NULL,
			balance_usd REAL NOT NULL DEFAULT 0,
			pnl_usd REAL NOT NULL DEFAULT 0,
			win_rate REAL DEFAULT 0,
			total_trades INTEGER DEFAULT 0,
			total_volume REAL DEFAULT 0,
			PRIMARY KEY (wallet_address, fetched_at),
			FOREIGN KEY (wallet_address) REFERENCES wallets(address)
		);
		CREATE INDEX IF NOT EXISTS idx_wallet_stats_recent
			ON wallet_statistics(wallet_address, fetched_at DESC);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create wallet_statistics: %w", err)
	}
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications_queue (
			id TEXT PRIMARY KEY,
			wallet_address TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retries INTEGER NOT NULL DEFAULT 3,
			next_retry_at INTEGER,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (wallet_address) REFERENCES wallets(address)
		);
		CREATE INDEX IF NOT EXISTS idx_notifications_pending
			ON notifications_queue(wallet_address, status, next_retry_at);
	`)
	if err != nil {
		return fmt.Errorf("sqlite: create notifications_queue: %w", err)
	}
	return nil
}

// --- SentAlertStore ---

// WasAlertSent проверяет, было ли уведомление отправлено в течение cooldown.
func (d *DB) WasAlertSent(ctx context.Context, alertType, conditionID string, cooldown time.Duration) (bool, error) {
	cutoff := time.Now().Add(-cooldown).Unix()
	var sentAt int64
	err := d.db.QueryRowContext(ctx,
		`SELECT sent_at FROM sent_alerts WHERE alert_type = ? AND condition_id = ? AND sent_at > ?`,
		alertType, conditionID, cutoff,
	).Scan(&sentAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkAlertSent записывает (или обновляет) время последней отправки уведомления.
func (d *DB) MarkAlertSent(ctx context.Context, alertType, conditionID string) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sent_alerts (alert_type, condition_id, sent_at) VALUES (?, ?, ?)`,
		alertType, conditionID, time.Now().Unix(),
	)
	return err
}

// --- MarketAlertStore ---

func (d *DB) SaveAlert(ctx context.Context, a *storage.MarketAlertRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO market_alerts (id, condition_id, token_id, direction, threshold, created_at, triggered)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ConditionID, a.TokenID, a.Direction, a.Threshold,
		a.CreatedAt.UTC().Format(time.RFC3339),
		boolToInt(a.Triggered),
	)
	return err
}

func (d *DB) DeleteAlert(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM market_alerts WHERE id = ?`, id)
	return err
}

func (d *DB) GetAlerts(ctx context.Context) ([]*storage.MarketAlertRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, condition_id, token_id, direction, threshold, created_at, triggered FROM market_alerts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.MarketAlertRecord
	for rows.Next() {
		var a storage.MarketAlertRecord
		var ca string
		var tr int
		if err := rows.Scan(&a.ID, &a.ConditionID, &a.TokenID, &a.Direction, &a.Threshold, &ca, &tr); err != nil {
			return nil, err
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		a.Triggered = tr != 0
		result = append(result, &a)
	}
	return result, rows.Err()
}

func (d *DB) MarkAlertTriggered(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, `UPDATE market_alerts SET triggered = 1 WHERE id = ?`, id)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- Old TradeStore and OrderStore methods (deprecated, for backward compatibility) ---

// SaveTrade saves a trade to the old 'trades' table (deprecated, use InsertTrade instead).
func (d *DB) SaveTrade(ctx context.Context, t *storage.TradeRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO trades (id, trade_id, order_id, asset_id, condition_id, side, price, size, fee, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.TradeID, t.OrderID, t.AssetID, t.ConditionID, t.Side,
		t.Price, t.Size, t.Fee, t.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

// SaveOrder saves an order to the old 'orders' table (deprecated).
func (d *DB) SaveOrder(ctx context.Context, o *storage.OrderRecord) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO orders (id, asset_id, condition_id, side, order_type, price, size, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.AssetID, o.ConditionID, o.Side, o.OrderType, o.Price, o.Size, o.Status,
		o.CreatedAt.UTC().Format(time.RFC3339), o.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// UpdateOrderStatus updates an order status in the old 'orders' table (deprecated).
func (d *DB) UpdateOrderStatus(ctx context.Context, id, status string) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
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

// --- WalletStatsStore ---


// --- MarketCacheStore ---

// UpsertMarkets вставляет или обновляет кеш маркетов.
// first_seen не перезаписывается если запись уже существует.
func (d *DB) UpsertMarkets(ctx context.Context, records []storage.MarketCacheRecord) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO markets_cache (condition_id, data, updated_at, first_seen)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(condition_id) DO UPDATE SET
			data       = excluded.data,
			updated_at = excluded.updated_at,
			first_seen = first_seen
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range records {
		if _, err := stmt.ExecContext(ctx,
			r.ConditionID, r.Data,
			r.UpdatedAt.Unix(), r.FirstSeen.Unix(),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetCachedMarkets возвращает все маркеты из кеша.
func (d *DB) GetCachedMarkets(ctx context.Context) ([]storage.MarketCacheRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT condition_id, data, updated_at, first_seen FROM markets_cache`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMarketCacheRows(rows)
}

// GetNewMarkets возвращает маркеты с first_seen >= since.
func (d *DB) GetNewMarkets(ctx context.Context, since time.Time) ([]storage.MarketCacheRecord, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT condition_id, data, updated_at, first_seen FROM markets_cache WHERE first_seen >= ?`,
		since.Unix(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMarketCacheRows(rows)
}

func scanMarketCacheRows(rows *sql.Rows) ([]storage.MarketCacheRecord, error) {
	var result []storage.MarketCacheRecord
	for rows.Next() {
		var r storage.MarketCacheRecord
		var ua, fs int64
		if err := rows.Scan(&r.ConditionID, &r.Data, &ua, &fs); err != nil {
			return nil, err
		}
		r.UpdatedAt = time.Unix(ua, 0).UTC()
		r.FirstSeen = time.Unix(fs, 0).UTC()
		result = append(result, r)
	}
	return result, rows.Err()
}


// --- NotificationQueueStore ---

// EnqueueNotification добавляет уведомление в очередь.
func (d *DB) EnqueueNotification(ctx context.Context, notif *storage.Notification) error {
	var nextRetryAtVal *int64
	if notif.NextRetryAt != nil {
		v := notif.NextRetryAt.Unix()
		nextRetryAtVal = &v
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO notifications_queue
		 (id, wallet_address, event_type, payload, status, retry_count, max_retries, next_retry_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		notif.ID, notif.WalletAddress, notif.EventType, notif.Payload, notif.Status, notif.RetryCount, notif.MaxRetries,
		nextRetryAtVal, notif.CreatedAt.UTC().Format(time.RFC3339), notif.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// GetPendingNotifications получает ожидающие уведомления.
func (d *DB) GetPendingNotifications(ctx context.Context, walletAddress string) ([]*storage.Notification, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, wallet_address, event_type, payload, status, retry_count, max_retries, next_retry_at, created_at, updated_at
		 FROM notifications_queue WHERE wallet_address = ? AND status IN ('PENDING', 'FAILED')`,
		walletAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.Notification
	for rows.Next() {
		var notif storage.Notification
		var nextRetryAtVal *int64
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&notif.ID, &notif.WalletAddress, &notif.EventType, &notif.Payload, &notif.Status, &notif.RetryCount,
			&notif.MaxRetries, &nextRetryAtVal, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		notif.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		notif.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if nextRetryAtVal != nil {
			t := time.Unix(*nextRetryAtVal, 0)
			notif.NextRetryAt = &t
		}
		result = append(result, &notif)
	}
	return result, rows.Err()
}

// UpdateNotificationStatus обновляет статус уведомления.
func (d *DB) UpdateNotificationStatus(ctx context.Context, id, status string, retryCount int, nextRetryAt *time.Time) error {
	var nextRetryAtVal *int64
	if nextRetryAt != nil {
		v := nextRetryAt.Unix()
		nextRetryAtVal = &v
	}
	_, err := d.db.ExecContext(ctx,
		`UPDATE notifications_queue SET status = ?, retry_count = ?, next_retry_at = ?, updated_at = ? WHERE id = ?`,
		status, retryCount, nextRetryAtVal, time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

// DeleteNotification удаляет уведомление из очереди.
func (d *DB) DeleteNotification(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM notifications_queue WHERE id = ?`, id)
	return err
}

// --- WalletStatisticsStore ---

// SaveWalletStats сохраняет статистику кошелька.
func (d *DB) SaveWalletStats(ctx context.Context, stats *storage.WalletStats) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO wallet_statistics
		 (wallet_address, fetched_at, balance_usd, pnl_usd, win_rate, total_trades, total_volume)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		stats.WalletAddress, stats.FetchedAt.Unix(), stats.BalanceUSD, stats.PnLUSD,
		stats.WinRate, stats.TotalTrades, stats.TotalVolume,
	)
	return err
}

// GetWalletStats получает статистику кошелька.
func (d *DB) GetWalletStats(ctx context.Context, walletAddress string, limit int) ([]*storage.WalletStats, error) {
	q := `SELECT wallet_address, fetched_at, balance_usd, pnl_usd, win_rate, total_trades, total_volume
	      FROM wallet_statistics WHERE wallet_address = ? ORDER BY fetched_at DESC`
	args := []any{walletAddress}
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.WalletStats
	for rows.Next() {
		var stats storage.WalletStats
		var fetchedAtVal int64
		if err := rows.Scan(&stats.WalletAddress, &fetchedAtVal, &stats.BalanceUSD, &stats.PnLUSD,
			&stats.WinRate, &stats.TotalTrades, &stats.TotalVolume); err != nil {
			return nil, err
		}
		stats.FetchedAt = time.Unix(fetchedAtVal, 0)
		result = append(result, &stats)
	}
	return result, rows.Err()
}

// Убедимся, что DB реализует storage.Store
var _ storage.Store = (*DB)(nil)
