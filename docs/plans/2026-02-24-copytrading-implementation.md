# Copytrading Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a fully functional copytrading subsystem that monitors selected Polymarket traders via the public Data API, replicates their positions with configurable size scaling, auto-closes when the trader exits, and persists history to SQLite.

**Architecture:** New `internal/copytrading/` package with `CopyTrader` orchestrator, per-trader `TraderTracker` goroutines, `SizeCalculator`, `OrderExecutor`. Separate `internal/auth/order_signer.go` for EIP-712 order signing. Full SQLite implementation replacing the current stub. Hot-reload of trader list via `fsnotify`.

**Tech Stack:** `modernc.org/sqlite` (pure-Go SQLite, no CGO), `github.com/fsnotify/fsnotify` (file watching), `github.com/google/uuid` (ID generation), `github.com/ethereum/go-ethereum` (EIP-712, already in go.mod)

---

### Task 1: Add dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add packages**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go get modernc.org/sqlite
go get github.com/fsnotify/fsnotify
go get github.com/google/uuid
go mod tidy
```

**Step 2: Verify build still compiles**

```bash
go build ./...
```

Expected: no errors.

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add sqlite, fsnotify, uuid dependencies"
```

---

### Task 2: Add CopytradingConfig to config.go

**Files:**
- Modify: `internal/config/config.go`
- Modify: `config.toml`

**Step 1: Add config structs to `internal/config/config.go`**

Add to the `Config` struct field (after `Database`):
```go
Copytrading CopytradingConfig `toml:"copytrading"`
```

Add new structs at the bottom of the file (before `Load`):
```go
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
```

Add defaults in `validate()`:
```go
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
```

**Step 2: Add section to `config.toml`**

Append at the end:
```toml
[copytrading]
  enabled          = false
  poll_interval_ms = 10000
  # Метод расчёта размера: "proportional" или "fixed_pct"
  size_mode        = "proportional"

  # Пример трейдера — укажи реальный адрес proxy-wallet
  # [[copytrading.traders]]
  #   address          = "0xABC..."
  #   label            = "whale1"
  #   enabled          = true
  #   allocation_pct   = 10.0
  #   max_position_usd = 50.0
  #   # size_mode = "fixed_pct"   # переопределяет глобальный
```

**Step 3: Verify build**

```bash
go build ./...
```

**Step 4: Commit**

```bash
git add internal/config/config.go config.toml
git commit -m "feat: add CopytradingConfig to config"
```

---

### Task 3: Extend storage interfaces for copy trades

**Files:**
- Modify: `internal/storage/storage.go`

**Step 1: Add CopyTradeRecord model and new interface methods**

Add to `internal/storage/storage.go` after the existing models:
```go
// CopyTradeRecord — запись о скопированной сделке.
type CopyTradeRecord struct {
    ID            string
    TraderAddress string
    AssetID       string
    ConditionID   string
    Side          string    // "BUY" или "SELL"
    Size          float64
    Price         float64
    OurOrderID    string    // ID ордера в CLOB
    Status        string    // "open", "closed", "failed"
    OpenedAt      time.Time
    ClosedAt      *time.Time
    PnL           *float64
}

// CopyTradeStore — хранилище копитрейд сделок.
type CopyTradeStore interface {
    SaveCopyTrade(ctx context.Context, r *CopyTradeRecord) error
    UpdateCopyTrade(ctx context.Context, id, status string, closedAt *time.Time, pnl *float64) error
    GetOpenCopyTrades(ctx context.Context, traderAddress string) ([]*CopyTradeRecord, error)
    GetAllOpenCopyTrades(ctx context.Context) ([]*CopyTradeRecord, error)
}
```

Extend `Store` interface:
```go
// Store — объединённый интерфейс хранилища.
type Store interface {
    TradeStore
    OrderStore
    CopyTradeStore
    Close() error
}
```

**Step 2: Verify build (sqlite stub will fail on interface — expected)**

```bash
go build ./... 2>&1 | head -20
```

Expected: compile error about missing methods in sqlite stub (we fix that in Task 4).

**Step 3: Commit**

```bash
git add internal/storage/storage.go
git commit -m "feat: add CopyTradeStore interface to storage"
```

---

### Task 4: Implement SQLite storage

**Files:**
- Modify: `internal/storage/sqlite/sqlite.go`

**Step 1: Replace the stub with full implementation**

Replace the entire content of `internal/storage/sqlite/sqlite.go`:

```go
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
```

**Step 2: Verify build**

```bash
go build ./...
```

Expected: no errors.

**Step 3: Write a quick smoke test for SQLite**

Create `internal/storage/sqlite/sqlite_test.go`:

```go
package sqlite_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/storage"
	"github.com/atlasdev/polytrade-bot/internal/storage/sqlite"
)

func TestSQLiteOpenAndMigrate(t *testing.T) {
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

func TestSQLiteCopyTrades(t *testing.T) {
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

	closedAt := time.Now().UTC().Truncate(time.Second)
	pnl := 2.5
	if err := db.UpdateCopyTrade(ctx, "test-id-1", "closed", &closedAt, &pnl); err != nil {
		t.Fatalf("UpdateCopyTrade: %v", err)
	}

	trades, err = db.GetOpenCopyTrades(ctx, "0xABC")
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 0 {
		t.Errorf("expected 0 open trades after close, got %d", len(trades))
	}
}
```

**Step 4: Run tests**

```bash
go test ./internal/storage/sqlite/... -v
```

Expected: PASS for both tests.

**Step 5: Commit**

```bash
git add internal/storage/sqlite/sqlite.go internal/storage/sqlite/sqlite_test.go
git commit -m "feat: implement SQLite storage with copy_trades table"
```

---

### Task 5: EIP-712 order signer

**Files:**
- Create: `internal/auth/order_signer.go`

**Context:** Polymarket orders are EIP-712 signed messages. To place an order, you must:
1. Build an order struct with token amounts in base units (6 decimals)
2. Hash it using EIP-712 with the CTF Exchange domain
3. Sign the hash with your Ethereum private key

**CTF Exchange contract addresses (Polygon mainnet, chainId=137):**
- Main: `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E`
- Neg Risk: `0xC5d563A36AE78145C45a50134d48A1215220f80a`

**Step 1: Create `internal/auth/order_signer.go`**

```go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Адреса CTF Exchange контрактов на Polygon (chainId=137)
const (
	CTFExchangeMain    = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	CTFExchangeNegRisk = "0xC5d563A36AE78145C45a50134d48A1215220f80a"
)

// OrderSide — сторона ордера в EIP-712 (0=BUY, 1=SELL)
type OrderSide int

const (
	Buy  OrderSide = 0
	Sell OrderSide = 1
)

// SignatureType — тип подписи (0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE)
type SignatureType int

const (
	EOA        SignatureType = 0
	PolyProxy  SignatureType = 1
	GnosisSafe SignatureType = 2
)

// RawOrder — неподписанный ордер для хеширования.
type RawOrder struct {
	Salt          *big.Int
	Maker         common.Address
	Signer        common.Address
	Taker         common.Address
	TokenID       *big.Int
	MakerAmount   *big.Int
	TakerAmount   *big.Int
	Expiration    *big.Int
	Nonce         *big.Int
	FeeRateBps    *big.Int
	Side          OrderSide
	SignatureType SignatureType
}

// ORDER_TYPEHASH — keccak256 хеш типа ордера EIP-712.
var ORDER_TYPEHASH = crypto.Keccak256Hash([]byte(
	"Order(uint256 salt,address maker,address signer,address taker,uint256 tokenId," +
		"uint256 makerAmount,uint256 takerAmount,uint256 expiration,uint256 nonce," +
		"uint256 feeRateBps,uint8 side,uint8 signatureType)",
))

// OrderSigner подписывает ордера приватным ключом Ethereum (EIP-712).
type OrderSigner struct {
	l1           *L1Signer
	exchangeAddr common.Address
	chainID      *big.Int
	domainSep    [32]byte
}

// NewOrderSigner создаёт OrderSigner для указанного exchange контракта.
// negRisk=true — использует NegRisk Exchange адрес.
func NewOrderSigner(l1 *L1Signer, chainID int64, negRisk bool) *OrderSigner {
	addr := CTFExchangeMain
	if negRisk {
		addr = CTFExchangeNegRisk
	}
	exchangeAddr := common.HexToAddress(addr)

	// EIP-712 domain separator
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	domainTypeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))
	nameHash := crypto.Keccak256Hash([]byte("Polymarket CTF Exchange"))
	versionHash := crypto.Keccak256Hash([]byte("1"))

	chainIDBig := big.NewInt(chainID)
	// Encode domain fields
	chainIDPadded := abi.U256(chainIDBig)
	addrPadded := make([]byte, 32)
	copy(addrPadded[12:], exchangeAddr.Bytes())

	domainSepBytes := crypto.Keccak256(
		domainTypeHash.Bytes(),
		nameHash.Bytes(),
		versionHash.Bytes(),
		chainIDPadded,
		addrPadded,
	)
	var domainSep [32]byte
	copy(domainSep[:], domainSepBytes)

	return &OrderSigner{
		l1:           l1,
		exchangeAddr: exchangeAddr,
		chainID:      chainIDBig,
		domainSep:    domainSep,
	}
}

// Sign вычисляет EIP-712 хеш ордера и подписывает его.
// Возвращает hex-строку подписи (без 0x).
func (s *OrderSigner) Sign(order *RawOrder) (string, error) {
	orderHash := s.hashOrder(order)

	// EIP-712 финальный хеш: keccak256("\x19\x01" + domainSeparator + structHash)
	finalHash := crypto.Keccak256(
		[]byte("\x19\x01"),
		s.domainSep[:],
		orderHash[:],
	)

	sig, err := crypto.Sign(finalHash, s.l1.privateKey)
	if err != nil {
		return "", fmt.Errorf("order signer: sign: %w", err)
	}
	// go-ethereum возвращает [R || S || V], V ∈ {0,1}; Ethereum ожидает V ∈ {27,28}
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), nil
}

// hashOrder вычисляет keccak256 хеш структуры ордера по EIP-712.
func (s *OrderSigner) hashOrder(o *RawOrder) [32]byte {
	// Каждое uint256/address поле ABI-кодируется как 32 байта
	encoded := make([]byte, 0, 32*13)
	encoded = append(encoded, ORDER_TYPEHASH.Bytes()...)
	encoded = append(encoded, padBigInt(o.Salt)...)
	encoded = append(encoded, padAddress(o.Maker)...)
	encoded = append(encoded, padAddress(o.Signer)...)
	encoded = append(encoded, padAddress(o.Taker)...)
	encoded = append(encoded, padBigInt(o.TokenID)...)
	encoded = append(encoded, padBigInt(o.MakerAmount)...)
	encoded = append(encoded, padBigInt(o.TakerAmount)...)
	encoded = append(encoded, padBigInt(o.Expiration)...)
	encoded = append(encoded, padBigInt(o.Nonce)...)
	encoded = append(encoded, padBigInt(o.FeeRateBps)...)
	encoded = append(encoded, padUint8(uint8(o.Side))...)
	encoded = append(encoded, padUint8(uint8(o.SignatureType))...)

	return crypto.Keccak256Hash(encoded)
}

// RandomSalt генерирует случайный salt для ордера.
func RandomSalt() (*big.Int, error) {
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	return n, nil
}

// --- Вспомогательные функции ABI-кодирования ---

func padBigInt(n *big.Int) []byte {
	if n == nil {
		return make([]byte, 32)
	}
	b := n.Bytes()
	pad := make([]byte, 32)
	copy(pad[32-len(b):], b)
	return pad
}

func padAddress(addr common.Address) []byte {
	pad := make([]byte, 32)
	copy(pad[12:], addr.Bytes())
	return pad
}

func padUint8(v uint8) []byte {
	pad := make([]byte, 32)
	pad[31] = v
	return pad
}
```

**Step 2: Write test for order signing**

Create `internal/auth/order_signer_test.go`:

```go
package auth_test

import (
	"math/big"
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/ethereum/go-ethereum/common"
)

func TestOrderSignerSign(t *testing.T) {
	// Тестовый приватный ключ (не использовать в продакшене!)
	l1, err := auth.NewL1Signer("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatalf("NewL1Signer: %v", err)
	}

	signer := auth.NewOrderSigner(l1, 137, false)

	order := &auth.RawOrder{
		Salt:        big.NewInt(12345),
		Maker:       common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
		Signer:      common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"),
		Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenID:     big.NewInt(1234567890),
		MakerAmount: big.NewInt(650000),  // 0.65 USDC
		TakerAmount: big.NewInt(1000000), // 1.0 share
		Expiration:  big.NewInt(0),
		Nonce:       big.NewInt(0),
		FeeRateBps:  big.NewInt(0),
		Side:        auth.Buy,
		SignatureType: auth.EOA,
	}

	sig, err := signer.Sign(order)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Подпись должна начинаться с 0x и иметь длину 132 символа (65 байт = 130 hex + "0x")
	if len(sig) != 132 {
		t.Errorf("expected sig length 132, got %d: %s", len(sig), sig)
	}
	if sig[:2] != "0x" {
		t.Errorf("expected sig to start with 0x, got: %s", sig[:2])
	}

	t.Logf("signature: %s", sig)
}
```

**Step 3: Run test**

```bash
go test ./internal/auth/... -v -run TestOrderSigner
```

Expected: PASS.

**Step 4: Commit**

```bash
git add internal/auth/order_signer.go internal/auth/order_signer_test.go
git commit -m "feat: add EIP-712 order signer for Polymarket CTF Exchange"
```

---

### Task 6: Update clob.SignedOrder model and add helper method

**Files:**
- Modify: `internal/api/clob/models.go`
- Modify: `internal/api/clob/orders.go`

**Context:** The existing `SignedOrder` struct is missing `salt` and `signer` fields required by the actual Polymarket API. We also need a helper to build a `CreateOrderRequest` from our auth package.

**Step 1: Update `SignedOrder` in `internal/api/clob/models.go`**

Replace the existing `SignedOrder` struct:
```go
// SignedOrder — подписанный EIP-712 ордер для POST /order.
type SignedOrder struct {
	// Salt — случайное число для уникальности ордера
	Salt string `json:"salt"`
	// Maker — адрес создателя (proxy wallet)
	Maker string `json:"maker"`
	// Signer — адрес подписанта (обычно совпадает с Maker)
	Signer string `json:"signer"`
	// Taker — обычно нулевой адрес
	Taker string `json:"taker"`
	// TokenID — token_id токена YES/NO
	TokenID string `json:"tokenId"`
	// MakerAmount — USDC (BUY) или токены (SELL) в base units (6 decimals)
	MakerAmount string `json:"makerAmount"`
	// TakerAmount — токены (BUY) или USDC (SELL) в base units (6 decimals)
	TakerAmount string `json:"takerAmount"`
	// Expiration — unix timestamp (0 для GTC)
	Expiration string `json:"expiration"`
	// Nonce
	Nonce string `json:"nonce"`
	// FeeRateBps — сбор в базисных пунктах
	FeeRateBps string `json:"feeRateBps"`
	// Side — 0=BUY, 1=SELL
	Side int `json:"side"`
	// SignatureType — 0=EOA, 1=POLY_PROXY, 2=POLY_GNOSIS_SAFE
	SignatureType int `json:"signatureType"`
	// Signature — EIP-712 подпись
	Signature string `json:"signature"`
}
```

Also update `CreateOrderRequest` to include `Owner` as the API key:
```go
// CreateOrderRequest — тело запроса POST /order
type CreateOrderRequest struct {
	// Order — подписанный EIP-712 ордер
	Order SignedOrder `json:"order"`
	// Owner — API key (L2 credentials)
	Owner string `json:"owner"`
	// OrderType — тип ордера
	OrderType OrderType `json:"orderType"`
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add internal/api/clob/models.go
git commit -m "fix: update SignedOrder struct to include salt and signer fields"
```

---

### Task 7: Create copytrading models

**Files:**
- Create: `internal/copytrading/models.go`

**Step 1: Create `internal/copytrading/models.go`**

```go
// Package copytrading реализует автоматическое копирование сделок трейдеров Polymarket.
package copytrading

import (
	"github.com/atlasdev/polytrade-bot/internal/api/data"
)

// TraderState — снимок позиций трейдера (map[assetID]Position).
type TraderState map[string]data.Position

// PositionDiff — результат сравнения двух снимков.
type PositionDiff struct {
	// Opened — новые позиции (появились в текущем снимке)
	Opened []data.Position
	// Closed — закрытые позиции (исчезли из текущего снимка)
	Closed []data.Position
}

// diffStates сравнивает два снимка и возвращает изменения.
func diffStates(prev, curr TraderState) PositionDiff {
	var diff PositionDiff

	for assetID, pos := range curr {
		if _, exists := prev[assetID]; !exists {
			diff.Opened = append(diff.Opened, pos)
		}
	}

	for assetID, pos := range prev {
		if _, exists := curr[assetID]; !exists {
			diff.Closed = append(diff.Closed, pos)
		}
	}

	return diff
}

// toTraderState конвертирует список позиций в map по assetID.
func toTraderState(positions []data.Position) TraderState {
	state := make(TraderState, len(positions))
	for _, p := range positions {
		state[p.Asset] = p
	}
	return state
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add internal/copytrading/models.go
git commit -m "feat: add copytrading models"
```

---

### Task 8: SizeCalculator

**Files:**
- Create: `internal/copytrading/sizer.go`
- Create: `internal/copytrading/sizer_test.go`

**Step 1: Write the failing test first**

Create `internal/copytrading/sizer_test.go`:

```go
package copytrading_test

import (
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/copytrading"
)

func TestSizeCalculatorProportional(t *testing.T) {
	calc := copytrading.NewSizeCalculator("proportional", 10.0, 100.0)

	// Трейдер: баланс 1000 USD, позиция 100 USDC (10% баланса)
	// Наш баланс: 500 USD, allocation: 10%
	// Ожидаем: 100/1000 * 500 * 0.10 = 5.0 USD
	size := calc.Calculate(100.0, 1000.0, 500.0)
	if size != 5.0 {
		t.Errorf("proportional: expected 5.0, got %f", size)
	}
}

func TestSizeCalculatorFixedPct(t *testing.T) {
	calc := copytrading.NewSizeCalculator("fixed_pct", 5.0, 100.0)

	// allocation_pct=5%, наш баланс=200 USD → 5% от 200 = 10.0
	// Позиция трейдера игнорируется
	size := calc.Calculate(999.0, 1000.0, 200.0)
	if size != 10.0 {
		t.Errorf("fixed_pct: expected 10.0, got %f", size)
	}
}

func TestSizeCalculatorMaxCap(t *testing.T) {
	calc := copytrading.NewSizeCalculator("fixed_pct", 50.0, 30.0)

	// 50% от 100 = 50, но max_position_usd=30 → должно вернуть 30
	size := calc.Calculate(0, 0, 100.0)
	if size != 30.0 {
		t.Errorf("max cap: expected 30.0, got %f", size)
	}
}

func TestSizeCalculatorZeroTraderBalance(t *testing.T) {
	calc := copytrading.NewSizeCalculator("proportional", 10.0, 50.0)

	// traderBalance=0 → не делить на ноль, вернуть 0
	size := calc.Calculate(100.0, 0.0, 500.0)
	if size != 0.0 {
		t.Errorf("zero trader balance: expected 0, got %f", size)
	}
}
```

**Step 2: Run test to see it fail**

```bash
go test ./internal/copytrading/... -v -run TestSizeCalculator
```

Expected: compile error (sizer.go doesn't exist).

**Step 3: Implement `internal/copytrading/sizer.go`**

```go
package copytrading

// SizeCalculator вычисляет размер нашей позиции при копировании.
type SizeCalculator struct {
	mode           string  // "proportional" или "fixed_pct"
	allocationPct  float64 // % нашего баланса выделяемый трейдеру
	maxPositionUSD float64 // лимит на одну позицию
}

// NewSizeCalculator создаёт калькулятор размера позиции.
func NewSizeCalculator(mode string, allocationPct, maxPositionUSD float64) *SizeCalculator {
	return &SizeCalculator{
		mode:           mode,
		allocationPct:  allocationPct,
		maxPositionUSD: maxPositionUSD,
	}
}

// Calculate вычисляет размер нашей позиции в USD.
// traderPositionUSD — текущая стоимость позиции трейдера в USD
// traderTotalBalance — оценочный общий баланс трейдера в USD (из currentValue всех позиций, или InitialValue как proxy)
// myBalance — наш текущий баланс в USD
func (c *SizeCalculator) Calculate(traderPositionUSD, traderTotalBalance, myBalance float64) float64 {
	var size float64

	switch c.mode {
	case "fixed_pct":
		size = myBalance * c.allocationPct / 100.0
	default: // "proportional"
		if traderTotalBalance <= 0 {
			return 0
		}
		ratio := traderPositionUSD / traderTotalBalance
		size = ratio * myBalance * c.allocationPct / 100.0
	}

	if c.maxPositionUSD > 0 && size > c.maxPositionUSD {
		size = c.maxPositionUSD
	}
	return size
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/copytrading/... -v -run TestSizeCalculator
```

Expected: all 4 tests PASS.

**Step 5: Commit**

```bash
git add internal/copytrading/sizer.go internal/copytrading/sizer_test.go
git commit -m "feat: add SizeCalculator with proportional and fixed_pct modes"
```

---

### Task 9: OrderExecutor

**Files:**
- Create: `internal/copytrading/executor.go`

**Context:** Places and closes orders on Polymarket via CLOB API. Uses EIP-712 signing. For opening: market-buy at the best ask price (FOK order type for immediate fill). For closing: market-sell at best bid.

Amount calculations:
- BUY side=0: `makerAmount = price * size * 1e6` (USDC you spend), `takerAmount = size * 1e6` (shares you receive)
- SELL side=1: `makerAmount = size * 1e6` (shares you give), `takerAmount = price * size * 1e6` (USDC you receive)

**Step 1: Create `internal/copytrading/executor.go`**

```go
package copytrading

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

const (
	// decimals — количество десятичных знаков (6 для USDC и условных токенов Polymarket)
	decimals   = 1_000_000
	feeRateBps = 0 // по умолчанию
)

// OrderExecutor размещает и закрывает ордера на Polymarket.
type OrderExecutor struct {
	clob        *clob.Client
	orderSigner *auth.OrderSigner
	apiKey      string
	makerAddr   string
	logger      zerolog.Logger
}

// NewOrderExecutor создаёт OrderExecutor.
// apiKey — значение поля "owner" в теле запроса (L2 api_key).
func NewOrderExecutor(
	clobClient *clob.Client,
	orderSigner *auth.OrderSigner,
	apiKey string,
	makerAddr string,
	log zerolog.Logger,
) *OrderExecutor {
	return &OrderExecutor{
		clob:        clobClient,
		orderSigner: orderSigner,
		apiKey:      apiKey,
		makerAddr:   makerAddr,
		logger:      log.With().Str("component", "order-executor").Logger(),
	}
}

// OpenResult — результат открытия позиции.
type OpenResult struct {
	OrderID string
	Price   float64
	Size    float64
}

// Open размещает market-buy ордер для указанного токена.
// assetID — token_id (ERC-1155), sizeUSD — размер позиции в USD.
// negRisk — для рынков с несколькими исходами.
func (e *OrderExecutor) Open(assetID string, sizeUSD float64, negRisk bool) (*OpenResult, error) {
	// Получаем лучшую цену ask для покупки
	priceResp, err := e.clob.GetPrice(assetID, "BUY")
	if err != nil {
		return nil, fmt.Errorf("executor: get BUY price for %s: %w", assetID, err)
	}
	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil || price <= 0 {
		return nil, fmt.Errorf("executor: invalid price %q for %s", priceResp.Price, assetID)
	}

	// Размер в токенах: sizeUSD / price
	sizeShares := sizeUSD / price

	// Строим и подписываем ордер
	req, err := e.buildOrderRequest(assetID, price, sizeShares, auth.Buy, negRisk)
	if err != nil {
		return nil, fmt.Errorf("executor: build BUY order: %w", err)
	}

	resp, err := e.clob.CreateOrder(req)
	if err != nil {
		return nil, fmt.Errorf("executor: CreateOrder BUY: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("executor: order rejected: %s", resp.ErrorMsg)
	}

	e.logger.Info().
		Str("asset_id", assetID).
		Str("order_id", resp.OrderID).
		Float64("price", price).
		Float64("size_shares", sizeShares).
		Msg("opened copy position")

	return &OpenResult{OrderID: resp.OrderID, Price: price, Size: sizeShares}, nil
}

// CloseResult — результат закрытия позиции.
type CloseResult struct {
	OrderID string
	Price   float64
	PnL     float64
}

// Close продаёт позицию по лучшей цене bid.
// sizeShares — количество токенов для продажи, avgBuyPrice — средняя цена покупки (для расчёта P&L).
func (e *OrderExecutor) Close(assetID string, sizeShares, avgBuyPrice float64, negRisk bool) (*CloseResult, error) {
	// Получаем лучшую цену bid для продажи
	priceResp, err := e.clob.GetPrice(assetID, "SELL")
	if err != nil {
		return nil, fmt.Errorf("executor: get SELL price for %s: %w", assetID, err)
	}
	price, err := strconv.ParseFloat(priceResp.Price, 64)
	if err != nil || price <= 0 {
		return nil, fmt.Errorf("executor: invalid SELL price %q for %s", priceResp.Price, assetID)
	}

	req, err := e.buildOrderRequest(assetID, price, sizeShares, auth.Sell, negRisk)
	if err != nil {
		return nil, fmt.Errorf("executor: build SELL order: %w", err)
	}

	resp, err := e.clob.CreateOrder(req)
	if err != nil {
		return nil, fmt.Errorf("executor: CreateOrder SELL: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("executor: sell order rejected: %s", resp.ErrorMsg)
	}

	pnl := (price - avgBuyPrice) * sizeShares

	e.logger.Info().
		Str("asset_id", assetID).
		Str("order_id", resp.OrderID).
		Float64("price", price).
		Float64("pnl", pnl).
		Msg("closed copy position")

	return &CloseResult{OrderID: resp.OrderID, Price: price, PnL: pnl}, nil
}

// buildOrderRequest строит подписанный CreateOrderRequest.
func (e *OrderExecutor) buildOrderRequest(
	assetID string,
	price float64,
	sizeShares float64,
	side auth.OrderSide,
	negRisk bool,
) (*clob.CreateOrderRequest, error) {
	salt, err := auth.RandomSalt()
	if err != nil {
		return nil, err
	}

	makerAddr := common.HexToAddress(e.makerAddr)
	zeroAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// TokenID из строки в big.Int
	tokenID := new(big.Int)
	tokenID.SetString(strings.TrimPrefix(assetID, "0x"), 16)

	// Суммы в base units (6 decimals)
	priceInt := int64(price * float64(decimals))
	sizeInt := int64(sizeShares * float64(decimals))

	var makerAmount, takerAmount *big.Int
	if side == auth.Buy {
		// BUY: отдаём USDC (makerAmount), получаем токены (takerAmount)
		makerAmount = big.NewInt(priceInt * sizeInt / int64(decimals))
		takerAmount = big.NewInt(sizeInt)
	} else {
		// SELL: отдаём токены (makerAmount), получаем USDC (takerAmount)
		makerAmount = big.NewInt(sizeInt)
		takerAmount = big.NewInt(priceInt * sizeInt / int64(decimals))
	}

	rawOrder := &auth.RawOrder{
		Salt:          salt,
		Maker:         makerAddr,
		Signer:        makerAddr,
		Taker:         zeroAddr,
		TokenID:       tokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(feeRateBps),
		Side:          side,
		SignatureType: auth.EOA,
	}

	sig, err := e.orderSigner.Sign(rawOrder)
	if err != nil {
		return nil, fmt.Errorf("executor: sign order: %w", err)
	}

	sideInt := 0
	if side == auth.Sell {
		sideInt = 1
	}

	return &clob.CreateOrderRequest{
		Order: clob.SignedOrder{
			Salt:          salt.String(),
			Maker:         e.makerAddr,
			Signer:        e.makerAddr,
			Taker:         "0x0000000000000000000000000000000000000000",
			TokenID:       assetID,
			MakerAmount:   makerAmount.String(),
			TakerAmount:   takerAmount.String(),
			Expiration:    "0",
			Nonce:         "0",
			FeeRateBps:    strconv.Itoa(feeRateBps),
			Side:          sideInt,
			SignatureType: int(auth.EOA),
			Signature:     sig,
		},
		Owner:     e.apiKey,
		OrderType: clob.OrderTypeGTC,
	}, nil
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add internal/copytrading/executor.go
git commit -m "feat: add OrderExecutor for EIP-712 signed order placement"
```

---

### Task 10: TraderTracker

**Files:**
- Create: `internal/copytrading/tracker.go`

**Context:** Polls `data.Client.GetPositions(traderAddress)` every `poll_interval_ms`. On first run after startup, loads open DB records and reconciles (closes orphaned positions). On each subsequent poll, diffs prev vs current snapshot.

**Step 1: Create `internal/copytrading/tracker.go`**

```go
package copytrading

import (
	"context"
	"fmt"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/atlasdev/polytrade-bot/internal/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TraderTracker следит за одним трейдером и копирует его позиции.
type TraderTracker struct {
	trader     config.TraderConfig
	dataClient *data.Client
	executor   *OrderExecutor
	sizer      *SizeCalculator
	store      storage.CopyTradeStore
	notifier   notify.Notifier
	logger     zerolog.Logger

	// Получить текущий баланс нашего кошелька (USDC)
	getMyBalance func() (float64, error)

	prev TraderState
}

// NewTraderTracker создаёт трекер для одного трейдера.
func NewTraderTracker(
	trader config.TraderConfig,
	dataClient *data.Client,
	executor *OrderExecutor,
	store storage.CopyTradeStore,
	notifier notify.Notifier,
	getMyBalance func() (float64, error),
	log zerolog.Logger,
) *TraderTracker {
	sizer := NewSizeCalculator(trader.SizeMode, trader.AllocationPct, trader.MaxPositionUSD)
	return &TraderTracker{
		trader:       trader,
		dataClient:   dataClient,
		executor:     executor,
		sizer:        sizer,
		store:        store,
		notifier:     notifier,
		logger:       log.With().Str("trader", trader.Label).Str("address", trader.Address).Logger(),
		getMyBalance: getMyBalance,
		prev:         make(TraderState),
	}
}

// Run запускает цикл мониторинга. Блокирует до отмены ctx.
func (t *TraderTracker) Run(ctx context.Context, pollInterval time.Duration) error {
	t.logger.Info().Dur("interval", pollInterval).Msg("trader tracker started")

	// При старте: загружаем открытые записи и сверяем с текущими позициями
	if err := t.reconcile(ctx); err != nil {
		t.logger.Warn().Err(err).Msg("reconcile failed on startup")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			t.poll(ctx)
		}
	}
}

// reconcile сверяет открытые DB-записи с реальными позициями трейдера.
// Закрывает записи для позиций, которые трейдер закрыл пока бот был оффлайн.
func (t *TraderTracker) reconcile(ctx context.Context) error {
	openTrades, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		return fmt.Errorf("load open trades: %w", err)
	}
	if len(openTrades) == 0 {
		return nil
	}

	positions, err := t.dataClient.GetPositions(data.PositionsParams{
		User:  t.trader.Address,
		Limit: 200,
	})
	if err != nil {
		return fmt.Errorf("fetch positions: %w", err)
	}

	curr := toTraderState(positions)
	t.prev = curr

	for _, trade := range openTrades {
		if _, exists := curr[trade.AssetID]; !exists {
			t.logger.Warn().
				Str("asset_id", trade.AssetID).
				Str("trade_id", trade.ID).
				Msg("orphaned position: trader closed while bot was offline, closing our position")
			t.closePosition(ctx, trade)
		}
	}
	return nil
}

// poll выполняет один цикл опроса.
func (t *TraderTracker) poll(ctx context.Context) {
	positions, err := t.dataClient.GetPositions(data.PositionsParams{
		User:  t.trader.Address,
		Limit: 200,
	})
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to fetch trader positions")
		return
	}

	curr := toTraderState(positions)
	diff := diffStates(t.prev, curr)
	t.prev = curr

	for _, pos := range diff.Opened {
		t.openPosition(ctx, pos)
	}

	for _, pos := range diff.Closed {
		t.handleTraderClosed(ctx, pos)
	}
}

// openPosition копирует открытие позиции трейдера.
func (t *TraderTracker) openPosition(ctx context.Context, pos data.Position) {
	// Проверяем, нет ли уже открытой позиции в DB
	existing, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to check existing trades")
		return
	}
	for _, e := range existing {
		if e.AssetID == pos.Asset {
			t.logger.Debug().Str("asset_id", pos.Asset).Msg("position already tracked, skipping")
			return
		}
	}

	// Получаем наш баланс
	myBalance, err := t.getMyBalance()
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to get our balance")
		return
	}

	// Рассчитываем размер позиции
	// Для proportional: traderTotalBalance аппроксимируем как сумму всех CurrentValue
	traderTotalBalance := t.estimateTraderBalance()
	sizeUSD := t.sizer.Calculate(pos.CurrentValue, traderTotalBalance, myBalance)
	if sizeUSD <= 0 {
		t.logger.Warn().Float64("size_usd", sizeUSD).Msg("calculated size is 0, skipping")
		return
	}

	t.logger.Info().
		Str("asset_id", pos.Asset).
		Str("market", pos.Title).
		Str("outcome", pos.Outcome).
		Float64("size_usd", sizeUSD).
		Float64("trader_price", pos.AvgPrice).
		Msg("opening copy position")

	result, err := t.executor.Open(pos.Asset, sizeUSD, false)
	if err != nil {
		t.logger.Error().Err(err).Str("asset_id", pos.Asset).Msg("failed to open position")
		t.sendAlert(ctx, fmt.Sprintf("❌ Failed to copy open: [%s] %s - %s\nError: %v",
			t.trader.Label, pos.Title, pos.Outcome, err))
		return
	}

	// Сохраняем в DB
	rec := &storage.CopyTradeRecord{
		ID:            uuid.New().String(),
		TraderAddress: t.trader.Address,
		AssetID:       pos.Asset,
		ConditionID:   pos.ConditionID,
		Side:          "BUY",
		Size:          result.Size,
		Price:         result.Price,
		OurOrderID:    result.OrderID,
		Status:        "open",
		OpenedAt:      time.Now().UTC(),
	}
	if err := t.store.SaveCopyTrade(ctx, rec); err != nil {
		t.logger.Error().Err(err).Msg("failed to save copy trade to DB")
	}

	t.sendAlert(ctx, fmt.Sprintf("📈 Opened copy trade: [%s] %s (%s)\nSize: $%.2f @ %.3f",
		t.trader.Label, pos.Title, pos.Outcome, sizeUSD, result.Price))
}

// handleTraderClosed обрабатывает закрытие позиции трейдером.
func (t *TraderTracker) handleTraderClosed(ctx context.Context, pos data.Position) {
	openTrades, err := t.store.GetOpenCopyTrades(ctx, t.trader.Address)
	if err != nil {
		t.logger.Warn().Err(err).Msg("failed to load open trades for close")
		return
	}

	for _, trade := range openTrades {
		if trade.AssetID == pos.Asset {
			t.closePosition(ctx, trade)
			return
		}
	}
}

// closePosition закрывает нашу скопированную позицию.
func (t *TraderTracker) closePosition(ctx context.Context, trade *storage.CopyTradeRecord) {
	t.logger.Info().
		Str("asset_id", trade.AssetID).
		Str("trade_id", trade.ID).
		Float64("size", trade.Size).
		Msg("closing copy position")

	result, err := t.executor.Close(trade.AssetID, trade.Size, trade.Price, false)
	status := "closed"
	if err != nil {
		t.logger.Error().Err(err).Str("trade_id", trade.ID).Msg("failed to close position")
		status = "failed"
		t.sendAlert(ctx, fmt.Sprintf("❌ Failed to close copy: [%s] asset=%s\nError: %v",
			t.trader.Label, trade.AssetID, err))
	}

	now := time.Now().UTC()
	var pnl *float64
	if result != nil {
		pnl = &result.PnL
	}

	if err := t.store.UpdateCopyTrade(ctx, trade.ID, status, &now, pnl); err != nil {
		t.logger.Error().Err(err).Msg("failed to update copy trade status in DB")
	}

	if result != nil {
		t.sendAlert(ctx, fmt.Sprintf("📉 Closed copy trade: [%s] asset=%s\nPnL: $%.2f",
			t.trader.Label, trade.AssetID, result.PnL))
	}
}

// estimateTraderBalance суммирует CurrentValue всех позиций трейдера как приближение баланса.
func (t *TraderTracker) estimateTraderBalance() float64 {
	total := 0.0
	for _, pos := range t.prev {
		total += pos.CurrentValue
	}
	return total
}

func (t *TraderTracker) sendAlert(ctx context.Context, msg string) {
	if err := t.notifier.Send(ctx, msg); err != nil {
		t.logger.Warn().Err(err).Msg("failed to send alert")
	}
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add internal/copytrading/tracker.go
git commit -m "feat: add TraderTracker with position diffing and open/close logic"
```

---

### Task 11: CopyTrader orchestrator with hot-reload

**Files:**
- Create: `internal/copytrading/copier.go`

**Step 1: Create `internal/copytrading/copier.go`**

```go
package copytrading

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/auth"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/atlasdev/polytrade-bot/internal/storage"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

// CopyTrader — оркестратор копитрейдинга.
// Управляет TraderTracker горутинами и поддерживает горячую перезагрузку конфига.
type CopyTrader struct {
	clobClient  *clob.Client
	dataClient  *data.Client
	store       storage.CopyTradeStore
	notifier    notify.Notifier
	executor    *OrderExecutor
	cfg         *config.CopytradingConfig
	cfgPath     string
	reloadCfg   func(path string) (*config.Config, error)
	logger      zerolog.Logger

	mu       sync.Mutex
	trackers map[string]context.CancelFunc // address → cancel func
}

// New создаёт CopyTrader.
// cfgPath — путь к config.toml для hot-reload.
// reloadCfg — функция для перечитывания конфига (обычно config.Load).
func New(
	clobClient *clob.Client,
	dataClient *data.Client,
	store storage.CopyTradeStore,
	notifier notify.Notifier,
	orderSigner *auth.OrderSigner,
	apiKey string,
	makerAddr string,
	cfg *config.CopytradingConfig,
	cfgPath string,
	reloadCfg func(string) (*config.Config, error),
	log zerolog.Logger,
) *CopyTrader {
	executor := NewOrderExecutor(clobClient, orderSigner, apiKey, makerAddr, log)
	return &CopyTrader{
		clobClient: clobClient,
		dataClient: dataClient,
		store:      store,
		notifier:   notifier,
		executor:   executor,
		cfg:        cfg,
		cfgPath:    cfgPath,
		reloadCfg:  reloadCfg,
		logger:     log.With().Str("component", "copytrader").Logger(),
		trackers:   make(map[string]context.CancelFunc),
	}
}

// Run запускает копитрейдер. Блокирует до отмены ctx.
func (c *CopyTrader) Run(ctx context.Context) error {
	// Запускаем трекеры из начального конфига
	for _, trader := range c.cfg.Traders {
		if trader.Enabled {
			c.startTracker(ctx, trader)
		}
	}

	// Следим за изменениями config.toml
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		c.logger.Warn().Err(err).Msg("failed to create file watcher, hot-reload disabled")
		<-ctx.Done()
		return nil
	}
	defer watcher.Close()

	if err := watcher.Add(c.cfgPath); err != nil {
		c.logger.Warn().Err(err).Str("path", c.cfgPath).Msg("failed to watch config file, hot-reload disabled")
		<-ctx.Done()
		return nil
	}

	c.logger.Info().Str("config", c.cfgPath).Msg("copytrader running with hot-reload")

	for {
		select {
		case <-ctx.Done():
			c.stopAllTrackers()
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				c.logger.Info().Str("event", event.Name).Msg("config changed, reloading traders")
				c.reloadTraders(ctx)
			}
		case err, ok := <-watcher.Errors:
			if ok {
				c.logger.Warn().Err(err).Msg("file watcher error")
			}
		}
	}
}

// reloadTraders перечитывает конфиг и добавляет/останавливает трекеры.
func (c *CopyTrader) reloadTraders(ctx context.Context) {
	newCfg, err := c.reloadCfg(c.cfgPath)
	if err != nil {
		c.logger.Warn().Err(err).Msg("failed to reload config")
		return
	}

	newTraderSet := make(map[string]config.TraderConfig)
	for _, t := range newCfg.Copytrading.Traders {
		if t.Enabled {
			newTraderSet[t.Address] = t
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Остановить удалённых трейдеров
	for addr, cancel := range c.trackers {
		if _, exists := newTraderSet[addr]; !exists {
			c.logger.Info().Str("address", addr).Msg("stopping removed trader tracker")
			cancel()
			delete(c.trackers, addr)
			_ = c.notifier.Send(ctx, fmt.Sprintf("ℹ️ Trader removed from config: %s. Open positions NOT auto-closed.", addr))
		}
	}

	// Запустить новых трейдеров
	for addr, trader := range newTraderSet {
		if _, exists := c.trackers[addr]; !exists {
			c.logger.Info().Str("address", addr).Str("label", trader.Label).Msg("starting new trader tracker")
			c.startTrackerLocked(ctx, trader)
		}
	}
}

// startTracker запускает горутину TraderTracker.
func (c *CopyTrader) startTracker(ctx context.Context, trader config.TraderConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startTrackerLocked(ctx, trader)
}

func (c *CopyTrader) startTrackerLocked(ctx context.Context, trader config.TraderConfig) {
	trackerCtx, cancel := context.WithCancel(ctx)
	c.trackers[trader.Address] = cancel

	pollInterval := time.Duration(c.cfg.PollIntervalMs) * time.Millisecond

	tracker := NewTraderTracker(
		trader,
		c.dataClient,
		c.executor,
		c.store,
		c.notifier,
		c.getMyBalance,
		c.logger,
	)

	go func() {
		if err := tracker.Run(trackerCtx, pollInterval); err != nil && trackerCtx.Err() == nil {
			c.logger.Error().Err(err).Str("trader", trader.Label).Msg("trader tracker error")
		}
	}()
}

// stopAllTrackers останавливает все запущенные трекеры.
func (c *CopyTrader) stopAllTrackers() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for addr, cancel := range c.trackers {
		c.logger.Info().Str("address", addr).Msg("stopping tracker on shutdown")
		cancel()
	}
	c.trackers = make(map[string]context.CancelFunc)
}

// getMyBalance получает наш текущий баланс USDC из CLOB API.
func (c *CopyTrader) getMyBalance() (float64, error) {
	ba, err := c.clobClient.GetBalanceAllowance("COLLATERAL", "")
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	// Balance в wei (6 decimals), конвертируем в USD
	var balanceWei float64
	fmt.Sscanf(ba.Balance, "%f", &balanceWei)
	return balanceWei / 1_000_000, nil
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add internal/copytrading/copier.go
git commit -m "feat: add CopyTrader orchestrator with fsnotify hot-reload"
```

---

### Task 12: Wire up in main.go and enable SQLite

**Files:**
- Modify: `cmd/bot/main.go`

**Step 1: Modify `cmd/bot/main.go`**

Add imports:
```go
"github.com/atlasdev/polytrade-bot/internal/copytrading"
"github.com/atlasdev/polytrade-bot/internal/storage/sqlite"
```

After the `l2Creds` block (around line 70), add SQLite initialization:
```go
// --- Storage ---
var store storage.Store
if cfg.Database.Enabled {
    db, err := sqlite.Open(cfg.Database.Path)
    if err != nil {
        return fmt.Errorf("open sqlite: %w", err)
    }
    defer db.Close()
    store = db
    log.Info().Str("path", cfg.Database.Path).Msg("sqlite storage enabled")
}
```

After the trading engine block (around line 100), add copytrading:
```go
// --- Copytrading ---
if cfg.Copytrading.Enabled {
    if l2Creds == nil {
        log.Warn().Msg("copytrading requires L2 credentials, skipping")
    } else if l2Creds.Address == "" {
        log.Warn().Msg("copytrading requires private_key for order signing, skipping")
    } else if store == nil {
        log.Warn().Msg("copytrading requires database.enabled=true, skipping")
    } else {
        l1, err := auth.NewL1Signer(cfg.Auth.PrivateKey)
        if err != nil {
            return fmt.Errorf("copytrading: l1 signer: %w", err)
        }
        orderSigner := auth.NewOrderSigner(l1, cfg.Auth.ChainID, cfg.Trading.NegRisk)
        copier := copytrading.New(
            clobClient,
            dataClient,
            store,
            notifier,
            orderSigner,
            cfg.Auth.APIKey,
            l2Creds.Address,
            &cfg.Copytrading,
            *cfgPath,
            config.Load,
            log,
        )
        errCh = make(chan error, 5) // расширяем буфер
        go func() {
            if err := copier.Run(ctx); err != nil && ctx.Err() == nil {
                errCh <- fmt.Errorf("copytrading: %w", err)
            }
        }()
        log.Info().Int("traders", len(cfg.Copytrading.Traders)).Msg("copytrading enabled")
    }
}
```

Also update the `storage` import and add `auth` if needed.

**Step 2: Verify build**

```bash
go build ./...
```

Expected: no errors.

**Step 3: Verify vet**

```bash
go vet ./...
```

**Step 4: Run all tests**

```bash
go test ./...
```

**Step 5: Commit**

```bash
git add cmd/bot/main.go
git commit -m "feat: integrate CopyTrader and SQLite into main.go"
```

---

### Task 13: End-to-end smoke test (dry run)

**Goal:** Verify the bot starts correctly with copytrading enabled (disabled traders, no actual orders placed).

**Step 1: Enable copytrading in config with no traders**

In `config.toml`, temporarily set:
```toml
[copytrading]
  enabled = true
  poll_interval_ms = 10000
  size_mode = "proportional"
```

Also enable database:
```toml
[database]
  enabled = true
  path    = "./data/polytrade.db"
```

**Step 2: Create data directory and run**

```bash
mkdir -p "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot/data"
go run ./cmd/bot/ --config config.toml
```

Expected log output:
- `sqlite storage enabled`
- `copytrading enabled` with `traders=0`
- No errors
- Bot running, Ctrl+C cleanly shuts down

**Step 3: Revert config changes**

Restore `copytrading.enabled = false` and `database.enabled = false` for safety.

**Step 4: Final commit**

```bash
git add .
git commit -m "chore: restore default config after smoke test"
```

---

## Summary

| Task | Files Created/Modified | Key Outcome |
|------|----------------------|-------------|
| 1 | go.mod | modernc/sqlite, fsnotify, uuid |
| 2 | config.go, config.toml | CopytradingConfig with hot-reload |
| 3 | storage.go | CopyTradeRecord + CopyTradeStore interface |
| 4 | sqlite/sqlite.go + test | Full SQLite impl + copy_trades table |
| 5 | auth/order_signer.go + test | EIP-712 CTF Exchange order signing |
| 6 | clob/models.go | Updated SignedOrder with salt/signer |
| 7 | copytrading/models.go | TraderState, PositionDiff |
| 8 | copytrading/sizer.go + test | SizeCalculator (proportional + fixed_pct) |
| 9 | copytrading/executor.go | OrderExecutor (open/close via CLOB) |
| 10 | copytrading/tracker.go | TraderTracker (poll + diff + reconcile) |
| 11 | copytrading/copier.go | CopyTrader orchestrator + fsnotify |
| 12 | cmd/bot/main.go | Wire everything up |
| 13 | — | Smoke test and verify |
