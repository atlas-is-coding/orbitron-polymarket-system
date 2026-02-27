# Test Suite Design — polytrade-bot

**Date:** 2026-02-26
**Scope:** Public + private (integration) tests for all packages
**Libraries:** Go stdlib `testing` + `github.com/stretchr/testify`

---

## Approach: Build Tags (Option B)

- **Without tag** — public tests, always run: `go test ./...`
- **`//go:build integration`** — private tests requiring auth: `go test -tags=integration ./...`

---

## Auth Setup

Only one env variable required:

```
POLY_PRIVATE_KEY   # hex, with or without 0x prefix (code strips it)
POLY_CHAIN_ID      # optional, default: 137 (Polygon Mainnet)
```

Flow: `POLY_PRIVATE_KEY` → `L1Signer` → `POST /auth/api-key` → L2 credentials (api_key, secret, passphrase) derived automatically. No manual credential copying needed.

```go
func TestMain(m *testing.M) {
    rawKey := os.Getenv("POLY_PRIVATE_KEY")
    if rawKey == "" {
        fmt.Println("skipping: POLY_PRIVATE_KEY not set")
        os.Exit(0)
    }
    rawKey = strings.TrimPrefix(rawKey, "0x")
    l1, _ := auth.NewL1Signer(rawKey)
    creds, _ = clobClient.CreateOrGetAPIKey(l1)
    os.Exit(m.Run())
}
```

---

## Package Coverage

### Public Tests (no auth, always run)

| Package | Tests |
|---|---|
| `internal/api/gamma` | `GetMarkets()`, `GetMarket()`, `GetEvents()`, `flexFloat64` parsing |
| `internal/api/data` | `GetPositions()`, `GetTrades()` with known public address |
| `internal/api/clob` | `GetMarkets()`, `GetOrderBook()` |
| `internal/auth` | `NewL1Signer()`, `L2Headers()`, `OrderSigner.Sign()`, `RandomSalt()` |
| `internal/config` | `Load()`, `Save()`, `ConfigWatcher` |
| `internal/copytrading` | `SizeCalculator` (all modes + edge cases) |
| `internal/storage/sqlite` | `TradeStore`, `OrderStore`, `CopyTradeStore` |
| `internal/i18n` | `SetLanguage()`, `T()`, all 5 locales load without error |
| `internal/trading` | `Engine` with fake Strategy — start/stop goroutines |
| `internal/tui` | `EventBus.Send()`, `Tap()`, message delivery |

### Integration Tests (`//go:build integration`, require `POLY_PRIVATE_KEY`)

| Package | Tests |
|---|---|
| `internal/api/clob` | `CreateOrGetAPIKey()`, `GetOpenOrders()`, `GetPositions()`, `GetTrades()` |
| `internal/monitor` | `TradesMonitor` — single poll cycle, correct orders/positions parsing |

---

## File Structure

```
internal/testutil/testutil.go          # shared helpers (NewCLOBClient, LoadL1Creds)
internal/api/gamma/gamma_test.go
internal/api/data/data_test.go
internal/api/clob/clob_test.go         # public + integration sections
internal/i18n/i18n_test.go
internal/trading/engine_test.go
internal/tui/eventbus_test.go
internal/monitor/trades_test.go        # integration only
```

---

## Test Patterns

```go
// Public API test
func TestGetMarkets(t *testing.T) {
    client := testutil.NewCLOBClient()
    markets, err := client.GetMarkets(ctx, clob.MarketsParams{Limit: 5})
    require.NoError(t, err)
    require.NotEmpty(t, markets)
    assert.NotEmpty(t, markets[0].ConditionID)
}

// Integration test
//go:build integration

func TestGetOpenOrders(t *testing.T) {
    _, creds := testutil.LoadL1Creds(t) // calls t.Skip() if no key
    client := clob.NewAuthClient(creds)
    orders, err := client.GetOpenOrders(context.Background())
    require.NoError(t, err)
    // orders may be empty — that's valid
    assert.NotNil(t, orders)
}
```

**Rule:** If a test fails → fix the bug in production code, not in the test.

---

## Bug Fixes

Any test failures encountered during implementation are fixed in the corresponding production code. Tests describe expected behavior; production code must conform.
