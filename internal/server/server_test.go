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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-1/tasks", nil)
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
	if body["action"] != "task.list" {
		t.Fatalf("expected task.list action, got %#v", body["action"])
	}
	if body["project_id"] != "project-1" {
		t.Fatalf("expected project id, got %#v", body["project_id"])
	}
}

type fakeProjectConfigReader struct {
	items []domain.ProjectConfig
	err   error
}

func (f fakeProjectConfigReader) ListProjectConfig(_ context.Context, _ string) ([]domain.ProjectConfig, error) {
	return f.items, f.err
}

type fakeProjectConfigWriter struct {
	item domain.ProjectConfig
	err  error
}

func (f fakeProjectConfigWriter) SetProjectConfig(_ context.Context, _ string, _ string, _ string) (domain.ProjectConfig, error) {
	return f.item, f.err
}

func TestProjectConfigListReturnsItems(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigReader: fakeProjectConfigReader{
			items: []domain.ProjectConfig{{Key: "issue_prefix", Value: "ABC"}},
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
	require.Equal(t, "issue_prefix", item["key"])
	require.Equal(t, "ABC", item["value"])
}

func TestProjectConfigSetReturnsItem(t *testing.T) {
	handler := newApp(Config{
		ProjectConfigWriter: fakeProjectConfigWriter{
			item: domain.ProjectConfig{Key: "issue_prefix", Value: "ABC"},
		},
	}).routes()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/project-1/config/issue_prefix", strings.NewReader(`{"value":"ABC"}`))
	req.Header.Set(AccessKeyHeader, "key")
	req.Header.Set(AccessSecretHeader, "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "issue_prefix", body["key"])
	require.Equal(t, "ABC", body["value"])
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/bootstrap", strings.NewReader(`{"username":"alice","password":"secret","workspace_name":"phatodo","project_name":"phatodo","issue_prefix":"PHA"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ws_1", body["workspace_id"])
	require.Equal(t, "prj_1", body["project_id"])
	require.Equal(t, "key_1", body["access_key"])
}
