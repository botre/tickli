package project

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type useProjectOptions struct {
	projectID string
	output    types.OutputFormat
}

func newUseProjectCmd(client *api.Client) *cobra.Command {
	opts := &useProjectOptions{}
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Set the active project",
		Long: `Switch the active project context for subsequent commands.

Without arguments, opens an interactive selector. With a project argument,
switches directly. The selected project becomes the default for future commands.`,
		Example: `  # Interactive project selection
  tickli project use

  # Switch by project name or ID
  tickli project use Chores`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.ProjectIDs(),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				opts.projectID = args[0]
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := client.ListProjects()
			if err != nil {
				return errors.Wrap(err, "could not fetch projects")
			}

			var selectedProject types.Project

			if opts.projectID != "" {
				project, err := client.ResolveProject(opts.projectID)
				if err != nil {
					return err
				}
				selectedProject = project
			} else {
				if !prompt.IsInteractive() {
					return fmt.Errorf("project argument required in non-interactive mode. Run 'tickli project list -o json' to see available projects")
				}
				project, err := utils.FuzzySelectProject(projects, "")
				if err != nil {
					return errors.Wrap(err, "could not select project")
				}
				selectedProject = project
			}

			cfg, err := config.Load()
			if err != nil {
				return errors.Wrap(err, "could not load config")
			}

			cfg.DefaultProject = selectedProject.ID
			if err := config.Save(cfg); err != nil {
				return errors.Wrap(err, "failed to save config")
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				jsonData, _ := json.MarshalIndent(selectedProject, "", "  ")
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(selectedProject.ID)
			default:
				fmt.Printf("Switched to project %s (%s)\n", selectedProject.Name, selectedProject.ID)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
