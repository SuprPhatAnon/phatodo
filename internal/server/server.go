package server

import (
	"context"
	"net/http"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
)

type Config struct {
	Addr                string
	PostgresDSN         string
	ProjectConfigReader ProjectConfigReader
	ProjectConfigWriter ProjectConfigWriter
	TaskCreator         TaskCreator
	TaskLister          TaskLister
	ReadyLister         ReadyLister
	BootstrapManager    BootstrapManager
}

type ProjectConfigReader interface {
	ListProjectConfig(context.Context, string) ([]domain.ProjectConfig, error)
	GetProjectConfig(context.Context, string, string) (domain.ProjectConfig, error)
}

type ProjectConfigWriter interface {
	SetProjectConfig(context.Context, string, string, string) (domain.ProjectConfig, error)
	DeleteProjectConfig(context.Context, string, string) (domain.ProjectConfig, error)
}

type TaskCreator interface {
	CreateTask(context.Context, string, domain.TaskCreateRequest, string) (domain.TaskCreateResponse, error)
}

type TaskLister interface {
	ListTasks(context.Context, string, string, string) (domain.TaskListResponse, error)
}

type ReadyLister interface {
	ListReadyTasks(context.Context, string, string) (domain.ReadyListResponse, error)
}

type BootstrapManager interface {
	InitAdmin(context.Context, domain.AdminInitRequest) (domain.AdminInitResponse, error)
	BootstrapProject(context.Context, domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error)
}

func New(config Config) *http.Server {
	app := newApp(config)

	return &http.Server{
		Addr:    config.Addr,
		Handler: app.routes(),
	}
}

func databaseState(dsn string) string {
	if dsn == "" {
		return "not_configured"
	}
	return "configured"
}
