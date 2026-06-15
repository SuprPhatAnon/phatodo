package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

type fakeAPIClient struct {
	listProjectConfigFn  func(context.Context, string) ([]ProjectConfigItem, error)
	getProjectConfigFn   func(context.Context, string, string) (ProjectConfigItem, error)
	setProjectConfigFn   func(context.Context, string, string, string) (ProjectConfigItem, error)
	unsetProjectConfigFn func(context.Context, string, string) (ProjectConfigItem, error)
	createTaskFn         func(context.Context, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error)
	listTasksFn          func(context.Context, string, string, string) (domain.TaskListResponse, error)
	listReadyTasksFn     func(context.Context, string, string) (domain.ReadyListResponse, error)
	initAdminFn          func(context.Context, domain.AdminInitRequest) (domain.AdminInitResponse, error)
	bootstrapAdminFn     func(context.Context, domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error)
}

func (f *fakeAPIClient) ListProjectConfig(ctx context.Context, projectID string) ([]ProjectConfigItem, error) {
	return f.listProjectConfigFn(ctx, projectID)
}

func (f *fakeAPIClient) GetProjectConfig(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
	return f.getProjectConfigFn(ctx, projectID, key)
}

func (f *fakeAPIClient) SetProjectConfig(ctx context.Context, projectID, key, value string) (ProjectConfigItem, error) {
	return f.setProjectConfigFn(ctx, projectID, key, value)
}

func (f *fakeAPIClient) UnsetProjectConfig(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
	return f.unsetProjectConfigFn(ctx, projectID, key)
}

func (f *fakeAPIClient) CreateTask(ctx context.Context, projectID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
	return f.createTaskFn(ctx, projectID, req)
}

func (f *fakeAPIClient) ListTasks(ctx context.Context, projectID, status, epicID string) (domain.TaskListResponse, error) {
	return f.listTasksFn(ctx, projectID, status, epicID)
}

func (f *fakeAPIClient) ListReadyTasks(ctx context.Context, projectID, epicID string) (domain.ReadyListResponse, error) {
	return f.listReadyTasksFn(ctx, projectID, epicID)
}

func (f *fakeAPIClient) InitAdmin(ctx context.Context, req domain.AdminInitRequest) (domain.AdminInitResponse, error) {
	return f.initAdminFn(ctx, req)
}

func (f *fakeAPIClient) BootstrapAdmin(ctx context.Context, req domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
	return f.bootstrapAdminFn(ctx, req)
}

func TestRunTaskListCallsServer(t *testing.T) {
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		require.Equal(t, "default", cfg.ProjectID)
		return &fakeAPIClient{
			listTasksFn: func(ctx context.Context, projectID, status, epicID string) (domain.TaskListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "in_progress", status)
				require.Equal(t, "epic-1", epicID)
				return domain.TaskListResponse{
					ProjectID: "default",
					Items: []domain.TaskListItem{
						{
							ID:       "ABC-1",
							Title:    "Write docs",
							Status:   domain.StatusInProgress,
							Priority: domain.PriorityMedium,
							EpicID:   "epic-1",
						},
					},
				}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"task", "list", "--status", "in_progress", "--epic", "epic-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "tasks[1]:")
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "status: in_progress")
	require.Contains(t, stdout.String(), "epicId: epic-1")
}

func TestRunReadyCallsServer(t *testing.T) {
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		require.Equal(t, "default", cfg.ProjectID)
		return &fakeAPIClient{
			listReadyTasksFn: func(ctx context.Context, projectID, epicID string) (domain.ReadyListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "epic-1", epicID)
				return domain.ReadyListResponse{
					ProjectID: "default",
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
				}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"ready", "--epic", "epic-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "ready[1]:")
	require.Contains(t, stdout.String(), "- id: CORE-1")
	require.Contains(t, stdout.String(), "dependents[1]{id,title,status,priority}:")
	require.Contains(t, stdout.String(), "CORE-5,Backups,todo,1")
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
	require.Contains(t, stdout.String(), "- config_path:")
	require.Contains(t, stdout.String(), "project_id: default")

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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		require.Equal(t, "default", cfg.ProjectID)
		return &fakeAPIClient{
			listProjectConfigFn: func(ctx context.Context, projectID string) ([]ProjectConfigItem, error) {
				require.Equal(t, "default", projectID)
				return []ProjectConfigItem{{Key: "theme", Value: "dark"}}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "list"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- theme: dark")
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			setProjectConfigFn: func(ctx context.Context, projectID, key, value string) (ProjectConfigItem, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "theme", key)
				require.Equal(t, "dark", value)
				return ProjectConfigItem{Key: "theme", Value: "dark"}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "set", "theme", "dark"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- theme: dark")
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			getProjectConfigFn: func(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "theme", key)
				return ProjectConfigItem{Key: "theme", Value: "dark"}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "get", "theme"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- theme: dark")
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			unsetProjectConfigFn: func(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "theme", key)
				return ProjectConfigItem{Key: "theme", Value: "dark"}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"config", "unset", "theme"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- theme: dark")
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			createTaskFn: func(ctx context.Context, projectID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "Write docs", req.Title)
				require.Equal(t, "ABC", req.IssuePrefix)
				require.Equal(t, []string{"dark"}, req.Tags)
				return domain.TaskCreateResponse{
					ID:          "ABC-1",
					IssuePrefix: "ABC",
					Title:       "Write docs",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
				}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"task", "create", "-t", "Write docs", "--issue-prefix", "ABC", "--tags", "dark"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "issue_prefix: ABC")
	require.Contains(t, stdout.String(), "title: \"Write docs\"")
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
	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			initAdminFn: func(ctx context.Context, req domain.AdminInitRequest) (domain.AdminInitResponse, error) {
				got = req
				return domain.AdminInitResponse{
					UserID:       "usr_1",
					Username:     req.Username,
					AccessKey:    "key_1",
					AccessSecret: "sec_1",
				}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	code := Run([]string{"admin", "init", "-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, "alice", got.Username)
	require.Equal(t, "secret", got.Password)
	require.Contains(t, stdout.String(), "- user_id: usr_1")
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
	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return &fakeAPIClient{
			bootstrapAdminFn: func(ctx context.Context, req domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
				got = req
				return domain.AdminBootstrapResponse{
					WorkspaceID:  "ws_1",
					ProjectID:    "prj_1",
					AccessKey:    "key_1",
					AccessSecret: "sec_1",
				}, nil
			},
		}, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	code := Run([]string{"admin", "bootstrap", "-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, "alice", got.Username)
	require.Equal(t, "secret", got.Password)
	require.Equal(t, "phatodo", got.WorkspaceName)
	require.Equal(t, "phatodo", got.ProjectName)
	configPath := filepath.Join(workdir, ".phatodo", "config.json")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Contains(t, string(data), `"project_id": "prj_1"`)
	require.Contains(t, stdout.String(), "- workspace_id: ws_1")
	require.Contains(t, stdout.String(), "project_id: prj_1")
	require.Contains(t, stdout.String(), "config_path:")
}
