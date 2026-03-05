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
	cursor    int
	detail    *gamma.Market

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

	sel := make(map[string]bool)
	if primaryID != "" {
		sel[primaryID] = true
	}

	return MarketsModel{
		mode:       modeList,
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
	case "f":
		m.cycleTag()
		m.cursor = 0
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
		m.mode = modeOrder
		m.priceInput.Focus()
	case "s":
		m.orderSide = "NO"
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

func (m MarketsModel) cycleTag() {
	if len(m.tags) == 0 {
		return
	}
	if m.activeTag == "" {
		if len(m.tags) > 0 {
			m.activeTag = m.tags[0].Slug
		}
		return
	}
	for i, tg := range m.tags {
		if tg.Slug == m.activeTag {
			if i+1 < len(m.tags) {
				m.activeTag = m.tags[i+1].Slug
			} else {
				m.activeTag = ""
			}
			return
		}
	}
	m.activeTag = ""
}

func (m MarketsModel) filtered() []gamma.Market {
	if m.activeTag == "" {
		return m.markets
	}
	var out []gamma.Market
	for _, mk := range m.markets {
		for _, tg := range mk.Tags {
			if tg.Slug == m.activeTag {
				out = append(out, mk)
				break
			}
		}
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

	tagLabel := "All"
	if m.activeTag != "" {
		tagLabel = m.activeTag
	}
	sb.WriteString(StyleMuted.Render(fmt.Sprintf(
		"Filter: [%s]  [f] cycle tag  [Enter] detail  [↑↓/jk] navigate\n\n", tagLabel,
	)))

	filtered := m.filtered()
	if len(filtered) == 0 {
		sb.WriteString(StyleMuted.Render("No markets found."))
		return sb.String()
	}

	visibleH := m.height - 8
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
		line := fmt.Sprintf("%-52s YES %-5s  Vol $%s",
			mktTruncate(mk.Question, 52),
			fmt.Sprintf("%.0f%%", yesPrice*100),
			mktFormatVolume(float64(mk.Volume)),
		)
		if i == m.cursor {
			sb.WriteString(StyleFieldActive.Render("▶ "+line) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}

	sb.WriteString("\n" + StyleHelpBar.Render("[b] Buy YES  [s] Buy NO  [a] Alert  [Esc] Back  in detail view"))
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
	sb.WriteString(StyleHelpBar.Render("[b] Buy YES  [s] Buy NO  [Esc] Back"))
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
