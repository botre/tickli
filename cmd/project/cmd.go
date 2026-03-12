package project

import (
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

func resolveOutput(cmd *cobra.Command, output types.OutputFormat) types.OutputFormat {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	quietFlag, _ := cmd.Flags().GetBool("quiet")
	return types.ResolveOutput(output, jsonFlag, quietFlag)
}

// NewProjectCommand returns a cobra command for `project` subcommands
func NewProjectCommand() *cobra.Command {
	var client api.Client
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Work with TickTick projects",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			client = utils.LoadClient()
			return nil
		},
	}

	cmd.AddCommand(
		newListCommand(&client),
		newCreateProjectCommand(&client),
		newUpdateProjectCommand(&client),
		newUseProjectCmd(&client),
		newShowCommand(&client),
		newDeleteCommand(&client),
	)

	return cmd
}
