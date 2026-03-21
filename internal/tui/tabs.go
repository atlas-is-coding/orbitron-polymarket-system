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

// sidebarWidth is removed for horizontal top navigation.
const topBarHeight = 3

// breakpoint returns the responsive layout tier for a given terminal width.
// Tiers: "tiny" ≤80, "mobile" ≤100, "standard" ≤140, "large" ≤180, "xl" >180.
func breakpoint(w int) string {
	switch {
	case w <= 80:
		return "tiny"
	case w <= 100:
		return "mobile"
	case w <= 140:
		return "standard"
	case w <= 180:
		return "large"
	default:
		return "xl"
	}
}

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

// RenderTopBar renders the horizontal top navigation bar.
func RenderTopBar(active TabID, width int) string {
	names := tabNames()
	var parts []string

	for i, name := range names {
		label := fmt.Sprintf(" %s %s: %s ", tabIcons[i], tabKeys[i], name)
		if TabID(i) == active {
			parts = append(parts, StyleTabActive.Render(label))
		} else {
			parts = append(parts, StyleTabInactive.Render(label))
		}
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	
	// Add a subtle bottom line for the bar
	line := StyleTabBarLine.Width(width).Render(strings.Repeat("━", width))
	
	return lipgloss.JoinVertical(lipgloss.Left, tabs, line)
}

// renderPanel renders a spec-compliant content panel.
// active=true uses ColorPrimary border; false uses ColorPrimaryDim.
func renderPanel(title, content string, width int, active bool) string {
	panelStyle := StylePanelInactive
	if active {
		panelStyle = StylePanelActive
	}
	innerW := width - 4 // 2 border + 2 padding cells each side
	if innerW < 1 {
		innerW = 1
	}
	var body string
	if title != "" {
		header := StylePageTitle.Render(title)
		body = header + "\n" + content
	} else {
		body = content
	}
	return panelStyle.Width(innerW).Render(body)
}


// renderHelpPanel renders the bottom keybind bar.
// keys format: "q=quit | Tab=switch | ?=help"
// Keys (before =) are rendered in ColorBright; separators in ColorMuted.
func renderHelpPanel(keys string, width int) string {
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	var sb strings.Builder
	parts := strings.Split(keys, " | ")
	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			sb.WriteString(StyleValue.Render(kv[0]))
			sb.WriteString(StyleMuted.Render("="))
			sb.WriteString(StyleBody.Render(kv[1]))
		} else {
			sb.WriteString(StyleBody.Render(part))
		}
		if i < len(parts)-1 {
			sb.WriteString(StyleMuted.Render(" | "))
		}
	}
	helpStyle := lipgloss.NewStyle().
		Background(ColorBgLight).
		BorderTop(true).
		BorderStyle(lipgloss.Border{Top: "─"}).
		BorderForeground(ColorPrimary).
		Width(innerW).
		Padding(0, 1)
	return helpStyle.Render(sb.String())
}

// renderEmptyState renders a centered empty-state (spec §7.8).
// icon: "○", line1: title text (bold), line2: subtitle (muted).
func renderEmptyState(icon, line1, line2 string, width int) string {
	center := func(s string) string {
		pad := max((width-lipgloss.Width(s))/2, 0)
		return strings.Repeat(" ", pad) + s
	}
	iconStr := center(StyleMuted.Render(icon))
	titleStr := center(StyleSectionHead.Render(line1))
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(iconStr + "\n\n")
	sb.WriteString(titleStr + "\n")
	if line2 != "" {
		sb.WriteString(center(StyleMuted.Render(line2)) + "\n")
	}
	sb.WriteString("\n\n")
	return sb.String()
}

