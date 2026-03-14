// Package picker provides Bubble Tea pickers for interactive selection.
package picker

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/tui/components"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// TaskPickerResult holds the result of a task picker session.
type TaskPickerResult struct {
	Task        types.Task
	ProjectName string
	Cancelled   bool
}

// taskPickerModel is the Bubble Tea model for the task picker.
type taskPickerModel struct {
	theme        theme.Theme
	table        table.Model
	detail       viewport.Model
	help         components.HelpBar
	tasks        []types.Task
	projectNames []string
	result       TaskPickerResult
	showDetail   bool
	width        int
	height       int
	title        string
}

func newTaskPickerModel(t theme.Theme, tasks []types.Task, projectNames []string, title string) taskPickerModel {
	columns := []table.Column{
		{Title: "", Width: 1},          // status icon
		{Title: "", Width: 1},          // priority
		{Title: "Title", Width: 30},
		{Title: "Project", Width: 16},
		{Title: "Due", Width: 12},
		{Title: "Tags", Width: 14},
	}

	rows := make([]table.Row, len(tasks))
	for i, tk := range tasks {
		statusIcon := theme.IconPending
		if tk.Status == task.StatusComplete {
			statusIcon = theme.IconComplete
		}

		priorityFlag := " "
		if tk.Priority >= task.PriorityLow {
			priorityFlag = theme.IconPriority
		}

		projectName := ""
		if i < len(projectNames) {
			projectName = projectNames[i]
		}

		dueStr := ""
		due := time.Time(tk.DueDate)
		if !due.IsZero() {
			dueStr = tk.DueDate.Humanize()
		}

		tagStr := ""
		if len(tk.Tags) > 0 {
			tagStr = strings.Join(tk.Tags, ", ")
		}

		rows[i] = table.Row{statusIcon, priorityFlag, tk.Title, projectName, dueStr, tagStr}
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

	detail := viewport.New(0, 0)

	return taskPickerModel{
		theme:        t,
		table:        tbl,
		detail:       detail,
		help:         components.NewHelpBar(t),
		tasks:        tasks,
		projectNames: projectNames,
		title:        title,
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
			m.table.SetWidth(m.width)
			m.table.SetHeight(m.height - 3) // title + help
			m.resizeColumns()
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.showDetail {
				m.showDetail = false
				m.table.SetWidth(m.width)
				m.table.SetHeight(m.height - 3)
				m.resizeColumns()
				return m, nil
			}
			m.result.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.showDetail {
				return m, tea.Quit
			}
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.tasks) {
				m.result.Task = m.tasks[idx]
				if idx < len(m.projectNames) {
					m.result.ProjectName = m.projectNames[idx]
				}
				m.showDetail = true
				content := components.RenderTaskDetail(m.theme, m.tasks[idx], m.result.ProjectName, m.width-4)
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
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *taskPickerModel) resizeColumns() {
	w := m.width
	if w < 40 {
		return
	}
	// Fixed widths for small columns
	statusW := 3
	priorityW := 3
	projectW := 16
	dueW := 12
	tagsW := 14
	// Title gets the rest
	titleW := w - statusW - priorityW - projectW - dueW - tagsW - 12 // padding
	if titleW < 10 {
		titleW = 10
	}
	m.table.SetColumns([]table.Column{
		{Title: "", Width: statusW},
		{Title: "", Width: priorityW},
		{Title: "Title", Width: titleW},
		{Title: "Project", Width: projectW},
		{Title: "Due", Width: dueW},
		{Title: "Tags", Width: tagsW},
	})
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

	titleStyle := m.theme.Title.Padding(0, 0, 1, 1)
	header := titleStyle.Render(m.title)

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "view details"},
		{Key: "esc", Help: "cancel"},
	}

	return header + "\n" + m.table.View() + "\n" + help.View()
}

// RunTaskPicker launches a Bubble Tea task picker and returns the selected task.
func RunTaskPicker(tasks []types.Task, projectNames []string, title string) (*TaskPickerResult, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}

	t := theme.Default()
	model := newTaskPickerModel(t, tasks, projectNames, title)
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
