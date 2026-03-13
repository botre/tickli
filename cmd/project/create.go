package project

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type createProjectOptions struct {
	name        string
	color       project.Color
	viewMode    project.ViewMode
	kind        project.Kind
	interactive bool
	output      types.OutputFormat
}

func newCreateProjectCommand(client *api.Client) *cobra.Command {
	opts := &createProjectOptions{
		kind:     project.KindTask,
		viewMode: project.ViewModeList,
		color:    project.DefaultColor,
	}
	cmd := &cobra.Command{
		Use:   "create [project-name]",
		Short: "Create a new project",
		Long: `Create a new TickTick project with customizable properties.
    
You can specify a name, color, view mode, and project type. The command
supports both direct parameter input and interactive mode.`,
		Example: `  # Create a basic task project
  tickli project create "My New Project"
  
  # Create a project with custom properties
  tickli project create -n "Work Tasks" -C "#FF5733" --view-mode kanban --kind TASK
  
  # Create a project in interactive mode
  tickli project create -i`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.interactive {
				if !prompt.IsInteractive() {
					return fmt.Errorf("--interactive requires a terminal (stdin is not a TTY)")
				}
				opts.name = prompt.String("Project name", opts.name)
				if opts.name == "" {
					return fmt.Errorf("project name is required")
				}

				colors := []string{"#3694FE (Default)", "#EC6665 (Red)", "#F2B04A (Orange)", "#FFD866 (Yellow)", "#5CD0A7 (Green)", "#9BECEC (Cyan)", "#4AA6EF (Blue)", "#CF66F6 (Purple)", "#EC70A5 (Pink)"}
				colorHexes := []string{"#3694FE", "#EC6665", "#F2B04A", "#FFD866", "#5CD0A7", "#9BECEC", "#4AA6EF", "#CF66F6", "#EC70A5"}
				idx, err := prompt.Select("Color:", colors)
				if err == nil {
					_ = opts.color.Set(colorHexes[idx])
				}

				viewModes := []string{"list", "kanban", "timeline"}
				idx, err = prompt.Select("View mode:", viewModes)
				if err == nil {
					_ = opts.viewMode.Set(viewModes[idx])
				}

				kinds := []string{"TASK", "NOTE"}
				idx, err = prompt.Select("Kind:", kinds)
				if err == nil {
					_ = opts.kind.Set(kinds[idx])
				}
			}

			p := &types.Project{
				Name:     opts.name,
				Color:    opts.color,
				ViewMode: opts.viewMode,
				Kind:     opts.kind,
			}

			p, err := client.CreateProject(p)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to create project %q", opts.name))
			}

			switch resolveOutput(cmd, opts.output) {
			case types.OutputJSON:
				jsonData, err := json.MarshalIndent(p, "", "  ")
				if err != nil {
					return errors.Wrap(err, "failed to marshal output")
				}
				fmt.Println(string(jsonData))
			case types.OutputQuiet:
				fmt.Println(p.ID)
			default:
				fmt.Println(utils.GetProjectDescription(*p))
				fmt.Println(p.ID)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Name of the new project (required unless -i)")
	cmd.Flags().VarP(&opts.color, "color", "C", "Color for the project (hex format, e.g., '#F18181')")
	_ = cmd.RegisterFlagCompletionFunc("color", project.ColorCompletionFunc)
	cmd.Flags().Var(&opts.viewMode, "view-mode", "How to display tasks: list, kanban, or timeline (default: list)")
	_ = cmd.RegisterFlagCompletionFunc("view-mode", project.ViewModeCompletionFunc)
	cmd.Flags().Var(&opts.kind, "kind", "Project type: TASK for action items or NOTE for information (default: TASK)")
	_ = cmd.RegisterFlagCompletionFunc("kind", project.KindCompletionFunc)
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Create project by answering prompts instead of using flags")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)
	cmd.MarkFlagsOneRequired("name", "interactive")

	return cmd
}
