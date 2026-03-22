# Web UI Redesign — Design Spec
**Date**: 2026-03-22
**Status**: Approved
**Stack**: Vue 3 + Vite + Pinia + vue-router + vue-i18n v11

---

## Overview

Full redesign of all Web UI pages from horizontal-tab layout to a collapsible sidebar layout. The redesign must be **production-ready** — all buttons, forms, API calls, and state management fully implemented. Text and icons sized slightly larger than current for better readability.

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Navigation | Collapsible sidebar (220px / 56px icon-only) | More content space, standard for trading dashboards |
| Theme | Purple/Violet (existing `#7c3aed` accent) | Continuity, no migration needed |
| Font size | Base 13px → 14px, labels 9px → 10px | User feedback: increase readability |
| Icons | 28px nav icons → 32px | User feedback: slightly larger |
| Approach | Evolutionary — same components, improved layout | Minimal risk, clear implementation path |

---

## Global Shell Changes

### AppSidebar.vue (new, replaces AppTabs.vue for layout)
- Width: `220px` expanded, `56px` collapsed (icon-only mode)
- Collapse button in footer toggles state, persisted to `localStorage`
- Sections: MAIN (Overview, Markets, Orders, Positions, Wallets), AUTOMATION (Strategies, Copytrading), SYSTEM (Logs, Settings)
- Each nav item: icon (32px box) + label + optional badge (count / status color)
- Badge colors: neutral (purple), danger (red for open orders with issues), success (green for open positions)
- Footer: wallet dot (animated pulse), truncated address, balance, collapse button

### AppHeader.vue (updated)
- Height: 54px
- Left: current page title + separator + subtitle
- Right: contextual status pill (LIVE / ENGINE ON / STREAMING), action buttons, alert chip

### Topbar is per-page, not global — each view renders its own topbar content via slot or props.

### Sidebar state
- Stored in Pinia `useUiStore()`: `sidebarCollapsed: boolean`
- Persisted to `localStorage` on toggle

---

## Pages

### 1. Overview (`/`)
**Data sources**: `GET /api/overview`, `GET /api/health`, Pinia `walletsMap`

**Sections**:
- KPI row (5 cards): Balance, Session P&L, Open Orders, Positions, Subsystems — `grid auto-fit minmax(160px,1fr)`
- Mid row (2:1): Session P&L chart (lightweight-charts, `LineSeries`) + Subsystems list
- API Health panel: 3-column grid, latency color-coded (ok/degraded/down), geoblock row
- Wallets summary: aggregate bar (Total Balance, Total P&L, Active count) + `<table>` with per-wallet rows

**New vs current**: Boot terminal animation retained on first load (2.2s then fades to dashboard). Animation lives inline in `OverviewView.vue` (existing `showWelcome` ref pattern). Sidebar renders normally during the 2.2s — animation is scoped to the content area only.

---

### 2. Markets (`/markets`)
**Data sources**: `GET /api/markets`, `GET /api/markets/tags`, `GET /api/markets/trending`, WebSocket price updates

**Sections**:
- Toolbar: TRENDING / BY CATEGORY toggle, search input, sort select, market count
- Tag filter bar (horizontal scroll, active tag highlighted)
- Sync progress bar (animated, shown during initial load)
- Cards grid (`auto-fill minmax(320px,1fr)`): each card shows category, badges (TRENDING/NEW/ALERT), question title, YES/NO probability bars with prices, footer meta (vol/liq/end date), BUY YES / BUY NO / ⋯ buttons
- Detail panel (right slide-in): full stats, orderbook (YES bids/asks), BUY YES / BUY NO / SET ALERT buttons
- Batch Action Bar (fixed bottom, appears when ≥1 market selected): count, YES/NO side toggle, size input, PLACE BATCH ORDER, CLEAR
- Dialogs: `PlaceOrderDialog`, `PriceAlertDialog`, `QuickBuyDialog` (existing, restyled)
- Pagination: prev/next + page numbers

**Fully working**:
- All API calls wired to existing `useMarketsStore()`
- BUY YES/NO opens `QuickBuyDialog` pre-filled with side
- SET ALERT opens `PriceAlertDialog`
- Batch order calls `store.placeBatchOrder(markets, side, size)`
- Detail panel orderbook populated from `GET /api/orderbook/:conditionId`

---

### 3. Orders + Positions + Trade History (`/orders`)
**Sub-tabs**: ORDERS / POSITIONS / TRADE HISTORY (URL hash or query param `?tab=`)
**Data sources**: `GET /api/orders`, `GET /api/positions`, `GET /api/trades`

**Orders tab**:
- Summary KPI (4 cards): Open Orders, Total Exposure, Filled Today, Cancelled
- Filter row: status tabs (ALL/OPEN/FILLED/CANCELLED), side filter (ALL/YES/NO), search, CANCEL ALL button
- Table: checkbox, market name (truncated), side pill, price, size, fill progress bar, remaining, status pill, created time, row actions (CANCEL / DETAIL) — visible on hover
- CANCEL triggers confirmation toast, then `DELETE /api/orders/:id`

**Positions tab**:
- Summary KPI (4 cards): Open Positions, Total Value, Unrealised P&L, Realised P&L
- Filter row: filter chips (ALL/WINNING/LOSING), side filter, search, CLOSE ALL
- Table: market, side pill, size, avg price, current price (live from WS), shares, unrealised P&L + %, mini trend sparkline (SVG), CLOSE / DETAIL row actions
- CLOSE opens `ClosePositionDialog` with quantity input and estimated return

**Trade History tab**:
- Table: time, market, side, type (BUY/SELL), price, size, fee, P&L
- Sortable columns, pagination

---

### 4. Wallets (`/wallets`)
**Data sources**: `GET /api/wallets`, `GET /api/wallets/:id/allowance`

**Sections**:
- Aggregate KPI row (4): Total Balance, Total P&L, Open Positions, Allowance status
- Wallet cards grid (`auto-fill minmax(340px,1fr)`):
  - Header: avatar (gradient initial), name + PRIMARY badge if main, address, status pill, edit/more buttons
  - Body: balance (large), session P&L, allocation bar (positions/orders/free), stats mini-grid (orders, positions, allowance status)
  - Footer: ORDERS / POSITIONS / HISTORY / DISABLE/ENABLE buttons
- Add Wallet card (dashed border placeholder) → opens `AddWalletDialog`
- USDC Allowance section: grid table, each wallet × contract (CTF Exchange, NegRisk Adapter), current allowance, APPROVE ∞ button → calls `POST /api/wallets/:id/approve`

**Fully working**:
- DISABLE/ENABLE calls `PATCH /api/wallets/:id` with `{enabled: bool}`
- APPROVE ∞ signs and broadcasts approval transaction via backend
- AddWalletDialog: private key input → `POST /api/wallets`

---

### 5. Strategies (`/strategies`)
**Data sources**: `GET /api/strategies`, `GET /api/strategies/:id/trades`

**Sections**:
- Summary KPI (4): Active/Total, Trades Today, Best P&L (name), Worst P&L (name)
- Performance Summary panel (6-col): Total P&L, Win Rate, Trades, Avg Hold, Fees, Exposure
- Strategy cards grid (`auto-fill minmax(360px,1fr)`):
  - Header: gradient icon, name, type, status pill, toggle switch
  - Body: description, metrics (P&L, trades, win rate / spread / hit rate), params chips, daily limit progress bar
  - Footer: CONFIGURE / TRADES / STOP or RESUME or START
- Toggle switch: `PATCH /api/strategies/:id` with `{enabled: bool}` — optimistic update
- CONFIGURE opens `StrategyConfigDialog` with all params as form fields
- TRADES opens filtered Orders page (`/orders?strategy=id`)

---

### 6. Copytrading (`/copytrading`)
**Data sources**: `GET /api/copytrading/traders`, `GET /api/copytrading/feed`

**Layout**: 2-column grid — left col (flex:1): trader cards list + add-trader placeholder. Right col (380px fixed): detail panel (top, updates on card click) + live feed (bottom, scrollable). On mobile (<768px) right col becomes a modal sheet.

**Trader cards**:
- Avatar, name/address, status pill (COPYING/PAUSED/STOPPED), copy wallet label
- Metrics: 30D ROI, Win Rate, Avg Size, Copy P&L
- Footer: size mode chips (FIXED/RATIO/SCALE), size input, TRADES / PAUSE or RESUME / REMOVE

**Detail panel** (right, updates on card click):
- Full stats grid
- Copy Settings form: size mode, fixed size, max daily exposure, wallet select, min/max size filters, YES-only toggle, NO-copy toggle
- SAVE CONFIG → `PATCH /api/copytrading/traders/:id`

**Live Feed**:
- Scrolling list of COPIED / SKIPPED / CLOSE events, color-coded dots
- Real-time via WebSocket `copytrading` channel

**Add Trader**: input for Polymarket address → `POST /api/copytrading/traders`

---

### 7. Logs (`/logs`)
**Data source**: WebSocket `logs` channel + `GET /api/logs?limit=500` on mount

**Layout**: log area (flex:1) + detail sidebar (320px) + stats bar (bottom)

**Toolbar**: level filter buttons (DEBUG/INFO/WARN/ERROR, toggle each), source filter chips, regex search input, PAUSE/EXPORT/CLEAR buttons

**Log lines**: timestamp | level (color) | source | message (syntax-highlighted: keys=purple, values=gold, numbers=green, errors=red)

**Click a line**: expands inline detail block + populates detail sidebar with structured fields, JSON dump, contextual action buttons (e.g. "→ GO TO WALLETS" for allowance errors)

**Stats bar**: per-level counts, FOLLOW toggle (auto-scroll to bottom), total count

**Fully working**:
- PAUSE stops WS subscription (buffer keeps last 5000 lines)
- EXPORT downloads current filtered log as `.log` file
- CLEAR wipes buffer
- FOLLOW: `scrollTop = scrollHeight` on new line when enabled
- Regex search: `new RegExp(query, 'i')` applied to full line string

---

### 8. Settings (`/settings`)
**Data source**: `GET /api/config` → flat config object; save via `POST /api/config`

**Layout**: settings left nav (200px) + scrollable form area

**Sections** (all on one page, left nav jumps to anchor):
- Auth & Keys: private_key (masked), api_key (auto), chain_id select, funder_address
- Network: per-endpoint URL inputs with live status badge + TEST button (calls `POST /api/health/test`)
- Trading Engine: enabled toggle, daily_loss_limit, max_concurrent_positions, min_order_size, tick_interval select
- Telegram: enabled toggle, bot_token (masked), chat_id, per-event toggles (fill, price alert, error)
- Logging: level select, format select, web_ui_buffer input
- Interface: language select, web_ui_port, sidebar_collapsed_default toggle, animations toggle
- Danger Zone: CANCEL ALL ORDERS, STOP ALL STRATEGIES, RESET CONFIG — each with confirmation dialog

**Unsaved changes**: tracked in `useSettingsStore()` as `isDirty`. Badge shown in topbar. SAVE calls `POST /api/config` then `toast.success("Saved")`. RESET reverts to server state.

---

## New Backend API Endpoints Needed

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/config` | GET / POST | Read/write config.toml |
| `/api/health/test` | POST `{url}` | Test single endpoint reachability |
| `/api/orderbook/:conditionId` | GET | Orderbook for detail panel |
| `/api/strategies/:id/trades` | GET | Trades for a specific strategy |
| `/api/copytrading/traders` | GET / POST | List / add traders |
| `/api/copytrading/traders/:id` | PATCH / DELETE | Update / remove trader |
| `/api/copytrading/feed` | WS | Live copy events stream |
| `/api/wallets/:id/approve` | POST | Trigger USDC approval tx |
| `/api/orders/batch` | POST | Place batch orders |

> **Note**: The table above lists only **new** endpoints required by the redesign. All other endpoints referenced in page sections (`/api/overview`, `/api/health`, `/api/markets`, `/api/orders`, `/api/positions`, `/api/trades`, `/api/wallets`, `/api/strategies`) are assumed to already exist in the Go backend.

---

## Component Architecture

```
src/
  components/
    layout/
      AppSidebar.vue          # new — collapsible sidebar
      AppHeader.vue           # updated — topbar with slots
    markets/
      MarketCard.vue          # updated — larger text, new badges
      MarketDetailPanel.vue   # updated — orderbook section
      BatchActionBar.vue      # new — floating batch bar
      ... (existing restyled)
    orders/
      OrdersTable.vue         # new
      PositionsTable.vue      # new
      TradeHistoryTable.vue   # new
      ClosePositionDialog.vue # new
    wallets/
      WalletCard.vue          # new
      AddWalletDialog.vue     # new
      AllowanceTable.vue      # new
    strategies/
      StrategyCard.vue        # new
      StrategyConfigDialog.vue # new
    copytrading/
      TraderCard.vue          # new
      CopySettingsPanel.vue   # new
      CopyFeed.vue            # new
    logs/
      LogLine.vue             # new
      LogDetailSidebar.vue    # new
    settings/
      SettingsNav.vue         # new
      SettingRow.vue          # new — label+desc+control pattern
  views/
    OverviewView.vue          # updated
    MarketsView.vue           # updated
    OrdersView.vue            # updated — includes Positions + History tabs
    WalletsView.vue           # updated
    StrategiesView.vue        # updated
    CopytradingView.vue       # updated
    LogsView.vue              # updated
    SettingsView.vue          # updated
  stores/
    ui.js                     # new — sidebarCollapsed, activeTab
    settings.js               # new — config state, isDirty
    logs.js                   # new — log buffer, filters
  assets/
    theme.css                 # updated — base font-size 14px, larger icons
```

---

## Responsive Breakpoints

| Breakpoint | Behavior |
|---|---|
| < 768px | Sidebar auto-collapses to icon-only, cards grid 1 col |
| 768–1200px | Sidebar expanded, cards grid 2 col, detail panels as modals |
| > 1200px | Full layout as designed |

---

## Typography Scale (updated from current)

| Element | Old | New |
|---|---|---|
| Base body | 13px | 14px |
| Small labels | 9px | 10px |
| Tiny labels | 9px | 10px |
| KPI values | 18px | 20px |
| Nav icons | 14px | 16px |
| Nav icon box | 28×28 | 32×32 |
| Table cells | 11px | 12px |

---

## Implementation Notes

- All existing API composables (`useApi`, `useWebSocket`) remain unchanged
- Existing stores (`useMarketsStore`, `useAppStore`, `useHealthStore`) remain, new stores added
- i18n keys added to all 5 locale files for new UI strings
- `theme.css` updated: `--font-size-base: 14px`, `--nav-icon-size: 32px`
- No breaking changes to Go backend except new endpoints listed above
- `AppTabs.vue` retained but no longer used in main layout (kept for potential reuse)
