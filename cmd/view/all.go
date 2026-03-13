package view

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewAllCommand() *cobra.Command {
	var client api.Client
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Show all incomplete tasks across all projects",
		Long: `Display all incomplete tasks from every project, including overdue tasks,
future tasks, and tasks without a due date.

Completed tasks are excluded by default. Use --all to include them.`,
		Example: `  # Show all incomplete tasks
  tickli all

  # Include completed tasks
  tickli all -a

  # Show all tasks in JSON
  tickli all -o json

  # Show only high priority tasks
  tickli all -p high`,
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

			tasks := filterByOpts(allTasks, opts)

			// Filter out completed tasks unless --all is set
			if !opts.all {
				var incomplete []projectTask
				for _, t := range tasks {
					if t.Status != task.StatusComplete {
						incomplete = append(incomplete, t)
					}
				}
				tasks = incomplete
			}

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
				for _, t := range tasks {
					fmt.Println(t.ID)
				}
			default:
				if !isInteractive() {
					printProjectTasksSimple(tasks)
					return nil
				}
				t, err := fuzzySelectProjectTask(tasks, "")
				if err != nil {
					return err
				}
				fmt.Println(getProjectTaskDescription(t))
			}
			return nil
		},
	}

	addViewFlags(cmd, opts)
	return cmd
}
