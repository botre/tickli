package picker

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

// fuzzyMatch checks whether all characters in query appear in s in order (case-insensitive).
func fuzzyMatch(s, query string) bool {
	s = strings.ToLower(s)
	query = strings.ToLower(query)
	qi := 0
	for _, r := range s {
		if qi < len(query) && byte(r) == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

// fuzzyMatchRow returns true if any cell in the row fuzzy-matches the query.
func fuzzyMatchRow(row table.Row, query string) bool {
	for _, cell := range row {
		if fuzzyMatch(cell, query) {
			return true
		}
	}
	return false
}

func tableStyles(t theme.Theme) table.Styles {
	var s table.Styles
	s.Header = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.SubText))).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(string(t.Palette.Subtle))).
		BorderBottom(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.Primary))).
		Bold(true)
	return s
}
