package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/SuprPhatAnon/phatodo/internal/storage/postgres"
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

func TestDecodeTaskCreateRequestDefaultsKindAndRejectsBugWithoutRootCause(t *testing.T) {
	req, err := decodeTaskCreateRequest(strings.NewReader(`{"title":"Write docs","issue_prefix":"ABC"}`))
	require.NoError(t, err)
	require.Equal(t, domain.TaskKindTask, req.Kind)

	req, err = decodeTaskCreateRequest(strings.NewReader(`{"title":"Add import","issue_prefix":"ABC","kind":"feature"}`))
	require.NoError(t, err)
	require.Equal(t, domain.TaskKindFeature, req.Kind)

	req, err = decodeTaskCreateRequest(strings.NewReader(`{"title":"Plan files","issue_prefix":"ABC","planned_files":["internal/cli/commands.go"]}`))
	require.NoError(t, err)
	require.Equal(t, []string{"internal/cli/commands.go"}, req.PlannedFiles)

	_, err = decodeTaskCreateRequest(strings.NewReader(`{"title":"Write docs","issue_prefix":"ABC","kind":"bug"}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "root_cause_analysis is required when kind is bug")
}

func TestDecodeSubtaskCreateRequestRejectsBugWithoutRootCause(t *testing.T) {
	_, err := decodeSubtaskCreateRequest(strings.NewReader(`{"title":"Write docs","kind":"bug"}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "root_cause_analysis is required when kind is bug")
}

func TestDecodeTaskUpdateRequestValidatesRootCauseAnalysisEmpty(t *testing.T) {
	_, err := decodeTaskUpdateRequest(strings.NewReader(`{"root_cause_analysis":"   "}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "root_cause_analysis cannot be empty")
}

func TestDecodeTaskUpdateRequestValidatesChangedFilesEmpty(t *testing.T) {
	_, err := decodeTaskUpdateRequest(strings.NewReader(`{"changed_files":["   "]}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "changed_files cannot be empty")

	req, err := decodeTaskUpdateRequest(strings.NewReader(`{"changed_files":["internal/server/handlers.go"]}`))
	require.NoError(t, err)
	require.NotNil(t, req.ChangedFiles)
	require.Equal(t, []string{"internal/server/handlers.go"}, *req.ChangedFiles)
}

func TestEpicListReturnsUnavailableWithoutStore(t *testing.T) {
	handler := newApp(Config{}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/epics", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "epic_store_unavailable" {
		t.Fatalf("expected epic_store_unavailable, got %#v", body["error"])
	}
	if body["message"] != "epic store is not configured" {
		t.Fatalf("expected epic store message, got %#v", body["message"])
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

type fakeAuthResolver struct {
	user domain.User
	err  error
}

func (f fakeAuthResolver) ResolveAPIPrincipal(_ context.Context, _ string, _ string) (domain.User, error) {
	return f.user, f.err
}

type fakeEpicStore struct {
	listResponse     domain.EpicListResponse
	getResponse      domain.Epic
	createResponse   domain.Epic
	updateResponse   domain.Epic
	completeResponse domain.Epic
	deleteResponse   domain.Epic
	err              error
}

func (f fakeEpicStore) ListEpics(_ context.Context, _ string, _ string, _ int) (domain.EpicListResponse, error) {
	return f.listResponse, f.err
}

func (f fakeEpicStore) GetEpic(_ context.Context, _ string, _ string) (domain.Epic, error) {
	return f.getResponse, f.err
}

func (f fakeEpicStore) CreateEpic(_ context.Context, _ string, _ domain.EpicCreateRequest, _ string) (domain.Epic, error) {
	return f.createResponse, f.err
}

func (f fakeEpicStore) UpdateEpic(_ context.Context, _ string, _ string, _ domain.EpicUpdateRequest, _ string) (domain.Epic, error) {
	return f.updateResponse, f.err
}

func (f fakeEpicStore) CompleteEpic(_ context.Context, _ string, _ string, _ string) (domain.Epic, error) {
	return f.completeResponse, f.err
}

func (f fakeEpicStore) DeleteEpic(_ context.Context, _ string, _ string, _ string) (domain.Epic, error) {
	return f.deleteResponse, f.err
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

func (f fakeSubtaskLister) ListSubtasks(_ context.Context, _ string, _ string, _ int) (domain.TaskListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeTaskLister struct {
	listResponse domain.TaskListResponse
	listErr      error
}

func (f fakeTaskLister) ListTasks(_ context.Context, _ string, _ string, _ string, _ int) (domain.TaskListResponse, error) {
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

type fakeDependencyLister struct {
	listResponse domain.DependencyListResponse
	listErr      error
}

func (f fakeDependencyLister) ListDependencies(_ context.Context, _ string, _ string) (domain.DependencyListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeDependencyAdder struct {
	addResponse domain.Dependency
	addErr      error
}

func (f fakeDependencyAdder) AddDependency(_ context.Context, _ string, _ string, _ string, _ string) (domain.Dependency, error) {
	return f.addResponse, f.addErr
}

type fakeDependencyRemover struct {
	removeResponse domain.Dependency
	removeErr      error
}

func (f fakeDependencyRemover) RemoveDependency(_ context.Context, _ string, _ string, _ string, _ string) (domain.Dependency, error) {
	return f.removeResponse, f.removeErr
}

type fakeLockStore struct {
	listResponse    domain.LockListResponse
	acquireResponse domain.WorkItemLock
	releaseResponse domain.WorkItemLock
	err             error
}

func (f fakeLockStore) ListLocks(_ context.Context, _ string, _ []string, _ string, _ bool) (domain.LockListResponse, error) {
	return f.listResponse, f.err
}

func (f fakeLockStore) AcquireLock(_ context.Context, _ string, _ domain.LockAcquireRequest, _ string) (domain.WorkItemLock, error) {
	return f.acquireResponse, f.err
}

func (f fakeLockStore) ReleaseLock(_ context.Context, _ string, _ string, _ string) (domain.WorkItemLock, error) {
	return f.releaseResponse, f.err
}

type fakeSearcher struct {
	searchResponse domain.SearchResponse
	searchErr      error
}

func (f fakeSearcher) Search(_ context.Context, _ string, _ string, _ string, _ string, _ int) (domain.SearchResponse, error) {
	return f.searchResponse, f.searchErr
}

type fakeHistorian struct {
	historyResponse domain.HistoryResponse
	historyErr      error
}

func (f fakeHistorian) History(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ int) (domain.HistoryResponse, error) {
	return f.historyResponse, f.historyErr
}

type fakeListLister struct {
	listResponse domain.ListResponse
	listErr      error
}

func (f fakeListLister) ListUnified(_ context.Context, _ string, _ string, _ string, _ string, _ string, _ int) (domain.ListResponse, error) {
	return f.listResponse, f.listErr
}

type fakeReadyLister struct {
	listResponse domain.ReadyListResponse
	listErr      error
}

func (f fakeReadyLister) ListReadyTasks(_ context.Context, _ string, _ string, _ int) (domain.ReadyListResponse, error) {
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
	creator := &recordingTaskCreator{
		createResponse: domain.TaskCreateResponse{
			ID:          "ABC-1",
			IssuePrefix: "ABC",
			Title:       "Write docs",
			Status:      domain.StatusTodo,
			Priority:    domain.PriorityMedium,
			ProjectID:   "project-1",
			WorkspaceID: "workspace-1",
		},
	}
	handler := newApp(Config{
		AuthResolver: fakeAuthResolver{
			user: domain.User{
				ID:          "usr_1",
				DisplayName: "CLI User",
				Role:        domain.UserRoleUser,
				AccessKey:   "key",
			},
		},
		TaskCreator: creator,
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/tasks", strings.NewReader(`{"title":"Write docs","issue_prefix":"ABC","priority":2,"tags":["docs"]}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ABC-1", body["id"])
	require.Equal(t, "ABC", body["issue_prefix"])
	require.Equal(t, "Write docs", body["title"])
	require.Equal(t, "usr_1", creator.actorUserID)
}

type recordingTaskCreator struct {
	createResponse domain.TaskCreateResponse
	createErr      error
	actorUserID    string
}

func (f *recordingTaskCreator) CreateTask(_ context.Context, _ string, _ domain.TaskCreateRequest, actorUserID string) (domain.TaskCreateResponse, error) {
	f.actorUserID = actorUserID
	return f.createResponse, f.createErr
}

func TestSearchReturnsItems(t *testing.T) {
	handler := newApp(Config{
		Searcher: fakeSearcher{
			searchResponse: domain.SearchResponse{
				ProjectID: "project-1",
				Query:     "auth bug",
				Items: []domain.SearchItem{
					{
						EntityType: "task",
						ID:         "ABC-1",
						Title:      "Fix auth bug",
						Status:     domain.StatusTodo,
					},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/search?q=auth+bug&type=task&status=todo&limit=10", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "project-1", body["project_id"])
	require.Equal(t, "auth bug", body["query"])
	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
}

func TestHistoryReturnsEvents(t *testing.T) {
	handler := newApp(Config{
		Historian: fakeHistorian{
			historyResponse: domain.HistoryResponse{
				ProjectID: "project-1",
				Items: []domain.HistoryEvent{
					{
						ID:         1,
						EntityType: "task",
						EntityID:   "ABC-1",
						Action:     "update",
					},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/history?entity=ABC-1&type=task&action=update&since=2025-01-01&limit=5", nil)
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

func TestListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		ListLister: fakeListLister{
			listResponse: domain.ListResponse{
				ProjectID: "project-1",
				Items: []domain.UnifiedListItem{
					{
						EntityType: "epic",
						ID:         "EPIC-1",
						Title:      "Track auth",
						Status:     domain.StatusTodo,
						Priority:   domain.PriorityCritical,
					},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/list?type=epic,task&status=todo&priority=0,1&sort=priority:asc,created:desc&limit=2", nil)
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

func TestDependencyAddReturnsEdge(t *testing.T) {
	handler := newApp(Config{
		DependencyAdder: fakeDependencyAdder{
			addResponse: domain.Dependency{
				ID:          "dep-1",
				TaskID:      "ABC-1",
				DependsOnID: "ABC-2",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/tasks/ABC-1/dependencies", strings.NewReader(`{"depends_on_id":"ABC-2"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "dep-1", body["id"])
	require.Equal(t, "ABC-1", body["task_id"])
	require.Equal(t, "ABC-2", body["depends_on_id"])
}

func TestDependencyListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		DependencyLister: fakeDependencyLister{
			listResponse: domain.DependencyListResponse{
				ProjectID: "project-1",
				TaskID:    "ABC-1",
				Items: []domain.Dependency{
					{ID: "dep-1", TaskID: "ABC-1", DependsOnID: "ABC-2"},
				},
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks/ABC-1/dependencies", nil)
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

func TestDependencyRemoveReturnsEdge(t *testing.T) {
	handler := newApp(Config{
		DependencyRemover: fakeDependencyRemover{
			removeResponse: domain.Dependency{
				ID:          "dep-1",
				TaskID:      "ABC-1",
				DependsOnID: "ABC-2",
			},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1/tasks/ABC-1/dependencies/ABC-2", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "dep-1", body["id"])
	require.Equal(t, "ABC-2", body["depends_on_id"])
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

func TestNewBuildsHTTPServer(t *testing.T) {
	server := New(Config{Addr: ":9090"})

	require.Equal(t, ":9090", server.Addr)
	require.NotNil(t, server.Handler)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"database":"not_configured"`)
}

func TestAPIIndexDashboardAndProjectPlaceholderRoutes(t *testing.T) {
	handler := newApp(Config{}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "phatodo API")

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "phatodo dashboard")

	req = httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotImplemented, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"actor_id":"usr_1"`)
}

func TestEpicRoutesReturnItems(t *testing.T) {
	store := fakeEpicStore{
		listResponse: domain.EpicListResponse{
			ProjectID: "project-1",
			Items:     []domain.Epic{{ID: "EPIC-1", Title: "Track auth", Status: domain.StatusTodo, Priority: domain.PriorityHigh}},
		},
		getResponse:      domain.Epic{ID: "EPIC-1", Title: "Track auth", Status: domain.StatusTodo, Priority: domain.PriorityHigh},
		createResponse:   domain.Epic{ID: "EPIC-2", Title: "Create auth", Status: domain.StatusTodo, Priority: domain.PriorityMedium},
		updateResponse:   domain.Epic{ID: "EPIC-1", Title: "Track auth v2", Status: domain.StatusInProgress, Priority: domain.PriorityHigh},
		completeResponse: domain.Epic{ID: "EPIC-1", Title: "Track auth", Status: domain.StatusCompleted, Priority: domain.PriorityHigh},
		deleteResponse:   domain.Epic{ID: "EPIC-1", Title: "Track auth", Status: domain.StatusArchived, Priority: domain.PriorityHigh},
	}
	handler := newApp(Config{
		EpicLister:    store,
		EpicReader:    store,
		EpicCreator:   store,
		EpicUpdater:   store,
		EpicCompleter: store,
		EpicDeleter:   store,
	}).routes()

	cases := []struct {
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{http.MethodGet, "/api/v1/projects/project-1/epics?status=todo&limit=5", "", http.StatusOK, "EPIC-1"},
		{http.MethodPost, "/api/v1/projects/project-1/epics", `{"title":"Create auth","priority":2}`, http.StatusCreated, "EPIC-2"},
		{http.MethodGet, "/api/v1/projects/project-1/epics/EPIC-1", "", http.StatusOK, "Track auth"},
		{http.MethodPatch, "/api/v1/projects/project-1/epics/EPIC-1", `{"title":"Track auth v2","status":"in_progress"}`, http.StatusOK, "Track auth v2"},
		{http.MethodPost, "/api/v1/projects/project-1/epics/EPIC-1/complete", "", http.StatusOK, "completed"},
		{http.MethodDelete, "/api/v1/projects/project-1/epics/EPIC-1", "", http.StatusOK, "archived"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set(AccessKeyHeader, "key")
		req.Header.Set(AccessSecretHeader, "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, tc.status, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), tc.want)
	}
}

func TestLockRoutesReturnLocks(t *testing.T) {
	store := fakeLockStore{
		listResponse: domain.LockListResponse{
			ProjectID: "project-1",
			Items: []domain.WorkItemLock{
				{ID: "lock-1", EntityType: "task", EntityID: "TASK-1", LockedBy: "usr_1", Reason: "editing"},
			},
		},
		acquireResponse: domain.WorkItemLock{ID: "lock-2", EntityType: "task", EntityID: "TASK-2", LockedBy: "usr_1", Reason: "editing"},
		releaseResponse: domain.WorkItemLock{ID: "lock-2", EntityType: "task", EntityID: "TASK-2", LockedBy: "usr_1", Reason: "editing"},
	}
	handler := newApp(Config{LockLister: store, LockAcquirer: store, LockReleaser: store}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/locks?type=task,epic&entity=TASK-1&active=true", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "lock-1")

	req = httptest.NewRequest(http.MethodPost, "/api/v1/projects/project-1/locks", strings.NewReader(`{"entity_type":"task","entity_id":"TASK-2","reason":"editing"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	req.Header.Set(UserIDHeader, "usr_1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "lock-2")

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/projects/project-1/locks/lock-2", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "lock-2")
}

func TestLockListRejectsInvalidActiveQuery(t *testing.T) {
	handler := newApp(Config{LockLister: fakeLockStore{}}).routes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/locks?active=maybe", nil)
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "active must be true or false")
}

func TestStoreUnavailableRoutesReturnServiceUnavailable(t *testing.T) {
	handler := newApp(Config{}).routes()
	cases := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{http.MethodPost, "/api/v1/projects/project-1/epics", `{"title":"Epic"}`, "epic_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/epics/EPIC-1", "", "epic_store_unavailable"},
		{http.MethodPatch, "/api/v1/projects/project-1/epics/EPIC-1", `{"title":"Epic"}`, "epic_store_unavailable"},
		{http.MethodPost, "/api/v1/projects/project-1/epics/EPIC-1/complete", "", "epic_store_unavailable"},
		{http.MethodDelete, "/api/v1/projects/project-1/epics/EPIC-1", "", "epic_store_unavailable"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks", `{"title":"Task","issue_prefix":"TASK"}`, "task_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1", "", "task_store_unavailable"},
		{http.MethodPatch, "/api/v1/projects/project-1/tasks/TASK-1", `{"title":"Task"}`, "task_store_unavailable"},
		{http.MethodDelete, "/api/v1/projects/project-1/tasks/TASK-1", "", "task_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1/subtasks", "", "task_store_unavailable"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/comments", `{"author":"codex","content":"hi"}`, "comment_store_unavailable"},
		{http.MethodPatch, "/api/v1/projects/project-1/comments/cmt-1", `{"content":"hi"}`, "comment_store_unavailable"},
		{http.MethodDelete, "/api/v1/projects/project-1/comments/cmt-1", "", "comment_store_unavailable"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", `{"depends_on_id":"TASK-0"}`, "dependency_store_unavailable"},
		{http.MethodDelete, "/api/v1/projects/project-1/tasks/TASK-1/dependencies/TASK-0", "", "dependency_store_unavailable"},
		{http.MethodPost, "/api/v1/projects/project-1/locks", `{"entity_type":"task","entity_id":"TASK-1"}`, "lock_store_unavailable"},
		{http.MethodDelete, "/api/v1/projects/project-1/locks/lock-1", "", "lock_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/search?q=test", "", "search_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/history", "", "history_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/list", "", "list_store_unavailable"},
		{http.MethodGet, "/api/v1/projects/project-1/ready", "", "ready_store_unavailable"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set(AccessKeyHeader, "key")
		req.Header.Set(AccessSecretHeader, "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusServiceUnavailable, rec.Code, tc.path+": "+rec.Body.String())
		require.Contains(t, rec.Body.String(), tc.want)
	}
}

func TestInvalidRequestRoutesReturnBadRequest(t *testing.T) {
	handler := newApp(Config{
		EpicCreator:     fakeEpicStore{},
		EpicUpdater:     fakeEpicStore{},
		TaskCreator:     fakeTaskCreator{},
		TaskUpdater:     fakeTaskUpdater{},
		CommentCreator:  fakeCommentCreator{},
		CommentUpdater:  fakeCommentUpdater{},
		DependencyAdder: fakeDependencyAdder{},
		LockAcquirer:    fakeLockStore{},
		Searcher:        fakeSearcher{},
	}).routes()
	cases := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{http.MethodPost, "/api/v1/projects/project-1/epics", `{}`, "title is required"},
		{http.MethodPatch, "/api/v1/projects/project-1/epics/EPIC-1", `{"status":"bad"}`, "epic status must"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks", `{"title":"Task"}`, "issue_prefix is required"},
		{http.MethodPatch, "/api/v1/projects/project-1/tasks/TASK-1", `{}`, "at least one field must be provided"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/comments", `{"author":"codex"}`, "content is required"},
		{http.MethodPatch, "/api/v1/projects/project-1/comments/cmt-1", `{}`, "content is required"},
		{http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", `{}`, "depends_on_id is required"},
		{http.MethodPost, "/api/v1/projects/project-1/locks", `{"entity_type":"task"}`, "entity_id is required"},
		{http.MethodGet, "/api/v1/projects/project-1/search", "", "search query is required"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set(AccessKeyHeader, "key")
		req.Header.Set(AccessSecretHeader, "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code, tc.path+": "+rec.Body.String())
		require.Contains(t, rec.Body.String(), tc.want)
	}
}

func TestKnownStoreErrorsMapToHTTPResponses(t *testing.T) {
	cases := []struct {
		name   string
		config Config
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{"epic list project missing", Config{EpicLister: fakeEpicStore{err: postgres.ErrProjectNotFound}}, http.MethodGet, "/api/v1/projects/project-1/epics", "", http.StatusNotFound, "project_not_found"},
		{"epic show missing", Config{EpicReader: fakeEpicStore{err: postgres.ErrEpicNotFound}}, http.MethodGet, "/api/v1/projects/project-1/epics/EPIC-1", "", http.StatusNotFound, "epic_not_found"},
		{"task update evidence", Config{TaskUpdater: fakeTaskUpdater{updateErr: postgres.ErrTaskCompletionRequiresEvidence}}, http.MethodPatch, "/api/v1/projects/project-1/tasks/TASK-1", `{"status":"completed"}`, http.StatusBadRequest, "completion_evidence_required"},
		{"dependency conflict", Config{DependencyAdder: fakeDependencyAdder{addErr: postgres.ErrDuplicateDependency}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", `{"depends_on_id":"TASK-0"}`, http.StatusConflict, "dependency_exists"},
		{"lock conflict", Config{LockAcquirer: fakeLockStore{err: postgres.ErrLockConflict}}, http.MethodPost, "/api/v1/projects/project-1/locks", `{"entity_type":"task","entity_id":"TASK-1"}`, http.StatusConflict, "lock_conflict"},
		{"admin exists", Config{BootstrapManager: fakeBootstrapManager{initErr: postgres.ErrAdminAlreadyExists}}, http.MethodPost, "/api/v1/admin/init", `{"username":"alice","password":"secret"}`, http.StatusConflict, "admin_already_exists"},
		{"admin credentials", Config{BootstrapManager: fakeBootstrapManager{bootstrapErr: postgres.ErrInvalidAdminCredentials}}, http.MethodPost, "/api/v1/admin/bootstrap", `{"username":"alice","password":"secret","workspace_name":"ws","project_name":"project"}`, http.StatusUnauthorized, "invalid_credentials"},
		{"config get missing", Config{ProjectConfigReader: fakeProjectConfigReader{err: postgres.ErrProjectConfigNotFound}}, http.MethodGet, "/api/v1/projects/project-1/config/theme", "", http.StatusNotFound, "project_config_not_found"},
		{"config set project missing", Config{ProjectConfigWriter: fakeProjectConfigWriter{err: postgres.ErrProjectNotFound}}, http.MethodPut, "/api/v1/projects/project-1/config/theme", `{"value":"dark"}`, http.StatusNotFound, "project_not_found"},
		{"task create invalid kind", Config{TaskCreator: fakeTaskCreator{createErr: postgres.ErrInvalidTaskKind}}, http.MethodPost, "/api/v1/projects/project-1/tasks", `{"title":"Task","issue_prefix":"TASK"}`, http.StatusBadRequest, "invalid_task_kind"},
		{"subtask create task missing", Config{TaskCreator: fakeTaskCreator{createErr: postgres.ErrTaskNotFound}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/subtasks", `{"title":"Subtask"}`, http.StatusNotFound, "task_not_found"},
		{"comment list task missing", Config{CommentLister: fakeCommentLister{listErr: postgres.ErrTaskNotFound}}, http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1/comments", "", http.StatusNotFound, "task_not_found"},
		{"comment update missing", Config{CommentUpdater: fakeCommentUpdater{updateErr: postgres.ErrCommentNotFound}}, http.MethodPatch, "/api/v1/projects/project-1/comments/cmt-1", `{"content":"hi"}`, http.StatusNotFound, "comment_not_found"},
		{"dependency cycle", Config{DependencyAdder: fakeDependencyAdder{addErr: postgres.ErrDependencyCycle}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", `{"depends_on_id":"TASK-0"}`, http.StatusBadRequest, "dependency_cycle"},
		{"dependency remove missing", Config{DependencyRemover: fakeDependencyRemover{removeErr: postgres.ErrDependencyNotFound}}, http.MethodDelete, "/api/v1/projects/project-1/tasks/TASK-1/dependencies/TASK-0", "", http.StatusNotFound, "dependency_not_found"},
		{"lock invalid type", Config{LockAcquirer: fakeLockStore{err: postgres.ErrInvalidLockEntityType}}, http.MethodPost, "/api/v1/projects/project-1/locks", `{"entity_type":"bad","entity_id":"TASK-1"}`, http.StatusBadRequest, "invalid_entity_type"},
		{"lock release missing", Config{LockReleaser: fakeLockStore{err: postgres.ErrLockNotFound}}, http.MethodDelete, "/api/v1/projects/project-1/locks/lock-1", "", http.StatusNotFound, "lock_not_found"},
		{"search project missing", Config{Searcher: fakeSearcher{searchErr: postgres.ErrProjectNotFound}}, http.MethodGet, "/api/v1/projects/project-1/search?q=test", "", http.StatusNotFound, "project_not_found"},
		{"history project missing", Config{Historian: fakeHistorian{historyErr: postgres.ErrProjectNotFound}}, http.MethodGet, "/api/v1/projects/project-1/history", "", http.StatusNotFound, "project_not_found"},
		{"list project missing", Config{ListLister: fakeListLister{listErr: postgres.ErrProjectNotFound}}, http.MethodGet, "/api/v1/projects/project-1/list", "", http.StatusNotFound, "project_not_found"},
		{"ready project missing", Config{ReadyLister: fakeReadyLister{listErr: postgres.ErrProjectNotFound}}, http.MethodGet, "/api/v1/projects/project-1/ready", "", http.StatusNotFound, "project_not_found"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := newApp(tc.config).routes()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set(AccessKeyHeader, "key")
			req.Header.Set(AccessSecretHeader, "secret")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			require.Contains(t, rec.Body.String(), tc.want)
		})
	}
}

func TestStoreInternalErrorsReturnServerErrors(t *testing.T) {
	boom := errors.New("boom")
	cases := []struct {
		name   string
		config Config
		method string
		path   string
		body   string
		want   string
	}{
		{"config list", Config{ProjectConfigReader: fakeProjectConfigReader{err: boom}}, http.MethodGet, "/api/v1/projects/project-1/config", "", "project_config_list_failed"},
		{"config get", Config{ProjectConfigReader: fakeProjectConfigReader{err: boom}}, http.MethodGet, "/api/v1/projects/project-1/config/theme", "", "project_config_get_failed"},
		{"config set", Config{ProjectConfigWriter: fakeProjectConfigWriter{err: boom}}, http.MethodPut, "/api/v1/projects/project-1/config/theme", `{"value":"dark"}`, "project_config_set_failed"},
		{"config unset", Config{ProjectConfigWriter: fakeProjectConfigWriter{err: boom}}, http.MethodDelete, "/api/v1/projects/project-1/config/theme", "", "project_config_unset_failed"},
		{"epic list", Config{EpicLister: fakeEpicStore{err: boom}}, http.MethodGet, "/api/v1/projects/project-1/epics", "", "epic_list_failed"},
		{"epic create", Config{EpicCreator: fakeEpicStore{err: boom}}, http.MethodPost, "/api/v1/projects/project-1/epics", `{"title":"Epic"}`, "epic_create_failed"},
		{"epic show", Config{EpicReader: fakeEpicStore{err: boom}}, http.MethodGet, "/api/v1/projects/project-1/epics/EPIC-1", "", "epic_show_failed"},
		{"epic update", Config{EpicUpdater: fakeEpicStore{err: boom}}, http.MethodPatch, "/api/v1/projects/project-1/epics/EPIC-1", `{"title":"Epic"}`, "epic_update_failed"},
		{"epic complete", Config{EpicCompleter: fakeEpicStore{err: boom}}, http.MethodPost, "/api/v1/projects/project-1/epics/EPIC-1/complete", "", "epic_complete_failed"},
		{"epic delete", Config{EpicDeleter: fakeEpicStore{err: boom}}, http.MethodDelete, "/api/v1/projects/project-1/epics/EPIC-1", "", "epic_delete_failed"},
		{"task create", Config{TaskCreator: fakeTaskCreator{createErr: boom}}, http.MethodPost, "/api/v1/projects/project-1/tasks", `{"title":"Task","issue_prefix":"TASK"}`, "task_create_failed"},
		{"subtask create", Config{TaskCreator: fakeTaskCreator{createErr: boom}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/subtasks", `{"title":"Task"}`, "subtask_create_failed"},
		{"task show", Config{TaskReader: fakeTaskReader{getErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1", "", "task_show_failed"},
		{"task list", Config{TaskLister: fakeTaskLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/tasks", "", "task_list_failed"},
		{"subtask list", Config{SubtaskLister: fakeSubtaskLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1/subtasks", "", "subtask_list_failed"},
		{"task update", Config{TaskUpdater: fakeTaskUpdater{updateErr: boom}}, http.MethodPatch, "/api/v1/projects/project-1/tasks/TASK-1", `{"title":"Task"}`, "task_update_failed"},
		{"task delete", Config{TaskDeleter: fakeTaskDeleter{deleteErr: boom}}, http.MethodDelete, "/api/v1/projects/project-1/tasks/TASK-1", "", "task_delete_failed"},
		{"comment list", Config{CommentLister: fakeCommentLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1/comments", "", "comment_list_failed"},
		{"comment create", Config{CommentCreator: fakeCommentCreator{createErr: boom}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/comments", `{"author":"codex","content":"hi"}`, "comment_create_failed"},
		{"comment update", Config{CommentUpdater: fakeCommentUpdater{updateErr: boom}}, http.MethodPatch, "/api/v1/projects/project-1/comments/cmt-1", `{"content":"hi"}`, "comment_update_failed"},
		{"comment delete", Config{CommentDeleter: fakeCommentDeleter{deleteErr: boom}}, http.MethodDelete, "/api/v1/projects/project-1/comments/cmt-1", "", "comment_delete_failed"},
		{"dependency list", Config{DependencyLister: fakeDependencyLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", "", "dependency_list_failed"},
		{"dependency add", Config{DependencyAdder: fakeDependencyAdder{addErr: boom}}, http.MethodPost, "/api/v1/projects/project-1/tasks/TASK-1/dependencies", `{"depends_on_id":"TASK-0"}`, "dependency_add_failed"},
		{"dependency remove", Config{DependencyRemover: fakeDependencyRemover{removeErr: boom}}, http.MethodDelete, "/api/v1/projects/project-1/tasks/TASK-1/dependencies/TASK-0", "", "dependency_remove_failed"},
		{"lock list", Config{LockLister: fakeLockStore{err: boom}}, http.MethodGet, "/api/v1/projects/project-1/locks", "", "lock_list_failed"},
		{"lock acquire", Config{LockAcquirer: fakeLockStore{err: boom}}, http.MethodPost, "/api/v1/projects/project-1/locks", `{"entity_type":"task","entity_id":"TASK-1"}`, "lock_acquire_failed"},
		{"lock release", Config{LockReleaser: fakeLockStore{err: boom}}, http.MethodDelete, "/api/v1/projects/project-1/locks/lock-1", "", "lock_release_failed"},
		{"search", Config{Searcher: fakeSearcher{searchErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/search?q=test", "", "search_failed"},
		{"history", Config{Historian: fakeHistorian{historyErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/history", "", "history_failed"},
		{"list", Config{ListLister: fakeListLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/list", "", "list_failed"},
		{"ready", Config{ReadyLister: fakeReadyLister{listErr: boom}}, http.MethodGet, "/api/v1/projects/project-1/ready", "", "ready_list_failed"},
		{"admin init", Config{BootstrapManager: fakeBootstrapManager{initErr: boom}}, http.MethodPost, "/api/v1/admin/init", `{"username":"alice","password":"secret"}`, "admin_init_failed"},
		{"admin bootstrap", Config{BootstrapManager: fakeBootstrapManager{bootstrapErr: boom}}, http.MethodPost, "/api/v1/admin/bootstrap", `{"username":"alice","password":"secret","workspace_name":"ws","project_name":"project"}`, "admin_bootstrap_failed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := newApp(tc.config).routes()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set(AccessKeyHeader, "key")
			req.Header.Set(AccessSecretHeader, "secret")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
			require.Contains(t, rec.Body.String(), tc.want)
		})
	}
}

func TestDecodeHelpersRejectMalformedAndMissingFields(t *testing.T) {
	_, err := decodeAdminInitRequest(strings.NewReader(`{"password":"secret"}`))
	require.Error(t, err)
	_, err = decodeAdminBootstrapRequest(strings.NewReader(`{"username":"alice","password":"secret"}`))
	require.Error(t, err)
	_, err = decodeEpicUpdateRequest(strings.NewReader(`{}`))
	require.Error(t, err)
	_, err = decodeProjectConfigSetRequest(strings.NewReader(`{bad-json`))
	require.Error(t, err)
	_, err = decodeTaskCreateRequest(strings.NewReader(`{"title":"Task","issue_prefix":"TASK","kind":"story"}`))
	require.Error(t, err)
	_, err = decodeSubtaskCreateRequest(strings.NewReader(`{"title":"Subtask","kind":"story"}`))
	require.Error(t, err)
	_, err = decodeCommentCreateRequest(strings.NewReader(`{"author":"codex","content":"hi","kind":"note"}`))
	require.Error(t, err)
	req, err := decodeLockAcquireRequest(strings.NewReader(`{"entity_type":"task","entity_id":"TASK-1"}`))
	require.NoError(t, err)
	require.Equal(t, "1h", req.TTL)
	require.False(t, isAllowedTaskKind("story"))
	require.False(t, isAllowedTaskStatus("blocked"))
	require.False(t, isAllowedEpicStatus("blocked"))
}
