package cmd

import (
	"github.com/botre/tickli/internal/update"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update tickli to the latest version",
		Long: `Update tickli to the latest released version.

This re-runs 'go install' against the latest tag, so the Go toolchain must be
installed. If you installed tickli another way, upgrade it with that tool
instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return update.Run(cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}
