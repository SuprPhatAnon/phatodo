package postgres

import (
	"encoding/json"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
)

func taskDetailFromSQLC(row db.Task) (domain.TaskDetail, error) {
	task := domain.TaskDetail{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Kind:        domain.TaskKind(row.Kind),
		Title:       row.Title,
		Priority:    domain.Priority(row.Priority),
		Status:      domain.Status(row.Status),
		Tags:        row.Tags,
	}
	if row.EpicID != nil {
		task.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		task.ParentTaskID = *row.ParentTaskID
	}
	if row.AssignedTo != nil {
		task.AssignedTo = *row.AssignedTo
	}
	if row.CreatedBy != nil {
		task.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		task.UpdatedBy = *row.UpdatedBy
	}
	if row.CompletedBy != nil {
		task.CompletedBy = *row.CompletedBy
	}
	if row.Description != nil {
		task.Description = *row.Description
	}
	task.RootCauseAnalysis = row.RootCauseAnalysis
	if row.CompletionSummary != nil {
		task.CompletionSummary = *row.CompletionSummary
	}
	if row.CompletedAt.Valid {
		task.CompletedAt = row.CompletedAt.Time
	}
	if len(row.AcceptanceCriteria) > 0 {
		if err := json.Unmarshal(row.AcceptanceCriteria, &task.AcceptanceCriteria); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.PlannedFiles) > 0 {
		if err := json.Unmarshal(row.PlannedFiles, &task.PlannedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.ChangedFiles) > 0 {
		if err := json.Unmarshal(row.ChangedFiles, &task.ChangedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.CompletionEvidence) > 0 {
		if err := json.Unmarshal(row.CompletionEvidence, &task.CompletionEvidence); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	return task, nil
}

func taskDetailFromUpdateTaskRow(row db.UpdateTaskRow) (domain.TaskDetail, error) {
	task := domain.TaskDetail{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Kind:        domain.TaskKind(row.Kind),
		Title:       row.Title,
		Priority:    domain.Priority(row.Priority),
		Status:      domain.Status(row.Status),
		Tags:        row.Tags,
	}
	if row.EpicID != nil {
		task.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		task.ParentTaskID = *row.ParentTaskID
	}
	if row.AssignedTo != nil {
		task.AssignedTo = *row.AssignedTo
	}
	if row.CreatedBy != nil {
		task.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		task.UpdatedBy = *row.UpdatedBy
	}
	if row.CompletedBy != nil {
		task.CompletedBy = *row.CompletedBy
	}
	if row.Description != nil {
		task.Description = *row.Description
	}
	task.RootCauseAnalysis = row.RootCauseAnalysis
	if row.CompletionSummary != nil {
		task.CompletionSummary = *row.CompletionSummary
	}
	if row.CompletedAt.Valid {
		task.CompletedAt = row.CompletedAt.Time
	}
	if len(row.AcceptanceCriteria) > 0 {
		if err := json.Unmarshal(row.AcceptanceCriteria, &task.AcceptanceCriteria); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.PlannedFiles) > 0 {
		if err := json.Unmarshal(row.PlannedFiles, &task.PlannedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.ChangedFiles) > 0 {
		if err := json.Unmarshal(row.ChangedFiles, &task.ChangedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.CompletionEvidence) > 0 {
		if err := json.Unmarshal(row.CompletionEvidence, &task.CompletionEvidence); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	return task, nil
}

func taskDetailFromGetTaskDetailRow(row db.GetTaskDetailRow) (domain.TaskDetail, error) {
	task := domain.TaskDetail{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Kind:        domain.TaskKind(row.Kind),
		Title:       row.Title,
		Priority:    domain.Priority(row.Priority),
		Status:      domain.Status(row.Status),
		Tags:        row.Tags,
	}
	if row.EpicID != nil {
		task.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		task.ParentTaskID = *row.ParentTaskID
	}
	if row.AssignedTo != nil {
		task.AssignedTo = *row.AssignedTo
	}
	if row.CreatedBy != nil {
		task.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		task.UpdatedBy = *row.UpdatedBy
	}
	if row.CompletedBy != nil {
		task.CompletedBy = *row.CompletedBy
	}
	if row.Description != nil {
		task.Description = *row.Description
	}
	task.RootCauseAnalysis = row.RootCauseAnalysis
	if row.CompletionSummary != nil {
		task.CompletionSummary = *row.CompletionSummary
	}
	if row.CompletedAt.Valid {
		task.CompletedAt = row.CompletedAt.Time
	}
	if len(row.AcceptanceCriteria) > 0 {
		if err := json.Unmarshal(row.AcceptanceCriteria, &task.AcceptanceCriteria); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.PlannedFiles) > 0 {
		if err := json.Unmarshal(row.PlannedFiles, &task.PlannedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.ChangedFiles) > 0 {
		if err := json.Unmarshal(row.ChangedFiles, &task.ChangedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.CompletionEvidence) > 0 {
		if err := json.Unmarshal(row.CompletionEvidence, &task.CompletionEvidence); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	return task, nil
}

func taskDetailFromDeleteTaskRow(row db.DeleteTaskRow) (domain.TaskDetail, error) {
	task := domain.TaskDetail{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Kind:        domain.TaskKind(row.Kind),
		Title:       row.Title,
		Priority:    domain.Priority(row.Priority),
		Status:      domain.Status(row.Status),
		Tags:        row.Tags,
	}
	if row.EpicID != nil {
		task.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		task.ParentTaskID = *row.ParentTaskID
	}
	if row.AssignedTo != nil {
		task.AssignedTo = *row.AssignedTo
	}
	if row.CreatedBy != nil {
		task.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		task.UpdatedBy = *row.UpdatedBy
	}
	if row.CompletedBy != nil {
		task.CompletedBy = *row.CompletedBy
	}
	if row.Description != nil {
		task.Description = *row.Description
	}
	task.RootCauseAnalysis = row.RootCauseAnalysis
	if row.CompletionSummary != nil {
		task.CompletionSummary = *row.CompletionSummary
	}
	if row.CompletedAt.Valid {
		task.CompletedAt = row.CompletedAt.Time
	}
	if len(row.AcceptanceCriteria) > 0 {
		if err := json.Unmarshal(row.AcceptanceCriteria, &task.AcceptanceCriteria); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.PlannedFiles) > 0 {
		if err := json.Unmarshal(row.PlannedFiles, &task.PlannedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.ChangedFiles) > 0 {
		if err := json.Unmarshal(row.ChangedFiles, &task.ChangedFiles); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	if len(row.CompletionEvidence) > 0 {
		if err := json.Unmarshal(row.CompletionEvidence, &task.CompletionEvidence); err != nil {
			return domain.TaskDetail{}, err
		}
	}
	return task, nil
}

func taskListItemFromSQLC(row db.ListTasksRow) domain.TaskListItem {
	item := domain.TaskListItem{
		ID:        row.ID,
		Title:     row.Title,
		Status:    domain.Status(row.Status),
		Priority:  domain.Priority(row.Priority),
		Tags:      row.Tags,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	item.Kind = domain.TaskKind(row.Kind)
	item.RootCauseAnalysis = row.RootCauseAnalysis
	decodeStringArray(row.PlannedFiles, &item.PlannedFiles)
	decodeStringArray(row.ChangedFiles, &item.ChangedFiles)
	if row.EpicID != nil {
		item.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		item.ParentTaskID = *row.ParentTaskID
	}
	return item
}

func readyListItemFromSQLC(row db.ListReadyTasksRow) domain.ReadyListItem {
	item := domain.ReadyListItem{
		ID:       row.ID,
		Title:    row.Title,
		Status:   domain.Status(row.Status),
		Priority: domain.Priority(row.Priority),
		Kind:     domain.TaskKind(row.Kind),
		Tags:     row.Tags,
	}
	item.RootCauseAnalysis = row.RootCauseAnalysis
	decodeStringArray(row.PlannedFiles, &item.PlannedFiles)
	decodeStringArray(row.ChangedFiles, &item.ChangedFiles)
	if row.Description != nil {
		item.Description = *row.Description
	}
	if row.EpicID != nil {
		item.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		item.ParentTaskID = *row.ParentTaskID
	}
	return item
}

func readyDependentFromSQLC(row db.ListReadyDependentsRow) domain.TaskListItem {
	item := domain.TaskListItem{
		ID:       row.ID,
		Title:    row.Title,
		Status:   domain.Status(row.Status),
		Priority: domain.Priority(row.Priority),
		Kind:     domain.TaskKind(row.Kind),
		Tags:     row.Tags,
	}
	item.RootCauseAnalysis = row.RootCauseAnalysis
	decodeStringArray(row.PlannedFiles, &item.PlannedFiles)
	decodeStringArray(row.ChangedFiles, &item.ChangedFiles)
	if row.EpicID != nil {
		item.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		item.ParentTaskID = *row.ParentTaskID
	}
	return item
}

func unifiedItemFromEpicSQLC(row db.Epic) domain.UnifiedListItem {
	return domain.UnifiedListItem{
		EntityType:  "epic",
		ID:          row.ID,
		Title:       row.Title,
		Description: derefString(row.Description),
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func unifiedItemFromTaskSQLC(row db.ListTasksUnifiedRow) domain.UnifiedListItem {
	item := domain.UnifiedListItem{
		EntityType:  "task",
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		Kind:        domain.TaskKind(row.Kind),
		Tags:        row.Tags,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	item.RootCauseAnalysis = row.RootCauseAnalysis
	decodeStringArray(row.PlannedFiles, &item.PlannedFiles)
	decodeStringArray(row.ChangedFiles, &item.ChangedFiles)
	if row.EpicID != nil {
		item.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		item.ParentTaskID = *row.ParentTaskID
		item.EntityType = "subtask"
	}
	return item
}

func searchItemFromEpicSQLC(row db.SearchEpicsRow) domain.SearchItem {
	return domain.SearchItem{
		EntityType:  "epic",
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func searchItemFromTaskSQLC(row db.SearchTasksRow) domain.SearchItem {
	item := domain.SearchItem{
		EntityType:  "task",
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Kind:        row.Kind,
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.EpicID != nil {
		item.EpicID = *row.EpicID
	}
	if row.ParentTaskID != nil {
		item.EntityType = "subtask"
		item.ParentTaskID = *row.ParentTaskID
	}
	return item
}

func searchItemFromCommentSQLC(row db.SearchCommentsRow) domain.SearchItem {
	return domain.SearchItem{
		EntityType:   "comment",
		ID:           row.ID,
		Author:       row.Author,
		Kind:         row.Kind,
		Content:      row.Content,
		ParentTaskID: row.TaskID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func historyEventFromSQLC(row db.Event) domain.HistoryEvent {
	evt := domain.HistoryEvent{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Action:      row.Action,
		EntityType:  row.EntityType,
		EntityID:    row.EntityID,
		CreatedAt:   row.CreatedAt,
	}
	if row.ActorUserID != nil {
		evt.ActorUserID = *row.ActorUserID
	}
	if row.ActorLabel != nil {
		evt.ActorLabel = *row.ActorLabel
	}
	if len(row.BeforeState) > 0 {
		evt.BeforeState = append(json.RawMessage(nil), row.BeforeState...)
	}
	if len(row.AfterState) > 0 {
		evt.AfterState = append(json.RawMessage(nil), row.AfterState...)
	}
	if len(row.Metadata) > 0 {
		evt.Metadata = append(json.RawMessage(nil), row.Metadata...)
	}
	return evt
}

func lockFromSQLC(row db.WorkItemLock) domain.WorkItemLock {
	lock := domain.WorkItemLock{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		EntityType:  row.EntityType,
		EntityID:    row.EntityID,
		LockedBy:    row.LockedBy,
		Reason:      derefString(row.Reason),
		LeasedAt:    row.LeasedAt,
		ExpiresAt:   row.ExpiresAt,
	}
	if row.ReleasedAt.Valid {
		lock.ReleasedAt = row.ReleasedAt.Time
	}
	return lock
}

func dependencyFromSQLC(row db.Dependency) domain.Dependency {
	return domain.Dependency{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		TaskID:      row.TaskID,
		DependsOnID: row.DependsOnID,
		CreatedAt:   row.CreatedAt,
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func decodeStringArray(data []byte, target *[]string) {
	if len(data) == 0 {
		return
	}
	_ = json.Unmarshal(data, target)
}

func epicFromSQLC(row db.Epic) (domain.Epic, error) {
	epic := domain.Epic{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		Title:       row.Title,
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.AssignedTo != nil {
		epic.AssignedTo = *row.AssignedTo
	}
	if row.CreatedBy != nil {
		epic.CreatedBy = *row.CreatedBy
	}
	if row.UpdatedBy != nil {
		epic.UpdatedBy = *row.UpdatedBy
	}
	if row.CompletedBy != nil {
		epic.CompletedBy = *row.CompletedBy
	}
	if row.Description != nil {
		epic.Description = *row.Description
	}
	if row.CompletionSummary != nil {
		epic.CompletionSummary = *row.CompletionSummary
	}
	if row.CompletedAt.Valid {
		epic.CompletedAt = row.CompletedAt.Time
	}
	if len(row.AcceptanceCriteria) > 0 {
		if err := json.Unmarshal(row.AcceptanceCriteria, &epic.AcceptanceCriteria); err != nil {
			return domain.Epic{}, err
		}
	}
	if len(row.CompletionEvidence) > 0 {
		if err := json.Unmarshal(row.CompletionEvidence, &epic.CompletionEvidence); err != nil {
			return domain.Epic{}, err
		}
	}
	return epic, nil
}
