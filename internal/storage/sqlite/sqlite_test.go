package sqlite_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/storage"
	"github.com/atlasdev/orbitron/internal/storage/sqlite"
)

func TestOpen(t *testing.T) {
	f, err := os.CreateTemp("", "polytrade-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	db, err := sqlite.Open(f.Name())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
}

func TestCopyTradeSaveAndGet(t *testing.T) {
	f, err := os.CreateTemp("", "polytrade-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	db, err := sqlite.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	rec := &storage.CopyTradeRecord{
		ID:            "test-id-1",
		TraderAddress: "0xABC",
		AssetID:       "token123",
		ConditionID:   "cond456",
		Side:          "BUY",
		Size:          10.0,
		Price:         0.65,
		OurOrderID:    "order-xyz",
		Status:        "open",
		OpenedAt:      now,
	}

	if err := db.SaveCopyTrade(ctx, rec); err != nil {
		t.Fatalf("SaveCopyTrade: %v", err)
	}

	trades, err := db.GetOpenCopyTrades(ctx, "0xABC")
	if err != nil {
		t.Fatalf("GetOpenCopyTrades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].AssetID != "token123" {
		t.Errorf("expected asset token123, got %s", trades[0].AssetID)
	}

	allOpen, err := db.GetAllOpenCopyTrades(ctx)
	if err != nil {
		t.Fatalf("GetAllOpenCopyTrades: %v", err)
	}
	if len(allOpen) != 1 {
		t.Fatalf("expected 1 in GetAllOpenCopyTrades, got %d", len(allOpen))
	}
}

func TestCopyTradeUpdateStatus(t *testing.T) {
	f, err := os.CreateTemp("", "polytrade-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	db, err := sqlite.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	rec := &storage.CopyTradeRecord{
		ID:            "test-id-2",
		TraderAddress: "0xDEF",
		AssetID:       "token789",
		ConditionID:   "cond111",
		Side:          "BUY",
		Size:          5.0,
		Price:         0.72,
		OurOrderID:    "order-abc",
		Status:        "open",
		OpenedAt:      now,
	}

	if err := db.SaveCopyTrade(ctx, rec); err != nil {
		t.Fatal(err)
	}

	closedAt := time.Now().UTC().Truncate(time.Second)
	pnl := 2.5
	if err := db.UpdateCopyTrade(ctx, "test-id-2", "closed", &closedAt, &pnl); err != nil {
		t.Fatalf("UpdateCopyTrade: %v", err)
	}

	trades, err := db.GetOpenCopyTrades(ctx, "0xDEF")
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 0 {
		t.Errorf("expected 0 open trades after close, got %d", len(trades))
	}
}

func TestWalletStats(t *testing.T) {
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Save a stats record
	err = db.SaveWalletStats(ctx, "wallet-1", 1234.56, 88.0)
	if err != nil {
		t.Fatalf("SaveWalletStats: %v", err)
	}

	// Retrieve it
	rows, err := db.GetWalletStats(ctx, "wallet-1", 10)
	if err != nil {
		t.Fatalf("GetWalletStats: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].BalanceUSD != 1234.56 {
		t.Errorf("BalanceUSD = %f, want 1234.56", rows[0].BalanceUSD)
	}
	if rows[0].PnLUSD != 88.0 {
		t.Errorf("PnLUSD = %f, want 88.0", rows[0].PnLUSD)
	}
	if rows[0].WalletID != "wallet-1" {
		t.Errorf("WalletID = %q, want wallet-1", rows[0].WalletID)
	}
	if rows[0].FetchedAt.IsZero() {
		t.Error("FetchedAt must not be zero")
	}

	// Different wallet should not appear
	rows2, err := db.GetWalletStats(ctx, "wallet-2", 10)
	if err != nil {
		t.Fatalf("GetWalletStats wallet-2: %v", err)
	}
	if len(rows2) != 0 {
		t.Errorf("expected 0 rows for wallet-2, got %d", len(rows2))
	}
}
