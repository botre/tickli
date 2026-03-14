package project

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/config"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/tui/forms"
	"github.com/botre/tickli/internal/tui/render"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
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
				t := theme.Default()
				kindStr := string(opts.kind)
				if kindStr == "" {
					kindStr = "TASK"
				}
				defaultColor := opts.color.String()
				if defaultColor == "" {
					if cfg, err := config.Load(); err == nil && cfg.DefaultProjectColor != "" {
						defaultColor = cfg.DefaultProjectColor
					}
				}
				result, err := forms.RunProjectCreateForm(t, forms.ProjectFormResult{
					Name:     opts.name,
					Color:    defaultColor,
					ViewMode: string(opts.viewMode),
					Kind:     kindStr,
				})
				if err != nil {
					return fmt.Errorf("form cancelled: %w", err)
				}
				opts.name = result.Name
				if result.Color != "" {
					_ = opts.color.Set(result.Color)
				}
				_ = opts.viewMode.Set(result.ViewMode)
				_ = opts.kind.Set(result.Kind)
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
				r := render.New()
				fmt.Println(r.SuccessMessage(fmt.Sprintf("Created project %s", p.ID)))
				fmt.Println(r.ProjectDetail(*p))
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
