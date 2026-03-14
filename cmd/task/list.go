package task

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/tui/picker"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	all       bool
	verbose   bool
	priority  task.Priority
	dueDate   string
	tag       string
	projectID string
	output    types.OutputFormat
}

func filterTasks(tasks []types.Task, opts *listOptions) []types.Task {
	// Filter by priority
	tasks = Filter(tasks, func(t types.Task) bool {
		return t.Priority >= opts.priority
	})

	// Filter by tags
	tasks = Filter(tasks, func(t types.Task) bool {
		if opts.tag != "" {
			return slices.Contains(t.Tags, opts.tag)
		}
		return true
	})

	// Filter by completion status
	if !opts.all {
		//	tasks = Filter(tasks, func(t types.Task) bool {
		//		return !t.
		//	})
	}

	if opts.dueDate != "" {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tomorrow := today.AddDate(0, 0, 1)
		weekEnd := today.AddDate(0, 0, 7)

		tasks = Filter(tasks, func(t types.Task) bool {
			due := time.Time(t.DueDate)
			if due.IsZero() {
				return false
			}
			switch opts.dueDate {
			case "today":
				return !due.Before(today) && due.Before(tomorrow)
			case "tomorrow":
				return !due.Before(tomorrow) && due.Before(tomorrow.AddDate(0, 0, 1))
			case "this-week":
				return !due.Before(today) && due.Before(weekEnd)
			case "overdue":
				return due.Before(now)
			default:
				return false
			}
		})
	}

	return tasks
}

func newListCommand(client *api.Client) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tasks",
		Long: `Display tasks in the current project or a specified project.
    
By default, only shows incomplete tasks. You can filter tasks by priority,
tags, and due date. Results are displayed in an interactive selector.`,
		Example: `  # List all incomplete tasks in current project
  tickli task list
  
  # List all tasks including completed ones
  tickli task list --all
  
  # List tasks with specific tag
  tickli task list --tag important
  
  # List high priority tasks
  tickli task list -p high
  
  # List tasks in a specific project (by name or ID)
  tickli task list --project Chores`,
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.dueDate != "" {
				switch opts.dueDate {
				case "today", "tomorrow", "this-week", "overdue":
				default:
					return fmt.Errorf("invalid --due-within value %q: must be one of today, tomorrow, this-week, overdue", opts.dueDate)
				}
			}
			if opts.projectID == "" {
				return fmt.Errorf("no project selected. Use -P <project> or run 'tickli project use' to set a default.\nRun 'tickli project list -o json' to see available projects")
			}

			resolvedProject, err := client.ResolveProject(opts.projectID)
			if err != nil {
				return err
			}
			opts.projectID = resolvedProject.ID

			filteredTasks, err := client.ListTasks(opts.projectID)
			if err != nil {
				return err
			}
			filteredTasks = filterTasks(filteredTasks, opts)

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				if filteredTasks == nil {
					filteredTasks = []types.Task{}
				}
				for i := range filteredTasks {
					utils.ComputeFields(&filteredTasks[i])
				}
				jsonData, err := json.MarshalIndent(filteredTasks, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				if len(filteredTasks) == 0 {
					fmt.Fprintln(os.Stderr, "No tasks found")
				}
				for _, t := range filteredTasks {
					fmt.Println(t.ID)
				}
			default:
				if !prompt.IsInteractive() {
					utils.PrintTasksSimple(filteredTasks)
					return nil
				}
				if len(filteredTasks) == 0 {
					fmt.Fprintln(os.Stderr, "No tasks found")
					return nil
				}
				result, err := picker.RunTaskPicker(filteredTasks, nil, "Select Task")
				if err != nil {
					return fmt.Errorf("failed to select task: %w", err)
				}
				r := render.New()
				fmt.Println(r.TaskDetail(result.Task, result.ProjectName))
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Include completed tasks in the results")
	cmd.Flags().StringVar(&opts.tag, "tag", "", "Only show tasks with this specific tag")
	cmd.Flags().VarP(&opts.priority, "priority", "p", "Only show tasks with this priority level or higher")
	_ = cmd.RegisterFlagCompletionFunc("priority", task.PriorityCompletionFunc)
	cmd.Flags().StringVar(&opts.dueDate, "due-within", "", "Filter by due date window: today, tomorrow, this-week, overdue")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Show more details for each task in the list")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
func Filter(tasks []types.Task, predicate func(task types.Task) bool) []types.Task {
	var result []types.Task
	for _, t := range tasks {
		if predicate(t) {
			result = append(result, t)
		}
	}
	return result
}
