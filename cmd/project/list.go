package project

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	filter string
	output types.OutputFormat
}

func filterProjectByName(projects []types.Project, name string) ([]types.Project, error) {
	var matched []types.Project
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

func newListCommand(client *api.Client) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List and select from available projects",
		Long: `Display all available projects and allow selection of one.

This command shows all projects matching the optional filter criteria,
then displays a fuzzy-search selector to choose a project.`,
		Example: `  # List all projects
  tickli project list
  
  # Filter projects by name
  tickli project list -f "work"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := client.ListProjects()
			if err != nil {
				return errors.Wrap(err, "failed to fetch projects")
			}

			projects, err = filterProjectByName(projects, opts.filter)
			if err != nil {
				return err
			}

			if opts.output == types.OutputJSON {
				jsonData, err := json.MarshalIndent(projects, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
				return nil
			}

			project, err := utils.FuzzySelectProject(projects, "")
			if err != nil {
				return errors.Wrap(err, "failed to select project")
			}

			fmt.Println(fmt.Sprintf("(%s) %s", project.ID, project.Name))
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.filter, "filter", "f", "", "Only show projects with names containing the provided text")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
