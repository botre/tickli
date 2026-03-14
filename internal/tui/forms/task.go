package forms

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"

	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

const newProjectSentinel = "__new_project__"

// TaskFormResult holds the values collected from the task creation form.
type TaskFormResult struct {
	Title          string
	Content        string
	Priority       task.Priority
	Date           string
	Tags           string
	Project        string // selected project ID
	NewProjectName string // set when user chose to create a new project
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

	var newTags string
	var groups []*huh.Group

	// Project selection (with inline "create new" option)
	if len(projects) > 0 {
		sorted := make([]types.Project, len(projects))
		copy(sorted, projects)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].ID == result.Project {
				return true
			}
			if sorted[j].ID == result.Project {
				return false
			}
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		})
		opts := []huh.Option[string]{
			huh.NewOption("+ Create new project", newProjectSentinel),
		}
		for _, p := range sorted {
			opts = append(opts, huh.NewOption(p.Name, p.ID))
		}
		// Pre-select default project if set, otherwise first real project
		if result.Project == "" && len(sorted) > 0 {
			result.Project = sorted[0].ID
		}
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title("Project").
				Description("Type to filter").
				Options(opts...).
				Filtering(true).
				Value(&result.Project),
		))

		// Conditional: new project name input
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("New Project Name").
				Placeholder("Enter project name…").
				Value(&result.NewProjectName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return result.Project != newProjectSentinel
		}))
	}

	// Title
	groups = append(groups, huh.NewGroup(
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
	))

	// Content
	groups = append(groups, huh.NewGroup(
		huh.NewText().
			Title("Content").
			Description("Additional details (optional)").
			Placeholder("Add notes, links, or context…").
			Value(&result.Content).
			Lines(3),
	))

	// Priority
	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Priority").
			Options(
				huh.NewOption("None", "none"),
				huh.NewOption("🔵 Low", "low"),
				huh.NewOption("🟡 Medium", "medium"),
				huh.NewOption("🔴 High", "high"),
			).
			Value(&priorityStr),
	))

	// Due date
	groups = append(groups, huh.NewGroup(
		huh.NewInput().
			Title("Due Date").
			Description("e.g. 'tomorrow at 2pm', 'next Friday', '2025-03-20'").
			Placeholder("Leave empty for no due date").
			Value(&result.Date),
	))

	// Tags (multi-select + inline new tag input)
	if len(knownTags) > 0 {
		tagOpts := make([]huh.Option[string], len(knownTags))
		for i, tag := range knownTags {
			tagOpts[i] = huh.NewOption(tag, tag).Selected(contains(selectedTags, tag))
		}
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Tags").
				Description("Type to filter, enter to toggle").
				Options(tagOpts...).
				Filterable(true).
				Value(&selectedTags),
			huh.NewInput().
				Title("Add New Tags").
				Placeholder("new-tag, another…").
				Value(&newTags),
		))
	} else {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Tags").
				Description("Comma-separated").
				Placeholder("work, important, meeting…").
				Value(&result.Tags),
		))
	}

	form := huh.NewForm(groups...).WithTheme(huhTheme(t))

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

	// Merge selected + new tags
	if len(knownTags) > 0 {
		allTags := append([]string{}, selectedTags...)
		if newTags != "" {
			for _, tag := range strings.Split(newTags, ",") {
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

// CollectAllTags fetches tags from all projects concurrently with a concurrency limit.
// listTasks is a function that fetches tasks for a project ID.
func CollectAllTags(projectIDs []string, listTasks func(projectID string) ([]types.Task, error)) []string {
	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var mu sync.Mutex
	seen := make(map[string]bool)

	var wg sync.WaitGroup
	for _, pid := range projectIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			tasks, err := listTasks(id)
			if err != nil {
				return
			}
			mu.Lock()
			for _, t := range tasks {
				for _, tag := range t.Tags {
					seen[tag] = true
				}
			}
			mu.Unlock()
		}(pid)
	}
	wg.Wait()

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
