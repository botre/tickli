// Package picker provides Bubble Tea pickers for interactive selection.
package picker

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

const previewHeight = 5

type taskPickerModel struct {
	theme        theme.Theme
	table        table.Model
	filter       textinput.Model
	help         components.HelpBar
	tasks        []types.Task
	projectNames []string
	allRows      []table.Row
	filteredIdx  []int
	result       TaskPickerResult
	showProject  bool
	width        int
	height       int
	title        string
}

func taskRow(tk types.Task, projectName string, showProject bool) table.Row {
	statusIcon := theme.IconPending
	if tk.Status == task.StatusComplete {
		statusIcon = theme.IconComplete
	}

	priorityFlag := " "
	if tk.Priority >= task.PriorityLow {
		priorityFlag = theme.IconPriority
	}

	dueStr := "—"
	due := time.Time(tk.DueDate)
	if !due.IsZero() {
		dueStr = tk.DueDate.Humanize()
	}

	tagStr := ""
	if len(tk.Tags) > 0 {
		tagStr = strings.Join(tk.Tags, ", ")
	}

	if showProject {
		return table.Row{statusIcon, priorityFlag, tk.Title, projectName, dueStr, tagStr}
	}
	return table.Row{statusIcon, priorityFlag, tk.Title, dueStr, tagStr}
}

func newTaskPickerModel(t theme.Theme, tasks []types.Task, projectNames []string, title string) taskPickerModel {
	showProject := len(projectNames) > 0
	rows := make([]table.Row, len(tasks))
	idx := make([]int, len(tasks))
	for i, tk := range tasks {
		name := ""
		if i < len(projectNames) {
			name = projectNames[i]
		}
		rows[i] = taskRow(tk, name, showProject)
		idx[i] = i
	}

	tbl := table.New(
		table.WithColumns(taskColumnsFor(0, showProject)),
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
		filter:       fi,
		help:         components.NewHelpBar(t),
		tasks:        tasks,
		projectNames: projectNames,
		allRows:      rows,
		filteredIdx:  idx,
		showProject:  showProject,
		title:        title,
	}
}

func (m taskPickerModel) Init() tea.Cmd {
	return textinput.Blink
}

// tableHeight returns the height available for the table.
// Layout: filter(1) + newline(1) + table + newline(1) + preview(previewHeight) + newline(1) + help(1)
func (m taskPickerModel) tableHeight() int {
	h := m.height - 1 - 1 - 1 - previewHeight - 1 - 1
	if h < 3 {
		h = 3
	}
	return h
}

func (m taskPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(m.width)
		m.table.SetHeight(m.tableHeight())
		m.resizeColumns()
		return m, nil

	case tea.KeyMsg:
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
				return m, tea.Quit
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
	query := m.filter.Value()
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
		if fuzzyMatchRow(row, query) {
			rows = append(rows, row)
			idx = append(idx, i)
		}
	}
	m.table.SetRows(rows)
	m.filteredIdx = idx
	m.table.GotoTop()
}

func (m *taskPickerModel) resizeColumns() {
	m.table.SetColumns(taskColumnsFor(m.width, m.showProject))
}

func taskColumnsFor(w int, showProject bool) []table.Column {
	if w < 40 {
		cols := []table.Column{
			{Title: "", Width: 1},
			{Title: "", Width: 1},
			{Title: "Title", Width: 30},
		}
		if showProject {
			cols = append(cols, table.Column{Title: "Project", Width: 16})
		}
		cols = append(cols,
			table.Column{Title: "Due", Width: 12},
			table.Column{Title: "Tags", Width: 14},
		)
		return cols
	}
	statusW := 3
	priorityW := 3
	projectW := 0
	if showProject {
		projectW = 16
	}
	dueW := 12
	tagsW := 14
	titleW := w - statusW - priorityW - projectW - dueW - tagsW - 12
	if titleW < 10 {
		titleW = 10
	}
	cols := []table.Column{
		{Title: "", Width: statusW},
		{Title: "", Width: priorityW},
		{Title: "Title", Width: titleW},
	}
	if showProject {
		cols = append(cols, table.Column{Title: "Project", Width: projectW})
	}
	cols = append(cols,
		table.Column{Title: "Due", Width: dueW},
		table.Column{Title: "Tags", Width: tagsW},
	)
	return cols
}

func (m taskPickerModel) renderPreview() string {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredIdx) {
		return ""
	}
	origIdx := m.filteredIdx[cursor]
	projectName := ""
	if origIdx < len(m.projectNames) {
		projectName = m.projectNames[origIdx]
	}
	return components.RenderTaskPreview(m.theme, m.tasks[origIdx], projectName, m.width-2)
}

func (m taskPickerModel) View() string {
	if m.width == 0 {
		return ""
	}

	help := m.help
	help.Width = m.width
	help.Bindings = []components.KeyBinding{
		{Key: "↑↓", Help: "navigate"},
		{Key: "⏎", Help: "select"},
		{Key: "esc", Help: "cancel"},
	}

	divider := m.theme.Divider.Render(strings.Repeat(theme.IconDivider, min(m.width-2, 40)))
	preview := lipgloss.NewStyle().Padding(0, 1).
		Height(previewHeight).
		MaxHeight(previewHeight).
		Render(m.renderPreview())

	return m.filter.View() + "\n" +
		m.table.View() + "\n" +
		divider + "\n" +
		preview + "\n" +
		help.View()
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
