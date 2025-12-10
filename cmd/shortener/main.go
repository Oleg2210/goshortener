package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/handler"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
	compres "github.com/Oleg2210/goshortener/pkg/middleware/compress"
	"github.com/Oleg2210/goshortener/pkg/middleware/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func chooseStorage() repository.URLRepository {
	if config.DatabaseInfo != "" {
		repo, err := repository.NewDBRepository(config.DatabaseInfo)

		if err == nil {
			return repo
		}

		fmt.Fprintf(os.Stderr, "failed to create db repo: %v\n", err)
	}

	if config.FileStoragePath != "" {
		repo, err := repository.NewFileRepository(config.FileStoragePath)

		if err == nil {
			return repo
		}

		fmt.Fprintf(os.Stderr, "failed to create file repo: %v\n", err)
	}

	return repository.NewMemoryRepository()
}

func main() {
	config.ParseFlags()
	router := chi.NewRouter()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}

	repo := chooseStorage()

	shortenerService := service.NewShortenerService(
		repo,
		config.MinLength,
		config.MaxLength,
	)

	app := handler.App{
		ShortenerService: shortenerService,
		Logger:           logger,
	}

	router.Use(logging.LoggingMiddleware(logger))
	router.Use(compres.GzipMiddleware)
	router.Get("/{id}", app.HandleGet)
	router.Post("/", app.HandlePost)
	router.Post("/api/shorten", app.HandlePostJSON)
	router.Post("/api/shorten/batch", app.HandlePostBatchJSON)
	router.Get("/ping", app.HandlePing)

	server := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
