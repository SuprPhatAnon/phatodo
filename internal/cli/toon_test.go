package cli

import (
	"bytes"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestWriteProjectConfigItemTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeProjectConfigItem(&buf, ProjectConfigItem{Key: "theme", Value: "dark"})

	require.Equal(t, "- theme: dark\n", buf.String())
}

func TestWriteProjectConfigItemHuman(t *testing.T) {
	setOutputMode(false)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeProjectConfigItem(&buf, ProjectConfigItem{Key: "theme", Value: "dark"})

	require.Equal(t, "theme: dark\n", buf.String())
}

func TestWriteTaskCreateResponseTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeTaskCreateResponse(&buf, domain.TaskCreateResponse{
		ID:          "ABC-1",
		IssuePrefix: "ABC",
		Title:       "Write docs",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityMedium,
		ProjectID:   "default",
		WorkspaceID: "workspace-1",
	})

	require.Equal(t, "- id: ABC-1\n  issue_prefix: ABC\n  title: \"Write docs\"\n  status: todo\n  priority: 2\n  project_id: default\n  workspace_id: workspace-1\n", buf.String())
}

func TestWriteTaskCreateResponseHuman(t *testing.T) {
	setOutputMode(false)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeTaskCreateResponse(&buf, domain.TaskCreateResponse{
		ID:          "ABC-1",
		IssuePrefix: "ABC",
		Title:       "Write docs",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityMedium,
		ProjectID:   "default",
		WorkspaceID: "workspace-1",
	})

	require.Equal(t, "id: ABC-1\n  issue_prefix: ABC\n  title: Write docs\n  status: todo\n  priority: 2\n  project_id: default\n  workspace_id: workspace-1\n", buf.String())
}

func TestWriteTaskListItemTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeTaskListItem(&buf, 0, domain.TaskListItem{
		ID:                "ABC-1",
		Title:             "Write docs",
		Kind:              domain.TaskKindBug,
		RootCauseAnalysis: "missing trace",
		Status:            domain.StatusInProgress,
		Priority:          domain.PriorityMedium,
		EpicID:            "epic-1",
		Tags:              []string{"infra", "api"},
		PlannedFiles:      []string{"internal/cli/toon.go"},
		ChangedFiles:      []string{"internal/cli/toon_test.go"},
	})

	require.Contains(t, buf.String(), "kind: bug\n")
	require.Contains(t, buf.String(), "rootCauseAnalysis: \"missing trace\"\n")
	require.Contains(t, buf.String(), "plannedFiles[1]:\n")
	require.Contains(t, buf.String(), "changedFiles[1]:\n")
}

func TestWriteReadyListItemTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeReadyListItem(&buf, 0, domain.ReadyListItem{
		ID:           "CORE-1",
		Title:        "Health endpoints",
		Kind:         domain.TaskKindFeature,
		Description:  "Add readiness and liveness checks",
		Status:       domain.StatusTodo,
		Priority:     domain.PriorityHigh,
		EpicID:       "epic-1",
		Tags:         []string{"infra", "api"},
		PlannedFiles: []string{"internal/server/handlers.go"},
		Unblocks: []domain.TaskListItem{
			{
				ID:       "CORE-5",
				Title:    "Backups",
				Kind:     domain.TaskKindTask,
				Status:   domain.StatusTodo,
				Priority: domain.PriorityHigh,
				EpicID:   "epic-1",
				Tags:     []string{"infra"},
			},
		},
	})

	require.Contains(t, buf.String(), "description: \"Add readiness and liveness checks\"")
	require.Contains(t, buf.String(), "kind: feature\n")
	require.Contains(t, buf.String(), "plannedFiles[1]:\n")
	require.Contains(t, buf.String(), "dependents[1]{id,title,status,priority}:\n")
	require.Contains(t, buf.String(), "    - CORE-5,Backups,todo,1\n")
}

func TestWriteUnifiedListItemTOONIncludesTaskMetadata(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeUnifiedListItem(&buf, 0, domain.UnifiedListItem{
		EntityType:        "task",
		ID:                "ABC-1",
		Title:             "Write docs",
		Kind:              domain.TaskKindBug,
		RootCauseAnalysis: "missing trace",
		Description:       "Docs need updates",
		Status:            domain.StatusCompleted,
		Priority:          domain.PriorityMedium,
		PlannedFiles:      []string{"docs/QUICKSTART.md"},
		ChangedFiles:      []string{"docs/CLI_OUTPUT.md"},
	})

	require.Contains(t, buf.String(), "kind: bug\n")
	require.Contains(t, buf.String(), "rootCauseAnalysis: \"missing trace\"\n")
	require.Contains(t, buf.String(), "plannedFiles[1]:\n")
	require.Contains(t, buf.String(), "changedFiles[1]:\n")
}

func TestWriteCommentTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeComment(&buf, 0, domain.Comment{
		ID:      "cmt-1",
		Author:  "agent",
		Kind:    "summary",
		Content: "Done",
	})

	require.Equal(t, "- id: cmt-1\n  author: agent\n  kind: summary\n  content: Done\n", buf.String())
}

func TestWriteDependencyTOON(t *testing.T) {
	setOutputMode(true)
	t.Cleanup(func() { setOutputMode(false) })
	var buf bytes.Buffer

	writeDependency(&buf, 0, domain.Dependency{
		ID:          "dep-1",
		TaskID:      "ABC-1",
		DependsOnID: "ABC-2",
	})

	require.Equal(t, "- id: dep-1\n  taskId: ABC-1\n  dependsOnId: ABC-2\n", buf.String())
}
