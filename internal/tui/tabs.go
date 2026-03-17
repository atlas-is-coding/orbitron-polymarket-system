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

// renderPanel renders a minimalist content box.
func renderPanel(title, content string, width int, active bool) string {
	panelStyle := StylePanelInactive
	titleStyle := StylePanelTitle
	if active {
		panelStyle = StylePanelActive
		titleStyle = StylePanelTitleActive
	}

	innerW := width - 2 // smaller padding for minimalist look
	if innerW < 1 {
		innerW = 1
	}

	var body string
	if title != "" {
		header := titleStyle.Render(fmt.Sprintf(" %s ", title))
		body = header + "\n" + content
	} else {
		body = content
	}

	return panelStyle.Width(innerW).Render(body)
}


// renderHelpPanel renders the contextual keys panel at the bottom of a tab.
func renderHelpPanel(keys string, width int) string {
	innerW := width - 4
	if innerW < 1 {
		innerW = 1
	}
	// Highlight keys using basic replacement to make it look nicer
	formattedKeys := strings.ReplaceAll(keys, "[", StyleAccent.Render("["))
	formattedKeys = strings.ReplaceAll(formattedKeys, "]", StyleAccent.Render("]"))

	return StylePanelHelp.Width(innerW).Render(" " + formattedKeys)
}

// renderEmptyState renders a centered empty-state message inside a panel body.
func renderEmptyState(icon, line1, line2 string, width int) string {
	pad := max((width-len([]rune(line1)))/2, 0)
	spaces := strings.Repeat(" ", pad)
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(spaces + StyleMuted.Render(icon) + " " + StyleFgDim.Render(line1) + "\n")
	if line2 != "" {
		pad2 := max((width-len([]rune(line2)))/2, 0)
		spaces2 := strings.Repeat(" ", pad2)
		sb.WriteString(spaces2 + StyleMuted.Render(line2) + "\n")
	}
	sb.WriteString("\n\n")
	return sb.String()
}

