package server

import (
	"encoding/json"
	"net/http"
)

type Config struct {
	Addr        string
	PostgresDSN string
}

func New(config Config) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, map[string]string{
			"status":   "ok",
			"database": databaseState(config.PostgresDSN),
		})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>trakkr</title></head>
<body>
<main>
<h1>trakkr dashboard</h1>
<p>Server scaffold is running. API and dashboard modules will be wired to Postgres-backed task data.</p>
</main>
</body>
</html>`))
	})

	return &http.Server{
		Addr:    config.Addr,
		Handler: mux,
	}
}

func databaseState(dsn string) string {
	if dsn == "" {
		return "not_configured"
	}
	return "configured"
}

func respondJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(value)
}
