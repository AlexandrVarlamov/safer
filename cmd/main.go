package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aleksandrvarlamov/safer/internal/config"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/handlers/slogpretty"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/sl"
	"github.com/aleksandrvarlamov/safer/internal/service/document"
	"github.com/aleksandrvarlamov/safer/internal/storage"
	"github.com/aleksandrvarlamov/safer/internal/transport/handlers/url/download"
	"github.com/aleksandrvarlamov/safer/internal/transport/handlers/url/save"
	mwLogger "github.com/aleksandrvarlamov/safer/internal/transport/middleware/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	config, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log := setupLogger(config.Env)

	log.Info("STARTING SAFER SERVER!")

	R2client, err := storage.NewR2(context.Background(), config)
	if err != nil {
		log.Error("Failed to init R2 client", sl.Err(err))
		os.Exit(1)
	}

	log.Info("R2 client initialized successfully!", slog.Any("client", R2client))

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	documentService := document.New(log, R2client)

	router.Put("/save", save.New(log, documentService))

	router.Get("/download/{bucket}/{key}", download.Get(log, documentService))

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Failed to start server", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
