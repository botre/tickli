package task

import (
	"github.com/pkg/errors"
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	projectID string
)

func NewTaskCommand() *cobra.Command {
	var client api.Client
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Work with TickTick tasks",
		Long: `Create, view, update, and manage tasks in your TickTick projects.

Single-task commands (show, update, delete, complete, uncomplete) work with just a task ID.
The --project flag is only needed for list and create.`,
		Example: `  # List all tasks in current project
  tickli task list
  
  # Create a new task
  tickli task create -t "Submit quarterly report"
  
  # Complete a task
  tickli task complete abc123def456`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			client = utils.LoadClient()
			if projectID == "" {
				cfg, err := config.Load()
				if err != nil {
					return errors.Wrap(err, "failed to load config")
				}
				projectID = cfg.DefaultProject
			}
			return nil
		},
	}

	cmd.AddCommand(
		newCompleteCmd(&client),
		newDeleteCommand(&client),
		newShowCommand(&client),
		newCreateCommand(&client),
		newListCommand(&client),
		newUncompleteCommand(&client),
		newUpdateCommand(&client),
		newMoveCommand(&client),
	)

	RegisterProjectOverride(cmd)

	return cmd
}

func resolveOutput(cmd *cobra.Command, output types.OutputFormat) types.OutputFormat {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	quietFlag, _ := cmd.Flags().GetBool("quiet")
	return types.ResolveOutput(output, jsonFlag, quietFlag)
}

func RegisterProjectOverride(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&projectID, "project", "P", "", "Project context for list and create commands (ID or name)")

	_ = cmd.RegisterFlagCompletionFunc("project", completion.ProjectIDs())
}
