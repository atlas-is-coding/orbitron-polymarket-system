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
	TabOrders
	TabPositions
	TabWallets     // NEW — index 3
	TabCopytrading // was 3, now 4
	TabLogs        // was 4, now 5
	TabSettings    // was 5, now 6
	tabCount       // sentinel
)

var tabKeys = []string{"1", "2", "3", "4", "5", "6", "7"}

// tabNames returns tab display names in the current locale.
func tabNames() []string {
	t := i18n.T()
	return []string{
		t.TabOverview,
		t.TabOrders,
		t.TabPositions,
		t.TabWallets,
		t.TabCopytrading,
		t.TabLogs,
		t.TabSettings,
	}
}

// RenderTabBar renders the tab bar with the active tab highlighted.
func RenderTabBar(active TabID, width int) string {
	names := tabNames()
	var sb strings.Builder
	for i, name := range names {
		label := fmt.Sprintf(" %s:%s ", tabKeys[i], name)
		if TabID(i) == active {
			sb.WriteString(StyleTabActive.Render(label))
		} else {
			sb.WriteString(StyleTabInactive.Render(label))
		}
		if i < len(names)-1 {
			sb.WriteString(StyleTabSep.String())
		}
	}
	return StyleTabBar.Width(width).Render(sb.String())
}
