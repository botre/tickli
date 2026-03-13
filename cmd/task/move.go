package task

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newMoveCommand(client *api.Client) *cobra.Command {
	var (
		targetProject string
		output        types.OutputFormat
	)
	cmd := &cobra.Command{
		Use:   "move <task-id> --to <project>",
		Short: "Move a task to a different project",
		Long: `Move a task from its current project to a different one.

The target project can be specified by name or ID.`,
		Example: `  # Move a task to the "Work" project
  tickli task move abc123def456 --to Work

  # Move a task to the inbox
  tickli task move abc123def456 --to inbox`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.TaskIDs(projectID),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			t, err := client.GetTask(taskID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get task %q", taskID))
			}

			resolvedProject, err := client.ResolveProject(targetProject)
			if err != nil {
				return fmt.Errorf("target project %q not found by ID or name. Run 'tickli project list -o json' to see available projects", targetProject)
			}

			err = client.MoveTask(t.ID, t.ProjectID, resolvedProject.ID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to move task %q", taskID))
			}
			t.ProjectID = resolvedProject.ID

			switch resolveOutput(cmd, output) {
			case types.OutputJSON:
				utils.ComputeFields(t)
				jsonData, err := json.MarshalIndent(t, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(t.ID)
			default:
				fmt.Printf("Task %s moved to %s\n", t.ID, resolvedProject.Name)
				fmt.Println(utils.GetTaskDescription(*t, project.DefaultColor))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&targetProject, "to", "", "Target project (name or ID)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.RegisterFlagCompletionFunc("to", completion.ProjectIDs())
	cmd.Flags().VarP(&output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
