package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// WalletProvider is a minimal interface that the WalletsModel uses to access
// wallet data. It is implemented by *wallet.Manager but avoids an import cycle
// (internal/wallet already imports internal/tui).
type WalletProvider interface {
	// Wallets returns a slice of wallet summaries suitable for display.
	// The actual type is returned as interface{} to break the cycle;
	// the full implementation in Task 6 will use concrete types.
	WalletIDs() []string
}

// WalletsModel is the Wallets tab sub-model.
// Full implementation in Task 6.
type WalletsModel struct {
	wm      WalletProvider
	cfgPath string
	width   int
	height  int
}

// NewWalletsModel creates a new WalletsModel stub.
func NewWalletsModel(wm WalletProvider, cfgPath string, width, height int) WalletsModel {
	return WalletsModel{
		wm:      wm,
		cfgPath: cfgPath,
		width:   width,
		height:  height,
	}
}

func (m WalletsModel) Init() tea.Cmd { return nil }

func (m WalletsModel) Update(msg tea.Msg) (WalletsModel, tea.Cmd) {
	return m, nil
}

func (m WalletsModel) View() string {
	return "Wallets tab — coming soon"
}

// IsEditing reports whether a form is currently being edited.
func (m WalletsModel) IsEditing() bool { return false }
