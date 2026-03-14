package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/tui/forms"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	taskID string
	force  bool
	output types.OutputFormat
}

func newDeleteCommand(client *api.Client) *cobra.Command {
	opts := &deleteOptions{}
	cmd := &cobra.Command{
		Use:     "delete <task-id>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a task",
		Long: `Delete a task completely from your TickTick account.

The task is found automatically across all projects; no --project flag needed.
This operation cannot be undone. By default, you will be asked to confirm
the deletion unless the --force flag is used or stdin is not a terminal.`,
		Example: `  # Delete with confirmation prompt
  tickli task delete abc123def456

  # Force delete without confirmation
  tickli task delete abc123def456 --force`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Fetch task so we can show its title in confirmation and JSON output
			t, getErr := client.GetTask(opts.taskID)
			if getErr != nil {
				return errors.Wrap(getErr, fmt.Sprintf("failed to get task %q", opts.taskID))
			}

			if !opts.force && prompt.IsInteractive() {
				confirmed, err := forms.RunConfirm(
					fmt.Sprintf("Delete \"%s\"?", t.Title),
					"This cannot be undone.",
				)
				if err != nil || !confirmed {
					fmt.Println("Deletion aborted")
					return nil
				}
			}

			var taskSnapshot *types.Task
			if resolveOutput(cmd, opts.output) == types.OutputJSON {
				taskSnapshot = t
			}

			err := client.DeleteTask(opts.taskID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to delete task %q", opts.taskID))
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				utils.ComputeFields(taskSnapshot)
				jsonData, _ := json.MarshalIndent(taskSnapshot, "", "  ")
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(opts.taskID)
			default:
				r := render.New()
				fmt.Println(r.SuccessMessage(fmt.Sprintf("Task %s deleted", opts.taskID)))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt and delete immediately")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
