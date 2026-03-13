package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// RenderTaskDetail returns a beautifully formatted task detail view.
func RenderTaskDetail(t theme.Theme, tsk types.Task, projectName string, width int) string {
	if width <= 0 {
		width = 80
	}

	contentWidth := width - 6 // padding
	var sections []string

	// Status + Title header
	var statusIcon string
	if tsk.Status == task.StatusComplete {
		statusIcon = t.StatusComplete.Render(theme.IconComplete + " Done")
	} else {
		statusIcon = t.StatusPending.Render(theme.IconPending + " Todo")
	}

	titleBlock := lipgloss.JoinVertical(lipgloss.Left,
		t.TaskTitle.Width(contentWidth).Render(tsk.Title),
		statusIcon,
	)
	sections = append(sections, titleBlock)

	// Priority
	priorityLine := renderPriorityLabel(t, tsk.Priority)
	sections = append(sections, priorityLine)

	// Project
	if projectName != "" {
		projectLine := t.Muted.Render("Project  ") + t.TaskProject.Render(projectName)
		sections = append(sections, projectLine)
	}

	// Dates
	var dates []string
	if s := time.Time(tsk.StartDate); !s.IsZero() {
		dates = append(dates, t.Muted.Render("Start    ")+t.TaskDue.Render(tsk.StartDate.Humanize()))
	}
	if d := time.Time(tsk.DueDate); !d.IsZero() {
		dueStr := tsk.DueDate.Humanize()
		if d.Before(time.Now()) && tsk.Status != task.StatusComplete {
			dates = append(dates, t.Muted.Render("Due      ")+t.TaskDueOverdue.Render(dueStr+" (overdue)"))
		} else {
			dates = append(dates, t.Muted.Render("Due      ")+t.TaskDue.Render(dueStr))
		}
	}
	if c := time.Time(tsk.CompletedTime); !c.IsZero() {
		dates = append(dates, t.Muted.Render("Done     ")+t.StatusComplete.Render(tsk.CompletedTime.Humanize()))
	}
	if len(dates) > 0 {
		sections = append(sections, strings.Join(dates, "\n"))
	}

	// Tags
	if len(tsk.Tags) > 0 {
		var tags []string
		for _, tag := range tsk.Tags {
			tags = append(tags, t.TaskTag.Render(theme.IconTag+tag))
		}
		sections = append(sections, t.Muted.Render("Tags     ")+strings.Join(tags, " "))
	}

	// Content
	if tsk.Content != "" {
		divider := t.Divider.Render(strings.Repeat(theme.IconDivider, min(contentWidth, 40)))
		content := t.TaskContent.Width(contentWidth).Render(tsk.Content)
		sections = append(sections, divider+"\n"+content)
	}

	// Checklist items
	if len(tsk.Items) > 0 {
		divider := t.Divider.Render(strings.Repeat(theme.IconDivider, min(contentWidth, 40)))
		var items []string
		for _, item := range tsk.Items {
			icon := t.StatusPending.Render(theme.IconPending)
			if item.Status == 2 {
				icon = t.StatusComplete.Render(theme.IconComplete)
			}
			items = append(items, fmt.Sprintf("  %s %s", icon, item.Title))
		}
		sections = append(sections, divider+"\n"+strings.Join(items, "\n"))
	}

	// ID
	sections = append(sections, t.Muted.Render("ID       ")+t.Muted.Render(tsk.ID))

	content := strings.Join(sections, "\n\n")

	card := t.Card.Width(contentWidth + 4).Render(content)
	return card
}

func renderPriorityLabel(t theme.Theme, p task.Priority) string {
	label := t.Muted.Render("Priority ")
	switch p {
	case task.PriorityHigh:
		return label + t.PriorityHigh.Render(theme.IconPriority+" High")
	case task.PriorityMedium:
		return label + t.PriorityMedium.Render(theme.IconPriority+" Medium")
	case task.PriorityLow:
		return label + t.PriorityLow.Render(theme.IconPriority+" Low")
	default:
		return label + t.PriorityNone.Render(theme.IconPriority+" None")
	}
}
