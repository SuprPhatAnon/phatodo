package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAPIClientCallsExpectedEndpoints(t *testing.T) {
	ctx := context.Background()
	type seenRequest struct {
		method string
		path   string
		query  string
	}
	var seen []seenRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "key-1", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret-1", r.Header.Get("X-Phatodo-Access-Secret"))
		seen = append(seen, seenRequest{method: r.Method, path: r.URL.Path, query: r.URL.RawQuery})
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/config":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"key":"theme","value":"dark"}]}`))
		case r.URL.Path == "/api/v1/projects/project-1/config/theme":
			_, _ = w.Write([]byte(`{"key":"theme","value":"dark"}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/projects/project-1/epics"):
			if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/epics" {
				_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":"EPIC-1","title":"Track auth","status":"todo","priority":1}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":"EPIC-1","title":"Track auth","status":"todo","priority":1}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/tasks":
			_, _ = w.Write([]byte(`{"id":"TASK-1","issue_prefix":"TASK","title":"Write tests","status":"todo","priority":2,"project_id":"project-1"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/subtasks":
			_, _ = w.Write([]byte(`{"id":"TASK-2","issue_prefix":"TASK","title":"Subtask","status":"todo","priority":2,"project_id":"project-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/tasks":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":"TASK-1","title":"Write tests","status":"todo","priority":2}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/subtasks":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":"TASK-2","title":"Subtask","status":"todo","priority":2}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/comments":
			_, _ = w.Write([]byte(`{"project_id":"project-1","task_id":"TASK-1","items":[{"id":"cmt-1","author":"codex","kind":"comment","content":"hello"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/comments":
			_, _ = w.Write([]byte(`{"id":"cmt-1","author":"codex","kind":"comment","content":"hello"}`))
		case r.URL.Path == "/api/v1/projects/project-1/comments/cmt-1":
			_, _ = w.Write([]byte(`{"id":"cmt-1","author":"codex","kind":"comment","content":"updated"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/dependencies":
			_, _ = w.Write([]byte(`{"project_id":"project-1","task_id":"TASK-1","items":[{"id":"dep-1","task_id":"TASK-1","depends_on_id":"TASK-0"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/dependencies":
			_, _ = w.Write([]byte(`{"id":"dep-1","task_id":"TASK-1","depends_on_id":"TASK-0"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/project-1/tasks/TASK-1/dependencies/TASK-0":
			_, _ = w.Write([]byte(`{"id":"dep-1","task_id":"TASK-1","depends_on_id":"TASK-0"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/locks":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":"lock-1","entity_type":"task","entity_id":"TASK-1","locked_by":"usr-1"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/project-1/locks":
			_, _ = w.Write([]byte(`{"id":"lock-1","entity_type":"task","entity_id":"TASK-1","locked_by":"usr-1"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/project-1/locks/lock-1":
			_, _ = w.Write([]byte(`{"id":"lock-1","entity_type":"task","entity_id":"TASK-1","locked_by":"usr-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/search":
			_, _ = w.Write([]byte(`{"project_id":"project-1","query":"tests","items":[{"entity_type":"task","id":"TASK-1","title":"Write tests","status":"todo"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/history":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":1,"entity_type":"task","entity_id":"TASK-1","action":"update"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/list":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"entity_type":"task","id":"TASK-1","title":"Write tests","status":"todo","priority":2}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/project-1/ready":
			_, _ = w.Write([]byte(`{"project_id":"project-1","items":[{"id":"TASK-1","title":"Write tests","status":"todo","priority":2}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/init":
			_, _ = w.Write([]byte(`{"user_id":"usr-1","username":"admin","access_key":"admin-key","access_secret":"admin-secret"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/bootstrap":
			_, _ = w.Write([]byte(`{"workspace_id":"workspace-1","project_id":"project-1","access_key":"key-1","access_secret":"secret-1"}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/projects/project-1/tasks/TASK-1"):
			_, _ = w.Write([]byte(`{"id":"TASK-1","title":"Write tests","status":"todo","priority":2}`))
		default:
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()

	client, err := NewAPIClient(config.LocalConfig{APIURL: server.URL, AccessKey: "key-1", AccessSecret: "secret-1"})
	require.NoError(t, err)

	configItems, err := client.ListProjectConfig(ctx, "project-1")
	require.NoError(t, err)
	require.Len(t, configItems, 1)
	item, err := client.GetProjectConfig(ctx, "project-1", "theme")
	require.NoError(t, err)
	require.Equal(t, "dark", item.Value)
	_, err = client.SetProjectConfig(ctx, "project-1", "theme", "dark")
	require.NoError(t, err)
	_, err = client.UnsetProjectConfig(ctx, "project-1", "theme")
	require.NoError(t, err)

	_, err = client.CreateEpic(ctx, "project-1", domain.EpicCreateRequest{Title: "Track auth"})
	require.NoError(t, err)
	_, err = client.ListEpics(ctx, "project-1", "todo", 5)
	require.NoError(t, err)
	_, err = client.GetEpic(ctx, "project-1", "EPIC-1")
	require.NoError(t, err)
	_, err = client.UpdateEpic(ctx, "project-1", "EPIC-1", domain.EpicUpdateRequest{})
	require.NoError(t, err)
	_, err = client.CompleteEpic(ctx, "project-1", "EPIC-1")
	require.NoError(t, err)
	_, err = client.DeleteEpic(ctx, "project-1", "EPIC-1")
	require.NoError(t, err)

	_, err = client.CreateTask(ctx, "project-1", domain.TaskCreateRequest{Title: "Write tests", IssuePrefix: "TASK"})
	require.NoError(t, err)
	_, err = client.CreateSubtask(ctx, "project-1", "TASK-1", domain.TaskCreateRequest{Title: "Subtask"})
	require.NoError(t, err)
	_, err = client.GetTask(ctx, "project-1", "TASK-1")
	require.NoError(t, err)
	_, err = client.UpdateTask(ctx, "project-1", "TASK-1", domain.TaskUpdateRequest{})
	require.NoError(t, err)
	_, err = client.DeleteTask(ctx, "project-1", "TASK-1")
	require.NoError(t, err)
	_, err = client.ListTasks(ctx, "project-1", "todo", "EPIC-1", 10)
	require.NoError(t, err)
	_, err = client.ListSubtasks(ctx, "project-1", "TASK-1", 10)
	require.NoError(t, err)

	_, err = client.ListComments(ctx, "project-1", "TASK-1")
	require.NoError(t, err)
	_, err = client.AddComment(ctx, "project-1", "TASK-1", domain.CommentCreateRequest{Author: "codex", Content: "hello"})
	require.NoError(t, err)
	_, err = client.UpdateComment(ctx, "project-1", "cmt-1", domain.CommentUpdateRequest{Content: "updated"})
	require.NoError(t, err)
	_, err = client.DeleteComment(ctx, "project-1", "cmt-1")
	require.NoError(t, err)

	_, err = client.ListDependencies(ctx, "project-1", "TASK-1")
	require.NoError(t, err)
	_, err = client.AddDependency(ctx, "project-1", "TASK-1", "TASK-0")
	require.NoError(t, err)
	_, err = client.RemoveDependency(ctx, "project-1", "TASK-1", "TASK-0")
	require.NoError(t, err)

	_, err = client.ListLocks(ctx, "project-1", "task,epic", "TASK-1", true)
	require.NoError(t, err)
	_, err = client.AcquireLock(ctx, "project-1", domain.LockAcquireRequest{EntityType: "task", EntityID: "TASK-1"})
	require.NoError(t, err)
	_, err = client.ReleaseLock(ctx, "project-1", "lock-1")
	require.NoError(t, err)
	_, err = client.Search(ctx, "project-1", "tests", "task", "todo", 3)
	require.NoError(t, err)
	_, err = client.History(ctx, "project-1", "TASK-1", "task", "update", "2026-01-01", 3)
	require.NoError(t, err)
	_, err = client.ListUnified(ctx, "project-1", "task", "todo", "2", "priority:asc", 3)
	require.NoError(t, err)
	_, err = client.ListReadyTasks(ctx, "project-1", "EPIC-1", 3)
	require.NoError(t, err)
	_, err = client.InitAdmin(ctx, domain.AdminInitRequest{Username: "admin", Password: "secret"})
	require.NoError(t, err)
	_, err = client.BootstrapAdmin(ctx, domain.AdminBootstrapRequest{Username: "admin", Password: "secret", WorkspaceName: "workspace", ProjectName: "project"})
	require.NoError(t, err)

	require.Greater(t, len(seen), 30)
	require.Contains(t, seen, seenRequest{method: http.MethodGet, path: "/api/v1/projects/project-1/tasks", query: "epic=EPIC-1&limit=10&status=todo"})
	require.Contains(t, seen, seenRequest{method: http.MethodGet, path: "/api/v1/projects/project-1/locks", query: "active=true&entity=TASK-1&type=task%2Cepic"})
}

func TestAPIClientReturnsAPIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "task_conflict", "message": "already locked"})
	}))
	defer server.Close()

	client, err := NewAPIClient(config.LocalConfig{APIURL: server.URL})
	require.NoError(t, err)

	_, err = client.GetTask(context.Background(), "project-1", "TASK-1")
	require.EqualError(t, err, "task_conflict: already locked")
}

func TestAPIClientReturnsDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{not-json"))
	}))
	defer server.Close()

	client, err := NewAPIClient(config.LocalConfig{APIURL: server.URL})
	require.NoError(t, err)

	_, err = client.GetTask(context.Background(), "project-1", "TASK-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode api response")
}

func TestAPIClientGETHelpersReturnAPIErrorMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_query", "message": "bad filter"})
	}))
	defer server.Close()

	client, err := NewAPIClient(config.LocalConfig{APIURL: server.URL})
	require.NoError(t, err)

	_, err = client.Search(context.Background(), "project-1", "tests", "", "", 0)
	require.EqualError(t, err, "invalid_query: bad filter")

	_, err = client.ListReadyTasks(context.Background(), "project-1", "", 0)
	require.EqualError(t, err, "invalid_query: bad filter")

	_, err = client.ListTasks(context.Background(), "project-1", "", "", 0)
	require.EqualError(t, err, "invalid_query: bad filter")
}

func TestNewAPIClientRejectsInvalidURL(t *testing.T) {
	_, err := NewAPIClient(config.LocalConfig{APIURL: "http://%zz"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse api url")
}
