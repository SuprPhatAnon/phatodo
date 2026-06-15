package server

import "net/http"

type app struct {
	config Config
}

func newApp(config Config) *app {
	return &app{config: config}
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.health)
	a.registerAdminRoutes(mux)
	mux.Handle("GET /api/v1", a.withAPIAuth(http.HandlerFunc(a.apiIndex)))

	a.registerProjectRoutes(mux)
	a.registerEpicRoutes(mux)
	a.registerTaskRoutes(mux)
	a.registerCommentRoutes(mux)
	a.registerDependencyRoutes(mux)
	a.registerConfigRoutes(mux)
	a.registerQueryRoutes(mux)

	mux.HandleFunc("GET /", a.dashboardPlaceholder)

	return mux
}

func (a *app) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/init", a.adminInit)
	mux.HandleFunc("POST /api/v1/admin/bootstrap", a.adminBootstrap)
}

func (a *app) registerProjectRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects", a.withAPIAuth(http.HandlerFunc(a.notImplemented("project.list"))))
	mux.Handle("POST /api/v1/projects", a.withAPIAuth(http.HandlerFunc(a.notImplemented("project.create"))))
	mux.Handle("GET /api/v1/projects/{projectID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("project.show"))))
	mux.Handle("PATCH /api/v1/projects/{projectID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("project.update"))))
	mux.Handle("DELETE /api/v1/projects/{projectID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("project.delete"))))
}

func (a *app) registerEpicRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/epics", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.list"))))
	mux.Handle("POST /api/v1/projects/{projectID}/epics", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.create"))))
	mux.Handle("GET /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.show"))))
	mux.Handle("PATCH /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.update"))))
	mux.Handle("POST /api/v1/projects/{projectID}/epics/{epicID}/complete", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.complete"))))
	mux.Handle("DELETE /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("epic.delete"))))
}

func (a *app) registerTaskRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks", a.withAPIAuth(http.HandlerFunc(a.notImplemented("task.list"))))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks", a.withAPIAuth(http.HandlerFunc(a.notImplemented("task.create"))))
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("task.show"))))
	mux.Handle("PATCH /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("task.update"))))
	mux.Handle("DELETE /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("task.delete"))))
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/subtasks", a.withAPIAuth(http.HandlerFunc(a.notImplemented("subtask.list"))))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/subtasks", a.withAPIAuth(http.HandlerFunc(a.notImplemented("subtask.create"))))
}

func (a *app) registerCommentRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/comments", a.withAPIAuth(http.HandlerFunc(a.notImplemented("comment.list"))))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/comments", a.withAPIAuth(http.HandlerFunc(a.notImplemented("comment.add"))))
	mux.Handle("PATCH /api/v1/projects/{projectID}/comments/{commentID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("comment.update"))))
	mux.Handle("DELETE /api/v1/projects/{projectID}/comments/{commentID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("comment.delete"))))
}

func (a *app) registerDependencyRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/dependencies", a.withAPIAuth(http.HandlerFunc(a.notImplemented("dep.list"))))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/dependencies", a.withAPIAuth(http.HandlerFunc(a.notImplemented("dep.add"))))
	mux.Handle("DELETE /api/v1/projects/{projectID}/tasks/{taskID}/dependencies/{dependsOnID}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("dep.remove"))))
}

func (a *app) registerConfigRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/config", a.withAPIAuth(http.HandlerFunc(a.listProjectConfig)))
	mux.Handle("GET /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("config.get"))))
	mux.Handle("PUT /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.setProjectConfig)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.notImplemented("config.unset"))))
}

func (a *app) registerQueryRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/search", a.withAPIAuth(http.HandlerFunc(a.notImplemented("search"))))
	mux.Handle("GET /api/v1/projects/{projectID}/history", a.withAPIAuth(http.HandlerFunc(a.notImplemented("history"))))
	mux.Handle("GET /api/v1/projects/{projectID}/list", a.withAPIAuth(http.HandlerFunc(a.notImplemented("list"))))
}
