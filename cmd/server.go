package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	dsn  string
	addr string
}

type application struct {
	config config
	db     *pgxpool.Pool
}

func (app *application) mount() (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", app.handleHealth)

	// TODO: register domain routes here, e.g.
	// mux.Handle("/v1/roles/", roles.NewHandler(app.db))

	handler := recoverMiddleware(loggingMiddleware(mux))

	srv := &http.Server{
		Addr:              app.config.addr,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	return srv, nil
}

func (app *application) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := app.db.Ping(r.Context()); err != nil {
		slog.Error("health check: database ping failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) run(srv *http.Server) error {
	lis, err := net.Listen("tcp", app.config.addr)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	served := make(chan error, 1)
	go func() {
		slog.Info("http server has started", "address", app.config.addr)
		served <- srv.Serve(lis)
	}()

	select {
	case err := <-served:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		return srv.Shutdown(context.Background())
	}
}
