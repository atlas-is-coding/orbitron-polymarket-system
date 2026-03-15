package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/atlasdev/orbitron/internal/i18n"
)

// TabID identifies a tab.
type TabID int

const (
	TabOverview   TabID = iota
	TabTrading          // Orders + Positions
	TabStrategies       // dedicated strategy management
	TabWallets
	TabCopytrading
	TabMarkets
	TabLogs
	TabSettings
	tabCount // sentinel
)

// sidebarWidth is the fixed width of the left navigation sidebar in chars.
const sidebarWidth = 22

var tabKeys = []string{"1", "2", "3", "4", "5", "6", "7", "8"}

// tabNames returns tab display names in the current locale.
func tabNames() []string {
	t := i18n.T()
	return []string{
		t.TabOverview,
		t.TabTrading,
		t.TabStrategies,
		t.TabWallets,
		t.TabCopytrading,
		t.TabMarkets,
		t.TabLogs,
		t.TabSettings,
	}
}

// tabIcons are unicode symbol prefixes for each tab.
var tabIcons = []string{"⊞", "≡", "⚡", "◎", "⇌", "⊛", "▦", "⚙"}

// RenderSidebar renders the left navigation sidebar.
// height is the full terminal height; the sidebar fills height-1 rows (leaving 1 for status bar).
func RenderSidebar(active TabID, height int, subsystems []SubsystemStatus) string {
	names := tabNames()
	inner := sidebarWidth - 2 // usable width inside padding
	sep := StyleSidebarSep.Render(strings.Repeat("─", inner))

	var sb strings.Builder

	// ── Logo ──────────────────────────────────────────────────────────────
	sb.WriteString(" " + StyleSidebarLogo.Render("◈ ORBITRON") + "\n")
	sb.WriteString(" " + StyleSidebarSubtitle.Render("NEXUS TERMINAL") + "\n")
	sb.WriteString("\n")
	sb.WriteString(" " + sep + "\n")
	sb.WriteString("\n")

	// ── Tabs ──────────────────────────────────────────────────────────────
	for i, name := range names {
		label := fmt.Sprintf("%s %s: %s", tabIcons[i], tabKeys[i], name)
		if TabID(i) == active {
			row := fmt.Sprintf(" ▶ %-*s", inner-3, label)
			sb.WriteString(StyleSidebarActive.Width(sidebarWidth).Render(row) + "\n")
		} else {
			row := fmt.Sprintf("   %-*s", inner-3, label)
			sb.WriteString(StyleSidebarInactive.Width(sidebarWidth).Render(row) + "\n")
		}
	}

	// ── Spacer ────────────────────────────────────────────────────────────
	headerRows := 5                   // logo(2) + blank(1) + sep(1) + blank(1)
	tabRows := len(names)             // one row per tab
	footerRows := 3 + len(subsystems) // sep(1) + blank(1) + label(1) + dots
	statusRows := 1
	used := headerRows + tabRows + footerRows + statusRows
	available := height - used
	for i := 0; i < available; i++ {
		sb.WriteString("\n")
	}

	// ── Subsystem health ──────────────────────────────────────────────────
	sb.WriteString(" " + sep + "\n")
	sb.WriteString("\n")
	sb.WriteString(" " + StyleSidebarLabel.Render("SUBSYSTEMS") + "\n")
	for _, s := range subsystems {
		if s.Active {
			sb.WriteString(" " + StyleSuccess.Render("●") + " " + StyleFgDim.Render(s.Name) + "\n")
		} else {
			sb.WriteString(" " + StyleMuted.Render("○") + " " + StyleMuted.Render(s.Name) + "\n")
		}
	}

	sidebarContent := sb.String()

	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Background(ColorSurface).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		Render(sidebarContent)
}

// renderPanel renders a titled content box with sharp borders.
// width is the total available width for the panel (border included).
// active controls border color: ColorPrimary when true, ColorBorder when false.
func renderPanel(title, content string, width int, active bool) string {
	borderColor := ColorBorder
	if active {
		borderColor = ColorPrimary
	}
	innerW := width - 4 // subtract border(2) + padding(2)
	if innerW < 1 {
		innerW = 1
	}

	var body string
	if title != "" {
		titleLine := StylePanelTitle.Render(" ─ "+title) + "\n"
		divLine := StylePanelDivider.Render(" "+strings.Repeat("─", max(innerW-1, 0))) + "\n"
		body = titleLine + divLine + content
	} else {
		body = content
	}

	return lipgloss.NewStyle().
		Border(BorderNormal).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(innerW).
		Render(body)
}

// renderHelpPanel renders the contextual keys panel at the bottom of a tab.
// width is the total available width.
func renderHelpPanel(keys string, width int) string {
	innerW := width - 4
	if innerW < 1 {
		innerW = 1
	}
	return lipgloss.NewStyle().
		Border(BorderNormal).
		BorderForeground(ColorBorder).
		Foreground(ColorMuted).
		Padding(0, 1).
		Width(innerW).
		Render(" " + keys)
}

// renderEmptyState renders a centered empty-state message inside a panel body.
func renderEmptyState(icon, line1, line2 string, width int) string {
	pad := max((width-len([]rune(line1)))/2, 0)
	spaces := strings.Repeat(" ", pad)
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(spaces + StyleMuted.Render(icon+"  "+line1) + "\n")
	if line2 != "" {
		sb.WriteString(spaces + StyleMuted.Render(line2) + "\n")
	}
	sb.WriteString("\n\n")
	return sb.String()
}
