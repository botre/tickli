package project

import (
	"encoding/json"
	"fmt"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/completion"
	"github.com/botre/tickli/internal/prompt"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/project"
	"github.com/botre/tickli/internal/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type updateProjectOptions struct {
	projectID   string
	name        string
	color       project.Color
	viewMode    project.ViewMode
	kind        project.Kind
	interactive bool
	output      types.OutputFormat
}

func newUpdateProjectCommand(client *api.Client) *cobra.Command {
	opts := &updateProjectOptions{}
	cmd := &cobra.Command{
		Use:   "update <project>",
		Short: "Update a project",
		Long: `Modify the properties of an existing project.
    
You can update a project's name, color, view mode, or kind.
Changes only the properties you specify - others remain unchanged.`,
		Example: `  # Update project name
  tickli project update abc123def456 -n "New Project Name"
  
  # Update multiple properties
  tickli project update abc123def456 --name "Work Tasks" --color "#00FF00" --view-mode kanban
  
  # Update interactively
  tickli project update abc123def456 -i`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ProjectIDs(),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.projectID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := client.GetProject(opts.projectID)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to fetch project %s", opts.projectID))
			}

			if opts.interactive {
				newName := prompt.String("Name", p.Name)
				if newName != "" {
					p.Name = newName
				}

				colors := []string{"#3694FE (Default)", "#EC6665 (Red)", "#F2B04A (Orange)", "#FFD866 (Yellow)", "#5CD0A7 (Green)", "#9BECEC (Cyan)", "#4AA6EF (Blue)", "#CF66F6 (Purple)", "#EC70A5 (Pink)"}
				colorHexes := []string{"#3694FE", "#EC6665", "#F2B04A", "#FFD866", "#5CD0A7", "#9BECEC", "#4AA6EF", "#CF66F6", "#EC70A5"}
				idx, selectErr := prompt.Select("Color:", colors)
				if selectErr == nil {
					_ = p.Color.Set(colorHexes[idx])
				}

				viewModes := []string{"list", "kanban", "timeline"}
				idx, selectErr = prompt.Select("View mode:", viewModes)
				if selectErr == nil {
					_ = p.ViewMode.Set(viewModes[idx])
				}

				kinds := []string{"TASK", "NOTE"}
				idx, selectErr = prompt.Select("Kind:", kinds)
				if selectErr == nil {
					_ = p.Kind.Set(kinds[idx])
				}
			}

			if cmd.Flags().Changed("name") {
				p.Name = opts.name
			}
			if cmd.Flags().Changed("color") {
				p.Color = opts.color
			}
			if cmd.Flags().Changed("view-mode") {
				p.ViewMode = opts.viewMode
			}
			if cmd.Flags().Changed("kind") {
				p.Kind = opts.kind
			}
			p, err = client.UpdateProject(p)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to update project %s", opts.projectID))
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
				fmt.Printf("Project %s updated successfully\n", p.ID)
				fmt.Println(utils.GetProjectDescription(p))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Change the project name")
	cmd.Flags().VarP(&opts.color, "color", "c", "Change the project color (hex format, e.g., '#F18181')")
	cmd.Flags().Lookup("color").DefValue = ""
	_ = cmd.RegisterFlagCompletionFunc("color", project.ColorCompletionFunc)
	cmd.Flags().Var(&opts.viewMode, "view-mode", "Change how tasks are displayed: list, kanban, or timeline")
	cmd.Flags().Lookup("view-mode").DefValue = ""
	_ = cmd.RegisterFlagCompletionFunc("view-mode", project.ViewModeCompletionFunc)
	cmd.Flags().Var(&opts.kind, "kind", "Change project type: TASK or NOTE")
	cmd.Flags().Lookup("kind").DefValue = ""
	_ = cmd.RegisterFlagCompletionFunc("kind", project.KindCompletionFunc)
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Update project by answering prompts instead of using flags")
	cmd.Flags().VarP(&opts.output, "output", "o", "Display format: simple (human-readable) or json (machine-readable)")
	_ = cmd.RegisterFlagCompletionFunc("output", types.OutputFormatCompletionFunc)

	return cmd
}
