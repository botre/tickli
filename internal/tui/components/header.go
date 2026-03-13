package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
)

// Header renders the top bar of the TUI app.
type Header struct {
	Theme      theme.Theme
	Width      int
	Title      string
	Breadcrumb []string
}

func NewHeader(t theme.Theme) Header {
	return Header{Theme: t, Title: "tickli"}
}

func (h Header) View() string {
	if h.Width <= 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Background(h.Theme.Palette.Primary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2)

	title := titleStyle.Render(h.Title)
	titleW := lipgloss.Width(title)

	crumb := ""
	if len(h.Breadcrumb) > 0 {
		sep := h.Theme.Muted.Render(" " + theme.IconArrowRight + " ")
		parts := make([]string, len(h.Breadcrumb))
		for i, b := range h.Breadcrumb {
			if i == len(h.Breadcrumb)-1 {
				parts[i] = h.Theme.Bold.Render(b)
			} else {
				parts[i] = h.Theme.Breadcrumb.Render(b)
			}
		}
		crumb = " " + strings.Join(parts, sep)
	}

	crumbW := lipgloss.Width(crumb)
	gap := h.Width - titleW - crumbW
	if gap < 0 {
		gap = 0
	}

	fill := strings.Repeat(" ", gap)

	return title + crumb + fill
}
