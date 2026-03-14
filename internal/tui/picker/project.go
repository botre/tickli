package picker

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

type projectPickerModel struct {
	theme       theme.Theme
	table       table.Model
	filter      textinput.Model
	help        components.HelpBar
	projects    []types.Project
	allRows     []table.Row
	filteredIdx []int
	result      ProjectPickerResult
	filtering   bool
	width       int
	height      int
	title       string
}

func newProjectPickerModel(t theme.Theme, projects []types.Project, title string) projectPickerModel {
	rows := make([]table.Row, len(projects))
	idx := make([]int, len(projects))
	for i, p := range projects {
		viewMode := ""
		if p.ViewMode.String() != "" {
			viewMode = p.ViewMode.String()
		}
		rows[i] = table.Row{theme.IconProject, p.Name, viewMode}
		idx[i] = i
	}

	tbl := table.New(
		table.WithColumns(projectColumns(0)),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(tableStyles(t)),
	)

	fi := textinput.New()
	fi.Prompt = "/ "
	fi.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(string(t.Palette.Primary)))
	fi.Placeholder = "filter…"

	return projectPickerModel{
		theme:       t,
		table:       tbl,
		filter:      fi,
		help:        components.NewHelpBar(t),
		projects:    projects,
		allRows:     rows,
		filteredIdx: idx,
		title:       title,
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
		if m.filtering {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.filtering = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.applyFilter()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m.filtering = false
				m.filter.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.filter.Value() != "" {
				m.filter.SetValue("")
				m.applyFilter()
				return m, nil
			}
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.filteredIdx) {
				m.result.Project = m.projects[m.filteredIdx[cursor]]
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *projectPickerModel) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.table.SetRows(m.allRows)
		m.filteredIdx = make([]int, len(m.allRows))
		for i := range m.allRows {
			m.filteredIdx[i] = i
		}
		return
	}

	var rows []table.Row
	var idx []int
	for i, row := range m.allRows {
		match := false
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), query) {
				match = true
				break
			}
		}
		if match {
			rows = append(rows, row)
			idx = append(idx, i)
		}
	}
	m.table.SetRows(rows)
	m.filteredIdx = idx
	m.table.GotoTop()
}

func (m *projectPickerModel) resizeColumns() {
	m.table.SetColumns(projectColumns(m.width))
}

func projectColumns(w int) []table.Column {
	if w < 20 {
		return []table.Column{
			{Title: "", Width: 1},
			{Title: "Name", Width: 30},
			{Title: "View", Width: 10},
		}
	}
	dotW := 3
	viewW := 10
	nameW := w - dotW - viewW - 8
	if nameW < 10 {
		nameW = 10
	}
	return []table.Column{
		{Title: "", Width: dotW},
		{Title: "Name", Width: nameW},
		{Title: "View", Width: viewW},
	}
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
		{Key: "/", Help: "filter"},
		{Key: "esc", Help: "cancel"},
	}

	view := header + "\n"
	if m.filtering || m.filter.Value() != "" {
		view += m.filter.View() + "\n"
	}
	view += m.table.View() + "\n" + help.View()

	return view
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
