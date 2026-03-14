package picker

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

func tableStyles(t theme.Theme) table.Styles {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.SubText))).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(string(t.Palette.Subtle))).
		BorderBottom(true).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.Primary))).
		Bold(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Padding(0, 1)
	return s
}
