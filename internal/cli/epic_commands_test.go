package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/config"
	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func withConfiguredCLI(t *testing.T, client *fakeAPIClient) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()

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
		return client, nil
	}
	t.Cleanup(func() { newAPIClient = oldFactory })

	_, err = config.WriteLocal(workdir, config.LocalConfig{
		APIURL:       "http://example.invalid",
		WorkspaceID:  "default",
		ProjectID:    "default",
		AccessKey:    "key",
		AccessSecret: "secret",
	})
	require.NoError(t, err)

	return &stdout, &stderr
}

func TestRunEpicCreateCallsServer(t *testing.T) {
	client := &fakeAPIClient{
		createEpicFn: func(ctx context.Context, projectID string, req domain.EpicCreateRequest) (domain.Epic, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "Track auth", req.Title)
			require.Equal(t, "Ship auth work", req.Description)
			require.Equal(t, domain.PriorityHigh, *req.Priority)
			require.Equal(t, "alice", req.AssignedTo)
			require.Equal(t, []string{"login works"}, req.AcceptanceCriteria)
			return domain.Epic{ID: "EPIC-1", Title: req.Title, Description: req.Description, Status: domain.StatusTodo, Priority: *req.Priority}, nil
		},
	}
	stdout, stderr := withConfiguredCLI(t, client)

	code := Run([]string{"--toon", "epic", "create", "-t", "Track auth", "-d", "Ship auth work", "-p", "1", "-a", "alice", "--criteria-json", `["login works"]`}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: EPIC-1")
	require.Contains(t, stdout.String(), `title: "Track auth"`)
}

func TestRunEpicShowUpdateCompleteAndDeleteCallServer(t *testing.T) {
	client := &fakeAPIClient{
		getEpicFn: func(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "EPIC-1", epicID)
			return domain.Epic{ID: epicID, Title: "Track auth", Status: domain.StatusTodo, Priority: domain.PriorityMedium}, nil
		},
		updateEpicFn: func(ctx context.Context, projectID, epicID string, req domain.EpicUpdateRequest) (domain.Epic, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "EPIC-1", epicID)
			require.NotNil(t, req.Title)
			require.Equal(t, "Track auth v2", *req.Title)
			require.NotNil(t, req.Status)
			require.Equal(t, domain.StatusInProgress, *req.Status)
			require.NotNil(t, req.CompletionEvidence)
			require.Equal(t, []string{"make coverage passes"}, *req.CompletionEvidence)
			return domain.Epic{ID: epicID, Title: *req.Title, Status: *req.Status, Priority: domain.PriorityHigh}, nil
		},
		completeEpicFn: func(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "EPIC-1", epicID)
			return domain.Epic{ID: epicID, Title: "Track auth", Status: domain.StatusCompleted, Priority: domain.PriorityHigh}, nil
		},
		deleteEpicFn: func(ctx context.Context, projectID, epicID string) (domain.Epic, error) {
			require.Equal(t, "default", projectID)
			require.Equal(t, "EPIC-1", epicID)
			return domain.Epic{ID: epicID, Title: "Track auth", Status: domain.StatusArchived, Priority: domain.PriorityHigh}, nil
		},
	}

	stdout, stderr := withConfiguredCLI(t, client)
	code := Run([]string{"--toon", "epic", "show", "EPIC-1"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "- id: EPIC-1")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "epic", "update", "EPIC-1", "-t", "Track auth v2", "-s", "in_progress", "--evidence-json", `["make coverage passes"]`}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: in_progress")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "epic", "complete", "EPIC-1"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: completed")

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--toon", "epic", "delete", "EPIC-1"}, stdout, stderr)
	require.Equal(t, 0, code, stderr.String())
	require.Contains(t, stdout.String(), "status: archived")
}

func TestParseEpicArgsRejectInvalidInputs(t *testing.T) {
	_, err := parseEpicCreateArgs(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires -t")

	_, err = parseEpicCreateArgs([]string{"-t", "Track", "extra"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not accept positional")

	_, err = parseEpicUpdateArgs([]string{"EPIC-1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires at least one change")

	_, err = parseEpicUpdateArgs([]string{"EPIC-1", "-s", "bad"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "status must be")
}
