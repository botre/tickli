package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

// StatusBar renders a bottom status bar with contextual information.
type StatusBar struct {
	Theme  theme.Theme
	Width  int
	Left   string
	Center string
	Right  string
}

func NewStatusBar(t theme.Theme) StatusBar {
	return StatusBar{Theme: t}
}

func (s StatusBar) View() string {
	if s.Width <= 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Background(s.Theme.Palette.Surface).
		Foreground(s.Theme.Palette.SubText).
		Width(s.Width)

	leftStyle := lipgloss.NewStyle().
		Background(s.Theme.Palette.Primary).
		Foreground(s.Theme.Palette.Base).
		Bold(true).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Background(s.Theme.Palette.Surface).
		Foreground(s.Theme.Palette.Muted).
		Padding(0, 1).
		Align(lipgloss.Right)

	centerStyle := lipgloss.NewStyle().
		Background(s.Theme.Palette.Surface).
		Foreground(s.Theme.Palette.SubText).
		Padding(0, 1)

	left := leftStyle.Render(s.Left)
	right := rightStyle.Render(s.Right)
	center := centerStyle.Render(s.Center)

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	centerW := lipgloss.Width(center)

	// Fill remaining space
	gap := s.Width - leftW - centerW - rightW
	if gap < 0 {
		gap = 0
	}

	fill := lipgloss.NewStyle().
		Background(s.Theme.Palette.Surface).
		Render(strings.Repeat(" ", gap))

	bar := left + center + fill + right

	return style.Render(bar)
}
