package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestHealthIsPublic(t *testing.T) {
	handler := newApp(Config{PostgresDSN: "postgres://example"}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIRoutesRequireCredentials(t *testing.T) {
	handler := newApp(Config{}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAPIRouteScaffoldReturnsAction(t *testing.T) {
	handler := newApp(Config{}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/epics", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "epic.list" {
		t.Fatalf("expected epic.list action, got %#v", body["action"])
	}
	if body["project_id"] != "project-1" {
		t.Fatalf("expected project id, got %#v", body["project_id"])
	}
}

type fakeProjectConfigReader struct {
	items []domain.ProjectConfig
	item  domain.ProjectConfig
	err   error
}

func (f fakeProjectConfigReader) ListProjectConfig(_ context.Context, _ string) ([]domain.ProjectConfig, error) {
	return f.items, f.err
}

func (f fakeProjectConfigReader) GetProjectConfig(_ context.Context, _ string, _ string) (domain.ProjectConfig, error) {
	return f.item, f.err
}

type fakeProjectConfigWriter struct {
	item domain.ProjectConfig
	err  error
}

func (f fakeProjectConfigWriter) SetProjectConfig(_ context.Context, _ string, _ string, _ string) (domain.ProjectConfig, error) {
	return f.item, f.err
}

func (f fakeProjectConfigWriter) DeleteProjectConfig(_ context.Context, _ string, _ string) (domain.ProjectConfig, error) {
	return f.item, f.err
}

func TestProjectConfigListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigReader: fakeProjectConfigReader{
			items: []domain.ProjectConfig{{Key: "theme", Value: "dark"}},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/config", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "project-1", body["project_id"])

	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "theme", item["key"])
	require.Equal(t, "dark", item["value"])
}

func TestProjectConfigSetReturnsItem(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigWriter: fakeProjectConfigWriter{
			item: domain.ProjectConfig{Key: "theme", Value: "dark"},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/project-1/config/theme", strings.NewReader(`{"value":"dark"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "theme", body["key"])
	require.Equal(t, "dark", body["value"])
}

func TestProjectConfigGetReturnsItem(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigReader: fakeProjectConfigReader{
			item: domain.ProjectConfig{Key: "theme", Value: "dark"},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/config/theme", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "theme", body["key"])
	require.Equal(t, "dark", body["value"])
}

func TestProjectConfigUnsetReturnsItem(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigWriter: fakeProjectConfigWriter{
			item: domain.ProjectConfig{Key: "theme", Value: "dark"},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1/config/theme", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "theme", body["key"])
	require.Equal(t, "dark", body["value"])
}

type fakeBootstrapManager struct {
	initResponse      domain.AdminInitResponse
	bootstrapResponse domain.AdminBootstrapResponse
	initErr           error
	bootstrapErr      error
}

func (f fakeBootstrapManager) InitAdmin(_ context.Context, _ domain.AdminInitRequest) (domain.AdminInitResponse, error) {
	return f.initResponse, f.initErr
}

func (f fakeBootstrapManager) BootstrapProject(_ context.Context, _ domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
	return f.bootstrapResponse, f.bootstrapErr
}

type fakeTaskCreator struct {
	createResponse domain.TaskCreateResponse
	createErr      error
}

func (f fakeTaskCreator) CreateTask(_ context.Context, _ string, _ domain.TaskCreateRequest, _ string) (domain.TaskCreateResponse, error) {
	return f.createResponse, f.createErr
}

type fakeSubtaskLister struct {
	listResponse domain.TaskListResponse
	listErr      error
}

func (f fakeSubtaskLister) ListSubtasks(_ context.Context, _ string, _ string) (domain.TaskListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeTaskLister struct {
	listResponse domain.TaskListResponse
	listErr      error
}

func (f fakeTaskLister) ListTasks(_ context.Context, _ string, _ string, _ string) (domain.TaskListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeTaskReader struct {
	getResponse domain.TaskDetail
	getErr      error
}

func (f fakeTaskReader) GetTask(_ context.Context, _ string, _ string) (domain.TaskDetail, error) {
	return f.getResponse, f.getErr
}

type fakeTaskUpdater struct {
	updateResponse domain.TaskDetail
	updateErr      error
}

func (f fakeTaskUpdater) UpdateTask(_ context.Context, _ string, _ string, _ domain.TaskUpdateRequest, _ string) (domain.TaskDetail, error) {
	return f.updateResponse, f.updateErr
}

type fakeTaskDeleter struct {
	deleteResponse domain.TaskDetail
	deleteErr      error
}

func (f fakeTaskDeleter) DeleteTask(_ context.Context, _ string, _ string, _ string) (domain.TaskDetail, error) {
	return f.deleteResponse, f.deleteErr
}

type fakeCommentLister struct {
	listResponse domain.CommentListResponse
	listErr      error
}

func (f fakeCommentLister) ListComments(_ context.Context, _ string, _ string) (domain.CommentListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeCommentCreator struct {
	createResponse domain.Comment
	createErr      error
}

func (f fakeCommentCreator) CreateComment(_ context.Context, _ string, _ string, _ domain.CommentCreateRequest, _ string) (domain.Comment, error) {
	return f.createResponse, f.createErr
}

type fakeCommentUpdater struct {
	updateResponse domain.Comment
	updateErr      error
}

func (f fakeCommentUpdater) UpdateComment(_ context.Context, _ string, _ string, _ domain.CommentUpdateRequest, _ string) (domain.Comment, error) {
	return f.updateResponse, f.updateErr
}

type fakeCommentDeleter struct {
	deleteResponse domain.Comment
	deleteErr      error
}

func (f fakeCommentDeleter) DeleteComment(_ context.Context, _ string, _ string, _ string) (domain.Comment, error) {
	return f.deleteResponse, f.deleteErr
}

type fakeReadyLister struct {
	listResponse domain.ReadyListResponse
	listErr      error
}

func (f fakeReadyLister) ListReadyTasks(_ context.Context, _ string, _ string) (domain.ReadyListResponse, error) {
	return f.listResponse, f.listErr
}

func TestAdminInitReturnsCreatedAdmin(t *testing.T) {
	handler := newApp(Config{
		BootstrapManager: fakeBootstrapManager{
			initResponse: domain.AdminInitResponse{
				UserID:       "usr_1",
				Username:     "alice",
				AccessKey:    "key_1",
				AccessSecret: "sec_1",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/init", strings.NewReader(`{"username":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "alice", body["username"])
	require.Equal(t, "key_1", body["access_key"])
	require.Equal(t, "sec_1", body["access_secret"])
}

func TestAdminBootstrapReturnsProjectConfig(t *testing.T) {
	handler := newApp(Config{
		BootstrapManager: fakeBootstrapManager{
			bootstrapResponse: domain.AdminBootstrapResponse{
				WorkspaceID:  "ws_1",
				ProjectID:    "prj_1",
				AccessKey:    "key_1",
				AccessSecret: "sec_1",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/bootstrap", strings.NewReader(`{"username":"alice","password":"secret","workspace_name":"phatodo","project_name":"phatodo"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ws_1", body["workspace_id"])
	require.Equal(t, "prj_1", body["project_id"])
	require.Equal(t, "key_1", body["access_key"])
}

func TestTaskShowReturnsDetail(t *testing.T) {
	handler := newApp(Config{
		TaskReader: fakeTaskReader{
			getResponse: domain.TaskDetail{
				ID:       "ABC-1",
				Title:    "Write docs",
				Status:   domain.StatusTodo,
				Priority: domain.PriorityMedium,
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks/ABC-1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ABC-1", body["id"])
	require.Equal(t, "Write docs", body["title"])
}

func TestTaskUpdateReturnsDetail(t *testing.T) {
	handler := newApp(Config{
		TaskUpdater: fakeTaskUpdater{
			updateResponse: domain.TaskDetail{
				ID:       "ABC-1",
				Title:    "Updated docs",
				Status:   domain.StatusInProgress,
				Priority: domain.PriorityHigh,
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/projects/project-1/tasks/ABC-1", strings.NewReader(`{"title":"Updated docs","status":"in_progress"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "Updated docs", body["title"])
	require.Equal(t, "in_progress", body["status"])
}

func TestTaskDeleteReturnsDetail(t *testing.T) {
	handler := newApp(Config{
		TaskDeleter: fakeTaskDeleter{
			deleteResponse: domain.TaskDetail{
				ID:     "ABC-1",
				Title:  "Write docs",
				Status: domain.StatusArchived,
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1/tasks/ABC-1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ABC-1", body["id"])
	require.Equal(t, "archived", body["status"])
}

func TestTaskCreateReturnsTask(t *testing.T) {
	handler := newApp(Config{
		TaskCreator: fakeTaskCreator{
			createResponse: domain.TaskCreateResponse{
				ID:          "ABC-1",
				IssuePrefix: "ABC",
				Title:       "Write docs",
				Status:      domain.StatusTodo,
				Priority:    domain.PriorityMedium,
				ProjectID:   "project-1",
				WorkspaceID: "workspace-1",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/tasks", strings.NewReader(`{"title":"Write docs","issue_prefix":"ABC","priority":2,"tags":["docs"]}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ABC-1", body["id"])
	require.Equal(t, "ABC", body["issue_prefix"])
	require.Equal(t, "Write docs", body["title"])
}

func TestSubtaskCreateReturnsTask(t *testing.T) {
	handler := newApp(Config{
		TaskCreator: fakeTaskCreator{
			createResponse: domain.TaskCreateResponse{
				ID:          "ABC-2",
				IssuePrefix: "ABC",
				Title:       "Write docs",
				Status:      domain.StatusTodo,
				Priority:    domain.PriorityMedium,
				ProjectID:   "project-1",
				WorkspaceID: "workspace-1",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/tasks/ABC-1/subtasks", strings.NewReader(`{"title":"Write docs","description":"child"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ABC-2", body["id"])
	require.Equal(t, "ABC", body["issue_prefix"])
	require.Equal(t, "Write docs", body["title"])
}

func TestTaskListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		TaskLister: fakeTaskLister{
			listResponse: domain.TaskListResponse{
				ProjectID: "project-1",
				Items: []domain.TaskListItem{
					{
						ID:       "ABC-1",
						Title:    "Write docs",
						Status:   domain.StatusInProgress,
						Priority: domain.PriorityMedium,
						EpicID:   "epic-1",
					},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks?status=in_progress&epic=epic-1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "project-1", body["project_id"])

	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "ABC-1", item["id"])
	require.Equal(t, "Write docs", item["title"])
	require.Equal(t, "in_progress", item["status"])
	require.Equal(t, float64(2), item["priority"])
	require.Equal(t, "epic-1", item["epic_id"])
}

func TestSubtaskListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		SubtaskLister: fakeSubtaskLister{
			listResponse: domain.TaskListResponse{
				ProjectID: "project-1",
				Items: []domain.TaskListItem{
					{ID: "ABC-2", Title: "Write docs", Status: domain.StatusTodo, Priority: domain.PriorityMedium, ParentTaskID: "ABC-1"},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks/ABC-1/subtasks", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "project-1", body["project_id"])
	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
}

func TestCommentAddReturnsComment(t *testing.T) {
	handler := newApp(Config{
		CommentCreator: fakeCommentCreator{
			createResponse: domain.Comment{
				ID:      "cmt-1",
				Author:  "agent",
				Kind:    "summary",
				Content: "Done",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/tasks/ABC-1/comments", strings.NewReader(`{"author":"agent","kind":"summary","content":"Done"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "cmt-1", body["id"])
	require.Equal(t, "agent", body["author"])
	require.Equal(t, "summary", body["kind"])
}

func TestCommentListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		CommentLister: fakeCommentLister{
			listResponse: domain.CommentListResponse{
				ProjectID: "project-1",
				TaskID:    "ABC-1",
				Items: []domain.Comment{
					{ID: "cmt-1", Author: "agent", Kind: "analysis", Content: "Working notes"},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks/ABC-1/comments", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
}

func TestCommentUpdateReturnsComment(t *testing.T) {
	handler := newApp(Config{
		CommentUpdater: fakeCommentUpdater{
			updateResponse: domain.Comment{
				ID:      "cmt-1",
				Author:  "agent",
				Kind:    "comment",
				Content: "Updated",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/projects/project-1/comments/cmt-1", strings.NewReader(`{"content":"Updated"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "cmt-1", body["id"])
	require.Equal(t, "Updated", body["content"])
}

func TestCommentDeleteReturnsComment(t *testing.T) {
	handler := newApp(Config{
		CommentDeleter: fakeCommentDeleter{
			deleteResponse: domain.Comment{
				ID:      "cmt-1",
				Author:  "agent",
				Kind:    "comment",
				Content: "Deleted",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1/comments/cmt-1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "cmt-1", body["id"])
	require.Equal(t, "Deleted", body["content"])
}

func TestReadyReturnsItems(t *testing.T) {
	handler := newApp(Config{
		ReadyLister: fakeReadyLister{
			listResponse: domain.ReadyListResponse{
				ProjectID: "project-1",
				Items: []domain.ReadyListItem{
					{
						ID:       "CORE-1",
						Title:    "Health endpoints",
						Status:   domain.StatusTodo,
						Priority: domain.PriorityHigh,
						EpicID:   "epic-1",
						Tags:     []string{"infra", "api"},
						Unblocks: []domain.TaskListItem{
							{
								ID:       "CORE-5",
								Title:    "Backups",
								Status:   domain.StatusTodo,
								Priority: domain.PriorityHigh,
								EpicID:   "epic-1",
								Tags:     []string{"infra"},
							},
						},
					},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/ready?epic=epic-1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "project-1", body["project_id"])

	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "CORE-1", item["id"])
	require.Equal(t, "Health endpoints", item["title"])
	require.Equal(t, "todo", item["status"])
	require.Equal(t, float64(1), item["priority"])

	unblocks, ok := item["unblocks"].([]any)
	require.True(t, ok)
	require.Len(t, unblocks, 1)
	blocked, ok := unblocks[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "CORE-5", blocked["id"])
	require.Equal(t, "Backups", blocked["title"])
}
