package api

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	cliErrors "github.com/botre/tickli/internal/errors"
	"github.com/botre/tickli/internal/types"
)

const (
	baseURL     = "https://api.ticktick.com/open/v1"
	authURL     = "https://ticktick.com/oauth/authorize"
	tokenURL    = "https://ticktick.com/oauth/token"
	scope       = "tasks:write tasks:read"
	redirectURL = "http://localhost:8080"
)

// tickTickError represents a TickTick API error response.
type tickTickError struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// apiErrorMessage extracts a user-friendly message from a TickTick API error response.
// Falls back to the raw body if the response cannot be parsed.
func apiErrorMessage(body string) string {
	var apiErr tickTickError
	if err := json.Unmarshal([]byte(body), &apiErr); err == nil && apiErr.ErrorMessage != "" {
		return apiErr.ErrorMessage
	}
	return body
}

type Client struct {
	http *resty.Client
}

func NewClient(token string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", "Bearer "+token)

	return &Client{http: client}
}

func GetAuthURL(clientID string) string {
	return fmt.Sprintf("%s?scope=%s&client_id=%s&state=state&redirect_uri=%s&response_type=code",
		authURL, scope, clientID, redirectURL)
}

func GetAccessToken(clientID, clientSecret, code string) (string, error) {
	client := resty.New()

	resp, err := client.R().
		SetBasicAuth(clientID, clientSecret).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":   "authorization_code",
			"code":         code,
			"redirect_uri": redirectURL,
		}).
		Post(tokenURL)

	if err != nil {
		return "", errors.Wrap(err, "requesting access token")
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", errors.Wrap(err, "parsing response")
	}

	return result.AccessToken, nil
}

func (c *Client) ListProjects() ([]types.Project, error) {
	var projects []types.Project
	resp, err := c.http.R().
		SetResult(&projects).
		Get("/project")

	if err != nil {
		return nil, errors.Wrap(err, "listing projects")
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list projects: %s", apiErrorMessage(resp.String()))
	}

	// Adds the default InboxProject - not appears by default
	projects = append(projects, types.InboxProject)

	return projects, nil
}

// ResolveProject finds a project by ID or name.
// It tries an exact ID lookup first, then falls back to fuzzy name matching
// across all projects (exact → prefix → contains → Levenshtein).
func (c *Client) ResolveProject(nameOrID string) (types.Project, error) {
	// Try direct ID lookup first
	if p, err := c.GetProject(nameOrID); err == nil {
		return p, nil
	}

	// Fall back to name-based matching
	projects, err := c.ListProjects()
	if err != nil {
		return types.NullProject, errors.Wrap(err, "listing projects for name lookup")
	}

	return matchProjectByName(projects, nameOrID)
}

// normalizeName strips emojis and surrounding whitespace, then lowercases
// the result so that "🧼Chores" becomes "chores".
func normalizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || unicode.IsPunct(r) {
			b.WriteRune(r)
		}
	}
	return strings.ToLower(strings.TrimSpace(b.String()))
}

// matchProjectByName finds the best project match by name using progressively
// looser matching: exact → prefix → contains → Levenshtein distance.
// Emojis are stripped from project names before comparison.
func matchProjectByName(projects []types.Project, query string) (types.Project, error) {
	queryNorm := normalizeName(query)

	// 1. Exact case-insensitive match (ignoring emojis)
	for _, p := range projects {
		if normalizeName(p.Name) == queryNorm {
			return p, nil
		}
	}

	// 2. Prefix match
	var prefixMatches []types.Project
	for _, p := range projects {
		if strings.HasPrefix(normalizeName(p.Name), queryNorm) {
			prefixMatches = append(prefixMatches, p)
		}
	}
	if len(prefixMatches) == 1 {
		return prefixMatches[0], nil
	}
	if len(prefixMatches) > 1 {
		return types.NullProject, ambiguousMatchError(query, prefixMatches)
	}

	// 3. Contains match
	var containsMatches []types.Project
	for _, p := range projects {
		if strings.Contains(normalizeName(p.Name), queryNorm) {
			containsMatches = append(containsMatches, p)
		}
	}
	if len(containsMatches) == 1 {
		return containsMatches[0], nil
	}
	if len(containsMatches) > 1 {
		return types.NullProject, ambiguousMatchError(query, containsMatches)
	}

	// 4. Levenshtein distance (typo tolerance)
	maxDist := len(query) / 3
	if maxDist < 1 {
		maxDist = 1
	}

	type scored struct {
		project  types.Project
		distance int
	}
	var fuzzyMatches []scored
	for _, p := range projects {
		d := levenshtein(normalizeName(p.Name), queryNorm)
		if d <= maxDist {
			fuzzyMatches = append(fuzzyMatches, scored{p, d})
		}
	}
	if len(fuzzyMatches) == 1 {
		return fuzzyMatches[0].project, nil
	}
	if len(fuzzyMatches) > 1 {
		// Pick the closest match if it's unambiguous
		best := fuzzyMatches[0]
		ambiguous := false
		for _, m := range fuzzyMatches[1:] {
			if m.distance < best.distance {
				best = m
				ambiguous = false
			} else if m.distance == best.distance {
				ambiguous = true
			}
		}
		if !ambiguous {
			return best.project, nil
		}
		matched := make([]types.Project, len(fuzzyMatches))
		for i, m := range fuzzyMatches {
			matched[i] = m.project
		}
		return types.NullProject, ambiguousMatchError(query, matched)
	}

	return types.NullProject, fmt.Errorf("no project found matching %q. Run 'tickli project list' to see available projects", query)
}

func ambiguousMatchError(query string, matches []types.Project) error {
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = fmt.Sprintf("  %s (id: %s)", m.Name, m.ID)
	}
	return fmt.Errorf("multiple projects match %q:\n%s\nUse a more specific name or pass the project ID", query, strings.Join(names, "\n"))
}

// levenshtein computes the Levenshtein edit distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func (c *Client) GetProject(id string) (types.Project, error) {
	if id == types.InboxProject.ID {
		return types.InboxProject, nil
	}
	var project types.Project
	resp, err := c.http.R().
		SetResult(&project).
		Get("/project/" + id)

	if err != nil {
		return types.NullProject, errors.Wrap(err, "getting project")
	}
	if resp.IsError() {
		return types.NullProject, fmt.Errorf("failed to get project: %s", apiErrorMessage(resp.String()))
	}
	if project == types.NullProject {
		return types.NullProject, &cliErrors.NotFoundError{Message: fmt.Sprintf("project not found: %s", id)}
	}

	return project, nil
}


func (c *Client) getTaskFromProject(projectID, taskID string) (*types.Task, error) {
	var task types.Task
	resp, err := c.http.R().
		SetResult(&task).
		Get(fmt.Sprintf("/project/%s/task/%s", projectID, taskID))

	if err != nil {
		return nil, errors.Wrap(err, "requesting task")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get task: %s", apiErrorMessage(resp.String()))
	}
	if task.ID == "" || task.Title == "" {
		return nil, &cliErrors.NotFoundError{Message: fmt.Sprintf("task %s not found in project %s", taskID, projectID)}
	}

	return &task, nil
}

// GetTask fetches a task by ID, searching across all projects if needed.
func (c *Client) GetTask(taskID string) (*types.Task, error) {
	// Try inbox first (most common case)
	if t, err := c.getTaskFromProject(types.InboxProject.ID, taskID); err == nil {
		return t, nil
	}

	// Search all projects
	projects, err := c.ListProjects()
	if err != nil {
		return nil, errors.Wrap(err, "listing projects for task lookup")
	}

	for _, p := range projects {
		if p.ID == types.InboxProject.ID {
			continue // already tried
		}
		if t, err := c.getTaskFromProject(p.ID, taskID); err == nil {
			return t, nil
		}
	}

	return nil, &cliErrors.NotFoundError{Message: fmt.Sprintf("task %s not found in any project", taskID)}
}

func (c *Client) ListTasks(projectID string) ([]types.Task, error) {
	var projectData struct {
		Tasks []types.Task `json:"tasks"`
	}
	resp, err := c.http.R().
		SetResult(&projectData).
		Get(fmt.Sprintf("/project/%s/data", projectID))

	if err != nil {
		return nil, errors.Wrap(err, "listing tasks")
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list tasks: %s", apiErrorMessage(resp.String()))
	}

	sortTasksByDueDate(projectData.Tasks)
	return projectData.Tasks, nil
}

func (c *Client) GetProjectWithTasks(projectID string) (*types.ProjectData, error) {
	var projectData types.ProjectData
	resp, err := c.http.R().
		SetResult(&projectData).
		Get(fmt.Sprintf("/project/%s/data", projectID))

	if err != nil {
		return nil, errors.Wrap(err, "getting project data")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to get project data: %s", apiErrorMessage(resp.String()))
	}

	// The API doesn't return project metadata for inbox, fill it in
	if projectData.Project.ID == "" && projectID == types.InboxProject.ID {
		projectData.Project = types.InboxProject
	}

	sortTasksByDueDate(projectData.Tasks)
	return &projectData, nil
}

// sortTasksByDueDate sorts tasks by due date ascending, with no-date tasks at the end.
func sortTasksByDueDate(tasks []types.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		di := time.Time(tasks[i].DueDate)
		dj := time.Time(tasks[j].DueDate)
		if di.IsZero() && dj.IsZero() {
			return false
		}
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.Before(dj)
	})
}

// TaskFilter specifies criteria for the filter tasks endpoint.
type TaskFilter struct {
	ProjectIDs []string `json:"projectIds,omitempty"`
	StartDate  string   `json:"startDate,omitempty"`
	EndDate    string   `json:"endDate,omitempty"`
	Priority   []int    `json:"priority,omitempty"`
	Tag        []string `json:"tag,omitempty"`
	Status     []int    `json:"status,omitempty"`
}

// FilterTasks retrieves tasks using the filter endpoint (single API call).
func (c *Client) FilterTasks(filter TaskFilter) ([]types.Task, error) {
	var tasks []types.Task
	resp, err := c.http.R().
		SetBody(filter).
		SetResult(&tasks).
		Post("/task/filter")

	if err != nil {
		return nil, errors.Wrap(err, "filtering tasks")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to filter tasks: %s", apiErrorMessage(resp.String()))
	}
	if tasks == nil {
		tasks = []types.Task{}
	}

	sortTasksByDueDate(tasks)
	return tasks, nil
}

func (c *Client) CreateTask(task *types.Task) (*types.Task, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	resp, err := c.http.R().
		SetBody(task).
		SetResult(task).
		Post("/task")

	if err != nil {
		return nil, errors.Wrap(err, "creating task")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create task: %s", apiErrorMessage(resp.String()))
	}

	return task, nil
}

func (c *Client) UpdateTask(task *types.Task) (*types.Task, error) {
	if task == nil {
		return nil, errors.New("task cannot be nil")
	}

	resp, err := c.http.R().
		SetBody(task).
		SetResult(task).
		Post(fmt.Sprintf("/task/%s", task.ID))

	if err != nil {
		return nil, errors.Wrap(err, "updating task")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to update task: %s", apiErrorMessage(resp.String()))
	}

	return task, nil
}

// resolveInboxID returns the real inbox project ID.
// The TickTick API uses "inbox" as an alias for some endpoints but not all.
// This method discovers the real ID by creating and deleting a temporary task.
func (c *Client) resolveInboxID() (string, error) {
	tmp := &types.Task{
		ProjectID: types.InboxProject.ID,
		Title:     "_tickli_inbox_probe",
	}
	created, err := c.CreateTask(tmp)
	if err != nil {
		return "", errors.Wrap(err, "creating inbox probe task")
	}
	realID := created.ProjectID
	// Clean up
	_ = c.DeleteTask(created.ID)
	return realID, nil
}

// MoveTask moves a task to a different project using the dedicated move endpoint.
func (c *Client) MoveTask(taskID, fromProjectID, toProjectID string) error {
	// The move endpoint doesn't accept "inbox" as a project ID alias,
	// so we resolve it to the real ID when needed.
	if fromProjectID == types.InboxProject.ID || toProjectID == types.InboxProject.ID {
		realInboxID, err := c.resolveInboxID()
		if err != nil {
			return errors.Wrap(err, "resolving inbox ID")
		}
		if fromProjectID == types.InboxProject.ID {
			fromProjectID = realInboxID
		}
		if toProjectID == types.InboxProject.ID {
			toProjectID = realInboxID
		}
	}

	type moveItem struct {
		TaskID        string `json:"taskId"`
		FromProjectID string `json:"fromProjectId"`
		ToProjectID   string `json:"toProjectId"`
	}

	resp, err := c.http.R().
		SetBody([]moveItem{{
			TaskID:        taskID,
			FromProjectID: fromProjectID,
			ToProjectID:   toProjectID,
		}}).
		Post("/task/move")

	if err != nil {
		return errors.Wrap(err, "moving task")
	}
	if resp.IsError() {
		return fmt.Errorf("failed to move task (status %d): %s", resp.StatusCode(), apiErrorMessage(resp.String()))
	}

	return nil
}

func (c *Client) UpdateProject(project types.Project) (types.Project, error) {
	resp, err := c.http.R().
		SetBody(project).
		SetResult(project).
		Post(fmt.Sprintf("/project/%s", project.ID))

	if err != nil {
		return types.NullProject, errors.Wrap(err, "updating project")
	}
	if resp.IsError() {
		return types.NullProject, fmt.Errorf("failed to update project: %s", apiErrorMessage(resp.String()))
	}

	return project, nil
}

func (c *Client) DeleteTask(taskID string) error {
	task, err := c.GetTask(taskID)
	if err != nil {
		return errors.Wrap(err, "resolving task for deletion")
	}

	resp, err := c.http.R().
		Delete(fmt.Sprintf("/project/%s/task/%s", task.ProjectID, taskID))

	if err != nil {
		return errors.Wrap(err, "deleting task")
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete task: %s", apiErrorMessage(resp.String()))
	}

	return nil
}

func (c *Client) CompleteTask(taskID string) error {
	task, err := c.GetTask(taskID)
	if err != nil {
		return errors.Wrap(err, "resolving task for completion")
	}

	resp, err := c.http.R().
		Post(fmt.Sprintf("/project/%s/task/%s/complete", task.ProjectID, taskID))

	if err != nil {
		return errors.Wrap(err, "completing task")
	}
	if resp.IsError() {
		return fmt.Errorf("failed to complete task: %s", apiErrorMessage(resp.String()))
	}

	return nil
}

func (c *Client) CreateProject(project *types.Project) (*types.Project, error) {
	if project == nil {
		return nil, errors.New("project cannot be nil")
	}

	resp, err := c.http.R().
		SetBody(project).
		SetResult(project).
		Post("/project")

	if err != nil {
		return nil, errors.Wrap(err, "creating project")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create project: %s", apiErrorMessage(resp.String()))
	}

	return project, nil
}

func (c *Client) DeleteProject(projectID string) error {
	resp, err := c.http.R().
		Delete(fmt.Sprintf("/project/%s", projectID))

	if err != nil {
		return errors.Wrap(err, "deleting project")
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete project: %s", apiErrorMessage(resp.String()))
	}

	return nil
}
