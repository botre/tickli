package task

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct {
	taskID string
	output types.OutputFormat
}

func newShowCommand(client *api.Client) *cobra.Command {
	opts := &showOptions{
		output: types.OutputSimple,
	}
	cmd := &cobra.Command{
		Use:     "show <task-id>",
		Aliases: []string{"info", "get"},
		Short:   "Display detailed information about a task",
		Long: `Show complete information about a specific task identified by its ID.
    
Displays title, content, dates, priority, tags, and other properties.
You can choose between human-readable output or machine-readable JSON.`,
		Example: `  # Show task details in human-readable format
  tickli task show abc123def456

  # Show task details in JSON format
  tickli task show abc123def456 -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.taskID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := client.GetTask(opts.taskID)
			if err != nil {
				return err
			}
			if task.ID != opts.taskID {
				log.Warn().Str("task-id", opts.taskID).Msg("task not found")
				return fmt.Errorf("task %s not found", opts.taskID)
			}
			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				jsonData, err := json.MarshalIndent(task, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(task.ID)
			default:
				fmt.Println(utils.GetTaskDescription(*task, project.DefaultColor))
				fmt.Println(task.ID)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	return cmd
}
