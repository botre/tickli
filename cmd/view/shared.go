package view

import (
	"fmt"
	"sync"
	"time"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/ktr0731/go-fuzzyfinder"
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
	projects, err := client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	type result struct {
		tasks   []types.Task
		project types.Project
		err     error
	}

	ch := make(chan result, len(projects))
	var wg sync.WaitGroup

	for _, p := range projects {
		wg.Add(1)
		go func(proj types.Project) {
			defer wg.Done()
			tasks, err := client.ListTasks(proj.ID)
			ch <- result{tasks: tasks, project: proj, err: err}
		}(p)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var all []projectTask
	for r := range ch {
		if r.err != nil {
			continue
		}
		for _, t := range r.tasks {
			all = append(all, projectTask{
				Task:         t,
				ProjectName:  r.project.Name,
				ProjectColor: r.project.Color,
			})
		}
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
	projectLine := t.ProjectColor.Sprint("----------------------")

	lines := fmt.Sprintf(`
Task Details:

%s %s
%s`,
		t.Status.ColorString(),
		projectLine,
		t.Title,
	)

	if t.Content != "" {
		lines += fmt.Sprintf("\nContent: %s", t.Content)
	}
	lines += fmt.Sprintf("\nPriority: %s", t.Priority.ColorString())
	lines += fmt.Sprintf("\nProject: %s", t.ProjectName)

	if s := formatTime(t.StartDate); s != "" {
		lines += fmt.Sprintf("\nStart: %s", s)
	}
	if s := formatTime(t.DueDate); s != "" {
		lines += fmt.Sprintf("\nDue: %s", s)
	}
	if len(t.Tags) > 0 {
		lines += fmt.Sprintf("\nTags: %v", t.Tags)
	}

	return lines
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
	for _, t := range tasks {
		due := formatTime(t.DueDate)
		if due == "" {
			due = "no due date"
		}
		fmt.Printf("%s\t[%s]\t%s\t%s\t%s\n", t.ID, t.ProjectName, t.Title, t.Priority, due)
	}
}

func fuzzySelectProjectTask(tasks []projectTask, query string) (projectTask, error) {
	if len(tasks) == 0 {
		return projectTask{}, fmt.Errorf("no tasks found")
	}
	idx, err := fuzzyfinder.Find(
		tasks,
		func(i int) string {
			return fmt.Sprintf("[%s] %s", tasks[i].ProjectName, tasks[i].Title)
		},
		fuzzyfinder.WithQuery(query),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return getProjectTaskDescription(tasks[i])
		}),
		fuzzyfinder.WithPromptString("Search Tasks: "),
	)
	if err != nil {
		return projectTask{}, err
	}
	return tasks[idx], nil
}
