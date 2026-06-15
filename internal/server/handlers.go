package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/SuprPhatAnon/phatodo/internal/storage/postgres"
)

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": databaseState(a.config.PostgresDSN),
	})
}

func (a *app) apiIndex(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"name":    "phatodo API",
		"version": "v1",
		"resources": []string{
			"projects",
			"epics",
			"tasks",
			"subtasks",
			"comments",
			"dependencies",
			"config",
			"locks",
			"search",
			"history",
			"list",
		},
	})
}

func (a *app) listProjectConfig(w http.ResponseWriter, r *http.Request) {
	if a.config.ProjectConfigReader == nil {
		respondError(w, http.StatusServiceUnavailable, "config_store_unavailable", "project config store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	items, err := a.config.ProjectConfigReader.ListProjectConfig(r.Context(), projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "project_config_list_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"project_id": projectID,
		"items":      items,
	})
}

func (a *app) listEpics(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicLister == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !isAllowedEpicStatus(status) {
		respondError(w, http.StatusBadRequest, "invalid_status", "status must be todo, in_progress, completed, or archived")
		return
	}

	result, err := a.config.EpicLister.ListEpics(r.Context(), projectID, status)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) createEpic(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicCreator == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	req, err := decodeEpicCreateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.EpicCreator.CreateEpic(r.Context(), projectID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_create_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) showEpic(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicReader == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	epicID := r.PathValue("epicID")
	result, err := a.config.EpicReader.GetEpic(r.Context(), projectID, epicID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_show_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) updateEpic(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicUpdater == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	epicID := r.PathValue("epicID")
	req, err := decodeEpicUpdateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.EpicUpdater.UpdateEpic(r.Context(), projectID, epicID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_update_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) completeEpic(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicCompleter == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	epicID := r.PathValue("epicID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.EpicCompleter.CompleteEpic(r.Context(), projectID, epicID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_complete_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) deleteEpic(w http.ResponseWriter, r *http.Request) {
	if a.config.EpicDeleter == nil {
		respondError(w, http.StatusServiceUnavailable, "epic_store_unavailable", "epic store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	epicID := r.PathValue("epicID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.EpicDeleter.DeleteEpic(r.Context(), projectID, epicID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "epic_delete_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) getProjectConfig(w http.ResponseWriter, r *http.Request) {
	if a.config.ProjectConfigReader == nil {
		respondError(w, http.StatusServiceUnavailable, "config_store_unavailable", "project config store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "config key is required")
		return
	}

	item, err := a.config.ProjectConfigReader.GetProjectConfig(r.Context(), projectID, key)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectConfigNotFound) {
			respondError(w, http.StatusNotFound, "project_config_not_found", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "project_config_get_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (a *app) setProjectConfig(w http.ResponseWriter, r *http.Request) {
	if a.config.ProjectConfigWriter == nil {
		respondError(w, http.StatusServiceUnavailable, "config_store_unavailable", "project config store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "config key is required")
		return
	}

	req, err := decodeProjectConfigSetRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	item, err := a.config.ProjectConfigWriter.SetProjectConfig(r.Context(), projectID, key, req.Value)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "project_config_set_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (a *app) unsetProjectConfig(w http.ResponseWriter, r *http.Request) {
	if a.config.ProjectConfigWriter == nil {
		respondError(w, http.StatusServiceUnavailable, "config_store_unavailable", "project config store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	key := r.PathValue("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "config key is required")
		return
	}

	item, err := a.config.ProjectConfigWriter.DeleteProjectConfig(r.Context(), projectID, key)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectConfigNotFound) {
			respondError(w, http.StatusNotFound, "project_config_not_found", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "project_config_unset_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (a *app) createTask(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskCreator == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	req, err := decodeTaskCreateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.TaskCreator.CreateTask(r.Context(), projectID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		case errors.Is(err, postgres.ErrAssignedUserNotFound):
			respondError(w, http.StatusNotFound, "assigned_user_not_found", err.Error())
		case errors.Is(err, postgres.ErrInvalidIssuePrefix):
			respondError(w, http.StatusBadRequest, "invalid_issue_prefix", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "task_create_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) createSubtask(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskCreator == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	parentTaskID := r.PathValue("taskID")
	req, err := decodeSubtaskCreateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	req.ParentTaskID = parentTaskID

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.TaskCreator.CreateTask(r.Context(), projectID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		case errors.Is(err, postgres.ErrAssignedUserNotFound):
			respondError(w, http.StatusNotFound, "assigned_user_not_found", err.Error())
		case errors.Is(err, postgres.ErrInvalidIssuePrefix):
			respondError(w, http.StatusBadRequest, "invalid_issue_prefix", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "subtask_create_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) listComments(w http.ResponseWriter, r *http.Request) {
	if a.config.CommentLister == nil {
		respondError(w, http.StatusServiceUnavailable, "comment_store_unavailable", "comment store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	result, err := a.config.CommentLister.ListComments(r.Context(), projectID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "comment_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) createComment(w http.ResponseWriter, r *http.Request) {
	if a.config.CommentCreator == nil {
		respondError(w, http.StatusServiceUnavailable, "comment_store_unavailable", "comment store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	req, err := decodeCommentCreateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.CommentCreator.CreateComment(r.Context(), projectID, taskID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "comment_create_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) updateComment(w http.ResponseWriter, r *http.Request) {
	if a.config.CommentUpdater == nil {
		respondError(w, http.StatusServiceUnavailable, "comment_store_unavailable", "comment store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	commentID := r.PathValue("commentID")
	req, err := decodeCommentUpdateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.CommentUpdater.UpdateComment(r.Context(), projectID, commentID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrCommentNotFound):
			respondError(w, http.StatusNotFound, "comment_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "comment_update_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) deleteComment(w http.ResponseWriter, r *http.Request) {
	if a.config.CommentDeleter == nil {
		respondError(w, http.StatusServiceUnavailable, "comment_store_unavailable", "comment store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	commentID := r.PathValue("commentID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.CommentDeleter.DeleteComment(r.Context(), projectID, commentID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrCommentNotFound):
			respondError(w, http.StatusNotFound, "comment_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "comment_delete_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listDependencies(w http.ResponseWriter, r *http.Request) {
	if a.config.DependencyLister == nil {
		respondError(w, http.StatusServiceUnavailable, "dependency_store_unavailable", "dependency store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	result, err := a.config.DependencyLister.ListDependencies(r.Context(), projectID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "dependency_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) addDependency(w http.ResponseWriter, r *http.Request) {
	if a.config.DependencyAdder == nil {
		respondError(w, http.StatusServiceUnavailable, "dependency_store_unavailable", "dependency store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	dependsOnID, err := decodeDependencyPair(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.DependencyAdder.AddDependency(r.Context(), projectID, taskID, dependsOnID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		case errors.Is(err, postgres.ErrDependencyNotFound):
			respondError(w, http.StatusNotFound, "dependency_not_found", err.Error())
		case errors.Is(err, postgres.ErrDuplicateDependency):
			respondError(w, http.StatusConflict, "dependency_exists", err.Error())
		case errors.Is(err, postgres.ErrDependencyCycle):
			respondError(w, http.StatusBadRequest, "dependency_cycle", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "dependency_add_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) removeDependency(w http.ResponseWriter, r *http.Request) {
	if a.config.DependencyRemover == nil {
		respondError(w, http.StatusServiceUnavailable, "dependency_store_unavailable", "dependency store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	dependsOnID := r.PathValue("dependsOnID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.DependencyRemover.RemoveDependency(r.Context(), projectID, taskID, dependsOnID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		case errors.Is(err, postgres.ErrDependencyNotFound):
			respondError(w, http.StatusNotFound, "dependency_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "dependency_remove_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listLocks(w http.ResponseWriter, r *http.Request) {
	if a.config.LockLister == nil {
		respondError(w, http.StatusServiceUnavailable, "lock_store_unavailable", "lock store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	entityTypes := splitCSVQuery(r.URL.Query().Get("type"))
	entityID := strings.TrimSpace(r.URL.Query().Get("entity"))
	active := false
	if activeRaw := strings.TrimSpace(r.URL.Query().Get("active")); activeRaw != "" {
		parsed, err := strconv.ParseBool(activeRaw)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_request", "active must be true or false")
			return
		}
		active = parsed
	}

	result, err := a.config.LockLister.ListLocks(r.Context(), projectID, entityTypes, entityID, active)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "lock_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) acquireLock(w http.ResponseWriter, r *http.Request) {
	if a.config.LockAcquirer == nil {
		respondError(w, http.StatusServiceUnavailable, "lock_store_unavailable", "lock store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	req, err := decodeLockAcquireRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.LockAcquirer.AcquireLock(r.Context(), projectID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		case errors.Is(err, postgres.ErrLockConflict):
			respondError(w, http.StatusConflict, "lock_conflict", err.Error())
		case errors.Is(err, postgres.ErrInvalidLockEntityType):
			respondError(w, http.StatusBadRequest, "invalid_entity_type", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "lock_acquire_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) releaseLock(w http.ResponseWriter, r *http.Request) {
	if a.config.LockReleaser == nil {
		respondError(w, http.StatusServiceUnavailable, "lock_store_unavailable", "lock store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	lockID := r.PathValue("lockID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.LockReleaser.ReleaseLock(r.Context(), projectID, lockID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrLockNotFound):
			respondError(w, http.StatusNotFound, "lock_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "lock_release_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) search(w http.ResponseWriter, r *http.Request) {
	if a.config.Searcher == nil {
		respondError(w, http.StatusServiceUnavailable, "search_store_unavailable", "search store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "search query is required")
		return
	}

	entityType := strings.TrimSpace(r.URL.Query().Get("type"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := parseQueryLimit(r.URL.Query().Get("limit"), 20)

	result, err := a.config.Searcher.Search(r.Context(), projectID, query, entityType, status, limit)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "search_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) history(w http.ResponseWriter, r *http.Request) {
	if a.config.Historian == nil {
		respondError(w, http.StatusServiceUnavailable, "history_store_unavailable", "history store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	entityID := strings.TrimSpace(r.URL.Query().Get("entity"))
	entityType := strings.TrimSpace(r.URL.Query().Get("type"))
	action := strings.TrimSpace(r.URL.Query().Get("action"))
	since := strings.TrimSpace(r.URL.Query().Get("since"))
	limit := parseQueryLimit(r.URL.Query().Get("limit"), 50)

	result, err := a.config.Historian.History(r.Context(), projectID, entityID, entityType, action, since, limit)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "history_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listUnified(w http.ResponseWriter, r *http.Request) {
	if a.config.ListLister == nil {
		respondError(w, http.StatusServiceUnavailable, "list_store_unavailable", "list store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	entityType := strings.TrimSpace(r.URL.Query().Get("type"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	priority := strings.TrimSpace(r.URL.Query().Get("priority"))
	sortSpec := strings.TrimSpace(r.URL.Query().Get("sort"))
	limit := parseQueryLimit(r.URL.Query().Get("limit"), 50)

	result, err := a.config.ListLister.ListUnified(r.Context(), projectID, entityType, status, priority, sortSpec, limit)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) showTask(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskReader == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	result, err := a.config.TaskReader.GetTask(r.Context(), projectID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "task_show_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listSubtasks(w http.ResponseWriter, r *http.Request) {
	if a.config.SubtaskLister == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	parentTaskID := r.PathValue("taskID")

	result, err := a.config.SubtaskLister.ListSubtasks(r.Context(), projectID, parentTaskID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "subtask_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) updateTask(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskUpdater == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")
	req, err := decodeTaskUpdateRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.TaskUpdater.UpdateTask(r.Context(), projectID, taskID, req, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		case errors.Is(err, postgres.ErrEpicNotFound):
			respondError(w, http.StatusNotFound, "epic_not_found", err.Error())
		case errors.Is(err, postgres.ErrAssignedUserNotFound):
			respondError(w, http.StatusNotFound, "assigned_user_not_found", err.Error())
		case errors.Is(err, postgres.ErrInvalidIssuePrefix):
			respondError(w, http.StatusBadRequest, "invalid_issue_prefix", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "task_update_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) deleteTask(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskDeleter == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	taskID := r.PathValue("taskID")

	actor, _ := principalFromContext(r.Context())
	result, err := a.config.TaskDeleter.DeleteTask(r.Context(), projectID, taskID, actor.UserID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		case errors.Is(err, postgres.ErrTaskNotFound):
			respondError(w, http.StatusNotFound, "task_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "task_delete_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listTasks(w http.ResponseWriter, r *http.Request) {
	if a.config.TaskLister == nil {
		respondError(w, http.StatusServiceUnavailable, "task_store_unavailable", "task store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	epicID := strings.TrimSpace(r.URL.Query().Get("epic"))
	if status != "" && !isAllowedTaskStatus(status) {
		respondError(w, http.StatusBadRequest, "invalid_status", "status must be todo, in_progress, completed, wont_fix, or archived")
		return
	}

	result, err := a.config.TaskLister.ListTasks(r.Context(), projectID, status, epicID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "task_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) listReadyTasks(w http.ResponseWriter, r *http.Request) {
	if a.config.ReadyLister == nil {
		respondError(w, http.StatusServiceUnavailable, "ready_store_unavailable", "ready store is not configured")
		return
	}

	projectID := r.PathValue("projectID")
	epicID := strings.TrimSpace(r.URL.Query().Get("epic"))

	result, err := a.config.ReadyLister.ListReadyTasks(r.Context(), projectID, epicID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectNotFound):
			respondError(w, http.StatusNotFound, "project_not_found", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "ready_list_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) adminInit(w http.ResponseWriter, r *http.Request) {
	if a.config.BootstrapManager == nil {
		respondError(w, http.StatusServiceUnavailable, "bootstrap_store_unavailable", "bootstrap store is not configured")
		return
	}

	req, err := decodeAdminInitRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result, err := a.config.BootstrapManager.InitAdmin(r.Context(), req)
	if err != nil {
		if errors.Is(err, postgres.ErrAdminAlreadyExists) {
			respondError(w, http.StatusConflict, "admin_already_exists", err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "admin_init_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (a *app) adminBootstrap(w http.ResponseWriter, r *http.Request) {
	if a.config.BootstrapManager == nil {
		respondError(w, http.StatusServiceUnavailable, "bootstrap_store_unavailable", "bootstrap store is not configured")
		return
	}

	req, err := decodeAdminBootstrapRequest(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result, err := a.config.BootstrapManager.BootstrapProject(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrProjectAlreadyExists):
			respondError(w, http.StatusConflict, "project_already_exists", err.Error())
		case errors.Is(err, postgres.ErrInvalidAdminCredentials):
			respondError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "admin_bootstrap_failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *app) dashboardPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>phatodo</title></head>
<body>
<main>
<h1>phatodo dashboard</h1>
<p>Server scaffold is running. API and dashboard modules will be wired to Postgres-backed task data.</p>
</main>
</body>
</html>`))
}

func (a *app) notImplemented(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("projectID")
		if projectID == "" {
			projectID = r.URL.Query().Get("project_id")
		}

		actor, _ := principalFromContext(r.Context())
		respondJSON(w, http.StatusNotImplemented, map[string]any{
			"error":      "not_implemented",
			"action":     action,
			"method":     r.Method,
			"path":       r.URL.Path,
			"project_id": projectID,
			"actor_id":   actor.UserID,
		})
	}
}

func decodeAdminInitRequest(body io.Reader) (domain.AdminInitRequest, error) {
	var req domain.AdminInitRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.AdminInitRequest{}, err
	}
	if req.Username == "" {
		return domain.AdminInitRequest{}, errors.New("username is required")
	}
	if req.Password == "" {
		return domain.AdminInitRequest{}, errors.New("password is required")
	}
	return req, nil
}

func decodeEpicCreateRequest(body io.Reader) (domain.EpicCreateRequest, error) {
	var payload struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		Priority           *int     `json:"priority"`
		AssignedTo         string   `json:"assigned_to"`
		AcceptanceCriteria []string `json:"acceptance_criteria"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return domain.EpicCreateRequest{}, err
	}
	if strings.TrimSpace(payload.Title) == "" {
		return domain.EpicCreateRequest{}, errors.New("title is required")
	}
	req := domain.EpicCreateRequest{
		Title:              payload.Title,
		Description:        payload.Description,
		AssignedTo:         payload.AssignedTo,
		AcceptanceCriteria: payload.AcceptanceCriteria,
	}
	if payload.Priority != nil {
		priority := domain.Priority(*payload.Priority)
		req.Priority = &priority
	}
	return req, nil
}

func decodeEpicUpdateRequest(body io.Reader) (domain.EpicUpdateRequest, error) {
	var payload struct {
		Title              *string   `json:"title"`
		Description        *string   `json:"description"`
		Priority           *int      `json:"priority"`
		Status             *string   `json:"status"`
		AssignedTo         *string   `json:"assigned_to"`
		AcceptanceCriteria *[]string `json:"acceptance_criteria"`
		CompletionSummary  *string   `json:"completion_summary"`
		CompletionEvidence *[]string `json:"completion_evidence"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return domain.EpicUpdateRequest{}, err
	}

	req := domain.EpicUpdateRequest{
		Title:              payload.Title,
		Description:        payload.Description,
		AssignedTo:         payload.AssignedTo,
		AcceptanceCriteria: payload.AcceptanceCriteria,
		CompletionSummary:  payload.CompletionSummary,
		CompletionEvidence: payload.CompletionEvidence,
	}
	if payload.Priority != nil {
		priority := domain.Priority(*payload.Priority)
		req.Priority = &priority
	}
	if payload.Status != nil {
		status := domain.Status(strings.TrimSpace(*payload.Status))
		if !isAllowedEpicStatus(string(status)) {
			return domain.EpicUpdateRequest{}, errors.New("epic status must be todo, in_progress, completed, or archived")
		}
		req.Status = &status
	}

	if req.Title == nil && req.Description == nil && req.Priority == nil && req.Status == nil && req.AssignedTo == nil && req.AcceptanceCriteria == nil && req.CompletionSummary == nil && req.CompletionEvidence == nil {
		return domain.EpicUpdateRequest{}, errors.New("epic update requires at least one change")
	}

	return req, nil
}

func decodeAdminBootstrapRequest(body io.Reader) (domain.AdminBootstrapRequest, error) {
	var req domain.AdminBootstrapRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.AdminBootstrapRequest{}, err
	}
	if req.Username == "" {
		return domain.AdminBootstrapRequest{}, errors.New("username is required")
	}
	if req.Password == "" {
		return domain.AdminBootstrapRequest{}, errors.New("password is required")
	}
	if req.WorkspaceName == "" {
		return domain.AdminBootstrapRequest{}, errors.New("workspace_name is required")
	}
	if req.ProjectName == "" {
		return domain.AdminBootstrapRequest{}, errors.New("project_name is required")
	}
	return req, nil
}

func decodeProjectConfigSetRequest(body io.Reader) (domain.ProjectConfigSetRequest, error) {
	var req domain.ProjectConfigSetRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.ProjectConfigSetRequest{}, err
	}
	return req, nil
}

func decodeTaskCreateRequest(body io.Reader) (domain.TaskCreateRequest, error) {
	var req domain.TaskCreateRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.TaskCreateRequest{}, err
	}
	if req.Title == "" {
		return domain.TaskCreateRequest{}, errors.New("title is required")
	}
	if req.IssuePrefix == "" {
		return domain.TaskCreateRequest{}, errors.New("issue_prefix is required")
	}
	return req, nil
}

func decodeSubtaskCreateRequest(body io.Reader) (domain.TaskCreateRequest, error) {
	var req domain.TaskCreateRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.TaskCreateRequest{}, err
	}
	if req.Title == "" {
		return domain.TaskCreateRequest{}, errors.New("title is required")
	}
	return req, nil
}

func decodeCommentCreateRequest(body io.Reader) (domain.CommentCreateRequest, error) {
	var req domain.CommentCreateRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.CommentCreateRequest{}, err
	}
	if req.Author == "" {
		return domain.CommentCreateRequest{}, errors.New("author is required")
	}
	if req.Content == "" {
		return domain.CommentCreateRequest{}, errors.New("content is required")
	}
	if req.Kind == "" {
		req.Kind = "comment"
	}
	switch req.Kind {
	case "comment", "analysis", "summary", "checkpoint", "handoff":
	default:
		return domain.CommentCreateRequest{}, errors.New("kind must be comment, analysis, summary, checkpoint, or handoff")
	}
	return req, nil
}

func decodeCommentUpdateRequest(body io.Reader) (domain.CommentUpdateRequest, error) {
	var req domain.CommentUpdateRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.CommentUpdateRequest{}, err
	}
	if req.Content == "" {
		return domain.CommentUpdateRequest{}, errors.New("content is required")
	}
	return req, nil
}

func decodeDependencyPair(body io.Reader) (string, error) {
	var req struct {
		DependsOnID string `json:"depends_on_id"`
	}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return "", err
	}
	if req.DependsOnID == "" {
		return "", errors.New("depends_on_id is required")
	}
	return req.DependsOnID, nil
}

func decodeLockAcquireRequest(body io.Reader) (domain.LockAcquireRequest, error) {
	var req domain.LockAcquireRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.LockAcquireRequest{}, err
	}
	if strings.TrimSpace(req.EntityType) == "" {
		return domain.LockAcquireRequest{}, errors.New("entity_type is required")
	}
	if strings.TrimSpace(req.EntityID) == "" {
		return domain.LockAcquireRequest{}, errors.New("entity_id is required")
	}
	if req.ExpiresAt == nil && strings.TrimSpace(req.TTL) == "" {
		req.TTL = "1h"
	}
	return req, nil
}

func parseQueryLimit(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	var limit int
	if _, err := fmt.Sscanf(value, "%d", &limit); err != nil || limit <= 0 {
		return fallback
	}
	return limit
}

func splitCSVQuery(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func decodeTaskUpdateRequest(body io.Reader) (domain.TaskUpdateRequest, error) {
	var req domain.TaskUpdateRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return domain.TaskUpdateRequest{}, err
	}
	if req.Title == nil &&
		req.Description == nil &&
		req.Priority == nil &&
		req.Status == nil &&
		req.Tags == nil &&
		req.EpicID == nil &&
		!req.NoEpic &&
		req.AssignedTo == nil &&
		req.AcceptanceCriteria == nil &&
		req.CompletionSummary == nil &&
		req.CompletionEvidence == nil {
		return domain.TaskUpdateRequest{}, errors.New("at least one field must be provided")
	}
	return req, nil
}

func isAllowedTaskStatus(value string) bool {
	switch domain.Status(value) {
	case domain.StatusTodo, domain.StatusInProgress, domain.StatusCompleted, domain.StatusWontFix, domain.StatusArchived:
		return true
	default:
		return false
	}
}

func isAllowedEpicStatus(value string) bool {
	switch domain.Status(value) {
	case domain.StatusTodo, domain.StatusInProgress, domain.StatusCompleted, domain.StatusArchived:
		return true
	default:
		return false
	}
}
