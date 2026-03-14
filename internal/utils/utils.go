package utils

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
)

// SortTasksByDueDate sorts tasks by due date ascending, with no-date tasks at the end.
func SortTasksByDueDate(tasks []types.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		di := time.Time(tasks[i].DueDate)
		dj := time.Time(tasks[j].DueDate)
		if di.IsZero() && dj.IsZero() {
			return false
		}
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.Before(dj)
	})
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

