package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/botre/tickli/internal/tui"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI dashboard",
		Long: `Launch a full-screen interactive terminal interface for managing your TickTick tasks.

Navigate between smart views (Today, Tomorrow, Week, Inbox, All), browse projects,
view task details, and complete tasks — all without leaving the terminal.`,
		Example: `  # Launch the TUI
  tickli tui`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := utils.LoadClient()
			if err != nil {
				return err
			}

			model := tui.New(&client)
			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
	return cmd
}
