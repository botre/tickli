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
	projectID string
	taskID    string

	title       string
	content     string
	description string
	priority    task.Priority
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

	output types.OutputFormat
}

func newUpdateCommand(client *api.Client) *cobra.Command {
	opts := &updateOptions{}
	cmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Modify an existing task's properties",
		Long: `Update any property of an existing task identified by its ID.
    
Changes only the properties you specify - others remain unchanged.
This command allows modifying title, content, priority, dates, and more.`,
		Example: `  # Update a task's title
  tickli task update abc123def456 -t "New title"
  
  # Update priority and add content
  tickli task update abc123def456 -p high -c "Additional details here"
  
  # Change due date
  tickli task update abc123def456 --due "next Friday 5pm"
  
  # Update interactively
  tickli task update abc123def456 -i`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := client.GetTask(opts.projectID, opts.taskID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get task with ID %s", opts.taskID))
			}

			if opts.interactive {
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
			if cmd.Flags().Changed("desc") {
				t.Desc = opts.description
			}
			if cmd.Flags().Changed("priority") {
				t.Priority = opts.priority
			}
			if cmd.Flags().Changed("tags") {
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
				startDate, err := time.Parse(time.RFC3339, opts.startDate)
				if err != nil {
					return errors.Wrap(err, "failed to parse start date")
				}
				t.StartDate = types.TickTickTime(startDate)
			}
			if opts.dueDate != "" {
				dueDate, err := time.Parse(time.RFC3339, opts.dueDate)
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
			}
			t, err = client.UpdateTask(t)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to update task %s", opts.taskID))
			}
			if opts.output == types.OutputJSON {
				jsonData, err := json.MarshalIndent(t, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			} else {
				fmt.Printf("Task %s updated successfully\n", t.ID)
				fmt.Println(utils.GetTaskDescription(*t, project.DefaultColor))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Change the task title")
	cmd.Flags().StringVarP(&opts.content, "content", "c", "", "Change or add content/description")
	cmd.Flags().StringVarP(&opts.description, "desc", "d", "", "New description (for checklist)")
	cmd.Flags().MarkDeprecated("desc", "please use --content")
	cmd.Flags().BoolVarP(&opts.allDay, "all-day", "a", false, "Toggle all-day status for the task")
	cmd.Flags().StringVar(&opts.startDate, "start", "", "Change when the task begins")
	cmd.Flags().StringVar(&opts.dueDate, "due", "", "Change when the task is due")
	cmd.Flags().StringVar(&opts.date, "date", "", "Set date with natural language (e.g., 'today', 'next week')")

	cmd.MarkFlagsMutuallyExclusive("date", "start")
	cmd.MarkFlagsMutuallyExclusive("date", "all-day")
	cmd.MarkFlagsMutuallyExclusive("date", "due")

	cmd.Flags().StringVar(&opts.timeZone, "timezone", "", "Change timezone for date calculations")
	cmd.Flags().StringSliceVar(&opts.reminders, "reminders", []string{}, "Set reminders (e.g., '10m', '1h before')")
	cmd.Flags().StringVar(&opts.repeat, "repeat", "", "New recurring rule (e.g., 'daily', 'weekly on monday')")
	cmd.Flags().Var(&opts.priority, "priority", "Change task importance: none, low, medium, high")
	_ = cmd.RegisterFlagCompletionFunc("priority", task.PriorityCompletionFunc)
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Update task by answering prompts")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
