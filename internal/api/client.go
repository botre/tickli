package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/botre/tickli/internal/types"
)

const (
	baseURL     = "https://api.ticktick.com/open/v1"
	authURL     = "https://ticktick.com/oauth/authorize"
	tokenURL    = "https://ticktick.com/oauth/token"
	scope       = "tasks:write tasks:read"
	redirectURL = "http://localhost:8080"
)

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
		return nil, fmt.Errorf("failed to list projects: %s", resp.String())
	}

	// Adds the default InboxProject - not appears by default
	projects = append(projects, types.InboxProject)

	return projects, nil
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
		return types.NullProject, fmt.Errorf("failed to get project: %s", resp.String())
	}
	if project == types.NullProject {
		return types.NullProject, fmt.Errorf("project not found: %s", id)
	}

	return project, nil
}

// ResolveProject finds a project by ID first, then by name if no ID match is found.
func (c *Client) ResolveProject(idOrName string) (types.Project, error) {
	// Try by ID first
	p, err := c.GetProject(idOrName)
	if err == nil {
		return p, nil
	}

	// Try by name
	projects, listErr := c.ListProjects()
	if listErr != nil {
		return types.NullProject, errors.Wrap(listErr, "listing projects for name lookup")
	}

	for _, proj := range projects {
		if strings.EqualFold(proj.Name, idOrName) {
			return proj, nil
		}
	}

	return types.NullProject, fmt.Errorf("project %q not found by ID or name", idOrName)
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
		return nil, fmt.Errorf("failed to get task: %s", resp.String())
	}
	if task.ID == "" || task.Title == "" {
		return nil, fmt.Errorf("task %s not found in project %s", taskID, projectID)
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

	return nil, fmt.Errorf("task %s not found in any project", taskID)
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
		return nil, fmt.Errorf("failed to list tasks: %s", resp.String())
	}

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
		return nil, fmt.Errorf("failed to get project data: %s", resp.String())
	}

	// The API doesn't return project metadata for inbox, fill it in
	if projectData.Project.ID == "" && projectID == types.InboxProject.ID {
		projectData.Project = types.InboxProject
	}

	return &projectData, nil
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
		return nil, fmt.Errorf("failed to create task: %s", resp.String())
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
		return nil, fmt.Errorf("failed to update task: %s", resp.String())
	}

	return task, nil
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
		return types.NullProject, fmt.Errorf("failed to update project: %s", resp.String())
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
		return fmt.Errorf("failed to delete task: %s", resp.String())
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
		return fmt.Errorf("failed to complete task: %s", resp.String())
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
		return nil, fmt.Errorf("failed to create project: %s", resp.String())
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
		return fmt.Errorf("failed to delete project: %s", resp.String())
	}

	return nil
}
