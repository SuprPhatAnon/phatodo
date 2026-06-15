package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
)

type APIClient struct {
	baseURL      *url.URL
	accessKey    string
	accessSecret string
	httpClient   *http.Client
}

type apiClient interface {
	ListProjectConfig(context.Context, string) ([]ProjectConfigItem, error)
	GetProjectConfig(context.Context, string, string) (ProjectConfigItem, error)
	SetProjectConfig(context.Context, string, string, string) (ProjectConfigItem, error)
	UnsetProjectConfig(context.Context, string, string) (ProjectConfigItem, error)
	CreateTask(context.Context, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error)
	CreateSubtask(context.Context, string, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error)
	GetTask(context.Context, string, string) (domain.TaskDetail, error)
	UpdateTask(context.Context, string, string, domain.TaskUpdateRequest) (domain.TaskDetail, error)
	DeleteTask(context.Context, string, string) (domain.TaskDetail, error)
	ListTasks(context.Context, string, string, string) (domain.TaskListResponse, error)
	ListSubtasks(context.Context, string, string) (domain.TaskListResponse, error)
	ListComments(context.Context, string, string) (domain.CommentListResponse, error)
	AddComment(context.Context, string, string, domain.CommentCreateRequest) (domain.Comment, error)
	UpdateComment(context.Context, string, string, domain.CommentUpdateRequest) (domain.Comment, error)
	DeleteComment(context.Context, string, string) (domain.Comment, error)
	ListDependencies(context.Context, string, string) (domain.DependencyListResponse, error)
	AddDependency(context.Context, string, string, string) (domain.Dependency, error)
	RemoveDependency(context.Context, string, string, string) (domain.Dependency, error)
	Search(context.Context, string, string, string, string, int) (domain.SearchResponse, error)
	History(context.Context, string, string, string, string, string, int) (domain.HistoryResponse, error)
	ListUnified(context.Context, string, string, string, string, string, int) (domain.ListResponse, error)
	ListReadyTasks(context.Context, string, string) (domain.ReadyListResponse, error)
	InitAdmin(context.Context, domain.AdminInitRequest) (domain.AdminInitResponse, error)
	BootstrapAdmin(context.Context, domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error)
}

type ProjectConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ProjectConfigListResponse struct {
	ProjectID string              `json:"project_id"`
	Items     []ProjectConfigItem `json:"items"`
}

type adminInitResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

type adminBootstrapResponse struct {
	WorkspaceID  string `json:"workspace_id"`
	ProjectID    string `json:"project_id"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

func NewAPIClient(cfg config.LocalConfig) (*APIClient, error) {
	baseURL, err := url.Parse(cfg.APIURL)
	if err != nil {
		return nil, fmt.Errorf("parse api url: %w", err)
	}

	return &APIClient{
		baseURL:      baseURL,
		accessKey:    cfg.AccessKey,
		accessSecret: cfg.AccessSecret,
		httpClient:   &http.Client{},
	}, nil
}

var newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
	return NewAPIClient(cfg)
}

func (c *APIClient) ListProjectConfig(ctx context.Context, projectID string) ([]ProjectConfigItem, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "config")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Phatodo-Access-Key", c.accessKey)
	req.Header.Set("X-Phatodo-Access-Secret", c.accessSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("list project config failed: %s", resp.Status)
	}

	var payload ProjectConfigListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode config list response: %w", err)
	}

	return payload.Items, nil
}

func (c *APIClient) GetProjectConfig(ctx context.Context, projectID string, key string) (ProjectConfigItem, error) {
	var payload ProjectConfigItem
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/config/%s", url.PathEscape(projectID), url.PathEscape(key)), nil, &payload); err != nil {
		return ProjectConfigItem{}, err
	}
	return payload, nil
}

func (c *APIClient) SetProjectConfig(ctx context.Context, projectID string, key string, value string) (ProjectConfigItem, error) {
	var payload ProjectConfigItem
	if err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/projects/%s/config/%s", url.PathEscape(projectID), url.PathEscape(key)), domain.ProjectConfigSetRequest{
		Value: value,
	}, &payload); err != nil {
		return ProjectConfigItem{}, err
	}
	return payload, nil
}

func (c *APIClient) UnsetProjectConfig(ctx context.Context, projectID string, key string) (ProjectConfigItem, error) {
	var payload ProjectConfigItem
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%s/config/%s", url.PathEscape(projectID), url.PathEscape(key)), nil, &payload); err != nil {
		return ProjectConfigItem{}, err
	}
	return payload, nil
}

func (c *APIClient) CreateTask(ctx context.Context, projectID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
	var payload domain.TaskCreateResponse
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID)), req, &payload); err != nil {
		return domain.TaskCreateResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) CreateSubtask(ctx context.Context, projectID string, taskID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
	var payload domain.TaskCreateResponse
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/subtasks", url.PathEscape(projectID), url.PathEscape(taskID)), req, &payload); err != nil {
		return domain.TaskCreateResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) GetTask(ctx context.Context, projectID string, taskID string) (domain.TaskDetail, error) {
	var payload domain.TaskDetail
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/tasks/%s", url.PathEscape(projectID), url.PathEscape(taskID)), nil, &payload); err != nil {
		return domain.TaskDetail{}, err
	}
	return payload, nil
}

func (c *APIClient) UpdateTask(ctx context.Context, projectID string, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
	var payload domain.TaskDetail
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/projects/%s/tasks/%s", url.PathEscape(projectID), url.PathEscape(taskID)), req, &payload); err != nil {
		return domain.TaskDetail{}, err
	}
	return payload, nil
}

func (c *APIClient) DeleteTask(ctx context.Context, projectID string, taskID string) (domain.TaskDetail, error) {
	var payload domain.TaskDetail
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%s/tasks/%s", url.PathEscape(projectID), url.PathEscape(taskID)), nil, &payload); err != nil {
		return domain.TaskDetail{}, err
	}
	return payload, nil
}

func (c *APIClient) ListTasks(ctx context.Context, projectID string, status string, epicID string) (domain.TaskListResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "tasks")
	query := endpoint.Query()
	if status != "" {
		query.Set("status", status)
	}
	if epicID != "" {
		query.Set("epic", epicID)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return domain.TaskListResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Phatodo-Access-Key", c.accessKey)
	req.Header.Set("X-Phatodo-Access-Secret", c.accessSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.TaskListResponse{}, fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		message := resp.Status
		if msg, ok := body["message"].(string); ok && msg != "" {
			message = msg
		}
		if code, ok := body["error"].(string); ok && code != "" {
			return domain.TaskListResponse{}, fmt.Errorf("%s: %s", code, message)
		}
		return domain.TaskListResponse{}, fmt.Errorf("list tasks failed: %s", message)
	}

	var payload domain.TaskListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.TaskListResponse{}, fmt.Errorf("decode task list response: %w", err)
	}

	return payload, nil
}

func (c *APIClient) ListSubtasks(ctx context.Context, projectID string, taskID string) (domain.TaskListResponse, error) {
	var payload domain.TaskListResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/subtasks", url.PathEscape(projectID), url.PathEscape(taskID)), nil, &payload); err != nil {
		return domain.TaskListResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) ListComments(ctx context.Context, projectID string, taskID string) (domain.CommentListResponse, error) {
	var payload domain.CommentListResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/comments", url.PathEscape(projectID), url.PathEscape(taskID)), nil, &payload); err != nil {
		return domain.CommentListResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) AddComment(ctx context.Context, projectID string, taskID string, req domain.CommentCreateRequest) (domain.Comment, error) {
	var payload domain.Comment
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/comments", url.PathEscape(projectID), url.PathEscape(taskID)), req, &payload); err != nil {
		return domain.Comment{}, err
	}
	return payload, nil
}

func (c *APIClient) UpdateComment(ctx context.Context, projectID string, commentID string, req domain.CommentUpdateRequest) (domain.Comment, error) {
	var payload domain.Comment
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/projects/%s/comments/%s", url.PathEscape(projectID), url.PathEscape(commentID)), req, &payload); err != nil {
		return domain.Comment{}, err
	}
	return payload, nil
}

func (c *APIClient) DeleteComment(ctx context.Context, projectID string, commentID string) (domain.Comment, error) {
	var payload domain.Comment
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%s/comments/%s", url.PathEscape(projectID), url.PathEscape(commentID)), nil, &payload); err != nil {
		return domain.Comment{}, err
	}
	return payload, nil
}

func (c *APIClient) ListDependencies(ctx context.Context, projectID string, taskID string) (domain.DependencyListResponse, error) {
	var payload domain.DependencyListResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/dependencies", url.PathEscape(projectID), url.PathEscape(taskID)), nil, &payload); err != nil {
		return domain.DependencyListResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) AddDependency(ctx context.Context, projectID string, taskID string, dependsOnID string) (domain.Dependency, error) {
	var payload domain.Dependency
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/dependencies", url.PathEscape(projectID), url.PathEscape(taskID)), map[string]string{"depends_on_id": dependsOnID}, &payload); err != nil {
		return domain.Dependency{}, err
	}
	return payload, nil
}

func (c *APIClient) RemoveDependency(ctx context.Context, projectID string, taskID string, dependsOnID string) (domain.Dependency, error) {
	var payload domain.Dependency
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%s/tasks/%s/dependencies/%s", url.PathEscape(projectID), url.PathEscape(taskID), url.PathEscape(dependsOnID)), nil, &payload); err != nil {
		return domain.Dependency{}, err
	}
	return payload, nil
}

func (c *APIClient) Search(ctx context.Context, projectID string, query string, entityType string, status string, limit int) (domain.SearchResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "search")
	q := endpoint.Query()
	q.Set("q", query)
	if entityType != "" {
		q.Set("type", entityType)
	}
	if status != "" {
		q.Set("status", status)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	endpoint.RawQuery = q.Encode()

	var payload domain.SearchResponse
	if err := c.doGETJSON(ctx, endpoint.String(), &payload); err != nil {
		return domain.SearchResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) History(ctx context.Context, projectID string, entityID string, entityType string, action string, since string, limit int) (domain.HistoryResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "history")
	q := endpoint.Query()
	if entityID != "" {
		q.Set("entity", entityID)
	}
	if entityType != "" {
		q.Set("type", entityType)
	}
	if action != "" {
		q.Set("action", action)
	}
	if since != "" {
		q.Set("since", since)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	endpoint.RawQuery = q.Encode()

	var payload domain.HistoryResponse
	if err := c.doGETJSON(ctx, endpoint.String(), &payload); err != nil {
		return domain.HistoryResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) ListUnified(ctx context.Context, projectID string, entityType string, status string, priority string, sort string, limit int) (domain.ListResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "list")
	q := endpoint.Query()
	if entityType != "" {
		q.Set("type", entityType)
	}
	if status != "" {
		q.Set("status", status)
	}
	if priority != "" {
		q.Set("priority", priority)
	}
	if sort != "" {
		q.Set("sort", sort)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	endpoint.RawQuery = q.Encode()

	var payload domain.ListResponse
	if err := c.doGETJSON(ctx, endpoint.String(), &payload); err != nil {
		return domain.ListResponse{}, err
	}
	return payload, nil
}

func (c *APIClient) ListReadyTasks(ctx context.Context, projectID string, epicID string) (domain.ReadyListResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), "/api/v1/projects", url.PathEscape(projectID), "ready")
	query := endpoint.Query()
	if epicID != "" {
		query.Set("epic", epicID)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Phatodo-Access-Key", c.accessKey)
	req.Header.Set("X-Phatodo-Access-Secret", c.accessSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		message := resp.Status
		if msg, ok := body["message"].(string); ok && msg != "" {
			message = msg
		}
		if code, ok := body["error"].(string); ok && code != "" {
			return domain.ReadyListResponse{}, fmt.Errorf("%s: %s", code, message)
		}
		return domain.ReadyListResponse{}, fmt.Errorf("ready list failed: %s", message)
	}

	var payload domain.ReadyListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("decode ready response: %w", err)
	}

	return payload, nil
}

func (c *APIClient) doGETJSON(ctx context.Context, urlString string, responseBody any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Phatodo-Access-Key", c.accessKey)
	req.Header.Set("X-Phatodo-Access-Secret", c.accessSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		message := resp.Status
		if msg, ok := body["message"].(string); ok && msg != "" {
			message = msg
		}
		if code, ok := body["error"].(string); ok && code != "" {
			return fmt.Errorf("%s: %s", code, message)
		}
		return fmt.Errorf("api request failed: %s", message)
	}

	if responseBody == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}
	return nil
}

func (c *APIClient) InitAdmin(ctx context.Context, req domain.AdminInitRequest) (domain.AdminInitResponse, error) {
	var payload adminInitResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/admin/init", req, &payload); err != nil {
		return domain.AdminInitResponse{}, err
	}
	return domain.AdminInitResponse{
		UserID:       payload.UserID,
		Username:     payload.Username,
		AccessKey:    payload.AccessKey,
		AccessSecret: payload.AccessSecret,
	}, nil
}

func (c *APIClient) BootstrapAdmin(ctx context.Context, req domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
	var payload adminBootstrapResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/admin/bootstrap", req, &payload); err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	return domain.AdminBootstrapResponse{
		WorkspaceID:  payload.WorkspaceID,
		ProjectID:    payload.ProjectID,
		AccessKey:    payload.AccessKey,
		AccessSecret: payload.AccessSecret,
	}, nil
}

func (c *APIClient) doJSON(ctx context.Context, method string, requestPath string, requestBody any, responseBody any) error {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(strings.TrimRight(c.baseURL.Path, "/"), requestPath)

	var body io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.accessKey != "" {
		req.Header.Set("X-Phatodo-Access-Key", c.accessKey)
	}
	if c.accessSecret != "" {
		req.Header.Set("X-Phatodo-Access-Secret", c.accessSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		message := resp.Status
		if msg, ok := body["message"].(string); ok && msg != "" {
			message = msg
		}
		if code, ok := body["error"].(string); ok && code != "" {
			return fmt.Errorf("%s: %s", code, message)
		}
		return fmt.Errorf("api request failed: %s", message)
	}

	if responseBody == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}
	return nil
}
