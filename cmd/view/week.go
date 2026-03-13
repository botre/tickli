package view

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewWeekCommand() *cobra.Command {
	var client api.Client
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "week",
		Short: "Show tasks for the next 7 days across all projects",
		Long: `Display tasks due within the next 7 days and overdue tasks from all projects.

This is equivalent to TickTick's "Next 7 Days" smart view.`,
		Example: `  # Show this week's tasks
  tickli week

  # Show this week's tasks in JSON
  tickli week -o json`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			client = utils.LoadClient()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			allTasks, err := fetchAllTasks(&client)
			if err != nil {
				return err
			}

			tasks := filterByDate(allTasks, matchWeek, matchOverdue)
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
