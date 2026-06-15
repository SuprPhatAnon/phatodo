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
	SubtaskLister       SubtaskLister
	TaskReader          TaskReader
	TaskUpdater         TaskUpdater
	TaskDeleter         TaskDeleter
	CommentLister       CommentLister
	CommentCreator      CommentCreator
	CommentUpdater      CommentUpdater
	CommentDeleter      CommentDeleter
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

type SubtaskLister interface {
	ListSubtasks(context.Context, string, string) (domain.TaskListResponse, error)
}

type TaskReader interface {
	GetTask(context.Context, string, string) (domain.TaskDetail, error)
}

type TaskUpdater interface {
	UpdateTask(context.Context, string, string, domain.TaskUpdateRequest, string) (domain.TaskDetail, error)
}

type TaskDeleter interface {
	DeleteTask(context.Context, string, string, string) (domain.TaskDetail, error)
}

type CommentLister interface {
	ListComments(context.Context, string, string) (domain.CommentListResponse, error)
}

type CommentCreator interface {
	CreateComment(context.Context, string, string, domain.CommentCreateRequest, string) (domain.Comment, error)
}

type CommentUpdater interface {
	UpdateComment(context.Context, string, string, domain.CommentUpdateRequest, string) (domain.Comment, error)
}

type CommentDeleter interface {
	DeleteComment(context.Context, string, string, string) (domain.Comment, error)
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
