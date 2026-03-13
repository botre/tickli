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

func NewTomorrowCommand() *cobra.Command {
	var client api.Client
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "tomorrow",
		Short: "Show tomorrow's tasks across all projects",
		Long: `Display tasks due tomorrow from all projects.

This is equivalent to TickTick's "Tomorrow" smart view.`,
		Example: `  # Show tomorrow's tasks
  tickli tomorrow

  # Show tomorrow's tasks in JSON
  tickli tomorrow -o json`,
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

			tasks := filterByDate(allTasks, matchTomorrow)
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
