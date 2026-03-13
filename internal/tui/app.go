package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/botre/tickli/internal/api"
	"github.com/botre/tickli/internal/tui/components"
	"github.com/botre/tickli/internal/tui/theme"
	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

// ViewState tracks which screen the app is showing.
type ViewState int

const (
	ViewLoading ViewState = iota
	ViewDashboard
	ViewTaskList
	ViewTaskDetail
	ViewProjectList
)

// SmartView represents which smart view filter is active.
type SmartView int

const (
	SmartViewToday SmartView = iota
	SmartViewTomorrow
	SmartViewWeek
	SmartViewInbox
	SmartViewAll
)

func (s SmartView) String() string {
	switch s {
	case SmartViewToday:
		return "Today"
	case SmartViewTomorrow:
		return "Tomorrow"
	case SmartViewWeek:
		return "This Week"
	case SmartViewInbox:
		return "Inbox"
	case SmartViewAll:
		return "All Tasks"
	default:
		return ""
	}
}

// Messages for async operations
type tasksLoadedMsg struct {
	tasks    []projectTask
	err      error
}

type projectsLoadedMsg struct {
	projects []projectWithCount
	err      error
}

type taskCompletedMsg struct {
	taskID string
	err    error
}

type projectTask struct {
	types.Task
	ProjectName string
}

type projectWithCount struct {
	types.Project
	TaskCount int
}

// Model is the main Bubble Tea model for the TUI app.
type Model struct {
	client *api.Client
	theme  theme.Theme
	keys   KeyMap
	width  int
	height int

	// State
	view       ViewState
	smartView  SmartView
	prevView   ViewState
	loading    bool
	err        error
	statusMsg  string

	// Data
	allTasks []projectTask
	projects []projectWithCount

	// Sub-models
	taskList    list.Model
	projectList list.Model
	detailView  viewport.Model
	spinner     components.LoadingSpinner

	// Header + status bar
	header    components.Header
	statusBar components.StatusBar
	helpBar   components.HelpBar

	// Selected state
	selectedTask *projectTask
}

// New creates a new TUI app model.
func New(client *api.Client) Model {
	t := theme.Default()
	keys := DefaultKeyMap()

	// Task list
	taskDelegate := components.NewTaskDelegate(t, true)
	taskList := list.New(nil, taskDelegate, 0, 0)
	taskList.SetShowTitle(false)
	taskList.SetShowStatusBar(true)
	taskList.SetShowFilter(true)
	taskList.SetFilteringEnabled(true)
	taskList.Styles.Title = t.Title
	taskList.Styles.FilterPrompt = t.FilterPrompt
	taskList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(t.Palette.Primary)
	taskList.StatusMessageLifetime = 3 * time.Second

	// Project list
	projectDelegate := components.NewProjectDelegate(t)
	projectList := list.New(nil, projectDelegate, 0, 0)
	projectList.SetShowTitle(false)
	projectList.SetShowStatusBar(true)
	projectList.SetShowFilter(true)
	projectList.SetFilteringEnabled(true)
	projectList.Styles.Title = t.Title
	projectList.Styles.FilterPrompt = t.FilterPrompt
	projectList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(t.Palette.Primary)

	// Detail viewport
	detailView := viewport.New(0, 0)

	// Spinner
	spinner := components.NewLoadingSpinner(t, "Loading tasks…")

	return Model{
		client:      client,
		theme:       t,
		keys:        keys,
		view:        ViewLoading,
		smartView:   SmartViewToday,
		loading:     true,
		taskList:    taskList,
		projectList: projectList,
		detailView:  detailView,
		spinner:     spinner,
		header:      components.NewHeader(t),
		statusBar:   components.NewStatusBar(t),
		helpBar:     components.NewHelpBar(t),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Init(),
		m.loadTasksCmd(),
		m.loadProjectsCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateLayout()

	case tea.KeyMsg:
		// Global keys that work regardless of view
		if key.Matches(msg, m.keys.Quit) && !m.taskList.SettingFilter() && !m.projectList.SettingFilter() {
			return m, tea.Quit
		}

		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tasksLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.allTasks = msg.tasks
			m = m.applySmartViewFilter()
			if m.view == ViewLoading {
				m.view = ViewDashboard
			}
		}

	case projectsLoadedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error loading projects: %v", msg.err)
		} else {
			m.projects = msg.projects
			m = m.updateProjectList()
		}

	case taskCompletedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.statusMsg = "Task completed!"
			// Remove from local data
			for i, t := range m.allTasks {
				if t.ID == msg.taskID {
					m.allTasks[i].Status = task.StatusComplete
					break
				}
			}
			m = m.applySmartViewFilter()
		}
	}

	// Update active sub-model
	switch m.view {
	case ViewLoading:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case ViewDashboard, ViewTaskList:
		var cmd tea.Cmd
		m.taskList, cmd = m.taskList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewProjectList:
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		cmds = append(cmds, cmd)
	case ViewTaskDetail:
		var cmd tea.Cmd
		m.detailView, cmd = m.detailView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Don't handle nav keys when filtering
	if m.taskList.SettingFilter() || m.projectList.SettingFilter() {
		return nil
	}

	switch {
	case key.Matches(msg, m.keys.Back):
		return m.goBack()

	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()

	case key.Matches(msg, m.keys.Complete):
		return m.handleComplete()

	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		m.statusMsg = "Refreshing…"
		return tea.Batch(m.loadTasksCmd(), m.loadProjectsCmd())

	case key.Matches(msg, m.keys.Today):
		m.smartView = SmartViewToday
		m.view = ViewDashboard
		m = m.applySmartViewFilter()

	case key.Matches(msg, m.keys.Tomorrow):
		m.smartView = SmartViewTomorrow
		m.view = ViewDashboard
		m = m.applySmartViewFilter()

	case key.Matches(msg, m.keys.Week):
		m.smartView = SmartViewWeek
		m.view = ViewDashboard
		m = m.applySmartViewFilter()

	case key.Matches(msg, m.keys.Inbox):
		m.smartView = SmartViewInbox
		m.view = ViewDashboard
		m = m.applySmartViewFilter()

	case key.Matches(msg, m.keys.All):
		m.smartView = SmartViewAll
		m.view = ViewDashboard
		m = m.applySmartViewFilter()

	case key.Matches(msg, m.keys.Projects):
		m.prevView = m.view
		m.view = ViewProjectList
	}
	return nil
}

func (m *Model) goBack() tea.Cmd {
	switch m.view {
	case ViewTaskDetail:
		m.view = m.prevView
		m.selectedTask = nil
	case ViewProjectList:
		m.view = ViewDashboard
	case ViewTaskList:
		m.view = ViewDashboard
	default:
		return tea.Quit
	}
	return nil
}

func (m *Model) handleEnter() tea.Cmd {
	switch m.view {
	case ViewDashboard, ViewTaskList:
		if item, ok := m.taskList.SelectedItem().(components.TaskItem); ok {
			pt := projectTask{Task: item.Task, ProjectName: item.ProjectName}
			m.selectedTask = &pt
			m.prevView = m.view
			m.view = ViewTaskDetail
			content := components.RenderTaskDetail(m.theme, item.Task, item.ProjectName, m.width-4)
			m.detailView.SetContent(content)
			m.detailView.GotoTop()
		}
	case ViewProjectList:
		if item, ok := m.projectList.SelectedItem().(components.ProjectItem); ok {
			// Switch to tasks for this project
			m.smartView = SmartViewAll
			m.view = ViewTaskList
			var filtered []projectTask
			for _, t := range m.allTasks {
				if t.ProjectID == item.Project.ID {
					filtered = append(filtered, t)
				}
			}
			m = m.setTaskListItems(filtered)
		}
	}
	return nil
}

func (m *Model) handleComplete() tea.Cmd {
	if m.view != ViewDashboard && m.view != ViewTaskList {
		return nil
	}
	item, ok := m.taskList.SelectedItem().(components.TaskItem)
	if !ok {
		return nil
	}
	if item.Task.Status == task.StatusComplete {
		return nil
	}
	taskID := item.Task.ID
	client := m.client
	return func() tea.Msg {
		err := client.CompleteTask(taskID)
		return taskCompletedMsg{taskID: taskID, err: err}
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Header
	m.header.Width = m.width
	m.header.Breadcrumb = m.breadcrumb()
	header := m.header.View()

	// Help bar
	m.helpBar.Width = m.width
	m.helpBar.Bindings = m.currentBindings()
	helpBar := m.helpBar.View()

	// Status bar
	m.statusBar.Width = m.width
	m.statusBar.Left = "tickli"
	m.statusBar.Center = m.statusMsg
	m.statusBar.Right = m.smartView.String()
	statusBar := m.statusBar.View()

	// Content area
	headerH := lipgloss.Height(header)
	helpH := lipgloss.Height(helpBar)
	statusH := lipgloss.Height(statusBar)
	contentH := m.height - headerH - helpH - statusH

	var content string
	switch m.view {
	case ViewLoading:
		content = lipgloss.Place(m.width, contentH,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View())
	case ViewDashboard, ViewTaskList:
		content = m.renderTaskView(contentH)
	case ViewTaskDetail:
		content = m.renderDetailView(contentH)
	case ViewProjectList:
		content = m.renderProjectView(contentH)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		helpBar,
		statusBar,
	)
}

func (m Model) renderTaskView(height int) string {
	// Tabs for smart views
	tabs := m.renderTabs()
	tabH := lipgloss.Height(tabs)

	listH := height - tabH
	if listH < 1 {
		listH = 1
	}

	m.taskList.SetSize(m.width, listH)
	return tabs + "\n" + m.taskList.View()
}

func (m Model) renderDetailView(height int) string {
	m.detailView.Width = m.width - 2
	m.detailView.Height = height
	return lipgloss.NewStyle().Padding(0, 1).Render(m.detailView.View())
}

func (m Model) renderProjectView(height int) string {
	m.projectList.SetSize(m.width, height)
	return m.projectList.View()
}

func (m Model) renderTabs() string {
	views := []struct {
		view SmartView
		name string
	}{
		{SmartViewToday, "Today"},
		{SmartViewTomorrow, "Tomorrow"},
		{SmartViewWeek, "Week"},
		{SmartViewInbox, "Inbox"},
		{SmartViewAll, "All"},
	}

	activeTab := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Primary).
		Bold(true).
		Underline(true).
		Padding(0, 2)

	inactiveTab := lipgloss.NewStyle().
		Foreground(m.theme.Palette.Muted).
		Padding(0, 2)

	var tabs []string
	for _, v := range views {
		if v.view == m.smartView {
			tabs = append(tabs, activeTab.Render(v.name))
		} else {
			tabs = append(tabs, inactiveTab.Render(v.name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	divider := m.theme.Divider.Render(
		lipgloss.NewStyle().Width(m.width).
			Render(repeatString(theme.IconDivider, m.width)),
	)

	return row + "\n" + divider
}

func (m Model) breadcrumb() []string {
	switch m.view {
	case ViewLoading:
		return []string{"Loading"}
	case ViewDashboard:
		return []string{m.smartView.String()}
	case ViewTaskList:
		return []string{"Tasks"}
	case ViewTaskDetail:
		crumbs := []string{"Tasks"}
		if m.selectedTask != nil {
			crumbs = append(crumbs, m.selectedTask.Title)
		}
		return crumbs
	case ViewProjectList:
		return []string{"Projects"}
	default:
		return nil
	}
}

func (m Model) currentBindings() []components.KeyBinding {
	base := []components.KeyBinding{
		{Key: "1-5", Help: "views"},
		{Key: "p", Help: "projects"},
		{Key: "r", Help: "refresh"},
	}

	switch m.view {
	case ViewDashboard, ViewTaskList:
		return append([]components.KeyBinding{
			{Key: "↑↓", Help: "navigate"},
			{Key: "⏎", Help: "details"},
			{Key: "/", Help: "filter"},
			{Key: "x", Help: "complete"},
		}, base...)
	case ViewTaskDetail:
		return append([]components.KeyBinding{
			{Key: "↑↓", Help: "scroll"},
			{Key: "esc", Help: "back"},
		}, base...)
	case ViewProjectList:
		return append([]components.KeyBinding{
			{Key: "↑↓", Help: "navigate"},
			{Key: "⏎", Help: "view tasks"},
			{Key: "/", Help: "filter"},
			{Key: "esc", Help: "back"},
		}, base...)
	}
	return append([]components.KeyBinding{{Key: "q", Help: "quit"}}, base...)
}

// Data loading commands

func (m Model) loadTasksCmd() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		projects, err := client.ListProjects()
		if err != nil {
			return tasksLoadedMsg{err: err}
		}

		type result struct {
			tasks   []types.Task
			project types.Project
			err     error
		}

		ch := make(chan result, len(projects))
		var wg sync.WaitGroup

		for _, p := range projects {
			wg.Add(1)
			go func(proj types.Project) {
				defer wg.Done()
				tasks, err := client.ListTasks(proj.ID)
				ch <- result{tasks: tasks, project: proj, err: err}
			}(p)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		var all []projectTask
		for r := range ch {
			if r.err != nil {
				continue
			}
			for _, t := range r.tasks {
				all = append(all, projectTask{
					Task:        t,
					ProjectName: r.project.Name,
				})
			}
		}

		return tasksLoadedMsg{tasks: all}
	}
}

func (m Model) loadProjectsCmd() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		projects, err := client.ListProjects()
		if err != nil {
			return projectsLoadedMsg{err: err}
		}

		type countResult struct {
			projectID string
			count     int
		}

		ch := make(chan countResult, len(projects))
		var wg sync.WaitGroup

		for _, p := range projects {
			wg.Add(1)
			go func(proj types.Project) {
				defer wg.Done()
				tasks, err := client.ListTasks(proj.ID)
				count := 0
				if err == nil {
					count = len(tasks)
				}
				ch <- countResult{projectID: proj.ID, count: count}
			}(p)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		counts := make(map[string]int)
		for r := range ch {
			counts[r.projectID] = r.count
		}

		var result []projectWithCount
		for _, p := range projects {
			result = append(result, projectWithCount{
				Project:   p,
				TaskCount: counts[p.ID],
			})
		}

		return projectsLoadedMsg{projects: result}
	}
}

// View filter logic

func (m Model) applySmartViewFilter() Model {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	weekEnd := today.AddDate(0, 0, 7)

	var filtered []projectTask

	switch m.smartView {
	case SmartViewToday:
		for _, t := range m.allTasks {
			if t.Status == task.StatusComplete {
				continue
			}
			due := time.Time(t.DueDate)
			if due.IsZero() {
				continue
			}
			// Today or overdue
			if due.Before(tomorrow) {
				filtered = append(filtered, t)
			}
		}
	case SmartViewTomorrow:
		for _, t := range m.allTasks {
			if t.Status == task.StatusComplete {
				continue
			}
			due := time.Time(t.DueDate)
			if due.IsZero() {
				continue
			}
			dayAfter := tomorrow.AddDate(0, 0, 1)
			if !due.Before(tomorrow) && due.Before(dayAfter) {
				filtered = append(filtered, t)
			}
		}
	case SmartViewWeek:
		for _, t := range m.allTasks {
			if t.Status == task.StatusComplete {
				continue
			}
			due := time.Time(t.DueDate)
			if due.IsZero() {
				continue
			}
			if due.Before(weekEnd) {
				filtered = append(filtered, t)
			}
		}
	case SmartViewInbox:
		for _, t := range m.allTasks {
			if t.Status == task.StatusComplete {
				continue
			}
			if t.ProjectID == "inbox" || t.ProjectName == "📥Inbox" {
				filtered = append(filtered, t)
			}
		}
	case SmartViewAll:
		for _, t := range m.allTasks {
			if t.Status == task.StatusComplete {
				continue
			}
			filtered = append(filtered, t)
		}
	}

	m = *m.setTaskListItems(filtered)
	return m
}

func (m *Model) setTaskListItems(tasks []projectTask) *Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = components.TaskItem{
			Task:        t.Task,
			ProjectName: t.ProjectName,
		}
	}
	m.taskList.SetItems(items)
	return m
}

func (m Model) updateProjectList() Model {
	items := make([]list.Item, len(m.projects))
	for i, p := range m.projects {
		items[i] = components.ProjectItem{
			Project:   p.Project,
			TaskCount: p.TaskCount,
		}
	}
	m.projectList.SetItems(items)
	return m
}

func (m Model) updateLayout() Model {
	contentH := m.height - 4 // header + help + statusbar
	if contentH < 1 {
		contentH = 1
	}
	m.taskList.SetSize(m.width, contentH)
	m.projectList.SetSize(m.width, contentH)
	m.detailView.Width = m.width - 2
	m.detailView.Height = contentH
	return m
}

func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
