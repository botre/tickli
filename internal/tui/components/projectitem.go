package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
)

// ProjectItem wraps a project for display in a Bubbles list.
type ProjectItem struct {
	Project   types.Project
	TaskCount int
}

func (p ProjectItem) Title() string       { return p.Project.Name }
func (p ProjectItem) Description() string { return fmt.Sprintf("%d tasks", p.TaskCount) }
func (p ProjectItem) FilterValue() string { return p.Project.Name }

// ProjectDelegate renders project items in the list.
type ProjectDelegate struct {
	Theme theme.Theme
}

func NewProjectDelegate(t theme.Theme) ProjectDelegate {
	return ProjectDelegate{Theme: t}
}

func (d ProjectDelegate) Height() int                         { return 2 }
func (d ProjectDelegate) Spacing() int                        { return 0 }
func (d ProjectDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d ProjectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	p, ok := item.(ProjectItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	width := m.Width() - 4

	// Project color indicator
	var colorDot string
	colorStr := p.Project.Color.String()
	if colorStr != "" {
		colorDot = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorStr)).
			Render(theme.IconProject)
	} else {
		colorDot = d.Theme.Muted.Render(theme.IconProject)
	}

	// Name
	name := p.Project.Name
	if isSelected {
		name = d.Theme.SelectedItem.Render(name)
	} else {
		name = d.Theme.Body.Render(name)
	}

	// First line: color dot + name
	firstLine := fmt.Sprintf(" %s %s", colorDot, name)

	// Second line: task count + view mode
	info := d.Theme.Muted.Render(fmt.Sprintf("     %d tasks", p.TaskCount))
	if p.Project.ViewMode.String() != "" {
		info += d.Theme.Muted.Render("  " + theme.IconDot + "  " + p.Project.ViewMode.String())
	}

	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(d.Theme.Palette.Primary).Render(theme.IconSelected + " ")
	}

	if width > 0 {
		firstLine = truncate(firstLine, width)
		info = truncate(info, width)
	}

	fmt.Fprintf(w, "%s%s\n%s%s", cursor, firstLine, cursor, info)
}
