package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/components"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
)

// ProjectPickerResult holds the result of a project picker session.
type ProjectPickerResult struct {
	Project   types.Project
	Cancelled bool
}

// projectPickerModel is the Bubble Tea model for the project picker.
type projectPickerModel struct {
	theme    theme.Theme
	table    table.Model
	help     components.HelpBar
	projects []types.Project
	result   ProjectPickerResult
	width    int
	height   int
	title    string
}

func newProjectPickerModel(t theme.Theme, projects []types.Project, title string) projectPickerModel {
	columns := []table.Column{
		{Title: "", Width: 1},         // color dot
		{Title: "Name", Width: 30},
		{Title: "View", Width: 10},
	}

	rows := make([]table.Row, len(projects))
	for i, p := range projects {
		colorDot := theme.IconProject
		viewMode := ""
		if p.ViewMode.String() != "" {
			viewMode = p.ViewMode.String()
		}
		rows[i] = table.Row{colorDot, p.Name, viewMode}
	}

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.SubText))).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(string(t.Palette.Subtle))).
		BorderBottom(true).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(string(t.Palette.Primary))).
		Bold(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Padding(0, 1)

	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(s),
	)

	return projectPickerModel{
		theme:    t,
		table:    tbl,
		help:     components.NewHelpBar(t),
		projects: projects,
		title:    title,
	}
}

func (m projectPickerModel) Init() tea.Cmd {
	return nil
}

func (m projectPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(m.width)
		m.table.SetHeight(m.height - 3)
		m.resizeColumns()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.projects) {
				m.result.Project = m.projects[idx]
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *projectPickerModel) resizeColumns() {
	w := m.width
	if w < 20 {
		return
	}
	dotW := 3
	viewW := 10
	nameW := w - dotW - viewW - 8 // padding
	if nameW < 10 {
		nameW = 10
	}
	m.table.SetColumns([]table.Column{
		{Title: "", Width: dotW},
		{Title: "Name", Width: nameW},
		{Title: "View", Width: viewW},
	})
}

func (m projectPickerModel) View() string {
	if m.width == 0 {
		return ""
	}

	titleStyle := m.theme.Title.Padding(0, 0, 0, 1)
	header := titleStyle.Render(m.title)

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "select"},
		{Key: "esc", Help: "cancel"},
	}

	return header + "\n" + m.table.View() + "\n" + help.View()
}

// RunProjectPicker launches a Bubble Tea project picker and returns the selected project.
func RunProjectPicker(projects []types.Project, title string) (*ProjectPickerResult, error) {
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects available for selection")
	}

	t := theme.Default()
	model := newProjectPickerModel(t, projects, title)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := final.(projectPickerModel).result
	if result.Cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}
	return &result, nil
}
