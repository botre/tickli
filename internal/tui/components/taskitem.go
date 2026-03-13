package components

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// TaskItem wraps a task for display in a Bubbles list.
type TaskItem struct {
	Task        types.Task
	ProjectName string
}

func (t TaskItem) Title() string       { return t.Task.Title }
func (t TaskItem) Description() string { return t.Task.Content }
func (t TaskItem) FilterValue() string { return t.Task.Title }

// TaskDelegate renders task items in the list.
type TaskDelegate struct {
	Theme       theme.Theme
	ShowProject bool
}

func NewTaskDelegate(t theme.Theme, showProject bool) TaskDelegate {
	return TaskDelegate{Theme: t, ShowProject: showProject}
}

func (d TaskDelegate) Height() int                         { return 2 }
func (d TaskDelegate) Spacing() int                        { return 0 }
func (d TaskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d TaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	t, ok := item.(TaskItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	width := m.Width() - 4

	// Status icon
	var statusIcon string
	if t.Task.Status == task.StatusComplete {
		statusIcon = d.Theme.StatusComplete.Render(theme.IconComplete)
	} else {
		statusIcon = d.Theme.StatusPending.Render(theme.IconPending)
	}

	// Priority flag
	priorityFlag := renderPriority(d.Theme, t.Task.Priority)

	// Title
	title := t.Task.Title
	if isSelected {
		title = d.Theme.SelectedItem.Render(title)
	} else if t.Task.Status == task.StatusComplete {
		title = d.Theme.DimmedItem.Render(title)
	} else {
		title = d.Theme.Body.Render(title)
	}

	// First line: status + priority + title
	firstLine := fmt.Sprintf(" %s %s %s", statusIcon, priorityFlag, title)

	// Second line: metadata (due date, project, tags)
	var meta []string

	if d.ShowProject && t.ProjectName != "" {
		meta = append(meta, d.Theme.TaskProject.Render(t.ProjectName))
	}

	due := time.Time(t.Task.DueDate)
	if !due.IsZero() {
		dueStr := t.Task.DueDate.Humanize()
		if due.Before(time.Now()) && t.Task.Status != task.StatusComplete {
			dueStr = d.Theme.TaskDueOverdue.Render(theme.IconCalendar + " " + dueStr)
		} else {
			dueStr = d.Theme.TaskDue.Render(theme.IconCalendar + " " + dueStr)
		}
		meta = append(meta, dueStr)
	}

	for _, tag := range t.Task.Tags {
		meta = append(meta, d.Theme.TaskTag.Render(theme.IconTag+tag))
	}

	secondLine := ""
	if len(meta) > 0 {
		secondLine = "     " + strings.Join(meta, d.Theme.Muted.Render("  "+theme.IconDot+"  "))
	}

	// Cursor indicator
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(d.Theme.Palette.Primary).Render(theme.IconSelected + " ")
	}

	// Truncate to width
	if width > 0 {
		firstLine = truncate(firstLine, width)
		if secondLine != "" {
			secondLine = truncate(secondLine, width)
		}
	}

	output := cursor + firstLine
	if secondLine != "" {
		output += "\n" + cursor + secondLine
	}

	fmt.Fprint(w, output)
}

// renderPriority returns a styled priority flag.
func renderPriority(t theme.Theme, p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return t.PriorityHigh.Render(theme.IconPriority)
	case task.PriorityMedium:
		return t.PriorityMedium.Render(theme.IconPriority)
	case task.PriorityLow:
		return t.PriorityLow.Render(theme.IconPriority)
	default:
		return t.PriorityNone.Render(theme.IconPriority)
	}
}

func truncate(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Simple truncation — keep it fast
	runes := []rune(s)
	if len(runes) > maxWidth-1 {
		return string(runes[:maxWidth-1]) + "…"
	}
	return s
}
