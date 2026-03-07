package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/polytrade-bot/internal/api/gamma"
)

type marketsMode int

const (
	modeList   marketsMode = iota
	modeDetail
	modeOrder
)

// marketWallet holds display info for a wallet in the Markets tab.
type marketWallet struct {
	ID    string
	Label string
}

// MarketsModel is the Markets tab sub-model.
type MarketsModel struct {
	mode      marketsMode
	markets   []gamma.Market
	tags      []gamma.Tag
	activeTag string
	tagIdx    int
	cursor    int
	detail    *gamma.Market

	// multi-select
	selected  map[string]bool // conditionID → selected for batch buy

	// batch buy form
	batchMode bool
	batchSide string
	batchSize textinput.Model

	// order form
	orderSide  string
	orderType  string
	priceInput textinput.Model
	sizeInput  textinput.Model
	wallets    []marketWallet
	selWallets map[string]bool
	primaryID  string

	width  int
	height int
}

// NewMarketsModel creates the initial MarketsModel.
func NewMarketsModel(wallets []marketWallet, primaryID string) MarketsModel {
	pi := textinput.New()
	pi.Placeholder = "0.50"
	pi.Width = 8

	si := textinput.New()
	si.Placeholder = "100"
	si.Width = 8

	bs := textinput.New()
	bs.Placeholder = "50"
	bs.Width = 8

	sel := make(map[string]bool)
	if primaryID != "" {
		sel[primaryID] = true
	}

	return MarketsModel{
		mode:       modeList,
		selected:   make(map[string]bool),
		batchSize:  bs,
		orderSide:  "YES",
		orderType:  "GTC",
		priceInput: pi,
		sizeInput:  si,
		wallets:    wallets,
		selWallets: sel,
		primaryID:  primaryID,
	}
}

func (m MarketsModel) Init() tea.Cmd { return nil }

func (m MarketsModel) Update(msg tea.Msg) (MarketsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MarketsUpdatedMsg:
		m.markets = msg.Markets
		m.tags = msg.Tags
		if m.cursor >= len(m.markets) && len(m.markets) > 0 {
			m.cursor = len(m.markets) - 1
		}

	case tea.KeyMsg:
		// Batch form intercepts keys when open
		if m.batchMode {
			return m.updateBatch(msg)
		}
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeDetail:
			return m.updateDetail(msg)
		case modeOrder:
			return m.updateOrder(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m MarketsModel) updateList(msg tea.KeyMsg) (MarketsModel, tea.Cmd) {
	filtered := m.filtered()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(filtered) {
			cp := filtered[m.cursor]
			m.detail = &cp
			m.mode = modeDetail
		}
	case "f", "F":
		if msg.String() == "F" {
			m.activeTag = reverseCycleTagSlug(m.activeTag, m.tags)
		} else {
			m.activeTag = cycleTagSlug(m.activeTag, m.tags)
		}
		m.cursor = 0
	case "space":
		if m.cursor < len(filtered) {
			cid := filtered[m.cursor].ConditionID
			if m.selected[cid] {
				delete(m.selected, cid)
			} else {
				m.selected[cid] = true
			}
		}
	case "b":
		if len(m.selected) > 0 {
			m.batchMode = true
			m.batchSide = "YES"
			m.batchSize.Focus()
		}
	case "n":
		if len(m.selected) > 0 {
			m.batchMode = true
			m.batchSide = "NO"
			m.batchSize.Focus()
		}
	case "escape", "esc":
		m.selected = make(map[string]bool)
		m.batchMode = false
	}
	return m, nil
}

func (m MarketsModel) updateDetail(msg tea.KeyMsg) (MarketsModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.mode = modeList
		m.detail = nil
	case "b":
		m.orderSide = "YES"
		if len(m.detail.OutcomePrices) > 0 {
			m.priceInput.SetValue(string(m.detail.OutcomePrices[0]))
		}
		m.mode = modeOrder
		m.priceInput.Focus()
	case "s":
		m.orderSide = "NO"
		if len(m.detail.OutcomePrices) > 1 {
			m.priceInput.SetValue(string(m.detail.OutcomePrices[1]))
		}
		m.mode = modeOrder
		m.priceInput.Focus()
	}
	return m, nil
}

func (m MarketsModel) updateOrder(msg tea.KeyMsg) (MarketsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeDetail
		m.priceInput.Blur()
		m.sizeInput.Blur()
		return m, nil
	case "enter":
		return m.submitOrder()
	case "tab":
		if m.priceInput.Focused() {
			m.priceInput.Blur()
			m.sizeInput.Focus()
		} else {
			m.sizeInput.Blur()
			m.priceInput.Focus()
		}
		return m, nil
	}
	var cmd tea.Cmd
	if m.priceInput.Focused() {
		m.priceInput, cmd = m.priceInput.Update(msg)
	} else {
		m.sizeInput, cmd = m.sizeInput.Update(msg)
	}
	return m, cmd
}

func (m MarketsModel) submitOrder() (MarketsModel, tea.Cmd) {
	if m.detail == nil {
		return m, nil
	}
	var walletIDs []string
	for id, ok := range m.selWallets {
		if ok {
			walletIDs = append(walletIDs, id)
		}
	}
	if len(walletIDs) == 0 && m.primaryID != "" {
		walletIDs = []string{m.primaryID}
	}
	var price, size float64
	fmt.Sscanf(m.priceInput.Value(), "%f", &price)
	fmt.Sscanf(m.sizeInput.Value(), "%f", &size)

	cid := m.detail.ConditionID
	side := m.orderSide
	ot := m.orderType

	cmd := func() tea.Msg {
		return PlaceOrderMsg{
			ConditionID: cid,
			WalletIDs:   walletIDs,
			Side:        side,
			Price:       price,
			Size:        size,
			OrderType:   ot,
		}
	}
	m.mode = modeDetail
	m.priceInput.Reset()
	m.sizeInput.Reset()
	return m, cmd
}

// cycleTagSlug returns the next tag slug after currentSlug, or "" (all) after the last tag.
func cycleTagSlug(currentSlug string, tags []gamma.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	if currentSlug == "" {
		return tags[0].Slug
	}
	for i, tg := range tags {
		if tg.Slug == currentSlug {
			if i+1 < len(tags) {
				return tags[i+1].Slug
			}
			return ""
		}
	}
	return ""
}

func reverseCycleTagSlug(currentSlug string, tags []gamma.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	if currentSlug == "" {
		return tags[len(tags)-1].Slug
	}
	for i, tg := range tags {
		if tg.Slug == currentSlug {
			if i > 0 {
				return tags[i-1].Slug
			}
			return ""
		}
	}
	return ""
}

func (m MarketsModel) filtered() []gamma.Market {
	if m.activeTag == "" {
		return m.markets
	}
	var out []gamma.Market
	for _, mk := range m.markets {
		// Match by tag slug
		for _, tg := range mk.Tags {
			if tg.Slug == m.activeTag {
				out = append(out, mk)
				goto next
			}
		}
		// Also match by category slug
		if mk.Category != "" {
			slug := strings.ToLower(strings.ReplaceAll(mk.Category, " ", "-"))
			if slug == m.activeTag {
				out = append(out, mk)
			}
		}
	next:
	}
	return out
}

func (m MarketsModel) View() string {
	switch m.mode {
	case modeList:
		return m.viewList()
	case modeDetail:
		return m.viewDetail()
	case modeOrder:
		return m.viewOrder()
	}
	return ""
}

func (m MarketsModel) viewList() string {
	var sb strings.Builder

	// Tag filter row
	tagLabel := "All"
	for _, tg := range m.tags {
		if tg.Slug == m.activeTag {
			tagLabel = tg.Label
			break
		}
	}
	filterLine := fmt.Sprintf("Filter: [%s]  [f/F] cycle  [Space] select  [b/n] batch buy", tagLabel)
	sb.WriteString(StyleMuted.Render(filterLine) + "\n")

	// Selection / batch status bar
	if len(m.selected) > 0 {
		status := fmt.Sprintf("  %d selected", len(m.selected))
		if m.batchMode {
			sb.WriteString(StyleFieldActive.Render(status+" · Batch "+m.batchSide) + "\n")
			sb.WriteString(fmt.Sprintf("  Size per market ($): %s\n", m.batchSize.View()))
			sb.WriteString(StyleHelpBar.Render("  [Enter] execute  [Esc] cancel") + "\n")
		} else {
			sb.WriteString(StyleFieldActive.Render(status+" · [b] Buy YES  [n] Buy NO  [Esc] clear") + "\n")
		}
	}
	sb.WriteString("\n")

	filtered := m.filtered()
	if len(filtered) == 0 {
		sb.WriteString(StyleMuted.Render("No markets found."))
		return sb.String()
	}

	visibleH := m.height - 10
	if visibleH < 5 {
		visibleH = 5
	}
	start := 0
	if m.cursor >= visibleH {
		start = m.cursor - visibleH + 1
	}
	end := start + visibleH
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := start; i < end; i++ {
		mk := filtered[i]
		yesPrice := mktYesProb(mk)
		sel := " "
		if m.selected[mk.ConditionID] {
			sel = "✓"
		}
		line := fmt.Sprintf("[%s] %-50s YES %-5s  Vol $%s",
			sel,
			mktTruncate(mk.Question, 50),
			fmt.Sprintf("%.0f%%", yesPrice*100),
			mktFormatVolume(float64(mk.Volume)),
		)
		if i == m.cursor {
			sb.WriteString(StyleFieldActive.Render("▶ "+line) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}

	if len(m.selected) == 0 {
		sb.WriteString("\n" + StyleHelpBar.Render("[Enter] detail  [Space] select  [f/F] filter  [↑↓/jk] navigate"))
	}
	return sb.String()
}

func (m MarketsModel) viewDetail() string {
	if m.detail == nil {
		return ""
	}
	mk := m.detail
	yes := mktYesProb(*mk)
	no := 1 - yes

	var sb strings.Builder
	sb.WriteString(StyleBold.Render(mk.Question) + "\n")
	sb.WriteString(StyleMuted.Render(fmt.Sprintf(
		"End: %s  |  Liq: $%s\n\n",
		mk.EndDateISO, mktFormatVolume(float64(mk.Liquidity)),
	)))

	yesBar := mktProgressBar(yes, 28)
	noBar := mktProgressBar(no, 28)
	sb.WriteString(fmt.Sprintf("YES  %s  %.1f¢\n", yesBar, yes*100))
	sb.WriteString(fmt.Sprintf("NO   %s  %.1f¢\n\n", noBar, no*100))

	walletLabel := "none"
	for _, w := range m.wallets {
		if w.ID == m.primaryID {
			walletLabel = "★ " + w.Label
			break
		}
	}
	sb.WriteString(StyleMuted.Render("Wallet: "+walletLabel) + "\n\n")
	sb.WriteString(StyleHelpBar.Render("[b] Buy YES  [s] Buy NO  [Esc] Back  (price auto-filled)"))
	return sb.String()
}

func (m MarketsModel) viewOrder() string {
	if m.detail == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(StyleBold.Render(fmt.Sprintf(
		"Buy %s — %s\n\n", m.orderSide, mktTruncate(m.detail.Question, 50),
	)))
	sb.WriteString(fmt.Sprintf("Price (0-1):  %s\n", m.priceInput.View()))
	sb.WriteString(fmt.Sprintf("Size (USDC):  %s\n", m.sizeInput.View()))

	var price, size float64
	fmt.Sscanf(m.priceInput.Value(), "%f", &price)
	fmt.Sscanf(m.sizeInput.Value(), "%f", &size)
	if price > 0 && size > 0 {
		sb.WriteString(StyleMuted.Render(fmt.Sprintf("\n→ Cost: $%.2f\n", price*size)))
	}
	sb.WriteString("\n" + StyleHelpBar.Render("[Tab] switch field  [Enter] submit  [Esc] cancel"))
	return sb.String()
}

func (m MarketsModel) updateBatch(msg tea.KeyMsg) (MarketsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.batchMode = false
		m.batchSize.Blur()
		m.batchSize.Reset()
		return m, nil
	case "enter":
		return m.submitBatch()
	}
	var cmd tea.Cmd
	m.batchSize, cmd = m.batchSize.Update(msg)
	return m, cmd
}

func (m MarketsModel) submitBatch() (MarketsModel, tea.Cmd) {
	var size float64
	fmt.Sscanf(m.batchSize.Value(), "%f", &size)
	if size <= 0 || len(m.selected) == 0 {
		return m, nil
	}
	cids := make([]string, 0, len(m.selected))
	for cid := range m.selected {
		cids = append(cids, cid)
	}
	side := m.batchSide
	walletID := m.primaryID
	cmd := func() tea.Msg {
		return BatchPlaceOrderMsg{
			ConditionIDs: cids,
			Side:         side,
			Size:         size,
			WalletID:     walletID,
		}
	}
	m.batchMode = false
	m.selected = make(map[string]bool)
	m.batchSize.Blur()
	m.batchSize.Reset()
	return m, cmd
}

// --- package-level helpers with mkt prefix to avoid conflicts ---

func mktYesProb(m gamma.Market) float64 {
	if len(m.OutcomePrices) > 0 {
		var f float64
		fmt.Sscanf(string(m.OutcomePrices[0]), "%f", &f)
		return f
	}
	return 0.5
}

func mktProgressBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa")).Render(strings.Repeat("█", filled)) +
		StyleMuted.Render(strings.Repeat("░", width-filled))
}

func mktFormatVolume(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.1fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

func mktTruncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
