package tui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap holds global keybindings for the TUI.
type GlobalKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Quit    key.Binding
}

// GlobalKeys is the default global keybinding set.
var GlobalKeys = GlobalKeyMap{
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "следующая вкладка"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "предыдущая вкладка"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "выход"),
	),
}

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
