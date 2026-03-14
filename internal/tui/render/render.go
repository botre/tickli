// Package render provides Lip Gloss styled output for CLI commands.
// This replaces the plain text formatters that used gookit/color.
package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// Renderer produces styled output for CLI commands.
type Renderer struct {
	Theme theme.Theme
}

// New creates a new Renderer with the default theme.
func New() Renderer {
	return Renderer{Theme: theme.Default()}
}

// TaskLine renders a single task as a one-line summary.
func (r Renderer) TaskLine(t types.Task) string {
	status := r.statusIcon(t.Status)
	priority := r.priorityFlag(t.Priority)
	title := r.Theme.Body.Render(t.Title)

	if t.Status == task.StatusComplete {
		title = r.Theme.DimmedItem.Render(t.Title)
	}

	return fmt.Sprintf("%s %s %s", status, priority, title)
}

// TaskLineWithProject renders a task line with project context.
func (r Renderer) TaskLineWithProject(t types.Task, projectName string) string {
	line := r.TaskLine(t)
	if projectName != "" {
		proj := r.Theme.TaskProject.Render("[" + projectName + "]")
		line = proj + " " + line
	}
	return line
}

// TaskDetail renders a full task detail view.
func (r Renderer) TaskDetail(t types.Task, projectName string) string {
	var sections []string

	// Title
	sections = append(sections, r.Theme.TaskTitle.Render(t.Title))

	// Status
	sections = append(sections, r.statusIcon(t.Status)+" "+r.statusLabel(t.Status))

	// Priority
	sections = append(sections, r.Theme.Muted.Render("Priority  ")+r.priorityLabel(t.Priority))

	// Project
	if projectName != "" {
		sections = append(sections, r.Theme.Muted.Render("Project   ")+r.Theme.TaskProject.Render(projectName))
	}

	// Dates
	if s := time.Time(t.StartDate); !s.IsZero() {
		sections = append(sections, r.Theme.Muted.Render("Start     ")+r.Theme.TaskDue.Render(t.StartDate.Humanize()))
	}
	if d := time.Time(t.DueDate); !d.IsZero() {
		dueStr := t.DueDate.Humanize()
		if d.Before(time.Now()) && t.Status != task.StatusComplete {
			sections = append(sections, r.Theme.Muted.Render("Due       ")+r.Theme.TaskDueOverdue.Render(dueStr+" (overdue)"))
		} else {
			sections = append(sections, r.Theme.Muted.Render("Due       ")+r.Theme.TaskDue.Render(dueStr))
		}
	}
	if c := time.Time(t.CompletedTime); !c.IsZero() {
		sections = append(sections, r.Theme.Muted.Render("Done      ")+r.Theme.StatusComplete.Render(t.CompletedTime.Humanize()))
	}

	// Tags
	if len(t.Tags) > 0 {
		var tags []string
		for _, tag := range t.Tags {
			tags = append(tags, r.Theme.TaskTag.Render(theme.IconTag+tag))
		}
		sections = append(sections, r.Theme.Muted.Render("Tags      ")+strings.Join(tags, " "))
	}

	// Content
	if t.Content != "" {
		divider := r.Theme.Divider.Render(strings.Repeat(theme.IconDivider, 40))
		sections = append(sections, divider)
		sections = append(sections, r.Theme.TaskContent.Render(t.Content))
	}

	// ID
	sections = append(sections, r.Theme.Muted.Render("ID        ")+r.Theme.Muted.Render(t.ID))

	return r.Theme.Card.Render(strings.Join(sections, "\n"))
}

// ProjectDetail renders a full project detail view.
func (r Renderer) ProjectDetail(p types.Project) string {
	var sections []string

	// Color bar
	colorStr := p.Color.String()
	if colorStr != "" {
		bar := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorStr)).
			Render("████████████████████████")
		sections = append(sections, bar)
	}

	// Name
	sections = append(sections, r.Theme.Title.Render(p.Name))

	// Status
	status := r.Theme.StatusComplete.Render("Open")
	if p.Closed {
		status = r.Theme.DimmedItem.Render("Closed")
	}
	sections = append(sections, r.Theme.Muted.Render("Status    ")+status)

	// Type
	sections = append(sections, r.Theme.Muted.Render("Type      ")+r.Theme.Body.Render(p.Kind.String()))

	// View mode
	sections = append(sections, r.Theme.Muted.Render("View      ")+r.Theme.Body.Render(p.ViewMode.String()))

	// ID
	sections = append(sections, r.Theme.Muted.Render("ID        ")+r.Theme.Muted.Render(p.ID))

	return r.Theme.Card.Render(strings.Join(sections, "\n"))
}

// ProjectLine renders a single project as a one-line summary.
func (r Renderer) ProjectLine(p types.Project) string {
	colorStr := p.Color.String()
	var dot string
	if colorStr != "" {
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color(colorStr)).Render(theme.IconProject)
	} else {
		dot = r.Theme.Muted.Render(theme.IconProject)
	}
	return fmt.Sprintf("%s %s", dot, r.Theme.Body.Render(p.Name))
}

// TaskList renders a list of tasks with indices.
func (r Renderer) TaskList(tasks []types.Task) string {
	if len(tasks) == 0 {
		return r.Theme.Muted.Render("No tasks found")
	}
	var lines []string
	for _, t := range tasks {
		lines = append(lines, r.TaskLine(t))
	}
	return strings.Join(lines, "\n")
}

// TaskListWithProjects renders a list of tasks with project names.
func (r Renderer) TaskListWithProjects(tasks []types.Task, projectNames []string) string {
	if len(tasks) == 0 {
		return r.Theme.Muted.Render("No tasks found")
	}
	var lines []string
	for i, t := range tasks {
		name := ""
		if i < len(projectNames) {
			name = projectNames[i]
		}
		lines = append(lines, r.TaskLineWithProject(t, name))
	}
	return strings.Join(lines, "\n")
}

// SuccessMessage renders a success message.
func (r Renderer) SuccessMessage(msg string) string {
	return r.Theme.SuccessMessage.Render(theme.IconSuccess + " " + msg)
}

// ErrorMessage renders an error message.
func (r Renderer) ErrorMessage(msg string) string {
	return r.Theme.ErrorMessage.Render(theme.IconError + " " + msg)
}

// InfoMessage renders an info message.
func (r Renderer) InfoMessage(msg string) string {
	return r.Theme.InfoMessage.Render(theme.IconInfo + " " + msg)
}

// Helpers

func (r Renderer) statusIcon(s task.Status) string {
	switch s {
	case task.StatusComplete:
		return r.Theme.StatusComplete.Render(theme.IconComplete)
	default:
		return r.Theme.StatusPending.Render(theme.IconPending)
	}
}

func (r Renderer) statusLabel(s task.Status) string {
	switch s {
	case task.StatusComplete:
		return r.Theme.StatusComplete.Render("Done")
	default:
		return r.Theme.StatusPending.Render("Todo")
	}
}

func (r Renderer) priorityFlag(p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return r.Theme.PriorityHigh.Render(theme.IconPriority)
	case task.PriorityMedium:
		return r.Theme.PriorityMedium.Render(theme.IconPriority)
	case task.PriorityLow:
		return r.Theme.PriorityLow.Render(theme.IconPriority)
	default:
		return r.Theme.PriorityNone.Render(theme.IconPriority)
	}
}

func (r Renderer) priorityLabel(p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return r.Theme.PriorityHigh.Render(theme.IconPriority + " High")
	case task.PriorityMedium:
		return r.Theme.PriorityMedium.Render(theme.IconPriority + " Medium")
	case task.PriorityLow:
		return r.Theme.PriorityLow.Render(theme.IconPriority + " Low")
	default:
		return r.Theme.PriorityNone.Render(theme.IconPriority + " None")
	}
}
