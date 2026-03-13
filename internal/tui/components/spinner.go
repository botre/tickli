package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/botre/tickli/internal/tui/theme"
)

// LoadingSpinner wraps the Bubbles spinner with themed styling.
type LoadingSpinner struct {
	Spinner spinner.Model
	Theme   theme.Theme
	Message string
}

func NewLoadingSpinner(t theme.Theme, msg string) LoadingSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = t.Spinner
	return LoadingSpinner{
		Spinner: s,
		Theme:   t,
		Message: msg,
	}
}

func (l LoadingSpinner) Init() tea.Cmd {
	return l.Spinner.Tick
}

func (l LoadingSpinner) Update(msg tea.Msg) (LoadingSpinner, tea.Cmd) {
	var cmd tea.Cmd
	l.Spinner, cmd = l.Spinner.Update(msg)
	return l, cmd
}

func (l LoadingSpinner) View() string {
	return l.Spinner.View() + " " + l.Theme.Muted.Render(l.Message)
}
