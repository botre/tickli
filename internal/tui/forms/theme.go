package forms

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

// huhTheme builds a Huh form theme from our app palette.
func huhTheme(t theme.Theme) *huh.Theme {
	th := huh.ThemeBase()
	p := t.Palette

	th.Focused.Base = th.Focused.Base.BorderForeground(p.Subtle)
	th.Focused.Card = th.Focused.Base
	th.Focused.Title = th.Focused.Title.Foreground(p.Primary).Bold(true)
	th.Focused.NoteTitle = th.Focused.NoteTitle.Foreground(p.Primary).Bold(true).MarginBottom(1)
	th.Focused.Description = th.Focused.Description.Foreground(p.SubText)
	th.Focused.ErrorIndicator = th.Focused.ErrorIndicator.Foreground(p.Error)
	th.Focused.ErrorMessage = th.Focused.ErrorMessage.Foreground(p.Error)
	th.Focused.SelectSelector = th.Focused.SelectSelector.Foreground(p.Accent)
	th.Focused.NextIndicator = th.Focused.NextIndicator.Foreground(p.Accent)
	th.Focused.PrevIndicator = th.Focused.PrevIndicator.Foreground(p.Accent)
	th.Focused.Option = th.Focused.Option.Foreground(p.Text)
	th.Focused.MultiSelectSelector = th.Focused.MultiSelectSelector.Foreground(p.Accent)
	th.Focused.SelectedOption = th.Focused.SelectedOption.Foreground(p.Success)
	th.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(p.Success).SetString("✓ ")
	th.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(p.Muted).SetString("○ ")
	th.Focused.UnselectedOption = th.Focused.UnselectedOption.Foreground(p.SubText)
	th.Focused.FocusedButton = th.Focused.FocusedButton.Foreground(p.Base).Background(p.Primary)
	th.Focused.BlurredButton = th.Focused.BlurredButton.Foreground(p.Text).Background(p.Surface)
	th.Focused.Next = th.Focused.FocusedButton
	th.Focused.TextInput.Cursor = th.Focused.TextInput.Cursor.Foreground(p.Primary)
	th.Focused.TextInput.Placeholder = th.Focused.TextInput.Placeholder.Foreground(p.Muted)
	th.Focused.TextInput.Prompt = th.Focused.TextInput.Prompt.Foreground(p.Accent)

	th.Blurred = th.Focused
	th.Blurred.Base = th.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	th.Blurred.Title = th.Blurred.Title.Foreground(p.SubText)
	th.Blurred.NextIndicator = lipgloss.NewStyle()
	th.Blurred.PrevIndicator = lipgloss.NewStyle()

	return th
}
