package server

import "net/http"

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
