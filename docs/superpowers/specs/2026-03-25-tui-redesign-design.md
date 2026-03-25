# TUI Redesign — Design Spec

**Date:** 2026-03-25
**Status:** Approved
**Scope:** Full visual redesign of `internal/tui/` — zero logic changes, pure chrome/layout overhaul.

---

## 1. Goals

- Replace `BorderSpec` (╔║╝) panels with Card + Accent Top Bar style
- Rework Overview to lead with 4 hero KPI cards
- Keep all existing message types, state, Nexus wiring, and business logic intact
- Stay in the `internal/tui/` flat package (no sub-packages — avoids import cycles)
- Follow the tui-component-design guideline: one concern per file, helper methods extracted, pure Views

---

## 2. Design Decisions (ratified)

| Decision | Choice | Rationale |
|---|---|---|
| Navigation chrome | Horizontal top bar (keep) | All 8 tabs always visible; familiar; no width cost |
| Panel / border style | Card + accent top bar | Cleaner hierarchy; active state via colored top border |
| Overview layout | 4 hero KPI cards + 2-col detail + wallets table | High-value numbers lead; subsystems/health secondary |
| Color palette | Deep Violet (keep) | Existing palette is strong; no migration cost |

---

## 3. Color Palette (unchanged)

```
ColorBg        #0e0b1a   — page background
ColorBgMid     #13102a   — card surface
ColorBgLight   #1a1535   — active row / tab bar
ColorPrimary   #7c3aed   — accent top border (focused card), active tab bg
ColorPrimary2  #4A1580   — accent top border (non-focused card)
ColorBright    #a78bfa   — values, titles, active tab text
ColorSuccess   #34d399   — positive PnL, active status, ON badge
ColorWarning   #fbbf24   — degraded, partial, skip
ColorDanger    #f87171   — error, negative PnL, ERR log
ColorText      #e0e0e0   — body text
ColorFgDim     #888888   — secondary labels, muted headers
ColorMuted     #888888   — dim labels, muted metadata (matches styles.go; #555555 = ColorBorder/ColorPrimaryDim)
```

---

## 4. Component: Card

The fundamental layout building block replacing `BorderSpec` panels.

**Visual anatomy:**
```
┌─ top accent border (2px, ColorPrimary or ColorPrimary2) ──────┐
│ background: ColorBgMid (#13102a)                               │
│ padding: 7px 9px                                               │
│                                                                │
│  ▸ SECTION TITLE   ← ColorBright/ColorFgDim, 9px, letter-2px  │
│  ─────────────────────────────── ← border-bottom ColorBgLight │
│  key          value               ← flex justify-between       │
│  key          value                                            │
└────────────────────────────────────────────────────────────────┘
```

**Focused vs unfocused:**
- Focused card: `border-top-color = ColorPrimary (#7c3aed)`, title = `ColorBright`
- Unfocused card: `border-top-color = ColorPrimary2 (#4A1580)`, title = `ColorFgDim`

**Implementation — `renderCard()` in `tabs.go`:**

> lipgloss v0.x (used here) does not support per-side border colors. The accent top bar is rendered as a separate line and joined vertically.

```go
func renderCard(title, body string, width int, focused bool) string {
    accentColor := ColorPrimary2
    titleColor  := ColorFgDim
    if focused {
        accentColor = ColorPrimary
        titleColor  = ColorBright
    }
    topBar  := lipgloss.NewStyle().Foreground(accentColor).Render(strings.Repeat("─", width))
    heading := lipgloss.NewStyle().Foreground(titleColor).Bold(true).
                   Render("▸ " + strings.ToUpper(title))
    sep     := lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", width-2))
    content := lipgloss.NewStyle().Background(ColorBgMid).Width(width).Padding(0,1).
                   Render(heading + "\n" + sep + "\n" + body)
    return lipgloss.JoinVertical(lipgloss.Left, topBar, content)
}
```

---

## 5. Tab: Overview (redesigned)

### Layout (standard ≥101 cols)

```
┌─ hero row (4 equal cards) ──────────────────────────────────────────┐
│  BALANCE       PNL TODAY      OPEN ORDERS     COPY TRADERS          │
│  $12,450       +$234          3               5                     │
│  USDC.e        +1.91%         2 positions     3 active              │
└─────────────────────────────────────────────────────────────────────┘
┌─ SUBSYSTEMS (half) ──────┐  ┌─ API HEALTH (half) ────────────────┐
│  ● Trading   ACTIVE      │  │  ● CLOB      42ms                  │
│  ● Monitor   ACTIVE      │  │  ● Gamma     78ms                  │
│  ○ WebSocket OFF         │  │  ◐ WebSocket degraded              │
│  ● Markets   ACTIVE      │  │  ● Data API  55ms                  │
└──────────────────────────┘  └────────────────────────────────────┘
┌─ WALLETS ── Total $12,450  PnL +$234  Active 2/3 ─────────────────┐
│  LABEL           BALANCE       P&L         STATUS                 │
│  ─────────────────────────────────────────────────               │
│  Main Wallet     $8,200.00     +$180.00    ● ON                   │
│  Strategy        $4,250.00     +$54.50     ● ON                   │
│  Reserve         $0.00         $0.00       ○ OFF                  │
└────────────────────────────────────────────────────────────────────┘
[help bar]
```

### Hero KPI cards
- 4 equal-width cards in a row
- Each: label (9px, muted, letter-spaced), large value (~18px bold), subtitle (9px, muted)
- Border top colors: Balance=ColorPrimary, PnL=ColorSuccess, Orders/Traders=ColorPrimary2
- PnL value color: ColorSuccess if ≥0, ColorDanger if <0

### Responsive breakpoints

**`tiny` (≤80):** single column plain text, no cards — same as current.

**`mobile` (≤100):** one stacked stats card, same as current.

**`standard` (≤140):** 2 hero cards (Balance + PnL only) + full subsystems/health row + wallets table.
```
┌─ BALANCE ──────┐  ┌─ PNL TODAY ────┐
│   $12,450      │  │   +$234        │
│   USDC.e       │  │   +1.91%       │
└────────────────┘  └────────────────┘
┌─ SUBSYSTEMS ───────┐  ┌─ API HEALTH ────────────┐
│  ● Trading ACTIVE  │  │  ● CLOB  42ms           │
│  ● Monitor ACTIVE  │  │  ● Gamma 78ms           │
└────────────────────┘  └─────────────────────────┘
[wallets table full width]
```

**`large`/`xl` (>140):** full 4-hero layout, as diagrammed above.

---

## 6. Tab: Trading

Two sub-tabs: **Orders** (`o`) and **Positions** (`p`).

```
[o Orders (3)]  [p Positions (2)]
┌─ OPEN ORDERS ──────────────────────────────────────────────────────┐
│  MARKET          SIDE   PRICE   SIZE   FILLED  STATUS   AGE  ID   │
│  Will Trump win? BUY    0.6200  $500   $0      LIVE     2m   a1…  │← selected
│  ETH >4000 Dec?  SELL   0.3800  $200   $0      LIVE     8m   d4…  │
│  Fed rate cut Q1 BUY    0.4500  $300   $150    PARTIAL  15m  g7…  │
└────────────────────────────────────────────────────────────────────┘
[help bar: ↑↓=select  D=cancel  A=cancel-all  o=orders  p=positions]
```

- Selected row: `background=ColorBgLight`, first column `ColorBright`
- SIDE: BUY=ColorSuccess, SELL=ColorDanger
- STATUS: LIVE=ColorBright, PARTIAL=ColorWarning, FILLED=ColorSuccess
- Sub-tab bar: use existing `StyleSubTabActive` / `StyleSubTabInactive` from `styles.go` — they already match this spec. Do not create new styles for sub-tabs.

---

## 7. Tab: Strategies

```
┌─ STRATEGIES ── 3 loaded · 2 running ──────────────────────────────┐
│  NAME        STATUS    WALLET        SIGNAL    ORDERS  PNL        │
│  momentum    RUNNING   Main Wallet   BUY 0.62  2       +$180.00   │← selected
│  mean_revert RUNNING   Strategy      —         1       +$54.50    │
│  arbitrage   STOPPED   —             —         0       $0.00      │
└────────────────────────────────────────────────────────────────────┘
┌─ momentum DETAIL ────────┐  ┌─ RECENT SIGNALS ─────────────────┐
│  Min Edge   0.015        │  │  12:34 BUY Will Trump win? 0.62  │
│  Max Pos    $500         │  │  12:28 SELL ETH >4000?     0.38  │
│  Stop Loss  5%           │  │  12:15 BUY Fed cut Q1?     0.45  │
│  Take Profit 12%         │  └──────────────────────────────────┘
└──────────────────────────┘
[help bar: ↑↓=select  Enter=start/stop  w=cycle-wallet]
```

- STATUS badges: RUNNING=ColorSuccess bg, STOPPED=ColorMuted bg
- Detail card renders only when a row is selected

---

## 8. Tab: Wallets

```
┌─ WALLETS ── 3 configured · 2 active ──────────────────────────────┐
│  LABEL           ADDRESS       BALANCE    P&L       ALLOW  STATUS │
│  Main Wallet     0xaBc1…d4E5   $8,200     +$180     6/6    ON     │← selected
│  Strategy Wallet 0xF9a2…3c77   $4,250     +$54      4/6    ON     │
│  Reserve         0x12b9…aF01   $0.00      $0.00     —      OFF    │
└────────────────────────────────────────────────────────────────────┘
┌─ Main Wallet ALLOWANCES ─┐  ┌─ ACTIONS ────────────────────────┐
│  ● CTF Exchange          │  │  a — Add wallet                  │
│  ● Neg Risk Exchange     │  │  e — Edit label                  │
│  ● CTF ExchangeProxy     │  │  space — Toggle on/off           │
│  ● Neg Risk ExchangeProxy│  │  D — Delete wallet               │
│  ● USDC Merge Wrapper    │  │  r — Refresh balances            │
│  ● CTF Merge Wrapper     │  └──────────────────────────────────┘
└──────────────────────────┘
```

- Allowances: full 6/6 = ColorSuccess dot; partial = ColorWarning dot + count colored
- ON/OFF badges: ON=ColorSuccess bg, OFF=ColorMuted bg

---

## 9. Tab: Copytrading

Two sub-tabs: **Traders** (`t`) and **Live Feed** (`l`).

```
[t Traders (5)]  [l Live Feed]
┌─ TRACKED TRADERS ── 5 configured · 3 active ──────────────────────┐
│  LABEL        ADDRESS       COPY SIZE  MIN EDGE  STATUS           │
│  whale_01     0xDead…Beef   $100       0.02      ON               │← selected
│  alpha_trader 0xCafe…1337   $50        0.01      ON               │
│  poly_pro     0xBabe…F00D   $200       0.03      ON               │
│  inactive_one 0xDead…0001   $50        0.02      OFF              │
└────────────────────────────────────────────────────────────────────┘
┌─ RECENT COPY TRADES ─────────────────────────────────────────────┐
│  12:34 whale_01 BUY  Will Trump win? 0.62 $100  [copied]         │
│  12:28 alpha    SELL ETH >4000?      0.38 $50   [copied]         │
│  12:15 poly_pro BUY  Fed cut Q1?     0.45 $200  [skipped: edge]  │
└───────────────────────────────────────────────────────────────────┘
```

- `[copied]` badge = ColorSuccess; `[skipped]` = ColorWarning

---

## 10. Tab: Markets

Three modes drill-down: list → detail → order form.

```
[h Trending]  [c Categories]
┌─ CATEGORIES ─┐  ┌─ MARKETS — Politics  182 markets  /=search ──┐
│ Politics(182) │  │  MARKET               YES   NO    VOL    END │
│ Crypto (94)   │  │  Will Trump win?      0.62  0.38  $4.2M  Nov5│← selected
│ Sports (67)   │  │  Harris wins popular? 0.71  0.29  $1.8M  Nov5│
│ Economy (45)  │  │  GOP keeps House?     0.55  0.45  $890K  Nov5│
│ Science (23)  │  └─────────────────────────────────────────────┘
└───────────────┘
[help bar: ↑↓=select  Enter=detail  b=buy  s=sell  a=alert  /=search]
```

- YES price: ColorSuccess if ≥0.5, else neutral
- NO price: ColorDanger if YES≥0.5, else neutral
- Price alert set: `⚑` (U+2691) indicator column — single-cell safe Unicode, avoids emoji width issues
- Search mode: `/` opens inline text input in card title area

---

## 11. Tab: Logs

```
┌─ SYSTEM LOGS  filter: all  [f=freeze ↑↓=scroll c=clear] ─────────┐
│ 12:34:51 [INF] strategy="momentum" signal=BUY price=0.62          │
│ 12:34:52 [WRN] websocket reconnect attempt=2 delay=5s             │
│ 12:34:53 [INF] order placed id=a1b2c3 side=BUY size=500           │
│ 12:34:55 [ERR] gamma api error status=429 retry_after=2s          │
│ 12:34:57 [INF] markets refreshed count=482 elapsed=1.2s           │
└────────────────────────────────────────────────────────────────────┘
[help bar: ↑↓=scroll  f=freeze  c=clear  /=filter]
```

- `[INF]` = ColorBright; `[WRN]` = ColorWarning; `[ERR]` = ColorDanger; `[DBG]` = ColorFgDim
- Timestamp column = ColorMuted
- Frozen state: card title shows `[FROZEN]` badge in ColorWarning

---

## 12. Tab: Settings

Two-column: section nav (left, ~14 col) + fields (right).

```
┌─ SECTIONS ────┐  ┌─ GENERAL ──────────────────────────────────┐
│ > General      │  │  Trading enabled        [OFF]              │
│   Trading      │  │  Log level          ◀   info   ▶           │
│   Notifications│  │  ✎ WebUI port       [8080____________]     │← editing
│   Language     │  │  WebUI enabled          [ON]               │
│   Auth / Keys  │  │  Telegram enabled       [OFF]              │
└────────────────┘  │  ──────────────────────────────────────    │
                    │  Config: ./config.toml   saved             │
                    └─────────────────────────────────────────────┘
[help bar: ↑↓=field  Enter=edit  space=toggle  ←→=enum  s=save  Esc=cancel]
```

- Active editing field: `background=ColorBgLight`, label=ColorText, input box with ColorPrimary border
- Toggle ON: ColorSuccess bg + ColorBg text; OFF: ColorMuted bg
- Enum value: ColorText bold, arrows ColorBright
- Saved indicator: ColorSuccess; unsaved changes: ColorWarning "unsaved"

---

## 13. Status Bar (unchanged structure)

```
[NORMAL]               12:34:56 UTC  [● LIVE]
```

- NORMAL pill: `background=ColorPrimary, color=ColorBg`
- LIVE pill: `background=ColorSuccess, color=ColorBg`
- Middle: ColorText time on ColorBgLight

---

## 14. Toast Notifications (unchanged)

Thick border (┏┓┗┛), `background=ColorBgLight`:
- Success: `border=ColorSuccess`, prefix `✓`
- Error: `border=ColorDanger`, prefix `✗`
- Warning: `border=ColorWarning`, prefix `⚠`
- Info: `border=ColorBright`, prefix `◈`

---

## 15. Splash Screen (light touch)

Keep existing ASCII logo and structure. Update box style to match new card aesthetic:
- Box border: `BorderThick` (┏━┓) with `ColorPrimary`
- Subtitle: `ColorPrimary` bold
- Status line: spinner + dimmed text

---

## 16. File Structure (unchanged — flat package)

```
internal/tui/
├── app.go             (exists) — NO CHANGE: AppModel root, message routing
├── styles.go          (exists) — UPDATE: remove dead StyleSidebar* vars if unused; keep all others
├── tabs.go            (exists) — UPDATE: keep RenderTopBar; replace renderPanel() with renderCard();
│                                          add renderHeroCard(); remove renderPanel() entirely
├── tab_overview.go    (exists) — UPDATE: hero KPI row + 2-col detail layout
├── tab_trading.go     (exists) — UPDATE: card wrapper around tables, sub-tab bar style
├── tab_strategies.go  (exists) — UPDATE: card wrapper, detail panel
├── tab_wallets.go     (exists) — UPDATE: card wrapper, allowances panel
├── tab_copytrading.go (exists) — UPDATE: card wrapper, live feed panel
├── tab_markets.go     (exists) — UPDATE: card wrapper, category sidebar
├── tab_logs.go        (exists) — UPDATE: card wrapper, freeze indicator
├── tab_settings.go    (exists) — UPDATE: section nav card + fields card
├── splash.go          (exists) — LIGHT UPDATE: box border style only
├── messages.go        (exists) — NO CHANGE
├── messages_test.go   (exists) — NO CHANGE
├── keys.go            (exists) — NO CHANGE
├── nexus.go           (exists) — NO CHANGE
├── root.go            (exists) — NO CHANGE
├── wizard.go          (exists) — NO CHANGE
```

**Critical:** `renderPanel()` in `tabs.go` (line ~91) is replaced by `renderCard()`. All callers across `tab_overview.go` and other tabs must switch to `renderCard()`. Delete `renderPanel()` — do not leave both.

---

## 17. Key Helper Functions to Add

### `renderCard(title, body string, width int, focused bool) string`
Location: `tabs.go`
Returns a card with colored top accent bar + title + separator + body.

### `renderHeroCard(label, value, sub string, width int, topColor lipgloss.Color) string`
Location: `tabs.go`
Returns a centered KPI hero card (large bold value, small label+sub).
Width formula: `heroW := (m.width - 6) / 4` — accounts for 3 single-space gaps between the 4 cards.
For `standard` breakpoint (2 cards): `heroW := (m.width - 3) / 2`.

### Badges — use existing style vars directly
Do not add a `renderBadge()` helper. Callers use existing `styles.go` vars:
- `StyleToggleOn.Render("ON")` / `StyleToggleOff.Render("OFF")`
- `StyleSuccess.Render("RUNNING")` / `StyleMuted.Render("STOPPED")`
- `StyleSuccess.Render("[copied]")` / `StyleWarning.Render("[skipped]")`

---

## 18. Out of Scope

- No changes to message types, EventBus, Nexus wiring
- No changes to wizard.go (first-run setup)
- No new keybindings beyond existing set
- No sub-package reorganization
- No i18n string changes
