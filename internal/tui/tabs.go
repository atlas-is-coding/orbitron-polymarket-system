package tui

import (
	"fmt"
	"strings"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

// TabID identifies a tab.
type TabID int

const (
	TabOverview    TabID = iota
	TabTrading           // Orders + Positions merged
	TabWallets
	TabCopytrading
	TabMarkets
	TabLogs
	TabSettings
	tabCount // sentinel
)

var tabKeys = []string{"1", "2", "3", "4", "5", "6", "7"}

// tabNames returns tab display names in the current locale.
func tabNames() []string {
	t := i18n.T()
	return []string{
		t.TabOverview,
		t.TabTrading,
		t.TabWallets,
		t.TabCopytrading,
		t.TabMarkets,
		t.TabLogs,
		t.TabSettings,
	}
}

// tabIcons are unicode prefixes for each tab.
var tabIcons = []string{"◈", "⊹", "◎", "⟳", "⊛", "≡", "⚙"}

// RenderTabBar renders a clean tab bar with the active tab highlighted.
func RenderTabBar(active TabID, width int) string {
	names := tabNames()
	var sb strings.Builder
	for i, name := range names {
		label := fmt.Sprintf(" %s %s:%s ", tabIcons[i], tabKeys[i], name)
		if TabID(i) == active {
			sb.WriteString(StyleTabActive.Render(label))
		} else {
			sb.WriteString(StyleTabInactive.Render(label))
		}
	}
	tabs := StyleTabBar.Width(width).Render(sb.String())

	// Decorative separator line
	line := StyleTabBarLine.Width(width).Render(strings.Repeat("─", width))
	return tabs + "\n" + line
}
