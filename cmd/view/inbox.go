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

func NewInboxCommand() *cobra.Command {
	var client api.Client
	opts := &viewOptions{}
	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Show tasks in the inbox",
		Long: `Display all tasks in the inbox project.

This is a shorthand for 'tickli task list --project inbox'.`,
		Example: `  # Show inbox tasks
  tickli inbox

  # Show inbox tasks in JSON
  tickli inbox -o json

  # Show high priority inbox tasks
  tickli inbox -p high`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			client = utils.LoadClient()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			inbox := types.InboxProject
			rawTasks, err := client.ListTasks(inbox.ID)
			if err != nil {
				return fmt.Errorf("listing inbox tasks: %w", err)
			}

			// Wrap as projectTasks for shared filtering
			tasks := make([]projectTask, len(rawTasks))
			for i, t := range rawTasks {
				tasks[i] = projectTask{
					Task:         t,
					ProjectName:  inbox.Name,
					ProjectColor: inbox.Color,
				}
			}
			tasks = filterByOpts(tasks, opts)

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				if tasks == nil {
					tasks = []projectTask{}
				}
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
