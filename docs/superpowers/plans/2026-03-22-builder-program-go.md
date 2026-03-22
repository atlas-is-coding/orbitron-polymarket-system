# BUILDER PROGRAM — Go (polytrade-bot) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add BuilderKeyValidator and OrderExecutionLogger to verify and audit that BuilderApiKey is present in every production order.

**Architecture:** Two new files in `internal/builder/` — validator validates credentials at startup and logs warnings, logger audits every order submission. Both are wired into existing `cmd/bot/main.go` and `internal/copytrading/executor.go` with minimal changes.

**Tech Stack:** Go 1.24, zerolog, go test.

**Module:** `github.com/atlasdev/orbitron` — use this exact path in ALL imports.
New package import path: `github.com/atlasdev/orbitron/internal/builder`

**Import dependency chain (no cycles):**
`internal/builder` → `internal/license` (for BuilderCredentials) — OK
`internal/copytrading` → `internal/builder` — OK (copytrading does not currently import builder)
`internal/wallet` → `internal/builder` — OK (follows same pattern as existing builder key wiring)

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/builder/validator.go` | BuilderKeyValidator — validates key/expiry |
| Create | `internal/builder/validator_test.go` | Unit tests for validator |
| Create | `internal/builder/logger.go` | OrderExecutionLogger — per-order audit counters |
| Create | `internal/builder/logger_test.go` | Unit tests for logger |
| Modify | `internal/copytrading/executor.go` | Add `orderLogger` field, `WithOrderLogger()`, call `LogOrder()` |
| Modify | `internal/wallet/manager.go` | Add `orderLogger` field, `SetOrderLogger()`, chain into BOTH NewOrderExecutor call sites (line 179 and line 596) |
| Modify | `cmd/bot/main.go` | Construct validator + logger, call validator after `license.Load()` |

---

## Task 1: BuilderKeyValidator

**Files:**
- Create: `internal/builder/validator.go`
- Create: `internal/builder/validator_test.go`

- [ ] **Step 1.1 — Write failing tests**

```go
// internal/builder/validator_test.go
package builder_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/atlasdev/orbitron/internal/license"
	"github.com/rs/zerolog"
)

func nopLog() zerolog.Logger { return zerolog.Nop() }

func TestValidator_NilCreds(t *testing.T) {
	v := builder.NewBuilderKeyValidator(nil, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("nil creds should not be valid")
	}
	if r.Reason == "" {
		t.Fatal("reason should be set")
	}
}

func TestValidator_EmptyKey(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("empty key should not be valid")
	}
}

func TestValidator_Expired(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if r.Valid {
		t.Fatal("expired key should not be valid")
	}
	if r.DaysUntilExpiry >= 0 {
		t.Fatalf("DaysUntilExpiry should be negative, got %d", r.DaysUntilExpiry)
	}
}

func TestValidator_Valid(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(10 * 24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if !r.Valid {
		t.Fatalf("expected valid, got reason: %s", r.Reason)
	}
	if r.DaysUntilExpiry < 9 || r.DaysUntilExpiry > 11 {
		t.Fatalf("expected ~10 days, got %d", r.DaysUntilExpiry)
	}
}

func TestValidator_SoonExpiry(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "testkey",
		ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
	}
	v := builder.NewBuilderKeyValidator(creds, nopLog())
	r := v.Check()
	if !r.Valid {
		t.Fatal("key expiring in 3 days should still be valid")
	}
	if r.DaysUntilExpiry > 7 {
		t.Fatal("should be flagged as soon-expiring")
	}
}
```

- [ ] **Step 1.2 — Run tests to confirm they fail**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/builder/... 2>&1
```

Expected: `cannot find package "github.com/atlasdev/orbitron/internal/builder"`

- [ ] **Step 1.3 — Implement validator**

```go
// internal/builder/validator.go
package builder

import (
	"time"

	"github.com/atlasdev/orbitron/internal/license"
	"github.com/rs/zerolog"
)

// ValidationResult holds the result of a builder key validation check.
type ValidationResult struct {
	Valid           bool
	DaysUntilExpiry int    // negative if expired
	Reason          string // empty if valid
}

// BuilderKeyValidator checks that builder credentials are present and not expired.
type BuilderKeyValidator struct {
	creds  *license.BuilderCredentials
	logger zerolog.Logger
}

// NewBuilderKeyValidator creates a validator. creds may be nil (no license token configured).
func NewBuilderKeyValidator(creds *license.BuilderCredentials, log zerolog.Logger) *BuilderKeyValidator {
	return &BuilderKeyValidator{creds: creds, logger: log}
}

// Check validates the credentials and logs the result. Always non-fatal.
func (v *BuilderKeyValidator) Check() ValidationResult {
	if v.creds == nil {
		v.logger.Debug().Msg("builder: no license token configured — builder attribution disabled")
		return ValidationResult{Reason: "no license token configured"}
	}
	if v.creds.APIKey == "" {
		v.logger.Error().Msg("builder: API key is empty — orders will NOT be attributed")
		return ValidationResult{Reason: "API key is empty"}
	}

	now := time.Now()
	days := int(v.creds.ExpiresAt.Sub(now).Hours() / 24)

	if now.After(v.creds.ExpiresAt) {
		v.logger.Error().
			Int("days_expired", -days).
			Msg("builder: API key has EXPIRED — orders are NOT attributed")
		return ValidationResult{DaysUntilExpiry: days, Reason: "key expired"}
	}

	if days < 7 {
		v.logger.Warn().
			Int("days_until_expiry", days).
			Msg("builder: API key expiring soon — renew via Polymarket")
	} else {
		v.logger.Info().
			Int("days_until_expiry", days).
			Str("key_prefix", v.creds.APIKey[:min(4, len(v.creds.APIKey))]+"***").
			Msg("builder: API key valid")
	}

	return ValidationResult{Valid: true, DaysUntilExpiry: days}
}
```

- [ ] **Step 1.4 — Run tests to confirm they pass**

```bash
go test ./internal/builder/... -v -run TestValidator
```

Expected: all 5 tests PASS

- [ ] **Step 1.5 — Commit**

```bash
git add internal/builder/validator.go internal/builder/validator_test.go
git commit -m "feat(builder): add BuilderKeyValidator"
```

---

## Task 2: OrderExecutionLogger

**Files:**
- Create: `internal/builder/logger.go`
- Create: `internal/builder/logger_test.go`

- [ ] **Step 2.1 — Write failing tests**

```go
// internal/builder/logger_test.go
package builder_test

import (
	"sync"
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/builder"
)

func TestLogger_CountsOrders(t *testing.T) {
	l := builder.NewOrderExecutionLogger(nopLog())

	l.LogOrder(builder.OrderLogEntry{OrderID: "a", BuilderKeySet: true, Timestamp: time.Now(), Success: true})
	l.LogOrder(builder.OrderLogEntry{OrderID: "b", BuilderKeySet: false, Timestamp: time.Now(), Success: true})
	l.LogOrder(builder.OrderLogEntry{OrderID: "c", BuilderKeySet: true, Timestamp: time.Now(), Success: false})

	total, withKey, withoutKey := l.Summary()
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if withKey != 2 {
		t.Fatalf("expected withKey=2, got %d", withKey)
	}
	if withoutKey != 1 {
		t.Fatalf("expected withoutKey=1, got %d", withoutKey)
	}
}

func TestLogger_ThreadSafe(t *testing.T) {
	l := builder.NewOrderExecutionLogger(nopLog())
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.LogOrder(builder.OrderLogEntry{BuilderKeySet: true, Timestamp: time.Now(), Success: true})
		}()
	}
	wg.Wait()
	total, _, _ := l.Summary()
	if total != 100 {
		t.Fatalf("expected 100, got %d", total)
	}
}
```

- [ ] **Step 2.2 — Run tests to confirm they fail**

```bash
go test ./internal/builder/... -run TestLogger 2>&1
```

Expected: FAIL — `builder.OrderLogEntry` undefined

- [ ] **Step 2.3 — Implement logger**

```go
// internal/builder/logger.go
package builder

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// OrderLogEntry describes a single order submission for audit purposes.
type OrderLogEntry struct {
	OrderID       string
	BuilderKeySet bool
	Timestamp     time.Time
	Success       bool
}

// OrderExecutionLogger audits order submissions and tracks builder key attribution.
// All fields protected by mu — no atomic primitives used (consistent single approach).
type OrderExecutionLogger struct {
	mu         sync.Mutex
	total      int64
	withKey    int64
	withoutKey int64
	log        zerolog.Logger
}

// NewOrderExecutionLogger creates a new logger. Thread-safe.
func NewOrderExecutionLogger(log zerolog.Logger) *OrderExecutionLogger {
	return &OrderExecutionLogger{
		log: log.With().Str("component", "builder-logger").Logger(),
	}
}

// LogOrder records an order. Logs summary every 100 orders.
// Snapshot totals are taken under the lock then used outside it — safe.
func (l *OrderExecutionLogger) LogOrder(entry OrderLogEntry) {
	l.mu.Lock()
	l.total++
	if entry.BuilderKeySet {
		l.withKey++
	} else {
		l.withoutKey++
	}
	total := l.total
	withKey := l.withKey
	withoutKey := l.withoutKey
	l.mu.Unlock()

	l.log.Debug().
		Str("order_id", entry.OrderID).
		Bool("builder_key_set", entry.BuilderKeySet).
		Bool("success", entry.Success).
		Msg("order submitted")

	// Log summary every 100 orders. total is a local snapshot captured under lock above.
	if total%100 == 0 {
		l.log.Info().
			Int64("total_orders", total).
			Int64("with_builder_key", withKey).
			Int64("without_builder_key", withoutKey).
			Msg("builder attribution summary")
	}
}

// Summary returns current counters. Thread-safe.
func (l *OrderExecutionLogger) Summary() (total, withKey, withoutKey int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.total, l.withKey, l.withoutKey
}
```

- [ ] **Step 2.4 — Run all builder tests**

```bash
go test ./internal/builder/... -v
```

Expected: all tests PASS

- [ ] **Step 2.5 — Commit**

```bash
git add internal/builder/logger.go internal/builder/logger_test.go
git commit -m "feat(builder): add OrderExecutionLogger"
```

---

## Task 3: Wire OrderExecutionLogger into executor.go

**Files:**
- Modify: `internal/copytrading/executor.go`

- [ ] **Step 3.1 — Add orderLogger field and WithOrderLogger() to OrderExecutor**

In `internal/copytrading/executor.go`, add to the `OrderExecutor` struct and constructor:

```go
// Add to imports:
"github.com/atlasdev/orbitron/internal/builder"

// Add field to OrderExecutor struct (after builderAPIKey line):
orderLogger *builder.OrderExecutionLogger

// Add method after WithBuilderKey():
func (e *OrderExecutor) WithOrderLogger(l *builder.OrderExecutionLogger) *OrderExecutor {
	e.orderLogger = l
	return e
}
```

- [ ] **Step 3.2 — Add LogOrder call after each order submission**

Find the place in `executor.go` where `clob.CreateOrder` (or equivalent) is called and a response is received. After each such call, add:

```go
if e.orderLogger != nil {
	e.orderLogger.LogOrder(builder.OrderLogEntry{
		OrderID:       resp.OrderID, // use actual field name from CreateOrderResponse
		BuilderKeySet: e.builderAPIKey != "",
		Timestamp:     time.Now(),
		Success:       resp.Success,
	})
}
```

Check what the actual return type and field names are in `internal/api/clob/models.go:CreateOrderResponse` before writing this.

- [ ] **Step 3.3 — Verify it compiles**

```bash
go build ./internal/copytrading/...
```

Expected: no errors

- [ ] **Step 3.4 — Commit**

```bash
git add internal/copytrading/executor.go
git commit -m "feat(builder): wire OrderExecutionLogger into executor"
```

---

## Task 4: Wire validator and logger into main.go

**Files:**
- Modify: `cmd/bot/main.go`

- [ ] **Step 4.1 — Add imports**

Add to imports in `cmd/bot/main.go`:

```go
"github.com/atlasdev/orbitron/internal/builder"
```

- [ ] **Step 4.2 — Construct and run validator after license.Load()**

Find the block starting at line ~274 where `builderCreds, licenseErr := license.Load()` is called.
After the existing log message (line ~279), add:

```go
// Validate builder key and log result.
validator := builder.NewBuilderKeyValidator(builderCreds, log)
validator.Check()

// Construct shared order logger (passed to all executors).
orderLogger := builder.NewOrderExecutionLogger(log)
```

- [ ] **Step 4.3 — Pass orderLogger to wallet manager / executors**

Find where `OrderExecutor` instances are created (via wallet manager). The wallet manager's `SetBuilderKey` already injects the API key. Add similar injection for the logger:

```go
// After wm.SetBuilderKey(builderCreds.APIKey):
wm.SetOrderLogger(orderLogger)
```

**Note:** This requires adding `SetOrderLogger(*builder.OrderExecutionLogger)` to `internal/wallet/manager.go` and propagating it to each executor via `WithOrderLogger()`. Follow the same pattern as `SetBuilderKey`.

- [ ] **Step 4.4 — Add SetOrderLogger to wallet/manager.go**

In `internal/wallet/manager.go`, following the exact same pattern as `SetBuilderKey` / `builderKey`:

```go
// Add import:
"github.com/atlasdev/orbitron/internal/builder"

// Add field to Manager struct:
orderLogger *builder.OrderExecutionLogger

// Add method:
func (m *Manager) SetOrderLogger(l *builder.OrderExecutionLogger) {
	m.orderLogger = l
}
```

Then find **both** call sites of `NewOrderExecutor` in manager.go and chain `.WithOrderLogger(m.orderLogger)` at each:

- **Line ~179** in `Activate()` — wallet instance activation
- **Line ~596** in `PlaceOrder()` — direct order placement path

Both must be updated. Leaving either one un-wired means manual orders skip attribution logging.

```go
// Both sites follow the same pattern:
executor := copytrading.NewOrderExecutor(...).
	WithBuilderKey(m.builderKey).
	WithOrderLogger(m.orderLogger)
```

- [ ] **Step 4.5 — Build the full binary**

```bash
go build ./cmd/bot/...
```

Expected: no errors

- [ ] **Step 4.6 — Run all tests**

```bash
go test ./...
```

Expected: all tests pass

- [ ] **Step 4.7 — Commit**

```bash
git add cmd/bot/main.go internal/wallet/manager.go
git commit -m "feat(builder): wire validator and logger into startup and wallet manager"
```

---

## Task 5: Integration test — verify full attribution flow

**Files:**
- Create: `internal/builder/integration_test.go`

- [ ] **Step 5.1 — Write integration test**

```go
// internal/builder/integration_test.go
package builder_test

import (
	"testing"
	"time"

	"github.com/atlasdev/orbitron/internal/builder"
	"github.com/atlasdev/orbitron/internal/license"
)

// TestFullAttributionFlow verifies that validator + logger work together correctly.
func TestFullAttributionFlow(t *testing.T) {
	creds := &license.BuilderCredentials{
		APIKey:    "test-attribution-key",
		ExpiresAt: time.Now().Add(15 * 24 * time.Hour),
	}
	log := nopLog()

	// Validator reports valid
	v := builder.NewBuilderKeyValidator(creds, log)
	result := v.Check()
	if !result.Valid {
		t.Fatalf("expected valid credentials, got: %s", result.Reason)
	}
	if result.DaysUntilExpiry < 14 {
		t.Fatalf("expected ~15 days, got %d", result.DaysUntilExpiry)
	}

	// Logger correctly separates attributed vs non-attributed orders
	l := builder.NewOrderExecutionLogger(log)
	for i := 0; i < 5; i++ {
		l.LogOrder(builder.OrderLogEntry{
			OrderID:       "order-with-key",
			BuilderKeySet: true,
			Timestamp:     time.Now(),
			Success:       true,
		})
	}
	l.LogOrder(builder.OrderLogEntry{
		OrderID:       "order-without-key",
		BuilderKeySet: false,
		Timestamp:     time.Now(),
		Success:       true,
	})

	total, withKey, withoutKey := l.Summary()
	if total != 6 {
		t.Fatalf("expected total=6, got %d", total)
	}
	if withKey != 5 {
		t.Fatalf("expected withKey=5, got %d", withKey)
	}
	if withoutKey != 1 {
		t.Fatalf("expected withoutKey=1, got %d", withoutKey)
	}
}
```

- [ ] **Step 5.2 — Run integration test**

```bash
go test ./internal/builder/... -v -run TestFullAttributionFlow
```

Expected: PASS

- [ ] **Step 5.3 — Run complete test suite**

```bash
go test ./...
```

Expected: all tests PASS

- [ ] **Step 5.4 — Build binary**

```bash
go build ./cmd/bot/...
```

Expected: no errors

- [ ] **Step 5.5 — Final commit**

```bash
git add internal/builder/integration_test.go
git commit -m "test(builder): add full attribution flow integration test"
git commit -m "feat: BUILDER PROGRAM Go monitoring complete" --allow-empty
```
