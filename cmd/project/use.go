package project

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func findProjectByID(projects []types.Project, id string) (types.Project, error) {
	for i := range projects {
		if projects[i].ID == id {
			return projects[i], nil
		}
	}
	return types.Project{}, fmt.Errorf("project not found with ID '%s'", id)
}

func findProjectsByName(projects []*types.Project, name string) ([]*types.Project, error) {
	var matched []*types.Project
	nameLower := strings.ToLower(name)
	for i := range projects {
		if strings.Contains(strings.ToLower(projects[i].Name), nameLower) {
			matched = append(matched, projects[i])
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no project found with name '%s'", name)
	}
	return matched, nil
}

type useProjectOptions struct {
	projectID string
	output    types.OutputFormat
}

func newUseProjectCmd(client *api.Client) *cobra.Command {
	opts := &useProjectOptions{}
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Switch active project context",
		Long: `Switch the active project context for subsequent commands.

Without arguments, opens an interactive selector. With a project ID argument,
switches directly. The selected project becomes the default for future commands.`,
		Example: `  # Interactive project selection
  tickli project use

  # Switch by project ID
  tickli project use abc123def456`,
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
				project, err := findProjectByID(projects, opts.projectID)
				if err != nil {
					return err
				}
				selectedProject = project
			} else {
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

			cfg.DefaultProjectID = selectedProject.ID
			if err := config.Save(cfg); err != nil {
				return errors.Wrap(err, "failed to save config")
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				result := map[string]string{"id": selectedProject.ID, "name": selectedProject.Name}
				jsonData, _ := json.MarshalIndent(result, "", "  ")
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
