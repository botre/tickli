package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	projectID string
	taskID    string
	force     bool
	output    types.OutputFormat
}

func newDeleteCommand(client *api.Client) *cobra.Command {
	opts := &deleteOptions{}
	cmd := &cobra.Command{
		Use:     "delete <task-id>",
		Aliases: []string{"rm", "remove"},
		Short:   "Remove a task permanently",
		Long: `Delete a task completely from your TickTick account.

This operation cannot be undone. By default, you will be asked to confirm
the deletion unless the --force flag is used or stdin is not a terminal.`,
		Example: `  # Delete with confirmation prompt
  tickli task delete abc123def456

  # Force delete without confirmation
  tickli task delete abc123def456 --force

  # Delete from specific project
  tickli task delete abc123def456 --project-id xyz789`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.force && prompt.IsInteractive() {
				var confirm string
				fmt.Printf("Are you sure you want to delete the task %s? (y/N): ", opts.taskID)
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Deletion aborted")
					return nil
				}
			}

			err := client.DeleteTask(opts.projectID, opts.taskID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to delete task %s", opts.taskID))
			}

			if opts.output == types.OutputJSON {
				result := map[string]string{"id": opts.taskID, "status": "deleted"}
				jsonData, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(jsonData))
			} else {
				fmt.Printf("Task %s deleted\n", opts.taskID)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt and delete immediately")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
