package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

type fakeAPIClient struct {
	listProjectConfigFn  func(context.Context, string) ([]ProjectConfigItem, error)
	getProjectConfigFn   func(context.Context, string, string) (ProjectConfigItem, error)
	setProjectConfigFn   func(context.Context, string, string, string) (ProjectConfigItem, error)
	unsetProjectConfigFn func(context.Context, string, string) (ProjectConfigItem, error)
	createEpicFn         func(context.Context, string, domain.EpicCreateRequest) (domain.Epic, error)
	listEpicsFn          func(context.Context, string, string, int) (domain.EpicListResponse, error)
	getEpicFn            func(context.Context, string, string) (domain.Epic, error)
	updateEpicFn         func(context.Context, string, string, domain.EpicUpdateRequest) (domain.Epic, error)
	completeEpicFn       func(context.Context, string, string) (domain.Epic, error)
	deleteEpicFn         func(context.Context, string, string) (domain.Epic, error)
	createTaskFn         func(context.Context, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error)
	createSubtaskFn      func(context.Context, string, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error)
	getTaskFn            func(context.Context, string, string) (domain.TaskDetail, error)
	updateTaskFn         func(context.Context, string, string, domain.TaskUpdateRequest) (domain.TaskDetail, error)
	deleteTaskFn         func(context.Context, string, string) (domain.TaskDetail, error)
	listTasksFn          func(context.Context, string, string, string, int) (domain.TaskListResponse, error)
	listSubtasksFn       func(context.Context, string, string, int) (domain.TaskListResponse, error)
	listCommentsFn       func(context.Context, string, string) (domain.CommentListResponse, error)
	addCommentFn         func(context.Context, string, string, domain.CommentCreateRequest) (domain.Comment, error)
	updateCommentFn      func(context.Context, string, string, domain.CommentUpdateRequest) (domain.Comment, error)
	deleteCommentFn      func(context.Context, string, string) (domain.Comment, error)
	listDependenciesFn   func(context.Context, string, string) (domain.DependencyListResponse, error)
	addDependencyFn      func(context.Context, string, string, string) (domain.Dependency, error)
	removeDependencyFn   func(context.Context, string, string, string) (domain.Dependency, error)
	listLocksFn          func(context.Context, string, string, string, bool) (domain.LockListResponse, error)
	acquireLockFn        func(context.Context, string, domain.LockAcquireRequest) (domain.WorkItemLock, error)
	releaseLockFn        func(context.Context, string, string) (domain.WorkItemLock, error)
	searchFn             func(context.Context, string, string, string, string, int) (domain.SearchResponse, error)
	historyFn            func(context.Context, string, string, string, string, string, int) (domain.HistoryResponse, error)
	listUnifiedFn        func(context.Context, string, string, string, string, string, int) (domain.ListResponse, error)
	listReadyTasksFn     func(context.Context, string, string, int) (domain.ReadyListResponse, error)
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

func (f *fakeAPIClient) CreateEpic(ctx context.Context, projectID string, req domain.EpicCreateRequest) (domain.Epic, error) {
	if f.createEpicFn == nil {
		return domain.Epic{}, nil
	}
	return f.createEpicFn(ctx, projectID, req)
}

func (f *fakeAPIClient) ListEpics(ctx context.Context, projectID, status string, limit int) (domain.EpicListResponse, error) {
	if f.listEpicsFn == nil {
		return domain.EpicListResponse{}, nil
	}
	return f.listEpicsFn(ctx, projectID, status, limit)
}

func (f *fakeAPIClient) GetEpic(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
	if f.getEpicFn == nil {
		return domain.Epic{}, nil
	}
	return f.getEpicFn(ctx, projectID, epicID)
}

func (f *fakeAPIClient) UpdateEpic(ctx context.Context, projectID, epicID string, req domain.EpicUpdateRequest) (domain.Epic, error) {
	if f.updateEpicFn == nil {
		return domain.Epic{}, nil
	}
	return f.updateEpicFn(ctx, projectID, epicID, req)
}

func (f *fakeAPIClient) CompleteEpic(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
	if f.completeEpicFn == nil {
		return domain.Epic{}, nil
	}
	return f.completeEpicFn(ctx, projectID, epicID)
}

func (f *fakeAPIClient) DeleteEpic(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
	if f.deleteEpicFn == nil {
		return domain.Epic{}, nil
	}
	return f.deleteEpicFn(ctx, projectID, epicID)
}

func (f *fakeAPIClient) CreateTask(ctx context.Context, projectID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
	return f.createTaskFn(ctx, projectID, req)
}

func (f *fakeAPIClient) CreateSubtask(ctx context.Context, projectID, taskID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
	return f.createSubtaskFn(ctx, projectID, taskID, req)
}

func (f *fakeAPIClient) GetTask(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
	return f.getTaskFn(ctx, projectID, taskID)
}

func (f *fakeAPIClient) UpdateTask(ctx context.Context, projectID, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
	return f.updateTaskFn(ctx, projectID, taskID, req)
}

func (f *fakeAPIClient) DeleteTask(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
	return f.deleteTaskFn(ctx, projectID, taskID)
}

func (f *fakeAPIClient) ListTasks(ctx context.Context, projectID, status, epicID string, limit int) (domain.TaskListResponse, error) {
	return f.listTasksFn(ctx, projectID, status, epicID, limit)
}

func (f *fakeAPIClient) ListSubtasks(ctx context.Context, projectID, taskID string, limit int) (domain.TaskListResponse, error) {
	return f.listSubtasksFn(ctx, projectID, taskID, limit)
}

func (f *fakeAPIClient) ListComments(ctx context.Context, projectID, taskID string) (domain.CommentListResponse, error) {
	return f.listCommentsFn(ctx, projectID, taskID)
}

func (f *fakeAPIClient) AddComment(ctx context.Context, projectID, taskID string, req domain.CommentCreateRequest) (domain.Comment, error) {
	return f.addCommentFn(ctx, projectID, taskID, req)
}

func (f *fakeAPIClient) UpdateComment(ctx context.Context, projectID, commentID string, req domain.CommentUpdateRequest) (domain.Comment, error) {
	return f.updateCommentFn(ctx, projectID, commentID, req)
}

func (f *fakeAPIClient) DeleteComment(ctx context.Context, projectID, commentID string) (domain.Comment, error) {
	return f.deleteCommentFn(ctx, projectID, commentID)
}

func (f *fakeAPIClient) ListDependencies(ctx context.Context, projectID, taskID string) (domain.DependencyListResponse, error) {
	return f.listDependenciesFn(ctx, projectID, taskID)
}

func (f *fakeAPIClient) AddDependency(ctx context.Context, projectID, taskID, dependsOnID string) (domain.Dependency, error) {
	return f.addDependencyFn(ctx, projectID, taskID, dependsOnID)
}

func (f *fakeAPIClient) RemoveDependency(ctx context.Context, projectID, taskID, dependsOnID string) (domain.Dependency, error) {
	return f.removeDependencyFn(ctx, projectID, taskID, dependsOnID)
}

func (f *fakeAPIClient) ListLocks(ctx context.Context, projectID, entityTypes, entityID string, active bool) (domain.LockListResponse, error) {
	if f.listLocksFn == nil {
		return domain.LockListResponse{}, nil
	}
	return f.listLocksFn(ctx, projectID, entityTypes, entityID, active)
}

func (f *fakeAPIClient) AcquireLock(ctx context.Context, projectID string, req domain.LockAcquireRequest) (domain.WorkItemLock, error) {
	if f.acquireLockFn == nil {
		return domain.WorkItemLock{}, nil
	}
	return f.acquireLockFn(ctx, projectID, req)
}

func (f *fakeAPIClient) ReleaseLock(ctx context.Context, projectID, lockID string) (domain.WorkItemLock, error) {
	if f.releaseLockFn == nil {
		return domain.WorkItemLock{}, nil
	}
	return f.releaseLockFn(ctx, projectID, lockID)
}

func (f *fakeAPIClient) Search(ctx context.Context, projectID, query, entityType, status string, limit int) (domain.SearchResponse, error) {
	return f.searchFn(ctx, projectID, query, entityType, status, limit)
}

func (f *fakeAPIClient) History(ctx context.Context, projectID, entityID, entityType, action, since string, limit int) (domain.HistoryResponse, error) {
	return f.historyFn(ctx, projectID, entityID, entityType, action, since, limit)
}

func (f *fakeAPIClient) ListUnified(ctx context.Context, projectID, entityType, status, priority, sortSpec string, limit int) (domain.ListResponse, error) {
	return f.listUnifiedFn(ctx, projectID, entityType, status, priority, sortSpec, limit)
}

func (f *fakeAPIClient) ListReadyTasks(ctx context.Context, projectID, epicID string, limit int) (domain.ReadyListResponse, error) {
	return f.listReadyTasksFn(ctx, projectID, epicID, limit)
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
			listTasksFn: func(ctx context.Context, projectID, status, epicID string, limit int) (domain.TaskListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "in_progress", status)
				require.Equal(t, "epic-1", epicID)
				require.Equal(t, 20, limit)
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

	code := Run([]string{"--toon", "task", "list", "--status", "in_progress", "--epic", "epic-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "tasks[1]:")
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "status: in_progress")
	require.Contains(t, stdout.String(), "epicId: epic-1")
}

func TestRunEpicListCallsServer(t *testing.T) {
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
			listEpicsFn: func(ctx context.Context, projectID, status string, limit int) (domain.EpicListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "in_progress", status)
				require.Equal(t, 20, limit)
				return domain.EpicListResponse{
					ProjectID: "default",
					Items: []domain.Epic{
						{
							ID:        "EPIC-1",
							Title:     "Track auth",
							Status:    domain.StatusInProgress,
							Priority:  domain.PriorityCritical,
							CreatedAt: time.Date(2026, 6, 9, 2, 13, 2, 0, time.UTC),
							UpdatedAt: time.Date(2026, 6, 13, 15, 1, 51, 0, time.UTC),
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

	code := Run([]string{"--toon", "epic", "list", "--status", "in_progress"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "epics[1]:")
	require.Contains(t, stdout.String(), "- id: EPIC-1")
	require.Contains(t, stdout.String(), "status: in_progress")
}

func TestRunLockAcquireCallsServer(t *testing.T) {
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
			acquireLockFn: func(ctx context.Context, projectID string, req domain.LockAcquireRequest) (domain.WorkItemLock, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "task", req.EntityType)
				require.Equal(t, "ABC-1", req.EntityID)
				require.Equal(t, "editing", req.Reason)
				require.Equal(t, "30m", req.TTL)
				return domain.WorkItemLock{
					ID:         "LOCK-1",
					EntityType: "task",
					EntityID:   "ABC-1",
					LockedBy:   "user-1",
					Reason:     "editing",
					LeasedAt:   time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
					ExpiresAt:  time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC),
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

	code := Run([]string{"--toon", "lock", "acquire", "task", "ABC-1", "--reason", "editing", "--expires", "30m"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: LOCK-1")
	require.Contains(t, stdout.String(), "entityType: task")
	require.Contains(t, stdout.String(), "entityId: ABC-1")
}

func TestRunLockListCallsServer(t *testing.T) {
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
			listLocksFn: func(ctx context.Context, projectID, entityTypes, entityID string, active bool) (domain.LockListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "epic,task", entityTypes)
				require.Equal(t, "EPIC-1", entityID)
				require.True(t, active)
				return domain.LockListResponse{
					ProjectID: "default",
					Items: []domain.WorkItemLock{
						{
							ID:         "LOCK-2",
							EntityType: "epic",
							EntityID:   "EPIC-1",
							LockedBy:   "user-1",
							Reason:     "planning",
							LeasedAt:   time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
							ExpiresAt:  time.Date(2026, 6, 15, 13, 0, 0, 0, time.UTC),
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

	code := Run([]string{"--toon", "lock", "list", "--type", "epic,task", "--entity", "EPIC-1", "--active"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "locks[1]:")
	require.Contains(t, stdout.String(), "- id: LOCK-2")
	require.Contains(t, stdout.String(), "entityType: epic")
}

func TestRunLockReleaseCallsServer(t *testing.T) {
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
			releaseLockFn: func(ctx context.Context, projectID, lockID string) (domain.WorkItemLock, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "LOCK-3", lockID)
				return domain.WorkItemLock{
					ID:         "LOCK-3",
					EntityType: "subtask",
					EntityID:   "ABC-2",
					LockedBy:   "user-1",
					Reason:     "review",
					LeasedAt:   time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
					ExpiresAt:  time.Date(2026, 6, 15, 13, 0, 0, 0, time.UTC),
					ReleasedAt: time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC),
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

	code := Run([]string{"--toon", "lock", "release", "LOCK-3"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: LOCK-3")
	require.Contains(t, stdout.String(), "releasedAt:")
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
			listReadyTasksFn: func(ctx context.Context, projectID, epicID string, limit int) (domain.ReadyListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "epic-1", epicID)
				require.Equal(t, 20, limit)
				return domain.ReadyListResponse{
					ProjectID: "default",
					Items: []domain.ReadyListItem{
						{
							ID:          "CORE-1",
							Title:       "Health endpoints",
							Description: "Add readiness and liveness checks",
							Status:      domain.StatusTodo,
							Priority:    domain.PriorityHigh,
							EpicID:      "epic-1",
							Tags:        []string{"infra", "api"},
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
	require.Contains(t, stdout.String(), "1 ready task(s)")
	require.Contains(t, stdout.String(), "CORE-1 | P1 | Health endpoints (epic-1) [infra,api]")
	require.Contains(t, stdout.String(), "  -> unblocks CORE-5 | todo | P1 | Backups [infra]")
}

func TestRunReadyCallsServerTOON(t *testing.T) {
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
			listReadyTasksFn: func(ctx context.Context, projectID, epicID string, limit int) (domain.ReadyListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "epic-1", epicID)
				require.Equal(t, 20, limit)
				return domain.ReadyListResponse{
					ProjectID: "default",
					Items: []domain.ReadyListItem{
						{
							ID:          "CORE-1",
							Title:       "Health endpoints",
							Description: "Add readiness and liveness checks",
							Status:      domain.StatusTodo,
							Priority:    domain.PriorityHigh,
							EpicID:      "epic-1",
							Tags:        []string{"infra", "api"},
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

	code := Run([]string{"--toon", "ready", "--epic", "epic-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "ready[1]:")
	require.Contains(t, stdout.String(), "- id: CORE-1")
	require.Contains(t, stdout.String(), "  description: \"Add readiness and liveness checks\"")
	require.Contains(t, stdout.String(), "dependents[1]{id,title,status,priority}:")
	require.Contains(t, stdout.String(), "CORE-5,Backups,todo,1")
}

func TestRunSearchCallsServer(t *testing.T) {
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
			searchFn: func(ctx context.Context, projectID, query, entityType, status string, limit int) (domain.SearchResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "auth bug", query)
				require.Equal(t, "task", entityType)
				require.Equal(t, "todo", status)
				require.Equal(t, 10, limit)
				return domain.SearchResponse{
					ProjectID: "default",
					Query:     query,
					Items: []domain.SearchItem{
						{
							EntityType: "task",
							ID:         "ABC-1",
							Title:      "Fix auth bug",
							Status:     domain.StatusTodo,
							Priority:   domain.PriorityHigh,
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

	code := Run([]string{"--toon", "search", "auth bug", "--type", "task", "--status", "todo", "--limit", "10"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "search[1]:")
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "entityType: task")
	require.Contains(t, stdout.String(), `title: "Fix auth bug"`)
}

func TestRunHistoryCallsServer(t *testing.T) {
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
			historyFn: func(ctx context.Context, projectID, entityID, entityType, action, since string, limit int) (domain.HistoryResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", entityID)
				require.Equal(t, "task", entityType)
				require.Equal(t, "update", action)
				require.Equal(t, "2025-01-01", since)
				require.Equal(t, 5, limit)
				return domain.HistoryResponse{
					ProjectID: "default",
					Items: []domain.HistoryEvent{
						{
							ID:         42,
							EntityType: "task",
							EntityID:   "ABC-1",
							Action:     "update",
							ActorLabel: "alice",
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

	code := Run([]string{"--toon", "history", "--entity", "ABC-1", "--type", "task", "--action", "update", "--since", "2025-01-01", "--limit", "5"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "history[1]:")
	require.Contains(t, stdout.String(), "- id: 42")
	require.Contains(t, stdout.String(), "entityType: task")
	require.Contains(t, stdout.String(), "actorLabel: alice")
}

func TestRunListCallsServer(t *testing.T) {
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
			listUnifiedFn: func(ctx context.Context, projectID, entityType, status, priority, sortSpec string, limit int) (domain.ListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "epic,task", entityType)
				require.Equal(t, "todo", status)
				require.Equal(t, "0,1", priority)
				require.Equal(t, "priority:asc,created:desc", sortSpec)
				require.Equal(t, 2, limit)
				return domain.ListResponse{
					ProjectID: "default",
					Items: []domain.UnifiedListItem{
						{
							EntityType: "epic",
							ID:         "EPIC-1",
							Title:      "Track auth",
							Status:     domain.StatusTodo,
							Priority:   domain.PriorityCritical,
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

	code := Run([]string{"--toon", "list", "--type", "epic,task", "--status", "todo", "--priority", "0,1", "--sort", "priority:asc,created:desc", "--limit", "2"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "list[1]:")
	require.Contains(t, stdout.String(), "- id: EPIC-1")
	require.Contains(t, stdout.String(), "entityType: epic")
	require.Contains(t, stdout.String(), `title: "Track auth"`)
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
	require.Contains(t, stdout.String(), "config_path:")
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

func TestRunTopLevelHelpIsFlat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.NotContains(t, stdout.String(), "Command groups:")
	require.Contains(t, stdout.String(), "ptodo quickstart")
	require.Contains(t, stdout.String(), "ptodo epic create")
	require.Contains(t, stdout.String(), "ptodo task update")
}

func TestRunQuickstart(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"quickstart"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "# Phatodo Quickstart")
	require.Contains(t, stdout.String(), "ptodo admin bootstrap")
	require.Contains(t, stdout.String(), "ptodo --toon ready")
	require.Contains(t, stdout.String(), "Good usage:")
	require.Contains(t, stdout.String(), "Bad usage:")
	require.Contains(t, stdout.String(), `ptodo task update ABC-1 --criteria-json '["docs written","tests passing"]'`)
}

func TestRunQuickstartHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"quickstart", "--help"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "Usage:")
	require.Contains(t, stdout.String(), "ptodo quickstart")
	require.Contains(t, stdout.String(), "Show quick reference for AI agents")
}

func TestRunTaskFamilyHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"task", "--help"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "ptodo task")
	require.Contains(t, stdout.String(), "ptodo task create")
	require.Contains(t, stdout.String(), "ptodo task delete")
}

func TestRunTaskCreateHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"task", "create", "--help"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "Usage:")
	require.Contains(t, stdout.String(), "ptodo task create -t \"Title\"")
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

	code := Run([]string{"--toon", "config", "list"}, &stdout, &stderr)
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

	code := Run([]string{"--toon", "config", "set", "theme", "dark"}, &stdout, &stderr)
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

	code := Run([]string{"--toon", "config", "get", "theme"}, &stdout, &stderr)
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

	code := Run([]string{"--toon", "config", "unset", "theme"}, &stdout, &stderr)
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
					Kind:        domain.TaskKindBug,
					RootCause:   "missing stack traces",
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

	code := Run([]string{"--toon", "task", "create", "-t", "Write docs", "--issue-prefix", "ABC", "--tags", "dark"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "issue_prefix: ABC")
	require.Contains(t, stdout.String(), "title: \"Write docs\"")
}

func TestRunTaskCreateAcceptsPrefixAlias(t *testing.T) {
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
				require.Equal(t, "ABC", req.IssuePrefix)
				return domain.TaskCreateResponse{
					ID:          "ABC-1",
					IssuePrefix: "ABC",
					Title:       "Write docs",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
					Kind:        domain.TaskKindBug,
					RootCause:   "timeout race",
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

	code := Run([]string{"--toon", "task", "create", "-t", "Write docs", "--prefix", "ABC"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
}

func TestRunTaskCreateRejectsBugWithoutRootCause(t *testing.T) {
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

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"--toon", "task", "create", "-t", "Write docs", "--issue-prefix", "ABC", "--kind", "bug"}, &stdout, &stderr)
	require.Equal(t, 2, code, stderr.String())
	require.Contains(t, stderr.String(), "root-cause-analysis")
}

func TestRunTaskCreateSendsKindAndRootCause(t *testing.T) {
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
				require.Equal(t, domain.TaskKindBug, req.Kind)
				require.Equal(t, "missing stack traces", req.RootCauseAnalysis)
				return domain.TaskCreateResponse{
					ID:          "ABC-1",
					IssuePrefix: "ABC",
					Title:       "Write docs",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
					Kind:        domain.TaskKindBug,
					RootCause:   "missing stack traces",
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

	code := Run([]string{"--toon", "task", "create", "-t", "Write docs", "--issue-prefix", "ABC", "--kind", "bug", "--root-cause-analysis", "missing stack traces"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "kind: bug")
	require.Contains(t, stdout.String(), `rootCauseAnalysis: "missing stack traces"`)
}

func TestRunTaskCreateSendsFeatureKind(t *testing.T) {
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
				require.Equal(t, domain.TaskKindFeature, req.Kind)
				require.Empty(t, req.RootCauseAnalysis)
				return domain.TaskCreateResponse{
					ID:          "ABC-1",
					IssuePrefix: "ABC",
					Title:       "Add import",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
					Kind:        domain.TaskKindFeature,
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

	code := Run([]string{"--toon", "task", "create", "-t", "Add import", "--issue-prefix", "ABC", "--kind", "feature"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "kind: feature")
}

func TestRunSubtaskCreateCallsServer(t *testing.T) {
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
			createSubtaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, "Write docs", req.Title)
				require.Empty(t, req.IssuePrefix)
				require.Equal(t, []string{"important"}, req.AcceptanceCriteria)
				return domain.TaskCreateResponse{
					ID:          "ABC-2",
					IssuePrefix: "ABC",
					Title:       "Write docs",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
					Kind:        domain.TaskKindBug,
					RootCause:   "timeout race",
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

	code := Run([]string{"--toon", "subtask", "create", "ABC-1", "-t", "Write docs", "--criteria-json", `["important"]`}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-2")
	require.Contains(t, stdout.String(), "issue_prefix: ABC")
	require.Contains(t, stdout.String(), "title: \"Write docs\"")
}

func TestRunSubtaskCreateRejectsBugWithoutRootCause(t *testing.T) {
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

	cfg := config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	}
	_, err = config.WriteLocal(workdir, cfg)
	require.NoError(t, err)

	code := Run([]string{"--toon", "subtask", "create", "ABC-1", "-t", "Write docs", "--kind", "bug"}, &stdout, &stderr)
	require.Equal(t, 2, code, stderr.String())
	require.Contains(t, stderr.String(), "root-cause-analysis")
}

func TestRunSubtaskCreateSendsKindAndRootCause(t *testing.T) {
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
			createSubtaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, domain.TaskKindBug, req.Kind)
				require.Equal(t, "timeout race", req.RootCauseAnalysis)
				return domain.TaskCreateResponse{
					ID:          "ABC-2",
					IssuePrefix: "ABC",
					Title:       "Write docs",
					Status:      domain.StatusTodo,
					Priority:    domain.PriorityMedium,
					ProjectID:   "default",
					WorkspaceID: "default",
					Kind:        domain.TaskKindBug,
					RootCause:   "timeout race",
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

	code := Run([]string{"--toon", "subtask", "create", "ABC-1", "-t", "Write docs", "--kind", "bug", "--root-cause", "timeout race"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "kind: bug")
	require.Contains(t, stdout.String(), `rootCauseAnalysis: "timeout race"`)
}

func TestRunTaskShowUsesFakeClient(t *testing.T) {
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
			getTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				return domain.TaskDetail{
					ID:         "ABC-1",
					Title:      "Write docs",
					Status:     domain.StatusInProgress,
					Priority:   domain.PriorityMedium,
					EpicID:     "epic-1",
					AssignedTo: "usr_1",
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

	code := Run([]string{"--toon", "task", "show", "ABC-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "epicId: epic-1")
	require.Contains(t, stdout.String(), "assignedTo: usr_1")
}

func TestRunSubtaskListCallsServer(t *testing.T) {
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
			listSubtasksFn: func(ctx context.Context, projectID, taskID string, limit int) (domain.TaskListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, 20, limit)
				return domain.TaskListResponse{
					ProjectID: "default",
					Items: []domain.TaskListItem{
						{
							ID:           "ABC-2",
							Title:        "Write docs",
							Status:       domain.StatusTodo,
							Priority:     domain.PriorityMedium,
							ParentTaskID: "ABC-1",
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

	code := Run([]string{"--toon", "subtask", "list", "ABC-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "subtasks[1]:")
	require.Contains(t, stdout.String(), "- id: ABC-2")
	require.Contains(t, stdout.String(), "parentTaskId: ABC-1")
}

func TestRunTaskUpdateUsesFakeClient(t *testing.T) {
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
			updateTaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.NotNil(t, req.Status)
				require.Equal(t, domain.StatusInProgress, *req.Status)
				require.NotNil(t, req.Title)
				return domain.TaskDetail{
					ID:       "ABC-1",
					Title:    "Updated docs",
					Status:   domain.StatusInProgress,
					Priority: domain.PriorityHigh,
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

	code := Run([]string{"--toon", "task", "update", "ABC-1", "-t", "Updated docs", "-s", "in_progress"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "title: \"Updated docs\"")
	require.Contains(t, stdout.String(), "status: in_progress")
}

func TestRunSubtaskUpdateUsesFakeClient(t *testing.T) {
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
			updateTaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-2", taskID)
				require.NotNil(t, req.Status)
				require.Equal(t, domain.StatusCompleted, *req.Status)
				require.NotNil(t, req.Title)
				return domain.TaskDetail{
					ID:     "ABC-2",
					Title:  "Updated subtask",
					Status: domain.StatusCompleted,
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

	code := Run([]string{"--toon", "subtask", "update", "ABC-2", "-t", "Updated subtask", "-s", "completed"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-2")
	require.Contains(t, stdout.String(), "title: \"Updated subtask\"")
	require.Contains(t, stdout.String(), "status: completed")
}

func TestRunTaskDeleteUsesFakeClient(t *testing.T) {
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
			deleteTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				return domain.TaskDetail{
					ID:     "ABC-1",
					Title:  "Write docs",
					Status: domain.StatusArchived,
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

	code := Run([]string{"--toon", "task", "delete", "ABC-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-1")
	require.Contains(t, stdout.String(), "status: archived")
}

func TestRunSubtaskDeleteUsesFakeClient(t *testing.T) {
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
			deleteTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-2", taskID)
				return domain.TaskDetail{
					ID:     "ABC-2",
					Title:  "Write docs",
					Status: domain.StatusArchived,
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

	code := Run([]string{"--toon", "subtask", "delete", "ABC-2"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: ABC-2")
	require.Contains(t, stdout.String(), "status: archived")
}

func TestRunCommentAddCallsServer(t *testing.T) {
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
			addCommentFn: func(ctx context.Context, projectID, taskID string, req domain.CommentCreateRequest) (domain.Comment, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, "agent", req.Author)
				require.Equal(t, "summary", req.Kind)
				require.Equal(t, "Done", req.Content)
				return domain.Comment{
					ID:      "cmt-1",
					Author:  "agent",
					Kind:    "summary",
					Content: "Done",
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

	code := Run([]string{"--toon", "comment", "add", "ABC-1", "-a", "agent", "-c", "Done", "-k", "summary"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: cmt-1")
	require.Contains(t, stdout.String(), "author: agent")
	require.Contains(t, stdout.String(), "kind: summary")
	require.Contains(t, stdout.String(), "content: Done")
}

func TestRunCommentListCallsServer(t *testing.T) {
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
			listCommentsFn: func(ctx context.Context, projectID, taskID string) (domain.CommentListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				return domain.CommentListResponse{
					ProjectID: "default",
					TaskID:    "ABC-1",
					Items: []domain.Comment{
						{
							ID:      "cmt-1",
							Author:  "agent",
							Kind:    "analysis",
							Content: "Working notes",
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

	code := Run([]string{"--toon", "comment", "list", "ABC-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "comments[1]:")
	require.Contains(t, stdout.String(), "- id: cmt-1")
	require.Contains(t, stdout.String(), "kind: analysis")
}

func TestRunCommentUpdateCallsServer(t *testing.T) {
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
			updateCommentFn: func(ctx context.Context, projectID, commentID string, req domain.CommentUpdateRequest) (domain.Comment, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "cmt-1", commentID)
				require.Equal(t, "Updated", req.Content)
				return domain.Comment{
					ID:      "cmt-1",
					Author:  "agent",
					Kind:    "comment",
					Content: "Updated",
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

	code := Run([]string{"--toon", "comment", "update", "cmt-1", "-c", "Updated"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: cmt-1")
	require.Contains(t, stdout.String(), "content: Updated")
}

func TestRunCommentDeleteCallsServer(t *testing.T) {
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
			deleteCommentFn: func(ctx context.Context, projectID, commentID string) (domain.Comment, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "cmt-1", commentID)
				return domain.Comment{
					ID:      "cmt-1",
					Author:  "agent",
					Kind:    "comment",
					Content: "Deleted",
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

	code := Run([]string{"--toon", "comment", "delete", "cmt-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: cmt-1")
	require.Contains(t, stdout.String(), "content: Deleted")
}

func TestRunDepAddCallsServer(t *testing.T) {
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
			addDependencyFn: func(ctx context.Context, projectID, taskID, dependsOnID string) (domain.Dependency, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, "ABC-2", dependsOnID)
				return domain.Dependency{
					ID:          "dep-1",
					TaskID:      "ABC-1",
					DependsOnID: "ABC-2",
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

	code := Run([]string{"--toon", "dep", "add", "ABC-1", "ABC-2"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: dep-1")
	require.Contains(t, stdout.String(), "taskId: ABC-1")
	require.Contains(t, stdout.String(), "dependsOnId: ABC-2")
}

func TestRunDepListCallsServer(t *testing.T) {
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
			listDependenciesFn: func(ctx context.Context, projectID, taskID string) (domain.DependencyListResponse, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				return domain.DependencyListResponse{
					ProjectID: "default",
					TaskID:    "ABC-1",
					Items: []domain.Dependency{
						{ID: "dep-1", TaskID: "ABC-1", DependsOnID: "ABC-2"},
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

	code := Run([]string{"--toon", "dep", "list", "ABC-1"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "dependencies[1]:")
	require.Contains(t, stdout.String(), "taskId: ABC-1")
	require.Contains(t, stdout.String(), "dependsOnId: ABC-2")
}

func TestRunDepRemoveCallsServer(t *testing.T) {
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
			removeDependencyFn: func(ctx context.Context, projectID, taskID, dependsOnID string) (domain.Dependency, error) {
				require.Equal(t, "default", projectID)
				require.Equal(t, "ABC-1", taskID)
				require.Equal(t, "ABC-2", dependsOnID)
				return domain.Dependency{
					ID:          "dep-1",
					TaskID:      "ABC-1",
					DependsOnID: "ABC-2",
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

	code := Run([]string{"--toon", "dep", "remove", "ABC-1", "ABC-2"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: dep-1")
	require.Contains(t, stdout.String(), "taskId: ABC-1")
	require.Contains(t, stdout.String(), "dependsOnId: ABC-2")
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

	code := Run([]string{"--toon", "admin", "init", "-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
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

	code := Run([]string{"--toon", "admin", "bootstrap", "-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
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

func TestRunAdminBootstrapAcceptsProjectAlias(t *testing.T) {
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

	code := Run([]string{"--toon", "admin", "bootstrap", "-u", "alice", "--url", "http://example.invalid", "--project", "renamed"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, "renamed", got.ProjectName)
	require.Equal(t, "phatodo", got.WorkspaceName)
}

func TestRunAdminBootstrapStopsWhenConfigExists(t *testing.T) {
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

	_, err = config.WriteLocal(workdir, config.DefaultLocalConfig())
	require.NoError(t, err)

	oldPrompt := readPasswordPrompt
	readPasswordPrompt = func(prompt string, _ io.Writer) (string, error) {
		t.Fatalf("password prompt should not be reached when config exists")
		return "", nil
	}
	t.Cleanup(func() { readPasswordPrompt = oldPrompt })

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		t.Fatalf("api client should not be created when config exists")
		return nil, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	code := Run([]string{"--toon", "admin", "bootstrap", "-u", "alice", "--url", "http://example.invalid"}, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "local config already exists:")
}
