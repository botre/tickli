package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	theme  theme.Theme
	list   list.Model
	help   components.HelpBar
	result ProjectPickerResult
	width  int
	height int
}

func newProjectPickerModel(t theme.Theme, projects []types.Project, title string) projectPickerModel {
	delegate := components.NewProjectDelegate(t)
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = components.ProjectItem{
			Project: p,
		}
	}

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = t.Title
	l.Styles.FilterPrompt = t.FilterPrompt
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(t.Palette.Primary)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return projectPickerModel{
		theme: t,
		list:  l,
		help:  components.NewHelpBar(t),
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
		m.list.SetSize(m.width, m.height-1)
		return m, nil

	case tea.KeyMsg:
		if m.list.SettingFilter() {
			break
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if item, ok := m.list.SelectedItem().(components.ProjectItem); ok {
				m.result.Project = item.Project
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m projectPickerModel) View() string {
	if m.width == 0 {
		return ""
	}

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "select"},
		{Key: "/", Help: "filter"},
		{Key: "esc", Help: "cancel"},
	}
	return m.list.View() + "\n" + help.View()
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
