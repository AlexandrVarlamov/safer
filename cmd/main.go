package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aleksandrvarlamov/safer/internal/config"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/handlers/multihandler"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/handlers/slogpretty"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/sl"
	"github.com/aleksandrvarlamov/safer/internal/service/document"
	"github.com/aleksandrvarlamov/safer/internal/storage"
	"github.com/aleksandrvarlamov/safer/internal/transport/handlers/url/download"
	"github.com/aleksandrvarlamov/safer/internal/transport/handlers/url/save"
	mwLogger "github.com/aleksandrvarlamov/safer/internal/transport/middleware/logger"
	"github.com/aleksandrvarlamov/safer/internal/transport/middleware/recoverer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envProd  = "prod"

	// Таймауты остаются в коде: они связаны с логикой загрузки
	// (см. обсуждение больших файлов), а не с окружением.
	serverReadTimeout  = 10 * time.Second
	serverWriteTimeout = 10 * time.Second
	serverIdleTimeout  = 60 * time.Second
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	config, err := config.Load(*configPath)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("config load failed", sl.Err(err))
		os.Exit(1)
	}

	log := setupLogger(config.Env, config.Log)

	log.Info("starting safer server",
		slog.String("addr", config.Server.Addr),
	)

	R2client, err := storage.NewR2(context.Background(), config)
	if err != nil {
		log.Error("failed to init R2 client", sl.Err(err))
		os.Exit(1)
	}

	log.Info("R2 client initialized")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(recoverer.New())

	documentService := document.New(R2client)

	router.Put("/save", save.New(config.Storage.MaxUploadBytes, documentService))

	router.Get("/download/{bucket}/{key}", download.Get(documentService))

	srv := &http.Server{
		Addr:         config.Server.Addr,
		Handler:      router,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string, logCfg config.Log) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(logCfg.Level)}

	var primary slog.Handler
	if env == envLocal {
		primary = slogpretty.PrettyHandlerOptions{SlogOpts: opts}.NewPrettyHandler(os.Stdout)
	} else {
		primary = slog.NewJSONHandler(os.Stdout, opts)
	}

	handler := primary
	if logCfg.File != "" {
		file, err := os.OpenFile(logCfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("failed to open log file", sl.Err(err))
			os.Exit(1)
		}
		handler = multihandler.New(primary, slog.NewJSONHandler(file, opts))
	}

	log := slog.New(handler).With(
		slog.String("service", "safer"),
		slog.String("env", env),
	)
	slog.SetDefault(log)

	return log
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		return slog.LevelInfo
	}
	return l
}
