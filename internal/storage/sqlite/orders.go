package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/atlasdev/orbitron/internal/storage"
)

// --- OrderHistoryStore implementation ---

// InsertOrder вставляет новый ордер в базу.
func (d *DB) InsertOrder(ctx context.Context, order *storage.Order) error {
	expiresAtStr := ""
	if order.ExpiresAt != nil {
		expiresAtStr = order.ExpiresAt.UTC().Format(time.RFC3339)
	}
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO orders_history
		 (id, wallet_address, condition_id, asset_id, side, order_type, price, size, status, expires_at, created_at, updated_at, synced_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.ID, order.WalletAddress, order.ConditionID, order.AssetID, order.Side, order.OrderType,
		order.Price, order.Size, order.Status, expiresAtStr,
		order.CreatedAt.UTC().Format(time.RFC3339),
		order.UpdatedAt.UTC().Format(time.RFC3339),
		nil,
	)
	return err
}

// UpdateOrder обновляет существующий ордер.
func (d *DB) UpdateOrder(ctx context.Context, order *storage.Order) error {
	expiresAtStr := ""
	if order.ExpiresAt != nil {
		expiresAtStr = order.ExpiresAt.UTC().Format(time.RFC3339)
	}
	var syncedAtStr *string
	if order.SyncedAt != nil {
		s := order.SyncedAt.UTC().Format(time.RFC3339)
		syncedAtStr = &s
	}
	_, err := d.db.ExecContext(ctx,
		`UPDATE orders_history SET status = ?, expires_at = ?, updated_at = ?, synced_at = ? WHERE id = ?`,
		order.Status, expiresAtStr,
		order.UpdatedAt.UTC().Format(time.RFC3339), syncedAtStr, order.ID,
	)
	return err
}

// GetOrder получает ордер по ID.
func (d *DB) GetOrder(ctx context.Context, id string) (*storage.Order, error) {
	var order storage.Order
	var expiresAtStr *string
	var syncedAtStr *string
	var createdAtStr, updatedAtStr string
	err := d.db.QueryRowContext(ctx,
		`SELECT id, wallet_address, condition_id, asset_id, side, order_type, price, size, status, expires_at, created_at, updated_at, synced_at
		 FROM orders_history WHERE id = ?`,
		id,
	).Scan(&order.ID, &order.WalletAddress, &order.ConditionID, &order.AssetID, &order.Side, &order.OrderType,
		&order.Price, &order.Size, &order.Status, &expiresAtStr, &createdAtStr, &updatedAtStr, &syncedAtStr)
	if err != nil {
		return nil, err
	}
	order.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	order.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	if expiresAtStr != nil {
		t, _ := time.Parse(time.RFC3339, *expiresAtStr)
		order.ExpiresAt = &t
	}
	if syncedAtStr != nil {
		t, _ := time.Parse(time.RFC3339, *syncedAtStr)
		order.SyncedAt = &t
	}
	return &order, nil
}

// GetOrders получает ордеры по фильтрам.
func (d *DB) GetOrders(ctx context.Context, filters storage.OrderFilters) ([]*storage.Order, error) {
	q := `SELECT id, wallet_address, condition_id, asset_id, side, order_type, price, size, status, expires_at, created_at, updated_at, synced_at
	      FROM orders_history WHERE 1=1`
	args := []any{}
	if filters.WalletAddress != "" {
		q += " AND wallet_address = ?"
		args = append(args, filters.WalletAddress)
	}
	if filters.ConditionID != "" {
		q += " AND condition_id = ?"
		args = append(args, filters.ConditionID)
	}
	if filters.Status != "" {
		q += " AND status = ?"
		args = append(args, filters.Status)
	}
	if filters.Side != "" {
		q += " AND side = ?"
		args = append(args, filters.Side)
	}
	if !filters.From.IsZero() {
		q += " AND created_at >= ?"
		args = append(args, filters.From.UTC().Format(time.RFC3339))
	}
	if !filters.To.IsZero() {
		q += " AND created_at <= ?"
		args = append(args, filters.To.UTC().Format(time.RFC3339))
	}
	q += " ORDER BY created_at DESC"
	if filters.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", filters.Limit)
	}
	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.Order
	for rows.Next() {
		var order storage.Order
		var expiresAtStr *string
		var syncedAtStr *string
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&order.ID, &order.WalletAddress, &order.ConditionID, &order.AssetID, &order.Side, &order.OrderType,
			&order.Price, &order.Size, &order.Status, &expiresAtStr, &createdAtStr, &updatedAtStr, &syncedAtStr); err != nil {
			return nil, err
		}
		order.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		order.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if expiresAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *expiresAtStr)
			order.ExpiresAt = &t
		}
		if syncedAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *syncedAtStr)
			order.SyncedAt = &t
		}
		result = append(result, &order)
	}
	return result, rows.Err()
}

// GetExpiredOrders получает истекшие GTD ордеры.
func (d *DB) GetExpiredOrders(ctx context.Context, before time.Time) ([]*storage.Order, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, wallet_address, condition_id, asset_id, side, order_type, price, size, status, expires_at, created_at, updated_at, synced_at
		 FROM orders_history WHERE expires_at IS NOT NULL AND expires_at < ? AND status IN ('PENDING', 'OPEN')`,
		before.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.Order
	for rows.Next() {
		var order storage.Order
		var expiresAtStr *string
		var syncedAtStr *string
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&order.ID, &order.WalletAddress, &order.ConditionID, &order.AssetID, &order.Side, &order.OrderType,
			&order.Price, &order.Size, &order.Status, &expiresAtStr, &createdAtStr, &updatedAtStr, &syncedAtStr); err != nil {
			return nil, err
		}
		order.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		order.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if expiresAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *expiresAtStr)
			order.ExpiresAt = &t
		}
		if syncedAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *syncedAtStr)
			order.SyncedAt = &t
		}
		result = append(result, &order)
	}
	return result, rows.Err()
}

// --- TradeHistoryStore implementation ---

// InsertTrade вставляет новую сделку.
func (d *DB) InsertTrade(ctx context.Context, trade *storage.Trade) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO trades_history
		 (id, wallet_address, order_id, trade_id, condition_id, asset_id, side, price, size, fee, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trade.ID, trade.WalletAddress, trade.OrderID, trade.TradeID, trade.ConditionID, trade.AssetID,
		trade.Side, trade.Price, trade.Size, trade.Fee, trade.Timestamp.UTC().Format(time.RFC3339),
	)
	return err
}

// GetTrades получает сделки по кошельку и дате.
func (d *DB) GetTrades(ctx context.Context, walletAddress string, from, to time.Time) ([]*storage.Trade, error) {
	q := `SELECT id, wallet_address, order_id, trade_id, condition_id, asset_id, side, price, size, fee, timestamp
	      FROM trades_history WHERE wallet_address = ?`
	args := []any{walletAddress}
	if !from.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, from.UTC().Format(time.RFC3339))
	}
	if !to.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, to.UTC().Format(time.RFC3339))
	}
	q += " ORDER BY timestamp DESC"
	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*storage.Trade
	for rows.Next() {
		var trade storage.Trade
		var timestampStr string
		if err := rows.Scan(&trade.ID, &trade.WalletAddress, &trade.OrderID, &trade.TradeID, &trade.ConditionID, &trade.AssetID,
			&trade.Side, &trade.Price, &trade.Size, &trade.Fee, &timestampStr); err != nil {
			return nil, err
		}
		trade.Timestamp, _ = time.Parse(time.RFC3339, timestampStr)
		result = append(result, &trade)
	}
	return result, rows.Err()
}

// --- Wallet statistics helpers ---

// GetWalletOrdersByCondition gets all orders for a wallet by condition.
func (d *DB) GetWalletOrdersByCondition(ctx context.Context, walletAddress, conditionID string) ([]*storage.Order, error) {
	filters := storage.OrderFilters{
		WalletAddress: walletAddress,
		ConditionID:   conditionID,
	}
	return d.GetOrders(ctx, filters)
}

// GetWalletTradesByCondition gets all trades for a wallet by condition.
func (d *DB) GetWalletTradesByCondition(ctx context.Context, walletAddress, conditionID string) ([]*storage.Trade, error) {
	q := `SELECT id, wallet_address, order_id, trade_id, condition_id, asset_id, side, price, size, fee, timestamp
	      FROM trades_history WHERE wallet_address = ? AND condition_id = ?
	      ORDER BY timestamp DESC`
	rows, err := d.db.QueryContext(ctx, q, walletAddress, conditionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*storage.Trade
	for rows.Next() {
		var trade storage.Trade
		var timestampStr string
		if err := rows.Scan(&trade.ID, &trade.WalletAddress, &trade.OrderID, &trade.TradeID, &trade.ConditionID, &trade.AssetID,
			&trade.Side, &trade.Price, &trade.Size, &trade.Fee, &timestampStr); err != nil {
			return nil, err
		}
		trade.Timestamp, _ = time.Parse(time.RFC3339, timestampStr)
		result = append(result, &trade)
	}
	return result, rows.Err()
}
