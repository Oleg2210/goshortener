package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/handler"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
	compres "github.com/Oleg2210/goshortener/pkg/middleware/compress"
	"github.com/Oleg2210/goshortener/pkg/middleware/cookies"
	"github.com/Oleg2210/goshortener/pkg/middleware/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func chooseStorage(ctx context.Context, logger *zap.Logger) repository.URLRepository {
	if config.DatabaseInfo != "" {
		repo, err := repository.NewDBRepository(config.DatabaseInfo)

		if err == nil {
			return repo
		}

		logger.Error("failed to create db repo", zap.Error(err))
	}

	if config.FileStoragePath != "" {
		repo, err := repository.NewFileRepository(ctx, config.FileStoragePath)

		if err == nil {
			return repo
		}

		logger.Error("failed to create file repo", zap.Error(err))
	}

	return repository.NewMemoryRepository()
}

func main() {
	config.Load()
	router := chi.NewRouter()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := chooseStorage(ctx, logger)

	shortenerService := service.NewShortenerService(
		repo,
		config.MinLength,
		config.MaxLength,
	)

	deleter := handler.NewDeleter(ctx, logger, shortenerService, 1)

	app := handler.App{
		ShortenerService: shortenerService,
		Logger:           logger,
		Deleter:          deleter,
	}

	router.Use(logging.LoggingMiddleware(logger))
	router.Use(cookies.AuthMiddleware([]byte(config.AuthSecret)))
	router.Use(compres.GzipMiddleware)
	router.Get("/{id}", app.HandleGet)
	router.Post("/", app.HandlePost)
	router.Post("/api/shorten", app.HandlePostJSON)
	router.Post("/api/shorten/batch", app.HandlePostBatchJSON)
	router.Get("/ping", app.HandlePing)
	router.Get("/api/user/urls", app.HandleGetAllUserUrls)
	router.Delete("/api/user/urls", app.HandleMarkDelete)

	server := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
