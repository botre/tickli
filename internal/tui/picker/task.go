// Package picker provides Bubble Tea list pickers that replace go-fuzzyfinder.
package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/components"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
)

// TaskPickerResult holds the result of a task picker session.
type TaskPickerResult struct {
	Task        types.Task
	ProjectName string
	Cancelled   bool
}

// TaskItem wraps a task for the picker list.
type taskPickerItem struct {
	task        types.Task
	projectName string
}

func (t taskPickerItem) Title() string       { return t.task.Title }
func (t taskPickerItem) Description() string { return t.projectName }
func (t taskPickerItem) FilterValue() string { return t.task.Title + " " + t.projectName }

// taskPickerModel is the Bubble Tea model for the task picker.
type taskPickerModel struct {
	theme      theme.Theme
	list       list.Model
	detail     viewport.Model
	help       components.HelpBar
	result     TaskPickerResult
	showDetail bool
	width      int
	height     int
	title      string
}

func newTaskPickerModel(t theme.Theme, tasks []taskPickerItem, title string) taskPickerModel {
	delegate := components.NewTaskDelegate(t, true)
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = components.TaskItem{
			Task:        task.task,
			ProjectName: task.projectName,
		}
	}

	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.Styles.Title = t.Title
	l.Styles.FilterPrompt = t.FilterPrompt
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(t.Palette.Primary)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	detail := viewport.New(0, 0)

	return taskPickerModel{
		theme:  t,
		list:   l,
		detail: detail,
		help:   components.NewHelpBar(t),
		title:  title,
	}
}

func (m taskPickerModel) Init() tea.Cmd {
	return nil
}

func (m taskPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.showDetail {
			m.detail.Width = m.width - 2
			m.detail.Height = m.height - 2
		} else {
			m.list.SetSize(m.width, m.height-1) // -1 for help bar
		}
		return m, nil

	case tea.KeyMsg:
		// Don't intercept keys when filtering
		if m.list.SettingFilter() {
			break
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.showDetail {
				m.showDetail = false
				m.list.SetSize(m.width, m.height-1)
				return m, nil
			}
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.showDetail {
				// Second enter = confirm selection
				return m, tea.Quit
			}
			// First enter = show detail
			if item, ok := m.list.SelectedItem().(components.TaskItem); ok {
				m.result.Task = item.Task
				m.result.ProjectName = item.ProjectName
				m.showDetail = true
				content := components.RenderTaskDetail(m.theme, item.Task, item.ProjectName, m.width-4)
				m.detail.SetContent(content)
				m.detail.GotoTop()
				m.detail.Width = m.width - 2
				m.detail.Height = m.height - 2
				return m, nil
			}
		}
	}

	if m.showDetail {
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m taskPickerModel) View() string {
	if m.width == 0 {
		return ""
	}

	if m.showDetail {
		help := m.help
		help.Width = m.width
		help.Bindings = []components.KeyBinding{
			{Key: "⏎", Help: "confirm"},
			{Key: "esc", Help: "back to list"},
			{Key: "↑↓", Help: "scroll"},
		}
		return lipgloss.NewStyle().Padding(0, 1).Render(m.detail.View()) + "\n" + help.View()
	}

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "view details"},
		{Key: "/", Help: "filter"},
		{Key: "esc", Help: "cancel"},
	}
	return m.list.View() + "\n" + help.View()
}

// RunTaskPicker launches a Bubble Tea task picker and returns the selected task.
// showProject controls whether the project name is shown next to each task.
func RunTaskPicker(tasks []types.Task, projectNames []string, title string) (*TaskPickerResult, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}

	t := theme.Default()
	items := make([]taskPickerItem, len(tasks))
	for i, task := range tasks {
		name := ""
		if i < len(projectNames) {
			name = projectNames[i]
		}
		items[i] = taskPickerItem{task: task, projectName: name}
	}

	model := newTaskPickerModel(t, items, title)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := final.(taskPickerModel).result
	if result.Cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}
	return &result, nil
}
