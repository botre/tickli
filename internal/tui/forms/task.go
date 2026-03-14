package forms

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// TaskFormResult holds the values collected from the task creation form.
type TaskFormResult struct {
	Title    string
	Content  string
	Priority task.Priority
	Date     string
	Tags     string
	Project  string // selected project ID
}

// RunTaskCreateForm displays an interactive task creation form using Huh.
// projects is optional — when provided, a project selector is shown.
// knownTags is optional — when provided, a multi-select for tags is shown.
func RunTaskCreateForm(t theme.Theme, defaults TaskFormResult, projects []types.Project, knownTags []string) (*TaskFormResult, error) {
	result := &TaskFormResult{
		Title:    defaults.Title,
		Content:  defaults.Content,
		Priority: defaults.Priority,
		Date:     defaults.Date,
		Tags:     defaults.Tags,
		Project:  defaults.Project,
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

	// Parse existing tags for pre-selection
	var selectedTags []string
	if defaults.Tags != "" {
		for _, tag := range strings.Split(defaults.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				selectedTags = append(selectedTags, tag)
			}
		}
	}

	var extraTags string

	var fields []huh.Field

	// Project selector if projects are provided and no project pre-selected
	if len(projects) > 0 && result.Project == "" {
		opts := make([]huh.Option[string], len(projects))
		for i, p := range projects {
			opts[i] = huh.NewOption(p.Name, p.ID)
		}
		fields = append(fields,
			huh.NewSelect[string]().
				Title("Project").
				Description("Which project does this belong to?").
				Options(opts...).
				Value(&result.Project),
		)
	}

	fields = append(fields,
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
				huh.NewOption("🔵 Low", "low"),
				huh.NewOption("🟡 Medium", "medium"),
				huh.NewOption("🔴 High", "high"),
			).
			Value(&priorityStr),

		huh.NewInput().
			Title("Due Date").
			Description("e.g. 'tomorrow at 2pm', 'next Friday', '2025-03-20'").
			Placeholder("Leave empty for no due date").
			Value(&result.Date),
	)

	// Tags: multi-select from known tags, plus free-text for new ones
	if len(knownTags) > 0 {
		tagOpts := make([]huh.Option[string], len(knownTags))
		for i, tag := range knownTags {
			tagOpts[i] = huh.NewOption(tag, tag).Selected(contains(selectedTags, tag))
		}
		fields = append(fields,
			huh.NewMultiSelect[string]().
				Title("Tags").
				Description("Select existing tags").
				Options(tagOpts...).
				Value(&selectedTags),
			huh.NewInput().
				Title("New Tags").
				Description("Add new tags (comma-separated)").
				Placeholder("new-tag, another-tag…").
				Value(&extraTags),
		)
	} else {
		fields = append(fields,
			huh.NewInput().
				Title("Tags").
				Description("Comma-separated").
				Placeholder("work, important, meeting…").
				Value(&result.Tags),
		)
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
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

	// Merge selected + extra tags
	if len(knownTags) > 0 {
		allTags := append([]string{}, selectedTags...)
		if extraTags != "" {
			for _, tag := range strings.Split(extraTags, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" && !contains(allTags, tag) {
					allTags = append(allTags, tag)
				}
			}
		}
		result.Tags = strings.Join(allTags, ", ")
	}

	return result, nil
}

// RunTaskUpdateForm displays an interactive task update form.
func RunTaskUpdateForm(t theme.Theme, defaults TaskFormResult, knownTags []string) (*TaskFormResult, error) {
	return RunTaskCreateForm(t, defaults, nil, knownTags)
}

// CollectTags extracts unique sorted tags from a list of tasks.
func CollectTags(tasks []types.Task) []string {
	seen := make(map[string]bool)
	for _, t := range tasks {
		for _, tag := range t.Tags {
			seen[tag] = true
		}
	}
	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
