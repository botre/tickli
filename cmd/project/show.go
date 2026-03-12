package project

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct {
	projectID string
	withTasks bool
	output    types.OutputFormat
}

func newShowCommand(client *api.Client) *cobra.Command {
	opts := &showOptions{
		output: types.OutputSimple,
	}

	cmd := &cobra.Command{
		Use:     "show [project-id]",
		Aliases: []string{"info", "get"},
		Short:   "Show details of a project",
		Long: `Display detailed information about a specific project.
    
If no project ID is provided, shows the currently active project.
Can include associated tasks and switch between output formats.`,
		Example: `  # Show current project
  tickli project show
  
  # Show specific project
  tickli project show abc123def456
  
  # Show project with all its tasks
  tickli project show --with-tasks
  
  # Output in JSON format
  tickli project show -o json`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.ProjectIDs(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.projectID = args[0]
			} else {
				cfg, err := config.Load()
				if err != nil {
					return errors.Wrap(err, "failed to load config")
				}
				opts.projectID = cfg.DefaultProjectID
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			output := resolveOutput(cmd, opts.output)
			if opts.withTasks {
				projectData, err := client.GetProjectWithTasks(opts.projectID)
				if err != nil {
					return errors.Wrap(err, "failed to get project data")
				}
				switch output {
				case types.OutputJSON:
					jsonData, err := json.MarshalIndent(projectData, "", "  ")
					if err != nil {
						return errors.Wrap(err, "failed to marshal output")
					}
					fmt.Println(string(jsonData))
				case types.OutputQuiet:
					fmt.Println(projectData.Project.ID)
				default:
					fmt.Println(utils.GetProjectDescription(projectData.Project))
					for _, task := range projectData.Tasks {
						fmt.Println(utils.GetTaskDescription(task, projectData.Project.Color))
					}
				}
			} else {
				project, err := client.GetProject(opts.projectID)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to get project %s", opts.projectID))
				}
				switch output {
				case types.OutputJSON:
					jsonData, err := json.MarshalIndent(project, "", "  ")
					if err != nil {
						return errors.Wrap(err, "failed to marshal output")
					}
					fmt.Println(string(jsonData))
				case types.OutputQuiet:
					fmt.Println(project.ID)
				default:
					fmt.Println(utils.GetProjectDescription(project))
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&opts.withTasks, "with-tasks", false, "Include all tasks belonging to this project")
	cmd.Flags().VarP(&opts.output, "output", "o", "Format for displaying results: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)
	return cmd
}
