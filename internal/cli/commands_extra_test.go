package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestRunConfigCommandsCallServer(t *testing.T) {
	client := &fakeAPIClient{
		listProjectConfigFn: func(ctx context.Context, projectID string) ([]ProjectConfigItem, error) {
			require.Equal(t, "default", projectID)
			return []ProjectConfigItem{{Key: "theme", Value: "dark"}}, nil
		},
		getProjectConfigFn: func(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "theme", key)
			return ProjectConfigItem{Key: key, Value: "dark"}, nil
		},
		setProjectConfigFn: func(ctx context.Context, projectID, key, value string) (ProjectConfigItem, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "theme", key)
			require.Equal(t, "light", value)
			return ProjectConfigItem{Key: key, Value: value}, nil
		},
		unsetProjectConfigFn: func(ctx context.Context, projectID, key string) (ProjectConfigItem, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "theme", key)
			return ProjectConfigItem{Key: key, Value: "light"}, nil
		},
	}
	stdout, stderr := withConfiguredCLI(t, client)

	code := Run([]string{"--toon", "config", "list"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "config", "get", "theme"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "dark")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "config", "set", "theme", "light"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "light")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "config", "unset", "theme"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "theme")
}

func TestRunTaskDetailMutationCommandsCallServer(t *testing.T) {
	client := &fakeAPIClient{
		getTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "TASK-1", taskID)
			return domain.TaskDetail{ID: taskID, Title: "Write tests", Status: domain.StatusTodo, Priority: domain.PriorityMedium}, nil
		},
		updateTaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "TASK-1", taskID)
			require.NotNil(t, req.Status)
			require.Equal(t, domain.StatusInProgress, *req.Status)
			return domain.TaskDetail{ID: taskID, Title: "Write tests", Status: *req.Status, Priority: domain.PriorityHigh}, nil
		},
		deleteTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "TASK-1", taskID)
			return domain.TaskDetail{ID: taskID, Title: "Write tests", Status: domain.StatusArchived, Priority: domain.PriorityMedium}, nil
		},
	}
	stdout, stderr := withConfiguredCLI(t, client)

	code := Run([]string{"--toon", "task", "show", "TASK-1"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: TASK-1")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "task", "update", "TASK-1", "-s", "in_progress", "-p", "1", "--changed-files-json", `["internal/cli/commands.go"]`}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: in_progress")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "task", "delete", "TASK-1"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: archived")
}

func TestRunSubtaskUpdateAndDeleteCallServer(t *testing.T) {
	client := &fakeAPIClient{
		updateTaskFn: func(ctx context.Context, projectID, taskID string, req domain.TaskUpdateRequest) (domain.TaskDetail, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "TASK-2", taskID)
			require.NotNil(t, req.Title)
			return domain.TaskDetail{ID: taskID, ParentTaskID: "TASK-1", Title: *req.Title, Status: domain.StatusTodo, Priority: domain.PriorityMedium}, nil
		},
		deleteTaskFn: func(ctx context.Context, projectID, taskID string) (domain.TaskDetail, error) {
			require.Equal(t, "TASK-2", taskID)
			return domain.TaskDetail{ID: taskID, ParentTaskID: "TASK-1", Title: "Subtask", Status: domain.StatusArchived, Priority: domain.PriorityMedium}, nil
		},
	}
	stdout, stderr := withConfiguredCLI(t, client)

	code := Run([]string{"--toon", "subtask", "update", "TASK-2", "-t", "Updated subtask"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "Updated subtask")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "subtask", "delete", "TASK-2"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: archived")
}

func TestRunCommandsFailWhenLocalConfigMissing(t *testing.T) {
	workdir := filepath.Join(t.TempDir(), "phatodo")
	require.NoError(t, os.MkdirAll(workdir, 0o755))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldwd))
	})
	require.NoError(t, os.Chdir(workdir))

	cases := [][]string{
		{"--toon", "config", "list"},
		{"--toon", "config", "get", "theme"},
		{"--toon", "config", "set", "theme", "dark"},
		{"--toon", "config", "unset", "theme"},
		{"--toon", "epic", "create", "-t", "Epic"},
		{"--toon", "epic", "list"},
		{"--toon", "epic", "show", "EPIC-1"},
		{"--toon", "epic", "update", "EPIC-1", "-t", "Epic"},
		{"--toon", "epic", "complete", "EPIC-1"},
		{"--toon", "epic", "delete", "EPIC-1"},
		{"--toon", "task", "create", "-t", "Task", "--prefix", "TASK"},
		{"--toon", "task", "list"},
		{"--toon", "task", "show", "TASK-1"},
		{"--toon", "task", "update", "TASK-1", "-s", "in_progress"},
		{"--toon", "task", "delete", "TASK-1"},
		{"--toon", "subtask", "create", "TASK-1", "-t", "Subtask"},
		{"--toon", "subtask", "list", "TASK-1"},
		{"--toon", "subtask", "update", "TASK-2", "-t", "Subtask"},
		{"--toon", "subtask", "delete", "TASK-2"},
		{"--toon", "comment", "add", "TASK-1", "-a", "codex", "-c", "hello"},
		{"--toon", "comment", "list", "TASK-1"},
		{"--toon", "comment", "update", "cmt-1", "-c", "hello"},
		{"--toon", "comment", "delete", "cmt-1"},
		{"--toon", "dep", "add", "TASK-1", "TASK-0"},
		{"--toon", "dep", "remove", "TASK-1", "TASK-0"},
		{"--toon", "dep", "list", "TASK-1"},
		{"--toon", "lock", "acquire", "task", "TASK-1"},
		{"--toon", "lock", "release", "lock-1"},
		{"--toon", "lock", "list"},
		{"--toon", "ready"},
		{"--toon", "search", "test"},
		{"--toon", "history"},
		{"--toon", "list"},
	}

	for _, args := range cases {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := Run(args, &stdout, &stderr)

		require.Equal(t, 1, code, strings.Join(args, " "))
		require.Contains(t, stderr.String(), "failed to read local config", strings.Join(args, " "))
	}
}

func TestRunCommandsFailWhenAPIClientFactoryFails(t *testing.T) {
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

	oldFactory := newAPIClient
	newAPIClient = func(cfg config.LocalConfig) (apiClient, error) {
		return nil, errors.New("factory failed")
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	cases := [][]string{
		{"--toon", "config", "list"},
		{"--toon", "config", "get", "theme"},
		{"--toon", "config", "set", "theme", "dark"},
		{"--toon", "config", "unset", "theme"},
		{"--toon", "epic", "create", "-t", "Epic"},
		{"--toon", "epic", "list"},
		{"--toon", "epic", "show", "EPIC-1"},
		{"--toon", "epic", "update", "EPIC-1", "-t", "Epic"},
		{"--toon", "epic", "complete", "EPIC-1"},
		{"--toon", "epic", "delete", "EPIC-1"},
		{"--toon", "task", "create", "-t", "Task", "--prefix", "TASK"},
		{"--toon", "task", "list"},
		{"--toon", "task", "show", "TASK-1"},
		{"--toon", "task", "update", "TASK-1", "-s", "in_progress"},
		{"--toon", "task", "delete", "TASK-1"},
		{"--toon", "subtask", "create", "TASK-1", "-t", "Subtask"},
		{"--toon", "subtask", "list", "TASK-1"},
		{"--toon", "subtask", "update", "TASK-2", "-t", "Subtask"},
		{"--toon", "subtask", "delete", "TASK-2"},
		{"--toon", "comment", "add", "TASK-1", "-a", "codex", "-c", "hello"},
		{"--toon", "comment", "list", "TASK-1"},
		{"--toon", "comment", "update", "cmt-1", "-c", "hello"},
		{"--toon", "comment", "delete", "cmt-1"},
		{"--toon", "dep", "add", "TASK-1", "TASK-0"},
		{"--toon", "dep", "remove", "TASK-1", "TASK-0"},
		{"--toon", "dep", "list", "TASK-1"},
		{"--toon", "lock", "acquire", "task", "TASK-1"},
		{"--toon", "lock", "release", "lock-1"},
		{"--toon", "lock", "list"},
		{"--toon", "ready"},
		{"--toon", "search", "test"},
		{"--toon", "history"},
		{"--toon", "list"},
	}

	for _, args := range cases {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := Run(args, &stdout, &stderr)

		require.Equal(t, 1, code, strings.Join(args, " "))
		require.Contains(t, stderr.String(), "failed to initialize api client", strings.Join(args, " "))
	}
}

func TestRunCommandsReturnServerErrors(t *testing.T) {
	serverErr := errors.New("server failed")
	cases := []struct {
		name   string
		args   []string
		client *fakeAPIClient
	}{
		{"config list", []string{"--toon", "config", "list"}, &fakeAPIClient{listProjectConfigFn: func(context.Context, string) ([]ProjectConfigItem, error) { return nil, serverErr }}},
		{"config get", []string{"--toon", "config", "get", "theme"}, &fakeAPIClient{getProjectConfigFn: func(context.Context, string, string) (ProjectConfigItem, error) {
			return ProjectConfigItem{}, serverErr
		}}},
		{"config set", []string{"--toon", "config", "set", "theme", "dark"}, &fakeAPIClient{setProjectConfigFn: func(context.Context, string, string, string) (ProjectConfigItem, error) {
			return ProjectConfigItem{}, serverErr
		}}},
		{"config unset", []string{"--toon", "config", "unset", "theme"}, &fakeAPIClient{unsetProjectConfigFn: func(context.Context, string, string) (ProjectConfigItem, error) {
			return ProjectConfigItem{}, serverErr
		}}},
		{"epic create", []string{"--toon", "epic", "create", "-t", "Epic"}, &fakeAPIClient{createEpicFn: func(context.Context, string, domain.EpicCreateRequest) (domain.Epic, error) {
			return domain.Epic{}, serverErr
		}}},
		{"epic list", []string{"--toon", "epic", "list"}, &fakeAPIClient{listEpicsFn: func(context.Context, string, string, int) (domain.EpicListResponse, error) {
			return domain.EpicListResponse{}, serverErr
		}}},
		{"epic show", []string{"--toon", "epic", "show", "EPIC-1"}, &fakeAPIClient{getEpicFn: func(context.Context, string, string) (domain.Epic, error) { return domain.Epic{}, serverErr }}},
		{"epic update", []string{"--toon", "epic", "update", "EPIC-1", "-t", "Epic"}, &fakeAPIClient{updateEpicFn: func(context.Context, string, string, domain.EpicUpdateRequest) (domain.Epic, error) {
			return domain.Epic{}, serverErr
		}}},
		{"epic complete", []string{"--toon", "epic", "complete", "EPIC-1"}, &fakeAPIClient{completeEpicFn: func(context.Context, string, string) (domain.Epic, error) { return domain.Epic{}, serverErr }}},
		{"epic delete", []string{"--toon", "epic", "delete", "EPIC-1"}, &fakeAPIClient{deleteEpicFn: func(context.Context, string, string) (domain.Epic, error) { return domain.Epic{}, serverErr }}},
		{"task create", []string{"--toon", "task", "create", "-t", "Task", "--prefix", "TASK"}, &fakeAPIClient{createTaskFn: func(context.Context, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
			return domain.TaskCreateResponse{}, serverErr
		}}},
		{"subtask create", []string{"--toon", "subtask", "create", "TASK-1", "-t", "Subtask"}, &fakeAPIClient{createSubtaskFn: func(context.Context, string, string, domain.TaskCreateRequest) (domain.TaskCreateResponse, error) {
			return domain.TaskCreateResponse{}, serverErr
		}}},
		{"task show", []string{"--toon", "task", "show", "TASK-1"}, &fakeAPIClient{getTaskFn: func(context.Context, string, string) (domain.TaskDetail, error) {
			return domain.TaskDetail{}, serverErr
		}}},
		{"subtask list", []string{"--toon", "subtask", "list", "TASK-1"}, &fakeAPIClient{listSubtasksFn: func(context.Context, string, string, int) (domain.TaskListResponse, error) {
			return domain.TaskListResponse{}, serverErr
		}}},
		{"task update", []string{"--toon", "task", "update", "TASK-1", "-s", "in_progress"}, &fakeAPIClient{updateTaskFn: func(context.Context, string, string, domain.TaskUpdateRequest) (domain.TaskDetail, error) {
			return domain.TaskDetail{}, serverErr
		}}},
		{"subtask update", []string{"--toon", "subtask", "update", "TASK-2", "-t", "Subtask"}, &fakeAPIClient{updateTaskFn: func(context.Context, string, string, domain.TaskUpdateRequest) (domain.TaskDetail, error) {
			return domain.TaskDetail{}, serverErr
		}}},
		{"task delete", []string{"--toon", "task", "delete", "TASK-1"}, &fakeAPIClient{deleteTaskFn: func(context.Context, string, string) (domain.TaskDetail, error) {
			return domain.TaskDetail{}, serverErr
		}}},
		{"subtask delete", []string{"--toon", "subtask", "delete", "TASK-2"}, &fakeAPIClient{deleteTaskFn: func(context.Context, string, string) (domain.TaskDetail, error) {
			return domain.TaskDetail{}, serverErr
		}}},
		{"comment add", []string{"--toon", "comment", "add", "TASK-1", "-a", "codex", "-c", "hi"}, &fakeAPIClient{addCommentFn: func(context.Context, string, string, domain.CommentCreateRequest) (domain.Comment, error) {
			return domain.Comment{}, serverErr
		}}},
		{"comment list", []string{"--toon", "comment", "list", "TASK-1"}, &fakeAPIClient{listCommentsFn: func(context.Context, string, string) (domain.CommentListResponse, error) {
			return domain.CommentListResponse{}, serverErr
		}}},
		{"comment update", []string{"--toon", "comment", "update", "cmt-1", "-c", "hi"}, &fakeAPIClient{updateCommentFn: func(context.Context, string, string, domain.CommentUpdateRequest) (domain.Comment, error) {
			return domain.Comment{}, serverErr
		}}},
		{"comment delete", []string{"--toon", "comment", "delete", "cmt-1"}, &fakeAPIClient{deleteCommentFn: func(context.Context, string, string) (domain.Comment, error) {
			return domain.Comment{}, serverErr
		}}},
		{"dep add", []string{"--toon", "dep", "add", "TASK-1", "TASK-0"}, &fakeAPIClient{addDependencyFn: func(context.Context, string, string, string) (domain.Dependency, error) {
			return domain.Dependency{}, serverErr
		}}},
		{"dep list", []string{"--toon", "dep", "list", "TASK-1"}, &fakeAPIClient{listDependenciesFn: func(context.Context, string, string) (domain.DependencyListResponse, error) {
			return domain.DependencyListResponse{}, serverErr
		}}},
		{"dep remove", []string{"--toon", "dep", "remove", "TASK-1", "TASK-0"}, &fakeAPIClient{removeDependencyFn: func(context.Context, string, string, string) (domain.Dependency, error) {
			return domain.Dependency{}, serverErr
		}}},
		{"lock acquire", []string{"--toon", "lock", "acquire", "task", "TASK-1"}, &fakeAPIClient{acquireLockFn: func(context.Context, string, domain.LockAcquireRequest) (domain.WorkItemLock, error) {
			return domain.WorkItemLock{}, serverErr
		}}},
		{"lock release", []string{"--toon", "lock", "release", "lock-1"}, &fakeAPIClient{releaseLockFn: func(context.Context, string, string) (domain.WorkItemLock, error) {
			return domain.WorkItemLock{}, serverErr
		}}},
		{"lock list", []string{"--toon", "lock", "list"}, &fakeAPIClient{listLocksFn: func(context.Context, string, string, string, bool) (domain.LockListResponse, error) {
			return domain.LockListResponse{}, serverErr
		}}},
		{"ready", []string{"--toon", "ready"}, &fakeAPIClient{listReadyTasksFn: func(context.Context, string, string, int) (domain.ReadyListResponse, error) {
			return domain.ReadyListResponse{}, serverErr
		}}},
		{"search", []string{"--toon", "search", "test"}, &fakeAPIClient{searchFn: func(context.Context, string, string, string, string, int) (domain.SearchResponse, error) {
			return domain.SearchResponse{}, serverErr
		}}},
		{"history", []string{"--toon", "history"}, &fakeAPIClient{historyFn: func(context.Context, string, string, string, string, string, int) (domain.HistoryResponse, error) {
			return domain.HistoryResponse{}, serverErr
		}}},
		{"list", []string{"--toon", "list"}, &fakeAPIClient{listUnifiedFn: func(context.Context, string, string, string, string, string, int) (domain.ListResponse, error) {
			return domain.ListResponse{}, serverErr
		}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := withConfiguredCLI(t, tc.client)

			code := Run(tc.args, stdout, stderr)

			require.Equal(t, 1, code)
			require.Contains(t, stderr.String(), "server failed")
		})
	}
}

func TestCommandValidationHelpersRejectUnknownValues(t *testing.T) {
	require.False(t, isAllowedTaskStatus("blocked"))
	require.False(t, isAllowedTaskKind("story"))
	require.False(t, isAllowedLockEntityType("milestone"))
	require.False(t, isAllowedCommentKind("note"))
	require.False(t, knownCommand([]string{"bogus"}))
}
