package view

import (
	"fmt"
	"os"
	"time"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/tui/picker"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

type projectTask struct {
	types.Task
	ProjectName  string       `json:"projectName"`
	ProjectColor project.Color `json:"-"`
}

type viewOptions struct {
	all      bool
	priority task.Priority
	tag      string
	output   types.OutputFormat
}

func addViewFlags(cmd *cobra.Command, opts *viewOptions) {
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Include completed tasks in the results")
	cmd.Flags().VarP(&opts.priority, "priority", "p", "Only show tasks with this priority level or higher")
	_ = cmd.RegisterFlagCompletionFunc("priority", task.PriorityCompletionFunc)
	cmd.Flags().StringVar(&opts.tag, "tag", "", "Only show tasks with this specific tag")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)
}

func resolveOutput(cmd *cobra.Command, output types.OutputFormat) types.OutputFormat {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	quietFlag, _ := cmd.Flags().GetBool("quiet")
	return types.ResolveOutput(output, jsonFlag, quietFlag)
}

func fetchAllTasks(client *api.Client) ([]projectTask, error) {
	// Two API calls: one for projects (name lookup), one for all tasks
	projects, err := client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	projectMap := make(map[string]types.Project, len(projects)+1)
	projectMap[types.InboxProject.ID] = types.InboxProject
	for _, p := range projects {
		projectMap[p.ID] = p
	}

	tasks, err := client.FilterTasks(api.TaskFilter{})
	if err != nil {
		return nil, fmt.Errorf("fetching tasks: %w", err)
	}

	var all []projectTask
	for _, t := range tasks {
		p := projectMap[t.ProjectID]
		all = append(all, projectTask{
			Task:         t,
			ProjectName:  p.Name,
			ProjectColor: p.Color,
		})
	}

	return all, nil
}

type dateMatcher func(due time.Time, now, today, tomorrow, weekEnd time.Time) bool

func matchToday(due, now, today, tomorrow, _ time.Time) bool {
	return !due.Before(today) && due.Before(tomorrow)
}

func matchTomorrow(due, _, _, tomorrow, _ time.Time) bool {
	dayAfter := tomorrow.AddDate(0, 0, 1)
	return !due.Before(tomorrow) && due.Before(dayAfter)
}

func matchWeek(due, _, today, _, weekEnd time.Time) bool {
	return !due.Before(today) && due.Before(weekEnd)
}

func matchOverdue(due, now, _, _, _ time.Time) bool {
	return due.Before(now)
}

func filterByDate(tasks []projectTask, matchers ...dateMatcher) []projectTask {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	weekEnd := today.AddDate(0, 0, 7)

	seen := make(map[string]bool)
	var result []projectTask
	for _, t := range tasks {
		due := time.Time(t.DueDate)
		if due.IsZero() {
			continue
		}
		for _, match := range matchers {
			if match(due, now, today, tomorrow, weekEnd) {
				if !seen[t.ID] {
					seen[t.ID] = true
					result = append(result, t)
				}
				break
			}
		}
	}
	return result
}

func filterByOpts(tasks []projectTask, opts *viewOptions) []projectTask {
	var result []projectTask
	for _, t := range tasks {
		if t.Priority < opts.priority {
			continue
		}
		if opts.tag != "" {
			found := false
			for _, tag := range t.Tags {
				if tag == opts.tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		result = append(result, t)
	}
	return result
}

func getProjectTaskDescription(t projectTask) string {
	r := render.New()
	return r.TaskDetail(t.Task, t.ProjectName)
}

func formatTime(t types.TickTickTime) string {
	if time.Time(t).IsZero() {
		return ""
	}
	return t.Humanize()
}

func isInteractive() bool {
	return prompt.IsInteractive()
}

func computeProjectTaskFields(tasks []projectTask) {
	for i := range tasks {
		utils.ComputeFields(&tasks[i].Task)
	}
}

func printProjectTasksSimple(tasks []projectTask) {
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found")
		return
	}
	if !isInteractive() {
		// Piped: tab-separated for scripting
		for _, t := range tasks {
			due := formatTime(t.DueDate)
			if due == "" {
				due = "no due date"
			}
			fmt.Printf("%s\t[%s]\t%s\t%s\t%s\n", t.ID, t.ProjectName, t.Title, t.Priority, due)
		}
		return
	}
	r := render.New()
	plainTasks := make([]types.Task, len(tasks))
	names := make([]string, len(tasks))
	for i, t := range tasks {
		plainTasks[i] = t.Task
		names[i] = t.ProjectName
	}
	fmt.Println(r.TaskListWithProjects(plainTasks, names))
}

func printProjectTaskIDs(tasks []projectTask) {
	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found")
		return
	}
	for _, t := range tasks {
		fmt.Println(t.ID)
	}
}

func fuzzySelectProjectTask(tasks []projectTask, _ string) (projectTask, bool, error) {
	if len(tasks) == 0 {
		return projectTask{}, false, nil
	}
	plainTasks := make([]types.Task, len(tasks))
	names := make([]string, len(tasks))
	for i, t := range tasks {
		plainTasks[i] = t.Task
		names[i] = t.ProjectName
	}
	result, err := picker.RunTaskPicker(plainTasks, names, "Select Task")
	if err != nil {
		return projectTask{}, false, err
	}
	// Find the original projectTask to preserve ProjectColor
	for _, t := range tasks {
		if t.ID == result.Task.ID {
			return t, true, nil
		}
	}
	return projectTask{}, false, fmt.Errorf("selected task not found")
}
