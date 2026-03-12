package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/types"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type completeOptions struct {
	projectID string
	taskID    string
	output    types.OutputFormat
}

func newCompleteCmd(client *api.Client) *cobra.Command {
	opts := &completeOptions{}
	cmd := &cobra.Command{
		Use:   "complete <task-id>",
		Short: "Mark a task as completed",
		Long: `Change a task's status to completed.

Takes a task ID and marks it as done. The task remains in the system
but will no longer appear in default listings unless using the --all flag.`,
		Example: `  # Complete a task in current project
  tickli task complete abc123def456

  # Complete a task in a specific project
  tickli task complete abc123def456 --project-id xyz789`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = projectID
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := client.CompleteTask(opts.projectID, opts.taskID)
			if err != nil {
				return errors.Wrap(err, "failed to complete task")
			}

			if opts.output == types.OutputJSON {
				result := map[string]string{"id": opts.taskID, "status": "completed"}
				jsonData, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(jsonData))
			} else {
				fmt.Printf("%s Task %s completed\n", color.Green.Sprint("☑"), opts.taskID)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
