package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/botre/tickli/cmd/project"
	"github.com/botre/tickli/cmd/task"
	"github.com/botre/tickli/cmd/view"
	cliErrors "github.com/botre/tickli/internal/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitUsage     = 2
	ExitNotFound  = 3
	ExitAuthError = 4
)

var (
	noColor     bool
	JSONOutput  bool
	QuietOutput bool
)

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
		SilenceUsage:  true,
	}
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	cmd.PersistentFlags().BoolVar(&JSONOutput, "json", false, "Output in JSON format (overrides per-command -o)")
	cmd.PersistentFlags().BoolVarP(&QuietOutput, "quiet", "q", false, "Only print IDs (overrides per-command -o)")
	cmd.AddCommand(
		NewInitCommand(),
		NewResetCommand(),
		NewVersionCommand(),
		task.NewTaskCommand(),
		project.NewProjectCommand(),
		view.NewTodayCommand(),
		view.NewTomorrowCommand(),
		view.NewWeekCommand(),
		view.NewInboxCommand(),
		view.NewAllCommand(),
	)

	return cmd
}

func Execute() {
	cmd := NewTickliCommand()

	// Parse flags early so --no-color is available before PersistentPreRun
	cmd.ParseFlags(os.Args[1:])

	if ColorDisabled() {
		color.Disable()
	}

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
		msg := err.Error()
		code := ExitError
		switch {
		case errors.As(err, new(*cliErrors.NotFoundError)):
			code = ExitNotFound
		case errors.As(err, new(*cliErrors.UsageError)):
			code = ExitUsage
		case errors.As(err, new(*cliErrors.AuthError)):
			code = ExitAuthError
		case strings.Contains(msg, "required flag") ||
			strings.Contains(msg, "flags in the group") ||
			strings.Contains(msg, "unknown flag") ||
			strings.Contains(msg, "unknown command") ||
			strings.Contains(msg, "invalid argument") ||
			strings.Contains(msg, "accepts") ||
			strings.Contains(msg, "arg(s)"):
			code = ExitUsage
		case strings.Contains(msg, "not found"):
			code = ExitNotFound
		case strings.Contains(msg, "token") || strings.Contains(msg, "auth") || strings.Contains(msg, "OAuth"):
			code = ExitAuthError
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
