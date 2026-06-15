package server

import "net/http"

type Config struct {
	Addr        string
	PostgresDSN string
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
