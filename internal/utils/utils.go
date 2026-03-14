package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
)

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

