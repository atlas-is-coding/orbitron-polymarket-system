package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// WalletProvider is implemented by *wallet.Manager (see internal/wallet).
// Using an interface here avoids an import cycle (wallet imports tui).
type WalletProvider interface {
	// WalletIDs returns the list of wallet identifiers to display.
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
