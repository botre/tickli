package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/botre/tickli/cmd/project"
	"github.com/botre/tickli/cmd/subtask"
	"github.com/botre/tickli/cmd/task"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var noColor bool

func ColorDisabled() bool {
	if noColor {
		return true
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	return false
}


func NewTickliCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tickli",
		Short: "TickTick CLI - A modern command line interface for TickTick",
		Long: `tickli is a CLI tool that helps you manage your TickTick tasks from the command line.
Complete documentation is available at https://github.com/botre/tickli`,
		SilenceErrors: true,
		SilenceUsage:  false,
	}
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	cmd.AddCommand(
		NewInitCommand(),
		NewResetCommand(),
		NewVersionCommand(),
		task.NewTaskCommand(),
		project.NewProjectCommand(),
		subtask.NewSubtaskCommand(),
	)

	return cmd
}

func Execute() {
	cmd := NewTickliCommand()

	// Parse flags early so --no-color is available before PersistentPreRun
	cmd.ParseFlags(os.Args[1:])

	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    ColorDisabled(),
		TimeFormat: "15:04:05",
		FormatFieldName: func(i interface{}) string {
			return i.(string) + ":"
		},
		FormatFieldValue: func(i interface{}) string {
			return "'" + i.(string) + "'"
		},
	})

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
