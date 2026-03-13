package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type completeOptions struct {
	taskID string
	output types.OutputFormat
}

func newCompleteCmd(client *api.Client) *cobra.Command {
	opts := &completeOptions{}
	cmd := &cobra.Command{
		Use:   "complete <task-id>",
		Short: "Complete a task",
		Long: `Change a task's status to completed.

Takes a task ID and marks it as done. The task remains in the system
but will no longer appear in default listings unless using the --all flag.`,
		Example: `  # Complete a task
  tickli task complete abc123def456`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := client.CompleteTask(opts.taskID)
			if err != nil {
				return errors.Wrap(err, "failed to complete task")
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				t, getErr := client.GetTask(opts.taskID)
				if getErr != nil {
					return errors.Wrap(getErr, "failed to get completed task")
				}
				utils.ComputeFields(t)
				jsonData, _ := json.MarshalIndent(t, "", "  ")
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(opts.taskID)
			default:
				fmt.Printf("%s Task %s completed\n", color.Green.Sprint("☑"), opts.taskID)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
