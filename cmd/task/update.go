package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	taskID string

	title      string
	content    string
	priority   task.Priority
	tags       []string
	moveToProject string

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

	output types.OutputFormat
}

func newUpdateCommand(client *api.Client) *cobra.Command {
	opts := &updateOptions{}
	cmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Update a task",
		Long: `Update any property of an existing task identified by its ID.

The task is found automatically across all projects; no --project flag needed.
Changes only the properties you specify - others remain unchanged.`,
		Example: `  # Update a task's title
  tickli task update abc123def456 -t "New title"
  
  # Update priority and add content
  tickli task update abc123def456 -p high -c "Additional details here"
  
  # Change due date
  tickli task update abc123def456 --due "2025-03-21"
  
  # Update interactively
  tickli task update abc123def456 -i`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := client.GetTask(opts.taskID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get task %q", opts.taskID))
			}

			if opts.interactive {
				if !prompt.IsInteractive() {
					return fmt.Errorf("--interactive requires a terminal (stdin is not a TTY)")
				}
				newTitle := prompt.String("Title", t.Title)
				if newTitle != t.Title {
					t.Title = newTitle
				}
				newContent := prompt.String("Content", t.Content)
				if newContent != t.Content {
					t.Content = newContent
				}

				priorities := []string{"none", "low", "medium", "high"}
				idx, selectErr := prompt.Select("Priority:", priorities)
				if selectErr == nil {
					var p task.Priority
					_ = p.Set(priorities[idx])
					t.Priority = p
				}

				tagsInput := prompt.String("Tags (comma-separated)", strings.Join(t.Tags, ", "))
				if tagsInput != "" {
					var tags []string
					for _, tag := range strings.Split(tagsInput, ",") {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							tags = append(tags, tag)
						}
					}
					t.Tags = tags
				} else {
					t.Tags = nil
				}

				dateInput := prompt.String("Date (e.g. 'tomorrow 2pm')", "")
				if dateInput != "" {
					r, parseErr := utils.ParseTimeExpression(dateInput)
					if parseErr == nil {
						t.StartDate = types.TickTickTime(r.Start())
						t.DueDate = types.TickTickTime(r.End())
						t.IsAllDay = r.IsAllDay()
					}
				}
			}

			if cmd.Flags().Changed("title") {
				t.Title = opts.title
			}
			if cmd.Flags().Changed("content") {
				t.Content = opts.content
			}
			if cmd.Flags().Changed("priority") {
				t.Priority = opts.priority
			}
			if cmd.Flags().Changed("tag") {
				t.Tags = opts.tags
			}
			if cmd.Flags().Changed("date") {
				r, err := utils.ParseTimeExpression(opts.date)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to parse range %s", opts.date))
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
			var movedToProject *types.Project
			if cmd.Flags().Changed("move-to") || cmd.Flags().Changed("to") {
				resolvedProject, resolveErr := client.ResolveProject(opts.moveToProject)
				if resolveErr != nil {
					return resolveErr
				}
				movedToProject = &resolvedProject
			}
			t, err = client.UpdateTask(t)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to update task %q", opts.taskID))
			}
			if movedToProject != nil {
				if moveErr := client.MoveTask(t.ID, t.ProjectID, movedToProject.ID); moveErr != nil {
					return errors.Wrap(moveErr, fmt.Sprintf("failed to move task %q", opts.taskID))
				}
				t.ProjectID = movedToProject.ID
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
				fmt.Printf("Task %s updated successfully\n", t.ID)
				fmt.Println(utils.GetTaskDescription(*t, project.DefaultColor))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Change the task title")
	cmd.Flags().StringVarP(&opts.content, "content", "c", "", "Change or add content/description")
	cmd.Flags().BoolVar(&opts.allDay, "all-day", false, "Change to all-day task without specific time (use --all-day=false to unset)")
	cmd.Flags().StringVar(&opts.startDate, "start", "", "Change start date/time (e.g. 'tomorrow', '2025-02-18'). Use with --due for ranges")
	cmd.Flags().StringVar(&opts.dueDate, "due", "", "Change deadline (e.g. 'friday', '2025-02-18'). Use with --start for ranges")
	cmd.Flags().StringVar(&opts.date, "date", "", "Set start+due together via natural language (e.g. 'today', 'tomorrow 2pm'). Cannot combine with --start/--due")

	cmd.MarkFlagsMutuallyExclusive("date", "start")
	cmd.MarkFlagsMutuallyExclusive("date", "due")

	cmd.Flags().StringVar(&opts.timeZone, "timezone", "", "Change timezone for date calculations")
	cmd.Flags().StringSliceVar(&opts.reminders, "reminders", []string{}, "Set reminders (e.g., '10m', '1h before')")
	_ = cmd.Flags().MarkHidden("reminders")
	cmd.Flags().StringVar(&opts.repeat, "repeat", "", "New recurring rule (e.g., 'daily', 'weekly on monday')")
	_ = cmd.Flags().MarkHidden("repeat")
	cmd.Flags().StringSliceVar(&opts.tags, "tag", []string{}, "Change tags on the task (comma-separated)")
	cmd.Flags().VarP(&opts.priority, "priority", "p", "Change task importance: none, low, medium, high")
	_ = cmd.RegisterFlagCompletionFunc("priority", task.PriorityCompletionFunc)
	cmd.Flags().StringVar(&opts.moveToProject, "move-to", "", "Move task to a different project (name or ID)")
	cmd.Flags().StringVar(&opts.moveToProject, "to", "", "Move task to a different project (alias for --move-to)")
	_ = cmd.RegisterFlagCompletionFunc("move-to", completion.ProjectIDs())
	_ = cmd.RegisterFlagCompletionFunc("to", completion.ProjectIDs())
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Update task by answering prompts")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
