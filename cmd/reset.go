package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/prompt"
	"github.com/spf13/cobra"
)

type resetOptions struct {
	force bool
}

func NewResetCommand() *cobra.Command {
	opts := &resetOptions{}
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset authentication",
		Long: `Reset tickli by removing the current access token and re-running the initialization process.
This is useful if you need to reauthenticate with TickTick.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.force && prompt.IsInteractive() {
				fmt.Printf("Are you sure you want to reset authentication? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Reset aborted")
					return nil
				}
			}

			if err := config.DeleteToken(); err != nil {
				return fmt.Errorf("failed to remove access token: %w", err)
			}

			log.Info().Msg("Successfully removed access token. Running initialization...")
			token, err := initTickli()
			if err != nil {
				return fmt.Errorf("failed to initialize tickli: %w", err)
			}
			log.Info().Str("token", token).Msg("Successfully initialized tickli")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Reset authentication without confirmation")
	return cmd
}
