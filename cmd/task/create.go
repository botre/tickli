package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
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
  tickli task create -t "Submit report" -p high --due "tomorrow 5pm"
  
  # Create a task in a specific project (by ID or name)
  tickli task create -t "Call client" --project Work
  
  # Create a task with content and tags
  tickli task create -t "Team meeting" -c "Discuss Q3 roadmap" --tag meeting,work
  
  # Create a task interactively
  tickli task create -i`,
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.projectID == "" {
				return fmt.Errorf("no project selected. Use -P <project> or run 'tickli project use' to set a default.\nRun 'tickli project list -o json' to see available projects")
			}

			resolvedProject, err := client.ResolveProject(opts.projectID)
			if err != nil {
				return fmt.Errorf("project %q not found by ID or name. Run 'tickli project list -o json' to see available projects: %w", opts.projectID, err)
			}
			opts.projectID = resolvedProject.ID

			if opts.interactive {
				if !prompt.IsInteractive() {
					return fmt.Errorf("--interactive requires a terminal (stdin is not a TTY)")
				}
				opts.title = prompt.String("Title", opts.title)
				if opts.title == "" {
					return fmt.Errorf("title is required")
				}
				opts.content = prompt.String("Content", opts.content)

				priorities := []string{"none", "low", "medium", "high"}
				idx, err := prompt.Select("Priority:", priorities)
				if err == nil {
					_ = opts.priority.Set(priorities[idx])
				}

				dateInput := prompt.String("Date (e.g. 'tomorrow 2pm')", "")
				if dateInput != "" {
					opts.date = dateInput
				}

				tagsInput := prompt.String("Tags (comma-separated)", "")
				if tagsInput != "" {
					for _, tag := range strings.Split(tagsInput, ",") {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							opts.tags = append(opts.tags, tag)
						}
					}
				}
			}

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
				fmt.Printf("Created task %s\n", t.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Title of the task (required unless -i)")
	cmd.Flags().StringVarP(&opts.content, "content", "c", "", "Additional details about the task")
	cmd.Flags().BoolVar(&opts.allDay, "all-day", false, "Set as an all-day task without specific time (use --all-day=false to unset)")
	cmd.Flags().StringVar(&opts.startDate, "start", "", "When the task begins (ISO 8601, e.g. '2025-02-18T15:04:05+01:00')")
	cmd.Flags().StringVar(&opts.dueDate, "due", "", "When the task is due (ISO 8601, e.g. '2025-02-18T18:00:00+01:00')")
	cmd.Flags().StringVar(&opts.date, "date", "", "Set date with natural language (e.g., 'today', 'next week')")

	cmd.MarkFlagsMutuallyExclusive("date", "all-day")
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
