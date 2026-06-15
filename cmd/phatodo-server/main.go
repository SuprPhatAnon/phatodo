package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/server"
	"github.com/SuprPhatAnon/phatodo/internal/storage/postgres"
)

func main() {
	addr := env("PHATODO_ADDR", ":8080")
	postgresDSN := env("PHATODO_DATABASE_URL", "")
	ctx := context.Background()

	var projectConfigStore server.ProjectConfigReader
	var projectConfigWriter server.ProjectConfigWriter
	var taskCreator server.TaskCreator
	var bootstrapManager server.BootstrapManager
	if postgresDSN != "" {
		store, err := postgres.NewStore(ctx, postgresDSN)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to postgres: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		projectConfigStore = store
		projectConfigWriter = store
		taskCreator = store
		bootstrapManager = store
	}

	srv := server.New(server.Config{
		Addr:                addr,
		PostgresDSN:         postgresDSN,
		ProjectConfigReader: projectConfigStore,
		ProjectConfigWriter: projectConfigWriter,
		TaskCreator:         taskCreator,
		BootstrapManager:    bootstrapManager,
	})

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		slog.Info("starting phatodo server", "addr", addr)
		errs <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
