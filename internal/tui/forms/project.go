package forms

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/botre/tickli/internal/tui/theme"
)

// ProjectFormResult holds the values collected from the project form.
type ProjectFormResult struct {
	Name     string
	Color    string
	ViewMode string
	Kind     string
}

// RunProjectCreateForm displays an interactive project creation form.
func RunProjectCreateForm(t theme.Theme, defaults ProjectFormResult) (*ProjectFormResult, error) {
	result := &ProjectFormResult{
		Name:     defaults.Name,
		Color:    defaults.Color,
		ViewMode: defaults.ViewMode,
		Kind:     defaults.Kind,
	}

	if result.ViewMode == "" {
		result.ViewMode = "list"
	}
	if result.Kind == "" {
		result.Kind = "TASK"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("Give your project a name").
				Placeholder("Enter project name…").
				Value(&result.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Color").
				Description("Choose a project color").
				Options(
					huh.NewOption("Red", "#EC6665"),
					huh.NewOption("Orange", "#F2B04A"),
					huh.NewOption("Yellow", "#FFD866"),
					huh.NewOption("Green", "#5CD0A7"),
					huh.NewOption("Cyan", "#9BECEC"),
					huh.NewOption("Blue", "#4AA6EF"),
					huh.NewOption("Purple", "#CF66F6"),
					huh.NewOption("Pink", "#EC70A5"),
					huh.NewOption("White", "#FDF8DC"),
				).
				Value(&result.Color),

			huh.NewSelect[string]().
				Title("View Mode").
				Description("How should this project be displayed?").
				Options(
					huh.NewOption("List", "list"),
					huh.NewOption("Kanban", "kanban"),
					huh.NewOption("Timeline", "timeline"),
				).
				Value(&result.ViewMode),

			huh.NewSelect[string]().
				Title("Type").
				Description("Project type").
				Options(
					huh.NewOption("Task", "TASK"),
					huh.NewOption("Note", "NOTE"),
				).
				Value(&result.Kind),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return result, nil
}
