package cli

import (
	"bytes"
	"testing"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestWriteProjectConfigItemTOON(t *testing.T) {
	var buf bytes.Buffer

	writeProjectConfigItem(&buf, ProjectConfigItem{Key: "theme", Value: "dark"})

	require.Equal(t, "- theme: dark\n", buf.String())
}

func TestWriteTaskCreateResponseTOON(t *testing.T) {
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

func TestWriteTaskListItemTOON(t *testing.T) {
	var buf bytes.Buffer

	writeTaskListItem(&buf, 0, domain.TaskListItem{
		ID:       "ABC-1",
		Title:    "Write docs",
		Status:   domain.StatusInProgress,
		Priority: domain.PriorityMedium,
		EpicID:   "epic-1",
		Tags:     []string{"infra", "api"},
	})

	require.Equal(t, "- id: ABC-1\n  title: \"Write docs\"\n  description: \"\"\n  priority: 2\n  status: in_progress\n  epicId: epic-1\n  tags: \"infra,api\"\n", buf.String())
}

func TestWriteReadyListItemTOON(t *testing.T) {
	var buf bytes.Buffer

	writeReadyListItem(&buf, 0, domain.ReadyListItem{
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
	})

	require.Contains(t, buf.String(), "dependents[1]{id,title,status,priority}:\n")
	require.Contains(t, buf.String(), "    - CORE-5,Backups,todo,1\n")
}
