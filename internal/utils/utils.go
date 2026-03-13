package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
)

func GetProjectDescription(project types.Project) string {
	var projectStatus string
	if project.Closed {
		projectStatus = "Closed"
	} else {
		projectStatus = "Open"
	}

	projectLine := project.Color.Sprint("■■■■■■■■■■■■■■■■■■■■■■■■", project.Color)

	description := fmt.Sprintf(`
Project Details:

%s
Name: %s
ID: %s
Type: %s 
Status: %s
Group: %s

Tasks:`,
		projectLine,
		project.Name,
		project.ID,
		project.Kind,
		projectStatus,
		project.GroupID,
	)

	return description
}

func formatTickTickTime(t types.TickTickTime) string {
	if time.Time(t).IsZero() {
		return ""
	}
	return t.Humanize()
}

func GetTaskDescription(task types.Task, projectColor project.Color) string {
	projectLine := projectColor.Sprint("----------------------")

	lines := fmt.Sprintf(`
Task Details:

%s %s
%s`,
		task.Status.ColorString(),
		projectLine,
		task.Title,
	)

	if task.Content != "" {
		lines += fmt.Sprintf("\nContent: %s", task.Content)
	}
	lines += fmt.Sprintf("\nPriority: %s", task.Priority.ColorString())
	lines += fmt.Sprintf("\nProject: %s", task.ProjectID)

	if s := formatTickTickTime(task.StartDate); s != "" {
		lines += fmt.Sprintf("\nStart: %s", s)
	}
	if s := formatTickTickTime(task.DueDate); s != "" {
		lines += fmt.Sprintf("\nDue: %s", s)
	}
	if s := formatTickTickTime(task.CompletedTime); s != "" {
		lines += fmt.Sprintf("\nCompleted: %s", s)
	}
	if len(task.Tags) > 0 {
		lines += fmt.Sprintf("\nTags: %v", task.Tags)
	}

	return lines
}

// PrintTasksSimple writes tab-separated task data to stdout for non-interactive use.
func PrintTasksSimple(tasks []types.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found")
		return
	}
	for _, t := range tasks {
		var due string
		if d := time.Time(t.DueDate); !d.IsZero() {
			due = t.DueDate.Humanize()
		} else {
			due = "no due date"
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", t.ID, t.Title, t.Priority, due)
	}
}

// PrintProjectsSimple writes tab-separated project data to stdout for non-interactive use.
func PrintProjectsSimple(projects []types.Project) {
	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		return
	}
	for _, p := range projects {
		fmt.Printf("%s\t%s\n", p.ID, p.Name)
	}
}

func FuzzySelectProject(projects []types.Project, query string) (types.Project, error) {
	if len(projects) == 0 {
		return types.Project{}, fmt.Errorf("no projects available for selection")
	}
	idx, err := fuzzyfinder.Find(
		projects,
		func(i int) string {
			return fmt.Sprintf("%s (%s)",
				projects[i].Name,
				projects[i].ID,
			)
		},
		fuzzyfinder.WithQuery(query),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return GetProjectDescription(projects[i])
		}),
		fuzzyfinder.WithPromptString("Search Project: "),
	)
	if err != nil {
		return types.Project{}, err
	}

	return projects[idx], nil
}

func FuzzySelectTask(tasks []types.Task, projectColor project.Color, query string) (types.Task, error) {
	if len(tasks) == 0 {
		return types.Task{}, fmt.Errorf("no tasks found")
	}
	idx, err := fuzzyfinder.Find(
		tasks,
		func(i int) string {
			return fmt.Sprintf("%s (%s)",
				tasks[i].Title,
				tasks[i].ID,
			)
		},
		fuzzyfinder.WithQuery(query),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return GetTaskDescription(tasks[i], projectColor)
		}),
		fuzzyfinder.WithPromptString("Search Project: "),
	)
	if err != nil {
		return types.Task{}, err
	}

	return tasks[idx], nil
}
