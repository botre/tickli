package project

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
	force     bool
	output    types.OutputFormat
}

func newDeleteCommand(client *api.Client) *cobra.Command {
	opts := deleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete <project>",
		Short: "Delete a project",
		Long: `Permanently delete a project by its ID.

This operation cannot be undone. By default, you will be asked to confirm
the deletion unless the --force flag is used or stdin is not a terminal.`,
		Example: `  # Delete with confirmation prompt
  tickli project delete abc123def456

  # Force delete without confirmation
  tickli project delete abc123def456 --force`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ProjectIDs(),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedProject, err := client.ResolveProject(opts.projectID)
			if err != nil {
				return fmt.Errorf("project %q not found by ID or name. Run 'tickli project list -o json' to see available projects: %w", opts.projectID, err)
			}

			if !opts.force && prompt.IsInteractive() {
				var confirm string
				fmt.Printf("Are you sure you want to delete the project %s? (y/N): ", resolvedProject.Name)
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Deletion aborted")
					return nil
				}
			}

			err = client.DeleteProject(resolvedProject.ID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to delete project %q", opts.projectID))
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				jsonData, _ := json.MarshalIndent(resolvedProject, "", "  ")
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(opts.projectID)
			default:
				fmt.Printf("Project %s deleted\n", opts.projectID)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt and delete immediately")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
