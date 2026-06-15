package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestRunAcceptsTrekkerCompatibleTaskCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--toon", "task", "list", "--status", "in_progress"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "task list") {
		t.Fatalf("expected accepted command output, got %q", stdout.String())
	}
}

func TestRunInitWritesLocalConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}

	code := Run([]string{"init"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}

	configPath := filepath.Join(workdir, ".phatodo", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"project_id": "default"`) {
		t.Fatalf("expected project id in config, got %s", string(data))
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"unknown"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}

func TestRunConfigListFetchesProjectConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/v1/projects/default/config", r.URL.Path)
		require.Equal(t, "key", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret", r.Header.Get("X-Phatodo-Access-Secret"))

		_ = json.NewEncoder(w).Encode(map[string]any{
			"project_id": "default",
			"items": []map[string]string{
				{"key": "theme", "value": "dark"},
			},
		})
	}))
	t.Cleanup(server.Close)

	cfg := config.LocalConfig{
		APIURL:       server.URL,
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "list"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme=dark")
}

func TestRunConfigSetCallsServer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/api/v1/projects/default/config/theme", r.URL.Path)
		require.Equal(t, "key", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret", r.Header.Get("X-Phatodo-Access-Secret"))

		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "dark", body["value"])

		_ = json.NewEncoder(w).Encode(map[string]string{
			"key":   "theme",
			"value": "dark",
		})
	}))
	t.Cleanup(server.Close)

	cfg := config.LocalConfig{
		APIURL:       server.URL,
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "set", "theme", "dark"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme=dark")
}

func TestRunConfigGetCallsServer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/v1/projects/default/config/theme", r.URL.Path)
		require.Equal(t, "key", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret", r.Header.Get("X-Phatodo-Access-Secret"))

		_ = json.NewEncoder(w).Encode(map[string]string{
			"key":   "theme",
			"value": "dark",
		})
	}))
	t.Cleanup(server.Close)

	cfg := config.LocalConfig{
		APIURL:       server.URL,
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "get", "theme"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme=dark")
}

func TestRunConfigUnsetCallsServer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/api/v1/projects/default/config/theme", r.URL.Path)
		require.Equal(t, "key", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret", r.Header.Get("X-Phatodo-Access-Secret"))

		_ = json.NewEncoder(w).Encode(map[string]string{
			"key":   "theme",
			"value": "dark",
		})
	}))
	t.Cleanup(server.Close)

	cfg := config.LocalConfig{
		APIURL:       server.URL,
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "unset", "theme"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme=dark")
}

func TestRunTaskCreateCallsServer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/projects/default/tasks", r.URL.Path)
		require.Equal(t, "key", r.Header.Get("X-Phatodo-Access-Key"))
		require.Equal(t, "secret", r.Header.Get("X-Phatodo-Access-Secret"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "Write docs", body["title"])
		require.Equal(t, "ABC", body["issue_prefix"])
		require.Equal(t, "dark", body["tags"].([]any)[0])

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "ABC-1",
			"issue_prefix": "ABC",
			"title":        "Write docs",
			"status":       "todo",
			"priority":     2,
			"project_id":   "default",
			"workspace_id": "default",
		})
	}))
	t.Cleanup(server.Close)

	cfg := config.LocalConfig{
		APIURL:       server.URL,
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"task", "create", "-t", "Write docs", "--issue-prefix", "ABC", "--tags", "dark"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "id=ABC-1")
	require.Contains(t, stdout.String(), "issue_prefix=ABC")
}

func TestRunAdminInitCallsServer(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "secret", nil
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	var got domain.AdminInitRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/admin/init", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":       "usr_1",
			"username":      got.Username,
			"access_key":    "key_1",
			"access_secret": "sec_1",
		})
	}))
	t.Cleanup(server.Close)

	code := Run([]string{"admin", "init", "-u", "alice", "--url", server.URL}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, "alice", got.Username)
	require.Equal(t, "secret", got.Password)
	require.Contains(t, stdout.String(), "access_key=key_1")
}

func TestRunAdminBootstrapWritesLocalConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		return "secret", nil
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	var got domain.AdminBootstrapRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/admin/bootstrap", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workspace_id":  "ws_1",
			"project_id":    "prj_1",
			"access_key":    "key_1",
			"access_secret": "sec_1",
		})
	}))
	t.Cleanup(server.Close)

	code := Run([]string{"admin", "bootstrap", "-u", "alice", "--url", server.URL}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, "alice", got.Username)
	require.Equal(t, "secret", got.Password)
	require.Equal(t, "phatodo", got.WorkspaceName)
	require.Equal(t, "phatodo", got.ProjectName)
	configPath := filepath.Join(workdir, ".phatodo", "config.json")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Contains(t, string(data), `"project_id": "prj_1"`)
}
