# Copytrading Wallet Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add add/remove/edit/toggle-enabled for copytrading wallets in the TUI (Tab 4) and Telegram Bot via commands + inline keyboard buttons.

**Architecture:** Extract pure config-mutation helpers (`addTrader`, `removeTrader`, `toggleTrader`, `editTrader`) into `internal/tui/tab_copytrading.go` and test them independently. The TUI `CopytradingModel` gains a modal form state machine (table / add / edit / confirm-delete). The Telegram Bot gains three commands and per-trader inline keyboard buttons. Both paths call `config.Save()` + emit `ConfigReloadedMsg`; `CopyTrader`'s existing fsnotify loop handles the rest automatically.

**Tech Stack:** Go 1.24, `charmbracelet/bubbletea` v1.3.10, `charmbracelet/bubbles` (textinput), `BurntSushi/toml`, `go-telegram-bot-api/v5`, `testify`

---

## Task 1: Add CopyKeyMap to keys.go

**Files:**
- Modify: `internal/tui/keys.go`

**Step 1: Add the key map struct and default bindings**

Append to `internal/tui/keys.go`:

```go
// CopyKeyMap holds keybindings for the Copytrading tab.
type CopyKeyMap struct {
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Toggle key.Binding
}

// CopyKeys is the default copytrading keybinding set.
var CopyKeys = CopyKeyMap{
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add trader"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit trader"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete trader"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle enabled"),
	),
}
```

**Step 2: Verify it compiles**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go build ./...
```

Expected: no errors.

**Step 3: Commit**

```bash
git add internal/tui/keys.go
git commit -m "feat(tui): add CopyKeyMap for copytrading tab keybindings"
```

---

## Task 2: Extract pure config-mutation helpers + tests

These helpers mutate `*config.Config` in-memory (no I/O). They are the testable core of the feature.

**Files:**
- Modify: `internal/tui/tab_copytrading.go` (add helpers at bottom)
- Create: `internal/tui/copytrading_helpers_test.go`

**Step 1: Write the failing tests**

Create `internal/tui/copytrading_helpers_test.go`:

```go
package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/config"
)

func baseCfg() *config.Config {
	return &config.Config{
		Copytrading: config.CopytradingConfig{
			Traders: []config.TraderConfig{
				{Address: "0xAAA", Label: "alice", Enabled: true, AllocationPct: 5.0, MaxPositionUSD: 50.0},
				{Address: "0xBBB", Label: "bob", Enabled: false, AllocationPct: 10.0, MaxPositionUSD: 100.0},
			},
		},
	}
}

func TestAddTrader_AppendNew(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "0xCCC", "carol", 3.0, 30.0)
	require.NoError(t, err)
	require.Len(t, cfg.Copytrading.Traders, 3)
	tr := cfg.Copytrading.Traders[2]
	assert.Equal(t, "0xCCC", tr.Address)
	assert.Equal(t, "carol", tr.Label)
	assert.InDelta(t, 3.0, tr.AllocationPct, 0.001)
	assert.InDelta(t, 30.0, tr.MaxPositionUSD, 0.001)
	assert.True(t, tr.Enabled)
}

func TestAddTrader_DuplicateAddress(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "0xAAA", "dup", 5.0, 50.0)
	require.Error(t, err)
	assert.Len(t, cfg.Copytrading.Traders, 2)
}

func TestAddTrader_EmptyAddress(t *testing.T) {
	cfg := baseCfg()
	err := addTrader(cfg, "", "nobody", 5.0, 50.0)
	require.Error(t, err)
}

func TestRemoveTrader_Existing(t *testing.T) {
	cfg := baseCfg()
	err := removeTrader(cfg, "0xAAA")
	require.NoError(t, err)
	require.Len(t, cfg.Copytrading.Traders, 1)
	assert.Equal(t, "0xBBB", cfg.Copytrading.Traders[0].Address)
}

func TestRemoveTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := removeTrader(cfg, "0xZZZ")
	require.Error(t, err)
}

func TestToggleTrader_EnablesDisabled(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xBBB")
	require.NoError(t, err)
	assert.True(t, cfg.Copytrading.Traders[1].Enabled)
}

func TestToggleTrader_DisablesEnabled(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xAAA")
	require.NoError(t, err)
	assert.False(t, cfg.Copytrading.Traders[0].Enabled)
}

func TestToggleTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := toggleTrader(cfg, "0xZZZ")
	require.Error(t, err)
}

func TestEditTrader_UpdatesFields(t *testing.T) {
	cfg := baseCfg()
	err := editTrader(cfg, "0xAAA", "ALICE", 7.5, 75.0)
	require.NoError(t, err)
	tr := cfg.Copytrading.Traders[0]
	assert.Equal(t, "ALICE", tr.Label)
	assert.InDelta(t, 7.5, tr.AllocationPct, 0.001)
	assert.InDelta(t, 75.0, tr.MaxPositionUSD, 0.001)
}

func TestEditTrader_NotFound(t *testing.T) {
	cfg := baseCfg()
	err := editTrader(cfg, "0xZZZ", "x", 5.0, 50.0)
	require.Error(t, err)
}
```

**Step 2: Run to confirm they fail**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go test ./internal/tui/ -run TestAddTrader -v 2>&1 | head -20
```

Expected: `undefined: addTrader` or similar.

**Step 3: Add helpers to tab_copytrading.go**

Append to bottom of `internal/tui/tab_copytrading.go`:

```go
// addTrader appends a new trader to cfg. Returns error if address is empty or already exists.
func addTrader(cfg *config.Config, address, label string, allocPct, maxPositionUSD float64) error {
	if address == "" {
		return fmt.Errorf("address is required")
	}
	for _, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			return fmt.Errorf("trader %q already exists", address)
		}
	}
	cfg.Copytrading.Traders = append(cfg.Copytrading.Traders, config.TraderConfig{
		Address:        address,
		Label:          label,
		Enabled:        true,
		AllocationPct:  allocPct,
		MaxPositionUSD: maxPositionUSD,
		SizeMode:       cfg.Copytrading.SizeMode,
	})
	return nil
}

// removeTrader removes the trader with the given address. Returns error if not found.
func removeTrader(cfg *config.Config, address string) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders = append(cfg.Copytrading.Traders[:i], cfg.Copytrading.Traders[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}

// toggleTrader flips the Enabled flag on the trader with the given address.
func toggleTrader(cfg *config.Config, address string) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders[i].Enabled = !t.Enabled
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}

// editTrader updates label, allocPct, and maxPositionUSD for the trader with the given address.
func editTrader(cfg *config.Config, address, label string, allocPct, maxPositionUSD float64) error {
	for i, t := range cfg.Copytrading.Traders {
		if t.Address == address {
			cfg.Copytrading.Traders[i].Label = label
			cfg.Copytrading.Traders[i].AllocationPct = allocPct
			cfg.Copytrading.Traders[i].MaxPositionUSD = maxPositionUSD
			return nil
		}
	}
	return fmt.Errorf("trader %q not found", address)
}
```

Add `"fmt"` to imports in `tab_copytrading.go` if not already present.

**Step 4: Run tests to confirm they pass**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go test ./internal/tui/ -run "TestAddTrader|TestRemoveTrader|TestToggleTrader|TestEditTrader" -v
```

Expected: all PASS.

**Step 5: Commit**

```bash
git add internal/tui/tab_copytrading.go internal/tui/copytrading_helpers_test.go
git commit -m "feat(tui): add pure trader mutation helpers with tests"
```

---

## Task 3: Add modal form state to CopytradingModel

**Files:**
- Modify: `internal/tui/tab_copytrading.go`

**Step 1: Add imports and mode type at top of file**

Add to the import block: `"github.com/charmbracelet/bubbles/textinput"` and `"strconv"` and `"fmt"` (if not present).

After the `TraderRow` type, add:

```go
type copyMode int

const (
	copyModeTable         copyMode = iota
	copyModeAddForm
	copyModeEditForm
	copyModeConfirmDelete
)
```

**Step 2: Update CopytradingModel struct**

Replace the existing `CopytradingModel` struct with:

```go
// CopytradingModel is the Copytrading tab sub-model.
type CopytradingModel struct {
	tradersTable table.Model
	recentTrades []string
	width        int
	height       int

	cfg     *config.Config
	cfgPath string

	mode      copyMode
	editIdx   int              // index in cfg.Copytrading.Traders for edit/delete
	formFocus int              // which textinput is active (0-3)
	inputs    []textinput.Model // Address, Label, Alloc%, MaxPositionUSD
	formErr   string           // last save error
}
```

**Step 3: Update NewCopytradingModel signature**

Replace the existing `NewCopytradingModel` with:

```go
// NewCopytradingModel creates a new CopytradingModel.
func NewCopytradingModel(cfg *config.Config, cfgPath string, width, height int) CopytradingModel {
	cols := []table.Column{
		{Title: i18n.T().CopyColAddress, Width: 20},
		{Title: i18n.T().CopyColLabel, Width: 18},
		{Title: i18n.T().CopyColStatus, Width: 10},
		{Title: i18n.T().CopyColAlloc, Width: 8},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(max(height/2-3, 1)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	t.SetStyles(s)

	return CopytradingModel{
		tradersTable: t,
		width:        width,
		height:       height,
		cfg:          cfg,
		cfgPath:      cfgPath,
		inputs:       makeFormInputs(),
	}
}

// makeFormInputs creates the four textinputs for the add/edit form.
func makeFormInputs() []textinput.Model {
	placeholders := []string{"0x… wallet address", "label (optional)", "alloc % (e.g. 5)", "max position USD (e.g. 50)"}
	inputs := make([]textinput.Model, 4)
	for i, ph := range placeholders {
		ti := textinput.New()
		ti.Placeholder = ph
		ti.CharLimit = 80
		inputs[i] = ti
	}
	inputs[0].Focus()
	return inputs
}
```

**Step 4: Add IsEditing helper (mirrors tab_settings.go pattern)**

```go
// IsEditing reports whether the model is in form-entry mode (blocks global tab switching).
func (m CopytradingModel) IsEditing() bool {
	return m.mode != copyModeTable
}
```

**Step 5: Verify compilation**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go build ./internal/tui/ 2>&1
```

Expected: errors about `NewCopytradingModel` call signature in `app.go` — that's fine, fix in Task 5.

**Step 6: Commit (partial — won't build yet)**

Skip commit here; continue to Task 4.

---

## Task 4: Implement Update and View for form modes

**Files:**
- Modify: `internal/tui/tab_copytrading.go`

**Step 1: Replace the Update method**

Replace the existing `Update` method with:

```go
func (m CopytradingModel) Update(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch m.mode {
	case copyModeTable:
		return m.updateTable(msg)
	case copyModeAddForm, copyModeEditForm:
		return m.updateForm(msg)
	case copyModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	}
	return m, nil
}

func (m CopytradingModel) updateTable(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, CopyKeys.Add):
			m.mode = copyModeAddForm
			m.inputs = makeFormInputs()
			m.formFocus = 0
			m.formErr = ""
			return m, textinput.Blink
		case key.Matches(msg, CopyKeys.Edit):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			m.editIdx = idx
			tr := m.cfg.Copytrading.Traders[idx]
			m.inputs = makeFormInputs()
			m.inputs[0].SetValue(tr.Address)
			m.inputs[0].Blur() // address not editable in edit mode
			m.inputs[1].SetValue(tr.Label)
			m.inputs[2].SetValue(strconv.FormatFloat(tr.AllocationPct, 'f', -1, 64))
			m.inputs[3].SetValue(strconv.FormatFloat(tr.MaxPositionUSD, 'f', -1, 64))
			m.formFocus = 1
			m.inputs[1].Focus()
			m.mode = copyModeEditForm
			m.formErr = ""
			return m, textinput.Blink
		case key.Matches(msg, CopyKeys.Delete):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			m.editIdx = idx
			m.mode = copyModeConfirmDelete
			return m, nil
		case key.Matches(msg, CopyKeys.Toggle):
			if m.cfg == nil || len(m.cfg.Copytrading.Traders) == 0 {
				return m, nil
			}
			idx := m.tradersTable.Cursor()
			if idx < 0 || idx >= len(m.cfg.Copytrading.Traders) {
				return m, nil
			}
			addr := m.cfg.Copytrading.Traders[idx].Address
			if err := toggleTrader(m.cfg, addr); err != nil {
				return m, nil
			}
			_ = config.Save(m.cfgPath, m.cfg)
			m.syncTable()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.tradersTable, cmd = m.tradersTable.Update(msg)
	return m, cmd
}

func (m CopytradingModel) updateForm(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = copyModeTable
			m.formErr = ""
			return m, nil
		case "tab", "down":
			m.inputs[m.formFocus].Blur()
			start := 0
			if m.mode == copyModeEditForm {
				start = 1 // address not editable
			}
			m.formFocus++
			if m.formFocus > 3 {
				m.formFocus = start
			}
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.inputs[m.formFocus].Blur()
			start := 0
			if m.mode == copyModeEditForm {
				start = 1
			}
			m.formFocus--
			if m.formFocus < start {
				m.formFocus = 3
			}
			m.inputs[m.formFocus].Focus()
			return m, textinput.Blink
		case "enter":
			if m.formFocus < 3 {
				// advance to next field
				m.inputs[m.formFocus].Blur()
				m.formFocus++
				m.inputs[m.formFocus].Focus()
				return m, textinput.Blink
			}
			// last field — save
			return m.saveForm()
		}
	}
	// Route to focused input
	var cmd tea.Cmd
	m.inputs[m.formFocus], cmd = m.inputs[m.formFocus].Update(msg)
	return m, cmd
}

func (m CopytradingModel) updateConfirmDelete(msg tea.Msg) (CopytradingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.cfg != nil && m.editIdx < len(m.cfg.Copytrading.Traders) {
				addr := m.cfg.Copytrading.Traders[m.editIdx].Address
				_ = removeTrader(m.cfg, addr)
				_ = config.Save(m.cfgPath, m.cfg)
				m.syncTable()
			}
			m.mode = copyModeTable
		default:
			m.mode = copyModeTable
		}
	}
	return m, nil
}

// saveForm validates inputs, calls addTrader or editTrader, saves config.
func (m CopytradingModel) saveForm() (CopytradingModel, tea.Cmd) {
	addr := strings.TrimSpace(m.inputs[0].Value())
	label := strings.TrimSpace(m.inputs[1].Value())
	allocStr := strings.TrimSpace(m.inputs[2].Value())
	maxStr := strings.TrimSpace(m.inputs[3].Value())

	allocPct := 5.0
	if allocStr != "" {
		if v, err := strconv.ParseFloat(allocStr, 64); err == nil {
			allocPct = v
		}
	}
	maxPos := 50.0
	if maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxPos = v
		}
	}

	var err error
	if m.mode == copyModeAddForm {
		err = addTrader(m.cfg, addr, label, allocPct, maxPos)
	} else {
		if m.editIdx < len(m.cfg.Copytrading.Traders) {
			addr = m.cfg.Copytrading.Traders[m.editIdx].Address
		}
		err = editTrader(m.cfg, addr, label, allocPct, maxPos)
	}
	if err != nil {
		m.formErr = err.Error()
		return m, nil
	}
	if saveErr := config.Save(m.cfgPath, m.cfg); saveErr != nil {
		m.formErr = saveErr.Error()
		return m, nil
	}
	m.syncTable()
	m.mode = copyModeTable
	m.formErr = ""
	return m, nil
}

// syncTable rebuilds table rows from cfg.
func (m *CopytradingModel) syncTable() {
	if m.cfg == nil {
		return
	}
	rows := make([]TraderRow, len(m.cfg.Copytrading.Traders))
	for i, t := range m.cfg.Copytrading.Traders {
		status := "disabled"
		if t.Enabled {
			status = "active"
		}
		rows[i] = TraderRow{
			Address:  t.Address,
			Label:    t.Label,
			Status:   status,
			AllocPct: strconv.FormatFloat(t.AllocationPct, 'f', 1, 64) + "%",
		}
	}
	m.SetTraderRows(rows)
}
```

**Step 2: Update the View method**

Replace the existing `View` with:

```go
func (m CopytradingModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleBold.Render(i18n.T().CopyTraders) + "\n")
	sb.WriteString(m.tradersTable.View() + "\n")

	switch m.mode {
	case copyModeTable:
		help := "  " + StyleMuted.Render("[a] add  [e] edit  [d] delete  [space] toggle")
		sb.WriteString(help + "\n")
	case copyModeAddForm:
		sb.WriteString(m.renderForm("Add Trader") + "\n")
	case copyModeEditForm:
		sb.WriteString(m.renderForm("Edit Trader") + "\n")
	case copyModeConfirmDelete:
		addr := ""
		if m.editIdx < len(m.cfg.Copytrading.Traders) {
			addr = m.cfg.Copytrading.Traders[m.editIdx].Address
		}
		prompt := StyleBold.Render(fmt.Sprintf("  Delete %s? [y/N]", addr))
		sb.WriteString(prompt + "\n")
	}

	sb.WriteString("\n" + StyleBold.Render(i18n.T().CopyRecentTrades) + "\n")
	if len(m.recentTrades) == 0 {
		sb.WriteString(StyleMuted.Render("  " + i18n.T().CopyNoData + "\n"))
	}
	for _, t := range m.recentTrades {
		sb.WriteString("  " + t + "\n")
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(sb.String())
}

// renderForm renders the add/edit textinput form.
func (m CopytradingModel) renderForm(title string) string {
	var sb strings.Builder
	sb.WriteString("\n  " + StyleBold.Render("── "+title+" ──") + "\n")
	labels := []string{"Address:     ", "Label:       ", "Alloc %:     ", "Max Pos USD: "}
	for i, inp := range m.inputs {
		prefix := "  "
		if m.formFocus == i {
			prefix = StyleAccent.Render("> ")
		}
		sb.WriteString(prefix + StyleMuted.Render(labels[i]) + inp.View() + "\n")
	}
	if m.formErr != "" {
		sb.WriteString("  " + StyleError.Render("Error: "+m.formErr) + "\n")
	}
	sb.WriteString("  " + StyleMuted.Render("[Enter] save  [Tab] next field  [Esc] cancel") + "\n")
	return sb.String()
}
```

> Note: `StyleAccent` and `StyleError` must exist in `styles.go`. Check that file. If they don't exist, add them (purple accent and red error).

**Step 3: Check styles.go for required styles**

```bash
grep -n "StyleAccent\|StyleError\|StyleMuted\|StyleBold" "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot/internal/tui/styles.go"
```

Add any missing styles to `styles.go`. For reference:

```go
// StyleAccent is purple-colored text used for focused indicators.
var StyleAccent = lipgloss.NewStyle().Foreground(ColorPurple)

// StyleError is red-colored text used for error messages.
var StyleError = lipgloss.NewStyle().Foreground(ColorRed)
```

**Step 4: Verify it compiles (after app.go fix in Task 5)**

Skip for now — proceed to Task 5.

---

## Task 5: Update app.go to pass cfg/cfgPath to CopytradingModel

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Update NewAppModel to pass cfg/cfgPath**

In `NewAppModel`, change the `copytrader` line from:
```go
copytrader: NewCopytradingModel(width, cw),
```
to:
```go
copytrader: NewCopytradingModel(cfg, cfgPath, width, cw),
```

**Step 2: Update the ConfigReloadedMsg handler in Update**

In the `ConfigReloadedMsg` case, after updating `m.cfg`, also update the copytrader's cfg pointer:

```go
case ConfigReloadedMsg:
    m.cfg = msg.Config
    m.copytrader.cfg = msg.Config   // keep copytrader in sync
    var cmd tea.Cmd
    m.settings, cmd = m.settings.Update(msg)
    return m, tea.Batch(cmd, m.bus.WaitForEvent())
```

**Step 3: Block tab switching when copytrading form is open**

In the `"tab"` and `"shift+tab"` cases, add the same guard as for settings:

```go
case "tab":
    if m.activeTab == TabSettings && m.settings.IsEditing() {
        break
    }
    if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
        break
    }
    m.activeTab = (m.activeTab + 1) % tabCount
    // ...
case "shift+tab":
    if m.activeTab == TabSettings && m.settings.IsEditing() {
        break
    }
    if m.activeTab == TabCopytrading && m.copytrader.IsEditing() {
        break
    }
    // ...
```

Also guard `"1"-"6"` digit switching the same way.

**Step 4: Update LanguageChangedMsg handler**

In the `LanguageChangedMsg` case, update the `NewCopytradingModel` call:
```go
m.copytrader = NewCopytradingModel(m.cfg, m.cfgPath, m.width, cw)
```

**Step 5: Build and run**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go build ./...
```

Expected: clean build.

**Step 6: Run all tests**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go test ./...
```

Expected: all pass.

**Step 7: Commit**

```bash
git add internal/tui/tab_copytrading.go internal/tui/app.go
git commit -m "feat(tui): add inline add/edit/delete/toggle form to Copytrading tab"
```

---

## Task 6: Add /addtrader, /removetrader, /toggletrader Telegram commands

**Files:**
- Modify: `internal/telegrambot/handlers.go`

**Step 1: Add three command cases to handleCommand**

In the `switch msg.Command()` block, after `case "copy":`, add:

```go
case "addtrader":
    args := strings.Fields(msg.CommandArguments())
    if len(args) < 1 {
        b.sendText(msg.Chat.ID, RenderError("Usage: /addtrader &lt;address&gt; [label] [alloc_pct]"))
        return
    }
    b.doAddTrader(ctx, msg.Chat.ID, args)
case "removetrader":
    addr := strings.TrimSpace(msg.CommandArguments())
    if addr == "" {
        b.sendText(msg.Chat.ID, RenderError("Usage: /removetrader &lt;address&gt;"))
        return
    }
    b.doRemoveTrader(ctx, msg.Chat.ID, addr)
case "toggletrader":
    addr := strings.TrimSpace(msg.CommandArguments())
    if addr == "" {
        b.sendText(msg.Chat.ID, RenderError("Usage: /toggletrader &lt;address&gt;"))
        return
    }
    b.doToggleTrader(ctx, msg.Chat.ID, addr)
```

**Step 2: Add the three action helpers**

Append to `handlers.go` (after `doSetSetting`):

```go
func (b *Bot) doAddTrader(_ context.Context, chatID int64, args []string) {
	addr := args[0]
	label := ""
	if len(args) > 1 {
		label = args[1]
	}
	allocPct := 5.0
	if len(args) > 2 {
		if v, err := strconv.ParseFloat(args[2], 64); err == nil {
			allocPct = v
		}
	}

	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	for _, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			b.cfgMu.Unlock()
			b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q already exists.", addr)))
			return
		}
	}
	if addr == "" {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError("Address is required."))
		return
	}

	cfgCopy.Copytrading.Traders = append(cfgCopy.Copytrading.Traders, config.TraderConfig{
		Address:        addr,
		Label:          label,
		Enabled:        true,
		AllocationPct:  allocPct,
		MaxPositionUSD: 50.0,
		SizeMode:       cfgCopy.Copytrading.SizeMode,
	})

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> added (label: %s, alloc: %.1f%%).", addr, label, allocPct)))
}

func (b *Bot) doRemoveTrader(_ context.Context, chatID int64, addr string) {
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	found := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders = append(cfgCopy.Copytrading.Traders[:i], cfgCopy.Copytrading.Traders[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q not found.", addr)))
		return
	}

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> removed.", addr)))
}

func (b *Bot) doToggleTrader(_ context.Context, chatID int64, addr string) {
	b.cfgMu.Lock()
	cfgCopy := *b.cfg
	traders := make([]config.TraderConfig, len(cfgCopy.Copytrading.Traders))
	copy(traders, cfgCopy.Copytrading.Traders)
	cfgCopy.Copytrading.Traders = traders

	found := false
	newState := false
	for i, t := range cfgCopy.Copytrading.Traders {
		if t.Address == addr {
			cfgCopy.Copytrading.Traders[i].Enabled = !t.Enabled
			newState = cfgCopy.Copytrading.Traders[i].Enabled
			found = true
			break
		}
	}
	if !found {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Trader %q not found.", addr)))
		return
	}

	if err := config.Save(b.cfgPath, &cfgCopy); err != nil {
		b.cfgMu.Unlock()
		b.sendText(chatID, RenderError(fmt.Sprintf("Failed to save config: %v", err)))
		return
	}
	*b.cfg = cfgCopy
	b.cfgMu.Unlock()

	b.bus.Send(tui.ConfigReloadedMsg{Config: b.cfg})
	state := "disabled"
	if newState {
		state = "enabled"
	}
	b.sendText(chatID, RenderSuccess(fmt.Sprintf("Trader <code>%s</code> %s.", addr, state)))
}
```

**Step 3: Build**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go build ./...
```

Expected: clean.

**Step 4: Commit**

```bash
git add internal/telegrambot/handlers.go
git commit -m "feat(telegram): add /addtrader, /removetrader, /toggletrader commands"
```

---

## Task 7: Update Telegram copytrading keyboard with per-trader buttons

**Files:**
- Modify: `internal/telegrambot/handlers.go`

**Step 1: Replace sendCopytrading to use a richer keyboard**

Replace `sendCopytrading`:

```go
func (b *Bot) sendCopytrading(chatID int64) {
	b.cfgMu.RLock()
	traders := b.cfg.Copytrading.Traders
	b.cfgMu.RUnlock()

	// Build display rows for the text (use the state cache for status)
	stateRows := b.state.Traders()
	statusMap := make(map[string]string, len(stateRows))
	for _, r := range stateRows {
		statusMap[r.Address] = r.Status
	}

	text := RenderCopytrading(b.state.Traders())
	b.sendWithKeyboard(chatID, text, copytradingKeyboard(traders, statusMap))
}
```

**Step 2: Add copytradingKeyboard function**

Add after `backKeyboard()`:

```go
func copytradingKeyboard(traders []config.TraderConfig, statusMap map[string]string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range traders {
		addr := t.Address
		short := addr
		if len(short) > 12 {
			short = short[:6] + "…" + short[len(short)-4:]
		}
		label := t.Label
		if label == "" {
			label = short
		}
		toggleIcon := "▶ Enable"
		if t.Enabled {
			toggleIcon = "⏸ Disable"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s  %s", label, toggleIcon),
				"trader:toggle:"+addr,
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑 Remove",
				"trader:remove:"+addr,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("← Back", "cmd:menu"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
```

Note: add `"github.com/atlasdev/polytrade-bot/internal/config"` to imports in `handlers.go` if not already there.

**Step 3: Handle trader:toggle: and trader:remove: callbacks**

In `handleCallback`, in the `switch` block after the `cancel:` case, add:

```go
case strings.HasPrefix(data, "trader:toggle:"):
    addr := strings.TrimPrefix(data, "trader:toggle:")
    b.doToggleTrader(ctx, chatID, addr)
    b.sendCopytrading(chatID)
case strings.HasPrefix(data, "trader:remove:"):
    addr := strings.TrimPrefix(data, "trader:remove:")
    b.doRemoveTrader(ctx, chatID, addr)
    b.sendCopytrading(chatID)
```

**Step 4: Build and test**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go build ./... && go test ./...
```

Expected: clean build, all tests pass.

**Step 5: Commit**

```bash
git add internal/telegrambot/handlers.go
git commit -m "feat(telegram): add per-trader toggle/remove buttons in copytrading keyboard"
```

---

## Task 8: Final verification

**Step 1: Run full test suite**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go test ./... -v 2>&1 | tail -30
```

Expected: all PASS.

**Step 2: Run vet**

```bash
cd "/home/atlasdev/Рабочий стол/Scripts/Golang/polytrade-bot" && go vet ./...
```

Expected: no warnings.

**Step 3: Manual smoke-test checklist**

TUI:
- [ ] Tab 4 shows traders table with `[a] add  [e] edit  [d] delete  [space] toggle` help
- [ ] Press `a` → blank form with 4 fields; `Esc` cancels back to table
- [ ] Add a trader → it appears in table and `config.toml` is updated
- [ ] Press `e` on a row → fields pre-filled (address disabled), can edit label/alloc/max
- [ ] Press `d` → confirm prompt `Delete 0x…? [y/N]`; `y` removes it; any other key cancels
- [ ] Press `space` → toggles enabled, saves immediately
- [ ] Tab/Shift+Tab don't switch tabs while form is open

Telegram:
- [ ] `/addtrader 0xABC alice 5` → success reply, trader appears in `/copy`
- [ ] `/removetrader 0xABC` → success reply, trader gone
- [ ] `/toggletrader 0xABC` → flips enabled
- [ ] Inline buttons in Copytrading screen: Disable/Enable toggle, Remove

**Step 4: Commit design doc if not already committed**

```bash
git add docs/plans/
git commit -m "docs: add copytrading wallet management design and implementation plan"
```
