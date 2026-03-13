package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
)

func GetProjectDescription(project types.Project) string {
	r := render.New()
	return r.ProjectDetail(project)
}

func GetTaskDescription(task types.Task, projectColor project.Color) string {
	r := render.New()
	return r.TaskDetail(task, "")
}

// PrintTasksSimple writes styled task data to stdout for non-interactive use.
// When piped (non-TTY), falls back to tab-separated format for scripting.
func PrintTasksSimple(tasks []types.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found")
		return
	}
	if !isTerminal() {
		// Piped output: keep tab-separated for scripting
		for _, t := range tasks {
			var due string
			if d := time.Time(t.DueDate); !d.IsZero() {
				due = t.DueDate.Humanize()
			} else {
				due = "no due date"
			}
			fmt.Printf("%s\t%s\t%s\t%s\n", t.ID, t.Title, t.Priority, due)
		}
		return
	}
	r := render.New()
	fmt.Println(r.TaskList(tasks))
}

// PrintProjectsSimple writes styled project data to stdout for non-interactive use.
// When piped (non-TTY), falls back to tab-separated format for scripting.
func PrintProjectsSimple(projects []types.Project) {
	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		return
	}
	if !isTerminal() {
		// Piped output: keep tab-separated for scripting
		for _, p := range projects {
			fmt.Printf("%s\t%s\n", p.ID, p.Name)
		}
		return
	}
	r := render.New()
	for _, p := range projects {
		fmt.Println(r.ProjectLine(p))
	}
}

func isTerminal() bool {
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
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
