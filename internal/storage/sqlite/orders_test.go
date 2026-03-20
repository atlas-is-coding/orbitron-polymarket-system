package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/storage"
)

// TestOrderStore_InsertOrder verifies that InsertOrder correctly stores an order in the database.
func TestOrderStore_InsertOrder(t *testing.T) {
	// Create in-memory database
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a test order
	testOrder := &storage.Order{
		ID:            "order-123",
		WalletAddress: "0xabcd1234",
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		OrderType:     "limit",
		Price:         0.65,
		Size:          100.0,
		Status:        "PENDING",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Insert the order
	err = db.InsertOrder(ctx, testOrder)
	if err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	// Retrieve the order
	retrieved, err := db.GetOrder(ctx, testOrder.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	// Verify the order was stored correctly
	if retrieved == nil {
		t.Fatal("retrieved order is nil")
	}
	if retrieved.ID != testOrder.ID {
		t.Errorf("expected ID %s, got %s", testOrder.ID, retrieved.ID)
	}
	if retrieved.WalletAddress != testOrder.WalletAddress {
		t.Errorf("expected wallet %s, got %s", testOrder.WalletAddress, retrieved.WalletAddress)
	}
	if retrieved.Side != testOrder.Side {
		t.Errorf("expected side %s, got %s", testOrder.Side, retrieved.Side)
	}
	if retrieved.Status != testOrder.Status {
		t.Errorf("expected status %s, got %s", testOrder.Status, retrieved.Status)
	}
	if retrieved.Price != testOrder.Price {
		t.Errorf("expected price %f, got %f", testOrder.Price, retrieved.Price)
	}
}

// TestOrderStore_GetOrders verifies that GetOrders returns orders matching the filters.
func TestOrderStore_GetOrders(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	wallet := "0xabcd1234"

	// Insert multiple orders
	orders := []*storage.Order{
		{
			ID:            "order-1",
			WalletAddress: wallet,
			ConditionID:   "0x111111",
			AssetID:       "0x222222",
			Side:          "BUY",
			OrderType:     "limit",
			Price:         0.50,
			Size:          50.0,
			Status:        "PENDING",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            "order-2",
			WalletAddress: wallet,
			ConditionID:   "0x111111",
			AssetID:       "0x222222",
			Side:          "SELL",
			OrderType:     "limit",
			Price:         0.75,
			Size:          100.0,
			Status:        "OPEN",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            "order-3",
			WalletAddress: "0xdifferent",
			ConditionID:   "0x333333",
			AssetID:       "0x444444",
			Side:          "BUY",
			OrderType:     "limit",
			Price:         0.60,
			Size:          75.0,
			Status:        "PENDING",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
	}

	for _, order := range orders {
		if err := db.InsertOrder(ctx, order); err != nil {
			t.Fatalf("InsertOrder failed for %s: %v", order.ID, err)
		}
	}

	// Test: Get orders for specific wallet
	filters := storage.OrderFilters{
		WalletAddress: wallet,
	}
	retrieved, err := db.GetOrders(ctx, filters)
	if err != nil {
		t.Fatalf("GetOrders failed: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("expected 2 orders for wallet, got %d", len(retrieved))
	}

	// Test: Get orders with status filter
	filters.Status = "PENDING"
	retrieved, err = db.GetOrders(ctx, filters)
	if err != nil {
		t.Fatalf("GetOrders with status filter failed: %v", err)
	}
	if len(retrieved) != 1 {
		t.Errorf("expected 1 PENDING order, got %d", len(retrieved))
	}
	if retrieved[0].ID != "order-1" {
		t.Errorf("expected order-1, got %s", retrieved[0].ID)
	}
}

// TestOrderStore_UpdateOrder verifies that UpdateOrder correctly updates order status.
func TestOrderStore_UpdateOrder(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create and insert initial order
	order := &storage.Order{
		ID:            "order-update-test",
		WalletAddress: "0xabcd1234",
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		OrderType:     "limit",
		Price:         0.65,
		Size:          100.0,
		Status:        "PENDING",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := db.InsertOrder(ctx, order); err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	// Update the order status
	order.Status = "OPEN"
	order.UpdatedAt = time.Now().UTC()
	if err := db.UpdateOrder(ctx, order); err != nil {
		t.Fatalf("UpdateOrder failed: %v", err)
	}

	// Verify the update
	retrieved, err := db.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if retrieved.Status != "OPEN" {
		t.Errorf("expected status OPEN, got %s", retrieved.Status)
	}
}

// TestOrderStore_InsertTrade verifies that InsertTrade correctly stores a trade in the database.
func TestOrderStore_InsertTrade(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// First insert an order (trades have foreign key to orders)
	order := &storage.Order{
		ID:            "order-for-trade",
		WalletAddress: "0xabcd1234",
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		OrderType:     "limit",
		Price:         0.65,
		Size:          100.0,
		Status:        "OPEN",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := db.InsertOrder(ctx, order); err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	// Create a test trade
	trade := &storage.Trade{
		ID:            "trade-123",
		WalletAddress: "0xabcd1234",
		OrderID:       "order-for-trade",
		TradeID:       "clob-trade-123",
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		Price:         0.65,
		Size:          50.0,
		Fee:           0.325,
		Timestamp:     time.Now().UTC(),
	}

	// Insert the trade
	err = db.InsertTrade(ctx, trade)
	if err != nil {
		t.Fatalf("InsertTrade failed: %v", err)
	}

	// Retrieve trades for the wallet
	trades, err := db.GetTrades(ctx, "0xabcd1234", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetTrades failed: %v", err)
	}

	// Alternative: use FilteredTrades helper
	_ = trades

	// Verify the trade was stored
	if len(trades) == 0 {
		t.Fatal("no trades retrieved")
	}
	retrieved := trades[0]
	if retrieved.ID != trade.ID {
		t.Errorf("expected trade ID %s, got %s", trade.ID, retrieved.ID)
	}
	if retrieved.Price != trade.Price {
		t.Errorf("expected price %f, got %f", trade.Price, retrieved.Price)
	}
}

// TestOrderStore_GetTrades verifies that GetTrades returns trades for a wallet within time range.
func TestOrderStore_GetTrades(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	wallet := "0xabcd1234"

	// Insert an order first
	order := &storage.Order{
		ID:            "order-trades-test",
		WalletAddress: wallet,
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		OrderType:     "limit",
		Price:         0.65,
		Size:          100.0,
		Status:        "FILLED",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := db.InsertOrder(ctx, order); err != nil {
		t.Fatalf("InsertOrder failed: %v", err)
	}

	// Insert multiple trades
	now := time.Now().UTC()
	trades := []*storage.Trade{
		{
			ID:            "trade-1",
			WalletAddress: wallet,
			OrderID:       "order-trades-test",
			TradeID:       "clob-trade-1",
			ConditionID:   "0x123456789",
			AssetID:       "0x0987654321",
			Side:          "BUY",
			Price:         0.65,
			Size:          50.0,
			Fee:           0.325,
			Timestamp:     now.Add(-1 * time.Hour),
		},
		{
			ID:            "trade-2",
			WalletAddress: wallet,
			OrderID:       "order-trades-test",
			TradeID:       "clob-trade-2",
			ConditionID:   "0x123456789",
			AssetID:       "0x0987654321",
			Side:          "BUY",
			Price:         0.65,
			Size:          50.0,
			Fee:           0.325,
			Timestamp:     now,
		},
	}

	for _, trade := range trades {
		if err := db.InsertTrade(ctx, trade); err != nil {
			t.Fatalf("InsertTrade failed for %s: %v", trade.ID, err)
		}
	}

	// Test: Get all trades
	retrieved, err := db.GetTrades(ctx, wallet, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetTrades failed: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("expected 2 trades, got %d", len(retrieved))
	}

	// Test: Get trades from last 30 minutes
	from := now.Add(-30 * time.Minute)
	retrieved, err = db.GetTrades(ctx, wallet, from, time.Time{})
	if err != nil {
		t.Fatalf("GetTrades with time filter failed: %v", err)
	}
	if len(retrieved) != 1 {
		t.Errorf("expected 1 trade in last 30 min, got %d", len(retrieved))
	}
}

// BenchmarkInsertOrder benchmarks the InsertOrder operation.
func BenchmarkInsertOrder(b *testing.B) {
	db, err := Open(":memory:")
	if err != nil {
		b.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	order := &storage.Order{
		ID:            "order-bench",
		WalletAddress: "0xabcd1234",
		ConditionID:   "0x123456789",
		AssetID:       "0x0987654321",
		Side:          "BUY",
		OrderType:     "limit",
		Price:         0.65,
		Size:          100.0,
		Status:        "PENDING",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order.ID = "order-" + string(rune(i))
		db.InsertOrder(ctx, order)
	}
}

// TestOrderStore_GetExpiredOrders verifies that expired GTD orders are retrieved.
func TestOrderStore_GetExpiredOrders(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	now := time.Now().UTC()
	expired := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Insert orders with different expiration times
	orders := []*storage.Order{
		{
			ID:            "order-expired",
			WalletAddress: "0xabcd1234",
			ConditionID:   "0x111111",
			AssetID:       "0x222222",
			Side:          "BUY",
			OrderType:     "limit",
			Price:         0.50,
			Size:          50.0,
			Status:        "PENDING",
			ExpiresAt:     &expired,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "order-valid",
			WalletAddress: "0xabcd1234",
			ConditionID:   "0x111111",
			AssetID:       "0x222222",
			Side:          "BUY",
			OrderType:     "limit",
			Price:         0.50,
			Size:          50.0,
			Status:        "PENDING",
			ExpiresAt:     &future,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}

	for _, order := range orders {
		if err := db.InsertOrder(ctx, order); err != nil {
			t.Fatalf("InsertOrder failed: %v", err)
		}
	}

	// Get expired orders
	expiredOrders, err := db.GetExpiredOrders(ctx, now)
	if err != nil {
		t.Fatalf("GetExpiredOrders failed: %v", err)
	}

	if len(expiredOrders) != 1 {
		t.Errorf("expected 1 expired order, got %d", len(expiredOrders))
	}
	if expiredOrders[0].ID != "order-expired" {
		t.Errorf("expected order-expired, got %s", expiredOrders[0].ID)
	}
}
