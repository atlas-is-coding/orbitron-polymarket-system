# TUI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `BorderSpec` (╔║╝) panels with Card + Accent Top Bar style across all 8 TUI tabs, and rework the Overview tab to lead with 4 hero KPI cards.

**Architecture:** Pure visual redesign — all message types, state, Nexus wiring, and business logic are untouched. The core change is replacing `renderPanel()` in `tabs.go` with `renderCard()` + `renderHeroCard()`, then updating each tab's `View()` method to use the new helpers. No new files are created.

**Tech Stack:** Go 1.24, `charmbracelet/bubbletea` v1.3.10, `charmbracelet/lipgloss` v1.1.0, `charmbracelet/bubbles` v1.0.0

---

## Verification commands (use after every task)

```bash
# From project root: /home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot
go build ./...
go test ./internal/tui/...
```

Both must pass with zero errors before committing.

---

## Task 1: Add `renderCard()` + `renderHeroCard()` to `tabs.go`, remove `renderPanel()`

**Files:**
- Modify: `internal/tui/tabs.go`

This is the foundation. Every subsequent task depends on `renderCard()` existing.

- [ ] **Step 1: Read `tabs.go` to find `renderPanel()` and its exact location**

  Open `internal/tui/tabs.go`. Note `renderPanel()` at line ~91 and `renderHelpPanel()` below it.

- [ ] **Step 2: Add `renderCard()` immediately after `RenderTopBar()`**

  Insert this function after `RenderTopBar()` (after line ~87), before `renderPanel()`:

  ```go
  // renderCard renders a card with a colored top accent line and titled body.
  // focused=true uses ColorPrimary accent; false uses ColorPrimary2.
  // Replaces the old renderPanel() for all tab content blocks.
  func renderCard(title, body string, width int, focused bool) string {
  	accentColor := ColorPrimary2
  	titleColor := ColorFgDim
  	if focused {
  		accentColor = ColorPrimary
  		titleColor = ColorBright
  	}
  	innerW := width - 2
  	if innerW < 1 {
  		innerW = 1
  	}
  	topBar := lipgloss.NewStyle().Foreground(accentColor).Render(strings.Repeat("─", width))
  	heading := lipgloss.NewStyle().Foreground(titleColor).Bold(true).
  		Render("▸ " + strings.ToUpper(title))
  	sep := lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", innerW))
  	content := lipgloss.NewStyle().
  		Background(ColorBgMid).
  		Width(innerW).
  		Padding(0, 1).
  		Render(heading + "\n" + sep + "\n" + body)
  	return lipgloss.JoinVertical(lipgloss.Left, topBar, content)
  }

  // renderHeroCard renders a centered KPI card with large bold value.
  // topColor sets the accent top bar color (use ColorPrimary, ColorSuccess, ColorPrimary2).
  // Width formula for 4-card row: heroW := (m.width - 6) / 4
  // Width formula for 2-card row: heroW := (m.width - 3) / 2
  func renderHeroCard(label, value, sub string, width int, topColor lipgloss.Color) string {
  	innerW := width - 2
  	if innerW < 1 {
  		innerW = 1
  	}
  	topBar := lipgloss.NewStyle().Foreground(topColor).Render(strings.Repeat("─", width))
  	lbl := lipgloss.NewStyle().Foreground(ColorMuted).
  		Width(innerW).Align(lipgloss.Center).Render(strings.ToUpper(label))
  	val := lipgloss.NewStyle().Foreground(ColorBright).Bold(true).
  		Width(innerW).Align(lipgloss.Center).Render(value)
  	subStr := lipgloss.NewStyle().Foreground(ColorMuted).
  		Width(innerW).Align(lipgloss.Center).Render(sub)
  	body := lipgloss.NewStyle().
  		Background(ColorBgMid).
  		Width(innerW).
  		Padding(1, 1).
  		Render(lbl + "\n" + val + "\n" + subStr)
  	return lipgloss.JoinVertical(lipgloss.Left, topBar, body)
  }
  ```

- [ ] **Step 3: Delete `renderPanel()` entirely**

  Remove the entire `renderPanel()` function (lines ~91–108). It will be replaced by `renderCard()` callers in each tab.

- [ ] **Step 4: Verify `tab_overview.go` is the only caller of `renderPanel()`**

  Search for `renderPanel(` in the codebase:
  ```bash
  grep -rn "renderPanel(" internal/tui/
  ```
  Expected: hits only in `tab_overview.go`. If other files use it, note them — they'll be updated in their respective tasks.

- [ ] **Step 5: Fix ALL `renderPanel()` calls in `tab_overview.go` to use `renderCard()`**

  `tab_overview.go` has calls in the `mobile`, `standard`, and `default` breakpoint branches. All must be renamed before the build passes, since `renderPanel()` is deleted in Step 3.

  ```bash
  # find every call
  grep -n "renderPanel(" internal/tui/tab_overview.go
  ```

  For every hit (across all breakpoints), change `renderPanel(` → `renderCard(`. The signatures are compatible: both take `(title, content string, width int, active bool)`. This is mechanical — just rename, no argument changes needed here. The `mobile` branch (`statsPanel := renderPanel(...)`) must also be updated.

- [ ] **Step 6: Build and test**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
  go build ./...
  go test ./internal/tui/...
  ```
  Expected: zero errors.

- [ ] **Step 7: Commit**

  ```bash
  git add internal/tui/tabs.go internal/tui/tab_overview.go
  git commit -m "feat(tui): add renderCard/renderHeroCard helpers, remove renderPanel"
  ```

---

## Task 2: Redesign `tab_overview.go` — hero KPI layout

**Files:**
- Modify: `internal/tui/tab_overview.go`

**Context:** The `OverviewModel` already holds all the data (`balance`, `pnlToday`, `openOrders`, `positions`, `traders`, `subsystems`, `health`, `wallets`). Only `View()` changes.

- [ ] **Step 1: Read the full current `View()` in `tab_overview.go`**

  Understand all 4 breakpoint branches (`tiny`, `mobile`, `standard`, default).

- [ ] **Step 2: Add `renderHeroCardColored()` to `tabs.go` first**

  This must exist before Steps 3–4 reference it. Add to `tabs.go` immediately after `renderHeroCard()`:

  ```go
  // renderHeroCardColored is like renderHeroCard but with a separate valueColor.
  // Use for PnL where the value can be ColorSuccess or ColorDanger.
  func renderHeroCardColored(label, value, sub string, width int, topColor, valueColor lipgloss.Color) string {
  	innerW := width - 2
  	if innerW < 1 {
  		innerW = 1
  	}
  	topBar := lipgloss.NewStyle().Foreground(topColor).Render(strings.Repeat("─", width))
  	lbl := lipgloss.NewStyle().Foreground(ColorMuted).
  		Width(innerW).Align(lipgloss.Center).Render(strings.ToUpper(label))
  	val := lipgloss.NewStyle().Foreground(valueColor).Bold(true).
  		Width(innerW).Align(lipgloss.Center).Render(value)
  	subStr := lipgloss.NewStyle().Foreground(ColorMuted).
  		Width(innerW).Align(lipgloss.Center).Render(sub)
  	body := lipgloss.NewStyle().
  		Background(ColorBgMid).
  		Width(innerW).
  		Padding(1, 1).
  		Render(lbl + "\n" + val + "\n" + subStr)
  	return lipgloss.JoinVertical(lipgloss.Left, topBar, body)
  }
  ```

- [ ] **Step 3: Replace the `standard` case view (≤140 cols)**

  The `standard` case currently shows a 2-col layout. Replace its `View()` return with:

  ```go
  case "standard":
  	heroW := max((m.width-3)/2, 10)

  	// Hero row: Balance + PnL
  	pnlVal := fmt.Sprintf("+$%.2f", m.pnlToday)
  	pnlColor := ColorSuccess
  	if m.pnlToday < 0 {
  		pnlVal = fmt.Sprintf("-$%.2f", -m.pnlToday)
  		pnlColor = ColorDanger
  	}
  	balCard := renderHeroCard(t.OverviewBalance, fmt.Sprintf("$%.2f", m.balance), "USDC.e", heroW, ColorPrimary)
  	pnlCard := renderHeroCardColored(t.OverviewPnLToday, pnlVal, "", heroW, ColorSuccess, pnlColor)
  	heroRow := lipgloss.JoinHorizontal(lipgloss.Top, balCard, " ", pnlCard)

  	// Detail row
  	half := (m.width - 4) / 2
  	subsPanel := renderCard(t.OverviewHealth, m.renderSubsystemsBlock(), half, false)
  	healthPanel := renderCard(t.OverviewHealth, m.renderHealthBlock(), half, false)
  	detailRow := lipgloss.JoinHorizontal(lipgloss.Top, subsPanel, " ", healthPanel)

  	// Wallets
  	walletsPanel := m.renderWalletsPanel(m.width)

  	parts := []string{" ", heroRow, " ", detailRow}
  	if walletsPanel != "" {
  		parts = append(parts, " ", walletsPanel)
  	}
  	parts = append(parts, " ", helpPanel)
  	return lipgloss.JoinVertical(lipgloss.Left, parts...)
  ```

- [ ] **Step 4: Replace the `default` (large/xl) case view (>140 cols)**

  ```go
  default:
  	heroW := max((m.width-6)/4, 10)

  	// PnL color
  	pnlVal := fmt.Sprintf("+$%.2f", m.pnlToday)
  	pnlColor := ColorSuccess
  	if m.pnlToday < 0 {
  		pnlVal = fmt.Sprintf("-$%.2f", -m.pnlToday)
  		pnlColor = ColorDanger
  	}

  	// 4 hero cards
  	balCard   := renderHeroCard(t.OverviewBalance, fmt.Sprintf("$%.2f", m.balance), "USDC.e", heroW, ColorPrimary)
  	pnlCard   := renderHeroCardColored(t.OverviewPnLToday, pnlVal, "", heroW, ColorSuccess, pnlColor)
  	ordCard   := renderHeroCard(t.OverviewOpenOrders, fmt.Sprintf("%d", m.openOrders), fmt.Sprintf("%d positions", m.positions), heroW, ColorPrimary2)
  	copyCard  := renderHeroCard(t.OverviewCopyTraders, fmt.Sprintf("%d", m.traders), "", heroW, ColorPrimary2)
  	heroRow   := lipgloss.JoinHorizontal(lipgloss.Top, balCard, " ", pnlCard, " ", ordCard, " ", copyCard)

  	// Detail row: subsystems + health side by side
  	half := (m.width - 4) / 2
  	subsPanel    := renderCard("Subsystems", m.renderSubsystemsBlock(), half, false)
  	healthPanel  := renderCard(t.OverviewHealth, m.renderHealthBlock(), half, false)
  	detailRow    := lipgloss.JoinHorizontal(lipgloss.Top, subsPanel, " ", healthPanel)

  	// Wallets full-width
  	walletsPanel := m.renderWalletsPanel(m.width)

  	parts := []string{" ", heroRow, " ", detailRow}
  	if walletsPanel != "" {
  		parts = append(parts, " ", walletsPanel)
  	}
  	parts = append(parts, " ", helpPanel)
  	return lipgloss.JoinVertical(lipgloss.Left, parts...)
  ```

- [ ] **Step 4: Extract `renderSubsystemsBlock()` helper from existing `rightContent` logic**

  The current `standard` and `default` branches both build subsystems inline. Extract to a method on `OverviewModel` (add near `renderHealthBlock()`):

  ```go
  func (m OverviewModel) renderSubsystemsBlock() string {
  	t := i18n.T()
  	var sb strings.Builder
  	for _, s := range m.subsystems {
  		dot := StyleSuccess.Render("●")
  		status := StyleSuccess.Render(t.OverviewActive)
  		if !s.Active {
  			dot = StyleMuted.Render("○")
  			status = StyleMuted.Render(t.OverviewInactive)
  		}
  		fmt.Fprintf(&sb, " %s %-16s %s\n", dot, StyleFgDim.Render(s.Name), status)
  	}
  	return sb.String()
  }
  ```

- [ ] **Step 6: Extract `renderWalletsPanel()` helper**

  The current wallets table block is 25+ lines inline. Extract to a method returning `string` (empty string if no wallets):

  ```go
  func (m OverviewModel) renderWalletsPanel(width int) string {
  	if len(m.wallets) == 0 {
  		return ""
  	}
  	t := i18n.T()
  	var totalBal, totalPnL float64
  	activeCount := 0
  	for _, w := range m.wallets {
  		totalBal += w.balance
  		totalPnL += w.pnl
  		if w.enabled {
  			activeCount++
  		}
  	}
  	subtitle := fmt.Sprintf("%s: %s  %s: %s  %s: %d/%d",
  		t.OverviewTotalBalance, StyleValue.Render(fmt.Sprintf("$%.2f", totalBal)),
  		t.OverviewTotalPnL, formatPnL(totalPnL),
  		t.OverviewActiveWallets, activeCount, len(m.wallets),
  	)

  	colW := max((width-12)/4, 12)
  	hdr := StyleFgDimBold.Render(fmt.Sprintf(" %-*s  %-14s  %-14s  %s", colW, "LABEL", "BALANCE", "P&L", "STATUS"))
  	sep := StyleMuted.Render(" " + strings.Repeat("─", width-6))
  	var sb strings.Builder
  	sb.WriteString(" " + subtitle + "\n\n")
  	sb.WriteString(hdr + "\n")
  	sb.WriteString(sep + "\n")
  	for _, w := range m.wallets {
  		lbl := w.label
  		if len(lbl) > colW {
  			lbl = lbl[:colW-1] + "…"
  		}
  		statusStr := StyleMuted.Render("○ OFF")
  		if w.enabled {
  			statusStr = StyleSuccess.Render("● ON")
  		}
  		fmt.Fprintf(&sb, " %-*s  %-14s  %-14s  %s\n", colW, lbl, fmt.Sprintf("$%.2f", w.balance), formatPnL(w.pnl), statusStr)
  	}
  	return renderCard(t.OverviewWallets, sb.String(), width, false)
  }
  ```

- [ ] **Step 7: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```
  Expected: zero errors.

- [ ] **Step 8: Commit**

  ```bash
  git add internal/tui/tabs.go internal/tui/tab_overview.go
  git commit -m "feat(tui): overview redesign — hero KPI cards, card panels"
  ```

---

## Task 3: Update `tab_trading.go` — card wrapper + sub-tab bar style

**Files:**
- Modify: `internal/tui/tab_trading.go`

**Context:** `TradingModel` has two `bubbles/table` models (`orders`, `positions`) and a `subTab` field (`SubTabOrders` / `SubTabPositions`). The table styles (`StyleTableHeader`, `StyleTableSelected`) are already correct — only the View() wrapper changes.

- [ ] **Step 1: Read the full `View()` in `tab_trading.go`**

  Understand how it currently renders the sub-tab bar and wraps each table.

- [ ] **Step 2: Update the sub-tab bar rendering**

  Replace any manual sub-tab rendering with the existing `StyleSubTabActive` / `StyleSubTabInactive` styles. Pattern:

  ```go
  func (m TradingModel) renderSubTabBar() string {
  	t := i18n.T()
  	var parts []string
  	labels := []struct {
  		id    TradingSubTab
  		label string
  	}{
  		{SubTabOrders, fmt.Sprintf("o %s (%d)", t.TabOrders, len(m.orderRows))},
  		{SubTabPositions, fmt.Sprintf("p %s (%d)", t.TabPositions, len(m.positionRows))},
  	}
  	for _, l := range labels {
  		if m.subTab == l.id {
  			parts = append(parts, StyleSubTabActive.Render(l.label))
  		} else {
  			parts = append(parts, StyleSubTabInactive.Render(l.label))
  		}
  	}
  	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
  }
  ```

- [ ] **Step 3: Wrap the table content in `renderCard()`**

  In `View()`, replace the bare table render with:

  ```go
  func (m TradingModel) View() string {
  	subTabBar := m.renderSubTabBar()
  	helpKeys := "↑↓=select | D=cancel | A=cancel-all | o=orders | p=positions | Tab=next"

  	var tableContent string
  	var cardTitle string
  	switch m.subTab {
  	case SubTabOrders:
  		cardTitle = fmt.Sprintf("Open Orders (%d)", len(m.orderRows))
  		tableContent = m.orders.View()
  	case SubTabPositions:
  		cardTitle = fmt.Sprintf("Positions (%d)", len(m.positionRows))
  		tableContent = m.positions.View()
  	}

  	card := renderCard(cardTitle, tableContent, m.width, true)
  	help := renderHelpPanel(helpKeys, m.width)
  	return lipgloss.JoinVertical(lipgloss.Left, " ", subTabBar, " ", card, help)
  }
  ```

  > Note: adjust `helpKeys` to match i18n strings if the file uses `i18n.T()` — check existing View() code.

- [ ] **Step 4: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add internal/tui/tab_trading.go
  git commit -m "feat(tui): trading tab — card wrapper, styled sub-tab bar"
  ```

---

## Task 4: Update `tab_strategies.go` — card wrapper + detail panel

**Files:**
- Modify: `internal/tui/tab_strategies.go`

**Context:** `StrategiesModel` has a `bubbles/table` (`table`) and `rows []StrategyRow`. It also has a `walletPicker bool`. The detail panel (showing per-strategy config + recent signals) is new.

- [ ] **Step 1: Read full `tab_strategies.go`**

  Note all fields in `StrategyRow` and the full current `View()`.

- [ ] **Step 2: Update `View()` to wrap the table in `renderCard()`**

  ```go
  func (m StrategiesModel) View() string {
  	t := i18n.T()
  	running := 0
  	for _, r := range m.rows {
  		if r.Status == "active" { // StrategyRow.Status values are "active"/"stopped" per messages.go
  			running++
  		}
  	}
  	subtitle := fmt.Sprintf("%d loaded · %d running", len(m.rows), running)
  	tableCard := renderCard(
  		fmt.Sprintf("Strategies — %s", subtitle),
  		m.table.View(),
  		m.width,
  		true,
  	)

  	helpKeys := "↑↓=select | Enter=start/stop | w=cycle-wallet | Tab=next"
  	help := renderHelpPanel(helpKeys, m.width)

  	// Detail row for selected strategy (if any row is selected)
  	var detailRow string
  	if sel := m.table.SelectedRow(); sel != nil {
  		half := (m.width - 4) / 2
  		detailCard  := renderCard(sel[0]+" detail", m.renderStrategyDetail(sel[0]), half, false)
  		signalsCard := renderCard("Recent Signals", m.renderRecentSignals(sel[0]), half, false)
  		detailRow = lipgloss.JoinHorizontal(lipgloss.Top, detailCard, " ", signalsCard)
  	}

  	parts := []string{" ", tableCard}
  	if detailRow != "" {
  		parts = append(parts, " ", detailRow)
  	}
  	_ = t // suppress unused if no i18n calls remain
  	parts = append(parts, help)
  	return lipgloss.JoinVertical(lipgloss.Left, parts...)
  }
  ```

- [ ] **Step 3: Add `renderStrategyDetail()` helper method**

  ```go
  func (m StrategiesModel) renderStrategyDetail(name string) string {
  	for _, r := range m.rows {
  		if r.Name == name {
  			var sb strings.Builder
  			fmt.Fprintf(&sb, " %-16s %s\n", StyleFgDim.Render("Wallet"), StyleValue.Render(r.WalletID))
  			fmt.Fprintf(&sb, " %-16s %s\n", StyleFgDim.Render("Status"), r.Status)
  			return sb.String()
  		}
  	}
  	return StyleMuted.Render("  no detail available")
  }
  ```

  > Adjust fields to match the actual `StrategyRow` struct fields — read the struct definition first.

- [ ] **Step 4: Add `renderRecentSignals()` stub**

  ```go
  func (m StrategiesModel) renderRecentSignals(_ string) string {
  	// Signals are not yet tracked per strategy in StrategiesModel.
  	// Return a placeholder — no logic change required by spec.
  	return StyleMuted.Render("  no recent signals")
  }
  ```

- [ ] **Step 5: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/tui/tab_strategies.go
  git commit -m "feat(tui): strategies tab — card wrapper, detail panel"
  ```

---

## Task 5: Update `tab_wallets.go` — card wrapper + allowances panel

**Files:**
- Modify: `internal/tui/tab_wallets.go`

**Context:** `WalletsModel` has a `bubbles/table` (`walletsTable`) and multiple modes (`walletModeTable`, `walletModeDetail`, `walletModeAddForm`, `walletModeEditForm`, `walletModeConfirmDelete`). Only the `walletModeTable` view changes (add card wrapper + allowances panel). All other modes/forms stay as-is.

- [ ] **Step 1: Read the full `View()` in `tab_wallets.go`**

  Find the `walletModeTable` branch in `View()`.

- [ ] **Step 2: Wrap the main wallet table in `renderCard()`**

  In the `walletModeTable` case, replace the bare table render with:

  ```go
  case walletModeTable:
  	active, total := 0, len(ids)
  	for _, id := range ids {
  		if m.wm.WalletEnabled(id) { active++ }
  	}
  	subtitle := fmt.Sprintf("%d configured · %d active", total, active)
  	tableCard := renderCard(
  		fmt.Sprintf("Wallets — %s", subtitle),
  		m.walletsTable.View(),
  		m.width,
  		true,
  	)

  	// Bottom row: allowances + actions (only when a wallet is selected)
  	var bottomRow string
  	if m.selectedID != "" {
  		half := (m.width - 4) / 2
  		allowCard  := renderCard(m.selectedID+" Allowances", m.renderAllowancesBlock(), half, false)
  		actionsCard := renderCard("Actions", m.renderActionsHelp(), half, false)
  		bottomRow = lipgloss.JoinHorizontal(lipgloss.Top, allowCard, " ", actionsCard)
  	}

  	help := renderHelpPanel("↑↓=select | a=add | e=edit | space=toggle | D=delete | r=refresh | Tab=next", m.width)
  	parts := []string{" ", tableCard}
  	if bottomRow != "" {
  		parts = append(parts, " ", bottomRow)
  	}
  	parts = append(parts, help)
  	return lipgloss.JoinVertical(lipgloss.Left, parts...)
  ```

- [ ] **Step 3: Add `renderAllowancesBlock()` helper**

  ```go
  func (m WalletsModel) renderAllowancesBlock() string {
  	// Allowance data is sent via WalletAddedMsg.Allowances — not stored in WalletsModel.
  	// Render a static placeholder. When allowance state tracking is added, update here.
  	return StyleMuted.Render("  select wallet to view allowances")
  }
  ```

  > If `WalletsModel` already stores allowance state, render it instead. Check the struct fields.

- [ ] **Step 4: Add `renderActionsHelp()` helper**

  ```go
  func (m WalletsModel) renderActionsHelp() string {
  	lines := []struct{ key, desc string }{
  		{"a", "Add wallet"},
  		{"e", "Edit label"},
  		{"space", "Toggle on/off"},
  		{"D", "Delete wallet"},
  		{"r", "Refresh balances"},
  	}
  	var sb strings.Builder
  	for _, l := range lines {
  		fmt.Fprintf(&sb, " %s — %s\n", StyleValue.Render(l.key), StyleFgDim.Render(l.desc))
  	}
  	return sb.String()
  }
  ```

- [ ] **Step 5: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/tui/tab_wallets.go
  git commit -m "feat(tui): wallets tab — card wrapper, allowances panel"
  ```

---

## Task 6: Update `tab_copytrading.go` — card wrapper + live feed panel

**Files:**
- Modify: `internal/tui/tab_copytrading.go`

**Context:** `CopytradingModel` has `tradersTable table.Model` and `recentTrades []string`. Multiple modes: `copyModeTable`, `copyModeAddForm`, `copyModeEditForm`, `copyModeConfirmDelete`. Only the `copyModeTable` view is redesigned.

- [ ] **Step 1: Read the full `View()` in `tab_copytrading.go`**

  Find the `copyModeTable` branch.

- [ ] **Step 2: Add sub-tab bar for Traders / Live Feed**

  ```go
  func (m CopytradingModel) renderSubTabBar() string {
  	traderCount := len(m.cfg.Copytrading.Traders)
  	tLabel := fmt.Sprintf("t Traders (%d)", traderCount)
  	lLabel := "l Live Feed"

  	// Use StyleSubTabActive / StyleSubTabInactive (from styles.go)
  	if m.showFeed {
  		return lipgloss.JoinHorizontal(lipgloss.Top,
  			StyleSubTabInactive.Render(tLabel),
  			StyleSubTabActive.Render(lLabel),
  		)
  	}
  	return lipgloss.JoinHorizontal(lipgloss.Top,
  		StyleSubTabActive.Render(tLabel),
  		StyleSubTabInactive.Render(lLabel),
  	)
  }
  ```

  > If `CopytradingModel` doesn't have a `showFeed bool` field, add one and wire `l` key in `Update()`. Check the existing Update() keybinding first.

- [ ] **Step 3: Wrap traders table in `renderCard()`**

  ```go
  case copyModeTable:
  	subBar := m.renderSubTabBar()
  	var card string
  	if !m.showFeed {
  		active := 0
  		for _, tr := range m.cfg.Copytrading.Traders {
  			if tr.Enabled { active++ }
  		}
  		subtitle := fmt.Sprintf("%d configured · %d active", len(m.cfg.Copytrading.Traders), active)
  		card = renderCard(
  			fmt.Sprintf("Tracked Traders — %s", subtitle),
  			m.tradersTable.View(),
  			m.width, true,
  		)
  	} else {
  		card = renderCard("Live Feed", m.renderFeed(), m.width, true)
  	}
  	help := renderHelpPanel("↑↓=select | a=add | e=edit | space=toggle | D=delete | t=traders | l=feed | Tab=next", m.width)
  	return lipgloss.JoinVertical(lipgloss.Left, " ", subBar, " ", card, help)
  ```

- [ ] **Step 4: Add `renderFeed()` helper**

  ```go
  func (m CopytradingModel) renderFeed() string {
  	if len(m.recentTrades) == 0 {
  		return renderEmptyState("○", "No copy trades yet", "waiting for signals…", m.width)
  	}
  	var sb strings.Builder
  	for _, line := range m.recentTrades {
  		sb.WriteString(" " + line + "\n")
  	}
  	return sb.String()
  }
  ```

- [ ] **Step 5: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/tui/tab_copytrading.go
  git commit -m "feat(tui): copytrading tab — card wrapper, live feed sub-tab"
  ```

---

## Task 7: Update `tab_markets.go` — card wrapper + category sidebar

**Files:**
- Modify: `internal/tui/tab_markets.go`

**Context:** `MarketsModel` has three modes (`modeList`, `modeDetail`, `modeOrder`) and two view modes (`viewTrending`, `viewCategories`). The redesign affects the `modeList` view only: add category sidebar card + market list card side by side.

- [ ] **Step 1: Read the full `View()` in `tab_markets.go`**

  Find the `modeList` branch.

- [ ] **Step 2: Add view-mode sub-tab bar**

  ```go
  func (m MarketsModel) renderViewModeBar() string {
  	hLabel := "h Trending"
  	cLabel := "c Categories"
  	if m.viewMode == viewCategories {
  		return lipgloss.JoinHorizontal(lipgloss.Top,
  			StyleSubTabInactive.Render(hLabel),
  			StyleSubTabActive.Render(cLabel),
  		)
  	}
  	return lipgloss.JoinHorizontal(lipgloss.Top,
  		StyleSubTabActive.Render(hLabel),
  		StyleSubTabInactive.Render(cLabel),
  	)
  }
  ```

- [ ] **Step 3: Add category sidebar card**

  ```go
  func (m MarketsModel) renderCategoryCard(height int) string {
  	var sb strings.Builder
  	for i, tag := range m.tags {
  		if i > 12 { break } // cap height
  		name := tag.Label
  		if len(name) > 12 { name = name[:12] }
  		label := fmt.Sprintf(" %s (%d)", name, tag.EventsCount)
  		if tag.Slug == m.activeTag {
  			sb.WriteString(StyleSubTabActive.Render(label) + "\n")
  		} else {
  			sb.WriteString(StyleFgDim.Render(label) + "\n")
  		}
  	}
  	return renderCard("Categories", sb.String(), 16, false)
  }
  ```

- [ ] **Step 4: Update `modeList` view to use side-by-side layout**

  ```go
  case modeList:
  	subBar := m.renderViewModeBar()
  	catCard := m.renderCategoryCard(m.height)
  	listW := m.width - 18 // 16 cat card + 2 gap

  	// build market list content (existing logic, just wrapped in renderCard)
  	listContent := m.renderMarketList(listW)
  	listTitle := fmt.Sprintf("Markets — %s  %d markets  /=search", m.activeTag, len(m.markets))
  	listCard := renderCard(listTitle, listContent, listW, true)

  	row := lipgloss.JoinHorizontal(lipgloss.Top, catCard, "  ", listCard)
  	help := renderHelpPanel("↑↓=select | Enter=detail | b=buy | s=sell | a=alert | /=search | h=trending | c=cat | Tab=next", m.width)
  	return lipgloss.JoinVertical(lipgloss.Left, " ", subBar, " ", row, help)
  ```

  > Extract existing market-list rendering (the loop over `m.markets`) into a `renderMarketList(width int) string` helper so the card wrapper can hold it cleanly.

- [ ] **Step 5: Replace `⚑` alert indicator**

  In the market list renderer, wherever price alerts are shown, use `⚑` (U+2691) instead of any emoji:
  ```go
  alertStr := " "
  if m.priceAlerts[cond.ConditionID] {
      alertStr = StyleWarning.Render("⚑")
  }
  ```

- [ ] **Step 6: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add internal/tui/tab_markets.go
  git commit -m "feat(tui): markets tab — card wrapper, category sidebar, safe alert symbol"
  ```

---

## Task 8: Update `tab_logs.go` — card wrapper + freeze indicator

**Files:**
- Modify: `internal/tui/tab_logs.go`

**Context:** `LogsModel` has a `viewport.Model`, `lines []BotEventMsg`, `filter string`, `freeze bool`. The viewport handles scrolling. Only `View()` changes.

- [ ] **Step 1: Read the full `View()` in `tab_logs.go`**

- [ ] **Step 2: Wrap viewport in `renderCard()` with freeze indicator in title**

  ```go
  func (m LogsModel) View() string {
  	title := "System Logs"
  	if m.freeze {
  		title += "  " + StyleWarning.Render("[FROZEN]")
  	}
  	if m.filter != "" {
  		title += "  filter:" + StyleValue.Render(m.filter)
  	}

  	card := renderCard(title, m.viewport.View(), m.width, true)
  	help := renderHelpPanel("↑↓=scroll | f=freeze | c=clear | /=filter | Tab=next", m.width)
  	return lipgloss.JoinVertical(lipgloss.Left, " ", card, help)
  }
  ```

- [ ] **Step 3: Update `renderLines()` to use spec log level colors**

  Ensure `renderLines()` applies:
  - `[INF]` → `StyleAccent.Render("[INF]")`  (ColorBright)
  - `[WRN]` → `StyleWarning.Render("[WRN]")`
  - `[ERR]` → `StyleError.Render("[ERR]")`
  - `[DBG]` → `StyleFgDim.Render("[DBG]")`
  - Timestamp → `StyleMuted.Render(ts)`

  Check the existing `renderLines()` and update any deviations.

- [ ] **Step 4: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add internal/tui/tab_logs.go
  git commit -m "feat(tui): logs tab — card wrapper, freeze indicator, log level colors"
  ```

---

## Task 9: Update `tab_settings.go` — section nav card + fields card

**Files:**
- Modify: `internal/tui/tab_settings.go`

**Context:** `SettingsModel` has `allFields []FieldDef` (defined as a package var), `cursor int`, `textInput textinput.Model`, and an `onSave` callback. The existing `View()` renders all fields in a single scrollable list. The spec adds a section nav sidebar.

- [ ] **Step 1: Read the full `View()` and `SettingsModel` struct in `tab_settings.go`**

- [ ] **Step 2: Add `renderSectionNav()` helper**

  ```go
  func (m SettingsModel) renderSectionNav(width int) string {
  	// Collect unique section names in order
  	seen := map[string]bool{}
  	var sections []string
  	for _, f := range allFields {
  		sec := f.Section()
  		if !seen[sec] {
  			seen[sec] = true
  			sections = append(sections, sec)
  		}
  	}

  	// Current section = section of field at cursor
  	currentSec := ""
  	if m.cursor < len(allFields) {
  		currentSec = allFields[m.cursor].Section()
  	}

  	var sb strings.Builder
  	for _, sec := range sections {
  		if sec == currentSec {
  			sb.WriteString(StyleSubTabActive.Render(" "+sec) + "\n")
  		} else {
  			sb.WriteString(StyleFgDim.Render(" "+sec) + "\n")
  		}
  	}
  	return renderCard("Sections", sb.String(), width, false)
  }
  ```

- [ ] **Step 3: Wrap the field list in `renderCard()` and add section nav**

  In `View()`, replace the current rendering with a two-column layout:

  ```go
  func (m SettingsModel) View() string {
  	navW  := 16
  	mainW := m.width - navW - 2

  	navCard  := m.renderSectionNav(navW)
  	// existing field rendering, wrapped in renderCard
  	fieldsBody := m.renderFields(mainW)
  	currentSec := ""
  	if m.cursor < len(allFields) {
  		currentSec = allFields[m.cursor].Section()
  	}
  	mainCard := renderCard(currentSec, fieldsBody, mainW, true)
  	row := lipgloss.JoinHorizontal(lipgloss.Top, navCard, "  ", mainCard)

  	help := renderHelpPanel("↑↓=field | Enter=edit | space=toggle | ←→=enum | s=save | Esc=cancel | Tab=next", m.width)
  	return lipgloss.JoinVertical(lipgloss.Left, " ", row, help)
  }
  ```

- [ ] **Step 4: Extract `renderFields()` from existing View() body**

  Move the existing per-field rendering loop into a `renderFields(width int) string` method. Keep all existing logic (active field highlight, toggle rendering, enum arrows, text input, save/unsaved indicator) exactly as-is.

- [ ] **Step 5: Build and test**

  ```bash
  go build ./...
  go test ./internal/tui/...
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add internal/tui/tab_settings.go
  git commit -m "feat(tui): settings tab — section nav card, fields card"
  ```

---

## Task 10: Update `splash.go` — box border style

**Files:**
- Modify: `internal/tui/splash.go`

**Context:** Light touch only. The existing `StyleSplashBox` uses `BorderThick` with `ColorPrimary` — already matches the spec. Verify and adjust only if the current code deviates.

- [ ] **Step 1: Read `splash.go` View() and check `StyleSplashBox`**

  Check `styles.go`:
  ```go
  StyleSplashBox = lipgloss.NewStyle().
      Border(BorderThick).
      BorderForeground(ColorPrimary).
      Background(ColorSurface).
      Padding(2, 4)
  ```
  This already matches the spec (`BorderThick` ┏━┓, `ColorPrimary`). No change needed unless `Background` should be `ColorBgMid` instead of `ColorSurface` (they are aliases — same value `#13102a`).

- [ ] **Step 2: Verify subtitle style**

  In `splash.go` View(), the subtitle uses `StyleSplashSubtitle` which is `Foreground(ColorPrimary).Bold(true)` — matches spec. No change.

- [ ] **Step 3: Only commit if changes were made**

  If no changes needed, skip commit. If a tweak was made:
  ```bash
  git add internal/tui/splash.go internal/tui/styles.go
  git commit -m "feat(tui): splash — align box style with card aesthetic"
  ```

---

## Task 11: Final check — remove unused `StyleSidebar*` vars (optional)

**Files:**
- Modify: `internal/tui/styles.go`

Per spec §16, remove `StyleSidebar*` vars if they are no longer referenced anywhere (they were added for a sidebar nav that was not chosen).

- [ ] **Step 1: Check for usages**

  ```bash
  grep -rn "StyleSidebar" internal/tui/
  ```

- [ ] **Step 2: If no remaining usages outside `styles.go` itself, remove the declarations**

  Remove `StyleSidebar`, `StyleSidebarActive`, `StyleSidebarInactive`, `StyleSidebarLogo`, `StyleSidebarSubtitle`, `StyleSidebarSep`, `StyleSidebarLabel` from `styles.go`.

  > Exception: `StyleSidebarLogo` and `StyleSidebarSubtitle` are used in `tab_overview.go` (in the old `standard`/`default` branches). After Task 2 replaces those branches, they may become unused. Verify after Task 2.

- [ ] **Step 3: Build**

  ```bash
  go build ./...
  ```

- [ ] **Step 4: Commit if changes made**

  ```bash
  git add internal/tui/styles.go
  git commit -m "chore(tui): remove unused StyleSidebar* vars"
  ```

---

## Task 12: End-to-end smoke test

- [ ] **Step 1: Run the bot in headless mode to verify no panics**

  ```bash
  cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot"
  go build -o /tmp/orbitron-test ./cmd/bot/
  /tmp/orbitron-test --no-tui --help 2>&1 | head -5
  ```
  Expected: prints usage, no panic.

- [ ] **Step 2: Run all tests**

  ```bash
  go test ./...
  ```
  Expected: all pass.

- [ ] **Step 3: Final commit if any remaining unstaged changes**

  ```bash
  git status
  # commit anything remaining
  ```
