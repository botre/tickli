// Package picker provides Bubble Tea pickers for interactive selection.
package picker

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

type taskPickerModel struct {
	theme        theme.Theme
	table        table.Model
	detail       viewport.Model
	filter       textinput.Model
	help         components.HelpBar
	tasks        []types.Task
	projectNames []string
	allRows      []table.Row
	filteredIdx  []int
	result       TaskPickerResult
	showDetail   bool
	width        int
	height       int
	title        string
}

func taskRow(tk types.Task, projectName string) table.Row {
	statusIcon := theme.IconPending
	if tk.Status == task.StatusComplete {
		statusIcon = theme.IconComplete
	}

	priorityFlag := " "
	if tk.Priority >= task.PriorityLow {
		priorityFlag = theme.IconPriority
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

	return table.Row{statusIcon, priorityFlag, tk.Title, projectName, dueStr, tagStr}
}

func newTaskPickerModel(t theme.Theme, tasks []types.Task, projectNames []string, title string) taskPickerModel {
	rows := make([]table.Row, len(tasks))
	idx := make([]int, len(tasks))
	for i, tk := range tasks {
		name := ""
		if i < len(projectNames) {
			name = projectNames[i]
		}
		rows[i] = taskRow(tk, name)
		idx[i] = i
	}

	tbl := table.New(
		table.WithColumns(taskColumns(0)),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(tableStyles(t)),
	)

	fi := textinput.New()
	fi.Prompt = "> "
	fi.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(string(t.Palette.Primary)))
	fi.Placeholder = "type to filter…"
	fi.Focus()

	return taskPickerModel{
		theme:        t,
		table:        tbl,
		detail:       viewport.New(0, 0),
		filter:       fi,
		help:         components.NewHelpBar(t),
		tasks:        tasks,
		projectNames: projectNames,
		allRows:      rows,
		filteredIdx:  idx,
		title:        title,
	}
}

func (m taskPickerModel) Init() tea.Cmd {
	return textinput.Blink
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
			m.table.SetHeight(m.height - 4)
			m.resizeColumns()
		}
		return m, nil

	case tea.KeyMsg:
		if m.showDetail {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.showDetail = false
				m.table.SetWidth(m.width)
				m.table.SetHeight(m.height - 4)
				m.resizeColumns()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				m.result.Cancelled = true
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
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

		case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
			m.table.MoveUp(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			m.table.MoveDown(1)
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.filteredIdx) {
				origIdx := m.filteredIdx[cursor]
				m.result.Task = m.tasks[origIdx]
				if origIdx < len(m.projectNames) {
					m.result.ProjectName = m.projectNames[origIdx]
				}
				m.showDetail = true
				content := components.RenderTaskDetail(m.theme, m.tasks[origIdx], m.result.ProjectName, m.width-4)
				m.detail.SetContent(content)
				m.detail.GotoTop()
				m.detail.Width = m.width - 2
				m.detail.Height = m.height - 2
				return m, nil
			}

		default:
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			m.applyFilter()
			return m, cmd
		}
	}

	return m, nil
}

func (m *taskPickerModel) applyFilter() {
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

func (m *taskPickerModel) resizeColumns() {
	m.table.SetColumns(taskColumns(m.width))
}

func taskColumns(w int) []table.Column {
	if w < 40 {
		return []table.Column{
			{Title: "", Width: 1},
			{Title: "", Width: 1},
			{Title: "Title", Width: 30},
			{Title: "Project", Width: 16},
			{Title: "Due", Width: 12},
			{Title: "Tags", Width: 14},
		}
	}
	statusW := 3
	priorityW := 3
	projectW := 16
	dueW := 12
	tagsW := 14
	titleW := w - statusW - priorityW - projectW - dueW - tagsW - 12
	if titleW < 10 {
		titleW = 10
	}
	return []table.Column{
		{Title: "", Width: statusW},
		{Title: "", Width: priorityW},
		{Title: "Title", Width: titleW},
		{Title: "Project", Width: projectW},
		{Title: "Due", Width: dueW},
		{Title: "Tags", Width: tagsW},
	}
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
			{Key: "esc", Help: "back"},
			{Key: "↑↓", Help: "scroll"},
		}
		return lipgloss.NewStyle().Padding(0, 1).Render(m.detail.View()) + "\n" + help.View()
	}

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "select"},
		{Key: "esc", Help: "cancel"},
	}

	return m.filter.View() + "\n" + m.table.View() + "\n" + help.View()
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
