package main

import (
	"fmt"
	"net/http"
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

func main() {
	config.ParseFlags()
	router := chi.NewRouter()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Errorf("failed to init zap logger: %w", err))
	}

	var repo repository.URLRepository
	var repoErr error

	if config.FileStoragePath != "" {
		repo, repoErr = repository.NewFileRepository(config.FileStoragePath)
	}

	if config.FileStoragePath == "" || repoErr != nil {
		repo = repository.NewMemoryRepository()
	}

	shortenerService := service.NewShortenerService(
		repo,
		config.MinLength,
		config.MaxLength,
	)

	app := handler.App{
		ShortenerService: shortenerService,
	}

	router.Use(logging.LoggingMiddleware(logger))
	router.Use(compres.GzipMiddleware)
	router.Get("/{id}", app.HandleGet)
	router.Post("/", app.HandlePost)
	router.Post("/api/shorten", app.HandlePostJSON)

	server := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
