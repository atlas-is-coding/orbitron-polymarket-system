# Test Suite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Write integration + unit tests for all public and private Polymarket API endpoints, plus core packages (auth, i18n, trading engine, EventBus).

**Architecture:** Подход B — build-теги. Публичные тесты без тега, приватные — `//go:build integration`. Shared helpers в `internal/testutil/`. Новый метод `DeriveAPIKey` добавляется в CLOB клиент для получения L2 credentials из приватного ключа.

**Tech Stack:** Go stdlib `testing`, `github.com/stretchr/testify` (assert/require), реальный Polymarket API (mainnet).

---

## Task 1: Add testify + DeriveAPIKey method

**Files:**
- Modify: `go.mod`, `go.sum` (добавить testify)
- Create: `internal/api/clob/auth.go`
- Create: `internal/testutil/testutil.go`

**Step 1: Добавить testify**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go get github.com/stretchr/testify@latest
go mod tidy
```

Expected: `go.mod` содержит `github.com/stretchr/testify`.

**Step 2: Создать `internal/api/clob/auth.go`**

```go
package clob

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/atlasdev/polytrade-bot/internal/auth"
)

// APIKeyCreds — ответ на /auth/derive-api-key и /auth/api-key (POST).
type APIKeyCreds struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKey получает существующие L2 credentials через L1 подпись.
// Вызывает GET /auth/derive-api-key с L1 заголовками (nonce=0).
func (c *Client) DeriveAPIKey(l1 *auth.L1Signer) (*auth.L2Credentials, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "0"
	headers, err := l1.L1Headers(ts, nonce)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: sign: %w", err)
	}
	resp, err := c.http.Get("/auth/derive-api-key", headers)
	if err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clob: DeriveAPIKey: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	var creds APIKeyCreds
	if err := json.Unmarshal(resp.Body, &creds); err != nil {
		return nil, fmt.Errorf("clob: DeriveAPIKey: decode: %w", err)
	}
	return &auth.L2Credentials{
		APIKey:     creds.APIKey,
		APISecret:  creds.Secret,
		Passphrase: creds.Passphrase,
		Address:    l1.Address(),
	}, nil
}
```

**Step 3: Создать `internal/testutil/testutil.go`**

```go
package testutil

import (
	"os"
	"strings"
	"testing"

	"github.com/atlasdev/polytrade-bot/internal/api"
	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
	"github.com/atlasdev/polytrade-bot/internal/auth"
)

const (
	ClobURL  = "https://clob.polymarket.com"
	GammaURL = "https://gamma-api.polymarket.com"
	DataURL  = "https://data-api.polymarket.com"
)

// NewCLOBClient возвращает публичный CLOB клиент (без auth).
func NewCLOBClient() *clob.Client {
	h := api.NewClient(ClobURL, 10, 2)
	return clob.NewClient(h, nil)
}

// NewGammaClient возвращает Gamma API клиент.
func NewGammaClient() *gamma.Client {
	h := api.NewClient(GammaURL, 10, 2)
	return gamma.NewClient(h)
}

// NewDataClient возвращает Data API клиент.
func NewDataClient() *data.Client {
	h := api.NewClient(DataURL, 10, 2)
	return data.NewClient(h)
}

// LoadPrivateKey читает POLY_PRIVATE_KEY из env, обрезает "0x", вызывает t.Skip если не задан.
func LoadPrivateKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("POLY_PRIVATE_KEY")
	if key == "" {
		t.Skip("POLY_PRIVATE_KEY not set — skipping integration test")
	}
	return strings.TrimPrefix(key, "0x")
}

// LoadL1Signer создаёт L1Signer из POLY_PRIVATE_KEY, пропускает тест если ключ не задан.
func LoadL1Signer(t *testing.T) *auth.L1Signer {
	t.Helper()
	rawKey := LoadPrivateKey(t)
	l1, err := auth.NewL1Signer(rawKey)
	if err != nil {
		t.Fatalf("testutil: NewL1Signer: %v", err)
	}
	return l1
}

// LoadL2Creds выводит L2 credentials через DeriveAPIKey. Вызывает t.Skip если нет ключа.
func LoadL2Creds(t *testing.T) (*auth.L1Signer, *auth.L2Credentials) {
	t.Helper()
	l1 := LoadL1Signer(t)
	pubClient := NewCLOBClient()
	creds, err := pubClient.DeriveAPIKey(l1)
	if err != nil {
		t.Fatalf("testutil: DeriveAPIKey: %v", err)
	}
	return l1, creds
}

// NewAuthCLOBClient возвращает CLOB клиент с L2 credentials.
func NewAuthCLOBClient(creds *auth.L2Credentials) *clob.Client {
	h := api.NewClient(ClobURL, 10, 2)
	return clob.NewClient(h, creds)
}
```

**Step 4: Убедиться что компилируется**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go build ./...
```

Expected: no errors.

**Step 5: Commit**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
git add go.mod go.sum internal/api/clob/auth.go internal/testutil/testutil.go
git commit -m "feat(test): add testify, DeriveAPIKey method, testutil package"
```

---

## Task 2: Tests for internal/auth

**Files:**
- Create: `internal/auth/l1_test.go`
- Create: `internal/auth/l2_test.go`

Тесты для L1Signer и L2Credentials. Не требуют сети. Используют фиктивный приватный ключ.

**Step 1: Создать `internal/auth/l1_test.go`**

```go
package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тестовый приватный ключ (не используется на mainnet).
const testPrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewL1Signer_Valid(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)
	assert.NotNil(t, l1)
	assert.NotEmpty(t, l1.Address())
	// Известный адрес для этого ключа (Foundry anvil key #0)
	assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", l1.Address())
}

func TestNewL1Signer_With0xPrefix(t *testing.T) {
	l1WithPrefix, err := NewL1Signer("0x" + testPrivKey)
	require.NoError(t, err)

	l1Without, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)

	// Оба должны дать одинаковый адрес
	assert.Equal(t, l1Without.Address(), l1WithPrefix.Address())
}

func TestNewL1Signer_InvalidKey(t *testing.T) {
	_, err := NewL1Signer("not-hex-key")
	assert.Error(t, err)
}

func TestNewL1Signer_EmptyKey(t *testing.T) {
	_, err := NewL1Signer("")
	assert.Error(t, err)
}

func TestL1Signer_Sign(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)

	sig, err := l1.Sign([]byte("hello"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(sig, "0x"), "signature must start with 0x")
	assert.Len(t, sig, 132, "65-byte signature = 130 hex chars + '0x' prefix")
}

func TestL1Signer_Sign_Deterministic(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)

	sig1, err := l1.Sign([]byte("test"))
	require.NoError(t, err)
	sig2, err := l1.Sign([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, sig1, sig2)
}

func TestL1Headers(t *testing.T) {
	l1, err := NewL1Signer(testPrivKey)
	require.NoError(t, err)

	headers, err := l1.L1Headers("1700000000", "0")
	require.NoError(t, err)

	assert.Equal(t, l1.Address(), headers["POLY_ADDRESS"])
	assert.Equal(t, "1700000000", headers["POLY_TIMESTAMP"])
	assert.Equal(t, "0", headers["POLY_NONCE"])
	assert.True(t, strings.HasPrefix(headers["POLY_SIGNATURE"], "0x"))
}
```

**Step 2: Создать `internal/auth/l2_test.go`**

```go
package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestL2Headers_Structure(t *testing.T) {
	creds := &L2Credentials{
		APIKey:     "test-api-key",
		APISecret:  "dGVzdC1zZWNyZXQ=", // base64 "test-secret"
		Passphrase: "test-pass",
		Address:    "0x1234",
	}

	headers, err := creds.L2Headers("GET", "/orders", "")
	require.NoError(t, err)

	assert.Equal(t, creds.Address, headers["POLY_ADDRESS"])
	assert.Equal(t, creds.APIKey, headers["POLY_API_KEY"])
	assert.Equal(t, creds.Passphrase, headers["POLY_PASSPHRASE"])
	assert.NotEmpty(t, headers["POLY_TIMESTAMP"])
	assert.NotEmpty(t, headers["POLY_SIGNATURE"])
}

func TestL2Headers_DifferentMethods(t *testing.T) {
	creds := &L2Credentials{
		APIKey:     "key",
		APISecret:  "c2VjcmV0",
		Passphrase: "pass",
		Address:    "0xabc",
	}

	h1, err := creds.L2Headers("GET", "/orders", "")
	require.NoError(t, err)
	h2, err := creds.L2Headers("POST", "/order", `{"test":"body"}`)
	require.NoError(t, err)

	// Разные запросы должны давать разные подписи
	assert.NotEqual(t, h1["POLY_SIGNATURE"], h2["POLY_SIGNATURE"])
}
```

**Step 3: Запустить тесты**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/auth/... -v -run "TestNewL1Signer|TestL1|TestL2"
```

Expected: все тесты PASS. Если падают — исправить баг в коде.

**Step 4: Commit**

```bash
git add internal/auth/l1_test.go internal/auth/l2_test.go
git commit -m "test(auth): add L1Signer and L2Credentials unit tests"
```

---

## Task 3: Tests for internal/api/gamma

**Files:**
- Create: `internal/api/gamma/gamma_test.go`

Публичные тесты, сеть нужна, auth не нужен.

**Step 1: Создать `internal/api/gamma/gamma_test.go`**

```go
package gamma_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/testutil"
)

func TestGetMarkets_ReturnsResults(t *testing.T) {
	client := testutil.NewGammaClient()
	active := true
	markets, err := client.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 5})
	require.NoError(t, err)
	require.NotEmpty(t, markets)

	m := markets[0]
	assert.NotEmpty(t, m.ConditionID, "conditionId should not be empty")
	assert.NotEmpty(t, m.Question, "question should not be empty")
}

func TestGetMarkets_WithLimit(t *testing.T) {
	client := testutil.NewGammaClient()
	markets, err := client.GetMarkets(gamma.MarketsParams{Limit: 3})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(markets), 3)
}

func TestGetMarket_ByConditionID(t *testing.T) {
	client := testutil.NewGammaClient()
	// Получаем первый рынок для test
	active := true
	markets, err := client.GetMarkets(gamma.MarketsParams{Active: &active, Limit: 1})
	require.NoError(t, err)
	require.NotEmpty(t, markets)

	condID := markets[0].ConditionID
	m, err := client.GetMarket(condID)
	require.NoError(t, err)
	assert.Equal(t, condID, m.ConditionID)
}

func TestGetEvents_ReturnsResults(t *testing.T) {
	client := testutil.NewGammaClient()
	active := true
	events, err := client.GetEvents(gamma.EventsParams{Active: &active, Limit: 5})
	require.NoError(t, err)
	require.NotEmpty(t, events)
	assert.NotEmpty(t, events[0].ID)
}

func TestGetEvents_WithLimit(t *testing.T) {
	client := testutil.NewGammaClient()
	events, err := client.GetEvents(gamma.EventsParams{Limit: 2})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(events), 2)
}

func TestGetEvent_ByID(t *testing.T) {
	client := testutil.NewGammaClient()
	events, err := client.GetEvents(gamma.EventsParams{Limit: 1})
	require.NoError(t, err)
	require.NotEmpty(t, events)

	ev, err := client.GetEvent(events[0].ID)
	require.NoError(t, err)
	assert.Equal(t, events[0].ID, ev.ID)
}

// TestFlexFloat64_ParsesStringAndNumber проверяет что flexFloat64 корректно
// десериализует поля liquidity/volume которые API возвращает как string или number.
func TestFlexFloat64_ParsesStringAndNumber(t *testing.T) {
	// Тест парсинга flexFloat64 через реальный ответ Gamma API
	client := testutil.NewGammaClient()
	markets, err := client.GetMarkets(gamma.MarketsParams{Limit: 10})
	require.NoError(t, err)
	// Если парсинг сломан — Unmarshal вернёт ошибку выше.
	// Дополнительно проверяем что Volume/Liquidity не отрицательные.
	for _, m := range markets {
		assert.GreaterOrEqual(t, float64(m.Volume), 0.0)
		assert.GreaterOrEqual(t, float64(m.Liquidity), 0.0)
	}
}
```

> **Замечание:** в файле нужен import `gamma "github.com/atlasdev/polytrade-bot/internal/api/gamma"`.

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/api/gamma/... -v -timeout 30s
```

Expected: все PASS. Если поле не парсится — это баг в `flexFloat64`, исправить.

**Step 3: Commit**

```bash
git add internal/api/gamma/gamma_test.go
git commit -m "test(gamma): add public API integration tests"
```

---

## Task 4: Tests for internal/api/data

**Files:**
- Create: `internal/api/data/data_test.go`

Публичные тесты, сеть нужна, auth не нужен.

**Step 1: Создать `internal/api/data/data_test.go`**

```go
package data_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/testutil"
)

// Публично известный адрес крупного трейдера Polymarket для тестов.
// Адрес взят из публичной статистики — это не секрет.
const knownTraderAddr = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

func TestGetPositions_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	positions, err := client.GetPositions(data.PositionsParams{
		User:  knownTraderAddr,
		Limit: 5,
	})
	require.NoError(t, err)
	// Трейдер может не иметь открытых позиций — это валидно.
	assert.NotNil(t, positions)
}

func TestGetPositions_EmptyAddress(t *testing.T) {
	client := testutil.NewDataClient()
	// Адрес без позиций должен вернуть пустой список, не ошибку
	positions, err := client.GetPositions(data.PositionsParams{
		User: "0x0000000000000000000000000000000000000000",
	})
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetClosedPositions_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	positions, err := client.GetClosedPositions(knownTraderAddr, 5, 0)
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetTrades_KnownTrader(t *testing.T) {
	client := testutil.NewDataClient()
	trades, err := client.GetTrades(data.TradesParams{
		User:  knownTraderAddr,
		Limit: 5,
	})
	require.NoError(t, err)
	assert.NotNil(t, trades)
}

func TestGetTrades_FieldsValid(t *testing.T) {
	client := testutil.NewDataClient()
	trades, err := client.GetTrades(data.TradesParams{
		User:  knownTraderAddr,
		Limit: 3,
	})
	require.NoError(t, err)
	for _, tr := range trades {
		assert.NotEmpty(t, tr.ID, "trade ID should not be empty")
		assert.NotEmpty(t, tr.AssetID, "asset ID should not be empty")
	}
}
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/api/data/... -v -timeout 30s
```

Expected: все PASS.

**Step 3: Commit**

```bash
git add internal/api/data/data_test.go
git commit -m "test(data): add public Data API integration tests"
```

---

## Task 5: Tests for internal/api/clob — Public Endpoints

**Files:**
- Create: `internal/api/clob/clob_public_test.go`

Публичные тесты для CLOB. Сеть нужна, auth не нужен.

**Step 1: Создать `internal/api/clob/clob_public_test.go`**

```go
package clob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/testutil"
)

func TestGetMarkets_FirstPage(t *testing.T) {
	client := testutil.NewCLOBClient()
	resp, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Data, "first page should have markets")
	assert.NotEmpty(t, resp.Data[0].ConditionID)
}

func TestGetMarkets_Pagination(t *testing.T) {
	client := testutil.NewCLOBClient()

	page1, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page1.NextCursor)

	page2, err := client.GetMarkets(page1.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, page2)
	// Страницы должны содержать разные рынки
	assert.NotEqual(t, page1.Data[0].ConditionID, page2.Data[0].ConditionID)
}

func TestGetMarket_ByConditionID(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data)

	condID := page.Data[0].ConditionID
	m, err := client.GetMarket(condID)
	require.NoError(t, err)
	assert.Equal(t, condID, m.ConditionID)
}

func TestGetOrderBook_ReturnsBook(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data)
	require.NotEmpty(t, page.Data[0].Tokens)

	tokenID := page.Data[0].Tokens[0].TokenID
	ob, err := client.GetOrderBook(tokenID)
	require.NoError(t, err)
	assert.Equal(t, tokenID, ob.AssetID)
	assert.NotNil(t, ob.Bids)
	assert.NotNil(t, ob.Asks)
}

func TestGetMidpoint(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data[0].Tokens)

	tokenID := page.Data[0].Tokens[0].TokenID
	mid, err := client.GetMidpoint(tokenID)
	require.NoError(t, err)
	assert.NotEmpty(t, mid.Mid)
}

func TestGetPrice_BuyAndSell(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data[0].Tokens)
	tokenID := page.Data[0].Tokens[0].TokenID

	buy, err := client.GetPrice(tokenID, "BUY")
	require.NoError(t, err)
	assert.NotEmpty(t, buy.Price)

	sell, err := client.GetPrice(tokenID, "SELL")
	require.NoError(t, err)
	assert.NotEmpty(t, sell.Price)
}

func TestGetSpread(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data[0].Tokens)
	tokenID := page.Data[0].Tokens[0].TokenID

	spread, err := client.GetSpread(tokenID)
	require.NoError(t, err)
	assert.NotEmpty(t, spread.Spread)
}

func TestGetMarketTrades_PublicEndpoint(t *testing.T) {
	client := testutil.NewCLOBClient()
	page, err := client.GetMarkets("")
	require.NoError(t, err)
	require.NotEmpty(t, page.Data[0].Tokens)
	tokenID := page.Data[0].Tokens[0].TokenID

	trades, err := client.GetMarketTrades(tokenID, 5)
	require.NoError(t, err)
	assert.NotNil(t, trades)
}
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/api/clob/... -v -run "TestGetMarkets|TestGetMarket|TestGetOrderBook|TestGetMidpoint|TestGetPrice|TestGetSpread|TestGetMarket" -timeout 30s
```

Expected: все PASS.

**Step 3: Commit**

```bash
git add internal/api/clob/clob_public_test.go
git commit -m "test(clob): add public endpoint integration tests"
```

---

## Task 6: Tests for internal/api/clob — Private (Integration)

**Files:**
- Create: `internal/api/clob/clob_integration_test.go`

Требует `POLY_PRIVATE_KEY`. Build-тег `integration`.

**Step 1: Создать `internal/api/clob/clob_integration_test.go`**

```go
//go:build integration

package clob_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/testutil"
)

func TestDeriveAPIKey_FromPrivateKey(t *testing.T) {
	l1 := testutil.LoadL1Signer(t)
	client := testutil.NewCLOBClient()

	creds, err := client.DeriveAPIKey(l1)
	require.NoError(t, err)
	assert.NotEmpty(t, creds.APIKey)
	assert.NotEmpty(t, creds.APISecret)
	assert.NotEmpty(t, creds.Passphrase)
	assert.Equal(t, l1.Address(), creds.Address)
}

func TestGetOrders_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)

	resp, err := client.GetOrders()
	require.NoError(t, err)
	// Может быть пустым — это валидно
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Data)
}

func TestGetPositions_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)

	positions, err := client.GetPositions()
	require.NoError(t, err)
	assert.NotNil(t, positions)
}

func TestGetTrades_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)

	resp, err := client.GetTrades()
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Data)
}

func TestGetBalanceAllowance_USDC(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)

	bal, err := client.GetBalanceAllowance("COLLATERAL", "")
	require.NoError(t, err)
	assert.NotNil(t, bal)
	assert.Equal(t, "COLLATERAL", bal.AssetType)
	assert.NotEmpty(t, bal.Balance)
}

func TestGetDataOrders_Authenticated(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)
	client := testutil.NewAuthCLOBClient(creds)

	resp, err := client.GetDataOrders(clob.OrdersFilter{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}
```

> **Замечание:** добавить `"github.com/atlasdev/polytrade-bot/internal/api/clob"` в imports для `clob.OrdersFilter`.

**Step 2: Запустить (нужен POLY_PRIVATE_KEY)**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
POLY_PRIVATE_KEY=<ваш_ключ> go test ./internal/api/clob/... -v -tags=integration -timeout 60s
```

Expected: все PASS. Если `DeriveAPIKey` возвращает 404 — значит у адреса нет ключа: создаём через `POST /auth/api-key`. Тогда добавить метод `CreateAPIKey` аналогично `DeriveAPIKey`.

**Step 3: Commit**

```bash
git add internal/api/clob/clob_integration_test.go
git commit -m "test(clob): add authenticated integration tests"
```

---

## Task 7: Tests for internal/i18n

**Files:**
- Create: `internal/i18n/i18n_test.go`

Юнит тесты, сеть не нужна, быстрые.

**Step 1: Создать `internal/i18n/i18n_test.go`**

```go
package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

func TestT_ReturnsNonNil(t *testing.T) {
	loc := i18n.T()
	require.NotNil(t, loc)
}

func TestT_DefaultEnglish(t *testing.T) {
	i18n.SetLanguage("en")
	loc := i18n.T()
	assert.NotEmpty(t, loc.TabOverview)
	assert.NotEmpty(t, loc.TabOrders)
	assert.NotEmpty(t, loc.TabPositions)
}

func TestSetLanguage_AllLocales(t *testing.T) {
	langs := i18n.Available()
	require.Equal(t, []string{"en", "ru", "zh", "ja", "ko"}, langs)

	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			i18n.SetLanguage(lang)
			loc := i18n.T()
			require.NotNil(t, loc)
			assert.NotEmpty(t, loc.TabOverview, "TabOverview should not be empty for lang %s", lang)
			assert.NotEmpty(t, loc.TabOrders, "TabOrders should not be empty for lang %s", lang)
		})
	}
}

func TestSetLanguage_UnknownFallsBackToEnglish(t *testing.T) {
	i18n.SetLanguage("en")
	engTab := i18n.T().TabOverview

	i18n.SetLanguage("xyz-unknown")
	loc := i18n.T()
	require.NotNil(t, loc)
	assert.Equal(t, engTab, loc.TabOverview, "unknown language should fall back to English")
}

func TestSetLanguage_ThreadSafe(t *testing.T) {
	// Одновременное переключение языков не должно паниковать
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			i18n.SetLanguage("ru")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		i18n.SetLanguage("en")
		_ = i18n.T()
	}
	<-done
}

func TestAvailable_Returns5Languages(t *testing.T) {
	langs := i18n.Available()
	assert.Len(t, langs, 5)
}
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/i18n/... -v
```

Expected: все PASS. Если поле пустое в каком-то locale — это баг в JSON файле, добавить недостающее поле.

**Step 3: Commit**

```bash
git add internal/i18n/i18n_test.go
git commit -m "test(i18n): add locale loading and SetLanguage tests"
```

---

## Task 8: Tests for internal/trading (Engine)

**Files:**
- Create: `internal/trading/engine_test.go`

Юнит тесты, сеть не нужна. Используем fake Strategy.

**Step 1: Создать `internal/trading/engine_test.go`**

```go
package trading_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/trading"
)

// fakeStrategy — тестовая реализация Strategy.
type fakeStrategy struct {
	name    string
	started atomic.Bool
	stopped atomic.Bool
	err     error
}

func (f *fakeStrategy) Name() string { return f.name }

func (f *fakeStrategy) Start(ctx context.Context) error {
	f.started.Store(true)
	<-ctx.Done() // ждём отмены контекста
	return f.err
}

func (f *fakeStrategy) Stop() error {
	f.stopped.Store(true)
	return nil
}

func TestEngine_NoStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	err := engine.Start(context.Background())
	// Без стратегий должен вернуть nil без ошибки
	assert.NoError(t, err)
}

func TestEngine_StartsAndStops(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s := &fakeStrategy{name: "test-strategy"}
	engine.Register(s)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := engine.Start(ctx)
	// context.DeadlineExceeded не пробрасывается как ошибка (ctx.Err() != nil)
	assert.NoError(t, err)
	assert.True(t, s.started.Load(), "strategy should have been started")
}

func TestEngine_MultipleStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s1 := &fakeStrategy{name: "strategy-1"}
	s2 := &fakeStrategy{name: "strategy-2"}
	engine.Register(s1)
	engine.Register(s2)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := engine.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, s1.started.Load())
	assert.True(t, s2.started.Load())
}

func TestEngine_StrategyError_Propagates(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	expectedErr := errors.New("strategy boom")
	s := &fakeStrategy{name: "err-strategy", err: expectedErr}
	engine.Register(s)

	// Стратегия с ошибкой которая стартует сразу
	errStrategy := &immediateErrStrategy{err: expectedErr}
	engine2 := trading.NewEngine(zerolog.Nop())
	engine2.Register(errStrategy)

	err := engine2.Start(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestEngine_Stop_CallsStopOnStrategies(t *testing.T) {
	engine := trading.NewEngine(zerolog.Nop())
	s := &fakeStrategy{name: "stop-test"}
	engine.Register(s)

	engine.Stop()
	assert.True(t, s.stopped.Load())
}

// immediateErrStrategy — стратегия которая сразу возвращает ошибку.
type immediateErrStrategy struct{ err error }

func (i *immediateErrStrategy) Name() string           { return "immediate-err" }
func (i *immediateErrStrategy) Start(_ context.Context) error { return i.err }
func (i *immediateErrStrategy) Stop() error            { return nil }
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/trading/... -v
```

Expected: все PASS.

**Step 3: Commit**

```bash
git add internal/trading/engine_test.go
git commit -m "test(trading): add Engine unit tests with fake strategies"
```

---

## Task 9: Tests for internal/tui (EventBus)

**Files:**
- Create: `internal/tui/eventbus_test.go`

Юнит тесты, быстрые.

**Step 1: Создать `internal/tui/eventbus_test.go`**

```go
package tui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestEventBus_SendAndReceive(t *testing.T) {
	bus := tui.NewEventBus()
	msg := tui.BotEventMsg{Level: "info", Message: "hello"}

	bus.Send(msg)

	// WaitForEvent возвращает tea.Cmd, вызываем её
	cmd := bus.WaitForEvent()
	received := cmd()
	assert.Equal(t, msg, received)
}

func TestEventBus_Tap_ReceivesCopy(t *testing.T) {
	bus := tui.NewEventBus()
	tap := bus.Tap()

	msg := tui.BotEventMsg{Level: "warn", Message: "tap test"}
	bus.Send(msg)

	select {
	case received := <-tap:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap did not receive message within timeout")
	}
}

func TestEventBus_MultipleTaps(t *testing.T) {
	bus := tui.NewEventBus()
	tap1 := bus.Tap()
	tap2 := bus.Tap()

	msg := tui.BalanceMsg{USDC: 100.5}
	bus.Send(msg)

	for _, tap := range []<-chan interface{}{} {
		_ = tap
	}

	select {
	case got := <-tap1:
		assert.Equal(t, msg, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap1 did not receive")
	}
	select {
	case got := <-tap2:
		assert.Equal(t, msg, got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tap2 did not receive")
	}
}

func TestEventBus_Send_NonBlocking_WhenFull(t *testing.T) {
	bus := tui.NewEventBus()
	// Заполняем буфер (512 элементов) — следующий Send не должен блокировать
	for i := 0; i < 512; i++ {
		bus.Send(tui.BotEventMsg{Message: "fill"})
	}
	// Должен вернуться сразу без блокировки
	done := make(chan struct{})
	go func() {
		bus.Send(tui.BotEventMsg{Message: "overflow"})
		close(done)
	}()
	select {
	case <-done:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send blocked on full buffer")
	}
}

func TestEventBus_DifferentMessageTypes(t *testing.T) {
	bus := tui.NewEventBus()
	tap := bus.Tap()

	msgs := []interface{}{
		tui.BotEventMsg{Level: "info", Message: "test"},
		tui.BalanceMsg{USDC: 50.0},
		tui.SubsystemStatusMsg{Name: "monitor", Active: true},
		tui.LanguageChangedMsg{},
	}

	for _, m := range msgs {
		bus.Send(m)
	}

	for _, expected := range msgs {
		select {
		case got := <-tap:
			assert.Equal(t, expected, got)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("did not receive %T", expected)
		}
	}
}
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./internal/tui/... -v -run "TestEventBus"
```

Expected: все PASS.

**Step 3: Commit**

```bash
git add internal/tui/eventbus_test.go
git commit -m "test(tui): add EventBus unit tests"
```

---

## Task 10: Tests for internal/monitor (Integration)

**Files:**
- Create: `internal/monitor/trades_test.go`

Integration тест для TradesMonitor — один poll-цикл.

**Step 1: Создать `internal/monitor/trades_test.go`**

```go
//go:build integration

package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/api"
	"github.com/atlasdev/polytrade-bot/internal/api/clob"
	"github.com/atlasdev/polytrade-bot/internal/api/data"
	"github.com/atlasdev/polytrade-bot/internal/config"
	"github.com/atlasdev/polytrade-bot/internal/monitor"
	"github.com/atlasdev/polytrade-bot/internal/notify"
	"github.com/atlasdev/polytrade-bot/internal/testutil"
	"github.com/atlasdev/polytrade-bot/internal/tui"
)

func TestTradesMonitor_SinglePoll(t *testing.T) {
	_, creds := testutil.LoadL2Creds(t)

	clobHTTP := api.NewClient(testutil.ClobURL, 10, 1)
	dataHTTP := api.NewClient(testutil.DataURL, 10, 1)

	clobClient := clob.NewClient(clobHTTP, creds)
	dataClient := data.NewClient(dataHTTP)

	cfg := &config.TradesMonitorConfig{
		PollIntervalMs: 5000,
		TrackPositions: true,
		TradesLimit:    10,
	}

	tm := monitor.NewTradesMonitor(clobClient, dataClient, notify.NewNoopNotifier(), cfg, zerolog.Nop())

	bus := tui.NewEventBus()
	tap := bus.Tap()
	tm.SetBus(bus)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Запускаем monitor в горутине, ждём первый poll
	go tm.Run(ctx)

	// Ждём OrdersUpdateMsg или PositionsUpdateMsg
	gotUpdate := false
	deadline := time.After(12 * time.Second)
	for !gotUpdate {
		select {
		case msg := <-tap:
			switch msg.(type) {
			case tui.OrdersUpdateMsg:
				gotUpdate = true
				t.Log("received OrdersUpdateMsg")
			case tui.PositionsUpdateMsg:
				gotUpdate = true
				t.Log("received PositionsUpdateMsg")
			}
		case <-deadline:
			t.Fatal("did not receive any update from TradesMonitor within timeout")
		}
	}

	assert.True(t, gotUpdate)
}
```

**Step 2: Запустить**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
POLY_PRIVATE_KEY=<ваш_ключ> go test ./internal/monitor/... -v -tags=integration -timeout 30s
```

Expected: PASS. Если `NewTradesMonitor` или `SetBus` имеют другую сигнатуру — исправить вызов в тесте под реальный API.

**Step 3: Commit**

```bash
git add internal/monitor/trades_test.go
git commit -m "test(monitor): add TradesMonitor integration test"
```

---

## Task 11: Run full suite + fix bugs

**Step 1: Публичные тесты (без сети auth)**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
go test ./... -v -timeout 60s 2>&1 | tee /tmp/test_public.log
```

**Step 2: Интеграционные тесты (с POLY_PRIVATE_KEY)**

```bash
POLY_PRIVATE_KEY=<ваш_ключ> go test ./... -v -tags=integration -timeout 90s 2>&1 | tee /tmp/test_integration.log
```

**Step 3: Анализ ошибок**

Для каждого FAIL:
1. Прочитать сообщение об ошибке
2. Найти соответствующий production-код
3. Исправить баг в production-коде (не в тесте)
4. Повторить тест

**Step 4: Финальный commit**

```bash
git add -A
git commit -m "test: full test suite — all tests passing"
```

---

## Команды запуска

```bash
# Публичные тесты (без auth)
go test ./... -timeout 60s

# Только юнит тесты (без сети)
go test ./internal/auth/... ./internal/i18n/... ./internal/trading/... ./internal/tui/... -v

# Интеграционные (с POLY_PRIVATE_KEY)
POLY_PRIVATE_KEY=0x<key> go test ./... -tags=integration -timeout 90s

# Конкретный пакет с подробным выводом
go test ./internal/api/gamma/... -v -timeout 30s
```
