package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

func isAllowedTaskStatus(value string) bool {
	switch domain.Status(value) {
	case domain.StatusTodo, domain.StatusInProgress, domain.StatusCompleted, domain.StatusWontFix, domain.StatusArchived:
		return true
	default:
		return false
	}
}
