package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/tui/forms"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type createOptions struct {
	title    string
	content  string
	priority task.Priority
	tags        []string

	// time specific vars
	allDay    bool
	date      string
	startDate string
	dueDate   string
	timeZone  string

	// reminders and repeat are more advanced features not implemented yet
	reminders []string
	repeat    string

	// interactive indicates if you should prompt to get title and content
	interactive bool

	projectID string
	output    types.OutputFormat
}

func newCreateCommand(client *api.Client) *cobra.Command {
	opts := &createOptions{}
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "a"},
		Short:   "Create a new task",
		Long: `Create a new task in the current project or a specified project.
    
You can set various properties including title, content, priority, due date,
and tags. At minimum, a title is required unless using interactive mode.`,
		Example: `  # Create a basic task with just a title
  tickli task create -t "Buy groceries"
  
  # Create a task with priority and due date
  tickli task create -t "Submit report" -p high --due "2025-03-14"
  
  # Create a task in a specific project (by name or ID)
  tickli task create -t "Call client" --project Chores
  
  # Create a task with content and tags
  tickli task create -t "Team meeting" -c "Discuss Q3 roadmap" --tag meeting,work
  
  # Create a task interactively
  tickli task create -i`,
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.interactive {
				if !prompt.IsInteractive() {
					return fmt.Errorf("--interactive requires a terminal (stdin is not a TTY)")
				}

				// Fetch projects for the picker if no project is pre-selected
				var projects []types.Project
				allProjects, listErr := client.ListProjects()
				if listErr != nil {
					return fmt.Errorf("failed to fetch projects: %w", listErr)
				}
				if opts.projectID == "" {
					projects = allProjects
				}

				// Collect known tags from all projects for the multi-select
				var knownTags []string
				for _, proj := range allProjects {
					tasks, taskErr := client.ListTasks(proj.ID)
					if taskErr == nil {
						knownTags = append(knownTags, forms.CollectTags(tasks)...)
					}
				}
				knownTags = dedupStrings(knownTags)

				t := theme.Default()
				result, err := forms.RunTaskCreateForm(t, forms.TaskFormResult{
					Title:    opts.title,
					Content:  opts.content,
					Priority: opts.priority,
					Date:     opts.date,
					Project:  opts.projectID,
				}, projects, knownTags)
				if err != nil {
					return fmt.Errorf("form cancelled: %w", err)
				}
				opts.title = result.Title
				opts.content = result.Content
				opts.priority = result.Priority
				if result.Project != "" {
					opts.projectID = result.Project
				}
				if result.Date != "" {
					opts.date = result.Date
				}
				if result.Tags != "" {
					for _, tag := range strings.Split(result.Tags, ",") {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							opts.tags = append(opts.tags, tag)
						}
					}
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

			t := &types.Task{
				ProjectID: opts.projectID,
				Title:     opts.title,
				Content:   opts.content,
				Priority:  opts.priority,
				Tags:      opts.tags,
			}

			if opts.date != "" {
				r, err := utils.ParseTimeExpression(opts.date)
				if err != nil {
					return errors.Wrap(err, "failed to parse date range")
				}
				t.StartDate = types.TickTickTime(r.Start())
				t.DueDate = types.TickTickTime(r.End())
				t.IsAllDay = r.IsAllDay()
			}
			if opts.startDate != "" {
				startDate, err := utils.ParseFlexibleTime(opts.startDate)
				if err != nil {
					return errors.Wrap(err, "failed to parse start date")
				}
				t.StartDate = types.TickTickTime(startDate)
			}
			if opts.dueDate != "" {
				dueDate, err := utils.ParseFlexibleTime(opts.dueDate)
				if err != nil {
					return errors.Wrap(err, "failed to parse due date")
				}
				t.DueDate = types.TickTickTime(dueDate)
			}
			if opts.timeZone != "" {
				t.TimeZone = opts.timeZone
			}
			if cmd.Flags().Changed("all-day") {
				t.IsAllDay = opts.allDay
				if opts.allDay {
					if s := time.Time(t.StartDate); !s.IsZero() {
						t.StartDate = types.TickTickTime(utils.TruncateToDate(s))
					}
					if d := time.Time(t.DueDate); !d.IsZero() {
						t.DueDate = types.TickTickTime(utils.TruncateToDate(d))
					}
				}
			}

			t, err = client.CreateTask(t)
			if err != nil {
				return errors.Wrap(err, "failed to create task")
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				utils.ComputeFields(t)
				jsonData, err := json.MarshalIndent(t, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(t.ID)
			default:
				r := render.New()
				fmt.Println(r.SuccessMessage(fmt.Sprintf("Created task %s", t.ID)))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Title of the task (required unless -i)")
	cmd.Flags().StringVarP(&opts.content, "content", "c", "", "Additional details about the task")
	cmd.Flags().BoolVar(&opts.allDay, "all-day", false, "Set as an all-day task without specific time (use --all-day=false to unset)")
	cmd.Flags().StringVar(&opts.startDate, "start", "", "Start date/time (e.g. 'tomorrow', '2025-02-18'). Use with --due for ranges")
	cmd.Flags().StringVar(&opts.dueDate, "due", "", "Deadline (e.g. 'friday', '2025-02-18'). Use with --start for ranges")
	cmd.Flags().StringVar(&opts.date, "date", "", "Set start+due together via natural language (e.g. 'today', 'tomorrow 2pm'). Cannot combine with --start/--due")

	cmd.MarkFlagsMutuallyExclusive("date", "start")
	cmd.MarkFlagsMutuallyExclusive("date", "due")

	cmd.Flags().StringVar(&opts.timeZone, "timezone", "", "Timezone for date calculations (e.g., 'America/Los_Angeles')")
	cmd.Flags().StringSliceVar(&opts.reminders, "reminders", []string{}, "List of reminder triggers (e.g., '9h', '0s')")
	_ = cmd.Flags().MarkHidden("reminders")
	cmd.Flags().StringSliceVar(&opts.tags, "tag", []string{}, "Apply tags to categorize the task (comma-separated)")
	cmd.Flags().StringVar(&opts.repeat, "repeat", "", "Recurring rule (e.g., 'daily', 'weekly on monday')")
	_ = cmd.Flags().MarkHidden("repeat")
	cmd.Flags().VarP(&opts.priority, "priority", "p", "Task importance: none, low, medium, high")
	_ = cmd.RegisterFlagCompletionFunc("priority", task.PriorityCompletionFunc)
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Create task by answering prompts")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)
	cmd.MarkFlagsOneRequired("title", "interactive")

	return cmd
}

func dedupStrings(s []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
