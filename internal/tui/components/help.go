package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

// KeyBinding represents a single key binding for help display.
type KeyBinding struct {
	Key  string
	Help string
}

// HelpBar renders a compact help bar showing available keybindings.
type HelpBar struct {
	Theme    theme.Theme
	Width    int
	Bindings []KeyBinding
}

func NewHelpBar(t theme.Theme) HelpBar {
	return HelpBar{Theme: t}
}

func (h HelpBar) View() string {
	if len(h.Bindings) == 0 || h.Width <= 0 {
		return ""
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(h.Theme.Palette.Accent).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(h.Theme.Palette.Muted)

	sep := helpStyle.Render("  " + theme.IconDot + "  ")

	var parts []string
	for _, b := range h.Bindings {
		parts = append(parts, keyStyle.Render(b.Key)+" "+helpStyle.Render(b.Help))
	}

	line := strings.Join(parts, sep)

	if lipgloss.Width(line) > h.Width {
		// Truncate to fit
		line = ""
		for i, b := range h.Bindings {
			part := keyStyle.Render(b.Key) + " " + helpStyle.Render(b.Help)
			if i > 0 {
				part = sep + part
			}
			if lipgloss.Width(line+part) > h.Width-3 {
				line += helpStyle.Render(" …")
				break
			}
			line += part
		}
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Width(h.Width).
		Render(line)
}
