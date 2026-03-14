package view

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewTodayCommand() *cobra.Command {
	var client api.Client
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show today's tasks and overdue tasks across all projects",
		Long: `Display tasks due today and overdue tasks from all projects.

This is equivalent to TickTick's "Today" smart view.`,
		Example: `  # Show today's tasks
  tickli today

  # Show today's tasks in JSON
  tickli today -o json

  # Show today's high priority tasks
  tickli today -p high`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			client, err = utils.LoadClient()
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			allTasks, err := fetchAllTasks(&client)
			if err != nil {
				return err
			}

			tasks := filterByDate(allTasks, matchToday, matchOverdue)
			tasks = filterByOpts(tasks, opts)

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				if tasks == nil {
					tasks = []projectTask{}
				}
				computeProjectTaskFields(tasks)
				jsonData, err := json.MarshalIndent(tasks, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				printProjectTaskIDs(tasks)
			default:
				if !isInteractive() {
					printProjectTasksSimple(tasks)
					return nil
				}
				t, ok, err := fuzzySelectProjectTask(tasks, "")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "No tasks found")
					return nil
				}
				fmt.Println(getProjectTaskDescription(t))
			}
			return nil
		},
	}

	addViewFlags(cmd, opts)
	return cmd
}
