package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the TUI app.
type KeyMap struct {
	Quit       key.Binding
	Back       key.Binding
	Enter      key.Binding
	Tab        key.Binding
	Help       key.Binding
	Refresh    key.Binding
	Filter     key.Binding
	Complete   key.Binding
	Create     key.Binding
	Delete     key.Binding
	Today      key.Binding
	Tomorrow   key.Binding
	Week       key.Binding
	Inbox      key.Binding
	All        key.Binding
	Projects   key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "select"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Complete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "complete"),
		),
		Create: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Today: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "today"),
		),
		Tomorrow: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "tomorrow"),
		),
		Week: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "week"),
		),
		Inbox: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "inbox"),
		),
		All: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "all"),
		),
		Projects: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "projects"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "l"),
			key.WithHelp("tab/l", "next view"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "h"),
			key.WithHelp("S-tab/h", "prev view"),
		),
	}
}
