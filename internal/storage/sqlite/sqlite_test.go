package sqlite_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/storage"
	"github.com/atlasdev/polytrade-bot/internal/storage/sqlite"
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
