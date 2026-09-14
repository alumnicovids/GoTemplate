package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alumnicovids/template-go/pkg/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	const defaultDSN = "host=localhost port=5432 user=postgres password=postgres dbname=myapp sslmode=disable"

	cfg := config{
		dsn:  env.GetString("DATABASE_URL", defaultDSN),
		addr: ":3000",
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	pool, err := pgxpool.New(ctx, cfg.dsn)
	if err != nil {
		slog.Error("connecting to database", "error", err)
	}
	defer pool.Close()

	slog.Info("connected to database")

	api := application{
		config: cfg,
		db:     pool,
	}

	srv, err := api.mount()
	if err != nil {
		slog.Error("mounting server", "error", err)
		return
	}

	if err := api.run(srv); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
