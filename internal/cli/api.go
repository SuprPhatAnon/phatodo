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
