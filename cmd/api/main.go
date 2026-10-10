package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/novapromptgobackend/internal/config"
	"github.com/yourname/novapromptgobackend/internal/db"
	"github.com/yourname/novapromptgobackend/internal/handlers"
	"github.com/yourname/novapromptgobackend/internal/repository"
	"github.com/yourname/novapromptgobackend/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns, cfg.DBMaxConnLife, cfg.DBMaxConnIdle)
	if err != nil {
		slog.Error("db_pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	router := server.NewRouter(server.Deps{
		Health:       &handlers.HealthHandler{Pool: pool},
		Images:       &handlers.ImageHandler{Repo: repository.NewImageRepo(pool)},
		Categories:   &handlers.CategoryHandler{Repo: repository.NewCategoryRepo(pool)},
		Tags:         &handlers.TagHandler{Repo: repository.NewTagRepo(pool)},
		Descriptions: &handlers.DescriptionHandler{Repo: repository.NewDescriptionRepo(pool)},
		AdKeys:       &handlers.AdKeyHandler{Repo: repository.NewAdKeyRepo(pool)},
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("server_start", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown_initiated")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("shutdown_complete")
}