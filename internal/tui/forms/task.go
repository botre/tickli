package forms

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types/task"
)

// TaskFormResult holds the values collected from the task creation form.
type TaskFormResult struct {
	Title    string
	Content  string
	Priority task.Priority
	Date     string
	Tags     string
}

// RunTaskCreateForm displays an interactive task creation form using Huh.
func RunTaskCreateForm(t theme.Theme, defaults TaskFormResult) (*TaskFormResult, error) {
	result := &TaskFormResult{
		Title:    defaults.Title,
		Content:  defaults.Content,
		Priority: defaults.Priority,
		Date:     defaults.Date,
		Tags:     defaults.Tags,
	}

	var priorityStr string
	switch defaults.Priority {
	case task.PriorityHigh:
		priorityStr = "high"
	case task.PriorityMedium:
		priorityStr = "medium"
	case task.PriorityLow:
		priorityStr = "low"
	default:
		priorityStr = "none"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Description("What needs to be done?").
				Placeholder("Enter task title…").
				Value(&result.Title).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Content").
				Description("Additional details (optional)").
				Placeholder("Add notes, links, or context…").
				Value(&result.Content).
				Lines(3),

			huh.NewSelect[string]().
				Title("Priority").
				Description("How important is this?").
				Options(
					huh.NewOption("None", "none"),
					huh.NewOption("Low", "low"),
					huh.NewOption("Medium", "medium"),
					huh.NewOption("High", "high"),
				).
				Value(&priorityStr),

			huh.NewInput().
				Title("Date").
				Description("When is it due? (e.g. 'tomorrow at 2pm', 'next Friday')").
				Placeholder("Enter date or leave empty…").
				Value(&result.Date),

			huh.NewInput().
				Title("Tags").
				Description("Comma-separated tags").
				Placeholder("work, important, meeting…").
				Value(&result.Tags),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Convert priority string back
	switch priorityStr {
	case "high":
		result.Priority = task.PriorityHigh
	case "medium":
		result.Priority = task.PriorityMedium
	case "low":
		result.Priority = task.PriorityLow
	default:
		result.Priority = task.PriorityNone
	}

	return result, nil
}

// RunTaskUpdateForm displays an interactive task update form.
func RunTaskUpdateForm(t theme.Theme, defaults TaskFormResult) (*TaskFormResult, error) {
	return RunTaskCreateForm(t, defaults)
}
