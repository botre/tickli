package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
	"github.com/botre/tickli/internal/utils"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type uncompleteOptions struct {
	taskID string
	output types.OutputFormat
}

func newUncompleteCommand(client *api.Client) *cobra.Command {
	opts := &uncompleteOptions{}
	cmd := &cobra.Command{
		Use:   "uncomplete <task-id>",
		Short: "Uncomplete a task",
		Long: `Change a task's status from completed back to active.

Reactivates tasks that were previously completed.`,
		Example: `  # Reactivate a completed task
  tickli task uncomplete abc123def456`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := client.GetTask(opts.taskID)
			if err != nil {
				return errors.Wrap(err, "failed to get task")
			}

			t.Status = task.StatusNormal
			t, err = client.UpdateTask(t)
			if err != nil {
				return errors.Wrap(err, "failed to uncomplete task")
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				utils.ComputeFields(t)
				jsonData, _ := json.MarshalIndent(t, "", "  ")
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(opts.taskID)
			default:
				fmt.Printf("%s Task %s uncompleted\n", color.Yellow.Sprint("☐"), opts.taskID)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
