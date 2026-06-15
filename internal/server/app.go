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
	a.registerLockRoutes(mux)
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
	mux.Handle("GET /api/v1/projects/{projectID}/epics", a.withAPIAuth(http.HandlerFunc(a.listEpics)))
	mux.Handle("POST /api/v1/projects/{projectID}/epics", a.withAPIAuth(http.HandlerFunc(a.createEpic)))
	mux.Handle("GET /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.showEpic)))
	mux.Handle("PATCH /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.updateEpic)))
	mux.Handle("POST /api/v1/projects/{projectID}/epics/{epicID}/complete", a.withAPIAuth(http.HandlerFunc(a.completeEpic)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/epics/{epicID}", a.withAPIAuth(http.HandlerFunc(a.deleteEpic)))
}

func (a *app) registerTaskRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks", a.withAPIAuth(http.HandlerFunc(a.listTasks)))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks", a.withAPIAuth(http.HandlerFunc(a.createTask)))
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.showTask)))
	mux.Handle("PATCH /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.updateTask)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/tasks/{taskID}", a.withAPIAuth(http.HandlerFunc(a.deleteTask)))
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/subtasks", a.withAPIAuth(http.HandlerFunc(a.listSubtasks)))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/subtasks", a.withAPIAuth(http.HandlerFunc(a.createSubtask)))
}

func (a *app) registerCommentRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/comments", a.withAPIAuth(http.HandlerFunc(a.listComments)))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/comments", a.withAPIAuth(http.HandlerFunc(a.createComment)))
	mux.Handle("PATCH /api/v1/projects/{projectID}/comments/{commentID}", a.withAPIAuth(http.HandlerFunc(a.updateComment)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/comments/{commentID}", a.withAPIAuth(http.HandlerFunc(a.deleteComment)))
}

func (a *app) registerDependencyRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/tasks/{taskID}/dependencies", a.withAPIAuth(http.HandlerFunc(a.listDependencies)))
	mux.Handle("POST /api/v1/projects/{projectID}/tasks/{taskID}/dependencies", a.withAPIAuth(http.HandlerFunc(a.addDependency)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/tasks/{taskID}/dependencies/{dependsOnID}", a.withAPIAuth(http.HandlerFunc(a.removeDependency)))
}

func (a *app) registerLockRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/locks", a.withAPIAuth(http.HandlerFunc(a.listLocks)))
	mux.Handle("POST /api/v1/projects/{projectID}/locks", a.withAPIAuth(http.HandlerFunc(a.acquireLock)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/locks/{lockID}", a.withAPIAuth(http.HandlerFunc(a.releaseLock)))
}

func (a *app) registerConfigRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/config", a.withAPIAuth(http.HandlerFunc(a.listProjectConfig)))
	mux.Handle("GET /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.getProjectConfig)))
	mux.Handle("PUT /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.setProjectConfig)))
	mux.Handle("DELETE /api/v1/projects/{projectID}/config/{key}", a.withAPIAuth(http.HandlerFunc(a.unsetProjectConfig)))
}

func (a *app) registerQueryRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/projects/{projectID}/search", a.withAPIAuth(http.HandlerFunc(a.search)))
	mux.Handle("GET /api/v1/projects/{projectID}/history", a.withAPIAuth(http.HandlerFunc(a.history)))
	mux.Handle("GET /api/v1/projects/{projectID}/list", a.withAPIAuth(http.HandlerFunc(a.listUnified)))
	mux.Handle("GET /api/v1/projects/{projectID}/ready", a.withAPIAuth(http.HandlerFunc(a.listReadyTasks)))
}
