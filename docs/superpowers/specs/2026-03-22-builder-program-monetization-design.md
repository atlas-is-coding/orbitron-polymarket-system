# Polymarket BUILDER PROGRAM Monetization — Design Spec
Date: 2026-03-22
Status: Approved (v3)

## Goal

Enable and verify monetization through the official Polymarket BUILDER PROGRAM.
Ensure BuilderApiKey is correctly attributed in every production order, provide
private monitoring to confirm attribution is working, and alert on credential expiry.

---

## Scope

- **polytrade-bot** (Go) — trading bot that places orders on Polymarket
- **orbitron-polymarket-website** (TypeScript/Next.js) — license server + private dashboard

Out of scope: migration to official Builder Signing SDK or Relayer.

---

## Architecture

```
polytrade-bot (Go)
  ├── license.Load() [existing]
  │     returns *BuilderCredentials{APIKey, Secret, ExpiresAt} or nil (graceful if no token)
  ├── internal/builder/validator.go [NEW]
  │     BuilderKeyValidator — validates key at startup, called from main.go
  ├── internal/builder/logger.go [NEW]
  │     OrderExecutionLogger — wired into copytrading/executor.go
  └── internal/copytrading/executor.go [existing, minor]
        injects BuilderApiKey into CreateOrderRequest (already implemented)
        + calls OrderExecutionLogger.LogOrder() after each order submission

orbitron-polymarket-website (TypeScript/Next.js)
  ├── src/lib/license-api.ts [existing]
  │     checkRateLimit(), safeEqual() — reused by all new endpoints
  ├── src/middleware.ts [extend or create]
  │     builderAuthMiddleware — Bearer token guard for /api/v1/builder/* and /dashboard/builder
  ├── src/app/api/v1/builder/status/route.ts [NEW]
  ├── src/app/api/v1/builder/analytics/route.ts [NEW]
  ├── src/app/api/v1/builder/health/route.ts [NEW]
  └── src/app/dashboard/builder/page.tsx [NEW]

Data flow:
  polytrade-bot → POST /api/v1/analytics/report [existing] → SQLite
  orbitron-website → GET /builder-analytics [Polymarket Data API, no auth] → official data
  dashboard reads from: orbitron-website endpoints only (no direct bot connection)
```

---

## BuilderApiKey Injection (polytrade-bot)

**Field:** JSON `"builderApiKey"` (omitempty) — top-level in POST /order body, NOT in signed EIP-712 body.

```go
// internal/api/clob/models.go (existing)
type CreateOrderRequest struct {
    Order         SignedOrder `json:"order"`
    Owner         string      `json:"owner"`
    OrderType     OrderType   `json:"orderType"`
    BuilderApiKey string      `json:"builderApiKey,omitempty"`
}
```

**Injection point:** `internal/copytrading/executor.go:buildOrderRequest()` — sets `req.BuilderApiKey = e.builderAPIKey`.

**diag.go note:** `internal/diag/diag.go` creates a `CreateOrderRequest` for diagnostic/test purposes only. It intentionally omits `BuilderApiKey`. This is acceptable — diag is not a production order path. Success criterion "0 orders without key" applies to production trading paths only (copytrading executor, strategy engine).

---

## Component Interfaces

### `internal/builder/validator.go`

```go
type ValidationResult struct {
    Valid           bool
    DaysUntilExpiry int    // negative if expired
    Reason          string // empty if valid; e.g. "key empty", "expired N days ago"
}

type BuilderKeyValidator struct {
    key       string
    expiresAt time.Time
    logger    zerolog.Logger
}

func NewBuilderKeyValidator(creds *license.BuilderCredentials, log zerolog.Logger) *BuilderKeyValidator
// creds may be nil (no license token configured) — validator will report Valid=false silently

func (v *BuilderKeyValidator) Check() ValidationResult
// Validates: creds not nil, key non-empty, expiresAt in future.
// Logs: INFO if valid and >7 days remaining
//       WARN if valid and <=7 days remaining
//       ERROR if expired or key empty
// Called from main.go after license.Load()
```

**ExpiresAt source:** `license.Load()` returns `BuilderCredentials.ExpiresAt` (populated from license server response field `"expires_at"`, set to `now + 30 days` by the server). No separate env var needed.

**Nil credentials:** If `license.Load()` returns `nil, nil` (no app token embedded), builder features are disabled. `BuilderKeyValidator.Check()` returns `Valid=false, Reason="no license token configured"` — non-fatal, logged at DEBUG level.

### `internal/builder/logger.go`

```go
type OrderLogEntry struct {
    OrderID       string
    BuilderKeySet bool
    Timestamp     time.Time
    Success       bool
}

type OrderExecutionLogger struct {
    mu            sync.Mutex
    totalOrders   int64
    withKey       int64
    withoutKey    int64
    log           zerolog.Logger
}

func NewOrderExecutionLogger(log zerolog.Logger) *OrderExecutionLogger

func (l *OrderExecutionLogger) LogOrder(entry OrderLogEntry)
// Thread-safe. Increments counters. Logs per-order at DEBUG level.
// Logs summary (totals) at INFO level every 100 orders.

func (l *OrderExecutionLogger) Summary() (total, withKey, withoutKey int64)
```

**Integration:** In `internal/copytrading/executor.go`, after `createAndPostOrder()` returns:
```go
e.orderLogger.LogOrder(OrderLogEntry{
    OrderID:       resp.OrderID,
    BuilderKeySet: e.builderAPIKey != "",
    Timestamp:     time.Now(),
    Success:       resp.Success,
})
```
`OrderExecutionLogger` is constructed in `main.go` and injected into `Executor` via `WithOrderLogger()`.

---

## Authentication (Private Dashboard & APIs)

**Mechanism:** Bearer token in `Authorization` header.
**Token:** existing `APP_TOKEN` env var (same token used by `/api/v1/license` endpoint).
**Comparison:** `safeEqual()` from `src/lib/license-api.ts` (constant-time, existing).
**On invalid/missing token:** `401 Unauthorized`.
**On missing `APP_TOKEN` env:** orbitron-website logs warning; builder endpoints return `503 Service Unavailable`.
  (Unlike the license endpoint, dashboard is non-critical — server still starts without it.)

**Security:**
- Dashboard served over HTTPS only (enforced at deployment level)
- `Content-Security-Policy: default-src 'self'` header on dashboard page
- Token stored in `sessionStorage` (clears on tab close, not persisted like localStorage)

**Dashboard auth flow:**
1. User visits `/dashboard/builder`
2. If no token in `sessionStorage` → inline token input form (no separate /login page)
3. Token entered → stored in `sessionStorage` → all API calls include `Authorization: Bearer <token>`

---

## Polymarket `/builder-analytics` Endpoint

**URL:** `GET https://data-api.polymarket.com/builder-analytics`
**Source in docs:** POLYMARKET_DOCS.md section 16 (Data API endpoints, line 834)
**Authentication:** No auth required (public Data API endpoint)
**Response schema:** Not documented in our docs — **schema must be verified against live API at implementation time.** Our proxy wraps whatever Polymarket returns with `last_synced` and `stale` fields.

---

## New Endpoints (orbitron-website)

### Rate Limiting
All new endpoints reuse `checkRateLimit()` from `src/lib/license-api.ts` (existing pattern):

| Endpoint | Limit |
|----------|-------|
| `GET /api/v1/builder/status` | 60 req/min per IP |
| `GET /api/v1/builder/analytics` | 10 req/min per IP |
| `GET /api/v1/builder/health` | 5 req/min per IP |

### `GET /api/v1/builder/status`
Auth: Bearer required → 401 if invalid
Source: reads `BUILDER_API_KEY` and `BUILDER_SECRET` env vars.
Expiry: computes `expires_at` dynamically? No — the website does not know the expiry without calling the license server itself.

**Resolution:** Add `BUILDER_EXPIRES_AT` env var (ISO8601). This is set manually by the operator when credentials are issued. It mirrors the value that the license server computes (`now + 30 days`).

```typescript
// .env
BUILDER_EXPIRES_AT=2026-04-21T00:00:00Z  // set by operator at credential issuance time
```

```typescript
Response 200:
{
  "api_key_prefix": "550e8...",   // first 8 chars only
  "expires_at": "2026-04-21T00:00:00Z",
  "days_until_expiry": 30,
  "is_valid": true,               // false if expired or BUILDER_API_KEY empty
  "last_checked": "2026-03-22T10:00:00Z"
}
```

### `GET /api/v1/builder/analytics`
Auth: Bearer required
Cache: module-level `Map<string, {data, fetchedAt: number}>`, TTL = 1 hour
`?force=true` bypasses cache (used by dashboard Refresh button)

```typescript
Response 200:
{
  // fields from Polymarket response (schema TBD at implementation time)
  "last_synced": "2026-03-22T10:00:00Z",
  "stale": false    // true if Polymarket down and cached data is returned
}
Response 503: { "error": "no_cached_data" }  // only if Polymarket down AND no cache
```

### `GET /api/v1/builder/health`
Auth: Bearer required
Calls: `GET https://clob.polymarket.com/ok` — **public endpoint, no auth required**.
Purpose: checks Polymarket API connectivity only. Does NOT validate the builder key itself.
Builder key validity is derived from `days_until_expiry` in the `/status` endpoint.

```typescript
Response 200:
{
  "healthy": true,        // false if /ok returned non-200 or network error
  "polymarket_status": 200,
  "checked_at": "2026-03-22T10:00:00Z"
}
```

---

## Dashboard UI (`/dashboard/builder`)

Private Next.js page (client component with sessionStorage auth check).

Sections:
1. **Key Status Card** — `api_key_prefix`, `expires_at`, `days_until_expiry`, health badge (green/yellow/red)
2. **Volume Bar Chart** — bars for volume_24h / volume_7d / volume_total (exact fields depend on Polymarket response schema)
3. **Alerts Panel** — derived from /status and /health responses (see Alerts table below)
4. **Refresh button** — calls `/analytics?force=true` and `/health`, re-renders all cards

---

## Alerts

| Condition | Severity | Message |
|-----------|----------|---------|
| `days_until_expiry` < 7 | WARNING | "Builder key expires in N days — renew via Polymarket" |
| `is_valid = false` (expired or missing) | CRITICAL | "Builder key invalid — orders NOT attributed" |
| `healthy = false` with polymarket_status 401/403 | CRITICAL | "Builder key rejected by Polymarket — check credentials" |
| `healthy = false` network error | WARNING | "Could not verify key — Polymarket unreachable" |

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `APP_TOKEN` missing | Dashboard endpoints return 503; server still starts |
| `BUILDER_API_KEY` missing | `/status` returns `is_valid: false`; `/health` skips upstream call |
| Polymarket analytics unavailable | Return cached data + `stale: true`; if no cache, 503 |
| `license.Load()` returns nil (no token embedded) | Bot starts, BuilderKeyValidator reports Valid=false at DEBUG; `builderAPIKey` stays empty; orders placed without key |
| `license.Load()` returns error (network/auth) | Main.go logs WARN, continues startup; builder features disabled |
| Key expired | Bot continues trading; dashboard shows CRITICAL alert |
| Unauthenticated request to builder endpoints | 401 Unauthorized |

---

## Testing

**polytrade-bot (Go):**
- Unit: `BuilderKeyValidator.Check()` — valid/expired/empty/nil credentials
- Unit: `OrderExecutionLogger.LogOrder()` — counter increments, thread safety
- Unit: `CreateOrderRequest` JSON — `builderApiKey` present when set, absent when empty (omitempty)
- Integration: full order request to POST /order includes `builderApiKey`

**orbitron-website (TypeScript):**
- Unit: `builderAuthMiddleware` — valid token passes, invalid/missing returns 401
- Unit: `/builder/status` — correct expiry from env, `is_valid: false` when expired
- Unit: `/builder/analytics` — cache hit; cache miss calls Polymarket; stale flag on failure; force=true bypasses cache
- Unit: `/builder/health` — handles 200, 401, network error responses
- E2E: dashboard shows token gate → enter token → renders key status and volume chart

---

## Environment Variables

**orbitron-polymarket-website `.env`:**
```
APP_TOKEN=<secret>                         # existing — dashboard auth + license endpoint
BUILDER_API_KEY=<key>                      # existing — official Polymarket builder key
BUILDER_SECRET=<secret>                    # existing — Go bot only; website does not read this
BUILDER_EXPIRES_AT=2026-04-21T00:00:00Z   # NEW — ISO8601, set by operator at issuance
```

**polytrade-bot:** No new env vars or config changes required.

---

## Success Criteria

1. Every production order (copytrading, strategies) contains non-empty `builderApiKey`
2. `polytrade-bot` logs show `withoutKey = 0` in OrderExecutionLogger summary
3. All `/api/v1/builder/*` endpoints return 401 for unauthenticated requests
4. `/api/v1/builder/analytics` returns Polymarket data when authenticated
5. Dashboard at `/dashboard/builder` renders key status and volume chart
6. Alert fires in dashboard when `days_until_expiry < 7`
7. `/api/v1/builder/health` returns `healthy: true` when builder key is valid

---

## Out of Scope

- Migration to `@polymarket/builder-signing-sdk` or relayer
- Multi-user dashboard
- Public analytics pages
- Background cron health check job
- Persistent order metrics storage
